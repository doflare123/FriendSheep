package tests

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestEventsServiceJoinEventAddsParticipantAndIncrementsCount(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 2)

	joined, err := service.JoinEvent(2, eventID)

	if err != nil {
		t.Fatalf("JoinEvent returned error: %v", err)
	}
	if !joined {
		t.Fatal("JoinEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
}

func TestEventsServiceJoinEventRejectsNonGroupMemberWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)

	joined, err := service.JoinEvent(2, eventID)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrNotGroupMember) {
		t.Fatalf("err = %v, want ErrNotGroupMember", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventsServiceJoinEventRejectsFullEventWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 1)

	joined, err := service.JoinEvent(2, eventID)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrEventFull) {
		t.Fatalf("err = %v, want ErrEventFull", err)
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventsServiceLeaveEventRemovesParticipantAndDecrementsCount(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 2, 3)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)

	left, err := service.LeaveEvent(2, eventID)

	if err != nil {
		t.Fatalf("LeaveEvent returned error: %v", err)
	}
	if !left {
		t.Fatal("LeaveEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, false)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func TestEventsServiceLeaveEventRejectsCreatorWithoutSideEffects(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventsService(&testLogger{}, repo)

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 1, 3)
	seedEventParticipant(t, db, eventID, 1)

	left, err := service.LeaveEvent(1, eventID)

	if left {
		t.Fatal("LeaveEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrCreatorCantLeave) {
		t.Fatalf("err = %v, want ErrCreatorCantLeave", err)
	}
	assertEventParticipantExists(t, db, eventID, 1, true)
	assertEventCurrentUsers(t, db, eventID, 1)
}

func newEventsServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&groupmodels.Group{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&eventmodels.Event{},
		&eventmodels.EventsUser{},
	); err != nil {
		t.Fatalf("auto migrate event service models: %v", err)
	}

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

func seedEvent(t *testing.T, db *gorm.DB, groupID, creatorID uint, currentUsers, maxUsers uint16) uint {
	t.Helper()

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

func testUintString(value uint) string {
	return fmt.Sprintf("%d", value)
}
