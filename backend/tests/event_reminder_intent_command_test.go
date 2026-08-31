package tests

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"

	"gorm.io/gorm"
)

func TestEventCommandServiceWritesReminderIntentForCreateRescheduleAndDelete(t *testing.T) {
	t.Run("create publishes three offsets", func(t *testing.T) {
		outbox := &eventReminderIntentOutboxStub{}
		service := newReminderIntentCommandService(newEventCommandStoreStub(), outbox)
		input := validCreateEventCommandInput()

		if _, err := service.CreateEvent(context.Background(), 5, input); err != nil {
			t.Fatalf("CreateEvent() returned error: %v", err)
		}
		assertReminderIntentUpsert(t, outbox.entries, 77, input.StartTime)
	})

	t.Run("start time reschedule publishes a new revision", func(t *testing.T) {
		outbox := &eventReminderIntentOutboxStub{}
		service := newReminderIntentCommandService(newEventCommandStoreStub(), outbox)
		start := time.Now().Add(72 * time.Hour).Truncate(time.Second)

		if _, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{StartTime: &start}); err != nil {
			t.Fatalf("UpdateEvent() returned error: %v", err)
		}
		assertReminderIntentUpsert(t, outbox.entries, 77, start)
	})

	t.Run("duration only does not recreate reminder schedule", func(t *testing.T) {
		outbox := &eventReminderIntentOutboxStub{}
		service := newReminderIntentCommandService(newEventCommandStoreStub(), outbox)
		duration := uint16(240)

		if _, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{Duration: &duration}); err != nil {
			t.Fatalf("UpdateEvent() returned error: %v", err)
		}
		if len(outbox.entries) != 0 {
			t.Fatalf("reminder intents = %#v, want none", outbox.entries)
		}
	})

	t.Run("delete publishes cancel without schedule fields", func(t *testing.T) {
		outbox := &eventReminderIntentOutboxStub{}
		service := newReminderIntentCommandService(newEventCommandStoreStub(), outbox)

		deleted, err := service.DeleteEvent(context.Background(), 5, 77)
		if err != nil || !deleted {
			t.Fatalf("DeleteEvent() = deleted:%t err:%v, want true/nil", deleted, err)
		}
		if len(outbox.entries) != 1 {
			t.Fatalf("reminder intents = %#v, want one cancel", outbox.entries)
		}
		entry := outbox.entries[0]
		if entry.EventID != 77 || entry.Operation != servicesevents.EventReminderOperationCancel ||
			entry.StartTime != nil || entry.ReminderOffsetMinutes != nil || entry.MessageID == "" ||
			entry.SchemaVersion != servicesevents.EventReminderIntentSchemaVersion ||
			entry.IntentType != servicesevents.EventReminderIntentType || entry.OccurredAt.IsZero() {
			t.Fatalf("cancel intent = %#v", entry)
		}
	})
}

func TestEventCommandServiceTreatsReminderIntentAsRequiredTransactionalWrite(t *testing.T) {
	injected := errors.New("injected reminder outbox failure")
	tests := []struct {
		name string
		run  func(servicesevents.EventCommandService) (bool, error)
	}{
		{
			name: "create",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				result, err := service.CreateEvent(context.Background(), 5, validCreateEventCommandInput())
				return result != nil, err
			},
		},
		{
			name: "reschedule",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				start := time.Now().Add(96 * time.Hour)
				result, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{StartTime: &start})
				return result != nil, err
			},
		},
		{
			name: "delete",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				return service.DeleteEvent(context.Background(), 5, 77)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventCommandStoreStub()
			audit := &eventAuditStoreStub{}
			outbox := &eventReminderIntentOutboxStub{err: injected}
			service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
				commands:  store,
				audit:     audit,
				reminders: outbox,
			})

			succeeded, err := test.run(service)

			if succeeded || !errors.Is(err, injected) {
				t.Fatalf("command = success:%t err:%v, want false/injected error", succeeded, err)
			}
			if len(outbox.entries) != 1 {
				t.Fatalf("reminder outbox attempts = %d, want 1", len(outbox.entries))
			}
			if len(audit.entries) != 0 || store.loadedEventID != 0 {
				t.Fatalf("post-outbox side effects: audit=%d loadedEventID=%d", len(audit.entries), store.loadedEventID)
			}
			if test.name == "delete" && len(store.deletedIDs) != 0 {
				t.Fatalf("delete ran after reminder cancel failure: %v", store.deletedIDs)
			}
		})
	}
}

