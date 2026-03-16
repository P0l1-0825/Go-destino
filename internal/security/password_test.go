package security

import (
	"strings"
	"testing"
)

func TestPasswordPolicy_Validate(t *testing.T) {
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{"valid password", "StrongPass1", false, ""},
		{"valid with special chars", "MyP@ss1234", false, ""},
		{"too short", "Ab1", true, "at least 8 characters"},
		{"too long", strings.Repeat("A", 129), true, "must not exceed 128"},
		{"missing uppercase", "lowercase1", true, "uppercase letter"},
		{"missing lowercase", "UPPERCASE1", true, "lowercase letter"},
		{"missing digit", "NoDigitsHere", true, "digit"},
		{"common password", "Password1", true, "too common"},
		{"common password 2", "Admin123", true, "too common"},
		{"empty password", "", true, "at least 8 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error should contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestPasswordPolicy_CustomPolicy(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"meets all requirements", "MyP@ssw0rd123", false},
		{"missing special", "MyPassw0rd123", true},
		{"too short for custom", "MyP@ss1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestDefaultPasswordPolicy(t *testing.T) {
	p := DefaultPasswordPolicy()

	if p.MinLength != 8 {
		t.Errorf("MinLength = %d, want 8", p.MinLength)
	}
	if !p.RequireUpper {
		t.Error("RequireUpper should be true")
	}
	if !p.RequireLower {
		t.Error("RequireLower should be true")
	}
	if !p.RequireDigit {
		t.Error("RequireDigit should be true")
	}
	if p.RequireSpecial {
		t.Error("RequireSpecial should be false by default")
	}
}

func TestIsCommonPassword(t *testing.T) {
	tests := []struct {
		password string
		want     bool
	}{
		{"password", true},
		{"12345678", true},
		{"qwerty12", true},
		{"admin123", true},
		{"xk3m9z2q", false},
		{"uniquepassword", false},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			got := isCommonPassword(tt.password)
			if got != tt.want {
				t.Errorf("isCommonPassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}
