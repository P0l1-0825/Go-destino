package mocks

import (
	"context"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

// --- UserRepo mock ---

type MockUserRepo struct {
	CreateFn          func(ctx context.Context, u *domain.User) error
	GetByIDFn         func(ctx context.Context, id string) (*domain.User, error)
	GetByEmailFn      func(ctx context.Context, tenantID, email string) (*domain.User, error)
	ExistsByEmailFn   func(ctx context.Context, tenantID, email string) (bool, error)
	ChangePasswordFn  func(ctx context.Context, userID, newHash string) error
	UpdateLastLoginFn func(ctx context.Context, id string) error
}

func (m *MockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, u)
	}
	return nil
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, tenantID, email)
	}
	return nil, nil
}

func (m *MockUserRepo) ExistsByEmail(ctx context.Context, tenantID, email string) (bool, error) {
	if m.ExistsByEmailFn != nil {
		return m.ExistsByEmailFn(ctx, tenantID, email)
	}
	return false, nil
}

func (m *MockUserRepo) ChangePassword(ctx context.Context, userID, newHash string) error {
	if m.ChangePasswordFn != nil {
		return m.ChangePasswordFn(ctx, userID, newHash)
	}
	return nil
}

func (m *MockUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	if m.UpdateLastLoginFn != nil {
		return m.UpdateLastLoginFn(ctx, id)
	}
	return nil
}

// --- BookingRepo mock ---

type MockBookingRepo struct {
	CreateFn          func(ctx context.Context, b *domain.Booking) error
	GetByIDFn         func(ctx context.Context, id string) (*domain.Booking, error)
	GetByIDTenantFn   func(ctx context.Context, id, tenantID string) (*domain.Booking, error)
	GetByNumberFn     func(ctx context.Context, number string) (*domain.Booking, error)
	GetByNumberTenantFn func(ctx context.Context, number, tenantID string) (*domain.Booking, error)
	UpdateStatusFn    func(ctx context.Context, id, tenantID string, status domain.BookingStatus) error
	AssignDriverFn    func(ctx context.Context, id, tenantID, driverID, vehicleID string) error
	SetStartedFn      func(ctx context.Context, id, tenantID string) error
	SetCompletedFn    func(ctx context.Context, id, tenantID string) error
	SetCancelledFn    func(ctx context.Context, id, tenantID, reason string) error
	ListByTenantFn    func(ctx context.Context, tenantID string, limit int) ([]domain.Booking, error)
	ListFilteredFn    func(ctx context.Context, f domain.ListBookingsFilter) ([]domain.Booking, int, error)
}

func (m *MockBookingRepo) Create(ctx context.Context, b *domain.Booking) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, b)
	}
	return nil
}

func (m *MockBookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockBookingRepo) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Booking, error) {
	if m.GetByIDTenantFn != nil {
		return m.GetByIDTenantFn(ctx, id, tenantID)
	}
	return nil, nil
}

func (m *MockBookingRepo) GetByNumber(ctx context.Context, number string) (*domain.Booking, error) {
	if m.GetByNumberFn != nil {
		return m.GetByNumberFn(ctx, number)
	}
	return nil, nil
}

func (m *MockBookingRepo) GetByNumberTenant(ctx context.Context, number, tenantID string) (*domain.Booking, error) {
	if m.GetByNumberTenantFn != nil {
		return m.GetByNumberTenantFn(ctx, number, tenantID)
	}
	return nil, nil
}

func (m *MockBookingRepo) UpdateStatus(ctx context.Context, id, tenantID string, status domain.BookingStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, tenantID, status)
	}
	return nil
}

func (m *MockBookingRepo) AssignDriver(ctx context.Context, id, tenantID, driverID, vehicleID string) error {
	if m.AssignDriverFn != nil {
		return m.AssignDriverFn(ctx, id, tenantID, driverID, vehicleID)
	}
	return nil
}

func (m *MockBookingRepo) SetStarted(ctx context.Context, id, tenantID string) error {
	if m.SetStartedFn != nil {
		return m.SetStartedFn(ctx, id, tenantID)
	}
	return nil
}

func (m *MockBookingRepo) SetCompleted(ctx context.Context, id, tenantID string) error {
	if m.SetCompletedFn != nil {
		return m.SetCompletedFn(ctx, id, tenantID)
	}
	return nil
}

