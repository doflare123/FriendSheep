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

func TestGORMEventUnitOfWorkJoinsAndLeavesAtomically(t *testing.T) {
	db := newEventsServiceDB(t)
	uow := servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db})

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)
	seedEventParticipant(t, db, eventID, 1)

	err := uow.WithinTransaction(context.Background(), func(tx servicesevents.EventTransaction) error {
		store := tx.Membership()
		event, err := store.FindEvent(eventID)
		if err != nil {
			return err
		}
		if event.ID != eventID || event.GroupID != groupID || event.CreatorID != 1 {
			t.Fatalf("event snapshot = %#v", event)
		}

		member, err := store.IsGroupMember(2, groupID)
		if err != nil {
			return err
		}
		if !member {
			t.Fatal("IsGroupMember returned false")
		}

		participant, err := store.IsParticipant(2, eventID)
		if err != nil {
			return err
		}
		if participant {
			t.Fatal("IsParticipant returned true before join")
		}

		if err := store.AddParticipant(2, eventID, time.Now()); err != nil {
			return err
		}
		incremented, err := store.IncrementParticipantsIfSpace(eventID)
		if err != nil {
			return err
		}
		if !incremented {
			return servicesevents.ErrEventFull
		}
		return nil
	})
	if err != nil {
		t.Fatalf("join transaction returned error: %v", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)

	err = uow.WithinTransaction(context.Background(), func(tx servicesevents.EventTransaction) error {
		store := tx.Membership()
		if err := store.RemoveParticipant(2, eventID); err != nil {
			return err
		}
		return store.DecrementParticipants(eventID)
	})
	if err != nil {
		t.Fatalf("leave transaction returned error: %v", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestGORMEventMembershipStoreMapsDuplicateParticipant(t *testing.T) {
	db := newEventsServiceDB(t)
	uow := servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db})

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 2, 3)
	seedEventParticipant(t, db, eventID, 2)

	err := uow.WithinTransaction(context.Background(), func(tx servicesevents.EventTransaction) error {
		return tx.Membership().AddParticipant(2, eventID, time.Now())
	})
	if !errors.Is(err, servicesevents.ErrAlreadyJoined) {
		t.Fatalf("AddParticipant err = %v, want ErrAlreadyJoined", err)
	}
	assertEventParticipantCount(t, db, eventID, 2, 1)
	assertEventCurrentUsers(t, db, eventID, 2)
}

func TestGORMEventUnitOfWorkRollsBackParticipantWhenEventIsFull(t *testing.T) {
	db := newEventsServiceDB(t)
	uow := servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db})

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 1)

	err := uow.WithinTransaction(context.Background(), func(tx servicesevents.EventTransaction) error {
		store := tx.Membership()
		if err := store.AddParticipant(2, eventID, time.Now()); err != nil {
			return err
		}
		incremented, err := store.IncrementParticipantsIfSpace(eventID)
		if err != nil {
			return err
		}
		if !incremented {
			return servicesevents.ErrEventFull
		}
		return nil
	})
	if !errors.Is(err, servicesevents.ErrEventFull) {
		t.Fatalf("transaction err = %v, want ErrEventFull", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestGORMEventUnitOfWorkRollsBackLeaveWhenCounterUpdateFails(t *testing.T) {
	db := newEventsServiceDB(t)
	uow := servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db})

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 2, 3)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	if err := db.Exec(`
		CREATE TRIGGER fail_event_counter_update
		BEFORE UPDATE OF current_users ON events
		BEGIN
			SELECT RAISE(ABORT, 'injected event counter failure');
		END
	`).Error; err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	err := uow.WithinTransaction(context.Background(), func(tx servicesevents.EventTransaction) error {
		store := tx.Membership()
		if err := store.RemoveParticipant(2, eventID); err != nil {
			return err
		}
		return store.DecrementParticipants(eventID)
	})
	if err == nil || !strings.Contains(err.Error(), "injected event counter failure") {
		t.Fatalf("transaction err = %v, want injected counter failure", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
}

func TestGORMEventUnitOfWorkRollsBackLateContextCancellation(t *testing.T) {
	db := newEventsServiceDB(t)
	uow := servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)

	const callbackName = "test:cancel_event_membership_after_counter_update"
	if err := db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
		if _, ok := tx.Statement.Model.(*eventmodels.Event); ok && tx.Error == nil {
			cancel()
		}
	}); err != nil {
		t.Fatalf("register cancellation callback: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Update().Remove(callbackName)
	})

	err := uow.WithinTransaction(ctx, func(tx servicesevents.EventTransaction) error {
		store := tx.Membership()
		if err := store.AddParticipant(2, eventID, time.Now()); err != nil {
			return err
		}
		incremented, err := store.IncrementParticipantsIfSpace(eventID)
		if err != nil {
			return err
		}
		if !incremented {
			return servicesevents.ErrEventFull
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("transaction err = %v, want context.Canceled", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestGORMEventAuditSavepointKeepsJoinAndLeaveCommittedAfterInsertFailure(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(repo),
	)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)
	seedEventParticipant(t, db, eventID, 1)

	if err := db.Exec(`
		CREATE TRIGGER fail_event_audit_insert
		BEFORE INSERT ON group_action_logs
		BEGIN
			SELECT RAISE(ABORT, 'injected event audit failure');
		END
	`).Error; err != nil {
		t.Fatalf("create audit failure trigger: %v", err)
	}

	joined, err := service.JoinEvent(context.Background(), 2, eventID)
	if err != nil {
		t.Fatalf("JoinEvent returned audit error: %v", err)
	}
	if !joined {
		t.Fatal("JoinEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
	assertGroupActionLogCount(t, db, groupID, groupmodels.ActionJoinEvent, 0)

	left, err := service.LeaveEvent(context.Background(), 2, eventID)
	if err != nil {
		t.Fatalf("LeaveEvent returned audit error: %v", err)
	}
	if !left {
		t.Fatal("LeaveEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
	assertGroupActionLogCount(t, db, groupID, groupmodels.ActionLeaveEvent, 0)
}
