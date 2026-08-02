package tests

import (
	"context"
	"friendship/models/groups"
	servicegroups "friendship/services/groups"
	"testing"

	"gorm.io/gorm"
)

func TestGroupServiceGetManagedGroupsSeparatesRolesAndUsesOnlyCurrentUser(t *testing.T) {
	db := newGroupServiceDB(t)
	service := servicegroups.NewGroupService(
		&testLogger{},
		servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: db}),
	)

	adminRoleID := seedGroupServiceRole(t, db, groups.RoleAdmin)
	moderatorRoleID := seedGroupServiceRole(t, db, groups.RoleModerator)
	memberRoleID := seedGroupServiceRole(t, db, groups.RoleMember)
	for _, userID := range []uint{1, 2, 3, 4} {
		seedGroupServiceUser(t, db, userID)
	}

	adminGroupID := seedGroupServiceGroup(t, db, 2, false)
	moderatorGroupID := seedGroupServiceGroup(t, db, 2, false)
	memberGroupID := seedGroupServiceGroup(t, db, 2, false)
	otherUserGroupID := seedGroupServiceGroup(t, db, 2, false)
	updateManagedGroupTestFields(t, db, adminGroupID, "Admin group", "Admin description", "https://example.com/admin.png")
	updateManagedGroupTestFields(t, db, moderatorGroupID, "Moderator group", "Moderator description", "https://example.com/moderator.png")
	setGroupEnterpriseMarker(t, db, adminGroupID, true)

	seedGroupServiceMembership(t, db, adminGroupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, adminGroupID, 2, memberRoleID)
	seedGroupServiceMembership(t, db, adminGroupID, 3, memberRoleID)
	seedGroupServiceMembership(t, db, moderatorGroupID, 1, moderatorRoleID)
	seedGroupServiceMembership(t, db, moderatorGroupID, 3, memberRoleID)
	seedGroupServiceMembership(t, db, memberGroupID, 1, memberRoleID)
	seedGroupServiceMembership(t, db, otherUserGroupID, 4, adminRoleID)

	seedGroupCategory(t, db, 1, "Travel")
	seedGroupCategory(t, db, 2, "Board games")
	seedManagedGroupCategory(t, db, adminGroupID, 1)
	seedManagedGroupCategory(t, db, adminGroupID, 2)
	seedManagedGroupCategory(t, db, moderatorGroupID, 1)

	result, err := service.GetManagedGroups(context.Background(), 1)

	if err != nil {
		t.Fatalf("GetManagedGroups returned error: %v", err)
	}
	if result == nil {
		t.Fatal("GetManagedGroups returned nil result")
	}
	if len(result.Admin) != 1 || result.Admin[0].ID != adminGroupID {
		t.Fatalf("admin groups = %#v, want only group %d", result.Admin, adminGroupID)
	}
	if len(result.Moderator) != 1 || result.Moderator[0].ID != moderatorGroupID {
		t.Fatalf("moderator groups = %#v, want only group %d", result.Moderator, moderatorGroupID)
	}

	admin := result.Admin[0]
	if admin.Name != "Admin group" || admin.SmallDescription != "Admin description" ||
		admin.MemberCount != 3 || admin.Image != "https://example.com/admin.png" || !admin.Enterprise {
		t.Fatalf("admin item = %#v, want mapped group fields and three members", admin)
	}
	if len(admin.Categories) != 2 || admin.Categories[0] != "Board games" || admin.Categories[1] != "Travel" {
		t.Fatalf("admin categories = %#v, want stable name order", admin.Categories)
	}

	moderator := result.Moderator[0]
	if moderator.Name != "Moderator group" || moderator.SmallDescription != "Moderator description" ||
		moderator.MemberCount != 2 || moderator.Image != "https://example.com/moderator.png" || moderator.Enterprise {
		t.Fatalf("moderator item = %#v, want mapped group fields and two members", moderator)
	}
	if len(moderator.Categories) != 1 || moderator.Categories[0] != "Travel" {
		t.Fatalf("moderator categories = %#v, want [Travel]", moderator.Categories)
	}
}

func TestGroupServiceGetManagedGroupsReturnsNonNilEmptyRoleSections(t *testing.T) {
	db := newGroupServiceDB(t)
	service := servicegroups.NewGroupService(
		&testLogger{},
		servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: db}),
	)
	seedGroupServiceUser(t, db, 1)

	result, err := service.GetManagedGroups(context.Background(), 1)

	if err != nil {
		t.Fatalf("GetManagedGroups returned error: %v", err)
	}
	if result == nil {
		t.Fatal("GetManagedGroups returned nil result")
	}
	if result.Admin == nil || len(result.Admin) != 0 {
		t.Fatalf("admin = %#v, want non-nil empty slice", result.Admin)
	}
	if result.Moderator == nil || len(result.Moderator) != 0 {
		t.Fatalf("moderator = %#v, want non-nil empty slice", result.Moderator)
	}
}

func updateManagedGroupTestFields(t *testing.T, db *gorm.DB, groupID uint, name, smallDescription, image string) {
	t.Helper()

	if err := db.Model(&groups.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]interface{}{
			"name":              name,
			"small_description": smallDescription,
			"image":             image,
		}).Error; err != nil {
		t.Fatalf("update managed group test fields for %d: %v", groupID, err)
	}
}

func setGroupEnterpriseMarker(t *testing.T, db *gorm.DB, groupID uint, enterprise bool) {
	t.Helper()

	if err := db.Model(&groups.Group{}).Where("id = ?", groupID).Update("enterprise", enterprise).Error; err != nil {
		t.Fatalf("set enterprise marker for group %d: %v", groupID, err)
	}
}

func seedManagedGroupCategory(t *testing.T, db *gorm.DB, groupID, categoryID uint) {
	t.Helper()

	link := groups.GroupGroupCategory{
		GroupID:         groupID,
		GroupCategoryID: categoryID,
	}
	if err := db.Create(&link).Error; err != nil {
		t.Fatalf("link category %d to group %d: %v", categoryID, groupID, err)
	}
}
