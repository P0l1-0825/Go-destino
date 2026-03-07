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
type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
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

	// Dispatch to channel (async in production via queue)
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

// SendBookingConfirmation sends a localized booking confirmation.
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

func (s *NotificationService) sendPush(req domain.SendNotificationRequest) error {
	log.Printf("[PUSH] → %s: %s - %s", req.UserID, req.Title, req.Body)
	return nil // FCM integration placeholder
}

func (s *NotificationService) sendSMS(req domain.SendNotificationRequest) error {
	log.Printf("[SMS] → %s: %s", req.UserID, req.Body)
	return nil // Twilio integration placeholder
}

func (s *NotificationService) sendEmail(req domain.SendNotificationRequest) error {
	log.Printf("[EMAIL] → %s: %s", req.UserID, req.Title)
	return nil // SES integration placeholder
}

func (s *NotificationService) sendWhatsApp(req domain.SendNotificationRequest) error {
	log.Printf("[WHATSAPP] → %s: %s", req.UserID, req.Body)
	return nil // WhatsApp Business API placeholder
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
