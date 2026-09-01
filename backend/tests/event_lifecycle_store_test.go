package tests

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	servicesevents "friendship/services/events"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestGORMEventLifecycleStoreLoadsUpdatesAndRollsBack(t *testing.T) {
	db := newEventsServiceDB(t)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	recruitmentID := seedLifecycleStatus(t, db, eventmodels.StatusRecruitment)
	activeID := seedLifecycleStatus(t, db, eventmodels.StatusActive)
	completedID := seedLifecycleStatus(t, db, eventmodels.StatusCompleted)
	now := time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC)
	setLifecycleEvent(t, db, eventID, recruitmentID, now, now.Add(time.Hour))

	store := servicesevents.NewGORMEventLifecycleStore(&testPostgresRepository{db: db})
	ctx := context.Background()
	err := store.WithinLifecycleTransaction(ctx, eventID, func(tx servicesevents.EventLifecycleTransaction) error {
		snapshot, err := tx.LoadEventForUpdate(ctx, eventID)
		if err != nil {
			return err
		}
		if snapshot.ID != eventID || snapshot.StatusName != eventmodels.StatusRecruitment ||
			!snapshot.StartTime.Equal(now) || !snapshot.EndTime.Equal(now.Add(time.Hour)) {
			t.Fatalf("snapshot = %#v", snapshot)
		}
		resolvedID, err := tx.ResolveStatusID(ctx, eventmodels.StatusActive)
		if err != nil {
			return err
		}
		if resolvedID != activeID {
			t.Fatalf("active status ID = %d, want %d", resolvedID, activeID)
		}
		return tx.UpdateEventStatus(ctx, eventID, resolvedID)
	})
	if err != nil {
		t.Fatalf("commit lifecycle transition: %v", err)
	}
	assertLifecycleEventStatus(t, db, eventID, activeID)

	injected := errors.New("rollback transition")
	err = store.WithinLifecycleTransaction(ctx, eventID, func(tx servicesevents.EventLifecycleTransaction) error {
		if err := tx.UpdateEventStatus(ctx, eventID, completedID); err != nil {
			return err
		}
		return injected
	})
	if !errors.Is(err, injected) {
		t.Fatalf("rollback transaction error = %v, want injected error", err)
	}
	assertLifecycleEventStatus(t, db, eventID, activeID)
}

func TestGORMEventLifecycleStoreMapsMissingEventAndStatus(t *testing.T) {
	db := newEventsServiceDB(t)
	store := servicesevents.NewGORMEventLifecycleStore(&testPostgresRepository{db: db})
	ctx := context.Background()

	err := store.WithinLifecycleTransaction(ctx, 999, func(tx servicesevents.EventLifecycleTransaction) error {
		_, err := tx.LoadEventForUpdate(ctx, 999)
		return err
	})
	if !errors.Is(err, servicesevents.ErrEventNotFound) {
		t.Fatalf("missing event error = %v, want ErrEventNotFound", err)
	}

	err = store.WithinLifecycleTransaction(ctx, 999, func(tx servicesevents.EventLifecycleTransaction) error {
		_, err := tx.ResolveStatusID(ctx, "missing lifecycle status")
		return err
	})
	if !errors.Is(err, servicesevents.ErrInvalidEventLifecycleState) {
		t.Fatalf("missing status error = %v, want ErrInvalidEventLifecycleState", err)
	}
}

func TestEventLifecycleConcurrentAdvanceUsesPostgresRowLock(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to verify PostgreSQL lifecycle row locking")
	}

	db := newPostgresEventCommandDB(t, dsn)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	recruitmentID := seedLifecycleStatus(t, db, eventmodels.StatusRecruitment)
	activeID := seedLifecycleStatus(t, db, eventmodels.StatusActive)
	seedLifecycleStatus(t, db, eventmodels.StatusCompleted)
	now := time.Date(2036, 8, 15, 12, 0, 0, 0, time.UTC)
	setLifecycleEvent(t, db, eventID, recruitmentID, now, now.Add(time.Hour))
	service := servicesevents.NewEventLifecycleService(
		servicesevents.NewGORMEventLifecycleStore(&testPostgresRepository{db: db}),
		lifecycleClockStub{now: now},
	)

	lockReady := make(chan struct{})
	releaseLock := make(chan struct{})
	lockDone := make(chan error, 1)
	released := false
	defer func() {
		if !released {
			close(releaseLock)
		}
	}()
	go func() {
		lockDone <- db.Transaction(func(tx *gorm.DB) error {
			var event eventmodels.Event
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, eventID).Error; err != nil {
				return err
			}
			close(lockReady)
			<-releaseLock
			return nil
		})
	}()
	select {
	case <-lockReady:
	case err := <-lockDone:
		t.Fatalf("lock transaction ended early: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("PostgreSQL test could not acquire event row lock")
	}

	blockedCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	_, err := service.Advance(blockedCtx, eventID)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Advance while row locked returned %v, want context deadline", err)
	}

	close(releaseLock)
	released = true
	if err := <-lockDone; err != nil {
		t.Fatalf("release row lock: %v", err)
	}

	results := make(chan servicesevents.EventLifecycleResult, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := service.Advance(context.Background(), eventID)
			results <- result
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Advance error = %v", err)
		}
	}
	applied := 0
	for result := range results {
		if result.Applied {
			applied++
		}
	}
	if applied != 1 {
		t.Fatalf("applied concurrent outcomes = %d, want 1", applied)
	}
	assertLifecycleEventStatus(t, db, eventID, activeID)
}

func seedLifecycleStatus(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()
	var status eventmodels.Status
	err := db.Where("name = ?", name).First(&status).Error
	if err == nil {
		return status.ID
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("seed lifecycle status %q: %v", name, err)
	}

	if err := syncLifecycleStatusSequence(db); err != nil {
		t.Fatalf("sync lifecycle status sequence for %q: %v", name, err)
	}

	status = eventmodels.Status{Name: name}
	if err := db.Create(&status).Error; err != nil {
		t.Fatalf("seed lifecycle status %q: %v", name, err)
	}
	return status.ID
}

func syncLifecycleStatusSequence(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}

	return db.Exec(`
SELECT setval(
	pg_get_serial_sequence('statuses', 'id'),
	COALESCE((SELECT MAX(id) FROM statuses), 1),
	EXISTS (SELECT 1 FROM statuses)
)`).Error
}

func setLifecycleEvent(t *testing.T, db *gorm.DB, eventID, statusID uint, startTime, endTime time.Time) {
	t.Helper()
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"status_id":  statusID,
		"start_time": startTime,
		"end_time":   endTime,
	}).Error; err != nil {
		t.Fatalf("set lifecycle event fields: %v", err)
	}
}

func assertLifecycleEventStatus(t *testing.T, db *gorm.DB, eventID, wantStatusID uint) {
	t.Helper()
	var event eventmodels.Event
	if err := db.First(&event, eventID).Error; err != nil {
		t.Fatalf("load lifecycle event: %v", err)
	}
	if event.StatusID != wantStatusID {
		t.Fatalf("event status ID = %d, want %d", event.StatusID, wantStatusID)
	}
}
