package geo

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name            string
		lat1, lng1      float64
		lat2, lng2      float64
		expectedKM      float64
		toleranceKM     float64
	}{
		{
			name:        "same point",
			lat1: 19.4326, lng1: -99.1332,
			lat2: 19.4326, lng2: -99.1332,
			expectedKM: 0, toleranceKM: 0.001,
		},
		{
			name:        "Cancun airport to Playa del Carmen (~65 km)",
			lat1: 21.0365, lng1: -86.8771, // CUN airport
			lat2: 20.6296, lng2: -87.0739, // Playa del Carmen
			expectedKM: 50, toleranceKM: 15, // rough estimate
		},
		{
			name:        "CDMX to Guadalajara (~460 km)",
			lat1: 19.4326, lng1: -99.1332,
			lat2: 20.6597, lng2: -103.3496,
			expectedKM: 460, toleranceKM: 30,
		},
		{
			name:        "equator points 1 degree apart (~111 km)",
			lat1: 0, lng1: 0,
			lat2: 0, lng2: 1,
			expectedKM: 111.19, toleranceKM: 1,
		},
		{
			name:        "North Pole to South Pole (~20000 km)",
			lat1: 90, lng1: 0,
			lat2: -90, lng2: 0,
			expectedKM: 20015, toleranceKM: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Haversine(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			if math.Abs(got-tt.expectedKM) > tt.toleranceKM {
				t.Errorf("Haversine(%f,%f → %f,%f) = %f km, expected ~%f km (±%f)",
					tt.lat1, tt.lng1, tt.lat2, tt.lng2, got, tt.expectedKM, tt.toleranceKM)
			}
		})
	}
}

func TestHaversine_Symmetry(t *testing.T) {
	// Distance A→B should equal B→A
	d1 := Haversine(19.4326, -99.1332, 20.6597, -103.3496)
	d2 := Haversine(20.6597, -103.3496, 19.4326, -99.1332)

	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("Haversine not symmetric: A→B = %f, B→A = %f", d1, d2)
	}
}

func TestHaversine_NonNegative(t *testing.T) {
	// Distance should never be negative
	d := Haversine(90, 180, -90, -180)
	if d < 0 {
		t.Errorf("Haversine returned negative distance: %f", d)
	}
}
