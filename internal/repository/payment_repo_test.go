package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

func TestPaymentRepo_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewPaymentRepository(db)

	p := testutil.NewTestPayment()

	mock.ExpectExec("INSERT INTO payments").
		WithArgs(p.ID, p.TenantID, p.BookingID, p.KioskID, p.UserID,
			p.Method, p.Status, p.AmountCents, p.Currency, p.Reference).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentRepo_GetByIDTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewPaymentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(testutil.PaymentColumns).
		AddRow(testutil.TestPaymentID, testutil.TestTenantID, testutil.TestBookingID,
			testutil.TestKioskID, testutil.TestUserID,
			"card", "completed", int64(15000), "MXN",
			"PAY-test", "", nil, now, now)

	mock.ExpectQuery("SELECT .+ FROM payments WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(testutil.TestPaymentID, testutil.TestTenantID).
		WillReturnRows(rows)

	payment, err := repo.GetByIDTenant(context.Background(), testutil.TestPaymentID, testutil.TestTenantID)
	if err != nil {
		t.Fatalf("GetByIDTenant: %v", err)
	}
	if payment.TenantID != testutil.TestTenantID {
		t.Errorf("tenant_id = %s, want %s", payment.TenantID, testutil.TestTenantID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentRepo_UpdateStatus_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectExec("UPDATE payments SET status = \\$1.+WHERE id = \\$2 AND tenant_id = \\$3").
		WithArgs(domain.PaymentCompleted, testutil.TestPaymentID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), testutil.TestPaymentID, testutil.TestTenantID, domain.PaymentCompleted)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentRepo_MarkFailed_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectExec("UPDATE payments SET status = 'failed', failure_reason = \\$1.+WHERE id = \\$2 AND tenant_id = \\$3").
		WithArgs("gateway timeout", testutil.TestPaymentID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkFailed(context.Background(), testutil.TestPaymentID, testutil.TestTenantID, "gateway timeout")
	if err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentRepo_Refund_Transaction(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewPaymentRepository(db)

	refund := testutil.NewTestPayment()
	refund.ID = "refund-id"
	refund.Reference = "REF-test"

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE payments SET status = 'refunded'.+WHERE id = \\$1 AND tenant_id = \\$2 AND status = 'completed'").
		WithArgs(testutil.TestPaymentID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO payments").
		WithArgs(refund.ID, refund.TenantID, refund.BookingID, refund.KioskID, refund.UserID,
			refund.Method, -refund.AmountCents, refund.Currency, refund.Reference).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Refund(context.Background(), testutil.TestPaymentID, testutil.TestTenantID, refund)
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPaymentRepo_ListByTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewPaymentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(testutil.PaymentColumns).
		AddRow(testutil.TestPaymentID, testutil.TestTenantID, testutil.TestBookingID,
			testutil.TestKioskID, testutil.TestUserID,
			"card", "completed", int64(15000), "MXN",
			"PAY-test", "", nil, now, now)

	mock.ExpectQuery("SELECT .+ FROM payments WHERE tenant_id = \\$1").
		WithArgs(testutil.TestTenantID, 50, 0).
		WillReturnRows(rows)

	payments, err := repo.ListByTenant(context.Background(), testutil.TestTenantID, 50, 0)
	if err != nil {
		t.Fatalf("ListByTenant: %v", err)
	}
	if len(payments) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(payments))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
