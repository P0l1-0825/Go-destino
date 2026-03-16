package security

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordResetStore_CreateToken(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	token, err := s.CreateToken("user-1", "tenant-1", "test@example.com", 1*time.Hour)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("CreateToken returned empty token")
	}

	if len(token) != 64 { // 32 bytes hex encoded = 64 chars
		t.Errorf("token length = %d, want 64", len(token))
	}
}

func TestPasswordResetStore_ValidateToken_HappyPath(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	token, _ := s.CreateToken("user-1", "tenant-1", "test@example.com", 1*time.Hour)

	userID, err := s.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if userID != "user-1" {
		t.Errorf("ValidateToken returned userID = %q, want %q", userID, "user-1")
	}
}

func TestPasswordResetStore_ValidateToken_InvalidToken(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	_, err := s.ValidateToken("nonexistent-token")
	if err == nil {
		t.Fatal("ValidateToken should fail for nonexistent token")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error should contain 'invalid', got: %v", err)
	}
}

func TestPasswordResetStore_ValidateToken_AlreadyUsed(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	token, _ := s.CreateToken("user-1", "tenant-1", "test@example.com", 1*time.Hour)

	// First use — should succeed
	_, err := s.ValidateToken(token)
	if err != nil {
		t.Fatalf("first ValidateToken should succeed, got: %v", err)
	}

	// Second use — should fail
	_, err = s.ValidateToken(token)
	if err == nil {
		t.Fatal("second ValidateToken should fail (already used)")
	}
	if !strings.Contains(err.Error(), "already used") {
		t.Errorf("error should contain 'already used', got: %v", err)
	}
}

func TestPasswordResetStore_ValidateToken_Expired(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	token, _ := s.CreateToken("user-1", "tenant-1", "test@example.com", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, err := s.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken should fail for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error should contain 'expired', got: %v", err)
	}
}

func TestPasswordResetStore_MultipleTokens_Independent(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	token1, _ := s.CreateToken("user-1", "tenant-1", "a@example.com", 1*time.Hour)
	token2, _ := s.CreateToken("user-2", "tenant-1", "b@example.com", 1*time.Hour)

	if token1 == token2 {
		t.Error("two tokens should be unique")
	}

	uid1, _ := s.ValidateToken(token1)
	uid2, _ := s.ValidateToken(token2)

	if uid1 != "user-1" || uid2 != "user-2" {
		t.Errorf("tokens should map to correct users, got %q and %q", uid1, uid2)
	}
}

func TestPasswordResetStore_Cleanup(t *testing.T) {
	s := &PasswordResetStore{
		tokens: make(map[string]*resetToken),
	}

	s.CreateToken("user-1", "tenant-1", "a@example.com", 1*time.Millisecond)
	s.CreateToken("user-2", "tenant-1", "b@example.com", 1*time.Hour)

	time.Sleep(5 * time.Millisecond)
	s.cleanup()

	s.mu.Lock()
	count := len(s.tokens)
	s.mu.Unlock()

	if count != 1 {
		t.Errorf("after cleanup, expected 1 token remaining, got %d", count)
	}
}
