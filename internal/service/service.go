package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/appraisal-crm/notification-service/internal/repository"
	"github.com/appraisal-crm/notification-service/internal/sender"
	"github.com/appraisal-crm/notification-service/internal/templates"
	"github.com/google/uuid"
)

type Service interface {
	DispatchEvent(ctx context.Context, env domain.EventEnvelope) error
	ListUserNotifications(ctx context.Context, recipientID uuid.UUID) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
}

type notificationService struct {
	repo            repository.NotificationRepository
	sender          sender.Sender
	renderer        *templates.Renderer
	clientPortalURL string
	officePortalURL string
	defaultLocale   templates.Locale
}

func NewNotificationService(
	repo repository.NotificationRepository,
	sender sender.Sender,
	renderer *templates.Renderer,
	clientPortalURL string,
	officePortalURL string,
) Service {
	return &notificationService{
		repo:            repo,
		sender:          sender,
		renderer:        renderer,
		clientPortalURL: clientPortalURL,
		officePortalURL: officePortalURL,
		defaultLocale:   templates.LocaleRU,
	}
}

// DispatchEvent routes an incoming Kafka domain event to the appropriate notification handler.
func (s *notificationService) DispatchEvent(ctx context.Context, env domain.EventEnvelope) error {
	switch env.EventType {
	case domain.EventTypeRequestCreated:
		return s.handleRequestCreated(ctx, env)
	case domain.EventTypeRequestStatusChanged:
		return s.handleRequestStatusChanged(ctx, env)
	case domain.EventTypeInspectCompleted:
		return s.handleInspectCompleted(ctx, env)
	case domain.EventTypeReportReady:
		return s.handleReportReady(ctx, env)
	default:
		slog.InfoContext(ctx, "ignoring unhandled event type", "event_type", env.EventType, "event_id", env.EventID)
		return nil
	}
}

func (s *notificationService) handleRequestCreated(ctx context.Context, env domain.EventEnvelope) error {
	var data domain.RequestCreatedData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return fmt.Errorf("unmarshal request.created payload: %w", err)
	}

	reqIDStr := env.RequestID.String()

	// 1. Client notification (Email)
	if data.Email != "" {
		subj, body, err := s.renderer.RenderRequestCreatedClient(
			s.defaultLocale,
			reqIDStr,
			data.Email,
			data.PhoneNumber,
			data.ObjectType,
			data.Address,
			s.clientPortalURL,
		)
		if err != nil {
			return fmt.Errorf("render request created client email: %w", err)
		}

		clientNotif := &domain.Notification{
			ID:               uuid.New(),
			RecipientID:      &data.ClientID,
			Channel:          domain.ChannelEmail,
			RecipientAddress: &data.Email,
			EventType:        env.EventType,
			RequestID:        &env.RequestID,
			Subject:          &subj,
			Body:             body,
			Status:           domain.StatusPending,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		if err := s.sendAndRecord(ctx, clientNotif); err != nil {
			slog.ErrorContext(ctx, "failed to send request created client notification", "error", err, "request_id", env.RequestID)
		}
	}

	return nil
}

