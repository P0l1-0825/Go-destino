package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

const maxTicketsPerPurchase = 20

type TicketService struct {
	ticketRepo  *repository.TicketRepository
	routeRepo   *repository.RouteRepository
	paymentRepo *repository.PaymentRepository
	notifSvc    *NotificationService
}

func NewTicketService(
	ticketRepo *repository.TicketRepository,
	routeRepo *repository.RouteRepository,
	paymentRepo *repository.PaymentRepository,
) *TicketService {
	return &TicketService{
		ticketRepo:  ticketRepo,
		routeRepo:   routeRepo,
		paymentRepo: paymentRepo,
	}
}

// SetNotificationService injects the notification service.
func (s *TicketService) SetNotificationService(notifSvc *NotificationService) {
	s.notifSvc = notifSvc
}

func (s *TicketService) PurchaseTickets(ctx context.Context, tenantID, kioskID string, req domain.PurchaseTicketRequest) (*domain.PurchaseTicketResponse, error) {
	// Validate quantity
	if req.Quantity < 1 {
		return nil, fmt.Errorf("quantity must be at least 1")
	}
	if req.Quantity > maxTicketsPerPurchase {
		return nil, fmt.Errorf("maximum %d tickets per purchase", maxTicketsPerPurchase)
	}

	// Validate payment method
	if !domain.ValidPaymentMethod(req.PaymentMethod) {
		return nil, fmt.Errorf("invalid payment method: %s", req.PaymentMethod)
	}

	route, err := s.routeRepo.GetByID(ctx, req.RouteID)
	if err != nil {
		return nil, fmt.Errorf("route not found: %w", err)
	}

	if !route.Active {
		return nil, fmt.Errorf("route is not active")
	}

	totalAmount := route.PriceCents * int64(req.Quantity)

	// Create payment as pending first
	payment := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		KioskID:     kioskID,
		Method:      domain.PaymentMethod(req.PaymentMethod),
		Status:      domain.PaymentPending,
		AmountCents: totalAmount,
		Currency:    route.Currency,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	// Generate tickets
	var tickets []domain.Ticket
	now := time.Now()
	validUntil := now.Add(24 * time.Hour)

	for i := 0; i < req.Quantity; i++ {
		qr, err := generateQRCode()
		if err != nil {
			// Mark payment as failed if ticket generation fails
			_ = s.paymentRepo.MarkFailed(ctx, payment.ID, tenantID, "ticket generation failed")
			return nil, fmt.Errorf("generating QR code: %w", err)
		}

		ticket := domain.Ticket{
			ID:         uuid.New().String(),
			TenantID:   tenantID,
			RouteID:    req.RouteID,
			KioskID:    kioskID,
			PaymentID:  payment.ID,
			QRCode:     qr,
			Status:     domain.TicketActive,
			PriceCents: route.PriceCents,
			Currency:   route.Currency,
			ValidFrom:  now,
			ValidUntil: validUntil,
		}

		if err := s.ticketRepo.Create(ctx, &ticket); err != nil {
			_ = s.paymentRepo.MarkFailed(ctx, payment.ID, tenantID, "ticket creation failed")
			return nil, fmt.Errorf("creating ticket: %w", err)
		}

		tickets = append(tickets, ticket)
	}

	// Mark payment as completed after all tickets created
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, tenantID, domain.PaymentCompleted); err != nil {
		return nil, fmt.Errorf("completing payment: %w", err)
	}
	payment.Status = domain.PaymentCompleted

	// Send ticket purchase confirmation (email + SMS + WhatsApp)
	if s.notifSvc != nil {
		go s.notifSvc.SendTicketPurchaseNotification(
			context.Background(), tenantID, "", // no userID for anonymous kiosk purchases
			tickets, totalAmount, route.Currency, req.PaymentMethod, "es",
		)
	}

	return &domain.PurchaseTicketResponse{
		Tickets: tickets,
		Payment: *payment,
	}, nil
}

func (s *TicketService) ValidateTicket(ctx context.Context, qrCode string) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.GetByQRCode(ctx, qrCode)
	if err != nil {
		return nil, fmt.Errorf("ticket not found")
	}

	if ticket.Status != domain.TicketActive {
		return nil, fmt.Errorf("ticket is %s", ticket.Status)
	}

	if time.Now().After(ticket.ValidUntil) {
		_ = s.ticketRepo.UpdateStatus(ctx, ticket.ID, domain.TicketExpired)
		return nil, fmt.Errorf("ticket has expired")
	}

	if err := s.ticketRepo.UpdateStatus(ctx, ticket.ID, domain.TicketUsed); err != nil {
		return nil, fmt.Errorf("updating ticket status: %w", err)
	}

	ticket.Status = domain.TicketUsed
	return ticket, nil
}

func (s *TicketService) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, id)
}

func (s *TicketService) CancelTicket(ctx context.Context, id string) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}
	if ticket.Status != domain.TicketActive {
		return fmt.Errorf("only active tickets can be cancelled")
	}
	return s.ticketRepo.UpdateStatus(ctx, ticket.ID, domain.TicketCanceled)
}

func generateQRCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "GD-" + hex.EncodeToString(b), nil
}
