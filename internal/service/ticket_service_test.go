package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Ticket mocks ---

type mockTicketRepo struct {
	CreateFn       func(ctx context.Context, t *domain.Ticket) error
	GetByIDFn      func(ctx context.Context, id string) (*domain.Ticket, error)
	GetByQRCodeFn  func(ctx context.Context, qr string) (*domain.Ticket, error)
	UpdateStatusFn func(ctx context.Context, id string, status domain.TicketStatus) error
}

func (m *mockTicketRepo) Create(ctx context.Context, t *domain.Ticket) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, t) }
	return nil
}
func (m *mockTicketRepo) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(ctx, id) }
	return nil, fmt.Errorf("not found")
}
func (m *mockTicketRepo) GetByQRCode(ctx context.Context, qr string) (*domain.Ticket, error) {
	if m.GetByQRCodeFn != nil { return m.GetByQRCodeFn(ctx, qr) }
	return nil, fmt.Errorf("not found")
}
func (m *mockTicketRepo) UpdateStatus(ctx context.Context, id string, status domain.TicketStatus) error {
	if m.UpdateStatusFn != nil { return m.UpdateStatusFn(ctx, id, status) }
	return nil
}

type mockRouteReader struct {
	GetByIDFn func(ctx context.Context, id string) (*domain.Route, error)
}

func (m *mockRouteReader) GetByID(ctx context.Context, id string) (*domain.Route, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(ctx, id) }
	return nil, fmt.Errorf("not found")
}

type mockPaymentWriter struct {
	CreateFn       func(ctx context.Context, p *domain.Payment) error
	UpdateStatusFn func(ctx context.Context, id, tenantID string, status domain.PaymentStatus) error
	MarkFailedFn   func(ctx context.Context, id, tenantID, reason string) error
}

func (m *mockPaymentWriter) Create(ctx context.Context, p *domain.Payment) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, p) }
	return nil
}
func (m *mockPaymentWriter) UpdateStatus(ctx context.Context, id, tenantID string, status domain.PaymentStatus) error {
	if m.UpdateStatusFn != nil { return m.UpdateStatusFn(ctx, id, tenantID, status) }
	return nil
}
func (m *mockPaymentWriter) MarkFailed(ctx context.Context, id, tenantID, reason string) error {
	if m.MarkFailedFn != nil { return m.MarkFailedFn(ctx, id, tenantID, reason) }
	return nil
}

func activeRoute() *domain.Route {
	return &domain.Route{
		ID: "route-1", TenantID: testutil.TestTenantID,
		Name: "MEX-Centro", PriceCents: 5000, Currency: "MXN", Active: true,
	}
}

// --- PurchaseTickets ---

func TestTicketService_PurchaseTickets(t *testing.T) {
	tests := []struct {
		name      string
		req       domain.PurchaseTicketRequest
		routeFn   func(context.Context, string) (*domain.Route, error)
		wantErr   string
	}{
		{
			name: "happy path - 2 tickets",
			req:  domain.PurchaseTicketRequest{RouteID: "route-1", Quantity: 2, PaymentMethod: "cash"},
			routeFn: func(_ context.Context, _ string) (*domain.Route, error) { return activeRoute(), nil },
		},
		{
			name:    "quantity zero",
			req:     domain.PurchaseTicketRequest{RouteID: "route-1", Quantity: 0, PaymentMethod: "cash"},
			wantErr: "quantity must be at least 1",
		},
		{
			name:    "quantity over max",
			req:     domain.PurchaseTicketRequest{RouteID: "route-1", Quantity: 21, PaymentMethod: "cash"},
			wantErr: "maximum 20 tickets",
		},
		{
			name:    "invalid payment method",
			req:     domain.PurchaseTicketRequest{RouteID: "route-1", Quantity: 1, PaymentMethod: "bitcoin"},
			wantErr: "invalid payment method",
		},
		{
			name: "route not found",
			req:  domain.PurchaseTicketRequest{RouteID: "bad", Quantity: 1, PaymentMethod: "cash"},
			routeFn: func(_ context.Context, _ string) (*domain.Route, error) {
				return nil, fmt.Errorf("not found")
			},
			wantErr: "route not found",
		},
		{
			name: "inactive route",
			req:  domain.PurchaseTicketRequest{RouteID: "route-1", Quantity: 1, PaymentMethod: "cash"},
			routeFn: func(_ context.Context, _ string) (*domain.Route, error) {
				r := activeRoute()
				r.Active = false
				return r, nil
			},
			wantErr: "route is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routeRepo := &mockRouteReader{}
			if tt.routeFn != nil {
				routeRepo.GetByIDFn = tt.routeFn
			}
			svc := NewTicketService(&mockTicketRepo{}, routeRepo, &mockPaymentWriter{})

			resp, err := svc.PurchaseTickets(context.Background(), testutil.TestTenantID, testutil.TestKioskID, tt.req)
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if len(resp.Tickets) != tt.req.Quantity {
				t.Errorf("got %d tickets, want %d", len(resp.Tickets), tt.req.Quantity)
			}
			if resp.Payment.Status != domain.PaymentCompleted {
				t.Errorf("payment status = %s, want completed", resp.Payment.Status)
			}
			expectedTotal := activeRoute().PriceCents * int64(tt.req.Quantity)
			if resp.Payment.AmountCents != expectedTotal {
				t.Errorf("amount = %d, want %d", resp.Payment.AmountCents, expectedTotal)
			}
		})
	}
}