func (m *MockBookingRepo) SetCancelled(ctx context.Context, id, tenantID, reason string) error {
	if m.SetCancelledFn != nil {
		return m.SetCancelledFn(ctx, id, tenantID, reason)
	}
	return nil
}

func (m *MockBookingRepo) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.Booking, error) {
	if m.ListByTenantFn != nil {
		return m.ListByTenantFn(ctx, tenantID, limit)
	}
	return nil, nil
}

func (m *MockBookingRepo) ListFiltered(ctx context.Context, f domain.ListBookingsFilter) ([]domain.Booking, int, error) {
	if m.ListFilteredFn != nil {
		return m.ListFilteredFn(ctx, f)
	}
	return nil, 0, nil
}

// --- PaymentRepo mock ---

type MockPaymentRepo struct {
	CreateFn             func(ctx context.Context, p *domain.Payment) error
	GetByIDTenantFn      func(ctx context.Context, id, tenantID string) (*domain.Payment, error)
	GetByBookingIDTenantFn func(ctx context.Context, bookingID, tenantID string) (*domain.Payment, error)
	UpdateStatusFn       func(ctx context.Context, id, tenantID string, status domain.PaymentStatus) error
	MarkFailedFn         func(ctx context.Context, id, tenantID, reason string) error
	RefundFn             func(ctx context.Context, originalID, tenantID string, refundPayment *domain.Payment) error
	ListByTenantFn       func(ctx context.Context, tenantID string, limit, offset int) ([]domain.Payment, error)
}

func (m *MockPaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, p)
	}
	return nil
}

func (m *MockPaymentRepo) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Payment, error) {
	if m.GetByIDTenantFn != nil {
		return m.GetByIDTenantFn(ctx, id, tenantID)
	}
	return nil, nil
}

func (m *MockPaymentRepo) GetByBookingIDTenant(ctx context.Context, bookingID, tenantID string) (*domain.Payment, error) {
	if m.GetByBookingIDTenantFn != nil {
		return m.GetByBookingIDTenantFn(ctx, bookingID, tenantID)
	}
	return nil, nil
}

func (m *MockPaymentRepo) UpdateStatus(ctx context.Context, id, tenantID string, status domain.PaymentStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, tenantID, status)
	}
	return nil
}

func (m *MockPaymentRepo) MarkFailed(ctx context.Context, id, tenantID, reason string) error {
	if m.MarkFailedFn != nil {
		return m.MarkFailedFn(ctx, id, tenantID, reason)
	}
	return nil
}

func (m *MockPaymentRepo) Refund(ctx context.Context, originalID, tenantID string, refundPayment *domain.Payment) error {
	if m.RefundFn != nil {
		return m.RefundFn(ctx, originalID, tenantID, refundPayment)
	}
	return nil
}

func (m *MockPaymentRepo) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.Payment, error) {
	if m.ListByTenantFn != nil {
		return m.ListByTenantFn(ctx, tenantID, limit, offset)
	}
	return nil, nil
}

// --- AuditService mock ---

type MockAuditService struct {
	LogFn func(ctx context.Context, tenantID, userID, action, resource, resourceID, details, ip, ua string)
}

func (m *MockAuditService) Log(ctx context.Context, tenantID, userID, action, resource, resourceID, details, ip, ua string) {
	if m.LogFn != nil {
		m.LogFn(ctx, tenantID, userID, action, resource, resourceID, details, ip, ua)
	}
}

// --- NotificationService mock ---

type MockNotificationService struct{}

func (m *MockNotificationService) SendBookingConfirmationFull(ctx context.Context, tenantID string, booking *domain.Booking, lang string) {
}
func (m *MockNotificationService) SendDriverAssignedFull(ctx context.Context, tenantID string, booking *domain.Booking, driverID, vehicleID, lang string) {
}
func (m *MockNotificationService) SendTripCompletedFull(ctx context.Context, tenantID string, booking *domain.Booking, paymentMethod, lang string) {
}
func (m *MockNotificationService) SendCancellationNotification(ctx context.Context, tenantID string, booking *domain.Booking, reason, lang string) {
}
func (m *MockNotificationService) SendPaymentReceipt(ctx context.Context, tenantID, userID string, payment *domain.Payment, bookingNumber, lang string) {
}
func (m *MockNotificationService) SendRefundNotification(ctx context.Context, tenantID, userID string, amountCents int64, currency, reference, reason, lang string) {
}
