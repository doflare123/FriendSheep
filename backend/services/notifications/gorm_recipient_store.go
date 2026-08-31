package notifications

import (
	"context"
	"errors"
	"fmt"

	eventmodels "friendship/models/events"
	"friendship/repository"

	"gorm.io/gorm"
)

type gormEventReminderRecipientStore struct {
	repo repository.PostgresRepository
}

func NewGORMEventReminderRecipientStore(repo repository.PostgresRepository) EventReminderRecipientStore {
	return &gormEventReminderRecipientStore{repo: repo}
}

func (s *gormEventReminderRecipientStore) LoadEventReminderSnapshot(ctx context.Context, eventID uint) (EventReminderSnapshot, error) {
	var snapshot EventReminderSnapshot
	err := s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		var event eventmodels.Event
		if err := tx.Select("id", "title", "start_time").First(&event, eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEventNotFound
			}
			return fmt.Errorf("ошибка чтения мероприятия для напоминания: %w", err)
		}
		var userIDs []uint
		if err := tx.Model(&eventmodels.EventsUser{}).
			Distinct("user_id").
			Where("event_id = ?", eventID).
			Order("user_id ASC").
			Pluck("user_id", &userIDs).Error; err != nil {
			return fmt.Errorf("ошибка чтения участников мероприятия: %w", err)
		}
		snapshot = EventReminderSnapshot{EventID: event.ID, Title: event.Title, StartTime: event.StartTime, UserIDs: userIDs}
		return nil
	})
	return snapshot, err
}
