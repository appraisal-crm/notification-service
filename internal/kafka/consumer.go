package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/appraisal-crm/notification-service/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// Deduplicator checks and marks event IDs for idempotent consumption.
type Deduplicator interface {
	Seen(ctx context.Context, eventID string) (bool, error)
	Forget(ctx context.Context, eventID string) error
}

// EventDispatcher processes consumed domain events.
type EventDispatcher interface {
	DispatchEvent(ctx context.Context, env domain.EventEnvelope) error
}

// Consumer consumes messages from a Kafka topic and routes them to the dispatcher.
type Consumer struct {
	reader *kafka.Reader
	dedup  Deduplicator
	svc    EventDispatcher
	topic  string
}

func NewConsumer(brokers []string, groupID, topic string, dedup Deduplicator, svc EventDispatcher) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		dedup: dedup,
		svc:   svc,
		topic: topic,
	}
}

// Run consumes messages in a blocking loop until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // Shutting down gracefully
			}
			return err
		}

		if err := c.process(ctx, m); err != nil {
			slog.ErrorContext(ctx, "consumer stopping after processing error", "error", err, "topic", c.topic)
			return err
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}

func (c *Consumer) process(ctx context.Context, m kafka.Message) error {
	var env domain.EventEnvelope
	if err := json.Unmarshal(m.Value, &env); err != nil {
		slog.WarnContext(ctx, "skipping malformed event", "error", err, "topic", c.topic, "offset", m.Offset)
		return nil
	}

	if env.EventID == uuid.Nil {
		slog.WarnContext(ctx, "skipping event without event_id", "topic", c.topic, "offset", m.Offset)
		return nil
	}

	dup, err := c.dedup.Seen(ctx, env.EventID.String())
	if err != nil {
		return fmt.Errorf("dedup check: %w", err)
	}
	if dup {
		slog.InfoContext(ctx, "duplicate event, skipping", "event_id", env.EventID, "event_type", env.EventType)
		return nil
	}

	if err := c.svc.DispatchEvent(ctx, env); err != nil {
		// Undo dedup mark so redelivery will re-process this event
		if ferr := c.dedup.Forget(ctx, env.EventID.String()); ferr != nil {
			slog.ErrorContext(ctx, "failed to forget dedup key", "error", ferr, "event_id", env.EventID)
		}
		return fmt.Errorf("dispatch event %s: %w", env.EventType, err)
	}

	slog.InfoContext(ctx, "processed Kafka domain event", "event_type", env.EventType, "event_id", env.EventID, "request_id", env.RequestID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
