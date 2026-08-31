package events

import (
	"context"
	"fmt"

	eventmodels "friendship/models/events"
	"friendship/repository"
)

type gormEventReminderIntentOutboxStore struct {
	repo repository.PostgresRepository
}

func NewGORMEventReminderIntentOutboxReader(repo repository.PostgresRepository) EventReminderIntentOutboxReader {
	return &gormEventReminderIntentOutboxStore{repo: repo}
}

func (s *gormEventReminderIntentOutboxStore) AppendEventReminderIntent(input EventReminderIntentOutboxInput) error {
	record := eventmodels.EventReminderIntentOutbox{
		MessageID:             input.MessageID,
		SchemaVersion:         input.SchemaVersion,
		IntentType:            input.IntentType,
		Operation:             input.Operation,
		EventID:               input.EventID,
		StartTime:             input.StartTime,
		ReminderOffsetMinutes: cloneReminderOffsets(input.ReminderOffsetMinutes),
		OccurredAt:            input.OccurredAt,
	}
	if err := s.repo.Create(&record).Error; err != nil {
		return fmt.Errorf("ошибка записи notification intent outbox: %w", err)
	}
	return nil
}

func (s *gormEventReminderIntentOutboxStore) ListEventReminderIntents(ctx context.Context, after int64, limit int) ([]EventReminderIntent, error) {
	var records []eventmodels.EventReminderIntentOutbox
	err := s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		return tx.Where("sequence > ?", after).Order("sequence ASC").Limit(limit).Find(&records).Error
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения notification intent outbox: %w", err)
	}

	items := make([]EventReminderIntent, 0, len(records))
	for _, record := range records {
		items = append(items, EventReminderIntent{
			Sequence:              record.Sequence,
			MessageID:             record.MessageID,
			SchemaVersion:         record.SchemaVersion,
			IntentType:            record.IntentType,
			Operation:             record.Operation,
			EventID:               record.EventID,
			StartTime:             record.StartTime,
			ReminderOffsetMinutes: cloneReminderOffsets(record.ReminderOffsetMinutes),
			OccurredAt:            record.OccurredAt,
		})
	}
	return items, nil
}
