package reminders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"notify_service/internal/database"
)

func TestStorePostgresIngestionIsOrderedIdempotentAndTransactional(t *testing.T) {
	db := requireReminderPostgres(t)
	now := time.Date(2036, 8, 31, 12, 0, 0, 0, time.UTC)
	source := uniqueReminderSource(t)
	eventID := uniqueReminderEventID()
	missingEventID := eventID + 1
	rolledBackEventID := eventID + 2
	rolledBackEventID2 := eventID + 3
	cleanupReminderFixture(t, db, source, eventID, missingEventID, rolledBackEventID, rolledBackEventID2)

	store := NewStore(db, mustReminderRegistry(t))
	start := now.Add(30 * time.Hour)

	first := validReminderIntent(1, eventID, OperationScheduleUpsert, reminderTimePointer(start), []int{60, 1440, 360}, now)
	result, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{first}, now)
	if err != nil || result.LastSequence != 1 || result.AppliedCount != 1 {
		t.Fatalf("apply first upsert = result:%#v error:%v", result, err)
	}
	assertReminderJob(t, db, eventID, 1, 1440, StateScheduled, reminderTimePointer(start), reminderTimePointer(start.Add(-24*time.Hour)), reminderTimePointer(start.Add(-6*time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 1, 360, StateScheduled, reminderTimePointer(start), reminderTimePointer(start.Add(-6*time.Hour)), reminderTimePointer(start.Add(-time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 1, 60, StateScheduled, reminderTimePointer(start), reminderTimePointer(start.Add(-time.Hour)), reminderTimePointer(start), 0, nil, "")

	replayed, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{first}, now.Add(time.Second))
	if err != nil || replayed.LastSequence != 1 || replayed.AppliedCount != 0 {
		t.Fatalf("replay first upsert = result:%#v error:%v", replayed, err)
	}

	outOfOrder := []ReminderIntent{
		validReminderIntent(3, eventID, OperationScheduleUpsert, reminderTimePointer(start.Add(2*time.Hour)), []int{60}, now.Add(2*time.Second)),
		validReminderIntent(2, eventID, OperationScheduleUpsert, reminderTimePointer(start.Add(time.Hour)), []int{60}, now.Add(3*time.Second)),
	}
	if _, err := store.ApplySourceEvents(context.Background(), source, outOfOrder, now.Add(2*time.Second)); err == nil {
		t.Fatal("out-of-order reminder batch unexpectedly committed")
	}
	assertReminderCursor(t, store, source, 1)
	assertReminderJobMissing(t, db, eventID, 2)

	updatedStart := start.Add(time.Hour)
	updated := validReminderIntent(2, eventID, OperationScheduleUpsert, reminderTimePointer(updatedStart), []int{60, 1440, 360}, now.Add(4*time.Second))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{updated}, now.Add(4*time.Second)); err != nil {
		t.Fatalf("apply reminder reschedule: %v", err)
	}
	assertReminderJob(t, db, eventID, 1, 1440, StateSuperseded, reminderTimePointer(start), reminderTimePointer(start.Add(-24*time.Hour)), reminderTimePointer(start.Add(-6*time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 2, 1440, StateScheduled, reminderTimePointer(updatedStart), reminderTimePointer(updatedStart.Add(-24*time.Hour)), reminderTimePointer(updatedStart.Add(-6*time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 2, 360, StateScheduled, reminderTimePointer(updatedStart), reminderTimePointer(updatedStart.Add(-6*time.Hour)), reminderTimePointer(updatedStart.Add(-time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 2, 60, StateScheduled, reminderTimePointer(updatedStart), reminderTimePointer(updatedStart.Add(-time.Hour)), reminderTimePointer(updatedStart), 0, nil, "")

	cancelled := validReminderIntent(3, eventID, OperationScheduleCancel, nil, nil, now.Add(5*time.Second))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{cancelled}, now.Add(5*time.Second)); err != nil {
		t.Fatalf("apply reminder cancel: %v", err)
	}
	assertReminderJob(t, db, eventID, 2, 1440, StateCancelled, reminderTimePointer(updatedStart), reminderTimePointer(updatedStart.Add(-24*time.Hour)), reminderTimePointer(updatedStart.Add(-6*time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 2, 360, StateCancelled, reminderTimePointer(updatedStart), reminderTimePointer(updatedStart.Add(-6*time.Hour)), reminderTimePointer(updatedStart.Add(-time.Hour)), 0, nil, "")
	assertReminderJob(t, db, eventID, 2, 60, StateCancelled, reminderTimePointer(updatedStart), reminderTimePointer(updatedStart.Add(-time.Hour)), reminderTimePointer(updatedStart), 0, nil, "")
	assertReminderJob(t, db, eventID, 3, 0, StateCancelled, nil, nil, nil, 0, nil, "")

	staleReplay, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{updated}, now.Add(6*time.Second))
	if err != nil || staleReplay.LastSequence != 3 || staleReplay.AppliedCount != 0 {
		t.Fatalf("stale replay after cancel = result:%#v error:%v", staleReplay, err)
	}

	resurrectedStart := start.Add(2 * time.Hour)
	resurrected := validReminderIntent(4, eventID, OperationScheduleUpsert, reminderTimePointer(resurrectedStart), []int{60}, now.Add(7*time.Second))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{resurrected}, now.Add(7*time.Second)); err != nil {
		t.Fatalf("apply newer upsert after cancel: %v", err)
	}
	assertReminderJob(t, db, eventID, 4, 60, StateScheduled, reminderTimePointer(resurrectedStart), reminderTimePointer(resurrectedStart.Add(-time.Hour)), reminderTimePointer(resurrectedStart), 0, nil, "")

	missingCancel := validReminderIntent(5, missingEventID, OperationScheduleCancel, nil, nil, now.Add(8*time.Second))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{missingCancel}, now.Add(8*time.Second)); err != nil {
		t.Fatalf("apply reminder cancel tombstone: %v", err)
	}
	assertReminderJob(t, db, missingEventID, 5, 0, StateCancelled, nil, nil, nil, 0, nil, "")

	duplicateMessage := validReminderIntent(6, missingEventID+10, OperationScheduleCancel, nil, nil, now.Add(9*time.Second))
	duplicateMessage.MessageID = first.MessageID
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{duplicateMessage}, now.Add(9*time.Second)); err == nil {
		t.Fatal("message_id reused with a different sequence was accepted")
	}
	assertReminderCursor(t, store, source, 5)

	rolledBackFirst := validReminderIntent(6, rolledBackEventID, OperationScheduleUpsert, reminderTimePointer(start.Add(3*time.Hour)), []int{60}, now.Add(10*time.Second))
	rolledBackSecond := validReminderIntent(7, rolledBackEventID2, OperationScheduleCancel, nil, nil, now.Add(11*time.Second))
	rolledBackSecond.MessageID = first.MessageID
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{rolledBackFirst, rolledBackSecond}, now.Add(10*time.Second)); err == nil {
		t.Fatal("batch with duplicate reminder source receipt unexpectedly committed")
	}
	assertReminderCursor(t, store, source, 5)
	assertReminderJobMissing(t, db, rolledBackEventID, 6)
	assertReminderReceiptMissing(t, db, source, rolledBackFirst.MessageID)

	restartedStore := NewStore(db, mustReminderRegistry(t))
	assertReminderCursor(t, restartedStore, source, 5)
}

func TestStorePostgresCursorLockSerializesDuplicatePollers(t *testing.T) {
	db := requireReminderPostgres(t)
	now := time.Date(2036, 8, 31, 12, 0, 0, 0, time.UTC)
	source := uniqueReminderSource(t)
	eventID := uniqueReminderEventID()
	cleanupReminderFixture(t, db, source, eventID)

	store := NewStore(db, mustReminderRegistry(t))
	startTime := now.Add(2 * time.Hour)
	item := validReminderIntent(1, eventID, OperationScheduleUpsert, reminderTimePointer(startTime), []int{60}, now)

	start := make(chan struct{})
	results := make(chan ApplyResult, 2)
	errorsChannel := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			result, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{item}, now)
			results <- result
			errorsChannel <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent ApplySourceEvents() returned error: %v", err)
		}
	}
	applied := 0
	for result := range results {
		applied += result.AppliedCount
	}
	if applied != 1 {
		t.Fatalf("total applied count across duplicate reminder pollers = %d, want 1", applied)
	}
	assertReminderCursor(t, store, source, 1)
	assertReminderJob(t, db, eventID, 1, 60, StateScheduled, reminderTimePointer(startTime), reminderTimePointer(startTime.Add(-time.Hour)), reminderTimePointer(startTime), 0, nil, "")
}

