package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

// PaymentService orchestrates the full payment lifecycle:
// charge → confirm → receipt → refund with notifications at each step.
// paymentRepoIface defines the subset of PaymentRepository methods used by PaymentService.
type paymentRepoIface interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Payment, error)
	GetByBookingIDTenant(ctx context.Context, bookingID, tenantID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id, tenantID string, status domain.PaymentStatus) error
	MarkFailed(ctx context.Context, id, tenantID, reason string) error
	Refund(ctx context.Context, originalID, tenantID string, refundPayment *domain.Payment) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.Payment, error)
}

// auditLogger defines the subset of AuditService methods used by PaymentService.
type auditLogger interface {
	Log(ctx context.Context, tenantID, userID, action, resource, resourceID, details, ip, ua string)
}

// notifier defines the subset of NotificationService methods used by PaymentService.
type notifier interface {
	SendPaymentReceipt(ctx context.Context, tenantID, userID string, payment *domain.Payment, bookingNumber, lang string)
	SendRefundNotification(ctx context.Context, tenantID, userID string, amountCents int64, currency, reference, reason, lang string)
}

type PaymentService struct {
	paymentRepo paymentRepoIface
	notifSvc    notifier
	auditSvc    auditLogger
}

func NewPaymentService(
	paymentRepo paymentRepoIface,
	notifSvc notifier,
	auditSvc auditLogger,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		notifSvc:    notifSvc,
		auditSvc:    auditSvc,
	}
}

// ProcessPaymentRequest describes a payment to process.
type ProcessPaymentRequest struct {
	TenantID    string
	UserID      string
	BookingID   string
	KioskID     string
	Method      domain.PaymentMethod
	AmountCents int64
	Currency    string
	Lang        string
}

// ProcessPayment creates a pending payment, charges via gateway, and sends receipt.
func (s *PaymentService) ProcessPayment(ctx context.Context, req ProcessPaymentRequest) (*domain.Payment, error) {
	ref, err := generatePaymentReference()
	if err != nil {
		return nil, fmt.Errorf("generating reference: %w", err)
	}

	payment := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		BookingID:   req.BookingID,
		KioskID:     req.KioskID,
		UserID:      req.UserID,
		Method:      req.Method,
		Status:      domain.PaymentPending,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Reference:   ref,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	// Process via payment gateway
	gatewayErr := s.chargeGateway(ctx, payment)
	if gatewayErr != nil {
		_ = s.paymentRepo.MarkFailed(ctx, payment.ID, req.TenantID, gatewayErr.Error())
		payment.Status = domain.PaymentFailed
		payment.FailureReason = gatewayErr.Error()

		// Audit failure
		go s.auditSvc.Log(ctx, req.TenantID, req.UserID, "payment.failed", "payment", payment.ID, gatewayErr.Error(), "", "")

		return payment, fmt.Errorf("payment failed: %w", gatewayErr)
	}

	// Mark completed
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, req.TenantID, domain.PaymentCompleted); err != nil {
		return nil, fmt.Errorf("completing payment: %w", err)
	}
	payment.Status = domain.PaymentCompleted

	// Audit success
	go s.auditSvc.Log(ctx, req.TenantID, req.UserID, "payment.completed", "payment", payment.ID,
		fmt.Sprintf("amount=%d currency=%s method=%s", req.AmountCents, req.Currency, req.Method), "", "")

	// Send payment receipt email (fire-and-forget)
	go func() {
		bookingNumber := ""
		if req.BookingID != "" {
			bookingNumber = req.BookingID // Simplified — in production, resolve booking number
		}
		s.notifSvc.SendPaymentReceipt(context.Background(), req.TenantID, req.UserID, payment, bookingNumber, req.Lang)
	}()

	return payment, nil
}

