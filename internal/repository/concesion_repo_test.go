package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

var concesionColumns = []string{
	"id", "tenant_id", "name", "code", "rfc", "type", "status",
	"manager_id", "phone", "email", "address",
	"max_vehicles", "max_drivers", "logo_url", "notes",
	"created_at", "updated_at",
}

func TestConcesionRepo_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	c := &domain.Concesion{
		TenantID:    testutil.TestTenantID,
		Name:        "Test Concesion",
		Code:        "CONC-001",
		Type:        domain.ConcesionTaxi,
		Status:      domain.ConcesionPending,
		MaxVehicles: 20,
		MaxDrivers:  30,
	}

	mock.ExpectExec("INSERT INTO concesiones").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), c)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID == "" {
		t.Error("ID should be generated")
	}
	if c.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_GetByID_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(concesionColumns).
		AddRow("c1", testutil.TestTenantID, "Test", "CONC-001", "", "taxi", "active",
			nil, "+52555", "test@c.com", "Addr",
			20, 30, "", "", now, now)

	// Query MUST contain both id AND tenant_id
	mock.ExpectQuery("SELECT .+ FROM concesiones WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs("c1", testutil.TestTenantID).
		WillReturnRows(rows)

	c, err := repo.GetByID(context.Background(), "c1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.TenantID != testutil.TestTenantID {
		t.Errorf("tenant_id = %s, want %s", c.TenantID, testutil.TestTenantID)
	}
	if c.Name != "Test" {
		t.Errorf("name = %s, want Test", c.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_Update_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	c := &domain.Concesion{
		ID:          "c1",
		TenantID:    testutil.TestTenantID,
		Name:        "Updated",
		Status:      domain.ConcesionActive,
		MaxVehicles: 50,
		MaxDrivers:  60,
	}

	// UPDATE query MUST contain WHERE id=$X AND tenant_id=$Y
	mock.ExpectExec("UPDATE concesiones SET .+ WHERE id=\\$11 AND tenant_id=\\$12").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), c)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_Delete_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	// DELETE MUST filter by tenant_id
	mock.ExpectExec("DELETE FROM concesiones WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs("c1", testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "c1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_CountDrivers_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery("SELECT COUNT.+ FROM drivers WHERE concesion_id = \\$1 AND tenant_id = \\$2").
		WithArgs("c1", testutil.TestTenantID).
		WillReturnRows(rows)

	count, err := repo.CountDrivers(context.Background(), "c1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("CountDrivers: %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_CountVehicles_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(12)
	mock.ExpectQuery("SELECT COUNT.+ FROM vehicles WHERE concesion_id = \\$1 AND tenant_id = \\$2").
		WithArgs("c1", testutil.TestTenantID).
		WillReturnRows(rows)

	count, err := repo.CountVehicles(context.Background(), "c1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("CountVehicles: %v", err)
	}
	if count != 12 {
		t.Errorf("count = %d, want 12", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_CountStaff_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(8)
	mock.ExpectQuery("SELECT COUNT.+ FROM users WHERE concesion_id = \\$1 AND tenant_id = \\$2").
		WithArgs("c1", testutil.TestTenantID).
		WillReturnRows(rows)

	count, err := repo.CountStaff(context.Background(), "c1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("CountStaff: %v", err)
	}
	if count != 8 {
		t.Errorf("count = %d, want 8", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_AssignStaff_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	mock.ExpectExec("UPDATE users SET concesion_id = \\$1.+WHERE id = \\$2 AND tenant_id = \\$3").
		WithArgs("c1", "u1", testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AssignStaff(context.Background(), "u1", "c1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("AssignStaff: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_RemoveStaff_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	mock.ExpectExec("UPDATE users SET concesion_id = NULL.+WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs("u1", testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveStaff(context.Background(), "u1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("RemoveStaff: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConcesionRepo_SetManager_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewConcesionRepository(db)

	mock.ExpectExec("UPDATE concesiones SET manager_id = \\$1.+WHERE id = \\$2 AND tenant_id = \\$3").
		WithArgs("u1", "c1", testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetManager(context.Background(), "c1", "u1", testutil.TestTenantID)
	if err != nil {
		t.Fatalf("SetManager: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
