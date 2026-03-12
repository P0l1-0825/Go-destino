package domain

import "time"

// Airport represents a registered airport in the platform.
type Airport struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	City        string    `json:"city" db:"city"`
	Country     string    `json:"country" db:"country"`
	CountryCode string    `json:"country_code" db:"country_code"`
	Lat         float64   `json:"lat" db:"lat"`
	Lng         float64   `json:"lng" db:"lng"`
	Timezone    string    `json:"timezone" db:"timezone"`
	Terminals   []AirportTerminal `json:"terminals,omitempty"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateAirportRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Timezone    string  `json:"timezone"`
}
