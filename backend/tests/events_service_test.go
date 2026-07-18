package tests

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"friendship/models"
	"friendship/models/dto"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestEventMembershipServiceJoinEventAddsParticipantAndIncrementsCount(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 2)

	joined, err := service.JoinEvent(context.Background(), 2, eventID)

	if err != nil {
		t.Fatalf("JoinEvent returned error: %v", err)
	}
	if !joined {
		t.Fatal("JoinEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
}

func TestEventMembershipServiceJoinEventRejectsNonGroupMemberWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)

	joined, err := service.JoinEvent(context.Background(), 2, eventID)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrNotGroupMember) {
		t.Fatalf("err = %v, want ErrNotGroupMember", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventMembershipServiceJoinEventRejectsFullEventWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 1)

	joined, err := service.JoinEvent(context.Background(), 2, eventID)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrEventFull) {
		t.Fatalf("err = %v, want ErrEventFull", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventMembershipServiceJoinEventRejectsDuplicateParticipant(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 2, 3)
	seedEventParticipant(t, db, eventID, 2)

	joined, err := service.JoinEvent(context.Background(), 2, eventID)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrAlreadyJoined) {
		t.Fatalf("err = %v, want ErrAlreadyJoined", err)
	}
	assertEventParticipantCount(t, db, eventID, 2, 1)
	assertEventCurrentUsers(t, db, eventID, 2)
}

func TestEventMembershipServiceJoinEventReturnsEventNotFound(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	joined, err := service.JoinEvent(context.Background(), 2, 999)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", err)
	}
}

