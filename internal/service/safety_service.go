package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/google/uuid"
)

type SafetyService struct {
	db       *sql.DB
	notifSvc *NotificationService
}

func NewSafetyService(db *sql.DB, notifSvc *NotificationService) *SafetyService {
	return &SafetyService{db: db, notifSvc: notifSvc}
}

func (s *SafetyService) ReportIncident(ctx context.Context, incident *domain.SafetyIncident) (*domain.SafetyIncident, error) {
	incident.ID = uuid.New().String()
	incident.Status = "reported"
	incident.CreatedAt = time.Now()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO safety_incidents (id, tenant_id, booking_id, reporter_id, incident_type, severity, description, lat, lng, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		incident.ID, incident.TenantID, incident.BookingID, incident.ReportedBy,
		incident.Type, incident.Severity, incident.Description,
		incident.Lat, incident.Lng, incident.Status, incident.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Async notify operations team for high/critical severity
	if incident.Severity == domain.SeverityHigh || incident.Severity == domain.SeverityCritical {
		go s.notifSvc.Send(context.Background(), incident.TenantID, domain.SendNotificationRequest{
			UserID:  "ops-team",
			Channel: domain.ChannelPush,
			Title:   "Safety Incident: " + string(incident.Severity),
			Body:    incident.Description,
		})
	}

	return incident, nil
}

func (s *SafetyService) TriggerSOS(ctx context.Context, tenantID, userID, bookingID string, lat, lng float64) (*domain.SOSAlert, error) {
	alert := &domain.SOSAlert{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		UserID:    userID,
		BookingID: bookingID,
		Lat:       lat,
		Lng:       lng,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sos_alerts (id, tenant_id, user_id, booking_id, lat, lng, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		alert.ID, alert.TenantID, alert.UserID, alert.BookingID,
		alert.Lat, alert.Lng, alert.Status, alert.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Notify emergency services asynchronously
	go s.notifSvc.SendSOSAlert(context.Background(), tenantID, bookingID, "MX")

	return alert, nil
}

func (s *SafetyService) ResolveSOS(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sos_alerts SET status = 'resolved', resolved_at = NOW() WHERE id = $1`, id)
	return err
}

func (s *SafetyService) ListIncidents(ctx context.Context, tenantID string, limit, offset int) ([]domain.SafetyIncident, error) {
	query := `SELECT id, tenant_id, booking_id, reporter_id, incident_type, severity, description, lat, lng, status, created_at
		FROM safety_incidents WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []domain.SafetyIncident
	for rows.Next() {
		var i domain.SafetyIncident
		if err := rows.Scan(&i.ID, &i.TenantID, &i.BookingID, &i.ReportedBy, &i.Type, &i.Severity, &i.Description, &i.Lat, &i.Lng, &i.Status, &i.CreatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	return incidents, rows.Err()
}

func (s *SafetyService) GetIncident(ctx context.Context, id string) (*domain.SafetyIncident, error) {
	i := &domain.SafetyIncident{}
	query := `SELECT id, tenant_id, booking_id, reporter_id, incident_type, severity, description, lat, lng, status, created_at
		FROM safety_incidents WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&i.ID, &i.TenantID, &i.BookingID, &i.ReportedBy, &i.Type, &i.Severity, &i.Description, &i.Lat, &i.Lng, &i.Status, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return i, nil
}
