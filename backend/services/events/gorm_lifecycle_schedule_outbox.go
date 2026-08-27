package events

import (
	"context"
	"fmt"

	eventmodels "friendship/models/events"
	"friendship/repository"
)

type gormEventLifecycleScheduleOutboxStore struct {
	repo repository.PostgresRepository
}

func NewGORMEventLifecycleScheduleOutboxReader(repo repository.PostgresRepository) EventLifecycleScheduleOutboxReader {
	return &gormEventLifecycleScheduleOutboxStore{repo: repo}
}

func (s *gormEventLifecycleScheduleOutboxStore) AppendLifecycleScheduleEvent(input EventLifecycleScheduleOutboxInput) error {
	record := eventmodels.EventLifecycleScheduleOutbox{
		MessageID:     input.MessageID,
		EventID:       input.EventID,
		Operation:     input.Operation,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		OccurredAt:    input.OccurredAt,
		SchemaVersion: input.SchemaVersion,
	}
	if err := s.repo.Create(&record).Error; err != nil {
		return fmt.Errorf("ошибка записи lifecycle schedule outbox: %w", err)
	}
	return nil
}

func (s *gormEventLifecycleScheduleOutboxStore) ListLifecycleScheduleEvents(ctx context.Context, after int64, limit int) ([]EventLifecycleScheduleEvent, error) {
	var records []eventmodels.EventLifecycleScheduleOutbox
	err := s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		return tx.Where("sequence > ?", after).
			Order("sequence ASC").
			Limit(limit).
			Find(&records).Error
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения lifecycle schedule outbox: %w", err)
	}

	items := make([]EventLifecycleScheduleEvent, 0, len(records))
	for _, record := range records {
		items = append(items, EventLifecycleScheduleEvent{
			Sequence:      record.Sequence,
			MessageID:     record.MessageID,
			SchemaVersion: record.SchemaVersion,
			Operation:     record.Operation,
			EventID:       record.EventID,
			StartTime:     record.StartTime,
			EndTime:       record.EndTime,
			OccurredAt:    record.OccurredAt,
		})
	}
	return items, nil
}
