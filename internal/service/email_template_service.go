package service

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// EmailTemplateService renders HTML email templates with i18n support.
type EmailTemplateService struct {
	templates map[string]*template.Template
}

func NewEmailTemplateService() *EmailTemplateService {
	svc := &EmailTemplateService{
		templates: make(map[string]*template.Template),
	}
	svc.registerAll()
	return svc
}

// TemplateData is the common data passed to all email templates.
type TemplateData struct {
	Lang           string
	BookingNumber  string
	TicketIDs      []string
	ReceiptNumber  string
	ServiceType    string
	Pickup         string
	Dropoff        string
	Passengers     int
	PriceCents     int64
	Currency       string
	PaymentMethod  string
	PaymentRef     string
	FlightNumber   string
	DriverName     string
	VehiclePlate   string
	QRCode         string
	ScheduledAt    string
	IssuedAt       string
	RefundAmount   int64
	RefundReason   string
	CancelReason   string
	CompanyName    string
	SupportEmail   string
	Year           int
}

func DefaultTemplateData() TemplateData {
	return TemplateData{
		CompanyName:  "GoDestino",
		SupportEmail: "soporte@godestino.com",
		Year:         time.Now().Year(),
	}
}

// RenderBookingConfirmation renders the booking confirmation email.
func (s *EmailTemplateService) RenderBookingConfirmation(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "booking_confirmation", data.BookingNumber)
	html, err = s.render("booking_confirmation", data)
	return
}

// RenderTicketPurchase renders the ticket purchase confirmation email.
func (s *EmailTemplateService) RenderTicketPurchase(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "ticket_purchase", "")
	html, err = s.render("ticket_purchase", data)
	return
}

// RenderPaymentReceipt renders the payment receipt email.
func (s *EmailTemplateService) RenderPaymentReceipt(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "payment_receipt", data.ReceiptNumber)
	html, err = s.render("payment_receipt", data)
	return
}

// RenderDriverAssigned renders the driver assignment notification email.
func (s *EmailTemplateService) RenderDriverAssigned(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "driver_assigned", data.BookingNumber)
	html, err = s.render("driver_assigned", data)
	return
}

// RenderTripCompleted renders the trip completion email with receipt.
func (s *EmailTemplateService) RenderTripCompleted(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "trip_completed", data.BookingNumber)
	html, err = s.render("trip_completed", data)
	return
}

// RenderRefundConfirmation renders the refund confirmation email.
func (s *EmailTemplateService) RenderRefundConfirmation(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "refund_confirmation", data.PaymentRef)
	html, err = s.render("refund_confirmation", data)
	return
}

// RenderBookingCancellation renders the booking cancellation email.
func (s *EmailTemplateService) RenderBookingCancellation(data TemplateData) (subject, html string, err error) {
	subject = s.localizedSubject(data.Lang, "booking_cancellation", data.BookingNumber)
	html, err = s.render("booking_cancellation", data)
	return
}

func (s *EmailTemplateService) render(name string, data TemplateData) (string, error) {
	tmpl, ok := s.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering template %s: %w", name, err)
	}
	return buf.String(), nil
}

func (s *EmailTemplateService) localizedSubject(lang, key, ref string) string {
	subjects := map[string]map[string]string{
		"booking_confirmation": {
			"es": "Reserva confirmada",
			"pt": "Reserva confirmada",
			"en": "Booking confirmed",
		},
		"ticket_purchase": {
			"es": "Ticket de transporte adquirido",
			"pt": "Bilhete de transporte adquirido",
			"en": "Transport ticket purchased",
		},
		"payment_receipt": {
			"es": "Recibo de pago",
			"pt": "Recibo de pagamento",
			"en": "Payment receipt",
		},
		"driver_assigned": {
			"es": "Conductor asignado a tu reserva",
			"pt": "Motorista designado para sua reserva",
			"en": "Driver assigned to your booking",
		},
		"trip_completed": {
			"es": "Viaje completado",
			"pt": "Viagem concluída",
			"en": "Trip completed",
		},
		"refund_confirmation": {
			"es": "Reembolso procesado",
			"pt": "Reembolso processado",
			"en": "Refund processed",
		},
		"booking_cancellation": {
			"es": "Reserva cancelada",
			"pt": "Reserva cancelada",
			"en": "Booking cancelled",
		},
	}

	m, ok := subjects[key]
	if !ok {
		return "GoDestino"
	}
	subj, ok := m[lang]
	if !ok {
		subj = m["en"]
	}
	if ref != "" {
		subj += " — " + ref
	}
	return subj
}

