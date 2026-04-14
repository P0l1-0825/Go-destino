package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/P0l1-0825/Go-destino/internal/config"
	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/security"
)

// userRepo defines the subset of UserRepository methods used by AuthService.
type userRepo interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, tenantID, email string) (bool, error)
	ChangePassword(ctx context.Context, userID, newHash string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type AuthService struct {
	userRepo       userRepo
	jwtCfg         config.JWTConfig
	loginLimiter   security.LoginLimiterStore
	tokenBlacklist security.TokenBlacklistStore
	resetStore     security.PasswordResetTokenStore
	passwordPolicy security.PasswordPolicy
	auditFn        func(tenantID, userID, action, resource, resourceID, details, ip, ua string)
}

type AuthServiceConfig struct {
	UserRepo       userRepo
	JWTCfg         config.JWTConfig
	LoginLimiter   security.LoginLimiterStore
	TokenBlacklist security.TokenBlacklistStore
	ResetStore     security.PasswordResetTokenStore
	AuditFn        func(tenantID, userID, action, resource, resourceID, details, ip, ua string)
}

func NewAuthService(userRepo userRepo, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtCfg:         jwtCfg,
		loginLimiter:   security.NewLoginLimiter(5, 15*time.Minute, 30*time.Minute),
		tokenBlacklist: security.NewTokenBlacklist(),
		resetStore:     security.NewPasswordResetStore(),
		passwordPolicy: security.DefaultPasswordPolicy(),
	}
}

func NewAuthServiceFull(cfg AuthServiceConfig) *AuthService {
	s := &AuthService{
		userRepo:       cfg.UserRepo,
		jwtCfg:         cfg.JWTCfg,
		loginLimiter:   cfg.LoginLimiter,
		tokenBlacklist: cfg.TokenBlacklist,
		resetStore:     cfg.ResetStore,
		passwordPolicy: security.DefaultPasswordPolicy(),
		auditFn:        cfg.AuditFn,
	}
	if s.loginLimiter == nil {
		s.loginLimiter = security.NewLoginLimiter(5, 15*time.Minute, 30*time.Minute)
	}
	if s.tokenBlacklist == nil {
		s.tokenBlacklist = security.NewTokenBlacklist()
	}
	if s.resetStore == nil {
		s.resetStore = security.NewPasswordResetStore()
	}
	return s
}

type Claims struct {
	jwt.RegisteredClaims
	Role        domain.UserRole    `json:"role"`
	TenantID    string             `json:"tenant_id"`
	Permissions []domain.Permission `json:"permissions"`
}