func (s *notificationService) handleRequestStatusChanged(ctx context.Context, env domain.EventEnvelope) error {
	var data domain.RequestStatusChangedData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return fmt.Errorf("unmarshal request.status_changed payload: %w", err)
	}

	reqIDStr := env.RequestID.String()

	switch data.NewStatus {
	case "inspection_scheduled":
		// Notify client about inspection scheduled
		subj, body, err := s.renderer.RenderInspectionScheduledClient(s.defaultLocale, reqIDStr, s.clientPortalURL)
		if err == nil {
			notif := &domain.Notification{
				ID:        uuid.New(),
				Channel:   domain.ChannelInApp,
				EventType: env.EventType,
				RequestID: &env.RequestID,
				Subject:   &subj,
				Body:      body,
				Status:    domain.StatusPending,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			_ = s.sendAndRecord(ctx, notif)
		}

	case "appraisal":
		// Notify client that appraisal is underway
		subj, body, err := s.renderer.RenderAppraisalStarted(s.defaultLocale, reqIDStr, s.clientPortalURL)
		if err == nil {
			notif := &domain.Notification{
				ID:        uuid.New(),
				Channel:   domain.ChannelInApp,
				EventType: env.EventType,
				RequestID: &env.RequestID,
				Subject:   &subj,
				Body:      body,
				Status:    domain.StatusPending,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			_ = s.sendAndRecord(ctx, notif)
		}

	case "report_sent":
		// Notify client that report is ready for download
		subj, body, err := s.renderer.RenderReportReadyClient(s.defaultLocale, reqIDStr, s.clientPortalURL)
		if err == nil {
			notif := &domain.Notification{
				ID:        uuid.New(),
				Channel:   domain.ChannelInApp,
				EventType: env.EventType,
				RequestID: &env.RequestID,
				Subject:   &subj,
				Body:      body,
				Status:    domain.StatusPending,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			_ = s.sendAndRecord(ctx, notif)
		}

	case "closed":
		// Notify client that request is closed
		subj, body, err := s.renderer.RenderRequestClosed(s.defaultLocale, reqIDStr, s.clientPortalURL)
		if err == nil {
			notif := &domain.Notification{
				ID:        uuid.New(),
				Channel:   domain.ChannelInApp,
				EventType: env.EventType,
				RequestID: &env.RequestID,
				Subject:   &subj,
				Body:      body,
				Status:    domain.StatusPending,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			_ = s.sendAndRecord(ctx, notif)
		}
	}

	return nil
}

func (s *notificationService) handleInspectCompleted(ctx context.Context, env domain.EventEnvelope) error {
	var data domain.InspectCompletedData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return fmt.Errorf("unmarshal inspect.completed payload: %w", err)
	}

	reqIDStr := env.RequestID.String()
	subj, body, err := s.renderer.RenderInspectionCompleted(s.defaultLocale, reqIDStr, data.PhotoCount, s.clientPortalURL, true)
	if err != nil {
		return err
	}

	notif := &domain.Notification{
		ID:        uuid.New(),
		Channel:   domain.ChannelInApp,
		EventType: env.EventType,
		RequestID: &env.RequestID,
		Subject:   &subj,
		Body:      body,
		Status:    domain.StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return s.sendAndRecord(ctx, notif)
}

func (s *notificationService) handleReportReady(ctx context.Context, env domain.EventEnvelope) error {
	var data domain.ReportReadyData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return fmt.Errorf("unmarshal report.ready payload: %w", err)
	}

	reqIDStr := env.RequestID.String()
	subj, body, err := s.renderer.RenderReportReadyClient(s.defaultLocale, reqIDStr, s.clientPortalURL)
	if err != nil {
		return err
	}

	notif := &domain.Notification{
		ID:        uuid.New(),
		Channel:   domain.ChannelInApp,
		EventType: env.EventType,
		RequestID: &env.RequestID,
		Subject:   &subj,
		Body:      body,
		Status:    domain.StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return s.sendAndRecord(ctx, notif)
}

// sendAndRecord persists the notification in DB as pending, delivers it via sender, and marks it sent or failed.
func (s *notificationService) sendAndRecord(ctx context.Context, n *domain.Notification) error {
	if err := s.repo.Create(ctx, n); err != nil {
		return fmt.Errorf("create notification in db: %w", err)
	}

	if n.Channel == domain.ChannelEmail && n.RecipientAddress != nil {
		sendErr := s.sender.Send(ctx, n)
		now := time.Now().UTC()
		if sendErr != nil {
			slog.ErrorContext(ctx, "failed to send notification email",
				"notification_id", n.ID,
				"recipient", *n.RecipientAddress,
				"error", sendErr,
			)
			_ = s.repo.MarkFailed(ctx, n.ID, now)
			return sendErr
		}
		_ = s.repo.MarkSent(ctx, n.ID, now, now)
		return nil
	}

	// In-app notifications are instantly ready
	now := time.Now().UTC()
	_ = s.repo.MarkSent(ctx, n.ID, now, now)
	return nil
}

func (s *notificationService) ListUserNotifications(ctx context.Context, recipientID uuid.UUID) ([]*domain.Notification, error) {
	return s.repo.ListByRecipient(ctx, recipientID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if n.RecipientID != nil && *n.RecipientID != recipientID {
		return domain.ErrForbidden
	}

	now := time.Now().UTC()
	return s.repo.MarkRead(ctx, id, now, now)
}

func (s *notificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return s.repo.GetByID(ctx, id)
}
