package tests

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	servicesevents "friendship/services/events"
)

func TestEventOutboxPublicationCommitOrderPostgres(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		name := "commit"
		if rollback {
			name = "rollback_gap"
		}
		t.Run(name, func(t *testing.T) {
			dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
			if dsn == "" {
				t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
			}
			db := newPostgresEventCommandDB(t, dsn)
			repo := &testPostgresRepository{db: db}
			uow := servicesevents.NewGORMEventUnitOfWork(repo)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ready, release := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			defer releaseOnce.Do(func() { close(release) })
			doneA, doneB := make(chan error, 1), make(chan error, 1)
			rollbackErr := errors.New("controlled rollback")
			appendBoth := func(tx servicesevents.EventTransaction, id uint, message string) error {
				now := time.Now().UTC()
				if err := tx.LifecycleScheduleOutbox().AppendLifecycleScheduleEvent(servicesevents.EventLifecycleScheduleOutboxInput{MessageID: message, EventID: id, Operation: servicesevents.LifecycleScheduleOperationCancel, OccurredAt: now, SchemaVersion: 1}); err != nil {
					return err
				}
				return tx.EventReminderIntentOutbox().AppendEventReminderIntent(servicesevents.EventReminderIntentOutboxInput{MessageID: message, EventID: id, IntentType: servicesevents.EventReminderIntentType, Operation: servicesevents.EventReminderOperationCancel, OccurredAt: now, SchemaVersion: 1})
			}
			go func() {
				doneA <- uow.WithinTransaction(ctx, func(tx servicesevents.EventTransaction) error {
					if err := appendBoth(tx, 1, "00000000-0000-4000-8000-000000000001"); err != nil {
						return err
					}
					close(ready)
					select {
					case <-release:
					case <-ctx.Done():
						return ctx.Err()
					}
					if rollback {
						return rollbackErr
					}
					return nil
				})
			}()
			select {
			case <-ready:
			case err := <-doneA:
				t.Fatalf("A before barrier: %v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			go func() {
				doneB <- uow.WithinTransaction(ctx, func(tx servicesevents.EventTransaction) error {
					return appendBoth(tx, 2, "00000000-0000-4000-8000-000000000002")
				})
			}()
			// Барьер подтверждает реальное ожидание PostgreSQL, а не предполагает его по задержке.
			for {
				var waiting bool
				err := db.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM pg_locks WHERE locktype = 'advisory' AND classid = 1179862864 AND objid = 1 AND NOT granted)").Scan(&waiting).Error
				if err != nil {
					t.Fatal(err)
				}
				if waiting {
					break
				}
				select {
				case err := <-doneB:
					t.Fatalf("B passed uncommitted A: %v", err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				default:
					runtime.Gosched()
				}
			}
			lifecycle := servicesevents.NewEventLifecycleScheduleOutboxService(servicesevents.NewGORMEventLifecycleScheduleOutboxReader(repo))
			cancelCtx, cancelWait := context.WithCancel(ctx)
			doneCancelled := make(chan error, 1)
			go func() {
				doneCancelled <- uow.WithinTransaction(cancelCtx, func(servicesevents.EventTransaction) error {
					return errors.New("cancelled waiter entered callback")
				})
			}()
			for {
				var count int64
				err := db.WithContext(ctx).Raw("SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND classid = 1179862864 AND objid = 1 AND NOT granted").Scan(&count).Error
				if err != nil {
					cancelWait()
					t.Fatal(err)
				}
				if count >= 2 {
					break
				}
				select {
				case err := <-doneCancelled:
					cancelWait()
					t.Fatalf("waiter returned before cancellation: %v", err)
				case <-ctx.Done():
					cancelWait()
					t.Fatal(ctx.Err())
				default:
					runtime.Gosched()
				}
			}
			cancelWait()
			if err := <-doneCancelled; !errors.Is(err, context.Canceled) {
				t.Fatalf("lock wait cancellation: %v", err)
			}
			reminders := servicesevents.NewEventReminderIntentService(servicesevents.NewGORMEventReminderIntentOutboxReader(repo))
			firstL, err := lifecycle.List(ctx, 0, 1)
			if err != nil {
				t.Fatal(err)
			}
			firstR, err := reminders.List(ctx, 0, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(firstL.Items) != 0 || len(firstR.Items) != 0 || firstL.NextCursor != 0 || firstR.NextCursor != 0 {
				t.Fatalf("uncommitted publications visible: %#v / %#v", firstL, firstR)
			}
			releaseOnce.Do(func() { close(release) })
			if err := <-doneA; (!rollback && err != nil) || (rollback && !errors.Is(err, rollbackErr)) {
				t.Fatalf("A result: %v", err)
			}
			if err := <-doneB; err != nil {
				t.Fatalf("B result: %v", err)
			}
			want := []uint{1, 2}
			if rollback {
				want = []uint{2}
			}
			// Независимые читатели и повтор страницы должны видеть одинаковую последовательность.
			for consumer := 0; consumer < 2; consumer++ {
				var cursorL, cursorR int64
				for _, id := range want {
					pageL, err := lifecycle.List(ctx, cursorL, 1)
					if err != nil {
						t.Fatal(err)
					}
					pageR, err := reminders.List(ctx, cursorR, 1)
					if err != nil {
						t.Fatal(err)
					}
					if len(pageL.Items) != 1 || len(pageR.Items) != 1 || pageL.Items[0].EventID != id || pageR.Items[0].EventID != id {
						t.Fatalf("lost publication for %d: %#v / %#v", id, pageL, pageR)
					}
					replayL, err := lifecycle.List(ctx, cursorL, 1)
					if err != nil {
						t.Fatal(err)
					}
					replayR, err := reminders.List(ctx, cursorR, 1)
					if err != nil {
						t.Fatal(err)
					}
					if replayL.NextCursor != pageL.NextCursor || replayR.NextCursor != pageR.NextCursor {
						t.Fatal("replay changed cursor")
					}
					cursorL, cursorR = pageL.NextCursor, pageR.NextCursor
				}
				if rollback && (cursorL <= 1 || cursorR <= 1) {
					t.Fatalf("rollback did not leave sequence gap: %d/%d", cursorL, cursorR)
				}
				endL, err := lifecycle.List(ctx, cursorL, 1)
				if err != nil {
					t.Fatal(err)
				}
				endR, err := reminders.List(ctx, cursorR, 1)
				if err != nil {
					t.Fatal(err)
				}
				if len(endL.Items) != 0 || len(endR.Items) != 0 || endL.NextCursor != cursorL || endR.NextCursor != cursorR {
					t.Fatal("unexpected trailing publication")
				}
			}
		})
	}
}