func TestEventMembershipServiceLeaveEventRemovesParticipantAndDecrementsCount(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 2, 3)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	left, err := service.LeaveEvent(context.Background(), 2, eventID)

	if err != nil {
		t.Fatalf("LeaveEvent returned error: %v", err)
	}
	if !left {
		t.Fatal("LeaveEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventMembershipServiceLeaveEventRejectsCreatorWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)
	seedEventParticipant(t, db, eventID, 1)

	left, err := service.LeaveEvent(context.Background(), 1, eventID)

	if left {
		t.Fatal("LeaveEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrCreatorCantLeave) {
		t.Fatalf("err = %v, want ErrCreatorCantLeave", err)
	}
	assertEventParticipantExists(t, db, eventID, 1, true)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventMembershipServiceLeaveEventRejectsMissingParticipantWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)
	seedEventParticipant(t, db, eventID, 1)

	left, err := service.LeaveEvent(context.Background(), 2, eventID)

	if left {
		t.Fatal("LeaveEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrNotJoined) {
		t.Fatalf("err = %v, want ErrNotJoined", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventMembershipServiceLeaveEventReturnsEventNotFound(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventMembershipService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	left, err := service.LeaveEvent(context.Background(), 2, 999)

	if left {
		t.Fatal("LeaveEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", err)
	}
}

func TestEventCommandServiceCreateEventCreatesEventRelationsAndActionLog(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	genreID := seedEventGenre(t, db, "Strategy")

	eventDTO, err := service.CreateEvent(context.Background(), 1, servicesevents.CreateEventInput{
		Title:       "Board Game Night",
		Description: "Long enough event description",
		GroupID:     groupID,
		EventTypeID: 1,
		LocationID:  1,
		ImageURL:    "https://example.com/event.png",
		StartTime:   time.Now().Add(24 * time.Hour),
		Duration:    120,
		MaxUsers:    5,
		Genres:      []uint{genreID},
		AgeLimitID:  1,
	})

	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if eventDTO == nil || eventDTO.ID == 0 {
		t.Fatalf("eventDTO = %#v, want created event", eventDTO)
	}
	assertEventCurrentUsers(t, db, eventDTO.ID, 1)
	assertEventParticipantExists(t, db, eventDTO.ID, 1, true)
	assertEventGenreCount(t, db, eventDTO.ID, 1)
	assertGroupActionLogCount(t, db, groupID, "create_event", 1)
}

func TestEventCommandServiceCreateEventRollsBackWhenGenreMissing(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")

	eventDTO, err := service.CreateEvent(context.Background(), 1, servicesevents.CreateEventInput{
		Title:       "Board Game Night",
		Description: "Long enough event description",
		GroupID:     groupID,
		EventTypeID: 1,
		LocationID:  1,
		ImageURL:    "https://example.com/event.png",
		StartTime:   time.Now().Add(24 * time.Hour),
		Duration:    120,
		MaxUsers:    5,
		Genres:      []uint{999},
		AgeLimitID:  1,
	})

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrInvalidGenres) {
		t.Fatalf("err = %v, want ErrInvalidGenres", err)
	}
	assertEventCount(t, db, 0)
	assertEventParticipantTotal(t, db, 0)
	assertEventGenreTotal(t, db, 0)
	assertGroupActionLogCount(t, db, groupID, "create_event", 0)
}

func TestEventCommandServiceCreateEventRollsBackWhenAgeLimitMissing(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventReferences(t, db)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	genreID := seedEventGenre(t, db, "Strategy")

	eventDTO, err := service.CreateEvent(context.Background(), 1, servicesevents.CreateEventInput{
		Title:       "Board Game Night",
		Description: "Long enough event description",
		GroupID:     groupID,
		EventTypeID: 1,
		LocationID:  1,
		ImageURL:    "https://example.com/event.png",
		StartTime:   time.Now().Add(24 * time.Hour),
		Duration:    120,
		MaxUsers:    5,
		Genres:      []uint{genreID},
		AgeLimitID:  999,
	})

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrAgeLimitNotFound) {
		t.Fatalf("err = %v, want ErrAgeLimitNotFound", err)
	}
	assertEventCount(t, db, 0)
	assertEventParticipantTotal(t, db, 0)
	assertEventGenreTotal(t, db, 0)
	assertGroupActionLogCount(t, db, groupID, "create_event", 0)
}

func TestEventCommandServiceDeleteEventRemovesEventRelationsAndWritesActionLog(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	genreID := seedEventGenre(t, db, "Strategy")
	seedEventGenreRelation(t, db, eventID, genreID)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	deleted, err := service.DeleteEvent(context.Background(), 1, eventID)

	if err != nil {
		t.Fatalf("DeleteEvent returned error: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteEvent returned false")
	}
	assertEventExists(t, db, eventID, false)
	assertEventParticipantTotal(t, db, 0)
	assertEventGenreTotal(t, db, 0)
	assertGroupActionLogCount(t, db, groupID, "delete_event", 1)
}

func TestEventsServiceKickUserFromEventRemovesTargetAndDecrementsCount(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	seedEventUser(t, db, 3)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 3, 5)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)
	seedEventParticipant(t, db, eventID, 3)

	kicked, err := service.KickUserFromEvent(1, eventID, 2)

	if err != nil {
		t.Fatalf("KickUserFromEvent returned error: %v", err)
	}
	if !kicked {
		t.Fatal("KickUserFromEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 1, true)
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventParticipantExists(t, db, eventID, 3, true)
	assertEventCurrentUsers(t, db, eventID, 2)
	assertGroupActionLogCount(t, db, groupID, "kick_from_event", 1)
}

func TestEventsServiceKickUserFromEventRejectsCreatorWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 2, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	kicked, err := service.KickUserFromEvent(2, eventID, 1)

	if kicked {
		t.Fatal("KickUserFromEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrCreatorCantLeave) {
		t.Fatalf("err = %v, want ErrCreatorCantLeave", err)
	}
	assertEventParticipantExists(t, db, eventID, 1, true)
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
	assertGroupActionLogCount(t, db, groupID, "kick_from_event", 0)
}

func TestEventCommandServiceUpdateEventRecomputesEndTimeReplacesGenresAndWritesActionLog(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	oldGenreID := seedEventGenre(t, db, "Old")
	newGenreID := seedEventGenre(t, db, "New")
	seedEventGenreRelation(t, db, eventID, oldGenreID)
	newLocationID := seedEventLocation(t, db, "Online")
	newStart := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	newDuration := uint16(180)
	newTitle := "Updated Event Title"

	eventDTO, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{
		Title:      &newTitle,
		LocationID: &newLocationID,
		StartTime:  &newStart,
		Duration:   &newDuration,
		Genres:     []uint{newGenreID},
	})

	if err != nil {
		t.Fatalf("UpdateEvent returned error: %v", err)
	}
	if eventDTO == nil || eventDTO.Title != newTitle {
		t.Fatalf("eventDTO = %#v, want updated title %q", eventDTO, newTitle)
	}
	assertEventTiming(t, db, eventID, newStart, newStart.Add(time.Duration(newDuration)*time.Minute), newDuration)
	assertEventLocation(t, db, eventID, newLocationID)
	assertEventGenreLinked(t, db, eventID, oldGenreID, false)
	assertEventGenreLinked(t, db, eventID, newGenreID, true)
	assertGroupActionLogCount(t, db, groupID, "update_event", 1)
}

func TestEventCommandServiceUpdateEventRejectsStartedEventWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	pastStart := time.Now().Add(-2 * time.Hour)
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Update("start_time", pastStart).Error; err != nil {
		t.Fatalf("make event started: %v", err)
	}
	newTitle := "Should Not Be Saved"

	eventDTO, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{Title: &newTitle})

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrEventAlreadyStarted) {
		t.Fatalf("err = %v, want ErrEventAlreadyStarted", err)
	}
	assertEventTitle(t, db, eventID, "Test Event")
	assertGroupActionLogCount(t, db, groupID, "update_event", 0)
}

