package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Topic names
const (
	TopicRequestEvents = "request.events"
	TopicInspectEvents = "inspect.events"
	TopicReviewEvents  = "review.events"
)

// Event types
const (
	EventTypeRequestCreated       = "request.created"
	EventTypeRequestStatusChanged = "request.status_changed"
	EventTypeInspectCompleted     = "inspect.completed"
	EventTypeReportReady          = "report.ready"
)

// EventEnvelope is the JSON wire format for every domain event consumed.
type EventEnvelope struct {
	EventID    uuid.UUID       `json:"event_id"`
	EventType  string          `json:"event_type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	RequestID  uuid.UUID       `json:"request_id"`
	Data       json.RawMessage `json:"data"`
}

// RequestCreatedData is the payload for request.created.
type RequestCreatedData struct {
	ClientID    uuid.UUID `json:"client_id"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	ObjectType  *string   `json:"object_type,omitempty"`
	Address     *string   `json:"address,omitempty"`
	Status      string    `json:"status"`
}

// RequestStatusChangedData is the payload for request.status_changed.
type RequestStatusChangedData struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

// InspectCompletedData is the payload for inspect.completed.
type InspectCompletedData struct {
	InspectionID uuid.UUID  `json:"inspection_id"`
	InspectorID  *uuid.UUID `json:"inspector_id,omitempty"`
	CompletedAt  time.Time  `json:"completed_at"`
	PhotoCount   int        `json:"photo_count"`
}

// ReportReadyData is the payload for report.ready.
type ReportReadyData struct {
	AppraisalID uuid.UUID  `json:"appraisal_id"`
	AppraiserID *uuid.UUID `json:"appraiser_id,omitempty"`
	ReportS3Key *string    `json:"report_s3_key,omitempty"`
	CompletedAt time.Time  `json:"completed_at"`
}