func (s *EmailTemplateService) registerAll() {
	funcMap := template.FuncMap{
		"formatMoney": formatMoney,
		"t":           translateLabel,
	}

	for name, body := range emailTemplates {
		tmpl := template.Must(template.New(name).Funcs(funcMap).Parse(baseLayout))
		template.Must(tmpl.New("content").Parse(body))
		s.templates[name] = tmpl
	}
}

func formatMoney(cents int64, currency string) string {
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	sym := "$"
	switch strings.ToUpper(currency) {
	case "BRL":
		sym = "R$"
	case "USD":
		sym = "US$"
	case "COP":
		sym = "COL$"
	}
	return fmt.Sprintf("%s%d.%02d %s", sym, whole, frac, currency)
}

func translateLabel(lang, key string) string {
	labels := map[string]map[string]string{
		"booking_number":  {"es": "Número de reserva", "pt": "Número da reserva", "en": "Booking number"},
		"service":         {"es": "Servicio", "pt": "Serviço", "en": "Service"},
		"pickup":          {"es": "Punto de recogida", "pt": "Ponto de embarque", "en": "Pickup"},
		"dropoff":         {"es": "Destino", "pt": "Destino", "en": "Dropoff"},
		"passengers":      {"es": "Pasajeros", "pt": "Passageiros", "en": "Passengers"},
		"price":           {"es": "Precio", "pt": "Preço", "en": "Price"},
		"payment_method":  {"es": "Método de pago", "pt": "Forma de pagamento", "en": "Payment method"},
		"flight":          {"es": "Vuelo", "pt": "Voo", "en": "Flight"},
		"driver":          {"es": "Conductor", "pt": "Motorista", "en": "Driver"},
		"vehicle":         {"es": "Vehículo", "pt": "Veículo", "en": "Vehicle"},
		"receipt":         {"es": "Recibo", "pt": "Recibo", "en": "Receipt"},
		"total":           {"es": "Total", "pt": "Total", "en": "Total"},
		"refund_amount":   {"es": "Monto reembolsado", "pt": "Valor reembolsado", "en": "Refund amount"},
		"cancel_reason":   {"es": "Motivo de cancelación", "pt": "Motivo do cancelamento", "en": "Cancellation reason"},
		"support":         {"es": "¿Necesitas ayuda?", "pt": "Precisa de ajuda?", "en": "Need help?"},
		"footer_rights":   {"es": "Todos los derechos reservados", "pt": "Todos os direitos reservados", "en": "All rights reserved"},
		"ticket_id":       {"es": "ID de ticket", "pt": "ID do bilhete", "en": "Ticket ID"},
		"valid_qr":        {"es": "Presenta este QR al abordar", "pt": "Apresente este QR ao embarcar", "en": "Show this QR when boarding"},
		"thank_you":       {"es": "¡Gracias por viajar con nosotros!", "pt": "Obrigado por viajar conosco!", "en": "Thank you for traveling with us!"},
		"scheduled_at":    {"es": "Fecha programada", "pt": "Data agendada", "en": "Scheduled date"},
		"on_the_way":      {"es": "Tu conductor va en camino", "pt": "Seu motorista está a caminho", "en": "Your driver is on the way"},
		"trip_summary":    {"es": "Resumen del viaje", "pt": "Resumo da viagem", "en": "Trip summary"},
		"refund_processed": {"es": "Tu reembolso ha sido procesado", "pt": "Seu reembolso foi processado", "en": "Your refund has been processed"},
		"booking_cancelled": {"es": "Tu reserva ha sido cancelada", "pt": "Sua reserva foi cancelada", "en": "Your booking has been cancelled"},
	}
	m, ok := labels[key]
	if !ok {
		return key
	}
	v, ok := m[lang]
	if !ok {
		return m["en"]
	}
	return v
}

// --- HTML Templates ---

