package domain

import "time"

// Tenant represents a SaaS customer (transport operator/company).
type Tenant struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	Logo      string    `json:"logo,omitempty" db:"logo"`
	Active    bool      `json:"active" db:"active"`
	Plan      string    `json:"plan" db:"plan"` // free, basic, pro, enterprise
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