func TestGORMEventReminderIntentOutboxIsOrderedReplayableAndTransactional(t *testing.T) {
	db := newEventsServiceDB(t)
	service := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)

	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	genreID := seedEventGenre(t, db, "Reminder outbox")
	input := validCreateEventCommandInput()
	input.GroupID = groupID
	input.EventTypeID = 1
	input.LocationID = 1
	input.AgeLimitID = 1
	input.Genres = []uint{genreID}

	created, err := service.CreateEvent(context.Background(), 1, input)
	if err != nil {
		t.Fatalf("CreateEvent() returned error: %v", err)
	}
	newStart := input.StartTime.Add(24 * time.Hour)
	if _, err := service.UpdateEvent(context.Background(), 1, created.ID, servicesevents.UpdateEventInput{StartTime: &newStart}); err != nil {
		t.Fatalf("reschedule returned error: %v", err)
	}
	duration := uint16(180)
	if _, err := service.UpdateEvent(context.Background(), 1, created.ID, servicesevents.UpdateEventInput{Duration: &duration}); err != nil {
		t.Fatalf("duration-only update returned error: %v", err)
	}
	if deleted, err := service.DeleteEvent(context.Background(), 1, created.ID); err != nil || !deleted {
		t.Fatalf("DeleteEvent() = deleted:%t err:%v", deleted, err)
	}

	reader := servicesevents.NewGORMEventReminderIntentOutboxReader(&testPostgresRepository{db: db})
	first, err := reader.ListEventReminderIntents(context.Background(), 0, 2)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	replay, err := reader.ListEventReminderIntents(context.Background(), 0, 2)
	if err != nil {
		t.Fatalf("replay page: %v", err)
	}
	last, err := reader.ListEventReminderIntents(context.Background(), first[1].Sequence, 2)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if len(first) != 2 || first[0].Sequence >= first[1].Sequence || !reflect.DeepEqual(first, replay) {
		t.Fatalf("first/replay pages = %#v / %#v, want two stable ordered records", first, replay)
	}
	if len(last) != 1 || last[0].Operation != servicesevents.EventReminderOperationCancel || last[0].StartTime != nil || len(last[0].ReminderOffsetMinutes) != 0 {
		t.Fatalf("exclusive last page = %#v, want cancel tombstone", last)
	}
}

func TestGORMEventCommandRollsBackMutationWhenReminderIntentWriteFails(t *testing.T) {
	db := newEventsServiceDB(t)
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}))
	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	genreID := seedEventGenre(t, db, "Reminder rollback")
	if err := db.Exec(`
		CREATE TRIGGER fail_reminder_outbox_insert
		BEFORE INSERT ON event_reminder_intent_outbox
		BEGIN
			SELECT RAISE(ABORT, 'injected reminder outbox failure');
		END
	`).Error; err != nil {
		t.Fatalf("create reminder failure trigger: %v", err)
	}
	input := validCreateEventCommandInput()
	input.GroupID = groupID
	input.EventTypeID = 1
	input.LocationID = 1
	input.AgeLimitID = 1
	input.Genres = []uint{genreID}

	result, err := service.CreateEvent(context.Background(), 1, input)

	if result != nil || err == nil || !strings.Contains(err.Error(), "injected reminder outbox failure") {
		t.Fatalf("CreateEvent() = result:%#v err:%v, want nil/injected failure", result, err)
	}
	assertEventCount(t, db, 0)
	assertEventParticipantTotal(t, db, 0)
	assertEventGenreTotal(t, db, 0)
	assertReminderIntentOutboxCount(t, db, 0)
	assertLifecycleScheduleOutboxCount(t, db, 0)
}

func TestGORMEventCommandRollsBackReminderIntentWhenMutationRollsBack(t *testing.T) {
	db := newEventsServiceDB(t)
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}))
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	if err := db.Exec(`
		CREATE TRIGGER fail_event_delete_after_reminder_cancel
		BEFORE DELETE ON events
		BEGIN
			SELECT RAISE(ABORT, 'injected event delete failure after reminder cancel');
		END
	`).Error; err != nil {
		t.Fatalf("create delete failure trigger: %v", err)
	}

	deleted, err := service.DeleteEvent(context.Background(), 1, eventID)

	if deleted || err == nil || !strings.Contains(err.Error(), "injected event delete failure") {
		t.Fatalf("DeleteEvent() = deleted:%t err:%v, want false/injected failure", deleted, err)
	}
	assertEventExists(t, db, eventID, true)
	assertReminderIntentOutboxCount(t, db, 0)
	assertLifecycleScheduleOutboxCount(t, db, 0)
}

func newReminderIntentCommandService(store *eventCommandStoreStub, outbox *eventReminderIntentOutboxStub) servicesevents.EventCommandService {
	return servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
		commands:  store,
		audit:     &eventAuditStoreStub{},
		reminders: outbox,
	})
}

func assertReminderIntentUpsert(t *testing.T, entries []servicesevents.EventReminderIntentOutboxInput, eventID uint, startTime time.Time) {
	t.Helper()
	if len(entries) != 1 {
		t.Fatalf("reminder intents = %#v, want one upsert", entries)
	}
	entry := entries[0]
	wantOffsets := []int{
		servicesevents.EventReminderOffset24Hours,
		servicesevents.EventReminderOffset6Hours,
		servicesevents.EventReminderOffset1Hour,
	}
	if entry.EventID != eventID || entry.Operation != servicesevents.EventReminderOperationUpsert ||
		entry.StartTime == nil || !entry.StartTime.Equal(startTime) ||
		!reflect.DeepEqual(entry.ReminderOffsetMinutes, wantOffsets) || entry.MessageID == "" ||
		entry.SchemaVersion != servicesevents.EventReminderIntentSchemaVersion ||
		entry.IntentType != servicesevents.EventReminderIntentType || entry.OccurredAt.IsZero() {
		t.Fatalf("upsert intent = %#v, want event/start/default offsets/v1", entry)
	}
}

func assertReminderIntentOutboxCount(t *testing.T, db interface{ Model(interface{}) *gorm.DB }, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(&eventmodels.EventReminderIntentOutbox{}).Count(&count).Error; err != nil {
		t.Fatalf("count reminder intent outbox: %v", err)
	}
	if count != want {
		t.Fatalf("reminder intent outbox count = %d, want %d", count, want)
	}
}

var _ servicesevents.EventReminderIntentOutboxWriter = (*eventReminderIntentOutboxStub)(nil)
