package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
	"github.com/P0l1-0825/Go-destino/internal/testutil/mocks"
)

func newTestPaymentService(paymentRepo *mocks.MockPaymentRepo) *PaymentService {
	return NewPaymentService(paymentRepo, &mocks.MockNotificationService{}, &mocks.MockAuditService{})
}

// --- ProcessPayment ---

func TestPaymentService_ProcessPayment(t *testing.T) {
	tests := []struct {
		name      string
		method    domain.PaymentMethod
		mockSetup func(*mocks.MockPaymentRepo)
		wantErr   string
		wantStatus domain.PaymentStatus
	}{
		{
			name:   "cash payment succeeds",
			method: domain.PaymentCash,
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Payment) error { return nil }
				m.UpdateStatusFn = func(_ context.Context, _, _ string, _ domain.PaymentStatus) error { return nil }
			},
			wantStatus: domain.PaymentCompleted,
		},
		{
			name:   "card payment succeeds",
			method: domain.PaymentCard,
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Payment) error { return nil }
				m.UpdateStatusFn = func(_ context.Context, _, _ string, _ domain.PaymentStatus) error { return nil }
			},
			wantStatus: domain.PaymentCompleted,
		},
		{
			name:   "QR payment succeeds",
			method: domain.PaymentQR,
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Payment) error { return nil }
				m.UpdateStatusFn = func(_ context.Context, _, _ string, _ domain.PaymentStatus) error { return nil }
			},
			wantStatus: domain.PaymentCompleted,
		},
		{
			name:   "unsupported payment method",
			method: "bitcoin",
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Payment) error { return nil }
				m.MarkFailedFn = func(_ context.Context, _, _, _ string) error { return nil }
			},
			wantErr: "payment failed",
		},
		{
			name:   "repo create error",
			method: domain.PaymentCash,
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Payment) error { return fmt.Errorf("db error") }
			},
			wantErr: "creating payment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockPaymentRepo{}
			tt.mockSetup(repo)
			svc := newTestPaymentService(repo)

			payment, err := svc.ProcessPayment(context.Background(), ProcessPaymentRequest{
				TenantID:    testutil.TestTenantID,
				UserID:      testutil.TestUserID,
				BookingID:   testutil.TestBookingID,
				KioskID:     testutil.TestKioskID,
				Method:      tt.method,
				AmountCents: 15000,
				Currency:    "MXN",
				Lang:        "es",
			})
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if payment == nil {
				t.Fatal("expected payment, got nil")
			}
			if payment.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", payment.Status, tt.wantStatus)
			}
			if payment.Reference == "" {
				t.Error("expected non-empty reference")
			}
		})
	}
}

// --- RefundPayment ---

func TestPaymentService_RefundPayment(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*mocks.MockPaymentRepo)
		wantErr   string
	}{
		{
			name: "happy path",
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.GetByIDTenantFn = func(_ context.Context, _, _ string) (*domain.Payment, error) {
					return testutil.NewTestPayment(), nil // status = completed
				}
				m.RefundFn = func(_ context.Context, _, _ string, _ *domain.Payment) error { return nil }
			},
		},
		{
			name: "payment not found",
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.GetByIDTenantFn = func(_ context.Context, _, _ string) (*domain.Payment, error) {
					return nil, fmt.Errorf("sql: no rows")
				}
			},
			wantErr: "payment not found",
		},
		{
			name: "payment not completed",
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.GetByIDTenantFn = func(_ context.Context, _, _ string) (*domain.Payment, error) {
					p := testutil.NewTestPayment()
					p.Status = domain.PaymentPending
					return p, nil
				}
			},
			wantErr: "can only refund completed payments",
		},
		{
			name: "refund repo error",
			mockSetup: func(m *mocks.MockPaymentRepo) {
				m.GetByIDTenantFn = func(_ context.Context, _, _ string) (*domain.Payment, error) {
					return testutil.NewTestPayment(), nil
				}
				m.RefundFn = func(_ context.Context, _, _ string, _ *domain.Payment) error {
					return fmt.Errorf("tx failed")
				}
			},
			wantErr: "processing refund",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockPaymentRepo{}
			tt.mockSetup(repo)
			svc := newTestPaymentService(repo)

			refund, err := svc.RefundPayment(context.Background(), testutil.TestPaymentID, testutil.TestTenantID, testutil.TestUserID, "test refund", "es")
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if refund == nil {
				t.Fatal("expected refund payment, got nil")
			}
		})
	}
}

// --- GetPayment ---

func TestPaymentService_GetPayment(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := &mocks.MockPaymentRepo{
			GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Payment, error) {
				return testutil.NewTestPayment(), nil
			},
		}
		svc := newTestPaymentService(repo)

		p, err := svc.GetPayment(context.Background(), testutil.TestPaymentID, testutil.TestTenantID)
		testutil.AssertNoError(t, err)
		if p.ID != testutil.TestPaymentID {
			t.Errorf("id = %s, want %s", p.ID, testutil.TestPaymentID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mocks.MockPaymentRepo{
			GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Payment, error) {
				return nil, fmt.Errorf("sql: no rows")
			},
		}
		svc := newTestPaymentService(repo)

		_, err := svc.GetPayment(context.Background(), "nonexistent", testutil.TestTenantID)
		if err == nil {
			t.Error("expected error for nonexistent payment")
		}
	})
}

// --- ListPayments ---

func TestPaymentService_ListPayments(t *testing.T) {
	repo := &mocks.MockPaymentRepo{
		ListByTenantFn: func(_ context.Context, _ string, limit, offset int) ([]domain.Payment, error) {
			return []domain.Payment{*testutil.NewTestPayment()}, nil
		},
	}
	svc := newTestPaymentService(repo)

	tests := []struct {
		name  string
		limit int
	}{
		{"default limit", 0},
		{"custom limit", 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payments, err := svc.ListPayments(context.Background(), testutil.TestTenantID, tt.limit, 0)
			testutil.AssertNoError(t, err)
			if len(payments) == 0 {
				t.Error("expected at least one payment")
			}
		})
	}
}
