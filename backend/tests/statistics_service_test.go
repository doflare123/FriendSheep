package tests

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"friendship/models"
	eventmodels "friendship/models/events"
	statmodels "friendship/models/statistics"
	stats "friendship/services/statistics"

	"gorm.io/gorm"
)

func migrateParticipationStatistics(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&statmodels.UserAggregate{}, &statmodels.UserPair{}, &statmodels.ProcessedEvent{}, &statmodels.Distribution{}); err != nil {
		t.Fatal(err)
	}
}

func seedStatisticsEvent(t *testing.T, db *gorm.DB, groupID, creatorID, typeID uint, start time.Time, minutes uint16, participants, genres []uint) uint {
	t.Helper()
	id := seedEvent(t, db, groupID, creatorID, uint16(len(participants)), 100)
	statusID := seedLifecycleStatus(t, db, eventmodels.StatusCompleted)
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", id).Updates(map[string]any{
		"event_type_id": typeID, "status_id": statusID, "start_time": start,
		"end_time": start.Add(time.Duration(minutes) * time.Minute), "duration": minutes,
	}).Error; err != nil {
		t.Fatal(err)
	}
	for _, userID := range participants {
		seedEventParticipant(t, db, id, userID)
	}
	for _, genreID := range genres {
		seedEventGenreRelation(t, db, id, genreID)
	}
	return id
}

func statisticsFixture(t *testing.T, db *gorm.DB) (uint, *stats.Service) {
	t.Helper()
	migrateParticipationStatistics(t, db)
	for _, id := range []uint{1, 2, 3, 4} {
		seedEventUser(t, db, id)
	}
	seedEventReferences(t, db)
	if err := db.Create(&models.Category{ID: 2, Name: "Спорт"}).Error; err != nil {
		t.Fatal(err)
	}
	groupID := seedEventGroup(t, db, 1, false)
	return groupID, stats.NewService(stats.NewGORMStore(&testPostgresRepository{db: db}))
}

func TestParticipationStatisticsAccumulatesPairsAndAllTimeContributions(t *testing.T) {
	db := newEventsServiceDB(t)
	groupID, service := statisticsFixture(t, db)
	genre1 := seedEventGenre(t, db, "Стратегия")
	genre2 := seedEventGenre(t, db, "Командное")
	// Локальный октябрь и четверг превращаются в сентябрь и среду по согласованному UTC.
	start1 := time.Date(2026, 10, 1, 0, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	start2 := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	ids := []uint{
		seedStatisticsEvent(t, db, groupID, 1, 1, start1, 90, []uint{3, 2}, []uint{genre2, genre1}),
		seedStatisticsEvent(t, db, groupID, 2, 1, start2, 120, []uint{4, 2, 3}, []uint{genre1}),
		seedStatisticsEvent(t, db, groupID, 1, 2, start2, 30, []uint{2}, []uint{genre2}),
	}
	for _, eventID := range ids {
		first, err := service.Process(context.Background(), eventID)
		if err != nil || first.Outcome != "processed" || first.ProcessedAt == nil {
			t.Fatalf("first result/error = %#v/%v", first, err)
		}
		// Успешный commit уже произошёл; повтор моделирует retry после утраты ответа.
		second, err := service.Process(context.Background(), eventID)
		if err != nil || second.Outcome != "already_processed" || second.ProcessedAt == nil || !second.ProcessedAt.Equal(*first.ProcessedAt) {
			t.Fatalf("repeat result/error = %#v/%v; first=%#v", second, err, first)
		}
	}
	view, err := service.Read(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if view.CompletedEvents != 3 || view.PlannedMinutes != 240 || view.CreatedCompletedEvents != 1 || view.UniquePeopleCount != 2 || view.Timezone != "UTC" || view.LastUpdatedAt == nil {
		t.Fatalf("view = %#v", view)
	}
	if !reflect.DeepEqual(view.Types, []stats.CountItem{{ID: 1, Count: 2}, {ID: 2, Count: 1}}) ||
		!reflect.DeepEqual(view.Genres, []stats.CountItem{{ID: genre1, Count: 2}, {ID: genre2, Count: 2}}) ||
		!reflect.DeepEqual(view.Months, []stats.MonthCount{{Month: "2026-09", Count: 1}, {Month: "2026-10", Count: 2}}) ||
		!reflect.DeepEqual(view.Weekdays, []stats.WeekdayCount{{Weekday: 1, Count: 2}, {Weekday: 3, Count: 1}}) {
		t.Fatalf("distributions = %#v", view)
	}
	if view.MostFrequentGenreID == nil || *view.MostFrequentGenreID != genre1 || view.MostFrequentTypeID == nil || *view.MostFrequentTypeID != 1 || view.MostFrequentWeekday == nil || *view.MostFrequentWeekday != 1 {
		t.Fatalf("maxima = %#v", view)
	}
	creator, err := service.Read(context.Background(), 1)
	if err != nil || creator.CreatedCompletedEvents != 2 || creator.CompletedEvents != 0 || creator.PlannedMinutes != 0 || creator.UniquePeopleCount != 0 || len(creator.Types) != 0 {
		t.Fatalf("creator without membership = %#v, error=%v", creator, err)
	}
	for _, userID := range []uint{3, 4} {
		participant, err := service.Read(context.Background(), userID)
		if err != nil || participant.UniquePeopleCount != 2 {
			t.Fatalf("participant %d = %#v, error=%v", userID, participant, err)
		}
	}
	var pairs, selfPairs, processed int64
	if err := db.Model(&statmodels.UserPair{}).Count(&pairs).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&statmodels.UserPair{}).Where("user_id = encountered_user_id").Count(&selfPairs).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&statmodels.ProcessedEvent{}).Count(&processed).Error; err != nil {
		t.Fatal(err)
	}
	if pairs != 6 || selfPairs != 0 || processed != 3 {
		t.Fatalf("pairs/self/processed = %d/%d/%d", pairs, selfPairs, processed)
	}
}

