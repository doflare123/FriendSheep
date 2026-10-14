package statistics

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

func postgresFixture(t *testing.T) (*sql.DB, string, uint64, time.Time) {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("NOTIFY_SERVICE_TEST_POSTGRES_DSN не задан; PostgreSQL statistics locks/leases не проверены")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, dsn, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = database.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	source := fmt.Sprintf("statistics_%d", time.Now().UnixNano())
	eventID := uint64(time.Now().UnixNano() & 0x3fffffffffffffff)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM notify_service.event_statistics_jobs WHERE event_id BETWEEN $1 AND $2`, int64(eventID), int64(eventID+10))
		_, _ = db.Exec(`DELETE FROM notify_service.event_statistics_source_receipts WHERE source_name=$1`, source)
		_, _ = db.Exec(`DELETE FROM notify_service.event_statistics_source_cursors WHERE source_name=$1`, source)
	})
	// Прошедшее время изолирует due claims от других fixtures с будущими часами.
	return db, source, eventID, time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestStatisticsPostgresDuplicatePollersAndStableEventIdentity(t *testing.T) {
	db, source, eventID, now := postgresFixture(t)
	store := NewStore(db)
	ctx := context.Background()
	message := Trigger{Sequence: 1, SchemaVersion: 1, EventID: eventID, CompletedAt: now}
	start := make(chan struct{})
	results := make(chan ApplyResult, 2)
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			result, err := NewStore(db).ApplyTriggers(ctx, source, []Trigger{message}, now)
			results <- result
			failures <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	total := 0
	for result := range results {
		total += result.AppliedCount
	}
	if total != 1 {
		t.Fatalf("duplicate jobs=%d", total)
	}
	message.Sequence = 2
	if result, err := store.ApplyTriggers(ctx, source, []Trigger{message}, now); err != nil || result.LastSequence != 2 || result.AppliedCount != 0 {
		t.Fatalf("backfill replay=%+v err=%v", result, err)
	}
	var jobs, receipts int
	if err := db.QueryRow(`SELECT count(*) FROM notify_service.event_statistics_jobs WHERE event_id=$1`, int64(eventID)).Scan(&jobs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM notify_service.event_statistics_source_receipts WHERE source_name=$1`, source).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if jobs != 1 || receipts != 2 {
		t.Fatalf("jobs=%d receipts=%d", jobs, receipts)
	}
	if cursor, err := NewStore(db).LoadCursor(ctx, source); err != nil || cursor != 2 {
		t.Fatalf("restart cursor=%d err=%v", cursor, err)
	}
}