func (s *AuthService) Register(ctx context.Context, tenantID string, req domain.CreateUserRequest) (*domain.User, error) {
	// Check for duplicate email
	exists, err := s.userRepo.ExistsByEmail(ctx, tenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// SECURITY: Self-registration always gets USUARIO role.
	// Elevated roles can only be assigned via admin CreateUser endpoint.
	if req.Role != domain.RoleUsuario && req.Role != "" {
		req.Role = domain.RoleUsuario
	}
	if req.Role == "" {
		req.Role = domain.RoleUsuario
	}
	if !validRole(req.Role) {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	// Validate password strength
	if err := s.passwordPolicy.Validate(req.Password); err != nil {
		return nil, err
	}

	// Default language
	if req.Lang == "" {
		req.Lang = "es"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Name:         req.Name,
		Role:         req.Role,
		SubRole:      req.SubRole,
		CompanyID:    req.CompanyID,
		Lang:         req.Lang,
		Active:       true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.audit(tenantID, user.ID, "user.register", "user", user.ID, "new user registered", "", "")

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, tenantID string, req domain.LoginRequest, ip, userAgent string) (*domain.LoginResponse, error) {
	loginKey := tenantID + ":" + req.Email

	// Check rate limit before doing any work
	if err := s.loginLimiter.Check(loginKey); err != nil {
		s.audit(tenantID, "", "login.locked", "auth", "", fmt.Sprintf("login blocked for %s: account locked", req.Email), ip, userAgent)
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, tenantID, req.Email)
	if err != nil {
		lockErr := s.loginLimiter.RecordFailure(loginKey)
		s.audit(tenantID, "", "login.failed", "auth", "", fmt.Sprintf("login failed for %s: user not found", req.Email), ip, userAgent)
		if lockErr != nil {
			return nil, lockErr
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		lockErr := s.loginLimiter.RecordFailure(loginKey)
		s.audit(tenantID, user.ID, "login.failed", "auth", user.ID, fmt.Sprintf("login failed for %s: wrong password", req.Email), ip, userAgent)
		if lockErr != nil {
			return nil, lockErr
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.Active {
		s.audit(tenantID, user.ID, "login.disabled", "auth", user.ID, fmt.Sprintf("login attempt on disabled account %s", req.Email), ip, userAgent)
		return nil, fmt.Errorf("account is disabled")
	}

	// Login success — clear rate limiter
	s.loginLimiter.RecordSuccess(loginKey)

	accessToken, err := s.generateToken(user, time.Duration(s.jwtCfg.ExpireHour)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	// Update last login asynchronously
	go func() {
		_ = s.userRepo.UpdateLastLogin(context.Background(), user.ID)
	}()

	s.audit(tenantID, user.ID, "login.success", "auth", user.ID, fmt.Sprintf("user %s logged in", req.Email), ip, userAgent)

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) Logout(tokenStr string) error {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	if claims.ExpiresAt != nil {
		s.tokenBlacklist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}

	s.audit(claims.TenantID, claims.Subject, "logout", "auth", claims.Subject, "user logged out", "", "")

	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.Active {
		return nil, fmt.Errorf("account is disabled")
	}

	// Revoke the old refresh token
	if claims.ExpiresAt != nil {
		s.tokenBlacklist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}

	newAccessToken, err := s.generateToken(user, time.Duration(s.jwtCfg.ExpireHour)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &domain.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	if err := s.passwordPolicy.Validate(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.ChangePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	s.audit(user.TenantID, userID, "password.changed", "user", userID, "password changed", "", "")

	return nil
}

// RequestPasswordReset generates a reset token (in production, send via email).
func (s *AuthService) RequestPasswordReset(ctx context.Context, tenantID, email string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal whether the email exists
		return "", nil
	}

	token, err := s.resetStore.CreateToken(user.ID, tenantID, email, 1*time.Hour)
	if err != nil {
		return "", err
	}

	s.audit(tenantID, user.ID, "password.reset_requested", "user", user.ID, "password reset requested", "", "")

	// In production: send email with the token — never log the full token
	log.Printf("[SECURITY] Password reset token generated for %s (token=%s…)", email, token[:8])

	return token, nil
}

// ResetPassword validates a reset token and sets a new password.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	userID, err := s.resetStore.ValidateToken(token)
	if err != nil {
		return err
	}

	if err := s.passwordPolicy.Validate(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.ChangePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	s.audit("", userID, "password.reset_completed", "user", userID, "password reset completed via token", "", "")

	return nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token has been revoked
	if claims.ID != "" && s.tokenBlacklist.IsRevoked(claims.ID) {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

func (s *AuthService) generateToken(user *domain.User, duration time.Duration) (string, error) {
	perms := domain.RolePermissions[user.Role]

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    "godestino",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		Role:        user.Role,
		TenantID:    user.TenantID,
		Permissions: perms,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) generateRefreshToken(user *domain.User) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    "godestino",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		Role:     user.Role,
		TenantID: user.TenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) audit(tenantID, userID, action, resource, resourceID, details, ip, ua string) {
	if s.auditFn != nil {
		s.auditFn(tenantID, userID, action, resource, resourceID, details, ip, ua)
	}
}

func validRole(role domain.UserRole) bool {
	switch role {
	case domain.RoleSuperAdmin, domain.RoleAdmin, domain.RoleClienteConcesion,
		domain.RoleTesoreriaCliente, domain.RoleMesaControl, domain.RoleOperador,
		domain.RoleTaxista, domain.RoleVendedor, domain.RoleBroker, domain.RoleUsuario:
		return true
	}
	return false
}
