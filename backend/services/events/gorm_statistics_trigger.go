package events

import (
	"context"
	"fmt"
	eventmodels "friendship/models/events"
	"friendship/repository"
	"gorm.io/gorm/clause"
	"time"
)

type gormStatisticsTriggerStore struct{ repo repository.PostgresRepository }

func NewGORMStatisticsTriggerReader(repo repository.PostgresRepository) StatisticsTriggerReader {
	return &gormStatisticsTriggerStore{repo: repo}
}

func NewGORMStatisticsTriggerBackfiller(repo repository.PostgresRepository) StatisticsTriggerBackfiller {
	return &gormStatisticsTriggerStore{repo: repo}
}

func appendStatisticsTrigger(ctx context.Context, repo repository.PostgresRepository, eventID uint, completedAt time.Time) (bool, error) {
	record := eventmodels.StatisticsTrigger{SchemaVersion: StatisticsTriggerSchemaVersion, EventID: eventID, CompletedAt: completedAt.UTC()}
	result := repo.Model(&eventmodels.StatisticsTrigger{}).WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_id"}}, DoNothing: true}).Create(&record)
	if result.Error != nil {
		return false, fmt.Errorf("ошибка публикации statistics trigger: %w", result.Error)
	}
	return result.RowsAffected == 1, nil
}

func (s *gormStatisticsTriggerStore) ListStatisticsTriggers(ctx context.Context, after int64, limit int) ([]StatisticsTriggerMessage, error) {
	var records []eventmodels.StatisticsTrigger
	err := s.repo.Model(&eventmodels.StatisticsTrigger{}).WithContext(ctx).
		Where("sequence > ?", after).Order("sequence ASC").Limit(limit).Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения statistics trigger: %w", err)
	}
	messages := make([]StatisticsTriggerMessage, 0, len(records))
	for _, record := range records {
		messages = append(messages, StatisticsTriggerMessage{Sequence: record.Sequence, SchemaVersion: record.SchemaVersion, EventID: record.EventID, CompletedAt: record.CompletedAt.UTC()})
	}
	return messages, nil
}

func (s *gormStatisticsTriggerStore) BackfillStatisticsTriggers(ctx context.Context, afterEventID uint, limit int) (StatisticsBackfillResult, error) {
	limit, err := statisticsTriggerLimit(limit)
	if err != nil {
		return StatisticsBackfillResult{}, err
	}
	result := StatisticsBackfillResult{NextEventID: afterEventID}
	err = s.repo.TransactionWithContext(ctx, func(tx repository.PostgresRepository) error {
		if err := lockEventOutboxPublication(tx); err != nil {
			return err
		}
		var records []eventmodels.Event
		// Страницы ограничены по ID: возобновление не требует повторного сканирования всей истории.
		if err := tx.Model(&eventmodels.Event{}).WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).Where("id > ?", afterEventID).
			Order("id ASC").Limit(limit + 1).Preload("Status").Find(&records).Error; err != nil {
			return err
		}
		result.HasMore = len(records) > limit
		if result.HasMore {
			records = records[:limit]
		}
		now := time.Now()
		for _, event := range records {
			result.NextEventID = event.ID
			result.Scanned++
			if event.Status.Name != eventmodels.StatusCompleted || !eventHasEnded(event.EndTime, now) {
				continue
			}
			published, err := appendStatisticsTrigger(ctx, tx, event.ID, event.EndTime)
			if err != nil {
				return err
			}
			if published {
				result.Published++
			}
		}
		return nil
	})
	if err != nil {
		return StatisticsBackfillResult{}, fmt.Errorf("ошибка backfill statistics trigger: %w", err)
	}
	return result, nil
}
