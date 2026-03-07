package service

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/P0l1-0825/Go-destino/internal/config"
)

// SMTPService handles email delivery via SMTP.
type SMTPService struct {
	cfg     config.SMTPConfig
	enabled bool
}

func NewSMTPService(cfg config.SMTPConfig) *SMTPService {
	return &SMTPService{
		cfg:     cfg,
		enabled: cfg.Enabled && cfg.Host != "",
	}
}

// SendEmail sends an HTML email to the given recipient.
func (s *SMTPService) SendEmail(to, subject, htmlBody string) error {
	if !s.enabled {
		log.Printf("[EMAIL] (dry-run) → %s | Subject: %s", to, subject)
		return nil
	}

	from := s.cfg.From
	// Extract email address from "Name <email>" format
	fromAddr := from
	if idx := strings.Index(from, "<"); idx != -1 {
		fromAddr = strings.Trim(from[idx:], "<>")
	}

	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, fromAddr, []string{to}, []byte(msg.String())); err != nil {
		log.Printf("[EMAIL] failed to send to %s: %v", to, err)
		return fmt.Errorf("sending email: %w", err)
	}

	log.Printf("[EMAIL] sent to %s | Subject: %s", to, subject)
	return nil
}

// IsEnabled returns true if SMTP is configured and enabled.
func (s *SMTPService) IsEnabled() bool {
	return s.enabled
}