func TestEventCommandServiceUpdateEventRollsBackInvalidGenreReplacement(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	oldGenreID := seedEventGenre(t, db, "Old")
	seedEventGenreRelation(t, db, eventID, oldGenreID)
	newTitle := "Should Roll Back"

	eventDTO, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{
		Title:  &newTitle,
		Genres: []uint{999},
	})

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrInvalidGenres) {
		t.Fatalf("err = %v, want ErrInvalidGenres", err)
	}
	assertEventTitle(t, db, eventID, "Test Event")
	assertEventGenreLinked(t, db, eventID, oldGenreID, true)
	assertEventGenreCount(t, db, eventID, 1)
	assertGroupActionLogCount(t, db, groupID, "update_event", 0)
}

func TestEventCommandServiceUpdateEventRejectsMissingAgeLimitWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventCommandService(&testLogger{}, servicesevents.NewGORMEventUnitOfWork(repo))

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	oldGenreID := seedEventGenre(t, db, "Old")
	seedEventGenreRelation(t, db, eventID, oldGenreID)
	missingAgeLimitID := uint(999)
	newTitle := "Should Roll Back"

	eventDTO, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{
		Title:    &newTitle,
		AgeLimit: &missingAgeLimitID,
		Genres:   []uint{oldGenreID},
	})

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrAgeLimitNotFound) {
		t.Fatalf("err = %v, want ErrAgeLimitNotFound", err)
	}
	assertEventTitle(t, db, eventID, "Test Event")
	assertEventGenreLinked(t, db, eventID, oldGenreID, true)
	assertEventGenreCount(t, db, eventID, 1)
	assertGroupActionLogCount(t, db, groupID, "update_event", 0)
}

func TestEventsServiceGetEventDetailsReturnsSubscribedStateForGroupMember(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	genreID := seedEventGenre(t, db, "Strategy")
	seedEventGenreRelation(t, db, eventID, genreID)
	seedEventParticipant(t, db, eventID, 2)

	eventDTO, err := service.GetEventDetails(2, eventID)

	if err != nil {
		t.Fatalf("GetEventDetails returned error: %v", err)
	}
	if eventDTO == nil || eventDTO.ID != eventID {
		t.Fatalf("eventDTO = %#v, want event id %d", eventDTO, eventID)
	}
	if !eventDTO.Subscribed {
		t.Fatal("Subscribed = false, want true")
	}
	if eventDTO.IsCreator {
		t.Fatal("IsCreator = true, want false")
	}
}

func TestEventsServiceGetEventDetailsAllowsNonGroupMemberInPublicGroup(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)

	eventDTO, err := service.GetEventDetails(2, eventID)

	if err != nil {
		t.Fatalf("GetEventDetails returned error: %v", err)
	}
	if eventDTO == nil || eventDTO.ID != eventID {
		t.Fatalf("eventDTO = %#v, want event id %d", eventDTO, eventID)
	}
	if eventDTO.Subscribed {
		t.Fatal("Subscribed = true, want false for non-member in public group")
	}
}

func TestEventsServiceGetEventDetailsRejectsNonGroupMemberInPrivateGroup(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, true)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)

	eventDTO, err := service.GetEventDetails(2, eventID)

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrNotGroupMember) {
		t.Fatalf("err = %v, want ErrNotGroupMember", err)
	}
}

