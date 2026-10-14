package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
)

func TestEventEndedHistoryCommandsPreserveCompositionDespiteDelayedStatus(t *testing.T) {
	for _, command := range []string{"update", "delete", "leave", "kick"} {
		for _, offset := range []time.Duration{0, -time.Nanosecond, time.Nanosecond} {
			t.Run(command+"/"+offset.String(), func(t *testing.T) {
				db := newEventsServiceDB(t)
				for _, id := range []uint{1, 2} {
					seedEventUser(t, db, id)
				}
				groupID := seedEventGroup(t, db, 1, false)
				seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
				eventID := seedEvent(t, db, groupID, 1, 2, 10)
				seedEventParticipant(t, db, eventID, 1)
				seedEventParticipant(t, db, eventID, 2)
				now := time.Date(2036, 10, 14, 12, 0, 0, 0, time.UTC)
				if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Updates(map[string]any{
					"start_time": now.Add(-time.Hour), "end_time": now.Add(offset),
				}).Error; err != nil {
					t.Fatal(err)
				}
				repo := &testPostgresRepository{db: db}
				uow := servicesevents.NewGORMEventUnitOfWork(repo)
				clock := lifecycleClockStub{now: now}
				commands := servicesevents.NewEventCommandServiceWithClock(&testLogger{}, uow, clock)
				var err error
				switch command {
				case "update":
					title := "Нельзя переписать историю"
					_, err = commands.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{Title: &title})
				case "delete":
					_, err = commands.DeleteEvent(context.Background(), 1, eventID)
				case "leave":
					_, err = servicesevents.NewEventMembershipServiceWithClock(&testLogger{}, uow, clock).LeaveEvent(context.Background(), 2, eventID)
				case "kick":
					_, err = servicesevents.NewEventAdminServiceWithClock(&testLogger{}, nil, uow, clock).KickUserFromEvent(context.Background(), 1, eventID, 2)
				}
				if offset > 0 {
					if command == "update" && !errors.Is(err, servicesevents.ErrEventAlreadyStarted) {
						t.Fatalf("active update error = %v, want ErrEventAlreadyStarted", err)
					}
					if command != "update" && err != nil {
						t.Fatalf("active %s error = %v", command, err)
					}
					return
				}
				if !errors.Is(err, servicesevents.ErrEventAlreadyEnded) {
					t.Fatalf("%s error = %v, want ErrEventAlreadyEnded", command, err)
				}
				var stored eventmodels.Event
				if err := db.First(&stored, eventID).Error; err != nil {
					t.Fatal(err)
				}
				if stored.Title != "Test Event" || stored.CurrentUsers != 2 || stored.StatusID != 1 {
					t.Fatalf("исторические факты изменились: %#v", stored)
				}
				assertEventParticipantExists(t, db, eventID, 1, true)
				assertEventParticipantExists(t, db, eventID, 2, true)
				var auditCount int64
				if err := db.Model(&groupmodels.GroupActionLog{}).Where("group_id = ?", groupID).Count(&auditCount).Error; err != nil || auditCount != 0 {
					t.Fatalf("audit count/error = %d/%v", auditCount, err)
				}
			})
		}
	}
}

func TestEventEndedHistoryPreservesAuthorizationAndMembershipErrors(t *testing.T) {
	db := newEventsServiceDB(t)
	for _, id := range []uint{1, 2, 3} {
		seedEventUser(t, db, id)
	}
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	seedEventGroupMembershipWithRole(t, db, 2, groupID, groupmodels.RoleMember)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	seedEventParticipant(t, db, eventID, 1)
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Updates(map[string]any{
		"start_time": time.Now().Add(-time.Hour), "end_time": time.Now().Add(-time.Second),
	}).Error; err != nil {
		t.Fatal(err)
	}
	uow := servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db})
	commands := servicesevents.NewEventCommandService(&testLogger{}, uow)
	title := "Закрытая история"
	if _, err := commands.UpdateEvent(context.Background(), 2, eventID, servicesevents.UpdateEventInput{Title: &title}); !errors.Is(err, servicesevents.ErrPermissionDenied) {
		t.Fatalf("unauthorized update = %v", err)
	}
	if _, err := commands.DeleteEvent(context.Background(), 2, eventID); !errors.Is(err, servicesevents.ErrPermissionDenied) {
		t.Fatalf("unauthorized delete = %v", err)
	}
	memberships := servicesevents.NewEventMembershipService(&testLogger{}, uow)
	if _, err := memberships.LeaveEvent(context.Background(), 3, eventID); !errors.Is(err, servicesevents.ErrNotJoined) {
		t.Fatalf("nonparticipant leave = %v", err)
	}
	if _, err := memberships.LeaveEvent(context.Background(), 1, eventID); !errors.Is(err, servicesevents.ErrCreatorCantLeave) {
		t.Fatalf("creator leave = %v", err)
	}
	admins := servicesevents.NewEventAdminService(&testLogger{}, nil, uow)
	if _, err := admins.KickUserFromEvent(context.Background(), 1, eventID, 3); !errors.Is(err, servicesevents.ErrNotJoined) {
		t.Fatalf("nonparticipant kick = %v", err)
	}
}
