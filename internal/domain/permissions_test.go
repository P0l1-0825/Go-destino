package domain

import "testing"

func TestHasPermission_SuperAdmin(t *testing.T) {
	// SuperAdmin should have ALL permissions
	allPerms := AllPermissions()
	for _, perm := range allPerms {
		if !HasPermission(RoleSuperAdmin, perm) {
			t.Errorf("SUPER_ADMIN should have permission %s", perm)
		}
	}
}

func TestHasPermission_RoleSeparation(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		perm Permission
		want bool
	}{
		// Taxista should have own-read but NOT all-read
		{"taxista has res.read.own", RoleTaxista, PermResReadOwn, true},
		{"taxista lacks res.read.all", RoleTaxista, PermResReadAll, false},

		// Taxista should have fleet.location.own but NOT fleet.driver.manage
		{"taxista has fleet.location.own", RoleTaxista, PermFleetLocationOwn, true},
		{"taxista lacks fleet.driver.manage", RoleTaxista, PermFleetDriverManage, false},

		// Vendedor should have kiosk permissions
		{"vendedor has kiosk.book.create", RoleVendedor, PermKioskBookCreate, true},
		{"vendedor has pay.charge", RoleVendedor, PermPayCharge, true},
		{"vendedor lacks sys.users.manage", RoleVendedor, PermSysUsersManage, false},

		// Admin should have user management
		{"admin has sys.users.manage", RoleAdmin, PermSysUsersManage, true},
		{"admin has fleet.driver.manage", RoleAdmin, PermFleetDriverManage, true},
		{"admin lacks sys.roles.create", RoleAdmin, PermSysRolesCreate, false},

		// Usuario should have limited permissions
		{"usuario has res.create.web", RoleUsuario, PermResCreateWeb, true},
		{"usuario has pay.charge", RoleUsuario, PermPayCharge, true},
		{"usuario lacks sys.users.read", RoleUsuario, PermSysUsersRead, false},
		{"usuario lacks fleet.location.view", RoleUsuario, PermFleetLocationView, false},

		// QR validate for specific roles
		{"taxista has qr.validate", RoleTaxista, PermQRValidate, true},
		{"vendedor has qr.validate", RoleVendedor, PermQRValidate, true},
		{"mesa_control has qr.validate", RoleMesaControl, PermQRValidate, true},
		{"usuario lacks qr.validate", RoleUsuario, PermQRValidate, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(tt.role, tt.perm)
			if got != tt.want {
				t.Errorf("HasPermission(%s, %s) = %v, want %v", tt.role, tt.perm, got, tt.want)
			}
		})
	}
}

func TestHasPermission_UnknownRole(t *testing.T) {
	if HasPermission("NONEXISTENT_ROLE", PermResReadAll) {
		t.Error("unknown role should not have any permissions")
	}
}

func TestAllPermissions_NoDuplicates(t *testing.T) {
	all := AllPermissions()
	seen := make(map[Permission]bool)
	for _, p := range all {
		if seen[p] {
			t.Errorf("duplicate permission: %s", p)
		}
		seen[p] = true
	}
}

func TestAllPermissions_Count(t *testing.T) {
	all := AllPermissions()
	// There should be a meaningful number of permissions (at least 70)
	if len(all) < 70 {
		t.Errorf("expected at least 70 permissions, got %d", len(all))
	}
}

func TestRolePermissions_AllRolesDefined(t *testing.T) {
	expectedRoles := []UserRole{
		RoleSuperAdmin, RoleAdmin, RoleClienteConcesion, RoleTesoreriaCliente,
		RoleMesaControl, RoleOperador, RoleTaxista, RoleVendedor, RoleBroker, RoleUsuario,
	}

	for _, role := range expectedRoles {
		perms, ok := RolePermissions[role]
		if !ok {
			t.Errorf("role %s not defined in RolePermissions", role)
			continue
		}
		if len(perms) == 0 {
			t.Errorf("role %s has zero permissions", role)
		}
	}
}