func TestEventsServiceGetEventDetailsForAdminReturnsParticipants(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, "Админ")
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	genreID := seedEventGenre(t, db, "Strategy")
	seedEventGenreRelation(t, db, eventID, genreID)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	eventDTO, err := service.GetEventDetailsForAdmin(1, eventID)

	if err != nil {
		t.Fatalf("GetEventDetailsForAdmin returned error: %v", err)
	}
	if eventDTO == nil || eventDTO.ID != eventID {
		t.Fatalf("eventDTO = %#v, want event id %d", eventDTO, eventID)
	}
	if len(eventDTO.AllParticipants) != 2 {
		t.Fatalf("participants = %d, want 2", len(eventDTO.AllParticipants))
	}
}

func TestEventsServiceGetEventDetailsForAdminRejectsPlainMember(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 2, groupID, "Участник")
	eventID := seedEvent(t, db, groupID, 1, 1, 5)

	eventDTO, err := service.GetEventDetailsForAdmin(2, eventID)

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
}

func TestEventsServiceGetGroupEventsAllowsPlainMember(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 2, groupID, groupmodels.RoleMember)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	genreID := seedEventGenre(t, db, "Strategy")
	seedEventGenreRelation(t, db, eventID, genreID)
	seedEventParticipant(t, db, eventID, 2)

	eventsList, err := service.GetGroupEvents(2, groupID)

	if err != nil {
		t.Fatalf("GetGroupEvents returned error: %v", err)
	}
	if len(eventsList) != 1 {
		t.Fatalf("events count = %d, want 1", len(eventsList))
	}
	if eventsList[0].ID != eventID {
		t.Fatalf("event ID = %d, want %d", eventsList[0].ID, eventID)
	}
	if !eventsList[0].Subscribed {
		t.Fatal("event subscribed = false, want true")
	}
}

func TestEventsServiceGetGroupEventsAllowsNonMemberInPublicGroup(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	seedEventParticipant(t, db, eventID, 1)

	eventsList, err := service.GetGroupEvents(2, groupID)

	if err != nil {
		t.Fatalf("GetGroupEvents returned error: %v", err)
	}
	if len(eventsList) != 1 {
		t.Fatalf("events count = %d, want 1", len(eventsList))
	}
	if eventsList[0].ID != eventID {
		t.Fatalf("event ID = %d, want %d", eventsList[0].ID, eventID)
	}
	if eventsList[0].Subscribed {
		t.Fatal("event subscribed = true, want false for different participant")
	}
}

func TestEventsServiceGetGroupEventsRejectsNonMemberInPrivateGroup(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, true)
	seedEvent(t, db, groupID, 1, 1, 5)

	eventsList, err := service.GetGroupEvents(2, groupID)

	if eventsList != nil {
		t.Fatalf("eventsList = %#v, want nil", eventsList)
	}
	if !errors.Is(err, servicesevents.ErrNotGroupMember) {
		t.Fatalf("err = %v, want ErrNotGroupMember", err)
	}
}

