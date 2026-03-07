package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/P0l1-0825/Go-destino/internal/config"
)

// TwilioService handles SMS and WhatsApp message delivery via Twilio API.
type TwilioService struct {
	cfg     config.TwilioConfig
	enabled bool
	client  *http.Client
	baseURL string
}

func NewTwilioService(cfg config.TwilioConfig) *TwilioService {
	return &TwilioService{
		cfg:     cfg,
		enabled: cfg.Enabled && cfg.AccountSID != "",
		client:  &http.Client{},
		baseURL: fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s", cfg.AccountSID),
	}
}

// SendSMS sends a text message via Twilio.
func (s *TwilioService) SendSMS(to, body string) error {
	if !s.enabled {
		log.Printf("[SMS] (dry-run) → %s | %s", to, body)
		return nil
	}

	return s.sendMessage(s.cfg.SMSFrom, to, body)
}

// SendWhatsApp sends a WhatsApp message via Twilio.
func (s *TwilioService) SendWhatsApp(to, body string) error {
	if !s.enabled {
		log.Printf("[WHATSAPP] (dry-run) → %s | %s", to, body)
		return nil
	}

	// Ensure WhatsApp prefix
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	return s.sendMessage(s.cfg.WhatsAppFrom, to, body)
}

// sendMessage posts a message to the Twilio Messages API.
func (s *TwilioService) sendMessage(from, to, body string) error {
	endpoint := s.baseURL + "/Messages.json"

	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	data.Set("Body", body)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(s.cfg.AccountSID, s.cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("twilio request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var twilioErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &twilioErr)
		log.Printf("[TWILIO] error %d → %s: %s", twilioErr.Code, to, twilioErr.Message)
		return fmt.Errorf("twilio error %d: %s", resp.StatusCode, twilioErr.Message)
	}

	var result struct {
		SID    string `json:"sid"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(respBody, &result)

	channel := "SMS"
	if strings.HasPrefix(to, "whatsapp:") {
		channel = "WHATSAPP"
	}
	log.Printf("[%s] sent → %s | SID: %s | Status: %s", channel, to, result.SID, result.Status)

	return nil
}

// IsEnabled returns true if Twilio is configured and enabled.
func (s *TwilioService) IsEnabled() bool {
	return s.enabled
}

// --- SMS/WhatsApp template messages ---

// FormatBookingConfirmationSMS returns a localized booking confirmation for SMS/WhatsApp.
func FormatBookingConfirmationSMS(lang, bookingNumber, pickup, dropoff string, priceCents int64, currency string) string {
	price := formatMoney(priceCents, currency)
	switch lang {
	case "es":
		return fmt.Sprintf("GoDestino ✓ Reserva %s confirmada.\nRecogida: %s\nDestino: %s\nTotal: %s\nPresenta tu QR al abordar.", bookingNumber, pickup, dropoff, price)
	case "pt":
		return fmt.Sprintf("GoDestino ✓ Reserva %s confirmada.\nEmbarque: %s\nDestino: %s\nTotal: %s\nApresente seu QR ao embarcar.", bookingNumber, pickup, dropoff, price)
	default:
		return fmt.Sprintf("GoDestino ✓ Booking %s confirmed.\nPickup: %s\nDropoff: %s\nTotal: %s\nShow your QR when boarding.", bookingNumber, pickup, dropoff, price)
	}
}

// FormatDriverAssignedSMS returns a localized driver assignment message.
func FormatDriverAssignedSMS(lang, bookingNumber, driverName, vehiclePlate string) string {
	switch lang {
	case "es":
		return fmt.Sprintf("GoDestino — Tu conductor %s va en camino. Vehículo: %s. Reserva: %s", driverName, vehiclePlate, bookingNumber)
	case "pt":
		return fmt.Sprintf("GoDestino — Seu motorista %s está a caminho. Veículo: %s. Reserva: %s", driverName, vehiclePlate, bookingNumber)
	default:
		return fmt.Sprintf("GoDestino — Your driver %s is on the way. Vehicle: %s. Booking: %s", driverName, vehiclePlate, bookingNumber)
	}
}

// FormatTicketPurchaseSMS returns a localized ticket purchase message.
func FormatTicketPurchaseSMS(lang string, quantity int, priceCents int64, currency, qrCode string) string {
	price := formatMoney(priceCents, currency)
	switch lang {
	case "es":
		return fmt.Sprintf("GoDestino ✓ %d ticket(s) adquirido(s).\nTotal: %s\nQR: %s\nPresenta tu QR al abordar.", quantity, price, qrCode)
	case "pt":
		return fmt.Sprintf("GoDestino ✓ %d bilhete(s) adquirido(s).\nTotal: %s\nQR: %s\nApresente seu QR ao embarcar.", quantity, price, qrCode)
	default:
		return fmt.Sprintf("GoDestino ✓ %d ticket(s) purchased.\nTotal: %s\nQR: %s\nShow your QR when boarding.", quantity, price, qrCode)
	}
}

// FormatTripCompletedSMS returns a localized trip completion message.
func FormatTripCompletedSMS(lang, bookingNumber string, priceCents int64, currency string) string {
	price := formatMoney(priceCents, currency)
	switch lang {
	case "es":
		return fmt.Sprintf("GoDestino — Viaje %s completado. Total: %s. ¡Gracias por viajar con nosotros!", bookingNumber, price)
	case "pt":
		return fmt.Sprintf("GoDestino — Viagem %s concluída. Total: %s. Obrigado por viajar conosco!", bookingNumber, price)
	default:
		return fmt.Sprintf("GoDestino — Trip %s completed. Total: %s. Thank you for traveling with us!", bookingNumber, price)
	}
}

// FormatRefundSMS returns a localized refund confirmation message.
func FormatRefundSMS(lang string, refundCents int64, currency, reference string) string {
	price := formatMoney(refundCents, currency)
	switch lang {
	case "es":
		return fmt.Sprintf("GoDestino — Reembolso de %s procesado. Ref: %s", price, reference)
	case "pt":
		return fmt.Sprintf("GoDestino — Reembolso de %s processado. Ref: %s", price, reference)
	default:
		return fmt.Sprintf("GoDestino — Refund of %s processed. Ref: %s", price, reference)
	}
}

// FormatCancellationSMS returns a localized booking cancellation message.
func FormatCancellationSMS(lang, bookingNumber string) string {
	switch lang {
	case "es":
		return fmt.Sprintf("GoDestino — Tu reserva %s ha sido cancelada. Si necesitas ayuda: soporte@godestino.com", bookingNumber)
	case "pt":
		return fmt.Sprintf("GoDestino — Sua reserva %s foi cancelada. Se precisar de ajuda: soporte@godestino.com", bookingNumber)
	default:
		return fmt.Sprintf("GoDestino — Your booking %s has been cancelled. Need help? soporte@godestino.com", bookingNumber)
	}
}
