package domain

import "testing"

func TestValidBookingTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    BookingStatus
		to      BookingStatus
		wantErr bool
	}{
		// Valid transitions
		{"pending to confirmed", BookingPending, BookingConfirmed, false},
		{"pending to cancelled", BookingPending, BookingCancelled, false},
		{"confirmed to assigned", BookingConfirmed, BookingAssigned, false},
		{"confirmed to cancelled", BookingConfirmed, BookingCancelled, false},
		{"assigned to started", BookingAssigned, BookingStarted, false},
		{"assigned to cancelled", BookingAssigned, BookingCancelled, false},
		{"started to completed", BookingStarted, BookingCompleted, false},
		// Invalid transitions
		{"pending to started", BookingPending, BookingStarted, true},
		{"pending to completed", BookingPending, BookingCompleted, true},
		{"confirmed to started", BookingConfirmed, BookingStarted, true},
		{"confirmed to completed", BookingConfirmed, BookingCompleted, true},
		{"assigned to confirmed", BookingAssigned, BookingConfirmed, true},
		{"started to cancelled", BookingStarted, BookingCancelled, true},
		{"completed to anything", BookingCompleted, BookingCancelled, true},
		{"cancelled to anything", BookingCancelled, BookingPending, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidBookingTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidBookingTransition(%s, %s) error = %v, wantErr %v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}

func TestValidServiceType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"taxi", true},
		{"shuttle", true},
		{"van", true},
		{"bus", true},
		{"", false},
		{"helicopter", false},
		{"TAXI", false}, // case-sensitive
		{"Taxi", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidServiceType(tt.input); got != tt.want {
				t.Errorf("ValidServiceType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidPaymentMethod(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"cash", true},
		{"card", true},
		{"qr", true},
		{"", false},
		{"bitcoin", false},
		{"CASH", false}, // case-sensitive
		{"Cash", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidPaymentMethod(tt.input); got != tt.want {
				t.Errorf("ValidPaymentMethod(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