func TestEventsServiceSearchEventsHidesPrivateGroupsFromAnonymousAndAllowsMember(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	publicGroupID := seedEventGroup(t, db, 1, false)
	privateGroupID := seedEventGroup(t, db, 1, true)
	seedEventGroupMembership(t, db, 2, privateGroupID)
	publicEventID := seedEvent(t, db, publicGroupID, 1, 1, 5)
	privateEventID := seedEvent(t, db, privateGroupID, 1, 1, 5)
	seedEventParticipant(t, db, publicEventID, 2)
	seedEventParticipant(t, db, privateEventID, 1)

	anonymousResult, err := service.SearchEvents(0, servicesevents.EventSearchInput{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("SearchEvents anonymous returned error: %v", err)
	}
	if anonymousResult.Total != 1 || len(anonymousResult.Items) != 1 || anonymousResult.Items[0].ID != publicEventID {
		t.Fatalf("anonymous result = %#v, want only public event %d", anonymousResult, publicEventID)
	}
	if anonymousResult.Items[0].Subscribed {
		t.Fatal("anonymous subscribed = true, want false")
	}

	memberResult, err := service.SearchEvents(2, servicesevents.EventSearchInput{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("SearchEvents member returned error: %v", err)
	}
	if memberResult.Total != 2 || len(memberResult.Items) != 2 {
		t.Fatalf("member result = %#v, want public and private events", memberResult)
	}
	assertSearchResultHasEvent(t, memberResult, publicEventID)
	assertSearchResultHasEvent(t, memberResult, privateEventID)
	if !findSearchResultEvent(t, memberResult, publicEventID).Subscribed {
		t.Fatal("public search event subscribed = false, want true")
	}
	if findSearchResultEvent(t, memberResult, privateEventID).Subscribed {
		t.Fatal("private search event subscribed = true, want false for different participant")
	}
}

func TestEventsServiceSearchEventsAppliesFiltersExclusionsAndPagination(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventReferences(t, db)
	seedEventCategory(t, db, 2, "Movie")
	strategyGenreID := seedEventGenre(t, db, "Strategy")
	horrorGenreID := seedEventGenre(t, db, "Horror")

	groupID := seedEventGroup(t, db, 1, false)
	setEventGroupCity(t, db, groupID, "Moscow")
	seedEventGroupCategory(t, db, groupID, 1)

	otherGroupID := seedEventGroup(t, db, 1, false)
	setEventGroupCity(t, db, otherGroupID, "Kazan")
	seedEventGroupCategory(t, db, otherGroupID, 2)

	start := time.Date(2027, 1, 10, 12, 0, 0, 0, time.UTC)
	firstEventID := seedEvent(t, db, groupID, 1, 1, 5)
	setEventSearchFields(t, db, firstEventID, "First Strategy", start, 1, 1)
	seedEventGenreRelation(t, db, firstEventID, strategyGenreID)

	secondEventID := seedEvent(t, db, groupID, 1, 1, 5)
	setEventSearchFields(t, db, secondEventID, "Second Strategy", start.Add(24*time.Hour), 1, 1)
	seedEventGenreRelation(t, db, secondEventID, strategyGenreID)

	excludedByGenreID := seedEvent(t, db, groupID, 1, 1, 5)
	setEventSearchFields(t, db, excludedByGenreID, "Horror Event", start.Add(48*time.Hour), 1, 1)
	seedEventGenreRelation(t, db, excludedByGenreID, horrorGenreID)

	excludedByCategoryID := seedEvent(t, db, otherGroupID, 1, 1, 5)
	setEventSearchFields(t, db, excludedByCategoryID, "Other City", start.Add(72*time.Hour), 1, 1)
	seedEventGenreRelation(t, db, excludedByCategoryID, strategyGenreID)

	dateFrom := start.Add(-time.Hour)
	dateTo := start.Add(25 * time.Hour)
	result, err := service.SearchEvents(0, servicesevents.EventSearchInput{
		CategoryIDs:     []uint{1},
		ExcludeGenreIDs: []uint{horrorGenreID},
		LocationTypes:   []string{"offline"},
		City:            "mos",
		DateFrom:        &dateFrom,
		DateTo:          &dateTo,
		Page:            1,
		Limit:           1,
	})

	if err != nil {
		t.Fatalf("SearchEvents returned error: %v", err)
	}
	if result.Total != 2 || result.TotalPages != 2 || !result.HasMore || len(result.Items) != 1 {
		t.Fatalf("result page = %#v, want total=2 totalPages=2 hasMore=true one item", result)
	}
	if result.Items[0].ID != firstEventID {
		t.Fatalf("first page event = %d, want %d", result.Items[0].ID, firstEventID)
	}
}

func TestEventsServiceSearchEventsSubscriptionNewsReturnsUnjoinedGroupEvents(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	privateGroupID := seedEventGroup(t, db, 1, true)
	seedEventGroupMembership(t, db, 2, privateGroupID)
	joinedEventID := seedEvent(t, db, privateGroupID, 1, 1, 5)
	freshEventID := seedEvent(t, db, privateGroupID, 1, 1, 5)
	seedEventParticipant(t, db, joinedEventID, 2)

	result, err := service.SearchEvents(2, servicesevents.EventSearchInput{
		OnlySubscriptionNews: true,
		Page:                 1,
		Limit:                20,
	})

	if err != nil {
		t.Fatalf("SearchEvents subscription news returned error: %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].ID != freshEventID {
		t.Fatalf("subscription news result = %#v, want only fresh event %d", result, freshEventID)
	}
	assertSearchResultMissingEvent(t, result, joinedEventID)
}

func TestEventsServiceGetAllReferencesIncludesGenres(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventReferences(t, db)
	seedEventGenre(t, db, "Strategy")

	references, err := service.GetAllReferences()

	if err != nil {
		t.Fatalf("GetAllReferences returned error: %v", err)
	}
	if references == nil {
		t.Fatal("references = nil")
	}
	if len(references.Genres) != 1 || references.Genres[0].Name != "Strategy" {
		t.Fatalf("genres = %#v, want Strategy reference", references.Genres)
	}
	if len(references.EventTypes) == 0 || len(references.Locations) == 0 || len(references.AgeLimits) == 0 || len(references.Statuses) == 0 || len(references.GroupActionTypes) == 0 {
		t.Fatalf("references missing required collections: %#v", references)
	}
}

func newEventsServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "-" + testUintString(uint(time.Now().UnixNano())) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&groupmodels.Group{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&groupmodels.GroupActionType{},
		&groupmodels.GroupActionLog{},
		&eventmodels.Event{},
		&eventmodels.EventsUser{},
		&eventmodels.EventLocation{},
		&eventmodels.Status{},
		&eventmodels.AgeLimit{},
		&eventmodels.Genre{},
		&eventmodels.EventGenre{},
	); err != nil {
		t.Fatalf("auto migrate event service models: %v", err)
	}

	seedGroupActionTypes(t, db)

	return db
}

func seedEventUser(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()

	user := models.User{
		ID:       userID,
		Name:     "Event User",
		Password: "Password123!",
		Us:       "event-user-" + testUintString(userID),
		Email:    "event-user-" + testUintString(userID) + "@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create event user %d: %v", userID, err)
	}
}

func seedEventGroup(t *testing.T, db *gorm.DB, creatorID uint, isPrivate bool) uint {
	t.Helper()

	group := groupmodels.Group{
		Name:             "Event Group",
		Description:      "Event Group Description",
		SmallDescription: "Event Group",
		Image:            "https://example.com/group.png",
		CreaterID:        creatorID,
		IsPrivate:        isPrivate,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create event group: %v", err)
	}
	return group.ID
}

func seedEventGroupMembership(t *testing.T, db *gorm.DB, userID, groupID uint) {
	t.Helper()

	role := groupmodels.Role_in_group{Name: "member-" + testUintString(userID)}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create event group role: %v", err)
	}

	membership := groupmodels.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: role.Id,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create event group membership: %v", err)
	}
}

