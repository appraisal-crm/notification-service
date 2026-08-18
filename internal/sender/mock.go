package sender

import (
	"context"
	"sync"

	"github.com/appraisal-crm/notification-service/internal/domain"
)

// MockSender records sent notifications in memory for testing.
type MockSender struct {
	mu       sync.Mutex
	Sent     []*domain.Notification
	ErrToRet error
}

func NewMockSender() *MockSender {
	return &MockSender{
		Sent: make([]*domain.Notification, 0),
	}
}

func (m *MockSender) Send(ctx context.Context, n *domain.Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ErrToRet != nil {
		return m.ErrToRet
	}

	copied := *n
	m.Sent = append(m.Sent, &copied)
	return nil
}

func (m *MockSender) GetAllSent() []*domain.Notification {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]*domain.Notification, len(m.Sent))
	copy(result, m.Sent)
	return result
}

func (m *MockSender) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Sent = make([]*domain.Notification, 0)
	m.ErrToRet = nil
}