// RefundPayment processes a full refund and sends notifications.
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID, tenantID, userID, reason, lang string) (*domain.Payment, error) {
	original, err := s.paymentRepo.GetByIDTenant(ctx, paymentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if original.Status != domain.PaymentCompleted {
		return nil, fmt.Errorf("can only refund completed payments, current status: %s", original.Status)
	}

	ref, _ := generatePaymentReference()
	refund := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		BookingID:   original.BookingID,
		KioskID:     original.KioskID,
		UserID:      original.UserID,
		Method:      original.Method,
		AmountCents: original.AmountCents,
		Currency:    original.Currency,
		Reference:   "REF-" + ref,
	}

	// Process gateway refund
	if err := s.refundGateway(ctx, original); err != nil {
		return nil, fmt.Errorf("gateway refund failed: %w", err)
	}

	// Persist refund in DB
	if err := s.paymentRepo.Refund(ctx, paymentID, tenantID, refund); err != nil {
		return nil, fmt.Errorf("processing refund: %w", err)
	}

	// Audit
	go s.auditSvc.Log(ctx, tenantID, userID, "payment.refunded", "payment", paymentID,
		fmt.Sprintf("refund_id=%s amount=%d reason=%s", refund.ID, refund.AmountCents, reason), "", "")

	// Send refund notification (all channels)
	go s.notifSvc.SendRefundNotification(
		context.Background(), tenantID, original.UserID,
		original.AmountCents, original.Currency, refund.Reference, reason, lang,
	)

	return refund, nil
}

// GetPayment retrieves a single payment (tenant-scoped).
func (s *PaymentService) GetPayment(ctx context.Context, id, tenantID string) (*domain.Payment, error) {
	return s.paymentRepo.GetByIDTenant(ctx, id, tenantID)
}

// GetPaymentByBooking retrieves the latest payment for a booking (tenant-scoped).
func (s *PaymentService) GetPaymentByBooking(ctx context.Context, bookingID, tenantID string) (*domain.Payment, error) {
	return s.paymentRepo.GetByBookingIDTenant(ctx, bookingID, tenantID)
}

// ListPayments lists payments for a tenant.
func (s *PaymentService) ListPayments(ctx context.Context, tenantID string, limit, offset int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.paymentRepo.ListByTenant(ctx, tenantID, limit, offset)
}

// --- Gateway integration (pluggable) ---

func (s *PaymentService) chargeGateway(ctx context.Context, payment *domain.Payment) error {
	// In production, this routes to:
	//   - Stripe/Conekta for card payments
	//   - MercadoPago for QR/OXXO
	//   - Cash register for cash (always succeeds)
	switch payment.Method {
	case domain.PaymentCash:
		log.Printf("[GATEWAY] Cash payment %s — auto-approved: %d %s", payment.ID, payment.AmountCents, payment.Currency)
		return nil
	case domain.PaymentCard:
		log.Printf("[GATEWAY] Card charge %s → ref: %s amount: %d %s", payment.ID, payment.Reference, payment.AmountCents, payment.Currency)
		// Stripe/Conekta integration placeholder
		return nil
	case domain.PaymentQR:
		log.Printf("[GATEWAY] QR payment %s → ref: %s amount: %d %s", payment.ID, payment.Reference, payment.AmountCents, payment.Currency)
		// MercadoPago/CoDi integration placeholder
		return nil
	default:
		return fmt.Errorf("unsupported payment method: %s", payment.Method)
	}
}

func (s *PaymentService) refundGateway(ctx context.Context, original *domain.Payment) error {
	switch original.Method {
	case domain.PaymentCash:
		log.Printf("[GATEWAY] Cash refund %s — manual process", original.ID)
		return nil
	case domain.PaymentCard:
		log.Printf("[GATEWAY] Card refund %s → original ref: %s", original.ID, original.Reference)
		// Stripe/Conekta refund placeholder
		return nil
	case domain.PaymentQR:
		log.Printf("[GATEWAY] QR refund %s → original ref: %s", original.ID, original.Reference)
		return nil
	default:
		return nil
	}
}

func generatePaymentReference() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "PAY-" + hex.EncodeToString(b), nil
}
