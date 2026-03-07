package validator

import (
	"fmt"
	"net/mail"
	"strings"
)

// ValidateEmail checks if the email format is valid.
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateRequired checks that a string field is not empty.
func ValidateRequired(field, name string) error {
	if strings.TrimSpace(field) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ValidateMinLength checks minimum string length.
func ValidateMinLength(field, name string, min int) error {
	if len(strings.TrimSpace(field)) < min {
		return fmt.Errorf("%s must be at least %d characters", name, min)
	}
	return nil
}

// ValidatePositive checks that an int64 is positive.
func ValidatePositive(value int64, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}
	return nil
}
