package sender

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockSender(t *testing.T) {
	ctx := context.Background()
	mock := NewMockSender()

	recipientEmail := "client@example.com"
	subject := "Test Subject"
	n := &domain.Notification{
		ID:               uuid.New(),
		Channel:          domain.ChannelEmail,
		RecipientAddress: &recipientEmail,
		EventType:        "request.created",
		Subject:          &subject,
		Body:             "<p>Hello</p>",
		Status:           domain.StatusPending,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	err := mock.Send(ctx, n)
	require.NoError(t, err)

	sent := mock.GetAllSent()
	require.Len(t, sent, 1)
	assert.Equal(t, n.ID, sent[0].ID)
	assert.Equal(t, "client@example.com", *sent[0].RecipientAddress)

	// Test error scenario
	mock.ErrToRet = errors.New("network error")
	err = mock.Send(ctx, n)
	assert.Error(t, err)

	mock.Reset()
	assert.Empty(t, mock.GetAllSent())
}

func TestSMTPSender_Disabled(t *testing.T) {
	ctx := context.Background()
	sender := NewSMTPSender("localhost", 1025, "", "", "noreply@appraisal-crm.ru", false)

	recipientEmail := "client@example.com"
	subject := "Test"
	n := &domain.Notification{
		ID:               uuid.New(),
		Channel:          domain.ChannelEmail,
		RecipientAddress: &recipientEmail,
		Subject:          &subject,
		Body:             "<p>Hello</p>",
	}

	err := sender.Send(ctx, n)
	require.NoError(t, err) // Should succeed without sending when disabled
}

func TestSMTPSender_EmptyRecipient(t *testing.T) {
	ctx := context.Background()
	sender := NewSMTPSender("localhost", 1025, "", "", "noreply@appraisal-crm.ru", true)

	n := &domain.Notification{
		ID:      uuid.New(),
		Channel: domain.ChannelEmail,
	}

	err := sender.Send(ctx, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient address is empty")
}
