package lifecycle

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"notify_service/internal/database"
)

func TestStorePostgresIngestionIsOrderedIdempotentAndTransactional(t *testing.T) {
	db := requireLifecyclePostgres(t)
	now := time.Date(2036, 8, 27, 12, 0, 0, 0, time.UTC)
	source := uniqueLifecycleSource(t)
	eventID := uniqueLifecycleEventID()
	missingEventID := eventID + 1
	rolledBackEventID := eventID + 2
	cleanupLifecycleFixture(t, db, source, eventID, missingEventID, rolledBackEventID)
	store := NewStore(db)

	first := validScheduleEvent(1, eventID, OperationScheduleUpsert, now)
	result, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{first}, now)
	if err != nil || result.LastSequence != 1 || result.AppliedCount != 1 {
		t.Fatalf("apply first upsert = result:%#v error:%v", result, err)
	}
	assertPostgresJob(t, db, eventID, 1, StateScheduled, first.StartTime, first.EndTime, 0, nil, "")

	replayed, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{first}, now.Add(time.Second))
	if err != nil || replayed.LastSequence != 1 || replayed.AppliedCount != 0 {
		t.Fatalf("replay first upsert = result:%#v error:%v", replayed, err)
	}

	updated := validScheduleEvent(2, eventID, OperationScheduleUpsert, now.Add(time.Hour))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{updated}, now.Add(time.Second)); err != nil {
		t.Fatalf("apply reschedule: %v", err)
	}
	assertPostgresJob(t, db, eventID, 2, StateScheduled, updated.StartTime, updated.EndTime, 0, nil, "")

	cancelled := validScheduleEvent(3, eventID, OperationScheduleCancel, now.Add(2*time.Hour))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{cancelled}, now.Add(2*time.Second)); err != nil {
		t.Fatalf("apply cancel: %v", err)
	}
	assertPostgresJob(t, db, eventID, 3, StateCancelled, nil, nil, 0, nil, "")

	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{updated}, now.Add(3*time.Second)); err != nil {
		t.Fatalf("apply stale upsert after cancel: %v", err)
	}
	assertPostgresJob(t, db, eventID, 3, StateCancelled, nil, nil, 0, nil, "")

	resurrected := validScheduleEvent(4, eventID, OperationScheduleUpsert, now.Add(3*time.Hour))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{resurrected}, now.Add(4*time.Second)); err != nil {
		t.Fatalf("apply newer upsert after cancel: %v", err)
	}
	assertPostgresJob(t, db, eventID, 4, StateScheduled, resurrected.StartTime, resurrected.EndTime, 0, nil, "")

	missingCancel := validScheduleEvent(5, missingEventID, OperationScheduleCancel, now.Add(4*time.Hour))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{missingCancel}, now.Add(5*time.Second)); err != nil {
		t.Fatalf("apply cancel tombstone for missing job: %v", err)
	}
	assertPostgresJob(t, db, missingEventID, 5, StateCancelled, nil, nil, 0, nil, "")

	duplicateMessage := validScheduleEvent(6, missingEventID+10, OperationScheduleCancel, now.Add(5*time.Hour))
	duplicateMessage.MessageID = first.MessageID
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{duplicateMessage}, now.Add(6*time.Second)); err == nil {
		t.Fatal("message_id reused with a different sequence was accepted")
	}
	assertPostgresCursor(t, store, source, 5)

	rolledBackFirst := validScheduleEvent(6, rolledBackEventID, OperationScheduleUpsert, now.Add(6*time.Hour))
	rolledBackSecond := validScheduleEvent(7, rolledBackEventID+1, OperationScheduleCancel, now.Add(7*time.Hour))
	rolledBackSecond.MessageID = first.MessageID
	if _, err := store.ApplySourceEvents(
		context.Background(), source, []ScheduleEvent{rolledBackFirst, rolledBackSecond}, now.Add(7*time.Second),
	); err == nil {
		t.Fatal("batch with duplicate source receipt unexpectedly committed")
	}
	assertPostgresCursor(t, store, source, 5)
	assertPostgresJobMissing(t, db, rolledBackEventID)
	assertPostgresReceiptMissing(t, db, source, rolledBackFirst.MessageID)

	restartedStore := NewStore(db)
	assertPostgresCursor(t, restartedStore, source, 5)
}

func TestStorePostgresCursorLockSerializesDuplicatePollers(t *testing.T) {
	db := requireLifecyclePostgres(t)
	now := time.Date(2036, 8, 27, 12, 0, 0, 0, time.UTC)
	source := uniqueLifecycleSource(t)
	eventID := uniqueLifecycleEventID()
	cleanupLifecycleFixture(t, db, source, eventID)
	store := NewStore(db)
	item := validScheduleEvent(1, eventID, OperationScheduleUpsert, now)

	start := make(chan struct{})
	results := make(chan ApplyResult, 2)
	errorsChannel := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			result, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{item}, now)
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
		t.Fatalf("total applied count across duplicate pollers = %d, want 1", applied)
	}
	assertPostgresCursor(t, store, source, 1)
	assertPostgresJob(t, db, eventID, 1, StateScheduled, item.StartTime, item.EndTime, 0, nil, "")
}

