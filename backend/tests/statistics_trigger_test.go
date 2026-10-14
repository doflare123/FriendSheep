package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	servicesevents "friendship/services/events"
)

func TestStatisticsTriggerCompletionAndCatchUpAreAtomicAndIdempotent(t *testing.T) {
	for _, initialStatus := range []string{eventmodels.StatusActive, eventmodels.StatusRecruitment} {
		t.Run(initialStatus, func(t *testing.T) {
			db := newEventsServiceDB(t)
			seedEventUser(t, db, 1)
			groupID := seedEventGroup(t, db, 1, false)
			eventID := seedEvent(t, db, groupID, 1, 0, 5)
			initialID := seedLifecycleStatus(t, db, initialStatus)
			completedID := seedLifecycleStatus(t, db, eventmodels.StatusCompleted)
			now := time.Date(2036, 1, 1, 12, 0, 0, 0, time.UTC)
			setLifecycleEvent(t, db, eventID, initialID, now.Add(-time.Hour), now)
			service := servicesevents.NewEventLifecycleService(servicesevents.NewGORMEventLifecycleStore(&testPostgresRepository{db: db}), lifecycleClockStub{now: now})
			if err := db.Exec(`CREATE TRIGGER fail_statistics_trigger BEFORE INSERT ON event_statistics_triggers BEGIN SELECT RAISE(ABORT, 'injected trigger failure'); END;`).Error; err != nil {
				t.Fatal(err)
			}
			if _, err := service.Advance(context.Background(), eventID); err == nil {
				t.Fatal("completion accepted despite failure to publish statistics")
			}
			assertLifecycleEventStatus(t, db, eventID, initialID)
			if err := db.Exec("DROP TRIGGER fail_statistics_trigger").Error; err != nil {
				t.Fatal(err)
			}
			first, err := service.Advance(context.Background(), eventID)
			if err != nil || !first.Applied || first.CurrentStatus != eventmodels.StatusCompleted {
				t.Fatalf("completion/error = %#v/%v", first, err)
			}
			assertLifecycleEventStatus(t, db, eventID, completedID)
			second, err := service.Advance(context.Background(), eventID)
			if err != nil || second.Applied || second.Outcome != servicesevents.LifecycleOutcomeAlreadyCompleted {
				t.Fatalf("repeat/error = %#v/%v", second, err)
			}
			var triggers []eventmodels.StatisticsTrigger
			if err := db.Find(&triggers).Error; err != nil || len(triggers) != 1 {
				t.Fatalf("triggers/error = %#v/%v", triggers, err)
			}
			trigger := triggers[0]
			if trigger.SchemaVersion != 1 || trigger.EventID != eventID || !trigger.CompletedAt.Equal(now) || trigger.Sequence <= 0 {
				t.Fatalf("trigger = %#v", trigger)
			}
		})
	}
}

func TestStatisticsTriggerBackfillIsBoundedResumableAndRepeatable(t *testing.T) {
	db := newEventsServiceDB(t)
	groupID, _ := statisticsFixture(t, db)
	ids := []uint{
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-3*time.Hour), 30, []uint{2}, nil),
		seedEvent(t, db, groupID, 1, 0, 5),
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-2*time.Hour), 15, nil, nil),
	}
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewStatisticsTriggerService(servicesevents.NewGORMStatisticsTriggerReader(repo), servicesevents.NewGORMStatisticsTriggerBackfiller(repo))
	first, err := service.Backfill(context.Background(), 0, 2)
	if err != nil || first.Scanned != 2 || first.Published != 1 || !first.HasMore || first.NextEventID != ids[1] {
		t.Fatalf("first backfill/error = %#v/%v", first, err)
	}
	second, err := service.Backfill(context.Background(), first.NextEventID, 2)
	if err != nil || second.Scanned != 1 || second.Published != 1 || second.HasMore || second.NextEventID != ids[2] {
		t.Fatalf("resumed backfill/error = %#v/%v", second, err)
	}
	repeated, err := service.Backfill(context.Background(), 0, 500)
	if err != nil || repeated.Published != 0 || repeated.Scanned != 3 {
		t.Fatalf("repeated backfill/error = %#v/%v", repeated, err)
	}
	page, err := service.List(context.Background(), 0, 1)
	if err != nil || len(page.Messages) != 1 || !page.HasMore || page.Messages[0].EventID != ids[0] || page.NextCursor != page.Messages[0].Sequence {
		t.Fatalf("first page/error = %#v/%v", page, err)
	}
	page2, err := service.List(context.Background(), page.NextCursor, 1)
	if err != nil || len(page2.Messages) != 1 || page2.HasMore || page2.Messages[0].EventID != ids[2] {
		t.Fatalf("second page/error = %#v/%v", page2, err)
	}
	empty, err := service.List(context.Background(), page2.NextCursor, 1)
	if err != nil || empty.Messages == nil || len(empty.Messages) != 0 || empty.NextCursor != page2.NextCursor || empty.HasMore {
		t.Fatalf("empty page/error = %#v/%v", empty, err)
	}
	if _, err := service.List(context.Background(), -1, 1); !errors.Is(err, servicesevents.ErrInvalidStatisticsTriggerCursor) {
		t.Fatalf("negative cursor error = %v", err)
	}
	if _, err := service.Backfill(context.Background(), 0, 501); !errors.Is(err, servicesevents.ErrInvalidStatisticsTriggerLimit) {
		t.Fatalf("over-limit error = %v", err)
	}
}
