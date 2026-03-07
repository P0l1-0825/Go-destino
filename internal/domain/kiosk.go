package domain

import "time"

type KioskStatus string

const (
	KioskOnline      KioskStatus = "online"
	KioskOffline     KioskStatus = "offline"
	KioskMaintenance KioskStatus = "maintenance"
)

// Kiosk represents a physical kiosk terminal managed by a tenant.
type Kiosk struct {
	ID         string      `json:"id" db:"id"`
	TenantID   string      `json:"tenant_id" db:"tenant_id"`
	Name       string      `json:"name" db:"name"`
	Location   string      `json:"location" db:"location"`
	AirportID  string      `json:"airport_id,omitempty" db:"airport_id"`
	TerminalID string      `json:"terminal_id,omitempty" db:"terminal_id"`
	Status     KioskStatus `json:"status" db:"status"`
	Config     KioskConfig `json:"config"`
	LastHeartbeat time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

type KioskConfig struct {
	Languages     []string `json:"languages"`
	DefaultLang   string   `json:"default_lang"`
	PrinterType   string   `json:"printer_type"`
	AcceptsNFC    bool     `json:"accepts_nfc"`
	AcceptsQR     bool     `json:"accepts_qr"`
	AcceptsCash   bool     `json:"accepts_cash"`
	IdleTimeoutSec int     `json:"idle_timeout_sec"`
}

type RegisterKioskRequest struct {
	Name       string `json:"name"`
	Location   string `json:"location"`
	AirportID  string `json:"airport_id"`
	TerminalID string `json:"terminal_id"`
}

type KioskHeartbeat struct {
	KioskID     string  `json:"kiosk_id"`
	Status      string  `json:"status"`
	PaperLevel  int     `json:"paper_level"`  // 0-100%
	Temperature float64 `json:"temperature"`
	Uptime      int64   `json:"uptime_seconds"`
}

// TransportCard represents a reloadable passenger card.
type TransportCard struct {
	ID         string    `json:"id" db:"id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	CardNumber string    `json:"card_number" db:"card_number"`
	BalanceCents int64   `json:"balance_cents" db:"balance_cents"`
	Currency   string    `json:"currency" db:"currency"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type RechargeCardRequest struct {
	CardNumber    string `json:"card_number"`
	AmountCents   int64  `json:"amount_cents"`
	PaymentMethod string `json:"payment_method"`
}

type RechargeCardResponse struct {
	Card    TransportCard `json:"card"`
	Payment Payment       `json:"payment"`
}