func TestStorePostgresSkipLockedLeaseAndCrashRecovery(t *testing.T) {
	db := requireReminderPostgres(t)
	now := time.Date(2036, 8, 31, 18, 0, 0, 0, time.UTC)
	source := uniqueReminderSource(t)
	eventID := uniqueReminderEventID()
	cleanupReminderFixture(t, db, source, eventID)

	registry := mustReminderRegistry(t)
	store := NewStore(db, registry)
	item := validReminderIntent(1, eventID, OperationScheduleUpsert, reminderTimePointer(now.Add(30*time.Minute)), []int{60}, now)
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{item}, now); err != nil {
		t.Fatalf("seed due reminder job: %v", err)
	}

	start := make(chan struct{})
	claims := make(chan *Job, 2)
	errorsChannel := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			job, err := NewStore(db, registry).ClaimDueJob(context.Background(), now, 30*time.Second)
			claims <- job
			errorsChannel <- err
		}()
	}
	close(start)
	wg.Wait()
	close(claims)
	close(errorsChannel)

	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent ClaimDueJob() returned error: %v", err)
		}
	}
	var claimed *Job
	claimCount := 0
	for job := range claims {
		if job != nil {
			claimCount++
			claimed = job
		}
	}
	if claimCount != 1 || claimed == nil {
		t.Fatalf("claimed reminder jobs = %d, want exactly 1", claimCount)
	}
	if claimed.State != StateProcessing || claimed.LeaseToken == "" || claimed.LeaseUntil == nil || !claimed.LeaseUntil.Equal(now.Add(30*time.Second)) {
		t.Fatalf("claimed reminder job = %#v", claimed)
	}

	if job, err := store.ClaimDueJob(context.Background(), now.Add(29*time.Second), 30*time.Second); err != nil || job != nil {
		t.Fatalf("claim before reminder lease expiry = job:%#v error:%v, want nil/nil", job, err)
	}
	recovered, err := NewStore(db, registry).ClaimDueJob(context.Background(), now.Add(30*time.Second), 30*time.Second)
	if err != nil || recovered == nil {
		t.Fatalf("claim at reminder lease expiry = job:%#v error:%v, want recovered job", recovered, err)
	}
	if recovered.ID != claimed.ID || recovered.LeaseToken == claimed.LeaseToken {
		t.Fatalf("recovered reminder job = %#v, original = %#v", recovered, claimed)
	}

	if err := store.MarkFinal(context.Background(), *claimed, StateTerminalFailed, ErrorCodeTimeout, now.Add(31*time.Second)); err == nil {
		t.Fatal("stale reminder lease acknowledgement was silently accepted")
	}
}

