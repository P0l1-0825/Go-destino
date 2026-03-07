package domain

import "time"

type TicketStatus string

const (
	TicketActive   TicketStatus = "active"
	TicketUsed     TicketStatus = "used"
	TicketExpired  TicketStatus = "expired"
	TicketCanceled TicketStatus = "canceled"
)

// Ticket represents a purchased transport ticket.
type Ticket struct {
	ID          string       `json:"id" db:"id"`
	TenantID    string       `json:"tenant_id" db:"tenant_id"`
	RouteID     string       `json:"route_id" db:"route_id"`
	KioskID     string       `json:"kiosk_id" db:"kiosk_id"`
	PaymentID   string       `json:"payment_id" db:"payment_id"`
	QRCode      string       `json:"qr_code" db:"qr_code"`
	Status      TicketStatus `json:"status" db:"status"`
	PriceCents  int64        `json:"price_cents" db:"price_cents"`
	Currency    string       `json:"currency" db:"currency"`
	PassengerID string       `json:"passenger_id,omitempty" db:"passenger_id"`
	ValidFrom   time.Time    `json:"valid_from" db:"valid_from"`
	ValidUntil  time.Time    `json:"valid_until" db:"valid_until"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
}

type PurchaseTicketRequest struct {
	RouteID       string `json:"route_id"`
	PaymentMethod string `json:"payment_method"` // cash, card, qr
	Quantity      int    `json:"quantity"`
}

type PurchaseTicketResponse struct {
	Tickets []Ticket `json:"tickets"`
	Payment Payment  `json:"payment"`
}
