package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"

	"gorm.io/gorm"
)

func newGORMEventAdminService(repo *testPostgresRepository) servicesevents.EventAdminService {
	return servicesevents.NewEventAdminService(
		&testLogger{},
		servicesevents.NewGORMEventAdminReader(repo),
		servicesevents.NewGORMEventUnitOfWork(repo),
	)
}

func TestGORMEventAdminServiceGetEventDetailsForAdminReturnsNotFound(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := newGORMEventAdminService(repo)

	eventDTO, err := service.GetEventDetailsForAdmin(context.Background(), 1, 999)

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", err)
	}
}

func TestGORMEventAdminReaderDoesNotLoadParticipantsWithoutCapability(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 2, groupID, groupmodels.RoleMember)
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	participantQueries := 0
	callbackName := "test:deny_admin_participant_preload"
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "events_users" || tx.Statement.Table == "users" {
			participantQueries++
		}
	}); err != nil {
		t.Fatalf("register query callback: %v", err)
	}
	defer func() {
		if err := db.Callback().Query().Remove(callbackName); err != nil {
			t.Errorf("remove query callback: %v", err)
		}
	}()

	reader := servicesevents.NewGORMEventAdminReader(repo)
	details, err := reader.LoadDetails(context.Background(), 2, eventID)

	if err != nil {
		t.Fatalf("LoadDetails returned error: %v", err)
	}
	if !details.Access.EventFound ||
		!details.Access.ActorIsGroupMember ||
		details.Access.ActorRole != groupmodels.RoleMember {
		t.Fatalf("access = %#v, want plain group member", details.Access)
	}
	if details.Found || len(details.Participants) != 0 {
		t.Fatalf("details = %#v, want no loaded event or participants", details)
	}
	if participantQueries != 0 {
		t.Fatalf("participant queries = %d, want 0 before capability approval", participantQueries)
	}
}

func TestGORMEventAdminServiceKickUserFromEventRollsBackWhenCounterUpdateFails(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := newGORMEventAdminService(repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	if err := db.Exec(`
		CREATE TRIGGER fail_admin_kick_counter_update
		BEFORE UPDATE OF current_users ON events
		BEGIN
			SELECT RAISE(ABORT, 'injected admin kick counter failure');
		END
	`).Error; err != nil {
		t.Fatalf("create counter failure trigger: %v", err)
	}

	kicked, err := service.KickUserFromEvent(context.Background(), 1, eventID, 2)

	if kicked {
		t.Fatal("KickUserFromEvent returned true")
	}
	if err == nil || !strings.Contains(err.Error(), "injected admin kick counter failure") {
		t.Fatalf("err = %v, want injected counter failure", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
	assertGroupActionLogCount(t, db, groupID, groupmodels.ActionKickFromEvent, 0)
}

func TestGORMEventAdminServiceKickUserFromEventIgnoresAuditInsertFailure(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := newGORMEventAdminService(repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	if err := db.Exec(`
		CREATE TRIGGER fail_admin_kick_audit_insert
		BEFORE INSERT ON group_action_logs
		BEGIN
			SELECT RAISE(ABORT, 'injected admin kick audit failure');
		END
	`).Error; err != nil {
		t.Fatalf("create audit failure trigger: %v", err)
	}

	kicked, err := service.KickUserFromEvent(context.Background(), 1, eventID, 2)

	if err != nil {
		t.Fatalf("KickUserFromEvent returned audit error: %v", err)
	}
	if !kicked {
		t.Fatal("KickUserFromEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
	assertGroupActionLogCount(t, db, groupID, groupmodels.ActionKickFromEvent, 0)
}
