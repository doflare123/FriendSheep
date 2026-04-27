package tests

import (
	"errors"
	"testing"

	"friendship/models"
	groupmodels "friendship/models/groups"
	servicegroups "friendship/services/groups"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestGroupServiceJoinGroupAddsMemberForPublicGroup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	result, err := service.JoinGroup(2, groupID)

	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
}

func TestGroupServiceJoinGroupCreatesRequestForPrivateGroup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)

	result, err := service.JoinGroup(2, groupID)

	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || result.Joined {
		t.Fatalf("result = %#v, want pending request result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 1)
}

func TestGroupServiceJoinGroupRejectsBlacklistedUserWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupBlacklist(t, db, groupID, 2, 1)

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrUserInBlacklist) {
		t.Fatalf("err = %v, want ErrUserInBlacklist", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
}

func TestGroupServiceJoinGroupRejectsDuplicatePendingRequest(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupJoinRequest(t, db, groupID, 2, "pending")

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrRequestAlreadyExists) {
		t.Fatalf("err = %v, want ErrRequestAlreadyExists", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 1)
}

func newGroupServiceDB(t *testing.T) *gorm.DB {
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
		&groupmodels.GroupJoinRequest{},
		&groupmodels.GroupBlacklist{},
	); err != nil {
		t.Fatalf("auto migrate group service models: %v", err)
	}

	return db
}

func seedGroupServiceUser(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()

	user := models.User{
		ID:       userID,
		Name:     "Group User",
		Password: "Password123!",
		Us:       "group-user-" + testUintString(userID),
		Email:    "group-user-" + testUintString(userID) + "@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create group user %d: %v", userID, err)
	}
}

func seedGroupServiceRole(t *testing.T, db *gorm.DB, roleName string) uint {
	t.Helper()

	role := groupmodels.Role_in_group{Name: roleName}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create group role %q: %v", roleName, err)
	}
	return role.Id
}

func seedGroupServiceGroup(t *testing.T, db *gorm.DB, creatorID uint, isPrivate bool) uint {
	t.Helper()

	group := groupmodels.Group{
		Name:             "Group Service Test",
		Description:      "Group Service Test Description",
		SmallDescription: "Group Service",
		Image:            "https://example.com/group.png",
		CreaterID:        creatorID,
		IsPrivate:        isPrivate,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group service group: %v", err)
	}
	return group.ID
}

func seedGroupBlacklist(t *testing.T, db *gorm.DB, groupID, userID, bannedBy uint) {
	t.Helper()

	entry := groupmodels.GroupBlacklist{
		GroupID:  groupID,
		UserID:   userID,
		BannedBy: bannedBy,
		Reason:   "test",
	}
	if err := db.Create(&entry).Error; err != nil {
		t.Fatalf("create group blacklist entry: %v", err)
	}
}

func seedGroupJoinRequest(t *testing.T, db *gorm.DB, groupID, userID uint, status string) {
	t.Helper()

	request := groupmodels.GroupJoinRequest{
		GroupID: groupID,
		UserID:  userID,
		Status:  status,
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatalf("create group join request: %v", err)
	}
}

func assertGroupMembershipExists(t *testing.T, db *gorm.DB, groupID, userID uint, want bool) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupUsers{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		t.Fatalf("count group membership: %v", err)
	}
	if (count > 0) != want {
		t.Fatalf("membership exists = %v, want %v", count > 0, want)
	}
}

func assertGroupJoinRequestCount(t *testing.T, db *gorm.DB, groupID, userID uint, status string, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupJoinRequest{}).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, status).
		Count(&count).Error; err != nil {
		t.Fatalf("count group join requests: %v", err)
	}
	if count != want {
		t.Fatalf("join request count = %d, want %d", count, want)
	}
}
