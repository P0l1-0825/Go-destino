package security

import (
	"fmt"
	"strings"
	"unicode"
)

// PasswordPolicy defines the password complexity requirements.
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordPolicy returns the standard GoDestino password policy.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: false, // not required but recommended
	}
}

// Validate checks a password against the policy.
func (p PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters", p.MinLength)
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	var missing []string
	if p.RequireUpper && !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if p.RequireLower && !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if p.RequireDigit && !hasDigit {
		missing = append(missing, "digit")
	}
	if p.RequireSpecial && !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain at least one %s", strings.Join(missing, ", "))
	}

	// Check common weak passwords
	if isCommonPassword(strings.ToLower(password)) {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	return nil
}

func isCommonPassword(pw string) bool {
	common := []string{
		"password", "12345678", "123456789", "1234567890",
		"qwerty12", "qwerty123", "abcdefgh", "abcd1234",
		"password1", "password123", "iloveyou", "admin123",
		"welcome1", "letmein12", "changeme", "football",
		"baseball", "trustno1", "sunshine", "princess",
		"master12", "dragon12", "monkey12", "shadow12",
	}
	for _, c := range common {
		if pw == c {
			return true
		}
	}
	return false
}
