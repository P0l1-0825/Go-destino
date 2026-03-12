package security

import "time"

// TokenBlacklistStore defines the interface for JWT token revocation.
// Implementations: TokenBlacklist (in-memory), RedisTokenBlacklist (Redis).
type TokenBlacklistStore interface {
	Revoke(jti string, expiresAt time.Time)
	IsRevoked(jti string) bool
}

// LoginLimiterStore defines the interface for login rate limiting.
// Implementations: LoginLimiter (in-memory), RedisLoginLimiter (Redis).
type LoginLimiterStore interface {
	Check(email string) error
	RecordFailure(email string) error
	RecordSuccess(email string)
	RemainingAttempts(email string) int
}

// PasswordResetTokenStore defines the interface for password reset tokens.
// Implementations: PasswordResetStore (in-memory), RedisPasswordResetStore (Redis).
type PasswordResetTokenStore interface {
	CreateToken(userID, tenantID, email string, ttl time.Duration) (string, error)
	ValidateToken(token string) (userID string, err error)
}
