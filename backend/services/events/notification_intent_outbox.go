package events

import (
	"context"
	"errors"
	"friendship/services/notifications"
	"time"
)

const (
	EventReminderIntentSchemaVersion uint16 = 1
	EventReminderIntentType                 = "event_reminder"

	EventReminderOperationUpsert = "schedule_upsert"
	EventReminderOperationCancel = "schedule_cancel"

	EventReminderOffset24Hours = notifications.EventReminderOffset24Hours
	EventReminderOffset6Hours  = notifications.EventReminderOffset6Hours
	EventReminderOffset1Hour   = notifications.EventReminderOffset1Hour

	DefaultEventReminderIntentLimit = 100
	MaxEventReminderIntentLimit     = 500
)

var (
	errEventReminderIntentOutboxUnavailable = errors.New("хранилище notification intent outbox не настроено")
	ErrInvalidEventReminderIntentCursor     = errors.New("некорректный cursor notification intent outbox")
	ErrInvalidEventReminderIntentLimit      = errors.New("некорректный limit notification intent outbox")
)

type EventReminderIntentOutboxInput struct {
	MessageID             string
	SchemaVersion         uint16
	IntentType            string
	Operation             string
	EventID               uint
	StartTime             *time.Time
	ReminderOffsetMinutes []int
	OccurredAt            time.Time
}

type EventReminderIntentOutboxWriter interface {
	AppendEventReminderIntent(EventReminderIntentOutboxInput) error
}

type EventReminderIntent struct {
	Sequence              int64      `json:"sequence"`
	MessageID             string     `json:"messageId"`
	SchemaVersion         uint16     `json:"schemaVersion"`
	IntentType            string     `json:"intentType"`
	Operation             string     `json:"operation"`
	EventID               uint       `json:"eventId"`
	StartTime             *time.Time `json:"startTime,omitempty"`
	ReminderOffsetMinutes []int      `json:"reminderOffsetMinutes,omitempty"`
	OccurredAt            time.Time  `json:"occurredAt"`
}

type EventReminderIntentPage struct {
	Items      []EventReminderIntent `json:"items"`
	NextCursor int64                 `json:"nextCursor"`
	HasMore    bool                  `json:"hasMore"`
}

type EventReminderIntentOutboxReader interface {
	ListEventReminderIntents(ctx context.Context, after int64, limit int) ([]EventReminderIntent, error)
}

type EventReminderIntentService interface {
	List(ctx context.Context, after int64, limit int) (EventReminderIntentPage, error)
}

type eventReminderIntentService struct {
	reader EventReminderIntentOutboxReader
}

func NewEventReminderIntentService(reader EventReminderIntentOutboxReader) EventReminderIntentService {
	return &eventReminderIntentService{reader: reader}
}

func (s *eventReminderIntentService) List(ctx context.Context, after int64, limit int) (EventReminderIntentPage, error) {
	if after < 0 {
		return EventReminderIntentPage{}, ErrInvalidEventReminderIntentCursor
	}
	if limit == 0 {
		limit = DefaultEventReminderIntentLimit
	}
	if limit < 1 || limit > MaxEventReminderIntentLimit {
		return EventReminderIntentPage{}, ErrInvalidEventReminderIntentLimit
	}
	if s.reader == nil {
		return EventReminderIntentPage{}, errEventReminderIntentOutboxUnavailable
	}

	items, err := s.reader.ListEventReminderIntents(ctx, after, limit+1)
	if err != nil {
		return EventReminderIntentPage{}, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	nextCursor := after
	if len(items) > 0 {
		nextCursor = items[len(items)-1].Sequence
	}
	if items == nil {
		items = []EventReminderIntent{}
	}
	return EventReminderIntentPage{Items: items, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func cloneReminderOffsets(source []int) []int {
	if source == nil {
		return nil
	}
	return append([]int(nil), source...)
}
