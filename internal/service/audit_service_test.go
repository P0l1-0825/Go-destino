package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Audit mocks ---

type mockAuditRepo struct {
	mu       sync.Mutex
	CreateFn       func(ctx context.Context, e *domain.AuditLogEntry) error
	ListByTenantFn func(ctx context.Context, tenantID string, limit int) ([]domain.AuditLogEntry, error)
	ListByUserFn   func(ctx context.Context, userID string, limit int) ([]domain.AuditLogEntry, error)
	createCalls    int
}

func (m *mockAuditRepo) Create(ctx context.Context, e *domain.AuditLogEntry) error {
	m.mu.Lock()
	m.createCalls++
	m.mu.Unlock()
	if m.CreateFn != nil { return m.CreateFn(ctx, e) }
	return nil
}
func (m *mockAuditRepo) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.AuditLogEntry, error) {
	if m.ListByTenantFn != nil { return m.ListByTenantFn(ctx, tenantID, limit) }
	return nil, nil
}
func (m *mockAuditRepo) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLogEntry, error) {
	if m.ListByUserFn != nil { return m.ListByUserFn(ctx, userID, limit) }
	return nil, nil
}

// --- Tests ---

func TestAuditService_Log_FireAndForget(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo)

	svc.Log(context.Background(), testutil.TestTenantID, testutil.TestUserID,
		"login.success", "auth", testutil.TestUserID, "user logged in", "127.0.0.1", "test-agent")

	// Wait for goroutine to complete
	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	calls := repo.createCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("Create called %d times, want 1", calls)
	}
}

func TestAuditService_ListByTenant_DefaultLimit(t *testing.T) {
	repo := &mockAuditRepo{
		ListByTenantFn: func(_ context.Context, _ string, limit int) ([]domain.AuditLogEntry, error) {
			if limit != 100 { t.Errorf("limit = %d, want 100", limit) }
			return []domain.AuditLogEntry{{ID: "a1"}}, nil
		},
	}
	svc := NewAuditService(repo)
	entries, err := svc.ListByTenant(context.Background(), testutil.TestTenantID, 0)
	testutil.AssertNoError(t, err)
	if len(entries) != 1 { t.Errorf("expected 1, got %d", len(entries)) }
}

func TestAuditService_ListByUser_DefaultLimit(t *testing.T) {
	repo := &mockAuditRepo{
		ListByUserFn: func(_ context.Context, _ string, limit int) ([]domain.AuditLogEntry, error) {
			if limit != 100 { t.Errorf("limit = %d, want 100", limit) }
			return nil, nil
		},
	}
	svc := NewAuditService(repo)
	_, err := svc.ListByUser(context.Background(), testutil.TestUserID, 0)
	testutil.AssertNoError(t, err)
}

func TestAuditService_ListByTenant_CustomLimit(t *testing.T) {
	repo := &mockAuditRepo{
		ListByTenantFn: func(_ context.Context, _ string, limit int) ([]domain.AuditLogEntry, error) {
			if limit != 25 { t.Errorf("limit = %d, want 25", limit) }
			return nil, nil
		},
	}
	svc := NewAuditService(repo)
	_, _ = svc.ListByTenant(context.Background(), testutil.TestTenantID, 25)
}
