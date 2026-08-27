package tests

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"

	"gorm.io/gorm"
)

func TestEventCommandServiceWritesLifecycleScheduleOutboxForCreateUpdateAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("create writes schedule upsert", func(t *testing.T) {
		store := newEventCommandStoreStub()
		outbox := &eventLifecycleScheduleOutboxStub{}
		service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
			commands: store,
			audit:    &eventAuditStoreStub{},
			outbox:   outbox,
		})
		input := validCreateEventCommandInput()

		if _, err := service.CreateEvent(context.Background(), 5, input); err != nil {
			t.Fatalf("CreateEvent() returned error: %v", err)
		}
		assertLifecycleScheduleUpsert(t, outbox.entries, 77, input.StartTime, input.StartTime.Add(time.Duration(input.Duration)*time.Minute))
	})

	t.Run("schedule update writes newer upsert", func(t *testing.T) {
		store := newEventCommandStoreStub()
		outbox := &eventLifecycleScheduleOutboxStub{}
		service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
			commands: store,
			audit:    &eventAuditStoreStub{},
			outbox:   outbox,
		})
		start := time.Now().Add(72 * time.Hour).Truncate(time.Second)
		duration := uint16(240)

		if _, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{
			StartTime: &start,
			Duration:  &duration,
		}); err != nil {
			t.Fatalf("UpdateEvent() returned error: %v", err)
		}
		assertLifecycleScheduleUpsert(t, outbox.entries, 77, start, start.Add(4*time.Hour))
	})

	t.Run("non schedule update writes nothing", func(t *testing.T) {
		store := newEventCommandStoreStub()
		outbox := &eventLifecycleScheduleOutboxStub{}
		service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
			commands: store,
			audit:    &eventAuditStoreStub{},
			outbox:   outbox,
		})
		title := "Only title changed"

		if _, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{Title: &title}); err != nil {
			t.Fatalf("UpdateEvent() returned error: %v", err)
		}
		if len(outbox.entries) != 0 {
			t.Fatalf("outbox entries = %#v, want none", outbox.entries)
		}
	})

	t.Run("delete writes schedule cancel", func(t *testing.T) {
		store := newEventCommandStoreStub()
		outbox := &eventLifecycleScheduleOutboxStub{}
		service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
			commands: store,
			audit:    &eventAuditStoreStub{},
			outbox:   outbox,
		})

		deleted, err := service.DeleteEvent(context.Background(), 5, 77)
		if err != nil || !deleted {
			t.Fatalf("DeleteEvent() = deleted:%t error:%v, want true/nil", deleted, err)
		}
		if len(outbox.entries) != 1 {
			t.Fatalf("outbox entries = %#v, want one cancel", outbox.entries)
		}
		entry := outbox.entries[0]
		if entry.EventID != 77 || entry.Operation != servicesevents.LifecycleScheduleOperationCancel ||
			entry.StartTime != nil || entry.EndTime != nil || entry.MessageID == "" || entry.OccurredAt.IsZero() ||
			entry.SchemaVersion != servicesevents.LifecycleScheduleSchemaVersion {
			t.Fatalf("cancel entry = %#v", entry)
		}
	})
}

func TestEventCommandServiceTreatsLifecycleOutboxAsRequiredTransactionalWrite(t *testing.T) {
	t.Parallel()

	injected := errors.New("injected lifecycle outbox failure")
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
			name: "schedule update",
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
			outbox := &eventLifecycleScheduleOutboxStub{err: injected}
			service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
				commands: store,
				audit:    audit,
				outbox:   outbox,
			})

			succeeded, err := test.run(service)

			if succeeded || !errors.Is(err, injected) {
				t.Fatalf("command = success:%t error:%v, want false/injected error", succeeded, err)
			}
			if len(outbox.entries) != 1 {
				t.Fatalf("outbox attempts = %d, want 1", len(outbox.entries))
			}
			if len(audit.entries) != 0 || store.loadedEventID != 0 {
				t.Fatalf("post-outbox side effects: audit=%d loadedEventID=%d", len(audit.entries), store.loadedEventID)
			}
			if test.name == "delete" && len(store.deletedIDs) != 0 {
				t.Fatalf("delete mutation ran after cancel outbox failure: %v", store.deletedIDs)
			}
		})
	}
}

