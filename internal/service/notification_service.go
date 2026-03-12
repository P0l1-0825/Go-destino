package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

// NotificationService handles push, SMS, email, and WhatsApp notifications.
// It integrates with SMTPService for email, TwilioService for SMS/WhatsApp,
// and FCM (placeholder) for push.
type NotificationService struct {
	notifRepo   *repository.NotificationRepository
	smtpSvc     *SMTPService
	twilioSvc   *TwilioService
	templateSvc *EmailTemplateService
	userRepo    *repository.UserRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

// NewNotificationServiceFull creates a NotificationService wired to real delivery services.
func NewNotificationServiceFull(
	notifRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	smtpSvc *SMTPService,
	twilioSvc *TwilioService,
	templateSvc *EmailTemplateService,
) *NotificationService {
	return &NotificationService{
		notifRepo:   notifRepo,
		smtpSvc:     smtpSvc,
		twilioSvc:   twilioSvc,
		templateSvc: templateSvc,
		userRepo:    userRepo,
	}
}

// Send dispatches a notification to the appropriate channel.
func (s *NotificationService) Send(ctx context.Context, tenantID string, req domain.SendNotificationRequest) (*domain.Notification, error) {
	notif := &domain.Notification{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		UserID:    req.UserID,
		Channel:   req.Channel,
		Status:    domain.NotifPending,
		Title:     req.Title,
		Body:      req.Body,
		Data:      req.Data,
		BookingID: req.BookingID,
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return nil, fmt.Errorf("creating notification: %w", err)
	}

	// Dispatch to channel (async — fire-and-forget)
	go func() {
		var err error
		switch req.Channel {
		case domain.ChannelPush:
			err = s.sendPush(req)
		case domain.ChannelSMS:
			err = s.sendSMS(req)
		case domain.ChannelEmail:
			err = s.sendEmail(req)
		case domain.ChannelWhatsApp:
			err = s.sendWhatsApp(req)
		}

		status := domain.NotifSent
		if err != nil {
			status = domain.NotifFailed
			log.Printf("notification %s failed: %v", notif.ID, err)
		}
		_ = s.notifRepo.UpdateStatus(context.Background(), notif.ID, status)
	}()

	return notif, nil
}

// --- High-level multi-channel notification helpers ---

// SendBookingConfirmation sends push + email + SMS/WhatsApp for a booking confirmation.
func (s *NotificationService) SendBookingConfirmation(ctx context.Context, tenantID, userID, bookingNumber, lang string) error {
	title, body := localizedBookingMsg(lang, "confirmed", bookingNumber)
	_, err := s.Send(ctx, tenantID, domain.SendNotificationRequest{
		UserID:  userID,
		Channel: domain.ChannelPush,
		Title:   title,
		Body:    body,
		Data:    map[string]string{"type": "booking_confirmed", "booking_number": bookingNumber},
	})
	return err
}

// SendBookingConfirmationFull sends multi-channel notifications with HTML email for a booking.
func (s *NotificationService) SendBookingConfirmationFull(ctx context.Context, tenantID string, booking *domain.Booking, lang string) {
	data := DefaultTemplateData()
	data.Lang = lang
	data.BookingNumber = booking.BookingNumber
	data.ServiceType = string(booking.ServiceType)
	data.Pickup = booking.PickupAddress
	data.Dropoff = booking.DropoffAddress
	data.Passengers = booking.PassengerCount
	data.PriceCents = booking.PriceCents
	data.Currency = booking.Currency
	data.FlightNumber = booking.FlightNumber
	if booking.ScheduledAt != nil {
		data.ScheduledAt = booking.ScheduledAt.Format("2006-01-02 15:04")
	}

	// Email
	s.sendEmailTemplate(ctx, tenantID, booking.UserID, func() (string, string, error) {
		return s.templateSvc.RenderBookingConfirmation(data)
	})

	// SMS
	smsBody := FormatBookingConfirmationSMS(lang, booking.BookingNumber, booking.PickupAddress, booking.DropoffAddress, booking.PriceCents, booking.Currency)
	s.sendSMSToUser(ctx, tenantID, booking.UserID, smsBody)

	// WhatsApp
	s.sendWhatsAppToUser(ctx, tenantID, booking.UserID, smsBody)

	// Push
	title, body := localizedBookingMsg(lang, "confirmed", booking.BookingNumber)
	_, _ = s.Send(ctx, tenantID, domain.SendNotificationRequest{
		UserID:  booking.UserID,
		Channel: domain.ChannelPush,
		Title:   title,
		Body:    body,
		Data:    map[string]string{"type": "booking_confirmed", "booking_number": booking.BookingNumber},
	})
}

// SendDriverAssignedFull sends multi-channel notifications when a driver is assigned.
func (s *NotificationService) SendDriverAssignedFull(ctx context.Context, tenantID string, booking *domain.Booking, driverName, vehiclePlate, lang string) {
	data := DefaultTemplateData()
	data.Lang = lang
	data.BookingNumber = booking.BookingNumber
	data.DriverName = driverName
	data.VehiclePlate = vehiclePlate
	data.Pickup = booking.PickupAddress
	data.Dropoff = booking.DropoffAddress

	// Email
	s.sendEmailTemplate(ctx, tenantID, booking.UserID, func() (string, string, error) {
		return s.templateSvc.RenderDriverAssigned(data)
	})

	// SMS + WhatsApp
	smsBody := FormatDriverAssignedSMS(lang, booking.BookingNumber, driverName, vehiclePlate)
	s.sendSMSToUser(ctx, tenantID, booking.UserID, smsBody)
	s.sendWhatsAppToUser(ctx, tenantID, booking.UserID, smsBody)

	// Push
	s.SendDriverAssigned(ctx, tenantID, booking.UserID, driverName, vehiclePlate, lang)
}

// SendDriverAssigned notifies tourist that a driver was assigned.
func (s *NotificationService) SendDriverAssigned(ctx context.Context, tenantID, userID, driverName, vehiclePlate, lang string) error {
	title := "Driver assigned"
	body := fmt.Sprintf("Your driver %s is on the way. Vehicle: %s", driverName, vehiclePlate)
	if lang == "es" {
		title = "Conductor asignado"
		body = fmt.Sprintf("Tu conductor %s va en camino. Vehículo: %s", driverName, vehiclePlate)
	}
	_, err := s.Send(ctx, tenantID, domain.SendNotificationRequest{
		UserID:  userID,
		Channel: domain.ChannelPush,
		Title:   title,
		Body:    body,
	})
	return err
}

// SendTripCompletedFull sends multi-channel notifications when a trip completes.
func (s *NotificationService) SendTripCompletedFull(ctx context.Context, tenantID string, booking *domain.Booking, paymentMethod, lang string) {
	data := DefaultTemplateData()
	data.Lang = lang
	data.BookingNumber = booking.BookingNumber
	data.ServiceType = string(booking.ServiceType)
	data.Pickup = booking.PickupAddress
	data.Dropoff = booking.DropoffAddress
	data.PriceCents = booking.PriceCents
	data.Currency = booking.Currency
	data.PaymentMethod = paymentMethod

	s.sendEmailTemplate(ctx, tenantID, booking.UserID, func() (string, string, error) {
		return s.templateSvc.RenderTripCompleted(data)
	})

	smsBody := FormatTripCompletedSMS(lang, booking.BookingNumber, booking.PriceCents, booking.Currency)
	s.sendSMSToUser(ctx, tenantID, booking.UserID, smsBody)
	s.sendWhatsAppToUser(ctx, tenantID, booking.UserID, smsBody)
}

// SendRefundNotification sends refund confirmation across all channels.
func (s *NotificationService) SendRefundNotification(ctx context.Context, tenantID, userID string, refundCents int64, currency, reference, reason, lang string) {
	data := DefaultTemplateData()
	data.Lang = lang
	data.RefundAmount = refundCents
	data.Currency = currency
	data.PaymentRef = reference
	data.RefundReason = reason

	s.sendEmailTemplate(ctx, tenantID, userID, func() (string, string, error) {
		return s.templateSvc.RenderRefundConfirmation(data)
	})

	smsBody := FormatRefundSMS(lang, refundCents, currency, reference)
	s.sendSMSToUser(ctx, tenantID, userID, smsBody)
	s.sendWhatsAppToUser(ctx, tenantID, userID, smsBody)
}

// SendCancellationNotification sends booking cancellation notification.
func (s *NotificationService) SendCancellationNotification(ctx context.Context, tenantID string, booking *domain.Booking, reason, lang string) {
	data := DefaultTemplateData()
	data.Lang = lang
	data.BookingNumber = booking.BookingNumber
	data.CancelReason = reason
	data.ServiceType = string(booking.ServiceType)
	data.Pickup = booking.PickupAddress
	data.Dropoff = booking.DropoffAddress

	s.sendEmailTemplate(ctx, tenantID, booking.UserID, func() (string, string, error) {
		return s.templateSvc.RenderBookingCancellation(data)
	})

	smsBody := FormatCancellationSMS(lang, booking.BookingNumber)
	s.sendSMSToUser(ctx, tenantID, booking.UserID, smsBody)
	s.sendWhatsAppToUser(ctx, tenantID, booking.UserID, smsBody)
}

// SendTicketPurchaseNotification sends ticket confirmation across all channels.
func (s *NotificationService) SendTicketPurchaseNotification(ctx context.Context, tenantID, userID string, tickets []domain.Ticket, totalCents int64, currency, paymentMethod, lang string) {
	if len(tickets) == 0 {
		return
	}

	ticketIDs := make([]string, len(tickets))
	for i, t := range tickets {
		ticketIDs[i] = t.ID
	}

	data := DefaultTemplateData()
	data.Lang = lang
	data.TicketIDs = ticketIDs
	data.PriceCents = totalCents
	data.Currency = currency
	data.PaymentMethod = paymentMethod
	data.QRCode = tickets[0].QRCode

	s.sendEmailTemplate(ctx, tenantID, userID, func() (string, string, error) {
		return s.templateSvc.RenderTicketPurchase(data)
	})

	smsBody := FormatTicketPurchaseSMS(lang, len(tickets), totalCents, currency, tickets[0].QRCode)
	s.sendSMSToUser(ctx, tenantID, userID, smsBody)
	s.sendWhatsAppToUser(ctx, tenantID, userID, smsBody)
}

// SendPaymentReceipt sends a payment receipt email.
func (s *NotificationService) SendPaymentReceipt(ctx context.Context, tenantID, userID string, payment *domain.Payment, bookingNumber, lang string) {
	data := DefaultTemplateData()
	data.Lang = lang
	data.ReceiptNumber = payment.Reference
	data.BookingNumber = bookingNumber
	data.PriceCents = payment.AmountCents
	data.Currency = payment.Currency
	data.PaymentMethod = string(payment.Method)
	data.PaymentRef = payment.Reference
	data.IssuedAt = payment.CreatedAt.Format("2006-01-02 15:04")

	s.sendEmailTemplate(ctx, tenantID, userID, func() (string, string, error) {
		return s.templateSvc.RenderPaymentReceipt(data)
	})
}

// SendSOSAlert sends emergency notifications to dispatch + authorities.
func (s *NotificationService) SendSOSAlert(ctx context.Context, tenantID, bookingID, country string) error {
	emergencyNum := domain.EmergencyNumbers[country]
	_, err := s.Send(ctx, tenantID, domain.SendNotificationRequest{
		Channel: domain.ChannelPush,
		Title:   "SOS EMERGENCY",
		Body:    fmt.Sprintf("Emergency alert triggered for booking. Local emergency: %s", emergencyNum),
		Data:    map[string]string{"type": "sos", "booking_id": bookingID, "emergency_number": emergencyNum},
	})
	return err
}

// GetUserNotifications retrieves notification history for a user.
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit int) ([]domain.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.notifRepo.ListByUser(ctx, userID, limit)
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, id string) error {
	return s.notifRepo.MarkRead(ctx, id)
}

