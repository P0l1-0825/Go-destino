package security

import (
	"strings"
	"testing"
	"time"
)

func TestLoginLimiter_Check_NoAttempts(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 5,
		window:      15 * time.Minute,
		lockout:     30 * time.Minute,
	}

	if err := l.Check("new@example.com"); err != nil {
		t.Errorf("Check on new email should not error, got: %v", err)
	}
}

func TestLoginLimiter_RecordFailure_CountsDown(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 5,
		window:      15 * time.Minute,
		lockout:     30 * time.Minute,
	}

	email := "test@example.com"

	for i := 1; i <= 4; i++ {
		err := l.RecordFailure(email)
		if err == nil {
			t.Fatalf("RecordFailure #%d should return error with remaining attempts", i)
		}
		expected := 5 - i
		if !strings.Contains(err.Error(), "attempts remaining") {
			t.Errorf("RecordFailure #%d error should contain 'attempts remaining', got: %v", i, err)
		}
		remaining := l.RemainingAttempts(email)
		if remaining != expected {
			t.Errorf("after %d failures, RemainingAttempts = %d, want %d", i, remaining, expected)
		}
	}
}

func TestLoginLimiter_RecordFailure_LocksAfterMax(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 3,
		window:      15 * time.Minute,
		lockout:     30 * time.Minute,
	}

	email := "brute@example.com"

	// 2 failures — not locked yet
	l.RecordFailure(email)
	l.RecordFailure(email)

	if err := l.Check(email); err != nil {
		t.Errorf("should not be locked after 2 failures, got: %v", err)
	}

	// 3rd failure — should lock
	err := l.RecordFailure(email)
	if err == nil {
		t.Fatal("3rd failure should return lockout error")
	}
	if !strings.Contains(err.Error(), "locked") {
		t.Errorf("lockout error should contain 'locked', got: %v", err)
	}

	// Check should now fail
	err = l.Check(email)
	if err == nil {
		t.Fatal("Check should return error when locked")
	}
	if !strings.Contains(err.Error(), "locked") {
		t.Errorf("Check error should contain 'locked', got: %v", err)
	}
}

func TestLoginLimiter_RecordSuccess_ClearsAttempts(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 5,
		window:      15 * time.Minute,
		lockout:     30 * time.Minute,
	}

	email := "success@example.com"
	l.RecordFailure(email)
	l.RecordFailure(email)

	if l.RemainingAttempts(email) == 5 {
		t.Error("should have fewer remaining attempts after failures")
	}

	l.RecordSuccess(email)

	if l.RemainingAttempts(email) != 5 {
		t.Errorf("after RecordSuccess, RemainingAttempts should be max (5), got %d", l.RemainingAttempts(email))
	}
}

func TestLoginLimiter_RemainingAttempts_NewEmail(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 5,
		window:      15 * time.Minute,
		lockout:     30 * time.Minute,
	}

	if got := l.RemainingAttempts("unknown@example.com"); got != 5 {
		t.Errorf("RemainingAttempts for new email = %d, want 5", got)
	}
}

func TestLoginLimiter_WindowExpiry_ResetsCount(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 5,
		window:      1 * time.Millisecond, // very short window
		lockout:     30 * time.Minute,
	}

	email := "window@example.com"
	l.RecordFailure(email)
	l.RecordFailure(email)

	time.Sleep(5 * time.Millisecond) // wait for window to expire

	// After window expires, Check should reset and allow
	if err := l.Check(email); err != nil {
		t.Errorf("after window expiry, Check should not error, got: %v", err)
	}

	if got := l.RemainingAttempts(email); got != 5 {
		t.Errorf("after window expiry, RemainingAttempts = %d, want 5", got)
	}
}

func TestLoginLimiter_Cleanup(t *testing.T) {
	l := &LoginLimiter{
		attempts:    make(map[string]*loginAttempts),
		maxAttempts: 5,
		window:      1 * time.Millisecond,
		lockout:     1 * time.Millisecond,
	}

	l.RecordFailure("stale@example.com")
	time.Sleep(5 * time.Millisecond)

	l.cleanup()

	l.mu.Lock()
	count := len(l.attempts)
	l.mu.Unlock()

	if count != 0 {
		t.Errorf("cleanup should remove stale entries, got %d remaining", count)
	}
}