// --- ValidateTicket ---

func TestTicketService_ValidateTicket(t *testing.T) {
	tests := []struct {
		name    string
		ticket  *domain.Ticket
		wantErr string
	}{
		{
			name:   "valid ticket",
			ticket: &domain.Ticket{ID: "t1", Status: domain.TicketActive, ValidUntil: time.Now().Add(1 * time.Hour)},
		},
		{
			name:    "already used",
			ticket:  &domain.Ticket{ID: "t2", Status: domain.TicketUsed, ValidUntil: time.Now().Add(1 * time.Hour)},
			wantErr: "ticket is used",
		},
		{
			name:    "expired",
			ticket:  &domain.Ticket{ID: "t3", Status: domain.TicketActive, ValidUntil: time.Now().Add(-1 * time.Hour)},
			wantErr: "ticket has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewTicketService(
				&mockTicketRepo{
					GetByQRCodeFn: func(_ context.Context, _ string) (*domain.Ticket, error) { return tt.ticket, nil },
				},
				&mockRouteReader{},
				&mockPaymentWriter{},
			)

			ticket, err := svc.ValidateTicket(context.Background(), "GD-qrcode")
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if ticket.Status != domain.TicketUsed {
				t.Errorf("status = %s, want used", ticket.Status)
			}
		})
	}
}

func TestTicketService_ValidateTicket_NotFound(t *testing.T) {
	svc := NewTicketService(&mockTicketRepo{}, &mockRouteReader{}, &mockPaymentWriter{})
	_, err := svc.ValidateTicket(context.Background(), "invalid-qr")
	testutil.AssertError(t, err, "ticket not found")
}

// --- CancelTicket ---

func TestTicketService_CancelTicket(t *testing.T) {
	t.Run("active ticket", func(t *testing.T) {
		svc := NewTicketService(
			&mockTicketRepo{
				GetByIDFn: func(_ context.Context, _ string) (*domain.Ticket, error) {
					return &domain.Ticket{ID: "t1", Status: domain.TicketActive}, nil
				},
			},
			&mockRouteReader{},
			&mockPaymentWriter{},
		)
		err := svc.CancelTicket(context.Background(), "t1")
		testutil.AssertNoError(t, err)
	})

	t.Run("already used", func(t *testing.T) {
		svc := NewTicketService(
			&mockTicketRepo{
				GetByIDFn: func(_ context.Context, _ string) (*domain.Ticket, error) {
					return &domain.Ticket{ID: "t2", Status: domain.TicketUsed}, nil
				},
			},
			&mockRouteReader{},
			&mockPaymentWriter{},
		)
		err := svc.CancelTicket(context.Background(), "t2")
		testutil.AssertError(t, err, "only active tickets can be cancelled")
	})
}
