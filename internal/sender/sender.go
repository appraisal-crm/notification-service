package sender

import (
	"context"

	"github.com/appraisal-crm/notification-service/internal/domain"
)

// Sender delivers a notification to the recipient through a specific communication channel.
type Sender interface {
	Send(ctx context.Context, n *domain.Notification) error
}
