package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/P0l1-0825/Go-destino/internal/config"
	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/security"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
	"github.com/P0l1-0825/Go-destino/internal/testutil/mocks"
)

// newTestAuthService creates an AuthService with mock dependencies for testing.
func newTestAuthService(mockRepo *mocks.MockUserRepo) *AuthService {
	return NewAuthServiceFull(AuthServiceConfig{
		UserRepo:       mockRepo,
		JWTCfg:         testutil.NewTestJWTConfig(),
		LoginLimiter:   security.NewLoginLimiter(5, 15*time.Minute, 30*time.Minute),
		TokenBlacklist: security.NewTokenBlacklist(),
		ResetStore:     security.NewPasswordResetStore(),
	})
}

// hashPassword is a test helper that creates a bcrypt hash.
func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4) // low cost for speed
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}
	return string(hash)
}

// activeTestUser returns a user with a valid bcrypt hash for the given password.
func activeTestUser(t *testing.T, password string) *domain.User {
	t.Helper()
	u := testutil.NewTestUser()
	u.PasswordHash = hashPassword(t, password)
	u.Active = true
	return u
}

// --- Register ---

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name      string
		req       domain.CreateUserRequest
		mockSetup func(*mocks.MockUserRepo)
		wantErr   string
	}{
		{
			name: "happy path",
			req: domain.CreateUserRequest{
				Email:    "new@test.com",
				Password: "StrongPass1",
				Name:     "New User",
				Role:     domain.RoleUsuario,
			},
			mockSetup: func(m *mocks.MockUserRepo) {
				m.ExistsByEmailFn = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
				m.CreateFn = func(_ context.Context, _ *domain.User) error { return nil }
			},
		},
		{
			name: "duplicate email",
			req: domain.CreateUserRequest{
				Email:    "existing@test.com",
				Password: "StrongPass1",
				Name:     "Dup",
				Role:     domain.RoleUsuario,
			},
			mockSetup: func(m *mocks.MockUserRepo) {
				m.ExistsByEmailFn = func(_ context.Context, _, _ string) (bool, error) { return true, nil }
			},
			wantErr: "email already registered",
		},
		{
			name: "invalid role",
			req: domain.CreateUserRequest{
				Email:    "role@test.com",
				Password: "StrongPass1",
				Name:     "Role",
				Role:     "INVALID_ROLE",
			},
			mockSetup: func(m *mocks.MockUserRepo) {
				m.ExistsByEmailFn = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
			},
			wantErr: "invalid role",
		},
		{
			name: "weak password",
			req: domain.CreateUserRequest{
				Email:    "weak@test.com",
				Password: "123",
				Name:     "Weak",
				Role:     domain.RoleUsuario,
			},
			mockSetup: func(m *mocks.MockUserRepo) {
				m.ExistsByEmailFn = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
			},
			wantErr: "at least 8 characters",
		},
		{
			name: "repo create error",
			req: domain.CreateUserRequest{
				Email:    "err@test.com",
				Password: "StrongPass1",
				Name:     "Err",
				Role:     domain.RoleUsuario,
			},
			mockSetup: func(m *mocks.MockUserRepo) {
				m.ExistsByEmailFn = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
				m.CreateFn = func(_ context.Context, _ *domain.User) error { return fmt.Errorf("db connection lost") }
			},
			wantErr: "creating user",
		},
		{
			name: "empty lang defaults to es",
			req: domain.CreateUserRequest{
				Email:    "lang@test.com",
				Password: "StrongPass1",
				Name:     "Lang",
				Role:     domain.RoleUsuario,
				Lang:     "",
			},
			mockSetup: func(m *mocks.MockUserRepo) {
				m.ExistsByEmailFn = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
				m.CreateFn = func(_ context.Context, u *domain.User) error {
					if u.Lang != "es" {
						return fmt.Errorf("expected lang=es, got %s", u.Lang)
					}
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mocks.MockUserRepo{}
			tt.mockSetup(mock)
			svc := newTestAuthService(mock)

			user, err := svc.Register(context.Background(), testutil.TestTenantID, tt.req)
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if user == nil {
				t.Fatal("expected user, got nil")
			}
			if user.TenantID != testutil.TestTenantID {
				t.Errorf("tenant_id = %s, want %s", user.TenantID, testutil.TestTenantID)
			}
		})
	}
}

// --- Login ---