func seedEventGroupMembershipWithRole(t *testing.T, db *gorm.DB, userID, groupID uint, roleName string) {
	t.Helper()

	role := groupmodels.Role_in_group{Name: roleName}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create event group role %q: %v", roleName, err)
	}

	membership := groupmodels.GroupUsers{
		UserID:        userID,
		GroupID:       groupID,
		RoleInGroupID: role.Id,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create event group membership: %v", err)
	}
}

func seedEventReferences(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Where("id = ?", 1).FirstOrCreate(&models.Category{}, models.Category{ID: 1, Name: "Game"}).Error; err != nil {
		t.Fatalf("create event type: %v", err)
	}
	if err := db.Where("id = ?", 1).FirstOrCreate(&eventmodels.EventLocation{}, eventmodels.EventLocation{ID: 1, Name: "Offline"}).Error; err != nil {
		t.Fatalf("create event location: %v", err)
	}
	if err := db.Where("id = ?", 1).FirstOrCreate(&eventmodels.Status{}, eventmodels.Status{ID: 1, Name: "Planned"}).Error; err != nil {
		t.Fatalf("create event status: %v", err)
	}
	if err := db.Where("id = ?", 1).FirstOrCreate(&eventmodels.AgeLimit{}, eventmodels.AgeLimit{ID: 1, Name: "18+"}).Error; err != nil {
		t.Fatalf("create event age limit: %v", err)
	}
}

func seedEventLocation(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()

	location := eventmodels.EventLocation{Name: name}
	if err := db.Create(&location).Error; err != nil {
		t.Fatalf("create event location: %v", err)
	}
	return location.ID
}

func seedEventCategory(t *testing.T, db *gorm.DB, categoryID uint, name string) {
	t.Helper()

	if err := db.Where("id = ?", categoryID).FirstOrCreate(&models.Category{}, models.Category{ID: categoryID, Name: name}).Error; err != nil {
		t.Fatalf("create event category: %v", err)
	}
}

