package service

import (
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

func TestConcesionService_CreateValidation(t *testing.T) {
	svc := &ConcesionService{} // nil repo — validation happens first

	tests := []struct {
		name    string
		req     domain.CreateConcesionRequest
		wantErr string
	}{
		{
			name:    "missing name",
			req:     domain.CreateConcesionRequest{Code: "CONC-001"},
			wantErr: "name is required",
		},
		{
			name:    "missing code",
			req:     domain.CreateConcesionRequest{Name: "Test"},
			wantErr: "code is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(nil, "tenant-1", tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.wantErr {
				t.Errorf("expected %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestConcesionService_AssignStaffValidation(t *testing.T) {
	svc := &ConcesionService{} // nil repo — validation happens first

	tests := []struct {
		name    string
		req     domain.AssignStaffRequest
		wantErr string
	}{
		{
			name:    "missing user_id",
			req:     domain.AssignStaffRequest{StaffRole: domain.StaffOperativo},
			wantErr: "user_id is required",
		},
		{
			name:    "invalid staff role",
			req:     domain.AssignStaffRequest{UserID: "u1", StaffRole: "invalid_role"},
			wantErr: "invalid staff role: invalid_role (valid: administrativo, operativo, taxista)",
		},
		{
			name:    "empty staff role is invalid",
			req:     domain.AssignStaffRequest{UserID: "u1", StaffRole: ""},
			wantErr: "invalid staff role:  (valid: administrativo, operativo, taxista)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.AssignStaff(nil, "conc-1", "tenant-1", tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.wantErr {
				t.Errorf("expected %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestValidStaffRoles(t *testing.T) {
	roles := domain.ValidStaffRoles()
	if len(roles) != 3 {
		t.Errorf("expected 3 staff roles, got %d", len(roles))
	}

	expected := map[domain.StaffRole]bool{
		domain.StaffAdministrativo: true,
		domain.StaffOperativo:      true,
		domain.StaffTaxista:        true,
	}
	for _, r := range roles {
		if !expected[r] {
			t.Errorf("unexpected role: %s", r)
		}
	}
}

func TestValidConcesionTypes(t *testing.T) {
	types := domain.ValidConcesionTypes()
	if len(types) != 4 {
		t.Errorf("expected 4 types, got %d", len(types))
	}

	expected := map[domain.ConcesionType]bool{
		domain.ConcesionTaxi:    true,
		domain.ConcesionVan:     true,
		domain.ConcesionShuttle: true,
		domain.ConcesionMixed:   true,
	}
	for _, ct := range types {
		if !expected[ct] {
			t.Errorf("unexpected type: %s", ct)
		}
	}
}

func TestConcesionStatusConstants(t *testing.T) {
	tests := []struct {
		status domain.ConcesionStatus
		val    string
	}{
		{domain.ConcesionActive, "active"},
		{domain.ConcesionInactive, "inactive"},
		{domain.ConcesionSuspended, "suspended"},
		{domain.ConcesionPending, "pending"},
	}
	for _, tc := range tests {
		if string(tc.status) != tc.val {
			t.Errorf("expected %q, got %q", tc.val, tc.status)
		}
	}
}
