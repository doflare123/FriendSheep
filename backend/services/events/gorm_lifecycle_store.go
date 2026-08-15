package events

import (
	"context"
	"errors"
	"fmt"
	eventmodels "friendship/models/events"
	"friendship/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormEventLifecycleStore struct {
	repo repository.PostgresRepository
}

func NewGORMEventLifecycleStore(repo repository.PostgresRepository) EventLifecycleStore {
	return &gormEventLifecycleStore{repo: repo}
}

func (s *gormEventLifecycleStore) WithinLifecycleTransaction(
	ctx context.Context,
	_ uint,
	fn func(EventLifecycleTransaction) error,
) error {
	return s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		return fn(&gormEventLifecycleTransaction{repo: tx})
	})
}

type gormEventLifecycleTransaction struct {
	repo repository.PostgresRepository
}

func (tx *gormEventLifecycleTransaction) LoadEventForUpdate(ctx context.Context, eventID uint) (EventLifecycleSnapshot, error) {
	var event eventmodels.Event
	err := tx.repo.Model(&eventmodels.Event{}).
		WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Status").
		First(&event, eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EventLifecycleSnapshot{}, ErrEventNotFound
		}
		return EventLifecycleSnapshot{}, fmt.Errorf("load event lifecycle state: %w", err)
	}
	if event.Status.ID == 0 || event.Status.Name == "" {
		return EventLifecycleSnapshot{}, fmt.Errorf("%w: статус мероприятия не найден", ErrInvalidEventLifecycleState)
	}
	return EventLifecycleSnapshot{
		ID:         event.ID,
		StatusID:   event.StatusID,
		StatusName: event.Status.Name,
		StartTime:  event.StartTime,
		EndTime:    event.EndTime,
	}, nil
}

func (tx *gormEventLifecycleTransaction) ResolveStatusID(ctx context.Context, statusName string) (uint, error) {
	var status eventmodels.Status
	err := tx.repo.Model(&eventmodels.Status{}).
		WithContext(ctx).
		Where("name = ?", statusName).
		First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("%w: статус %q не настроен", ErrInvalidEventLifecycleState, statusName)
		}
		return 0, fmt.Errorf("load lifecycle status: %w", err)
	}
	return status.ID, nil
}

func (tx *gormEventLifecycleTransaction) UpdateEventStatus(ctx context.Context, eventID uint, statusID uint) error {
	result := tx.repo.Model(&eventmodels.Event{}).
		WithContext(ctx).
		Where("id = ?", eventID).
		Update("status_id", statusID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrEventNotFound
	}
	return nil
}
