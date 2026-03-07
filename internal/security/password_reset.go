package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// PasswordResetStore stores password reset tokens.
// In production, persist to Redis or database.
type PasswordResetStore struct {
	mu     sync.Mutex
	tokens map[string]*resetToken
}

type resetToken struct {
	UserID    string
	TenantID  string
	Email     string
	ExpiresAt time.Time
	Used      bool
}

// NewPasswordResetStore creates a new in-memory reset token store.
func NewPasswordResetStore() *PasswordResetStore {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.cleanup()
		}
	}()

	return s
}

// CreateToken generates a password reset token valid for the given duration.
func (s *PasswordResetStore) CreateToken(userID, tenantID, email string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating reset token: %w", err)
	}

	token := hex.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token] = &resetToken{
		UserID:    userID,
		TenantID:  tenantID,
		Email:     email,
		ExpiresAt: time.Now().Add(ttl),
	}

	return token, nil
}

// ValidateToken checks if a reset token is valid and returns the user ID.
func (s *PasswordResetStore) ValidateToken(token string) (userID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rt, ok := s.tokens[token]
	if !ok {
		return "", fmt.Errorf("invalid or expired reset token")
	}

	if rt.Used {
		return "", fmt.Errorf("reset token already used")
	}

	if time.Now().After(rt.ExpiresAt) {
		delete(s.tokens, token)
		return "", fmt.Errorf("reset token has expired")
	}

	// Mark as used
	rt.Used = true

	return rt.UserID, nil
}

func (s *PasswordResetStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, rt := range s.tokens {
		if now.After(rt.ExpiresAt) || rt.Used {
			delete(s.tokens, token)
		}
	}
}
