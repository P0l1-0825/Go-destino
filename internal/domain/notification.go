package domain

import "time"

type NotificationChannel string

const (
	ChannelPush     NotificationChannel = "push"
	ChannelSMS      NotificationChannel = "sms"
	ChannelEmail    NotificationChannel = "email"
	ChannelWhatsApp NotificationChannel = "whatsapp"
)

type NotificationStatus string

const (
	NotifPending   NotificationStatus = "pending"
	NotifSent      NotificationStatus = "sent"
	NotifDelivered NotificationStatus = "delivered"
	NotifFailed    NotificationStatus = "failed"
)

// Notification represents a message sent to a user.
type Notification struct {
	ID        string              `json:"id" db:"id"`
	TenantID  string              `json:"tenant_id" db:"tenant_id"`
	UserID    string              `json:"user_id" db:"user_id"`
	Channel   NotificationChannel `json:"channel" db:"channel"`
	Status    NotificationStatus  `json:"status" db:"status"`
	Title     string              `json:"title" db:"title"`
	Body      string              `json:"body" db:"body"`
	Data      map[string]string   `json:"data,omitempty"`
	BookingID string              `json:"booking_id,omitempty" db:"booking_id"`
	CreatedAt time.Time           `json:"created_at" db:"created_at"`
	SentAt    *time.Time          `json:"sent_at,omitempty" db:"sent_at"`
}

// NotificationPreference stores user notification preferences.
type NotificationPreference struct {
	UserID          string `json:"user_id" db:"user_id"`
	PushEnabled     bool   `json:"push_enabled" db:"push_enabled"`
	SMSEnabled      bool   `json:"sms_enabled" db:"sms_enabled"`
	EmailEnabled    bool   `json:"email_enabled" db:"email_enabled"`
	WhatsAppEnabled bool   `json:"whatsapp_enabled" db:"whatsapp_enabled"`
	Lang            string `json:"lang" db:"lang"`
}

type SendNotificationRequest struct {
	UserID    string              `json:"user_id"`
	Channel   NotificationChannel `json:"channel"`
	Title     string              `json:"title"`
	Body      string              `json:"body"`
	Data      map[string]string   `json:"data,omitempty"`
	BookingID string              `json:"booking_id,omitempty"`
}

// NotificationTemplate for i18n multi-language messages.
type NotificationTemplate struct {
	ID       string `json:"id" db:"id"`
	Key      string `json:"key" db:"key"`      // e.g. "booking.confirmed"
	Lang     string `json:"lang" db:"lang"`     // es, en, pt
	Channel  string `json:"channel" db:"channel"`
	Title    string `json:"title" db:"title"`
	Body     string `json:"body" db:"body"`     // supports {{.BookingNumber}} placeholders
}
