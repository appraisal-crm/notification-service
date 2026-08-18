package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/appraisal-crm/notification-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	notifications map[uuid.UUID]*domain.Notification
}

func newMockService() *mockService {
	return &mockService{
		notifications: make(map[uuid.UUID]*domain.Notification),
	}
}

func (m *mockService) DispatchEvent(ctx context.Context, env domain.EventEnvelope) error {
	return nil
}

func (m *mockService) ListUserNotifications(ctx context.Context, recipientID uuid.UUID) ([]*domain.Notification, error) {
	var list []*domain.Notification
	for _, n := range m.notifications {
		if n.RecipientID != nil && *n.RecipientID == recipientID {
			list = append(list, n)
		}
	}
	return list, nil
}

func (m *mockService) MarkAsRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error {
	n, ok := m.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	if n.RecipientID != nil && *n.RecipientID != recipientID {
		return domain.ErrForbidden
	}
	now := time.Now().UTC()
	n.ReadAt = &now
	return nil
}

func (m *mockService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	n, ok := m.notifications[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return n, nil
}

func TestHandler_List(t *testing.T) {
	mockSvc := newMockService()
	h := NewHandler(mockSvc)

	userID := uuid.New()
	notifID := uuid.New()
	mockSvc.notifications[notifID] = &domain.Notification{
		ID:          notifID,
		RecipientID: &userID,
		Channel:     domain.ChannelInApp,
		EventType:   domain.EventTypeRequestCreated,
		Body:        "Test notification",
		Status:      domain.StatusSent,
	}

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	ctx := middleware.ContextWithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var list []*domain.Notification
	err := json.NewDecoder(rec.Body).Decode(&list)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, notifID, list[0].ID)
}

func TestHandler_MarkRead(t *testing.T) {
	mockSvc := newMockService()
	h := NewHandler(mockSvc)

	userID := uuid.New()
	notifID := uuid.New()
	mockSvc.notifications[notifID] = &domain.Notification{
		ID:          notifID,
		RecipientID: &userID,
		Channel:     domain.ChannelInApp,
		EventType:   domain.EventTypeRequestCreated,
		Body:        "Test notification",
		Status:      domain.StatusSent,
	}

	r := chi.NewRouter()
	r.Patch("/notifications/{id}/read", h.MarkRead)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/"+notifID.String()+"/read", nil)
	ctx := middleware.ContextWithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotNil(t, mockSvc.notifications[notifID].ReadAt)
}
