package config

import (
	"os"
	"testing"
)

func TestParseDatabaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(t *testing.T, cfg DatabaseConfig)
	}{
		{
			name: "full URL with sslmode",
			url:  "postgresql://user:pass@host.example.com:5432/mydb?sslmode=require",
			check: func(t *testing.T, cfg DatabaseConfig) {
				if cfg.Host != "host.example.com" {
					t.Errorf("Host = %q, want host.example.com", cfg.Host)
				}
				if cfg.Port != "5432" {
					t.Errorf("Port = %q, want 5432", cfg.Port)
				}
				if cfg.User != "user" {
					t.Errorf("User = %q, want user", cfg.User)
				}
				if cfg.Password != "pass" {
					t.Errorf("Password = %q, want pass", cfg.Password)
				}
				if cfg.Name != "mydb" {
					t.Errorf("Name = %q, want mydb", cfg.Name)
				}
				if cfg.SSLMode != "require" {
					t.Errorf("SSLMode = %q, want require", cfg.SSLMode)
				}
			},
		},
		{
			name: "URL without sslmode defaults to require",
			url:  "postgresql://user:pass@host.example.com:5432/mydb",
			check: func(t *testing.T, cfg DatabaseConfig) {
				if cfg.SSLMode != "require" {
					t.Errorf("SSLMode = %q, want require (default)", cfg.SSLMode)
				}
			},
		},
		{
			name: "URL without port defaults to 5432",
			url:  "postgresql://user:pass@host.example.com/mydb",
			check: func(t *testing.T, cfg DatabaseConfig) {
				if cfg.Port != "5432" {
					t.Errorf("Port = %q, want 5432 (default)", cfg.Port)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseDatabaseURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDatabaseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestParseRedisURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		check func(t *testing.T, cfg RedisConfig)
	}{
		{
			name: "standard redis URL",
			url:  "redis://default:mypassword@redis.example.com:6379",
			check: func(t *testing.T, cfg RedisConfig) {
				if cfg.Host != "redis.example.com" {
					t.Errorf("Host = %q, want redis.example.com", cfg.Host)
				}
				if cfg.Port != "6379" {
					t.Errorf("Port = %q, want 6379", cfg.Port)
				}
				if cfg.Password != "mypassword" {
					t.Errorf("Password = %q, want mypassword", cfg.Password)
				}
				if cfg.DB != 0 {
					t.Errorf("DB = %d, want 0", cfg.DB)
				}
			},
		},
		{
			name: "redis URL with password as username",
			url:  "redis://:secretpass@redis.example.com:6379",
			check: func(t *testing.T, cfg RedisConfig) {
				if cfg.Password != "secretpass" {
					t.Errorf("Password = %q, want secretpass", cfg.Password)
				}
			},
		},
		{
			name: "redis URL with db number",
			url:  "redis://default:pass@redis.example.com:6379/3",
			check: func(t *testing.T, cfg RedisConfig) {
				if cfg.DB != 3 {
					t.Errorf("DB = %d, want 3", cfg.DB)
				}
			},
		},
		{
			name: "redis URL without port defaults to 6379",
			url:  "redis://default:pass@redis.example.com",
			check: func(t *testing.T, cfg RedisConfig) {
				if cfg.Port != "6379" {
					t.Errorf("Port = %q, want 6379 (default)", cfg.Port)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseRedisURL(tt.url)
			if err != nil {
				t.Fatalf("parseRedisURL() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Errorf("DSN() = %q, want %q", dsn, expected)
	}
}

func TestGetEnvList(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int // expected count
	}{
		{"empty", "", 0},
		{"single", "http://localhost:3000", 1},
		{"multiple", "http://localhost:3000, http://localhost:5173, https://app.example.com", 3},
		{"with spaces", "  a , b , c  ", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_ENV_LIST", tt.value)
			defer os.Unsetenv("TEST_ENV_LIST")

			result := getEnvList("TEST_ENV_LIST")
			if len(result) != tt.want {
				t.Errorf("getEnvList() returned %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	// Test with valid int
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")
	if got := getEnvInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvInt() = %d, want 42", got)
	}

	// Test with invalid int — should return fallback
	os.Setenv("TEST_INT_INVALID", "not-a-number")
	defer os.Unsetenv("TEST_INT_INVALID")
	if got := getEnvInt("TEST_INT_INVALID", 99); got != 99 {
		t.Errorf("getEnvInt() with invalid value = %d, want fallback 99", got)
	}

	// Test with missing env — should return fallback
	if got := getEnvInt("TEST_INT_MISSING", 7); got != 7 {
		t.Errorf("getEnvInt() with missing key = %d, want fallback 7", got)
	}
}