// MarkAllRead marks all unread notifications for a user as read.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

// --- Internal delivery helpers ---

func (s *NotificationService) sendEmailTemplate(ctx context.Context, tenantID, userID string, renderFn func() (string, string, error)) {
	if s.templateSvc == nil || s.smtpSvc == nil {
		return
	}

	go func() {
		subject, html, err := renderFn()
		if err != nil {
			log.Printf("[EMAIL] template render error: %v", err)
			return
		}

		email := s.resolveUserEmail(userID)
		if email == "" {
			return
		}

		if err := s.smtpSvc.SendEmail(email, subject, html); err != nil {
			log.Printf("[EMAIL] delivery error: %v", err)
			_, _ = s.Send(context.Background(), tenantID, domain.SendNotificationRequest{
				UserID:  userID,
				Channel: domain.ChannelEmail,
				Title:   subject,
				Body:    "delivery failed",
			})
		}
	}()
}

func (s *NotificationService) sendSMSToUser(ctx context.Context, tenantID, userID, body string) {
	if s.twilioSvc == nil {
		return
	}

	go func() {
		phone := s.resolveUserPhone(userID)
		if phone == "" {
			return
		}
		if err := s.twilioSvc.SendSMS(phone, body); err != nil {
			log.Printf("[SMS] delivery error to %s: %v", userID, err)
		}
		_, _ = s.Send(context.Background(), tenantID, domain.SendNotificationRequest{
			UserID:  userID,
			Channel: domain.ChannelSMS,
			Title:   "GoDestino",
			Body:    body,
		})
	}()
}

