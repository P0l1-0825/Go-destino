package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `INSERT INTO notifications (id, tenant_id, user_id, channel, status, title, body, booking_id, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())`
	_, err := r.db.ExecContext(ctx, query, n.ID, n.TenantID, n.UserID, n.Channel, n.Status, n.Title, n.Body, n.BookingID)
	return err
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET status=$1, sent_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, channel, status, title, body, booking_id, created_at, sent_at
		FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Channel, &n.Status, &n.Title, &n.Body, &n.BookingID, &n.CreatedAt, &n.SentAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, rows.Err()
}
