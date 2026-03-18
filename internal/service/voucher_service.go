package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type voucherRepoIface interface {
	Create(ctx context.Context, v *domain.Voucher) error
	GetByID(ctx context.Context, id string) (*domain.Voucher, error)
	GetByCode(ctx context.Context, code string) (*domain.Voucher, error)
	GetByCodeTenant(ctx context.Context, code, tenantID string) (*domain.Voucher, error)
	Redeem(ctx context.Context, id, redeemedBy string) error
	List(ctx context.Context, tenantID string, limit, offset int) ([]domain.Voucher, error)
}

type paymentCreator interface {
	Create(ctx context.Context, p *domain.Payment) error
}

type VoucherService struct {
	voucherRepo voucherRepoIface
	paymentRepo paymentCreator
}

func NewVoucherService(voucherRepo voucherRepoIface, paymentRepo paymentCreator) *VoucherService {
	return &VoucherService{voucherRepo: voucherRepo, paymentRepo: paymentRepo}
}

func (s *VoucherService) Create(ctx context.Context, tenantID, createdBy string, req domain.CreateVoucherRequest) (*domain.Voucher, error) {
	code, err := generateVoucherCode()
	if err != nil {
		return nil, err
	}

	voucher := &domain.Voucher{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Code:        code,
		BookingID:   req.BookingID,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Status:      domain.VoucherActive,
		QRCodeURL:   fmt.Sprintf("https://api.godestino.com/voucher/qr/%s", code),
		CreatedBy:   createdBy,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	if err := s.voucherRepo.Create(ctx, voucher); err != nil {
		return nil, err
	}
	return voucher, nil
}

func (s *VoucherService) Redeem(ctx context.Context, tenantID, redeemedBy string, req domain.RedeemVoucherRequest) (*domain.RedeemVoucherResponse, error) {
	voucher, err := s.voucherRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("voucher not found")
	}

	if voucher.Status != domain.VoucherActive {
		return nil, fmt.Errorf("voucher is %s", voucher.Status)
	}

	if time.Now().After(voucher.ExpiresAt) {
		return nil, fmt.Errorf("voucher has expired")
	}

	if req.AmountReceived < voucher.AmountCents {
		return nil, fmt.Errorf("insufficient cash: need %d, received %d", voucher.AmountCents, req.AmountReceived)
	}

	// Create payment
	payment := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Method:      domain.PaymentCash,
		Status:      domain.PaymentCompleted,
		AmountCents: voucher.AmountCents,
		Currency:    voucher.Currency,
		Reference:   "VOUCHER:" + voucher.Code,
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// Redeem voucher
	if err := s.voucherRepo.Redeem(ctx, voucher.ID, redeemedBy); err != nil {
		return nil, err
	}

	voucher.Status = domain.VoucherRedeemed
	return &domain.RedeemVoucherResponse{
		Voucher: *voucher,
		Payment: *payment,
		Change:  req.AmountReceived - voucher.AmountCents,
	}, nil
}

func (s *VoucherService) List(ctx context.Context, tenantID string, limit, offset int) ([]domain.Voucher, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.voucherRepo.List(ctx, tenantID, limit, offset)
}

func (s *VoucherService) GetByID(ctx context.Context, id string) (*domain.Voucher, error) {
	return s.voucherRepo.GetByID(ctx, id)
}

func (s *VoucherService) GetByCode(ctx context.Context, code, tenantID string) (*domain.Voucher, error) {
	return s.voucherRepo.GetByCodeTenant(ctx, code, tenantID)
}

func generateVoucherCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "V-" + hex.EncodeToString(b), nil
}
