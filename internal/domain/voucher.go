package domain

import "time"

type VoucherStatus string

const (
	VoucherActive   VoucherStatus = "active"
	VoucherRedeemed VoucherStatus = "redeemed"
	VoucherExpired  VoucherStatus = "expired"
	VoucherVoided   VoucherStatus = "voided"
)

// Voucher represents a cash payment voucher generated at POS/kiosk.
type Voucher struct {
	ID          string        `json:"id" db:"id"`
	TenantID    string        `json:"tenant_id" db:"tenant_id"`
	Code        string        `json:"code" db:"code"`
	BookingID   string        `json:"booking_id,omitempty" db:"booking_id"`
	AmountCents int64         `json:"amount_cents" db:"amount_cents"`
	Currency    string        `json:"currency" db:"currency"`
	Status      VoucherStatus `json:"status" db:"status"`
	QRCodeURL   string        `json:"qr_code_url" db:"qr_code_url"`
	CreatedBy   string        `json:"created_by" db:"created_by"` // seller/kiosk ID
	RedeemedBy  string        `json:"redeemed_by,omitempty" db:"redeemed_by"`
	ExpiresAt   time.Time     `json:"expires_at" db:"expires_at"`
	RedeemedAt  *time.Time    `json:"redeemed_at,omitempty" db:"redeemed_at"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}

type CreateVoucherRequest struct {
	BookingID   string `json:"booking_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type RedeemVoucherRequest struct {
	Code           string `json:"code"`
	AmountReceived int64  `json:"amount_received"` // cash tendered
}

type RedeemVoucherResponse struct {
	Voucher Voucher `json:"voucher"`
	Payment Payment `json:"payment"`
	Change  int64   `json:"change_cents"` // amount_received - voucher amount
}