func TestParticipationStatisticsTiesAndReadUsesOnlyAggregates(t *testing.T) {
	db := newEventsServiceDB(t)
	groupID, service := statisticsFixture(t, db)
	genre1 := seedEventGenre(t, db, "Первый жанр")
	genre2 := seedEventGenre(t, db, "Второй жанр")
	for _, id := range []uint{
		seedStatisticsEvent(t, db, groupID, 1, 2, time.Date(2026, 10, 6, 8, 0, 0, 0, time.UTC), 20, []uint{2}, []uint{genre2}),
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC), 20, []uint{2}, []uint{genre1}),
	} {
		if _, err := service.Process(context.Background(), id); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Callback().Query().Before("gorm:query").Register("reject_statistics_history_scan", func(tx *gorm.DB) {
		switch tx.Statement.Table {
		case "events", "events_users", "event_genres", "participation_statistics_pairs":
			tx.AddError(errors.New("read endpoint попытался пересчитать историю"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	view, err := service.Read(context.Background(), 2)
	if err != nil || view.MostFrequentTypeID == nil || *view.MostFrequentTypeID != 1 || view.MostFrequentGenreID == nil || *view.MostFrequentGenreID != genre1 || view.MostFrequentWeekday == nil || *view.MostFrequentWeekday != 1 {
		t.Fatalf("tie maxima/read result/error=%#v/%v", view, err)
	}
}

func TestParticipationStatisticsEmptyHistoryAndNotCompleted(t *testing.T) {
	db := newEventsServiceDB(t)
	groupID, service := statisticsFixture(t, db)
	view, err := service.Read(context.Background(), 2)
	if err != nil || view.CompletedEvents != 0 || view.PlannedMinutes != 0 || view.CreatedCompletedEvents != 0 || view.UniquePeopleCount != 0 || view.LastUpdatedAt != nil || view.MostFrequentTypeID != nil || view.MostFrequentGenreID != nil || view.MostFrequentWeekday != nil || view.Timezone != "UTC" {
		t.Fatalf("empty view/error = %#v/%v", view, err)
	}
	if view.Types == nil || view.Genres == nil || view.Months == nil || view.Weekdays == nil || len(view.Types)+len(view.Genres)+len(view.Months)+len(view.Weekdays) != 0 {
		t.Fatalf("empty distributions must be JSON arrays: %#v", view)
	}
	for _, variant := range []string{"delayed_status", "future_end"} {
		t.Run(variant, func(t *testing.T) {
			id := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-time.Hour), 20, []uint{2}, nil)
			if variant == "delayed_status" {
				db.Model(&eventmodels.Event{}).Where("id = ?", id).Update("status_id", 1)
			} else {
				db.Model(&eventmodels.Event{}).Where("id = ?", id).Update("end_time", time.Now().Add(time.Hour))
			}
			result, err := service.Process(context.Background(), id)
			if err != nil || result.Outcome != "not_completed" || result.ProcessedAt != nil {
				t.Fatalf("unready process result/error = %#v/%v", result, err)
			}
		})
	}
	missing, err := service.Process(context.Background(), 99999)
	if err != nil || missing.Outcome != "missing" {
		t.Fatalf("missing result/error = %#v/%v", missing, err)
	}
	var count int64
	if err := db.Model(&statmodels.ProcessedEvent{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unready event processed count/error = %d/%v", count, err)
	}
}

func TestParticipationStatisticsRollbackLeavesNoPartialFacts(t *testing.T) {
	db := newEventsServiceDB(t)
	groupID, service := statisticsFixture(t, db)
	eventID := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-2*time.Hour), 45, []uint{2, 3}, nil)
	if err := db.Exec(`CREATE TRIGGER fail_statistics_distribution BEFORE INSERT ON participation_statistics_distributions BEGIN SELECT RAISE(ABORT, 'injected statistics failure'); END;`).Error; err != nil {
		t.Fatal(err)
	}
	result, err := service.Process(context.Background(), eventID)
	if err == nil || !reflect.DeepEqual(result, stats.ProcessResult{}) {
		t.Fatalf("failed process = %#v/%v", result, err)
	}
	for _, model := range []any{&statmodels.UserAggregate{}, &statmodels.UserPair{}, &statmodels.ProcessedEvent{}, &statmodels.Distribution{}} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("rollback %T count/error = %d/%v", model, count, err)
		}
	}
	if err := db.Exec("DROP TRIGGER fail_statistics_distribution").Error; err != nil {
		t.Fatal(err)
	}
	if result, err := service.Process(context.Background(), eventID); err != nil || result.Outcome != "processed" {
		t.Fatalf("retry result/error = %#v/%v", result, err)
	}
	if err := db.Create(&statmodels.UserPair{UserID: 2, EncounteredUserID: 2}).Error; err == nil {
		t.Fatal("DB приняла пару с самим собой")
	}
	if err := db.Create(&statmodels.UserPair{UserID: 2, EncounteredUserID: 3}).Error; err == nil {
		t.Fatal("DB приняла повторную направленную пару")
	}
}
