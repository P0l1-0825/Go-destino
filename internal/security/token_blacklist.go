package security

import (
	"sync"
	"time"
)

// TokenBlacklist stores revoked JWT token IDs (JTI).
// In production, replace with Redis SET with TTL.
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // JTI → expiration time
}

// NewTokenBlacklist creates a new in-memory token blacklist.
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	// Cleanup expired entries every 10 minutes
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			bl.cleanup()
		}
	}()

	return bl
}

// Revoke adds a token JTI to the blacklist until its expiration.
func (bl *TokenBlacklist) Revoke(jti string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.tokens[jti] = expiresAt
}

// IsRevoked checks if a token JTI has been revoked.
func (bl *TokenBlacklist) IsRevoked(jti string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	exp, ok := bl.tokens[jti]
	if !ok {
		return false
	}

	// If the token has naturally expired, remove it
	if time.Now().After(exp) {
		return false
	}

	return true
}

func (bl *TokenBlacklist) cleanup() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	for jti, exp := range bl.tokens {
		if now.After(exp) {
			delete(bl.tokens, jti)
		}
	}
}