func TestGORMEventLifecycleScheduleOutboxPersistsOrderedCreateRescheduleAndCancel(t *testing.T) {
	db := newEventsServiceDB(t)
	service := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)

	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	genreID := seedEventGenre(t, db, "Lifecycle outbox")
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
		t.Fatalf("schedule UpdateEvent() returned error: %v", err)
	}
	title := "Irrelevant update"
	if _, err := service.UpdateEvent(context.Background(), 1, created.ID, servicesevents.UpdateEventInput{Title: &title}); err != nil {
		t.Fatalf("non-schedule UpdateEvent() returned error: %v", err)
	}
	if deleted, err := service.DeleteEvent(context.Background(), 1, created.ID); err != nil || !deleted {
		t.Fatalf("DeleteEvent() = deleted:%t error:%v", deleted, err)
	}

	var records []eventmodels.EventLifecycleScheduleOutbox
	if err := db.Order("sequence ASC").Find(&records).Error; err != nil {
		t.Fatalf("load lifecycle outbox: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("outbox records = %#v, want create/upsert, reschedule/upsert, delete/cancel", records)
	}
	if records[0].Sequence >= records[1].Sequence || records[1].Sequence >= records[2].Sequence {
		t.Fatalf("sequences are not strictly increasing: %d, %d, %d", records[0].Sequence, records[1].Sequence, records[2].Sequence)
	}
	if records[0].Operation != servicesevents.LifecycleScheduleOperationUpsert ||
		records[1].Operation != servicesevents.LifecycleScheduleOperationUpsert ||
		records[2].Operation != servicesevents.LifecycleScheduleOperationCancel {
		t.Fatalf("outbox operations = %q, %q, %q", records[0].Operation, records[1].Operation, records[2].Operation)
	}
	if records[0].MessageID == records[1].MessageID || records[1].MessageID == records[2].MessageID || records[0].MessageID == "" {
		t.Fatalf("message IDs are not unique and non-empty: %#v", records)
	}
	if records[2].EventID != created.ID || records[2].StartTime != nil || records[2].EndTime != nil {
		t.Fatalf("cancel tombstone = %#v", records[2])
	}
}

func TestGORMEventCommandRollsBackMutationWhenLifecycleOutboxWriteFails(t *testing.T) {
	db := newEventsServiceDB(t)
	service := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)

	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	genreID := seedEventGenre(t, db, "Outbox failure")
	if err := db.Exec(`
		CREATE TRIGGER fail_lifecycle_outbox_insert
		BEFORE INSERT ON event_lifecycle_schedule_outbox
		BEGIN
			SELECT RAISE(ABORT, 'injected lifecycle outbox failure');
		END
	`).Error; err != nil {
		t.Fatalf("create outbox failure trigger: %v", err)
	}
	input := validCreateEventCommandInput()
	input.GroupID = groupID
	input.EventTypeID = 1
	input.LocationID = 1
	input.AgeLimitID = 1
	input.Genres = []uint{genreID}

	result, err := service.CreateEvent(context.Background(), 1, input)

	if result != nil || err == nil || !strings.Contains(err.Error(), "injected lifecycle outbox failure") {
		t.Fatalf("CreateEvent() = result:%#v error:%v, want nil/injected failure", result, err)
	}
	assertEventCount(t, db, 0)
	assertEventParticipantTotal(t, db, 0)
	assertEventGenreTotal(t, db, 0)
	assertLifecycleScheduleOutboxCount(t, db, 0)
}

