package events

import (
	"context"
	"errors"
	"fmt"
	eventmodels "friendship/models/events"
	"time"
)

const (
	LifecycleOutcomeNotDue           = "not_due"
	LifecycleOutcomeStarted          = "started"
	LifecycleOutcomeCompleted        = "completed"
	LifecycleOutcomeCompletedCatchUp = "completed_catch_up"
	LifecycleOutcomeAlreadyActive    = "already_active"
	LifecycleOutcomeAlreadyCompleted = "already_completed"
)

var ErrInvalidEventLifecycleState = errors.New("некорректное состояние жизненного цикла мероприятия")

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

type EventLifecycleSnapshot struct {
	ID         uint
	StatusID   uint
	StatusName string
	StartTime  time.Time
	EndTime    time.Time
}

type EventLifecycleTransaction interface {
	LoadEventForUpdate(ctx context.Context, eventID uint) (EventLifecycleSnapshot, error)
	ResolveStatusID(ctx context.Context, statusName string) (uint, error)
	UpdateEventStatus(ctx context.Context, eventID uint, statusID uint) error
}

type EventLifecycleStore interface {
	WithinLifecycleTransaction(ctx context.Context, eventID uint, fn func(EventLifecycleTransaction) error) error
}

type EventLifecycleResult struct {
	EventID        uint      `json:"eventId"`
	PreviousStatus string    `json:"previousStatus"`
	CurrentStatus  string    `json:"currentStatus"`
	Outcome        string    `json:"outcome"`
	Applied        bool      `json:"applied"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	ProcessedAt    time.Time `json:"processedAt"`
}

type EventLifecycleService interface {
	Advance(ctx context.Context, eventID uint) (EventLifecycleResult, error)
}

type eventLifecycleService struct {
	store EventLifecycleStore
	clock Clock
}

func NewEventLifecycleService(store EventLifecycleStore, clock Clock) EventLifecycleService {
	if clock == nil {
		clock = SystemClock{}
	}
	return &eventLifecycleService{store: store, clock: clock}
}

func (s *eventLifecycleService) Advance(ctx context.Context, eventID uint) (EventLifecycleResult, error) {
	if eventID == 0 {
		return EventLifecycleResult{}, ErrEventNotFound
	}
	var result EventLifecycleResult
	err := s.store.WithinLifecycleTransaction(ctx, eventID, func(tx EventLifecycleTransaction) error {
		snapshot, err := tx.LoadEventForUpdate(ctx, eventID)
		if err != nil {
			return err
		}
		if !snapshot.EndTime.After(snapshot.StartTime) {
			return fmt.Errorf("%w: end_time должен быть позже start_time", ErrInvalidEventLifecycleState)
		}
		processedAt := s.clock.Now()

		result = EventLifecycleResult{
			EventID:        snapshot.ID,
			PreviousStatus: snapshot.StatusName,
			CurrentStatus:  snapshot.StatusName,
			StartTime:      snapshot.StartTime,
			EndTime:        snapshot.EndTime,
			ProcessedAt:    processedAt,
		}

		targetStatus, outcome, applied, err := lifecycleDecision(snapshot.StatusName, snapshot.StartTime, snapshot.EndTime, processedAt)
		if err != nil {
			return err
		}
		result.Outcome = outcome
		result.Applied = applied
		if !applied {
			return nil
		}

		targetStatusID, err := tx.ResolveStatusID(ctx, targetStatus)
		if err != nil {
			return fmt.Errorf("resolve lifecycle status %q: %w", targetStatus, err)
		}
		if targetStatusID == 0 {
			return fmt.Errorf("%w: статус %q не настроен", ErrInvalidEventLifecycleState, targetStatus)
		}
		if err := tx.UpdateEventStatus(ctx, snapshot.ID, targetStatusID); err != nil {
			return fmt.Errorf("update event lifecycle status: %w", err)
		}
		result.CurrentStatus = targetStatus
		return nil
	})
	if err != nil {
		return EventLifecycleResult{}, err
	}
	return result, nil
}

func lifecycleDecision(status string, startTime, endTime, now time.Time) (string, string, bool, error) {
	switch status {
	case eventmodels.StatusRecruitment:
		if !now.Before(endTime) {
			return eventmodels.StatusCompleted, LifecycleOutcomeCompletedCatchUp, true, nil
		}
		if !now.Before(startTime) {
			return eventmodels.StatusActive, LifecycleOutcomeStarted, true, nil
		}
		return status, LifecycleOutcomeNotDue, false, nil
	case eventmodels.StatusActive:
		if !now.Before(endTime) {
			return eventmodels.StatusCompleted, LifecycleOutcomeCompleted, true, nil
		}
		return status, LifecycleOutcomeAlreadyActive, false, nil
	case eventmodels.StatusCompleted:
		return status, LifecycleOutcomeAlreadyCompleted, false, nil
	default:
		return "", "", false, fmt.Errorf("%w: неизвестный статус %q", ErrInvalidEventLifecycleState, status)
	}
}
