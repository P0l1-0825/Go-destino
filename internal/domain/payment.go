package domain

import "time"

type PaymentMethod string

const (
	PaymentCash PaymentMethod = "cash"
	PaymentCard PaymentMethod = "card"
	PaymentQR   PaymentMethod = "qr"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

// Payment represents a transaction for ticket purchase or card recharge.
type Payment struct {
	ID          string        `json:"id" db:"id"`
	TenantID    string        `json:"tenant_id" db:"tenant_id"`
	KioskID     string        `json:"kiosk_id" db:"kiosk_id"`
	Method      PaymentMethod `json:"method" db:"method"`
	Status      PaymentStatus `json:"status" db:"status"`
	AmountCents int64         `json:"amount_cents" db:"amount_cents"`
	Currency    string        `json:"currency" db:"currency"`
	Reference   string        `json:"reference,omitempty" db:"reference"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}
