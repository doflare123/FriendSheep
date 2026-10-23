package statistics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	eventmodels "friendship/models/events"
	statmodels "friendship/models/statistics"
	"friendship/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormStore struct {
	repo          repository.PostgresRepository
	initialSpread time.Duration
}
type gormTransaction struct {
	db            *gorm.DB
	initialSpread time.Duration
}

func NewGORMStore(repo repository.PostgresRepository) Store {
	return NewGORMStoreWithUniquePeopleConfig(repo, DefaultUniquePeopleConfig())
}

func NewGORMStoreWithUniquePeopleConfig(repo repository.PostgresRepository, config UniquePeopleConfig) Store {
	spread := config.InitialSpread
	if spread <= 0 {
		spread = DefaultUniquePeopleConfig().InitialSpread
	}
	return &gormStore{repo: repo, initialSpread: spread}
}
func (s *gormStore) WithinTransaction(ctx context.Context, fn func(Transaction) error) error {
	return s.repo.TransactionWithContext(ctx, func(repo repository.PostgresRepository) error {
		return fn(&gormTransaction{db: repo.Model(nil).WithContext(ctx), initialSpread: s.initialSpread})
	})
}
func (tx *gormTransaction) LoadEvent(ctx context.Context, id uint) (Snapshot, error) {
	var event eventmodels.Event
	err := tx.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Status").First(&event, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Snapshot{}, ErrEventMissing
	}
	if err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{EventID: id, CreatorID: event.CreatorID, TypeID: event.EventTypeID, StartTime: event.StartTime, EndTime: event.EndTime, DurationMinutes: int64(event.Duration), Completed: event.Status.Name == eventmodels.StatusCompleted}
	var processed statmodels.ProcessedEvent
	err = tx.db.First(&processed, "event_id = ?", id).Error
	if err == nil {
		result.ProcessedAt = &processed.ProcessedAt
		return result, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Snapshot{}, err
	}
	if err = tx.db.Model(&eventmodels.EventsUser{}).Where("event_id = ?", id).Order("user_id ASC").Pluck("user_id", &result.Participants).Error; err != nil {
		return Snapshot{}, err
	}
	if err = tx.db.Model(&eventmodels.EventGenre{}).Where("event_id = ?", id).Order("genre_id ASC").Pluck("genre_id", &result.Genres).Error; err != nil {
		return Snapshot{}, err
	}
	result.Participants = uniqueIDs(result.Participants)
	result.Genres = uniqueIDs(result.Genres)
	return result, nil
}
func uniqueIDs(ids []uint) []uint {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := ids[:0]
	for _, id := range ids {
		if len(out) == 0 || out[len(out)-1] != id {
			out = append(out, id)
		}
	}
	return out
}
func (tx *gormTransaction) Apply(ctx context.Context, c Contribution, now time.Time) (bool, error) {
	db := tx.db.WithContext(ctx)
	mark := statmodels.ProcessedEvent{EventID: c.EventID, ProcessedAt: now}
	insert := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&mark)
	if insert.Error != nil {
		return false, insert.Error
	}
	if insert.RowsAffected == 0 {
		return false, nil
	}
	// Создание и блокировки агрегатов идут в едином порядке между всеми workers.
	users := uniqueIDs(append(append([]uint{}, c.Participants...), c.CreatorID))
	for _, id := range users {
		row := statmodels.UserAggregate{UserID: id, UpdatedAt: now}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return false, err
		}
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "user_id = ?", id).Error; err != nil {
			return false, err
		}
	}
	updatedAt := gorm.Expr("CASE WHEN updated_at > ? THEN updated_at ELSE ? END", now, now)
	if err := db.Model(&statmodels.UserAggregate{}).Where("user_id = ?", c.CreatorID).Updates(map[string]any{"created_completed_events": gorm.Expr("created_completed_events + 1"), "updated_at": updatedAt}).Error; err != nil {
		return false, err
	}
	for _, id := range c.Participants {
		if err := db.Model(&statmodels.UserAggregate{}).Where("user_id = ?", id).Updates(map[string]any{"completed_events": gorm.Expr("completed_events + 1"), "planned_minutes": gorm.Expr("planned_minutes + ?", c.DurationMinutes), "updated_at": updatedAt}).Error; err != nil {
			return false, err
		}
		if err := markUniquePeopleDirty(db, id, now, tx.initialSpread); err != nil {
			return false, err
		}
		items := []statmodels.Distribution{{UserID: id, Kind: "type", Key: numericKey(c.TypeID), Count: 1}, {UserID: id, Kind: "month", Key: c.Month, Count: 1}, {UserID: id, Kind: "weekday", Key: strconv.Itoa(c.Weekday), Count: 1}}
		for _, genre := range c.Genres {
			items = append(items, statmodels.Distribution{UserID: id, Kind: "genre", Key: numericKey(genre), Count: 1})
		}
		for i := range items {
			if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "kind"}, {Name: "key"}}, DoUpdates: clause.Assignments(map[string]any{"count": gorm.Expr("participation_statistics_distributions.count + 1")})}).Create(&items[i]).Error; err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func markUniquePeopleDirty(db *gorm.DB, userID uint, now time.Time, spread time.Duration) error {
	job := statmodels.UniquePeopleJob{UserID: userID, NextAttemptAt: now}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&job).Error; err != nil {
		return err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, "user_id = ?", userID).Error; err != nil {
		return err
	}
	wasFresh := job.RequestedGeneration == job.AppliedGeneration
	nextGeneration := job.RequestedGeneration + 1
	updates := map[string]any{"requested_generation": nextGeneration}
	if job.LeaseToken != nil {
		if job.FollowupDirtySince == nil {
			updates["followup_dirty_since"] = now
		}
	} else if wasFresh || job.DirtySince == nil {
		updates["dirty_since"] = now
		updates["next_attempt_at"] = now.Add(uniquePeopleSpreadOffset(userID, nextGeneration, spread))
	}
	return db.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", userID).Updates(updates).Error
}
func (s *gormStore) Read(ctx context.Context, userID uint) (View, error) {
	view := emptyView()
	db := s.repo.Model(nil).WithContext(ctx)
	options := &sql.TxOptions{ReadOnly: true}
	if db.Dialector.Name() == "postgres" {
		options.Isolation = sql.LevelRepeatableRead
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		var row statmodels.UserAggregate
		err := tx.First(&row, "user_id = ?", userID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		view.CompletedEvents = row.CompletedEvents
		view.PlannedMinutes = row.PlannedMinutes
		view.CreatedCompletedEvents = row.CreatedCompletedEvents
		view.UniquePeopleCount = row.UniquePeopleCount
		view.UniquePeopleCalculatedAt = row.UniquePeopleCalculatedAt
		view.LastUpdatedAt = &row.UpdatedAt
		var job statmodels.UniquePeopleJob
		jobErr := tx.First(&job, "user_id = ?", userID).Error
		if jobErr != nil && !errors.Is(jobErr, gorm.ErrRecordNotFound) {
			return jobErr
		}
		if row.UniquePeopleCount == nil || row.UniquePeopleCalculatedAt == nil {
			view.UniquePeopleStatus = UniquePeopleStatusNotCalculated
		} else if jobErr == nil && job.RequestedGeneration > job.AppliedGeneration {
			view.UniquePeopleStatus = UniquePeopleStatusPending
		} else {
			view.UniquePeopleStatus = UniquePeopleStatusFresh
		}
		var items []statmodels.Distribution
		if err := tx.Where("user_id = ?", userID).Order("kind ASC, key ASC").Find(&items).Error; err != nil {
			return err
		}
		var maxType, maxGenre, maxDay int64
		for _, item := range items {
			if item.Kind == "month" {
				view.Months = append(view.Months, MonthCount{Month: item.Key, Count: item.Count})
				continue
			}
			id, err := strconv.ParseUint(item.Key, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ключ распределения: %w", err)
			}
			key := uint(id)
			switch item.Kind {
			case "type":
				view.Types = append(view.Types, CountItem{ID: key, Count: item.Count})
				if item.Count > maxType || (item.Count == maxType && (view.MostFrequentTypeID == nil || key < *view.MostFrequentTypeID)) {
					maxType = item.Count
					v := key
					view.MostFrequentTypeID = &v
				}
			case "genre":
				view.Genres = append(view.Genres, CountItem{ID: key, Count: item.Count})
				if item.Count > maxGenre || (item.Count == maxGenre && (view.MostFrequentGenreID == nil || key < *view.MostFrequentGenreID)) {
					maxGenre = item.Count
					v := key
					view.MostFrequentGenreID = &v
				}
			case "weekday":
				day := int(key)
				view.Weekdays = append(view.Weekdays, WeekdayCount{Weekday: day, Count: item.Count})
				if item.Count > maxDay || (item.Count == maxDay && (view.MostFrequentWeekday == nil || day < *view.MostFrequentWeekday)) {
					maxDay = item.Count
					view.MostFrequentWeekday = &day
				}
			default:
				return fmt.Errorf("неизвестный вид распределения %q", item.Kind)
			}
		}
		sort.Slice(view.Types, func(i, j int) bool { return view.Types[i].ID < view.Types[j].ID })
		sort.Slice(view.Genres, func(i, j int) bool { return view.Genres[i].ID < view.Genres[j].ID })
		sort.Slice(view.Weekdays, func(i, j int) bool { return view.Weekdays[i].Weekday < view.Weekdays[j].Weekday })
		return nil
	}, options)
	if err != nil {
		return View{}, err
	}
	return view, nil
}
