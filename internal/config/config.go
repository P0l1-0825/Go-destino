package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	SMTP        SMTPConfig
	Twilio      TwilioConfig
	CORSOrigins []string
}

// SMTPConfig holds email delivery settings.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string // "GoDestino <noreply@godestino.com>"
	Enabled  bool
}

// TwilioConfig holds SMS and WhatsApp delivery settings.
type TwilioConfig struct {
	AccountSID     string
	AuthToken      string
	SMSFrom        string // e.g. "+15551234567"
	WhatsAppFrom   string // e.g. "whatsapp:+14155238886"
	Enabled        bool
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "destino"),
			Password: getEnv("DB_PASSWORD", "destino"),
			Name:     getEnv("DB_NAME", "destino"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			ExpireHour: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvInt("SMTP_PORT", 587),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "GoDestino <noreply@godestino.com>"),
			Enabled:  getEnv("SMTP_ENABLED", "false") == "true",
		},
		Twilio: TwilioConfig{
			AccountSID:   getEnv("TWILIO_ACCOUNT_SID", ""),
			AuthToken:    getEnv("TWILIO_AUTH_TOKEN", ""),
			SMSFrom:      getEnv("TWILIO_SMS_FROM", ""),
			WhatsAppFrom: getEnv("TWILIO_WHATSAPP_FROM", ""),
			Enabled:      getEnv("TWILIO_ENABLED", "false") == "true",
		},
		CORSOrigins: getEnvList("CORS_ORIGINS"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvList(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	var result []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
