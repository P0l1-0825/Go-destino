package security

import (
	"fmt"
	"sync"
	"time"
)

// LoginLimiter tracks failed login attempts per email and enforces
// rate limiting with progressive lockout.
type LoginLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempts

	maxAttempts int
	window      time.Duration
	lockout     time.Duration
}

type loginAttempts struct {
	count    int
	firstAt  time.Time
	lockedAt *time.Time
}

// NewLoginLimiter creates a limiter: maxAttempts failures within window
// triggers a lockout duration.
func NewLoginLimiter(maxAttempts int, window, lockout time.Duration) *LoginLimiter {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: maxAttempts,
		window:      window,
		lockout:     lockout,
	}

	// Cleanup stale entries every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			l.cleanup()
		}
	}()

	return l
}

// Check returns an error if the email is currently locked out.
func (l *LoginLimiter) Check(email string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		return nil
	}

	// Check if locked out
	if a.lockedAt != nil {
		remaining := l.lockout - time.Since(*a.lockedAt)
		if remaining > 0 {
			mins := int(remaining.Minutes()) + 1
			return fmt.Errorf("account temporarily locked, try again in %d minutes", mins)
		}
		// Lockout expired, reset
		delete(l.attempts, email)
		return nil
	}

	// Check if window expired
	if time.Since(a.firstAt) > l.window {
		delete(l.attempts, email)
		return nil
	}

	return nil
}

// RecordFailure records a failed login attempt.
// Returns an error if the account is now locked.
func (l *LoginLimiter) RecordFailure(email string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		a = &loginAttempts{firstAt: time.Now()}
		l.attempts[email] = a
	}

	// Reset if window expired
	if time.Since(a.firstAt) > l.window {
		a.count = 0
		a.firstAt = time.Now()
		a.lockedAt = nil
	}

	a.count++

	if a.count >= l.maxAttempts {
		now := time.Now()
		a.lockedAt = &now
		mins := int(l.lockout.Minutes())
		return fmt.Errorf("too many failed attempts, account locked for %d minutes", mins)
	}

	remaining := l.maxAttempts - a.count
	return fmt.Errorf("invalid credentials (%d attempts remaining)", remaining)
}

// RecordSuccess clears the failure count for an email.
func (l *LoginLimiter) RecordSuccess(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, email)
}

// RemainingAttempts returns how many attempts are left for an email.
func (l *LoginLimiter) RemainingAttempts(email string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		return l.maxAttempts
	}

	if time.Since(a.firstAt) > l.window {
		return l.maxAttempts
	}

	return l.maxAttempts - a.count
}

func (l *LoginLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-l.window - l.lockout)
	for email, a := range l.attempts {
		if a.firstAt.Before(cutoff) {
			delete(l.attempts, email)
		}
	}
}
