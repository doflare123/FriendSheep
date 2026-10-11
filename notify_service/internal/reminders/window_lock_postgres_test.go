package reminders

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

type lockWindowClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *lockWindowClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *lockWindowClock) set(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
}

func TestStorePostgresWindowDecisionAfterLock(t *testing.T) {
	for _, offset := range []int{1440, 360, 60} {
		for _, delta := range []time.Duration{0, time.Nanosecond} {
			t.Run(fmt.Sprintf("offset_%d_delta_%d", offset, delta), func(t *testing.T) {
				db := requireReminderPostgres(t)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				source, eventID := uniqueReminderSource(t), uniqueReminderEventID()
				cleanupReminderFixture(t, db, source, eventID)
				store := NewStore(db, mustReminderRegistry(t))
				start := time.Date(2036, 9, 1, 12, 0, 0, 0, time.UTC)
				boundary := start
				if offset == 1440 {
					boundary = start.Add(-6 * time.Hour)
				}
				if offset == 360 {
					boundary = start.Add(-time.Hour)
				}
				before := boundary.Add(-time.Second)
				intent := validReminderIntent(1, eventID, OperationScheduleUpsert, &start, []int{1440, 360, 60}, before)
				if _, err := store.ApplySourceEvents(ctx, source, []ReminderIntent{intent}, before); err != nil {
					t.Fatal(err)
				}
				// Остальные окна не участвуют в проверке выбранной блокировки.
				if _, err := db.ExecContext(ctx, `UPDATE notify_service.event_reminder_jobs SET state = $3 WHERE event_id = $1 AND reminder_offset_minutes <> $2`, int64(eventID), offset, StateCancelled); err != nil {
					t.Fatal(err)
				}
				job, err := store.ClaimDueJob(ctx, before, time.Minute)
				if err != nil || job == nil {
					t.Fatalf("claim: %v %v", job, err)
				}
				blocker, err := db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				defer blocker.Rollback()
				var blockerPID, id int
				if err := blocker.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
					t.Fatal(err)
				}
				if err := blocker.QueryRowContext(ctx, `SELECT id FROM notify_service.event_reminder_jobs WHERE id=$1 FOR UPDATE`, job.ID).Scan(&id); err != nil {
					t.Fatal(err)
				}
				clock := &lockWindowClock{now: before}
				done := make(chan error, 1)
				go func() {
					count, err := store.MaterializeJob(ctx, *job, snapshotForReminderJob(*job), clock)
					if err == nil && count != 0 {
						err = fmt.Errorf("materialized %d recipients", count)
					}
					done <- err
				}()
				// Наблюдаем реальное ожидание нашей транзакции; задержки не служат барьером.
				for {
					var waiting bool
					if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid)))`, blockerPID).Scan(&waiting); err != nil {
						t.Fatal(err)
					}
					if waiting {
						break
					}
					select {
					case err := <-done:
						t.Fatalf("materialization finished before barrier: %v", err)
					default:
					}
					runtime.Gosched()
				}
				clock.set(boundary.Add(delta))
				if err := blocker.Commit(); err != nil {
					t.Fatal(err)
				}
				if err := <-done; err != nil {
					t.Fatal(err)
				}
				want := StateSkippedMissedWindow
				if offset == 60 {
					want = StateExpired
				}
				var state string
				if err := db.QueryRowContext(ctx, `SELECT state FROM notify_service.event_reminder_jobs WHERE id=$1`, job.ID).Scan(&state); err != nil || state != want {
					t.Fatalf("state=%s want=%s err=%v", state, want, err)
				}
				assertNotificationCountForEvent(t, db, eventID, 0)
				assertDeliveredTargetCountForEvents(t, db, []uint64{eventID}, 0)
			})
		}
	}
}
