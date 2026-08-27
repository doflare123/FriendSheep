package tests

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"

	"gorm.io/gorm"
)

func TestEventCommandUpdateWaitsForMembershipWriteAndSurvivesAuditFailurePostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
	}

	db := newPostgresEventCommandDB(t, dsn)
	service := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	seedEventParticipant(t, db, eventID, 1)
	createFailingEventCommandAuditTriggerPostgres(t, db)

	lockReady := make(chan struct{})
	releaseMembership := make(chan struct{})
	membershipDone := make(chan error, 1)
	released := false
	defer func() {
		if !released {
			close(releaseMembership)
		}
	}()

	go func() {
		membershipDone <- db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&eventmodels.EventsUser{
				EventID:  eventID,
				UserID:   2,
				JoinedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&eventmodels.Event{}).
				Where("id = ?", eventID).
				UpdateColumn("current_users", gorm.Expr("current_users + 1")).Error; err != nil {
				return err
			}
			close(lockReady)
			<-releaseMembership
			return nil
		})
	}()

	select {
	case <-lockReady:
	case err := <-membershipDone:
		t.Fatalf("membership write finished before lock check: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("membership write did not acquire the event row lock")
	}

	blockedTitle := "Обновление должно ждать"
	blockedContext, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	_, err := service.UpdateEvent(blockedContext, 1, eventID, servicesevents.UpdateEventInput{Title: &blockedTitle})
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("UpdateEvent while membership lock is held returned %v, want context deadline", err)
	}

	close(releaseMembership)
	released = true
	if err := <-membershipDone; err != nil {
		t.Fatalf("commit membership write: %v", err)
	}

	updatedTitle := "Обновление после освобождения блокировки"
	result, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{Title: &updatedTitle})
	if err != nil {
		t.Fatalf("UpdateEvent returned audit error: %v", err)
	}
	if result == nil || result.Title != updatedTitle || result.CurrentUsers != 2 {
		t.Fatalf("result = %#v, want updated title and two participants", result)
	}

	var stored eventmodels.Event
	if err := db.First(&stored, eventID).Error; err != nil {
		t.Fatalf("load updated event: %v", err)
	}
	if stored.Title != updatedTitle || stored.CurrentUsers != 2 {
		t.Fatalf("stored event title/current users = %q/%d, want %q/2", stored.Title, stored.CurrentUsers, updatedTitle)
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertGroupActionLogCount(t, db, groupID, groupmodels.ActionUpdateEvent, 0)
}

func newPostgresEventCommandDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	db := newPostgresGroupServiceDB(t, dsn)
	if err := db.AutoMigrate(
		&models.Category{},
		&eventmodels.EventLocation{},
		&eventmodels.Status{},
		&eventmodels.AgeLimit{},
		&eventmodels.Genre{},
		&eventmodels.Event{},
		&eventmodels.EventsUser{},
		&eventmodels.EventGenre{},
		&eventmodels.EventLifecycleScheduleOutbox{},
	); err != nil {
		t.Fatalf("auto migrate event models in postgres test schema: %v", err)
	}
	return db
}

func createFailingEventCommandAuditTriggerPostgres(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION fail_event_command_audit_insert()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			RAISE EXCEPTION 'injected event command audit failure';
		END;
		$$
	`).Error; err != nil {
		t.Fatalf("create postgres audit failure function: %v", err)
	}
	if err := db.Exec(`
		CREATE TRIGGER fail_event_command_audit_insert
		BEFORE INSERT ON group_action_logs
		FOR EACH ROW
		EXECUTE FUNCTION fail_event_command_audit_insert()
	`).Error; err != nil {
		t.Fatalf("create postgres audit failure trigger: %v", err)
	}
}