func TestStatisticsPostgresBatchRollback(t *testing.T) {
	db, source, eventID, now := postgresFixture(t)
	store := NewStore(db)
	ctx := context.Background()
	if _, err := store.LoadCursor(ctx, source); err != nil {
		t.Fatal(err)
	}
	// Конфликт receipt второй записи происходит после первой job-вставки.
	if _, err := db.Exec(`INSERT INTO notify_service.event_statistics_source_receipts(source_name,sequence,event_id,schema_version,completed_at,received_at) VALUES($1,2,$2,1,$3,$3)`, source, int64(eventID+1), now); err != nil {
		t.Fatal(err)
	}
	messages := []Trigger{{Sequence: 1, SchemaVersion: 1, EventID: eventID, CompletedAt: now}, {Sequence: 2, SchemaVersion: 1, EventID: eventID + 1, CompletedAt: now}}
	if _, err := store.ApplyTriggers(ctx, source, messages, now); err == nil {
		t.Fatal("conflict batch committed")
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM notify_service.event_statistics_jobs WHERE event_id=$1`, int64(eventID)).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial job=%d err=%v", count, err)
	}
	if cursor, err := store.LoadCursor(ctx, source); err != nil || cursor != 0 {
		t.Fatalf("partial cursor=%d err=%v", cursor, err)
	}
}

func TestStatisticsPostgresSkipLockedLeaseRecoveryRetryAndAck(t *testing.T) {
	db, source, eventID, now := postgresFixture(t)
	store := NewStore(db)
	ctx := context.Background()
	messages := []Trigger{{Sequence: 1, SchemaVersion: 1, EventID: eventID, CompletedAt: now}, {Sequence: 2, SchemaVersion: 1, EventID: eventID + 1, CompletedAt: now}}
	if _, err := store.ApplyTriggers(ctx, source, messages, now); err != nil {
		t.Fatal(err)
	}
	lock, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	var lockedID int64
	if err = lock.QueryRow(`SELECT id FROM notify_service.event_statistics_jobs WHERE event_id=$1 FOR UPDATE`, int64(eventID)).Scan(&lockedID); err != nil {
		t.Fatal(err)
	}
	claimCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	second, err := store.ClaimDueJob(claimCtx, now, 30*time.Second)
	if err != nil || second == nil || second.EventID != eventID+1 {
		t.Fatalf("SKIP LOCKED claim=%+v err=%v", second, err)
	}
	if err = lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	first, err := store.ClaimDueJob(ctx, now, 30*time.Second)
	if err != nil || first == nil || first.EventID != eventID {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	if job, err := store.ClaimDueJob(ctx, now.Add(29*time.Second), 30*time.Second); err != nil || job != nil {
		t.Fatalf("early recovery=%+v err=%v", job, err)
	}
	recovered, err := NewStore(db).ClaimDueJob(ctx, now.Add(30*time.Second), 30*time.Second)
	if err != nil || recovered == nil || recovered.ID != first.ID || recovered.LeaseToken == first.LeaseToken {
		t.Fatalf("recovery=%+v err=%v", recovered, err)
	}
	if err = store.Complete(ctx, *first, ProcessResult{EventID: eventID, Outcome: "processed", ProcessedAt: &now}, now); err == nil {
		t.Fatal("stale lease completed")
	}
	if err = store.Retry(ctx, *recovered, now.Add(time.Minute), "timeout", now.Add(30*time.Second)); err != nil {
		t.Fatal(err)
	}
	// Второй job завершён отдельно, чтобы persisted retry оставался единственным due.
	if err = store.Complete(ctx, *second, ProcessResult{EventID: eventID + 1, Outcome: "missing"}, now); err != nil {
		t.Fatal(err)
	}
	if job, err := store.ClaimDueJob(ctx, now.Add(59*time.Second), 30*time.Second); err != nil || job != nil {
		t.Fatalf("early retry=%+v err=%v", job, err)
	}
	retry, err := NewStore(db).ClaimDueJob(ctx, now.Add(time.Minute), 30*time.Second)
	if err != nil || retry == nil || retry.AttemptCount != 1 {
		t.Fatalf("retry=%+v err=%v", retry, err)
	}
	if err = store.Complete(ctx, *retry, ProcessResult{EventID: eventID, Outcome: "already_processed", ProcessedAt: &now}, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if job, err := store.ClaimDueJob(ctx, now.Add(time.Hour), 30*time.Second); err != nil || job != nil {
		t.Fatalf("terminal job claim=%+v err=%v", job, err)
	}
}

func TestStatisticsPostgresConcurrentWorkersClaimOneJob(t *testing.T) {
	db, source, eventID, now := postgresFixture(t)
	ctx := context.Background()
	store := NewStore(db)
	if _, err := store.ApplyTriggers(ctx, source, []Trigger{{Sequence: 1, SchemaVersion: 1, EventID: eventID, CompletedAt: now}}, now); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	claims := make(chan *Job, 2)
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			job, err := NewStore(db).ClaimDueJob(ctx, now, time.Minute)
			claims <- job
			failures <- err
		}()
	}
	close(start)
	wg.Wait()
	close(claims)
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	for job := range claims {
		if job != nil {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("claims=%d", count)
	}
}