func TestAuthService_Login(t *testing.T) {
	password := "ValidPass1"

	tests := []struct {
		name      string
		email     string
		password  string
		mockSetup func(*mocks.MockUserRepo)
		wantErr   string
	}{
		{
			name:     "happy path",
			email:    "user@test.com",
			password: password,
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByEmailFn = func(_ context.Context, _, _ string) (*domain.User, error) {
					return activeTestUser(t, password), nil
				}
				m.UpdateLastLoginFn = func(_ context.Context, _ string) error { return nil }
			},
		},
		{
			name:     "user not found",
			email:    "unknown@test.com",
			password: password,
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByEmailFn = func(_ context.Context, _, _ string) (*domain.User, error) {
					return nil, fmt.Errorf("sql: no rows")
				}
			},
			wantErr: "invalid credentials",
		},
		{
			name:     "wrong password",
			email:    "user@test.com",
			password: "WrongPass1",
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByEmailFn = func(_ context.Context, _, _ string) (*domain.User, error) {
					return activeTestUser(t, password), nil
				}
			},
			wantErr: "invalid credentials",
		},
		{
			name:     "inactive account",
			email:    "disabled@test.com",
			password: password,
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByEmailFn = func(_ context.Context, _, _ string) (*domain.User, error) {
					u := activeTestUser(t, password)
					u.Active = false
					return u, nil
				}
			},
			wantErr: "account is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mocks.MockUserRepo{}
			tt.mockSetup(mock)
			svc := newTestAuthService(mock)

			resp, err := svc.Login(context.Background(), testutil.TestTenantID,
				domain.LoginRequest{Email: tt.email, Password: tt.password}, "127.0.0.1", "test-agent")
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if resp.AccessToken == "" {
				t.Error("expected access token")
			}
			if resp.RefreshToken == "" {
				t.Error("expected refresh token")
			}
		})
	}
}

// --- Logout ---

func TestAuthService_Logout(t *testing.T) {
	mock := &mocks.MockUserRepo{}
	svc := newTestAuthService(mock)

	// Generate a valid token first
	user := testutil.NewTestUser()
	token, err := svc.generateToken(user, 1*time.Hour)
	testutil.AssertNoError(t, err)

	// Logout should succeed
	err = svc.Logout(token)
	testutil.AssertNoError(t, err)

	// Token should now be revoked
	_, err = svc.ValidateToken(token)
	testutil.AssertError(t, err, "revoked")
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	mock := &mocks.MockUserRepo{}
	svc := newTestAuthService(mock)

	err := svc.Logout("invalid-token-string")
	testutil.AssertError(t, err, "invalid token")
}

// --- RefreshToken ---

func TestAuthService_RefreshToken(t *testing.T) {
	password := "ValidPass1"

	tests := []struct {
		name      string
		setupUser func() *domain.User
		wantErr   string
	}{
		{
			name: "happy path",
			setupUser: func() *domain.User {
				return activeTestUser(t, password)
			},
		},
		{
			name: "inactive user",
			setupUser: func() *domain.User {
				u := activeTestUser(t, password)
				u.Active = false
				return u
			},
			wantErr: "account is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setupUser()
			mock := &mocks.MockUserRepo{
				GetByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return user, nil
				},
			}
			svc := newTestAuthService(mock)

			// Generate a refresh token
			refreshToken, err := svc.generateRefreshToken(user)
			testutil.AssertNoError(t, err)

			resp, err := svc.RefreshToken(context.Background(), refreshToken)
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if resp.AccessToken == "" {
				t.Error("expected new access token")
			}

			// Old refresh token should be revoked
			_, err = svc.ValidateToken(refreshToken)
			testutil.AssertError(t, err, "revoked")
		})
	}
}

func TestAuthService_RefreshToken_Invalid(t *testing.T) {
	mock := &mocks.MockUserRepo{}
	svc := newTestAuthService(mock)

	_, err := svc.RefreshToken(context.Background(), "garbage")
	testutil.AssertError(t, err, "invalid refresh token")
}

// --- ChangePassword ---

func TestAuthService_ChangePassword(t *testing.T) {
	oldPass := "OldPass123"
	newPass := "NewPass456"

	tests := []struct {
		name      string
		oldPass   string
		newPass   string
		mockSetup func(*mocks.MockUserRepo)
		wantErr   string
	}{
		{
			name:    "happy path",
			oldPass: oldPass,
			newPass: newPass,
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByIDFn = func(_ context.Context, _ string) (*domain.User, error) {
					return activeTestUser(t, oldPass), nil
				}
				m.ChangePasswordFn = func(_ context.Context, _, _ string) error { return nil }
			},
		},
		{
			name:    "wrong old password",
			oldPass: "WrongOld1",
			newPass: newPass,
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByIDFn = func(_ context.Context, _ string) (*domain.User, error) {
					return activeTestUser(t, oldPass), nil
				}
			},
			wantErr: "current password is incorrect",
		},
		{
			name:    "weak new password",
			oldPass: oldPass,
			newPass: "weak",
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByIDFn = func(_ context.Context, _ string) (*domain.User, error) {
					return activeTestUser(t, oldPass), nil
				}
			},
			wantErr: "at least 8 characters",
		},
		{
			name:    "user not found",
			oldPass: oldPass,
			newPass: newPass,
			mockSetup: func(m *mocks.MockUserRepo) {
				m.GetByIDFn = func(_ context.Context, _ string) (*domain.User, error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mocks.MockUserRepo{}
			tt.mockSetup(mock)
			svc := newTestAuthService(mock)

			err := svc.ChangePassword(context.Background(), testutil.TestUserID, tt.oldPass, tt.newPass)
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
		})
	}
}