func TestStorePostgresCompleteJobPersistsInboxReadStateAndDeliveryAttempts(t *testing.T) {
	db := requireReminderPostgres(t)
	now := time.Date(2036, 8, 31, 18, 0, 0, 0, time.UTC)
	source := uniqueReminderSource(t)
	eventID := uniqueReminderEventID()
	eventID2 := eventID + 1
	cleanupReminderFixture(t, db, source, eventID, eventID2)

	store := NewStore(db, mustReminderRegistry(t))
	itemOne := validReminderIntent(1, eventID, OperationScheduleUpsert, reminderTimePointer(now.Add(30*time.Minute)), []int{60}, now)
	itemTwo := validReminderIntent(2, eventID2, OperationScheduleUpsert, reminderTimePointer(now.Add(35*time.Minute)), []int{60}, now.Add(time.Second))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{itemOne, itemTwo}, now); err != nil {
		t.Fatalf("seed due reminder jobs: %v", err)
	}

	jobOne, err := store.ClaimDueJob(context.Background(), now, 30*time.Second)
	if err != nil || jobOne == nil {
		t.Fatalf("claim first reminder job = job:%#v error:%v", jobOne, err)
	}
	snapshotOne := snapshotForReminderJob(*jobOne)
	deliveredOne, err := store.CompleteJob(context.Background(), *jobOne, snapshotOne, now)
	if err != nil {
		t.Fatalf("complete first reminder job: %v", err)
	}
	if deliveredOne != 2 {
		t.Fatalf("delivered users for first reminder job = %d, want 2", deliveredOne)
	}
	if _, err := store.CompleteJob(context.Background(), *jobOne, snapshotOne, now.Add(2*time.Second)); err == nil {
		t.Fatal("repeated CompleteJob() with stale lease unexpectedly succeeded")
	}

	jobTwo, err := store.ClaimDueJob(context.Background(), now.Add(time.Second), 30*time.Second)
	if err != nil || jobTwo == nil {
		t.Fatalf("claim second reminder job = job:%#v error:%v", jobTwo, err)
	}
	snapshotTwo := snapshotForReminderJob(*jobTwo)
	deliveredTwo, err := store.CompleteJob(context.Background(), *jobTwo, snapshotTwo, now.Add(time.Second))
	if err != nil {
		t.Fatalf("complete second reminder job: %v", err)
	}
	if deliveredTwo != 1 {
		t.Fatalf("delivered users for second reminder job = %d, want 1", deliveredTwo)
	}

	assertReminderJob(t, db, eventID, 1, 60, StateCompleted, reminderTimePointer(now.Add(30*time.Minute)), reminderTimePointer(now.Add(-30*time.Minute)), reminderTimePointer(now.Add(30*time.Minute)), 0, nil, "")
	assertReminderJob(t, db, eventID2, 2, 60, StateCompleted, reminderTimePointer(now.Add(35*time.Minute)), reminderTimePointer(now.Add(-25*time.Minute)), reminderTimePointer(now.Add(35*time.Minute)), 0, nil, "")
	assertNotificationCountForEvent(t, db, eventID, 2)
	assertNotificationCountForEvent(t, db, eventID2, 1)
	assertDeliveredAttemptCountForEvents(t, db, []uint64{eventID, eventID2}, 3)

	pageOne, err := store.ListNotifications(context.Background(), 501, "", 1, false)
	if err != nil {
		t.Fatalf("ListNotifications(first page): %v", err)
	}
	if len(pageOne.Items) != 1 || !pageOne.HasMore || pageOne.NextCursor == "" {
		t.Fatalf("first notification page = %#v", pageOne)
	}
	pageTwo, err := store.ListNotifications(context.Background(), 501, pageOne.NextCursor, 10, false)
	if err != nil {
		t.Fatalf("ListNotifications(second page): %v", err)
	}
	if len(pageTwo.Items) != 1 || pageTwo.HasMore || pageTwo.Items[0].ID == pageOne.Items[0].ID {
		t.Fatalf("second notification page = %#v", pageTwo)
	}

	unreadBefore, err := store.CountUnread(context.Background(), 501)
	if err != nil {
		t.Fatalf("CountUnread(before): %v", err)
	}
	if unreadBefore != 2 {
		t.Fatalf("CountUnread(before) = %d, want 2", unreadBefore)
	}

	readAt := now.Add(3 * time.Second)
	marked, err := store.MarkAsRead(context.Background(), 501, pageOne.Items[0].ID, readAt)
	if err != nil {
		t.Fatalf("MarkAsRead(first): %v", err)
	}
	if marked.ReadAt == nil || !marked.ReadAt.Equal(readAt) {
		t.Fatalf("MarkAsRead(first) readAt = %#v, want %s", marked.ReadAt, readAt)
	}

	markedAgain, err := store.MarkAsRead(context.Background(), 501, pageOne.Items[0].ID, readAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("MarkAsRead(second): %v", err)
	}
	if markedAgain.ReadAt == nil || !markedAgain.ReadAt.Equal(readAt) {
		t.Fatalf("MarkAsRead(second) changed readAt = %#v, want original %s", markedAgain.ReadAt, readAt)
	}

	unreadOnly, err := store.ListNotifications(context.Background(), 501, "", 10, true)
	if err != nil {
		t.Fatalf("ListNotifications(unreadOnly): %v", err)
	}
	if len(unreadOnly.Items) != 1 || unreadOnly.Items[0].ID == pageOne.Items[0].ID {
		t.Fatalf("unread-only page = %#v", unreadOnly)
	}

	unreadAfter, err := store.CountUnread(context.Background(), 501)
	if err != nil {
		t.Fatalf("CountUnread(after): %v", err)
	}
	if unreadAfter != 1 {
		t.Fatalf("CountUnread(after) = %d, want 1", unreadAfter)
	}

	_, err = store.MarkAsRead(context.Background(), 502, pageOne.Items[0].ID, readAt)
	var clientErr *ClientError
	if !errors.As(err, &clientErr) || clientErr.Code != ErrorCodeNotificationNotFound {
		t.Fatalf("cross-user MarkAsRead error = %#v", err)
	}
}

