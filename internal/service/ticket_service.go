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

type TicketService struct {
	ticketRepo  *repository.TicketRepository
	routeRepo   *repository.RouteRepository
	paymentRepo *repository.PaymentRepository
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

func (s *TicketService) PurchaseTickets(ctx context.Context, tenantID, kioskID string, req domain.PurchaseTicketRequest) (*domain.PurchaseTicketResponse, error) {
	route, err := s.routeRepo.GetByID(ctx, req.RouteID)
	if err != nil {
		return nil, fmt.Errorf("route not found: %w", err)
	}

	if !route.Active {
		return nil, fmt.Errorf("route is not active")
	}

	totalAmount := route.PriceCents * int64(req.Quantity)

	payment := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		KioskID:     kioskID,
		Method:      domain.PaymentMethod(req.PaymentMethod),
		Status:      domain.PaymentCompleted,
		AmountCents: totalAmount,
		Currency:    route.Currency,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	var tickets []domain.Ticket
	now := time.Now()
	validUntil := now.Add(24 * time.Hour)

	for i := 0; i < req.Quantity; i++ {
		qr, err := generateQRCode()
		if err != nil {
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
			return nil, fmt.Errorf("creating ticket: %w", err)
		}

		tickets = append(tickets, ticket)
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

func generateQRCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "GD-" + hex.EncodeToString(b), nil
}