const baseLayout = `<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  body { margin: 0; padding: 0; background: #040C1F; font-family: 'Plus Jakarta Sans', -apple-system, sans-serif; color: #CBD5E1; }
  .container { max-width: 600px; margin: 0 auto; background: #070F28; border-radius: 12px; overflow: hidden; }
  .header { background: #0D1B5E; padding: 24px; text-align: center; }
  .header h1 { color: #F1F5F9; font-family: 'Syne', sans-serif; font-size: 24px; margin: 0; }
  .body { padding: 32px 24px; }
  .row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #1C2F62; }
  .row .label { color: #94A3B8; font-size: 13px; }
  .row .value { color: #F1F5F9; font-weight: 600; font-size: 14px; text-align: right; }
  .highlight { background: #0B1535; border-radius: 8px; padding: 20px; margin: 20px 0; text-align: center; }
  .highlight .amount { font-size: 28px; font-weight: 800; color: #38BDF8; font-family: 'JetBrains Mono', monospace; }
  .badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; }
  .badge-blue { background: rgba(37,99,235,0.2); color: #38BDF8; }
  .badge-green { background: rgba(16,185,129,0.2); color: #10B981; }
  .badge-orange { background: rgba(232,112,32,0.2); color: #E87020; }
  .qr-box { background: #FFFFFF; border-radius: 8px; padding: 16px; text-align: center; margin: 20px auto; max-width: 200px; }
  .qr-code { font-family: 'JetBrains Mono', monospace; font-size: 14px; color: #0D1B5E; word-break: break-all; }
  .cta { display: block; background: #2563EB; color: #FFFFFF; text-decoration: none; text-align: center; padding: 14px 24px; border-radius: 8px; font-weight: 700; margin: 24px 0; }
  .cta:hover { background: #1D4ED8; }
  .footer { background: #040C1F; padding: 24px; text-align: center; font-size: 12px; color: #64748B; }
  .footer a { color: #38BDF8; text-decoration: none; }
  .divider { border: none; border-top: 1px solid #1C2F62; margin: 20px 0; }
  .text-center { text-align: center; }
  .text-muted { color: #94A3B8; font-size: 13px; }
  .mb-16 { margin-bottom: 16px; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>GoDestino</h1>
  </div>
  <div class="body">
    {{template "content" .}}
  </div>
  <div class="footer">
    <p>{{t .Lang "support"}} <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
    <p>&copy; {{.Year}} {{.CompanyName}} — {{t .Lang "footer_rights"}}</p>
  </div>
</div>
</body>
</html>`

var emailTemplates = map[string]string{
	"booking_confirmation": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "booking_number"}}: <span style="color:#38BDF8;">{{.BookingNumber}}</span></h2>
<div class="row"><span class="label">{{t .Lang "service"}}</span><span class="value"><span class="badge badge-blue">{{.ServiceType}}</span></span></div>
<div class="row"><span class="label">{{t .Lang "pickup"}}</span><span class="value">{{.Pickup}}</span></div>
<div class="row"><span class="label">{{t .Lang "dropoff"}}</span><span class="value">{{.Dropoff}}</span></div>
<div class="row"><span class="label">{{t .Lang "passengers"}}</span><span class="value">{{.Passengers}}</span></div>
{{if .FlightNumber}}<div class="row"><span class="label">{{t .Lang "flight"}}</span><span class="value">{{.FlightNumber}}</span></div>{{end}}
{{if .ScheduledAt}}<div class="row"><span class="label">{{t .Lang "scheduled_at"}}</span><span class="value">{{.ScheduledAt}}</span></div>{{end}}
<div class="highlight">
  <div class="text-muted mb-16">{{t .Lang "total"}}</div>
  <div class="amount">{{formatMoney .PriceCents .Currency}}</div>
  <div class="text-muted" style="margin-top:8px;">{{.PaymentMethod}}</div>
</div>
{{if .QRCode}}
<div class="qr-box">
  <div class="qr-code">{{.QRCode}}</div>
</div>
<p class="text-center text-muted">{{t .Lang "valid_qr"}}</p>
{{end}}
<p class="text-center" style="color:#F1F5F9;">{{t .Lang "thank_you"}}</p>`,

	"ticket_purchase": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "receipt"}}</h2>
{{range .TicketIDs}}
<div class="row"><span class="label">{{t $.Lang "ticket_id"}}</span><span class="value" style="font-family:'JetBrains Mono',monospace; font-size:12px;">{{.}}</span></div>
{{end}}
<div class="highlight">
  <div class="text-muted mb-16">{{t .Lang "total"}}</div>
  <div class="amount">{{formatMoney .PriceCents .Currency}}</div>
  <div class="text-muted" style="margin-top:8px;">{{.PaymentMethod}}</div>
</div>
{{if .QRCode}}
<div class="qr-box">
  <div class="qr-code">{{.QRCode}}</div>
</div>
<p class="text-center text-muted">{{t .Lang "valid_qr"}}</p>
{{end}}
<p class="text-center" style="color:#F1F5F9;">{{t .Lang "thank_you"}}</p>`,

	"payment_receipt": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "receipt"}} #{{.ReceiptNumber}}</h2>
{{if .BookingNumber}}<div class="row"><span class="label">{{t .Lang "booking_number"}}</span><span class="value">{{.BookingNumber}}</span></div>{{end}}
<div class="row"><span class="label">{{t .Lang "payment_method"}}</span><span class="value">{{.PaymentMethod}}</span></div>
{{if .PaymentRef}}<div class="row"><span class="label">Ref</span><span class="value" style="font-family:'JetBrains Mono',monospace;">{{.PaymentRef}}</span></div>{{end}}
<div class="highlight">
  <div class="text-muted mb-16">{{t .Lang "total"}}</div>
  <div class="amount">{{formatMoney .PriceCents .Currency}}</div>
