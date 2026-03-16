package validator

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
		errMsg  string
	}{
		{"user@example.com", false, ""},
		{"admin@godestino.com", false, ""},
		{"", true, "required"},
		{"   ", true, "required"},
		{"not-an-email", true, "invalid email"},
		{"@missing-local.com", true, "invalid email"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error should contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		field   string
		name    string
		wantErr bool
	}{
		{"value", "field", false},
		{"", "field", true},
		{"   ", "field", true},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			err := ValidateRequired(tt.field, tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired(%q, %q) error = %v, wantErr %v", tt.field, tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		min     int
		wantErr bool
	}{
		{"meets min", "longvalue", 5, false},
		{"exactly min", "12345", 5, false},
		{"below min", "abc", 5, true},
		{"empty", "", 1, true},
		{"spaces trimmed", "   ab   ", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMinLength(tt.field, "test", tt.min)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMinLength(%q, %d) error = %v, wantErr %v", tt.field, tt.min, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePositive(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr bool
	}{
		{"positive", 10, false},
		{"one", 1, false},
		{"zero", 0, true},
		{"negative", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositive(tt.value, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePositive(%d) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