// --- RequestPasswordReset ---

func TestAuthService_RequestPasswordReset(t *testing.T) {
	t.Run("user exists", func(t *testing.T) {
		mock := &mocks.MockUserRepo{
			GetByEmailFn: func(_ context.Context, _, _ string) (*domain.User, error) {
				return testutil.NewTestUser(), nil
			},
		}
		svc := newTestAuthService(mock)

		token, err := svc.RequestPasswordReset(context.Background(), testutil.TestTenantID, "user@test.com")
		testutil.AssertNoError(t, err)
		if token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("user not found returns no error", func(t *testing.T) {
		mock := &mocks.MockUserRepo{
			GetByEmailFn: func(_ context.Context, _, _ string) (*domain.User, error) {
				return nil, fmt.Errorf("not found")
			},
		}
		svc := newTestAuthService(mock)

		token, err := svc.RequestPasswordReset(context.Background(), testutil.TestTenantID, "unknown@test.com")
		testutil.AssertNoError(t, err) // Should NOT leak that email doesn't exist
		if token != "" {
			t.Error("expected empty token for unknown user")
		}
	})
}

// --- ResetPassword ---

func TestAuthService_ResetPassword(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mock := &mocks.MockUserRepo{
			GetByEmailFn: func(_ context.Context, _, _ string) (*domain.User, error) {
				return testutil.NewTestUser(), nil
			},
			ChangePasswordFn: func(_ context.Context, _, _ string) error { return nil },
		}
		svc := newTestAuthService(mock)

		// Generate a valid reset token
		token, err := svc.RequestPasswordReset(context.Background(), testutil.TestTenantID, testutil.TestEmail)
		testutil.AssertNoError(t, err)

		err = svc.ResetPassword(context.Background(), token, "NewStrong1")
		testutil.AssertNoError(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		mock := &mocks.MockUserRepo{}
		svc := newTestAuthService(mock)

		err := svc.ResetPassword(context.Background(), "bad-token", "NewStrong1")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("weak new password", func(t *testing.T) {
		mock := &mocks.MockUserRepo{
			GetByEmailFn: func(_ context.Context, _, _ string) (*domain.User, error) {
				return testutil.NewTestUser(), nil
			},
		}
		svc := newTestAuthService(mock)

		token, _ := svc.RequestPasswordReset(context.Background(), testutil.TestTenantID, testutil.TestEmail)

		err := svc.ResetPassword(context.Background(), token, "weak")
		testutil.AssertError(t, err, "at least 8 characters")
	})
}

// --- ValidateToken ---

func TestAuthService_ValidateToken(t *testing.T) {
	mock := &mocks.MockUserRepo{}
	svc := newTestAuthService(mock)

	user := testutil.NewTestUser()

	t.Run("valid token", func(t *testing.T) {
		token, err := svc.generateToken(user, 1*time.Hour)
		testutil.AssertNoError(t, err)

		claims, err := svc.ValidateToken(token)
		testutil.AssertNoError(t, err)
		if claims.Subject != user.ID {
			t.Errorf("subject = %s, want %s", claims.Subject, user.ID)
		}
		if claims.TenantID != user.TenantID {
			t.Errorf("tenant_id = %s, want %s", claims.TenantID, user.TenantID)
		}
		if claims.Role != user.Role {
			t.Errorf("role = %s, want %s", claims.Role, user.Role)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := svc.generateToken(user, -1*time.Hour)
		testutil.AssertNoError(t, err)

		_, err = svc.ValidateToken(token)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("revoked token", func(t *testing.T) {
		token, err := svc.generateToken(user, 1*time.Hour)
		testutil.AssertNoError(t, err)

		// Revoke via logout
		_ = svc.Logout(token)

		_, err = svc.ValidateToken(token)
		testutil.AssertError(t, err, "revoked")
	})

	t.Run("wrong signing key", func(t *testing.T) {
		// Create a service with a different secret
		otherSvc := NewAuthServiceFull(AuthServiceConfig{
			UserRepo: mock,
			JWTCfg:   config.JWTConfig{Secret: "other-secret-that-is-at-least-32-characters-long", ExpireHour: 1},
		})
		token, _ := otherSvc.generateToken(user, 1*time.Hour)

		_, err := svc.ValidateToken(token)
		if err == nil {
			t.Error("expected error for wrong signing key")
		}
	})

	t.Run("garbage token", func(t *testing.T) {
		_, err := svc.ValidateToken("not.a.jwt")
		if err == nil {
			t.Error("expected error for garbage token")
		}
	})
}
