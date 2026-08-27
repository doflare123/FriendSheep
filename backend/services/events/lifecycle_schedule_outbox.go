package events

import (
	"context"
	"errors"
	"time"
)

const (
	LifecycleScheduleSchemaVersion uint16 = 1

	LifecycleScheduleOperationUpsert = "schedule_upsert"
	LifecycleScheduleOperationCancel = "schedule_cancel"

	DefaultLifecycleScheduleEventsLimit = 100
	MaxLifecycleScheduleEventsLimit     = 500
)

var (
	errEventScheduleOutboxUnavailable = errors.New("хранилище lifecycle schedule outbox не настроено")
	ErrInvalidScheduleEventsCursor    = errors.New("некорректный cursor lifecycle schedule outbox")
	ErrInvalidScheduleEventsLimit     = errors.New("некорректный limit lifecycle schedule outbox")
)

type EventLifecycleScheduleOutboxInput struct {
	MessageID     string
	EventID       uint
	Operation     string
	StartTime     *time.Time
	EndTime       *time.Time
	OccurredAt    time.Time
	SchemaVersion uint16
}

type EventLifecycleScheduleOutboxWriter interface {
	AppendLifecycleScheduleEvent(input EventLifecycleScheduleOutboxInput) error
}

type EventLifecycleScheduleEvent struct {
	Sequence      int64      `json:"sequence"`
	MessageID     string     `json:"messageId"`
	SchemaVersion uint16     `json:"schemaVersion"`
	Operation     string     `json:"operation"`
	EventID       uint       `json:"eventId"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	OccurredAt    time.Time  `json:"occurredAt"`
}

type EventLifecycleScheduleEventsPage struct {
	Items      []EventLifecycleScheduleEvent `json:"items"`
	NextCursor int64                         `json:"nextCursor"`
	HasMore    bool                          `json:"hasMore"`
}

type EventLifecycleScheduleOutboxReader interface {
	ListLifecycleScheduleEvents(ctx context.Context, after int64, limit int) ([]EventLifecycleScheduleEvent, error)
}

type EventLifecycleScheduleOutboxService interface {
	List(ctx context.Context, after int64, limit int) (EventLifecycleScheduleEventsPage, error)
}

type eventLifecycleScheduleOutboxService struct {
	reader EventLifecycleScheduleOutboxReader
}

func NewEventLifecycleScheduleOutboxService(reader EventLifecycleScheduleOutboxReader) EventLifecycleScheduleOutboxService {
	return &eventLifecycleScheduleOutboxService{reader: reader}
}

func (s *eventLifecycleScheduleOutboxService) List(ctx context.Context, after int64, limit int) (EventLifecycleScheduleEventsPage, error) {
	if after < 0 {
		return EventLifecycleScheduleEventsPage{}, ErrInvalidScheduleEventsCursor
	}
	if limit == 0 {
		limit = DefaultLifecycleScheduleEventsLimit
	}
	if limit < 1 || limit > MaxLifecycleScheduleEventsLimit {
		return EventLifecycleScheduleEventsPage{}, ErrInvalidScheduleEventsLimit
	}
	if s.reader == nil {
		return EventLifecycleScheduleEventsPage{}, errEventScheduleOutboxUnavailable
	}

	items, err := s.reader.ListLifecycleScheduleEvents(ctx, after, limit+1)
	if err != nil {
		return EventLifecycleScheduleEventsPage{}, err
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
		items = []EventLifecycleScheduleEvent{}
	}
	return EventLifecycleScheduleEventsPage{Items: items, NextCursor: nextCursor, HasMore: hasMore}, nil
}
