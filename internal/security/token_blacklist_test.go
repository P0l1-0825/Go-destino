package security

import (
	"testing"
	"time"
)

func TestTokenBlacklist_Revoke_And_IsRevoked(t *testing.T) {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	tests := []struct {
		name      string
		jti       string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "revoked token is detected",
			jti:       "jti-001",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired revoked token is not detected",
			jti:       "jti-expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl.Revoke(tt.jti, tt.expiresAt)
			got := bl.IsRevoked(tt.jti)
			if got != tt.want {
				t.Errorf("IsRevoked(%q) = %v, want %v", tt.jti, got, tt.want)
			}
		})
	}
}

func TestTokenBlacklist_IsRevoked_Unknown(t *testing.T) {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	if bl.IsRevoked("nonexistent-jti") {
		t.Error("IsRevoked returned true for unknown JTI")
	}
}

func TestTokenBlacklist_Cleanup(t *testing.T) {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	bl.Revoke("active", time.Now().Add(1*time.Hour))
	bl.Revoke("expired", time.Now().Add(-1*time.Hour))

	bl.cleanup()

	if !bl.IsRevoked("active") {
		t.Error("active token should still be revoked after cleanup")
	}

	bl.mu.RLock()
	_, exists := bl.tokens["expired"]
	bl.mu.RUnlock()
	if exists {
		t.Error("expired token should be removed after cleanup")
	}
}

func TestTokenBlacklist_ConcurrentAccess(t *testing.T) {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			jti := "jti-concurrent"
			bl.Revoke(jti, time.Now().Add(1*time.Hour))
			bl.IsRevoked(jti)
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
