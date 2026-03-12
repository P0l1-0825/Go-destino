package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLoginLimiter tracks failed login attempts in Redis with automatic
// TTL-based expiration. Uses two keys per email:
//   - ll:<email>:count — INCR counter with window TTL
//   - ll:<email>:lock  — lockout flag with lockout TTL
type RedisLoginLimiter struct {
	rdb         *redis.Client
	prefix      string
	maxAttempts int
	window      time.Duration
	lockout     time.Duration
}

// NewRedisLoginLimiter creates a Redis-backed login limiter.
func NewRedisLoginLimiter(rdb *redis.Client, maxAttempts int, window, lockout time.Duration) *RedisLoginLimiter {
	return &RedisLoginLimiter{
		rdb:         rdb,
		prefix:      "ll:",
		maxAttempts: maxAttempts,
		window:      window,
		lockout:     lockout,
	}
}

// Check returns an error if the email is currently locked out.
func (l *RedisLoginLimiter) Check(email string) error {
	ctx := context.Background()
	lockKey := l.prefix + email + ":lock"

	ttl, err := l.rdb.TTL(ctx, lockKey).Result()
	if err != nil {
		return nil // Redis error — fail open
	}
	if ttl > 0 {
		mins := int(ttl.Minutes()) + 1
		return fmt.Errorf("account temporarily locked, try again in %d minutes", mins)
	}
	return nil
}

// RecordFailure records a failed login attempt.
// Returns an error if the account is now locked.
func (l *RedisLoginLimiter) RecordFailure(email string) error {
	ctx := context.Background()
	countKey := l.prefix + email + ":count"
	lockKey := l.prefix + email + ":lock"

	count, err := l.rdb.Incr(ctx, countKey).Result()
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}

	// Set TTL on first attempt (INCR creates the key if missing)
	if count == 1 {
		l.rdb.Expire(ctx, countKey, l.window)
	}

	if int(count) >= l.maxAttempts {
		l.rdb.Set(ctx, lockKey, "1", l.lockout)
		l.rdb.Del(ctx, countKey)
		mins := int(l.lockout.Minutes())
		return fmt.Errorf("too many failed attempts, account locked for %d minutes", mins)
	}

	remaining := l.maxAttempts - int(count)
	return fmt.Errorf("invalid credentials (%d attempts remaining)", remaining)
}

// RecordSuccess clears the failure count for an email.
func (l *RedisLoginLimiter) RecordSuccess(email string) {
	ctx := context.Background()
	l.rdb.Del(ctx, l.prefix+email+":count", l.prefix+email+":lock")
}

// RemainingAttempts returns how many attempts are left for an email.
func (l *RedisLoginLimiter) RemainingAttempts(email string) int {
	ctx := context.Background()
	count, err := l.rdb.Get(ctx, l.prefix+email+":count").Int()
	if err != nil {
		return l.maxAttempts // no record or Redis error — full attempts
	}
	remaining := l.maxAttempts - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