func TestStorePostgresSkipLockedLeaseAndCrashRecovery(t *testing.T) {
	db := requireLifecyclePostgres(t)
	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	source := uniqueLifecycleSource(t)
	eventID := uniqueLifecycleEventID()
	cleanupLifecycleFixture(t, db, source, eventID)
	store := NewStore(db)
	item := validScheduleEvent(1, eventID, OperationScheduleUpsert, now.Add(-time.Hour))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{item}, now); err != nil {
		t.Fatalf("seed due job: %v", err)
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
			job, err := NewStore(db).ClaimDueJob(context.Background(), now, 30*time.Second)
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
		t.Fatalf("claimed jobs = %d, want exactly 1", claimCount)
	}
	if claimed.State != StateProcessing || claimed.LeaseToken == "" || claimed.LeaseUntil == nil || !claimed.LeaseUntil.Equal(now.Add(30*time.Second)) {
		t.Fatalf("claimed job = %#v", claimed)
	}

	if job, err := store.ClaimDueJob(context.Background(), now.Add(29*time.Second), 30*time.Second); err != nil || job != nil {
		t.Fatalf("claim before lease expiry = job:%#v error:%v, want nil/nil", job, err)
	}
	recovered, err := NewStore(db).ClaimDueJob(context.Background(), now.Add(30*time.Second), 30*time.Second)
	if err != nil || recovered == nil {
		t.Fatalf("claim at lease expiry = job:%#v error:%v, want recovered job", recovered, err)
	}
	if recovered.ID != claimed.ID || recovered.LeaseToken == claimed.LeaseToken {
		t.Fatalf("recovered job = %#v, original = %#v", recovered, claimed)
	}

	staleLease := *claimed
	if err := store.MarkTerminal(context.Background(), staleLease, StateTerminalFailed, ErrorCodeEventNotFound, now.Add(31*time.Second)); err == nil {
		t.Fatal("stale lease terminal ack was silently accepted")
	}
}

func TestStorePostgresPersistsStartEndRetryTerminalAndCancelledStates(t *testing.T) {
	db := requireLifecyclePostgres(t)
	now := time.Date(2036, 8, 27, 18, 0, 0, 0, time.UTC)
	source := uniqueLifecycleSource(t)
	eventID := uniqueLifecycleEventID()
	cancelledEventID := eventID + 1
	cleanupLifecycleFixture(t, db, source, eventID, cancelledEventID)
	store := NewStore(db)
	item := validScheduleEvent(1, eventID, OperationScheduleUpsert, now.Add(-time.Hour))
	item.EndTime = lifecycleTimePointer(now.Add(2 * time.Hour))
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{item}, now); err != nil {
		t.Fatalf("seed due job: %v", err)
	}

	claimed, err := store.ClaimDueJob(context.Background(), now, 30*time.Second)
	if err != nil || claimed == nil {
		t.Fatalf("claim start job = job:%#v error:%v", claimed, err)
	}
	started := testAdvanceResult(OutcomeStarted, now)
	started.StartTime = *item.StartTime
	started.EndTime = *item.EndTime
	if err := store.RescheduleAfterSuccess(context.Background(), *claimed, started, now); err != nil {
		t.Fatalf("persist started outcome: %v", err)
	}
	assertPostgresJob(t, db, eventID, 1, StateActiveWait, &started.StartTime, &started.EndTime, 0, nil, "")
	if job, err := NewStore(db).ClaimDueJob(context.Background(), now.Add(time.Hour), 30*time.Second); err != nil || job != nil {
		t.Fatalf("claim active_wait before end = job:%#v error:%v", job, err)
	}

	endClaim, err := NewStore(db).ClaimDueJob(context.Background(), started.EndTime, 30*time.Second)
	if err != nil || endClaim == nil {
		t.Fatalf("claim active_wait at end = job:%#v error:%v", endClaim, err)
	}
	nextAttempt := now.Add(3 * time.Hour)
	if err := store.RescheduleRetry(context.Background(), *endClaim, nextAttempt, ErrorCodeNetwork, started.EndTime); err != nil {
		t.Fatalf("persist retry: %v", err)
	}
	assertPostgresJob(t, db, eventID, 1, StateRetryWait, &started.StartTime, &started.EndTime, 1, &nextAttempt, ErrorCodeNetwork)
	if job, err := NewStore(db).ClaimDueJob(context.Background(), nextAttempt.Add(-time.Second), 30*time.Second); err != nil || job != nil {
		t.Fatalf("claim retry before next_attempt = job:%#v error:%v", job, err)
	}
	retryClaim, err := NewStore(db).ClaimDueJob(context.Background(), nextAttempt, 30*time.Second)
	if err != nil || retryClaim == nil || retryClaim.AttemptCount != 1 {
		t.Fatalf("claim persisted retry = job:%#v error:%v", retryClaim, err)
	}
	if err := store.MarkTerminal(context.Background(), *retryClaim, StateTerminalFailed, ErrorCodeEventNotFound, nextAttempt); err != nil {
		t.Fatalf("persist terminal event-not-found: %v", err)
	}
	assertPostgresJob(t, db, eventID, 1, StateTerminalFailed, &started.StartTime, &started.EndTime, 1, nil, ErrorCodeEventNotFound)

	cancel := validScheduleEvent(2, cancelledEventID, OperationScheduleCancel, now)
	if _, err := store.ApplySourceEvents(context.Background(), source, []ScheduleEvent{cancel}, now); err != nil {
		t.Fatalf("seed cancelled tombstone: %v", err)
	}
	if job, err := NewStore(db).ClaimDueJob(context.Background(), now.Add(24*time.Hour), 30*time.Second); err != nil || job != nil {
		t.Fatalf("cancelled job was claimable: job:%#v error:%v", job, err)
	}
}

func requireLifecyclePostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("NOTIFY_SERVICE_TEST_POSTGRES_DSN не задан; PostgreSQL lifecycle integration test пропущен (SQLite не доказывает cursor locks, SKIP LOCKED или leases)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, dsn, 5*time.Second)
	if err != nil {
		t.Fatalf("open lifecycle test PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate lifecycle test PostgreSQL: %v", err)
	}
	return db
}

func uniqueLifecycleSource(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("test_%d_%s", time.Now().UnixNano(), strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
}

func uniqueLifecycleEventID() uint64 {
	return uint64(time.Now().UnixNano()&0x3fffffffffffffff) + 1000
}

func cleanupLifecycleFixture(t *testing.T, db *sql.DB, source string, eventIDs ...uint64) {
	t.Helper()
	t.Cleanup(func() {
		for _, eventID := range eventIDs {
			_, _ = db.Exec(`DELETE FROM notify_service.event_lifecycle_jobs WHERE event_id = $1`, int64(eventID))
		}
		_, _ = db.Exec(`DELETE FROM notify_service.event_lifecycle_source_messages WHERE source_name = $1`, source)
		_, _ = db.Exec(`DELETE FROM notify_service.event_lifecycle_source_cursors WHERE source_name = $1`, source)
	})
}

func assertPostgresCursor(t *testing.T, store *Store, source string, want uint64) {
	t.Helper()
	got, err := store.LoadCursor(context.Background(), source)
	if err != nil {
		t.Fatalf("LoadCursor(%q): %v", source, err)
	}
	if got != want {
		t.Fatalf("cursor %q = %d, want %d", source, got, want)
	}
}

func assertPostgresJob(
	t *testing.T,
	db *sql.DB,
	eventID uint64,
	wantSequence uint64,
	wantState string,
	wantStart *time.Time,
	wantEnd *time.Time,
	wantAttempts int,
	wantNextAttempt *time.Time,
	wantErrorCode string,
) {
	t.Helper()
	var (
		sequence      int64
		state         string
		start         sql.NullTime
		end           sql.NullTime
		attempts      int
		nextAttempt   sql.NullTime
		lastErrorCode sql.NullString
	)
	err := db.QueryRow(`
		SELECT source_sequence, state, start_time, end_time, attempt_count, next_attempt_at, last_error_code
		FROM notify_service.event_lifecycle_jobs
		WHERE event_id = $1
	`, int64(eventID)).Scan(&sequence, &state, &start, &end, &attempts, &nextAttempt, &lastErrorCode)
	if err != nil {
		t.Fatalf("load lifecycle job %d: %v", eventID, err)
	}
	if uint64(sequence) != wantSequence || state != wantState || attempts != wantAttempts || lastErrorCode.String != wantErrorCode {
		t.Fatalf("job %d = sequence:%d state:%q attempts:%d error:%q; want %d/%q/%d/%q",
			eventID, sequence, state, attempts, lastErrorCode.String, wantSequence, wantState, wantAttempts, wantErrorCode)
	}
	assertNullTime(t, "start_time", start, wantStart)
	assertNullTime(t, "end_time", end, wantEnd)
	assertNullTime(t, "next_attempt_at", nextAttempt, wantNextAttempt)
}

func assertNullTime(t *testing.T, name string, got sql.NullTime, want *time.Time) {
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

func assertPostgresJobMissing(t *testing.T, db *sql.DB, eventID uint64) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notify_service.event_lifecycle_jobs WHERE event_id = $1`, int64(eventID)).Scan(&count); err != nil {
		t.Fatalf("count lifecycle job %d: %v", eventID, err)
	}
	if count != 0 {
		t.Fatalf("lifecycle job %d exists after rollback", eventID)
	}
}

func assertPostgresReceiptMissing(t *testing.T, db *sql.DB, source, messageID string) {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM notify_service.event_lifecycle_source_messages
		WHERE source_name = $1 AND message_id = $2
	`, source, messageID).Scan(&count); err != nil {
		t.Fatalf("count lifecycle receipt: %v", err)
	}
	if count != 0 {
		t.Fatalf("lifecycle receipt %s/%s exists after rollback", source, messageID)
	}
}

func lifecycleTimePointer(value time.Time) *time.Time {
	return &value
}
