package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisPasswordResetStore stores password reset tokens in Redis with automatic
// TTL-based expiration. Each token is stored as a JSON-encoded value
// containing user metadata and a one-time-use flag.
type RedisPasswordResetStore struct {
	rdb    *redis.Client
	prefix string
}

type redisResetToken struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Used     bool   `json:"used"`
}

// NewRedisPasswordResetStore creates a Redis-backed password reset store.
func NewRedisPasswordResetStore(rdb *redis.Client) *RedisPasswordResetStore {
	return &RedisPasswordResetStore{rdb: rdb, prefix: "pr:"}
}

// CreateToken generates a password reset token valid for the given duration.
func (s *RedisPasswordResetStore) CreateToken(userID, tenantID, email string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating reset token: %w", err)
	}

	token := hex.EncodeToString(b)

	data, err := json.Marshal(redisResetToken{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
	})
	if err != nil {
		return "", fmt.Errorf("marshaling reset token: %w", err)
	}

	ctx := context.Background()
	if err := s.rdb.Set(ctx, s.prefix+token, data, ttl).Err(); err != nil {
		return "", fmt.Errorf("storing reset token: %w", err)
	}

	return token, nil
}

// ValidateToken checks if a reset token is valid and returns the user ID.
// The token is marked as used after validation (one-time use).
func (s *RedisPasswordResetStore) ValidateToken(token string) (string, error) {
	ctx := context.Background()
	key := s.prefix + token

	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return "", fmt.Errorf("invalid or expired reset token")
	}

	var rt redisResetToken
	if err := json.Unmarshal(data, &rt); err != nil {
		return "", fmt.Errorf("invalid reset token data")
	}

	if rt.Used {
		return "", fmt.Errorf("reset token already used")
	}

	// Mark as used and keep for 5 minutes so repeated attempts get "already used"
	rt.Used = true
	updated, _ := json.Marshal(rt)
	s.rdb.Set(ctx, key, updated, 5*time.Minute)

	return rt.UserID, nil
}
