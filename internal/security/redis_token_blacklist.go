package security

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenBlacklist stores revoked JWT token IDs in Redis with automatic
// TTL-based expiration. Each revoked JTI is stored as a key with the
// remaining token lifetime as the TTL — no background cleanup needed.
type RedisTokenBlacklist struct {
	rdb    *redis.Client
	prefix string
}

// NewRedisTokenBlacklist creates a Redis-backed token blacklist.
func NewRedisTokenBlacklist(rdb *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{rdb: rdb, prefix: "bl:"}
}

// Revoke adds a token JTI to the blacklist until its expiration.
func (bl *RedisTokenBlacklist) Revoke(jti string, expiresAt time.Time) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return // token already expired, no need to blacklist
	}
	ctx := context.Background()
	bl.rdb.Set(ctx, bl.prefix+jti, "1", ttl)
}

// IsRevoked checks if a token JTI has been revoked.
// SECURITY: Fails CLOSED on Redis error — revoked tokens stay revoked even during outages.
func (bl *RedisTokenBlacklist) IsRevoked(jti string) bool {
	ctx := context.Background()
	result, err := bl.rdb.Exists(ctx, bl.prefix+jti).Result()
	if err != nil {
		// Fail CLOSED: if Redis is unreachable, treat token as revoked.
		// This prevents previously revoked tokens from being accepted during outages.
		log.Printf("[SECURITY] Redis blacklist check failed for JTI %s: %v — failing closed", jti[:8], err)
		return true
	}
	return result > 0
}
