package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/appraisal-crm/notification-service/internal/sender"
	"github.com/appraisal-crm/notification-service/internal/templates"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	mu            sync.Mutex
	notifications map[uuid.UUID]*domain.Notification
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		notifications: make(map[uuid.UUID]*domain.Notification),
	}
}

func (m *mockRepository) Create(ctx context.Context, n *domain.Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := *n
	m.notifications[n.ID] = &copied
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notifications[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return n, nil
}

func (m *mockRepository) ListByRecipient(ctx context.Context, recipientID uuid.UUID) ([]*domain.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*domain.Notification
	for _, n := range m.notifications {
		if n.RecipientID != nil && *n.RecipientID == recipientID {
			list = append(list, n)
		}
	}
	return list, nil
}

func (m *mockRepository) MarkSent(ctx context.Context, id uuid.UUID, sentAt, updatedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.Status = domain.StatusSent
	n.SentAt = &sentAt
	n.UpdatedAt = updatedAt
	return nil
}

func (m *mockRepository) MarkFailed(ctx context.Context, id uuid.UUID, updatedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.Status = domain.StatusFailed
	n.UpdatedAt = updatedAt
	return nil
}

func (m *mockRepository) MarkRead(ctx context.Context, id uuid.UUID, readAt, updatedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.ReadAt = &readAt
	n.UpdatedAt = updatedAt
	return nil
}

func setupTestService(t *testing.T) (Service, *mockRepository, *sender.MockSender) {
	repo := newMockRepository()
	mockSender := sender.NewMockSender()
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	svc := NewNotificationService(repo, mockSender, renderer, "http://localhost:5173", "http://localhost:5174")
	return svc, repo, mockSender
}

func TestService_HandleRequestCreated(t *testing.T) {
	svc, repo, mockSender := setupTestService(t)
	ctx := context.Background()

	clientID := uuid.New()
	reqID := uuid.New()
	objType := "apartment"
	address := "ул. Ленина, 1"

	dataBytes, err := json.Marshal(domain.RequestCreatedData{
		ClientID:    clientID,
		Email:       "client@test.com",
		PhoneNumber: "+79991234567",
		ObjectType:  &objType,
		Address:     &address,
		Status:      "new",
	})
	require.NoError(t, err)

	env := domain.EventEnvelope{
		EventID:    uuid.New(),
		EventType:  domain.EventTypeRequestCreated,
		Version:    1,
		OccurredAt: time.Now().UTC(),
		RequestID:  reqID,
		Data:       dataBytes,
	}

	err = svc.DispatchEvent(ctx, env)
	require.NoError(t, err)

	// Check mock sender received email
	sentEmails := mockSender.GetAllSent()
	require.Len(t, sentEmails, 1)
	assert.Equal(t, "client@test.com", *sentEmails[0].RecipientAddress)
	assert.Contains(t, *sentEmails[0].Subject, "принята")

	// Check DB record
	list, err := repo.ListByRecipient(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, domain.StatusSent, list[0].Status)
	assert.NotNil(t, list[0].SentAt)
}

func TestService_HandleRequestCreated_SenderFailure(t *testing.T) {
	svc, repo, mockSender := setupTestService(t)
	mockSender.ErrToRet = errors.New("smtp connection refused")
	ctx := context.Background()

	clientID := uuid.New()
	reqID := uuid.New()

	dataBytes, _ := json.Marshal(domain.RequestCreatedData{
		ClientID: clientID,
		Email:    "fail@test.com",
	})

	env := domain.EventEnvelope{
		EventID:   uuid.New(),
		EventType: domain.EventTypeRequestCreated,
		RequestID: reqID,
		Data:      dataBytes,
	}

	err := svc.DispatchEvent(ctx, env)
	require.NoError(t, err) // Dispatch should log error and not crash

	list, _ := repo.ListByRecipient(ctx, clientID)
	require.Len(t, list, 1)
	assert.Equal(t, domain.StatusFailed, list[0].Status)
}

func TestService_HandleStatusChanged_InspectionScheduled(t *testing.T) {
	svc, repo, _ := setupTestService(t)
	ctx := context.Background()

	reqID := uuid.New()
	dataBytes, _ := json.Marshal(domain.RequestStatusChangedData{
		OldStatus: "in_progress",
		NewStatus: "inspection_scheduled",
	})

	env := domain.EventEnvelope{
		EventID:   uuid.New(),
		EventType: domain.EventTypeRequestStatusChanged,
		RequestID: reqID,
		Data:      dataBytes,
	}

	err := svc.DispatchEvent(ctx, env)
	require.NoError(t, err)

	assert.Len(t, repo.notifications, 1)
}

func TestService_MarkAsRead(t *testing.T) {
	svc, repo, _ := setupTestService(t)
	ctx := context.Background()

	userA := uuid.New()
	userB := uuid.New()
	notifID := uuid.New()

	n := &domain.Notification{
		ID:          notifID,
		RecipientID: &userA,
		Channel:     domain.ChannelInApp,
		EventType:   domain.EventTypeRequestCreated,
		Body:        "Test",
		Status:      domain.StatusSent,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, n))

	// User B should be forbidden
	err := svc.MarkAsRead(ctx, notifID, userB)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// User A should succeed
	err = svc.MarkAsRead(ctx, notifID, userA)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, notifID)
	require.NoError(t, err)
	assert.NotNil(t, updated.ReadAt)
}