func TestStorePostgresDeliveryRollbackBeforeJobAckIsSafelyReplayable(t *testing.T) {
	db := requireReminderPostgres(t)
	now := time.Date(2036, 8, 31, 18, 0, 0, 0, time.UTC)
	source := uniqueReminderSource(t)
	eventID := uniqueReminderEventID()
	cleanupReminderFixture(t, db, source, eventID)

	store := NewStore(db, mustReminderRegistry(t))
	intent := validReminderIntent(1, eventID, OperationScheduleUpsert, reminderTimePointer(now.Add(30*time.Minute)), []int{60}, now)
	if _, err := store.ApplySourceEvents(context.Background(), source, []ReminderIntent{intent}, now); err != nil {
		t.Fatalf("seed crash-recovery reminder job: %v", err)
	}
	job, err := store.ClaimDueJob(context.Background(), now, 30*time.Second)
	if err != nil || job == nil {
		t.Fatalf("claim crash-recovery reminder job = job:%#v error:%v", job, err)
	}

	functionName := fmt.Sprintf("fail_reminder_complete_%d", eventID)
	triggerName := fmt.Sprintf("fail_reminder_complete_%d", eventID)
	createFunction := fmt.Sprintf(`
		CREATE FUNCTION notify_service.%s() RETURNS trigger AS $$
		BEGIN
			IF NEW.event_id = %d AND NEW.state = 'completed' THEN
				RAISE EXCEPTION 'injected failure before reminder job acknowledgement';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`, functionName, eventID)
	if _, err := db.Exec(createFunction); err != nil {
		t.Fatalf("create reminder failure function: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s
		BEFORE UPDATE ON notify_service.event_reminder_jobs
		FOR EACH ROW EXECUTE FUNCTION notify_service.%s()
	`, triggerName, functionName)); err != nil {
		t.Fatalf("create reminder failure trigger: %v", err)
	}
	dropFailureInjection := func() {
		_, _ = db.Exec(fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON notify_service.event_reminder_jobs`, triggerName))
		_, _ = db.Exec(fmt.Sprintf(`DROP FUNCTION IF EXISTS notify_service.%s()`, functionName))
	}
	t.Cleanup(dropFailureInjection)

	snapshot := RecipientSnapshot{
		EventID: eventID, Title: "Проверка атомарности", StartTime: *job.StartTime,
		ReminderOffsetMinutes: 60,
		Recipients:            []Recipient{{UserID: 601, Channels: []string{ChannelCodeInApp}}},
	}
	if _, err := store.CompleteJob(context.Background(), *job, snapshot, now); err == nil {
		t.Fatal("CompleteJob() неожиданно зафиксировал транзакцию при ошибке job ack")
	}
	assertNotificationCountForEvent(t, db, eventID, 0)
	assertDeliveredAttemptCountForEvents(t, db, []uint64{eventID}, 0)

	dropFailureInjection()
	if delivered, err := store.CompleteJob(context.Background(), *job, snapshot, now.Add(time.Second)); err != nil || delivered != 1 {
		t.Fatalf("replayed CompleteJob() = delivered:%d error:%v, want 1/nil", delivered, err)
	}
	assertNotificationCountForEvent(t, db, eventID, 1)
	assertDeliveredAttemptCountForEvents(t, db, []uint64{eventID}, 1)

	page, err := store.ListNotifications(context.Background(), 601, "", 10, false)
	if err != nil || len(page.Items) != 1 || page.Items[0].Payload.SchemaVersion != 1 {
		t.Fatalf("replayed inbox page = %#v error:%v", page, err)
	}
}

func requireReminderPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("NOTIFY_SERVICE_TEST_POSTGRES_DSN не задан; PostgreSQL reminder integration test пропущен (SQLite не доказывает cursor locks, SKIP LOCKED, leases или inbox идемпотентность)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, dsn, 5*time.Second)
	if err != nil {
		t.Fatalf("open reminder test PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate reminder test PostgreSQL: %v", err)
	}
	return db
}

func uniqueReminderSource(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("test_%d_%s", time.Now().UnixNano(), strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
}

func uniqueReminderEventID() uint64 {
	return uint64(time.Now().UnixNano()&0x3fffffffffffffff) + 5000
}

func cleanupReminderFixture(t *testing.T, db *sql.DB, source string, eventIDs ...uint64) {
	t.Helper()
	t.Cleanup(func() {
		for _, eventID := range eventIDs {
			_, _ = db.Exec(`DELETE FROM notify_service.notifications WHERE resource_type = $1 AND resource_id = $2`, ResourceTypeEvent, int64(eventID))
			_, _ = db.Exec(`DELETE FROM notify_service.event_reminder_jobs WHERE event_id = $1`, int64(eventID))
		}
		_, _ = db.Exec(`DELETE FROM notify_service.event_reminder_source_messages WHERE source_name = $1`, source)
		_, _ = db.Exec(`DELETE FROM notify_service.event_reminder_source_cursors WHERE source_name = $1`, source)
	})
}

func mustReminderRegistry(t *testing.T) *ChannelRegistry {
	t.Helper()
	registry, err := NewChannelRegistry(InAppChannel{})
	if err != nil {
		t.Fatalf("NewChannelRegistry(): %v", err)
	}
	return registry
}

func validReminderIntent(sequence, eventID uint64, operation string, start *time.Time, offsets []int, occurredAt time.Time) ReminderIntent {
	item := ReminderIntent{
		Sequence:      sequence,
		MessageID:     reminderMessageID(eventID, sequence),
		SchemaVersion: 1,
		IntentType:    IntentTypeEventReminder,
		Operation:     operation,
		EventID:       eventID,
		OccurredAt:    occurredAt,
	}
	if start != nil {
		startCopy := *start
		item.StartTime = &startCopy
	}
	if len(offsets) > 0 {
		item.ReminderOffsetMinutes = append([]int(nil), offsets...)
	}
	return item
}

func snapshotForReminderJob(job Job) RecipientSnapshot {
	start := time.Time{}
	if job.StartTime != nil {
		start = *job.StartTime
	}
	if job.SourceRevision == 1 {
		return RecipientSnapshot{
			EventID:               job.EventID,
			Title:                 "Первое напоминание",
			StartTime:             start,
			ReminderOffsetMinutes: job.ReminderOffsetMinutes,
			Recipients: []Recipient{
				{UserID: 501, Channels: []string{ChannelCodeInApp, ChannelCodeInApp}},
				{UserID: 501, Channels: []string{ChannelCodeInApp}},
				{UserID: 502, Channels: []string{ChannelCodeInApp}},
			},
		}
	}
	return RecipientSnapshot{
		EventID:               job.EventID,
		Title:                 "Второе напоминание",
		StartTime:             start,
		ReminderOffsetMinutes: job.ReminderOffsetMinutes,
		Recipients: []Recipient{
			{UserID: 501, Channels: []string{ChannelCodeInApp}},
		},
	}
}

func assertReminderCursor(t *testing.T, store *Store, source string, want uint64) {
	t.Helper()
	got, err := store.LoadCursor(context.Background(), source)
	if err != nil {
		t.Fatalf("LoadCursor(%q): %v", source, err)
	}
	if got != want {
		t.Fatalf("cursor %q = %d, want %d", source, got, want)
	}
}

func assertReminderJob(
	t *testing.T,
	db *sql.DB,
	eventID uint64,
	wantSequence uint64,
	wantOffset int,
	wantState string,
	wantStart *time.Time,
	wantDue *time.Time,
	wantWindow *time.Time,
	wantAttempts int,
	wantNextAttempt *time.Time,
	wantErrorCode string,
) {
	t.Helper()
	var (
		sequence      int64
		offset        int
		state         string
		start         sql.NullTime
		due           sql.NullTime
		window        sql.NullTime
		attempts      int
		nextAttempt   sql.NullTime
		lastErrorCode sql.NullString
	)
	err := db.QueryRow(`
		SELECT source_revision, reminder_offset_minutes, state, start_time, due_at, window_end_at, attempt_count, next_attempt_at, last_error_code
		FROM notify_service.event_reminder_jobs
		WHERE event_id = $1 AND source_revision = $2 AND reminder_offset_minutes = $3
	`, int64(eventID), int64(wantSequence), wantOffset).Scan(&sequence, &offset, &state, &start, &due, &window, &attempts, &nextAttempt, &lastErrorCode)
	if err != nil {
		t.Fatalf("load reminder job event=%d sequence=%d offset=%d: %v", eventID, wantSequence, wantOffset, err)
	}
	if uint64(sequence) != wantSequence || offset != wantOffset || state != wantState || attempts != wantAttempts || lastErrorCode.String != wantErrorCode {
		t.Fatalf("reminder job %d/%d/%d = sequence:%d offset:%d state:%q attempts:%d error:%q; want %d/%d/%q/%d/%q",
			eventID, wantSequence, wantOffset, sequence, offset, state, attempts, lastErrorCode.String, wantSequence, wantOffset, wantState, wantAttempts, wantErrorCode)
	}
	assertReminderNullTime(t, "start_time", start, wantStart)
	assertReminderNullTime(t, "due_at", due, wantDue)
	assertReminderNullTime(t, "window_end_at", window, wantWindow)
	assertReminderNullTime(t, "next_attempt_at", nextAttempt, wantNextAttempt)
}

func assertReminderNullTime(t *testing.T, name string, got sql.NullTime, want *time.Time) {
	t.Helper()
	if want == nil {
		if got.Valid {
			t.Fatalf("%s = %s, want NULL", name, got.Time)
		}
		return
	}
	if !got.Valid || !got.Time.Equal(*want) {
		t.Fatalf("%s = valid:%t time:%s, want %s", name, got.Valid, got.Time, *want)
	}
}

func assertReminderJobMissing(t *testing.T, db *sql.DB, eventID uint64, sourceRevision uint64) {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notify_service.event_reminder_jobs
		WHERE event_id = $1 AND source_revision = $2
	`, int64(eventID), int64(sourceRevision)).Scan(&count); err != nil {
		t.Fatalf("count reminder jobs event=%d source_revision=%d: %v", eventID, sourceRevision, err)
	}
	if count != 0 {
		t.Fatalf("reminder jobs event=%d source_revision=%d exist after rollback", eventID, sourceRevision)
	}
}

func assertReminderReceiptMissing(t *testing.T, db *sql.DB, source, messageID string) {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notify_service.event_reminder_source_messages
		WHERE source_name = $1 AND message_id = $2
	`, source, messageID).Scan(&count); err != nil {
		t.Fatalf("count reminder receipt: %v", err)
	}
	if count != 0 {
		t.Fatalf("reminder receipt %s/%s exists after rollback", source, messageID)
	}
}