func seedEventGroupCategory(t *testing.T, db *gorm.DB, groupID, categoryID uint) {
	t.Helper()

	if err := db.Exec(
		"INSERT INTO group_group_categories (group_id, group_category_id) VALUES (?, ?)",
		groupID,
		categoryID,
	).Error; err != nil {
		t.Fatalf("create group category relation: %v", err)
	}
}

func setEventGroupCity(t *testing.T, db *gorm.DB, groupID uint, city string) {
	t.Helper()

	if err := db.Model(&groupmodels.Group{}).
		Where("id = ?", groupID).
		Update("city", city).Error; err != nil {
		t.Fatalf("update group city: %v", err)
	}
}

func seedEventGenre(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()

	genre := eventmodels.Genre{Name: name}
	if err := db.Create(&genre).Error; err != nil {
		t.Fatalf("create event genre: %v", err)
	}
	return genre.ID
}

func seedEventGenreRelation(t *testing.T, db *gorm.DB, eventID, genreID uint) {
	t.Helper()

	eventGenre := eventmodels.EventGenre{
		EventID: eventID,
		GenreID: genreID,
	}
	if err := db.Create(&eventGenre).Error; err != nil {
		t.Fatalf("create event genre relation: %v", err)
	}
}

func seedEvent(t *testing.T, db *gorm.DB, groupID, creatorID uint, currentUsers, maxUsers uint16) uint {
	t.Helper()

	seedEventReferences(t, db)

	event := eventmodels.Event{
		Title:           "Test Event",
		Description:     "Test Event Description",
		GroupID:         groupID,
		EventTypeID:     1,
		EventLocationID: 1,
		CreatorID:       creatorID,
		StartTime:       time.Now().Add(24 * time.Hour),
		EndTime:         time.Now().Add(26 * time.Hour),
		Duration:        120,
		CurrentUsers:    currentUsers,
		MaxUsers:        maxUsers,
		StatusID:        1,
		AgeLimitID:      1,
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatalf("create event: %v", err)
	}
	return event.ID
}

func setEventSearchFields(t *testing.T, db *gorm.DB, eventID uint, title string, startTime time.Time, eventTypeID, locationID uint) {
	t.Helper()

	if err := db.Model(&eventmodels.Event{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"title":             title,
			"description":       title + " description",
			"start_time":        startTime,
			"end_time":          startTime.Add(2 * time.Hour),
			"event_type_id":     eventTypeID,
			"event_location_id": locationID,
		}).Error; err != nil {
		t.Fatalf("update event search fields: %v", err)
	}
}

