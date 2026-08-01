package tests

import (
	"friendship/models/dto"
	"friendship/models/groups"
	servicegroups "friendship/services/groups"
	"reflect"
	"testing"
)

func TestGroupServiceGetSubscribedGroupsReturnsOnlyOrdinaryMembershipsWithStablePages(t *testing.T) {
	db := newGroupServiceDB(t)
	service := servicegroups.NewGroupService(
		&testLogger{},
		servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: db}),
	)

	memberRoleID := seedGroupServiceRole(t, db, groups.RoleMember)
	adminRoleID := seedGroupServiceRole(t, db, groups.RoleAdmin)
	moderatorRoleID := seedGroupServiceRole(t, db, groups.RoleModerator)
	for _, userID := range []uint{1, 2, 3} {
		seedGroupServiceUser(t, db, userID)
	}

	memberFirstID := seedGroupServiceGroup(t, db, 2, false)
	memberSecondID := seedGroupServiceGroup(t, db, 2, false)
	memberThirdID := seedGroupServiceGroup(t, db, 2, false)
	adminGroupID := seedGroupServiceGroup(t, db, 2, false)
	moderatorGroupID := seedGroupServiceGroup(t, db, 2, false)
	updateManagedGroupTestFields(t, db, memberThirdID, "Newest member group", "Newest subscription", "https://example.com/newest.png")

	seedGroupServiceMembership(t, db, memberFirstID, 1, memberRoleID)
	seedGroupServiceMembership(t, db, memberSecondID, 1, memberRoleID)
	seedGroupServiceMembership(t, db, memberThirdID, 1, memberRoleID)
	seedGroupServiceMembership(t, db, memberThirdID, 2, memberRoleID)
	seedGroupServiceMembership(t, db, memberThirdID, 3, memberRoleID)
	seedGroupServiceMembership(t, db, adminGroupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, moderatorGroupID, 1, moderatorRoleID)

	seedGroupCategory(t, db, 1, "Travel")
	seedGroupCategory(t, db, 2, "Board games")
	seedManagedGroupCategory(t, db, memberThirdID, 1)
	seedManagedGroupCategory(t, db, memberThirdID, 2)

	firstPage, err := service.GetSubscribedGroups(t.Context(), 1, 1, 2)
	if err != nil {
		t.Fatalf("GetSubscribedGroups first page returned error: %v", err)
	}
	if firstPage.Total != 3 || firstPage.Page != 1 || firstPage.Limit != 2 || !firstPage.HasMore {
		t.Fatalf("first page metadata = %#v, want total:3 page:1 limit:2 hasMore:true", firstPage)
	}
	if got, want := subscribedGroupIDs(firstPage), []uint{memberThirdID, memberSecondID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first page IDs = %v, want stable descending %v", got, want)
	}
	newest := firstPage.Items[0]
	if newest.Name != "Newest member group" || newest.SmallDescription != "Newest subscription" || newest.Image != "https://example.com/newest.png" || newest.MemberCount != 3 {
		t.Fatalf("newest item = %#v, want mapped group data and all three members", newest)
	}
	if got, want := newest.Categories, []string{"Board games", "Travel"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("newest categories = %v, want stable name order %v", got, want)
	}

	secondPage, err := service.GetSubscribedGroups(t.Context(), 1, 2, 2)
	if err != nil {
		t.Fatalf("GetSubscribedGroups second page returned error: %v", err)
	}
	if secondPage.Total != 3 || secondPage.Page != 2 || secondPage.Limit != 2 || secondPage.HasMore {
		t.Fatalf("second page metadata = %#v, want total:3 page:2 limit:2 hasMore:false", secondPage)
	}
	if got, want := subscribedGroupIDs(secondPage), []uint{memberFirstID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("second page IDs = %v, want %v", got, want)
	}

	for _, item := range append(firstPage.Items, secondPage.Items...) {
		if item.ID == adminGroupID || item.ID == moderatorGroupID {
			t.Fatalf("admin/moderator group %d leaked into ordinary subscriptions: %#v", item.ID, append(firstPage.Items, secondPage.Items...))
		}
	}
}

func TestGroupServiceGetSubscribedGroupsReturnsEmptyItemsArray(t *testing.T) {
	db := newGroupServiceDB(t)
	service := servicegroups.NewGroupService(
		&testLogger{},
		servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: db}),
	)
	seedGroupServiceUser(t, db, 1)

	result, err := service.GetSubscribedGroups(t.Context(), 1, 1, 20)
	if err != nil {
		t.Fatalf("GetSubscribedGroups returned error: %v", err)
	}
	if result.Items == nil || len(result.Items) != 0 || result.Total != 0 || result.HasMore {
		t.Fatalf("empty subscriptions result = %#v, want non-nil empty items with zero metadata", result)
	}
}

func subscribedGroupIDs(result *dto.SubscribedGroupsResponseDto) []uint {
	ids := make([]uint, 0, len(result.Items))
	for _, item := range result.Items {
		ids = append(ids, item.ID)
	}
	return ids
}
