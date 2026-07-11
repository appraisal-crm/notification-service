package repository

import (
	"context"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByRecipient(ctx context.Context, recipientID uuid.UUID) ([]*domain.Notification, error)
	MarkSent(ctx context.Context, id uuid.UUID, sentAt, updatedAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID, updatedAt time.Time) error
	MarkRead(ctx context.Context, id uuid.UUID, readAt, updatedAt time.Time) error
}