func assertNotificationCountForEvent(t *testing.T, db *sql.DB, eventID uint64, want int) {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notify_service.notifications
		WHERE resource_type = $1 AND resource_id = $2
	`, ResourceTypeEvent, int64(eventID)).Scan(&count); err != nil {
		t.Fatalf("count notifications for event %d: %v", eventID, err)
	}
	if count != want {
		t.Fatalf("notification count for event %d = %d, want %d", eventID, count, want)
	}
}

func assertDeliveredAttemptCountForEvents(t *testing.T, db *sql.DB, eventIDs []uint64, want int) {
	t.Helper()
	if len(eventIDs) == 0 {
		t.Fatal("eventIDs for delivered-attempt assertion are empty")
	}
	args := make([]any, 0, len(eventIDs)+1)
	placeholders := make([]string, 0, len(eventIDs))
	for index, eventID := range eventIDs {
		args = append(args, int64(eventID))
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}
	args = append(args, DeliveryStatusDelivered)
	statusPlaceholder := fmt.Sprintf("$%d", len(args))
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM notify_service.notification_delivery_attempts attempt
		JOIN notify_service.notifications notification ON notification.id = attempt.notification_id
		WHERE notification.resource_type = '%s'
		  AND notification.resource_id IN (%s)
		  AND attempt.status = %s
	`, ResourceTypeEvent, strings.Join(placeholders, ", "), statusPlaceholder)
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("count delivered attempts: %v", err)
	}
	if count != want {
		t.Fatalf("delivered attempt count = %d, want %d", count, want)
	}
}

func reminderTimePointer(value time.Time) *time.Time {
	return &value
}

func reminderMessageID(eventID, sequence uint64) string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		uint32(sequence&0xffffffff),
		uint16((sequence>>32)&0xffff),
		uint16(eventID&0xffff),
		uint16((eventID>>16)&0xffff),
		eventID&0xffffffffffff,
	)
}