func TestGORMEventCommandRollsBackOutboxWhenMutationRollsBack(t *testing.T) {
	db := newEventsServiceDB(t)
	service := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	if err := db.Exec(`
		CREATE TRIGGER fail_event_delete_after_cancel
		BEFORE DELETE ON events
		BEGIN
			SELECT RAISE(ABORT, 'injected event delete failure');
		END
	`).Error; err != nil {
		t.Fatalf("create event delete failure trigger: %v", err)
	}

	deleted, err := service.DeleteEvent(context.Background(), 1, eventID)

	if deleted || err == nil || !strings.Contains(err.Error(), "injected event delete failure") {
		t.Fatalf("DeleteEvent() = deleted:%t error:%v, want false/injected failure", deleted, err)
	}
	assertEventExists(t, db, eventID, true)
	assertLifecycleScheduleOutboxCount(t, db, 0)
}

func TestGORMEventLifecycleScheduleReaderOrdersExcludesCursorAndReplays(t *testing.T) {
	db := newEventsServiceDB(t)
	now := time.Date(2036, 8, 27, 12, 0, 0, 0, time.UTC)
	for _, record := range []eventmodels.EventLifecycleScheduleOutbox{
		{MessageID: "00000000-0000-0000-0000-000000000001", EventID: 1, Operation: servicesevents.LifecycleScheduleOperationCancel, OccurredAt: now, SchemaVersion: 1},
		{MessageID: "00000000-0000-0000-0000-000000000002", EventID: 2, Operation: servicesevents.LifecycleScheduleOperationCancel, OccurredAt: now, SchemaVersion: 1},
		{MessageID: "00000000-0000-0000-0000-000000000003", EventID: 3, Operation: servicesevents.LifecycleScheduleOperationCancel, OccurredAt: now, SchemaVersion: 1},
	} {
		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("seed outbox: %v", err)
		}
	}
	reader := servicesevents.NewGORMEventLifecycleScheduleOutboxReader(&testPostgresRepository{db: db})

	first, err := reader.ListLifecycleScheduleEvents(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("first read: %v", err)
	}
	replayed, err := reader.ListLifecycleScheduleEvents(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("replay read: %v", err)
	}
	if len(first) != 2 || first[0].Sequence != 2 || first[1].Sequence != 3 {
		t.Fatalf("exclusive ordered page = %#v, want sequences 2,3", first)
	}
	if len(replayed) != 2 || replayed[0] != first[0] || replayed[1] != first[1] {
		t.Fatalf("replayed page = %#v, want stable %#v", replayed, first)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := reader.ListLifecycleScheduleEvents(cancelled, 0, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read error = %v, want context.Canceled", err)
	}
}

func assertLifecycleScheduleUpsert(
	t *testing.T,
	entries []servicesevents.EventLifecycleScheduleOutboxInput,
	eventID uint,
	startTime time.Time,
	endTime time.Time,
) {
	t.Helper()
	if len(entries) != 1 {
		t.Fatalf("outbox entries = %#v, want one upsert", entries)
	}
	entry := entries[0]
	if entry.EventID != eventID || entry.Operation != servicesevents.LifecycleScheduleOperationUpsert ||
		entry.StartTime == nil || !entry.StartTime.Equal(startTime) ||
		entry.EndTime == nil || !entry.EndTime.Equal(endTime) ||
		entry.MessageID == "" || entry.OccurredAt.IsZero() ||
		entry.SchemaVersion != servicesevents.LifecycleScheduleSchemaVersion {
		t.Fatalf("upsert entry = %#v", entry)
	}
}

func assertLifecycleScheduleOutboxCount(t *testing.T, db interface {
	Model(interface{}) *gorm.DB
}, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(&eventmodels.EventLifecycleScheduleOutbox{}).Count(&count).Error; err != nil {
		t.Fatalf("count lifecycle outbox: %v", err)
	}
	if count != want {
		t.Fatalf("lifecycle outbox count = %d, want %d", count, want)
	}
}

var _ servicesevents.EventLifecycleScheduleOutboxWriter = (*eventLifecycleScheduleOutboxStub)(nil)
