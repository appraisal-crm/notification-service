package domain

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelInApp Channel = "in_app"
	ChannelPush  Channel = "push"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

// Notification is one message queued for delivery to a recipient over a single
// channel. It is created from a consumed event and dispatched by a sender.
type Notification struct {
	ID               uuid.UUID  `json:"id"`
	RecipientID      *uuid.UUID `json:"recipient_id,omitempty"`
	Channel          Channel    `json:"channel"`
	RecipientAddress *string    `json:"recipient_address,omitempty"`
	EventType        string     `json:"event_type"`
	RequestID        *uuid.UUID `json:"request_id,omitempty"`
	Subject          *string    `json:"subject,omitempty"`
	Body             string     `json:"body"`
	Status           Status     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	SentAt           *time.Time `json:"sent_at,omitempty"`
	ReadAt           *time.Time `json:"read_at,omitempty"`
}
