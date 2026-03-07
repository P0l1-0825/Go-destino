package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type VoucherRepository struct {
	db *sql.DB
}

func NewVoucherRepository(db *sql.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

func (r *VoucherRepository) Create(ctx context.Context, v *domain.Voucher) error {
	query := `INSERT INTO vouchers (id, tenant_id, code, booking_id, amount_cents, currency, status, qr_code_url, created_by, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())`
	_, err := r.db.ExecContext(ctx, query, v.ID, v.TenantID, v.Code, v.BookingID, v.AmountCents, v.Currency, v.Status, v.QRCodeURL, v.CreatedBy, v.ExpiresAt)
	return err
}

func (r *VoucherRepository) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	v := &domain.Voucher{}
	query := `SELECT id, tenant_id, code, booking_id, amount_cents, currency, status, qr_code_url, created_by, redeemed_by, expires_at, redeemed_at, created_at
		FROM vouchers WHERE code=$1`
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&v.ID, &v.TenantID, &v.Code, &v.BookingID, &v.AmountCents, &v.Currency, &v.Status, &v.QRCodeURL,
		&v.CreatedBy, &v.RedeemedBy, &v.ExpiresAt, &v.RedeemedAt, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VoucherRepository) Redeem(ctx context.Context, id, redeemedBy string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vouchers SET status='redeemed', redeemed_by=$1, redeemed_at=NOW() WHERE id=$2`, redeemedBy, id)
	return err
}
