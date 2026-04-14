package domain

import "time"

type UserRole string

const (
	RoleSuperAdmin      UserRole = "SUPER_ADMIN"
	RoleAdmin           UserRole = "ADMINISTRADOR"
	RoleClienteConcesion UserRole = "CLIENTE_CONCESION"
	RoleTesoreriaCliente UserRole = "TESORERIA_CLIENTE"
	RoleMesaControl     UserRole = "MESA_CONTROL"
	RoleOperador        UserRole = "OPERADOR"
	RoleTaxista         UserRole = "TAXISTA"
	RoleVendedor        UserRole = "VENDEDOR"
	RoleBroker          UserRole = "BROKER"
	RoleUsuario         UserRole = "USUARIO"
)

// RoleLevel returns the privilege level of a role (lower = more privileged).
// Used for role hierarchy enforcement: a user cannot assign a role at or above their own level.
func RoleLevel(role UserRole) int {
	switch role {
	case RoleSuperAdmin:
		return 0
	case RoleAdmin:
		return 1
	case RoleClienteConcesion, RoleTesoreriaCliente:
		return 2
	case RoleMesaControl, RoleOperador:
		return 3
	case RoleTaxista, RoleVendedor, RoleBroker:
		return 4
	case RoleUsuario:
		return 5
	default:
		return 99
	}
}

// User represents an operator, admin, driver, seller, or end-user.
type User struct {
	ID           string    `json:"id" db:"id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	Email        string    `json:"email" db:"email"`
	Phone        string    `json:"phone,omitempty" db:"phone"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	Role         UserRole  `json:"role" db:"role"`
	SubRole      string    `json:"sub_role,omitempty" db:"sub_role"`
	ConcesionID  string    `json:"concesion_id,omitempty" db:"concesion_id"`
	CompanyID    string    `json:"company_id,omitempty" db:"company_id"` // deprecated: use concesion_id
	AirportIDs   []string  `json:"airport_ids,omitempty"`
	Lang         string    `json:"lang" db:"lang"`
	Active       bool      `json:"active" db:"active"`
	MFAEnabled   bool      `json:"mfa_enabled" db:"mfa_enabled"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type CreateUserRequest struct {
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Password  string   `json:"password"`
	Name      string   `json:"name"`
	Role      UserRole `json:"role"`
	SubRole   string   `json:"sub_role,omitempty"`
	ConcesionID string `json:"concesion_id,omitempty"`
	CompanyID   string `json:"company_id,omitempty"` // deprecated: use concesion_id
	Lang        string `json:"lang"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