func (s *NotificationService) sendWhatsAppToUser(ctx context.Context, tenantID, userID, body string) {
	if s.twilioSvc == nil {
		return
	}

	go func() {
		phone := s.resolveUserPhone(userID)
		if phone == "" {
			return
		}
		if err := s.twilioSvc.SendWhatsApp(phone, body); err != nil {
			log.Printf("[WHATSAPP] delivery error to %s: %v", userID, err)
		}
		_, _ = s.Send(context.Background(), tenantID, domain.SendNotificationRequest{
			UserID:  userID,
			Channel: domain.ChannelWhatsApp,
			Title:   "GoDestino",
			Body:    body,
		})
	}()
}

func (s *NotificationService) resolveUserEmail(userID string) string {
	if s.userRepo == nil || userID == "" {
		return ""
	}
	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		return ""
	}
	return user.Email
}

func (s *NotificationService) resolveUserPhone(userID string) string {
	if s.userRepo == nil || userID == "" {
		return ""
	}
	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil || user.Phone == "" {
		return ""
	}
	return user.Phone
}

func (s *NotificationService) sendPush(req domain.SendNotificationRequest) error {
	log.Printf("[PUSH] → %s: %s - %s", req.UserID, req.Title, req.Body)
	return nil // FCM integration placeholder
}

