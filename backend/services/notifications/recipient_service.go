package notifications

import (
	"context"
	"errors"
	"time"
)

const ChannelInApp = "in_app"

const (
	EventReminderOffset24Hours = 1440
	EventReminderOffset6Hours  = 360
	EventReminderOffset1Hour   = 60
)

func DefaultEventReminderOffsetsMinutes() []int {
	return []int{EventReminderOffset24Hours, EventReminderOffset6Hours, EventReminderOffset1Hour}
}

var (
	ErrEventNotFound               = errors.New("мероприятие не найдено")
	ErrInvalidReminderOffset       = errors.New("некорректный интервал напоминания")
	ErrRecipientStoreUnavailable   = errors.New("хранилище получателей напоминания не настроено")
	ErrReminderPreferenceReaderNil = errors.New("политика напоминаний не настроена")
)

type EventReminderSnapshot struct {
	EventID   uint
	Title     string
	StartTime time.Time
	UserIDs   []uint
}

type EventReminderRecipientStore interface {
	LoadEventReminderSnapshot(ctx context.Context, eventID uint) (EventReminderSnapshot, error)
}

type EventReminderPreferenceDecision struct {
	Enabled  bool
	Channels []string
}

// EventReminderPreferenceReader — порт для будущих пользовательских настроек.
// Замена его адаптера не требует менять scheduler или контракт inbox.
type EventReminderPreferenceReader interface {
	ReadEventReminderPreference(ctx context.Context, userID uint, eventID uint, reminderOffsetMinutes int) (EventReminderPreferenceDecision, error)
}

type DefaultEventReminderPreferenceReader struct{}

func NewDefaultEventReminderPreferenceReader() EventReminderPreferenceReader {
	return DefaultEventReminderPreferenceReader{}
}

func (DefaultEventReminderPreferenceReader) ReadEventReminderPreference(_ context.Context, _ uint, _ uint, offset int) (EventReminderPreferenceDecision, error) {
	if offset <= 0 {
		return EventReminderPreferenceDecision{}, ErrInvalidReminderOffset
	}
	if !IsDefaultReminderOffset(offset) {
		return EventReminderPreferenceDecision{Enabled: false}, nil
	}
	return EventReminderPreferenceDecision{Enabled: true, Channels: []string{ChannelInApp}}, nil
}

func IsDefaultReminderOffset(offset int) bool {
	switch offset {
	case EventReminderOffset24Hours, EventReminderOffset6Hours, EventReminderOffset1Hour:
		return true
	default:
		return false
	}
}

type EventReminderRecipient struct {
	UserID   uint     `json:"userId"`
	Channels []string `json:"channels"`
}

type EventReminderRecipients struct {
	EventID               uint                     `json:"eventId"`
	Title                 string                   `json:"title"`
	StartTime             time.Time                `json:"startTime"`
	ReminderOffsetMinutes int                      `json:"reminderOffsetMinutes"`
	Recipients            []EventReminderRecipient `json:"recipients"`
}

type EventReminderRecipientService interface {
	Resolve(ctx context.Context, eventID uint, reminderOffsetMinutes int) (EventReminderRecipients, error)
}

type eventReminderRecipientService struct {
	store       EventReminderRecipientStore
	preferences EventReminderPreferenceReader
}

func NewEventReminderRecipientService(store EventReminderRecipientStore, preferences EventReminderPreferenceReader) EventReminderRecipientService {
	return &eventReminderRecipientService{store: store, preferences: preferences}
}

func (s *eventReminderRecipientService) Resolve(ctx context.Context, eventID uint, offset int) (EventReminderRecipients, error) {
	if eventID == 0 || offset <= 0 {
		return EventReminderRecipients{}, ErrInvalidReminderOffset
	}
	if s.store == nil {
		return EventReminderRecipients{}, ErrRecipientStoreUnavailable
	}
	if s.preferences == nil {
		return EventReminderRecipients{}, ErrReminderPreferenceReaderNil
	}

	snapshot, err := s.store.LoadEventReminderSnapshot(ctx, eventID)
	if err != nil {
		return EventReminderRecipients{}, err
	}
	recipients := make([]EventReminderRecipient, 0, len(snapshot.UserIDs))
	seen := make(map[uint]struct{}, len(snapshot.UserIDs))
	for _, userID := range snapshot.UserIDs {
		if userID == 0 {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		decision, err := s.preferences.ReadEventReminderPreference(ctx, userID, eventID, offset)
		if err != nil {
			return EventReminderRecipients{}, err
		}
		if !decision.Enabled || len(decision.Channels) == 0 {
			continue
		}
		recipients = append(recipients, EventReminderRecipient{UserID: userID, Channels: append([]string(nil), decision.Channels...)})
	}
	if recipients == nil {
		recipients = []EventReminderRecipient{}
	}
	return EventReminderRecipients{
		EventID: snapshot.EventID, Title: snapshot.Title, StartTime: snapshot.StartTime,
		ReminderOffsetMinutes: offset, Recipients: recipients,
	}, nil
}
