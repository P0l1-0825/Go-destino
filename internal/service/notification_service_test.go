package service

import (
	"strings"
	"testing"
)

// ─── SMS formatters ───

func TestFormatBookingConfirmationSMS(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"es", "Reserva BK-001 confirmada"},
		{"pt", "Reserva BK-001 confirmada"},
		{"en", "Booking BK-001 confirmed"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			msg := FormatBookingConfirmationSMS(tt.lang, "BK-001", "Airport T1", "Hotel Ritz", 150000, "MXN")
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected %q in message, got: %s", tt.contains, msg)
			}
			if !strings.Contains(msg, "Hotel Ritz") {
				t.Errorf("expected destination in message")
			}
			if !strings.Contains(msg, "Airport T1") {
				t.Errorf("expected pickup in message")
			}
		})
	}
}

func TestFormatDriverAssignedSMS(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"es", "conductor Juan"},
		{"pt", "motorista Juan"},
		{"en", "driver Juan"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			msg := FormatDriverAssignedSMS(tt.lang, "BK-002", "Juan", "ABC-123")
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected %q, got: %s", tt.contains, msg)
			}
			if !strings.Contains(msg, "ABC-123") {
				t.Errorf("expected vehicle plate in message")
			}
		})
	}
}

func TestFormatTicketPurchaseSMS(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"es", "2 ticket(s) adquirido(s)"},
		{"pt", "2 bilhete(s) adquirido(s)"},
		{"en", "2 ticket(s) purchased"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			msg := FormatTicketPurchaseSMS(tt.lang, 2, 50000, "MXN", "QR-ABC-123")
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected %q, got: %s", tt.contains, msg)
			}
			if !strings.Contains(msg, "QR-ABC-123") {
				t.Errorf("expected QR code in message")
			}
		})
	}
}

func TestFormatTripCompletedSMS(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"es", "Viaje BK-003 completado"},
		{"pt", "Viagem BK-003 concluída"},
		{"en", "Trip BK-003 completed"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			msg := FormatTripCompletedSMS(tt.lang, "BK-003", 200000, "MXN")
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected %q, got: %s", tt.contains, msg)
			}
		})
	}
}

func TestFormatRefundSMS(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"es", "Reembolso de"},
		{"pt", "Reembolso de"},
		{"en", "Refund of"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			msg := FormatRefundSMS(tt.lang, 75000, "MXN", "REF-001")
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected %q, got: %s", tt.contains, msg)
			}
			if !strings.Contains(msg, "REF-001") {
				t.Errorf("expected reference in message")
			}
		})
	}
}

func TestFormatCancellationSMS(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"es", "reserva BK-004 ha sido cancelada"},
		{"pt", "reserva BK-004 foi cancelada"},
		{"en", "booking BK-004 has been cancelled"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			msg := FormatCancellationSMS(tt.lang, "BK-004")
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected %q, got: %s", tt.contains, msg)
			}
		})
	}
}

// ─── Localized booking messages ───

func TestLocalizedBookingMsg(t *testing.T) {
	tests := []struct {
		lang      string
		event     string
		wantTitle string
	}{
		{"es", "confirmed", "Reserva confirmada"},
		{"pt", "confirmed", "Reserva confirmada"},
		{"en", "confirmed", "Booking confirmed"},
		{"fr", "confirmed", "Booking confirmed"}, // fallback to English
		{"es", "unknown_event", "GoDestino"},      // default case
	}
	for _, tt := range tests {
		t.Run(tt.lang+"_"+tt.event, func(t *testing.T) {
			title, body := localizedBookingMsg(tt.lang, tt.event, "BK-TEST")
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if body == "" {
				t.Error("body should not be empty")
			}
		})
	}
}

