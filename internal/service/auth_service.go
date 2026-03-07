package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/P0l1-0825/Go-destino/internal/config"
	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, jwtCfg: jwtCfg}
}

type Claims struct {
	jwt.RegisteredClaims
	Role        domain.UserRole    `json:"role"`
	TenantID    string             `json:"tenant_id"`
	Permissions []domain.Permission `json:"permissions"`
}

func (s *AuthService) Register(ctx context.Context, tenantID string, req domain.CreateUserRequest) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
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

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, tenantID string, req domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, tenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.Active {
		return nil, fmt.Errorf("account is disabled")
	}

	accessToken, err := s.generateToken(user, time.Duration(s.jwtCfg.ExpireHour)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := s.generateToken(user, 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
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
		},
		Role:        user.Role,
		TenantID:    user.TenantID,
		Permissions: perms,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}