func seedEventParticipant(t *testing.T, db *gorm.DB, eventID, userID uint) {
	t.Helper()

	participant := eventmodels.EventsUser{
		EventID:  eventID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	if err := db.Create(&participant).Error; err != nil {
		t.Fatalf("create event participant: %v", err)
	}
}

func assertEventParticipantExists(t *testing.T, db *gorm.DB, eventID, userID uint, want bool) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.EventsUser{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Count(&count).Error; err != nil {
		t.Fatalf("count event participant: %v", err)
	}
	if (count > 0) != want {
		t.Fatalf("participant exists = %v, want %v", count > 0, want)
	}
}

func assertEventParticipantCount(t *testing.T, db *gorm.DB, eventID, userID uint, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.EventsUser{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Count(&count).Error; err != nil {
		t.Fatalf("count event participant: %v", err)
	}
	if count != want {
		t.Fatalf("participant count = %d, want %d", count, want)
	}
}

func assertEventCurrentUsers(t *testing.T, db *gorm.DB, eventID uint, want uint16) {
	t.Helper()

	var event eventmodels.Event
	if err := db.First(&event, eventID).Error; err != nil {
		t.Fatalf("find event: %v", err)
	}
	if event.CurrentUsers != want {
		t.Fatalf("current users = %d, want %d", event.CurrentUsers, want)
	}
}

func assertEventTitle(t *testing.T, db *gorm.DB, eventID uint, want string) {
	t.Helper()

	var event eventmodels.Event
	if err := db.First(&event, eventID).Error; err != nil {
		t.Fatalf("find event: %v", err)
	}
	if event.Title != want {
		t.Fatalf("event title = %q, want %q", event.Title, want)
	}
}

func assertEventTiming(t *testing.T, db *gorm.DB, eventID uint, wantStart, wantEnd time.Time, wantDuration uint16) {
	t.Helper()

	var event eventmodels.Event
	if err := db.First(&event, eventID).Error; err != nil {
		t.Fatalf("find event: %v", err)
	}
	if !event.StartTime.Equal(wantStart) {
		t.Fatalf("start time = %s, want %s", event.StartTime, wantStart)
	}
	if !event.EndTime.Equal(wantEnd) {
		t.Fatalf("end time = %s, want %s", event.EndTime, wantEnd)
	}
	if event.Duration != wantDuration {
		t.Fatalf("duration = %d, want %d", event.Duration, wantDuration)
	}
}

func assertEventLocation(t *testing.T, db *gorm.DB, eventID, wantLocationID uint) {
	t.Helper()

	var event eventmodels.Event
	if err := db.First(&event, eventID).Error; err != nil {
		t.Fatalf("find event: %v", err)
	}
	if event.EventLocationID != wantLocationID {
		t.Fatalf("event location ID = %d, want %d", event.EventLocationID, wantLocationID)
	}
}

func assertEventGenreCount(t *testing.T, db *gorm.DB, eventID uint, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.EventGenre{}).
		Where("event_id = ?", eventID).
		Count(&count).Error; err != nil {
		t.Fatalf("count event genres: %v", err)
	}
	if count != want {
		t.Fatalf("event genre count = %d, want %d", count, want)
	}
}

func assertEventGenreLinked(t *testing.T, db *gorm.DB, eventID, genreID uint, want bool) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.EventGenre{}).
		Where("event_id = ? AND genre_id = ?", eventID, genreID).
		Count(&count).Error; err != nil {
		t.Fatalf("count event genre relation: %v", err)
	}
	if (count > 0) != want {
		t.Fatalf("event genre linked = %v, want %v", count > 0, want)
	}
}

func assertEventCount(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.Event{}).Count(&count).Error; err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != want {
		t.Fatalf("event count = %d, want %d", count, want)
	}
}

func assertEventExists(t *testing.T, db *gorm.DB, eventID uint, want bool) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.Event{}).
		Where("id = ?", eventID).
		Count(&count).Error; err != nil {
		t.Fatalf("count event: %v", err)
	}
	if (count > 0) != want {
		t.Fatalf("event exists = %v, want %v", count > 0, want)
	}
}

func assertEventParticipantTotal(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.EventsUser{}).Count(&count).Error; err != nil {
		t.Fatalf("count event participants: %v", err)
	}
	if count != want {
		t.Fatalf("event participant count = %d, want %d", count, want)
	}
}

func assertEventGenreTotal(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&eventmodels.EventGenre{}).Count(&count).Error; err != nil {
		t.Fatalf("count event genres: %v", err)
	}
	if count != want {
		t.Fatalf("event genre count = %d, want %d", count, want)
	}
}

func assertGroupActionLogCount(t *testing.T, db *gorm.DB, groupID uint, action string, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupActionLog{}).
		Joins("JOIN group_action_types ON group_action_types.id = group_action_logs.action_type_id").
		Where("group_action_logs.group_id = ? AND group_action_types.code = ?", groupID, action).
		Count(&count).Error; err != nil {
		t.Fatalf("count group action logs: %v", err)
	}
	if count != want {
		t.Fatalf("group action log count = %d, want %d", count, want)
	}
}

func assertSearchResultHasEvent(t *testing.T, result *dto.EventSearchResponse, eventID uint) {
	t.Helper()

	for _, item := range result.Items {
		if item.ID == eventID {
			return
		}
	}
	t.Fatalf("search result missing event %d: %#v", eventID, result)
}

func assertSearchResultMissingEvent(t *testing.T, result *dto.EventSearchResponse, eventID uint) {
	t.Helper()

	for _, item := range result.Items {
		if item.ID == eventID {
			t.Fatalf("search result contains event %d: %#v", eventID, result)
		}
	}
}

func findSearchResultEvent(t *testing.T, result *dto.EventSearchResponse, eventID uint) dto.EventSearchItemDto {
	t.Helper()

	for _, item := range result.Items {
		if item.ID == eventID {
			return item
		}
	}
	t.Fatalf("search result missing event %d: %#v", eventID, result)
	return dto.EventSearchItemDto{}
}

func testUintString(value uint) string {
	return fmt.Sprintf("%d", value)
}
