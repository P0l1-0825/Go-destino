package domain

import "time"

type DriverStatus string

const (
	DriverOffline   DriverStatus = "offline"
	DriverAvailable DriverStatus = "available"
	DriverBusy      DriverStatus = "busy"
	DriverOnTrip    DriverStatus = "on_trip"
	DriverEmergency DriverStatus = "emergency"
)

type VehicleType string

const (
	VehicleSedan   VehicleType = "sedan"
	VehicleSUV     VehicleType = "suv"
	VehicleVan     VehicleType = "van"
	VehicleMinibus VehicleType = "minibus"
	VehicleBus     VehicleType = "bus"
)

type DocType string

const (
	DocLicense       DocType = "license"
	DocInsurance     DocType = "insurance"
	DocRegistration  DocType = "registration"
	DocBackground    DocType = "background_check"
	DocProfilePhoto  DocType = "profile_photo"
	DocVehiclePhoto  DocType = "vehicle_photo"
)

// Driver represents a registered transport operator.
type Driver struct {
	ID                string       `json:"id" db:"id"`
	TenantID          string       `json:"tenant_id" db:"tenant_id"`
	UserID            string       `json:"user_id" db:"user_id"`
	ConcesionID       string       `json:"concesion_id,omitempty" db:"concesion_id"`
	CompanyID         string       `json:"company_id,omitempty" db:"company_id"` // deprecated: use concesion_id
	LicenseNumber     string       `json:"license_number" db:"license_number"`
	Status            DriverStatus `json:"status" db:"status"`
	SubRole           string       `json:"sub_role" db:"sub_role"` // DRIVER_TAXI, DRIVER_VAN, DRIVER_BUS, DRIVER_SHUTTLE
	Rating            float64      `json:"rating" db:"rating"`
	TotalTrips        int          `json:"total_trips" db:"total_trips"`
	DocsVerified      bool         `json:"docs_verified" db:"docs_verified"`
	BiometricVerified bool         `json:"biometric_verified" db:"biometric_verified"`
	AirportIDs        []string     `json:"airport_ids,omitempty"`
	CurrentLat        float64      `json:"current_lat" db:"current_lat"`
	CurrentLng        float64      `json:"current_lng" db:"current_lng"`
	Heading           float64      `json:"heading" db:"heading"`
	Speed             float64      `json:"speed" db:"speed"`
	LastLocationAt    *time.Time   `json:"last_location_at,omitempty" db:"last_location_at"`
	CreatedAt         time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
}

// Vehicle represents a transport vehicle.
type Vehicle struct {
	ID          string      `json:"id" db:"id"`
	TenantID    string      `json:"tenant_id" db:"tenant_id"`
	DriverID    string      `json:"driver_id" db:"driver_id"`
	ConcesionID string      `json:"concesion_id,omitempty" db:"concesion_id"`
	CompanyID   string      `json:"company_id,omitempty" db:"company_id"` // deprecated: use concesion_id
	Plate       string      `json:"plate" db:"plate"`
	Brand       string      `json:"brand" db:"brand"`
	Model       string      `json:"model" db:"model"`
	Year        int         `json:"year" db:"year"`
	Color       string      `json:"color" db:"color"`
	Type        VehicleType `json:"type" db:"type"`
	Capacity    int         `json:"capacity" db:"capacity"`
	Features    []string    `json:"features,omitempty"`
	Status      string      `json:"status" db:"status"` // active, maintenance, inactive
	InsuranceID string      `json:"insurance_id,omitempty" db:"insurance_id"`
	InsuranceExp *time.Time `json:"insurance_exp,omitempty" db:"insurance_exp"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// DriverDocument represents uploaded verification documents.
type DriverDocument struct {
	ID         string    `json:"id" db:"id"`
	DriverID   string    `json:"driver_id" db:"driver_id"`
	DocType    DocType   `json:"doc_type" db:"doc_type"`
	FileURL    string    `json:"file_url" db:"file_url"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Verified   bool      `json:"verified" db:"verified"`
	VerifiedBy string    `json:"verified_by,omitempty" db:"verified_by"`
	VerifiedAt *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// DriverLocation is a GPS update from the driver app (every 3-5s).
type DriverLocation struct {
	DriverID  string  `json:"driver_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
	Accuracy  float64 `json:"accuracy"`
	Timestamp int64   `json:"timestamp"`
}

type RegisterDriverRequest struct {
	UserID        string `json:"user_id"`
	LicenseNumber string `json:"license_number"`
	SubRole       string `json:"sub_role"`
	ConcesionID   string `json:"concesion_id,omitempty"`
	CompanyID     string `json:"company_id,omitempty"` // deprecated: use concesion_id
}

type RegisterVehicleRequest struct {
	DriverID string      `json:"driver_id"`
	Plate    string      `json:"plate"`
	Brand    string      `json:"brand"`
	Model    string      `json:"model"`
	Year     int         `json:"year"`
	Color    string      `json:"color"`
	Type     VehicleType `json:"type"`
	Capacity int         `json:"capacity"`
}

type NearbyDriversRequest struct {
	Lat       float64     `json:"lat"`
	Lng       float64     `json:"lng"`
	RadiusKM  float64     `json:"radius_km"`
	Type      VehicleType `json:"type,omitempty"`
	MinRating float64     `json:"min_rating,omitempty"`
}

type DriverEarnings struct {
	DriverID    string `json:"driver_id"`
	Period      string `json:"period"` // today, week, month
	TotalCents  int64  `json:"total_cents"`
	TripCount   int    `json:"trip_count"`
	AvgRating   float64 `json:"avg_rating"`
	Currency    string `json:"currency"`
}
