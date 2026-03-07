package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type TransportCardRepository struct {
	db *sql.DB
}

func NewTransportCardRepository(db *sql.DB) *TransportCardRepository {
	return &TransportCardRepository{db: db}
}

func (r *TransportCardRepository) Create(ctx context.Context, card *domain.TransportCard) error {
	query := `INSERT INTO transport_cards (id, tenant_id, card_number, balance_cents, currency, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query,
		card.ID, card.TenantID, card.CardNumber, card.BalanceCents, card.Currency, card.Active,
	)
	return err
}

func (r *TransportCardRepository) GetByID(ctx context.Context, id string) (*domain.TransportCard, error) {
	card := &domain.TransportCard{}
	query := `SELECT id, tenant_id, card_number, balance_cents, currency, active, created_at, updated_at
		FROM transport_cards WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&card.ID, &card.TenantID, &card.CardNumber, &card.BalanceCents,
		&card.Currency, &card.Active, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (r *TransportCardRepository) GetByNumber(ctx context.Context, tenantID, cardNumber string) (*domain.TransportCard, error) {
	card := &domain.TransportCard{}
	query := `SELECT id, tenant_id, card_number, balance_cents, currency, active, created_at, updated_at
		FROM transport_cards WHERE tenant_id = $1 AND card_number = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, cardNumber).Scan(
		&card.ID, &card.TenantID, &card.CardNumber, &card.BalanceCents,
		&card.Currency, &card.Active, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (r *TransportCardRepository) AddBalance(ctx context.Context, id string, amountCents int64) error {
	query := `UPDATE transport_cards SET balance_cents = balance_cents + $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, amountCents, id)
	return err
}

func (r *TransportCardRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE transport_cards SET active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *TransportCardRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.TransportCard, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, tenant_id, card_number, balance_cents, currency, active, created_at, updated_at
		FROM transport_cards WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []domain.TransportCard
	for rows.Next() {
		var c domain.TransportCard
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.CardNumber, &c.BalanceCents,
			&c.Currency, &c.Active, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}
