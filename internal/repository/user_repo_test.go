package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

func TestUserRepo_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	user := testutil.NewTestUser()

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.TenantID, user.Email, user.Phone, user.PasswordHash,
			user.Name, user.Role, user.SubRole, user.CompanyID, user.Lang, user.Active).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepo_GetByEmail_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(testutil.UserColumns).
		AddRow(testutil.TestUserID, testutil.TestTenantID, "user@test.com", "+52555", "hash",
			"Test", domain.RoleUsuario, "", "", "es", true, false, now, now, nil)

	mock.ExpectQuery("SELECT .+ FROM users WHERE tenant_id = \\$1 AND email = \\$2").
		WithArgs(testutil.TestTenantID, "user@test.com").
		WillReturnRows(rows)

	user, err := repo.GetByEmail(context.Background(), testutil.TestTenantID, "user@test.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if user.TenantID != testutil.TestTenantID {
		t.Errorf("tenant_id = %s, want %s", user.TenantID, testutil.TestTenantID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepo_GetByIDTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(testutil.UserColumns).
		AddRow(testutil.TestUserID, testutil.TestTenantID, "user@test.com", "+52555", "hash",
			"Test", domain.RoleUsuario, "", "", "es", true, false, now, now, nil)

	// The query MUST contain both id AND tenant_id
	mock.ExpectQuery("SELECT .+ FROM users WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(testutil.TestUserID, testutil.TestTenantID).
		WillReturnRows(rows)

	user, err := repo.GetByIDTenant(context.Background(), testutil.TestUserID, testutil.TestTenantID)
	if err != nil {
		t.Fatalf("GetByIDTenant: %v", err)
	}
	if user.ID != testutil.TestUserID {
		t.Errorf("id = %s, want %s", user.ID, testutil.TestUserID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepo_ExistsByEmail_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery("SELECT EXISTS.+FROM users WHERE tenant_id = \\$1 AND email = \\$2").
		WithArgs(testutil.TestTenantID, "exists@test.com").
		WillReturnRows(rows)

	exists, err := repo.ExistsByEmail(context.Background(), testutil.TestTenantID, "exists@test.com")
	if err != nil {
		t.Fatalf("ExistsByEmail: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepo_ListByTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	// ListByTenant doesn't include password_hash in SELECT
	listCols := []string{
		"id", "tenant_id", "email", "phone", "name", "role",
		"sub_role", "company_id", "lang", "active", "mfa_enabled",
		"created_at", "updated_at", "last_login",
	}
	now := time.Now()
	rows := sqlmock.NewRows(listCols).
		AddRow(testutil.TestUserID, testutil.TestTenantID, "user@test.com", "+52555",
			"Test", domain.RoleUsuario, "", "", "es", true, false, now, now, nil)

	mock.ExpectQuery("SELECT .+ FROM users WHERE tenant_id = \\$1").
		WithArgs(testutil.TestTenantID).
		WillReturnRows(rows)

	users, err := repo.ListByTenant(context.Background(), testutil.TestTenantID)
	if err != nil {
		t.Fatalf("ListByTenant: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].TenantID != testutil.TestTenantID {
		t.Errorf("tenant_id = %s, want %s", users[0].TenantID, testutil.TestTenantID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepo_ChangePassword(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	mock.ExpectExec("UPDATE users SET password_hash = \\$1").
		WithArgs("newhash", testutil.TestUserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.ChangePassword(context.Background(), testutil.TestUserID, "newhash")
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepo_DeactivateTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewUserRepository(db)

	mock.ExpectExec("UPDATE users SET active = false.+WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(testutil.TestUserID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeactivateTenant(context.Background(), testutil.TestUserID, testutil.TestTenantID)
	if err != nil {
		t.Fatalf("DeactivateTenant: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
