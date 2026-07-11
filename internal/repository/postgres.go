package repository

import (
	"context"
	"errors"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) NotificationRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, recipient_id, channel, recipient_address, event_type, request_id, subject, body, status, created_at, updated_at, sent_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		n.ID,
		n.RecipientID,
		n.Channel,
		n.RecipientAddress,
		n.EventType,
		n.RequestID,
		n.Subject,
		n.Body,
		n.Status,
		n.CreatedAt,
		n.UpdatedAt,
		n.SentAt,
		n.ReadAt,
	)
	return err
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT id, recipient_id, channel, recipient_address, event_type, request_id, subject, body, status, created_at, updated_at, sent_at, read_at
		FROM notifications
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	n, err := scanNotification(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return n, nil
}

func (r *postgresRepository) ListByRecipient(ctx context.Context, recipientID uuid.UUID) ([]*domain.Notification, error) {
	query := `
		SELECT id, recipient_id, channel, recipient_address, event_type, request_id, subject, body, status, created_at, updated_at, sent_at, read_at
		FROM notifications
		WHERE recipient_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, recipientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *postgresRepository) MarkSent(ctx context.Context, id uuid.UUID, sentAt, updatedAt time.Time) error {
	query := `
		UPDATE notifications SET status = $1, sent_at = $2, updated_at = $3
		WHERE id = $4
	`
	return r.exec(ctx, query, domain.StatusSent, sentAt, updatedAt, id)
}

func (r *postgresRepository) MarkFailed(ctx context.Context, id uuid.UUID, updatedAt time.Time) error {
	query := `
		UPDATE notifications SET status = $1, updated_at = $2
		WHERE id = $3
	`
	return r.exec(ctx, query, domain.StatusFailed, updatedAt, id)
}

func (r *postgresRepository) MarkRead(ctx context.Context, id uuid.UUID, readAt, updatedAt time.Time) error {
	query := `
		UPDATE notifications SET read_at = $1, updated_at = $2
		WHERE id = $3
	`
	return r.exec(ctx, query, readAt, updatedAt, id)
}

// exec runs an UPDATE and maps "no row matched" to ErrNotFound.
func (r *postgresRepository) exec(ctx context.Context, query string, args ...any) error {
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(row scanner) (*domain.Notification, error) {
	var n domain.Notification
	err := row.Scan(
		&n.ID,
		&n.RecipientID,
		&n.Channel,
		&n.RecipientAddress,
		&n.EventType,
		&n.RequestID,
		&n.Subject,
		&n.Body,
		&n.Status,
		&n.CreatedAt,
		&n.UpdatedAt,
		&n.SentAt,
		&n.ReadAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
}