</div>
<p class="text-center text-muted">{{.IssuedAt}}</p>`,

	"driver_assigned": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "on_the_way"}}</h2>
<div class="row"><span class="label">{{t .Lang "booking_number"}}</span><span class="value" style="color:#38BDF8;">{{.BookingNumber}}</span></div>
<div class="row"><span class="label">{{t .Lang "driver"}}</span><span class="value">{{.DriverName}}</span></div>
<div class="row"><span class="label">{{t .Lang "vehicle"}}</span><span class="value"><span class="badge badge-orange">{{.VehiclePlate}}</span></span></div>
<div class="row"><span class="label">{{t .Lang "pickup"}}</span><span class="value">{{.Pickup}}</span></div>
<div class="row"><span class="label">{{t .Lang "dropoff"}}</span><span class="value">{{.Dropoff}}</span></div>
<hr class="divider">
<p class="text-center" style="color:#F1F5F9;">{{t .Lang "thank_you"}}</p>`,

	"trip_completed": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "trip_summary"}} <span class="badge badge-green">✓</span></h2>
<div class="row"><span class="label">{{t .Lang "booking_number"}}</span><span class="value">{{.BookingNumber}}</span></div>
<div class="row"><span class="label">{{t .Lang "service"}}</span><span class="value">{{.ServiceType}}</span></div>
<div class="row"><span class="label">{{t .Lang "pickup"}}</span><span class="value">{{.Pickup}}</span></div>
<div class="row"><span class="label">{{t .Lang "dropoff"}}</span><span class="value">{{.Dropoff}}</span></div>
<div class="row"><span class="label">{{t .Lang "payment_method"}}</span><span class="value">{{.PaymentMethod}}</span></div>
<div class="highlight">
  <div class="text-muted mb-16">{{t .Lang "total"}}</div>
  <div class="amount">{{formatMoney .PriceCents .Currency}}</div>
</div>
<p class="text-center" style="color:#F1F5F9;">{{t .Lang "thank_you"}}</p>`,

	"refund_confirmation": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "refund_processed"}}</h2>
{{if .BookingNumber}}<div class="row"><span class="label">{{t .Lang "booking_number"}}</span><span class="value">{{.BookingNumber}}</span></div>{{end}}
{{if .PaymentRef}}<div class="row"><span class="label">Ref</span><span class="value">{{.PaymentRef}}</span></div>{{end}}
{{if .RefundReason}}<div class="row"><span class="label">{{t .Lang "cancel_reason"}}</span><span class="value">{{.RefundReason}}</span></div>{{end}}
<div class="highlight">
  <div class="text-muted mb-16">{{t .Lang "refund_amount"}}</div>
  <div class="amount">{{formatMoney .RefundAmount .Currency}}</div>
</div>`,

	"booking_cancellation": `
<h2 style="color:#F1F5F9; margin-top:0;">{{t .Lang "booking_cancelled"}}</h2>
<div class="row"><span class="label">{{t .Lang "booking_number"}}</span><span class="value">{{.BookingNumber}}</span></div>
{{if .CancelReason}}<div class="row"><span class="label">{{t .Lang "cancel_reason"}}</span><span class="value">{{.CancelReason}}</span></div>{{end}}
<div class="row"><span class="label">{{t .Lang "service"}}</span><span class="value">{{.ServiceType}}</span></div>
<div class="row"><span class="label">{{t .Lang "pickup"}}</span><span class="value">{{.Pickup}}</span></div>
<div class="row"><span class="label">{{t .Lang "dropoff"}}</span><span class="value">{{.Dropoff}}</span></div>
<hr class="divider">
<p class="text-center text-muted">{{t .Lang "support"}} <a href="mailto:{{.SupportEmail}}" style="color:#38BDF8;">{{.SupportEmail}}</a></p>`,
}
