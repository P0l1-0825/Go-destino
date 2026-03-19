package domain

import "time"

// ConcesionStatus represents the lifecycle state of a concession.
type ConcesionStatus string

const (
	ConcesionActive    ConcesionStatus = "active"
	ConcesionInactive  ConcesionStatus = "inactive"
	ConcesionSuspended ConcesionStatus = "suspended"
	ConcesionPending   ConcesionStatus = "pending" // awaiting approval
)

// ConcesionType categorizes the type of transport franchise.
type ConcesionType string

const (
	ConcesionTaxi    ConcesionType = "taxi"
	ConcesionVan     ConcesionType = "van"
	ConcesionShuttle ConcesionType = "shuttle"
	ConcesionMixed   ConcesionType = "mixed" // operates multiple vehicle types
)

// StaffRole represents the internal hierarchy within a concession.
type StaffRole string

const (
	StaffAdministrativo StaffRole = "administrativo" // manages the concession
	StaffOperativo      StaffRole = "operativo"      // dispatches, monitors operations
	StaffTaxista        StaffRole = "taxista"         // drives vehicles
)

// Concesion represents a transport franchise/concession that operates
// within a tenant. Each concesion owns vehicles (unidades) and has
// its own staff hierarchy: administrativo → operativo → taxista.
type Concesion struct {
	ID          string          `json:"id" db:"id"`
	TenantID    string          `json:"tenant_id" db:"tenant_id"`
	Name        string          `json:"name" db:"name"`
	Code        string          `json:"code" db:"code"`           // business identifier (e.g., "CONC-001")
	RFC         string          `json:"rfc,omitempty" db:"rfc"`   // tax ID (Mexico)
	Type        ConcesionType   `json:"type" db:"type"`
	Status      ConcesionStatus `json:"status" db:"status"`
	ManagerID   string          `json:"manager_id,omitempty" db:"manager_id"` // FK to users (administrativo principal)
	Phone       string          `json:"phone,omitempty" db:"phone"`
	Email       string          `json:"email,omitempty" db:"email"`
	Address     string          `json:"address,omitempty" db:"address"`
	AirportIDs  []string        `json:"airport_ids,omitempty"`                // airports where they operate
	MaxVehicles int             `json:"max_vehicles" db:"max_vehicles"`       // contract limit
	MaxDrivers  int             `json:"max_drivers" db:"max_drivers"`         // contract limit
	LogoURL     string          `json:"logo_url,omitempty" db:"logo_url"`
	Notes       string          `json:"notes,omitempty" db:"notes"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`

	// Aggregated stats (populated by service layer, not stored in DB)
	VehicleCount int `json:"vehicle_count,omitempty" db:"-"`
	DriverCount  int `json:"driver_count,omitempty" db:"-"`
	StaffCount   int `json:"staff_count,omitempty" db:"-"`
}

// ConcesionSummary is a lightweight view for lists and dropdowns.
type ConcesionSummary struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Code         string          `json:"code"`
	Type         ConcesionType   `json:"type"`
	Status       ConcesionStatus `json:"status"`
	VehicleCount int             `json:"vehicle_count"`
	DriverCount  int             `json:"driver_count"`
}

// ── Request/Response types ──

type CreateConcesionRequest struct {
	Name        string        `json:"name"`
	Code        string        `json:"code"`
	RFC         string        `json:"rfc,omitempty"`
	Type        ConcesionType `json:"type"`
	Phone       string        `json:"phone,omitempty"`
	Email       string        `json:"email,omitempty"`
	Address     string        `json:"address,omitempty"`
	MaxVehicles int           `json:"max_vehicles"`
	MaxDrivers  int           `json:"max_drivers"`
}

type UpdateConcesionRequest struct {
	Name        *string          `json:"name,omitempty"`
	Phone       *string          `json:"phone,omitempty"`
	Email       *string          `json:"email,omitempty"`
	Address     *string          `json:"address,omitempty"`
	Status      *ConcesionStatus `json:"status,omitempty"`
	MaxVehicles *int             `json:"max_vehicles,omitempty"`
	MaxDrivers  *int             `json:"max_drivers,omitempty"`
	LogoURL     *string          `json:"logo_url,omitempty"`
	Notes       *string          `json:"notes,omitempty"`
}

type AssignStaffRequest struct {
	UserID    string    `json:"user_id"`
	StaffRole StaffRole `json:"staff_role"`
}

type ListConcesionesFilter struct {
	TenantID string          `json:"tenant_id"`
	Status   ConcesionStatus `json:"status,omitempty"`
	Type     ConcesionType   `json:"type,omitempty"`
	Search   string          `json:"search,omitempty"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}

// ValidConcesionTypes returns all valid concesion types.
func ValidConcesionTypes() []ConcesionType {
	return []ConcesionType{ConcesionTaxi, ConcesionVan, ConcesionShuttle, ConcesionMixed}
}

// ValidStaffRoles returns all valid staff roles within a concesion.
func ValidStaffRoles() []StaffRole {
	return []StaffRole{StaffAdministrativo, StaffOperativo, StaffTaxista}
}