func (s *NotificationService) sendSMS(req domain.SendNotificationRequest) error {
	phone := s.resolveUserPhone(req.UserID)
	if phone == "" {
		log.Printf("[SMS] → %s: %s (no phone)", req.UserID, req.Body)
		return nil
	}
	if s.twilioSvc != nil {
		return s.twilioSvc.SendSMS(phone, req.Body)
	}
	log.Printf("[SMS] → %s: %s", phone, req.Body)
	return nil
}

func (s *NotificationService) sendEmail(req domain.SendNotificationRequest) error {
	email := s.resolveUserEmail(req.UserID)
	if email == "" {
		log.Printf("[EMAIL] → %s: %s (no email)", req.UserID, req.Title)
		return nil
	}
	if s.smtpSvc != nil {
		return s.smtpSvc.SendEmail(email, req.Title, req.Body)
	}
	log.Printf("[EMAIL] → %s: %s", email, req.Title)
	return nil
}

func (s *NotificationService) sendWhatsApp(req domain.SendNotificationRequest) error {
	phone := s.resolveUserPhone(req.UserID)
	if phone == "" {
		log.Printf("[WHATSAPP] → %s: %s (no phone)", req.UserID, req.Body)
		return nil
	}
	if s.twilioSvc != nil {
		return s.twilioSvc.SendWhatsApp(phone, req.Body)
	}
	log.Printf("[WHATSAPP] → %s: %s", phone, req.Body)
	return nil
}

func localizedBookingMsg(lang, event, bookingNumber string) (string, string) {
	switch {
	case event == "confirmed" && strings.HasPrefix(lang, "es"):
		return "Reserva confirmada", fmt.Sprintf("Tu reserva %s ha sido confirmada. Tu conductor será asignado pronto.", bookingNumber)
	case event == "confirmed" && strings.HasPrefix(lang, "pt"):
		return "Reserva confirmada", fmt.Sprintf("Sua reserva %s foi confirmada. Seu motorista será designado em breve.", bookingNumber)
	case event == "confirmed":
		return "Booking confirmed", fmt.Sprintf("Your booking %s has been confirmed. Your driver will be assigned shortly.", bookingNumber)
	default:
		return "GoDestino", bookingNumber
	}
}
