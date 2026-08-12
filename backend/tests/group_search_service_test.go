package tests

import (
	"context"
	"reflect"
	"testing"
	"time"

	"friendship/models/dto"
	"friendship/models/groups"
	servicegroups "friendship/services/groups"

	"gorm.io/gorm"
)

func TestGORMGroupSearchAppliesCaseInsensitiveFiltersAndCategoryOR(t *testing.T) {
	database := newGroupServiceDB(t)
	service := servicegroups.NewGroupService(
		&testLogger{},
		servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: database}),
	)

	for _, userID := range []uint{1, 2, 3} {
		seedGroupServiceUser(t, database, userID)
	}
	memberRoleID := seedGroupServiceRole(t, database, groups.RoleMember)
	seedGroupCategory(t, database, 1, "Board Games")
	seedGroupCategory(t, database, 2, "Travel")
	seedGroupCategory(t, database, 3, "Sport")

	alphaID := seedGroupServiceGroup(t, database, 1, false)
	betaID := seedGroupServiceGroup(t, database, 1, false)
	privateID := seedGroupServiceGroup(t, database, 1, true)
	updateGroupSearchFields(t, database, alphaID, "Alpha Board Club", "Friendly TABLE games", "Moscow Center", true, time.Date(2028, 1, 1, 12, 0, 0, 0, time.UTC))
	updateGroupSearchFields(t, database, betaID, "Beta Travelers", "Trips", "Moscow", false, time.Date(2028, 1, 2, 12, 0, 0, 0, time.UTC))
	updateGroupSearchFields(t, database, privateID, "Private Sport", "Training", "Saint Petersburg", false, time.Date(2028, 1, 3, 12, 0, 0, 0, time.UTC))
	seedManagedGroupCategory(t, database, alphaID, 1)
	seedManagedGroupCategory(t, database, alphaID, 2)
	seedManagedGroupCategory(t, database, betaID, 2)
	seedManagedGroupCategory(t, database, privateID, 3)
	seedGroupServiceMembership(t, database, alphaID, 1, memberRoleID)
	seedGroupServiceMembership(t, database, alphaID, 2, memberRoleID)
	seedGroupServiceMembership(t, database, betaID, 1, memberRoleID)

	isPrivate := false
	filtered, err := service.SearchGroups(context.Background(), 2, servicegroups.GroupSearchInput{
		Query:       "  tAbLe  ",
		CategoryIDs: []uint{1, 2},
		IsPrivate:   &isPrivate,
		City:        "  OSC  ",
		SortBy:      "createdAt",
		SortOrder:   "desc",
		Page:        1,
		Limit:       20,
	})
	if err != nil {
		t.Fatalf("filtered SearchGroups returned error: %v", err)
	}
	if filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].ID != alphaID {
		t.Fatalf("filtered result = %#v, want only group %d", filtered, alphaID)
	}
	item := filtered.Items[0]
	if item.Name != "Alpha Board Club" || item.SmallDescription != "Friendly TABLE games" || item.MemberCount != 2 ||
		item.Image != "https://example.com/group.png" || item.IsPrivate || !item.Enterprise || !item.IsSubscribed || item.CreatedAt.IsZero() ||
		!reflect.DeepEqual(item.Categories, []string{"Board Games", "Travel"}) {
		t.Fatalf("filtered item = %#v, want exact mapped fields", item)
	}

	categoryPage, err := service.SearchGroups(context.Background(), 3, servicegroups.GroupSearchInput{
		CategoryIDs: []uint{1, 2},
		SortBy:      "createdAt",
		SortOrder:   "asc",
		Page:        1,
		Limit:       20,
	})
	if err != nil {
		t.Fatalf("category SearchGroups returned error: %v", err)
	}
	if categoryPage.Total != 2 || len(categoryPage.Items) != 2 {
		t.Fatalf("category OR result = %#v, want two unique groups", categoryPage)
	}
	if got, want := groupSearchItemIDs(categoryPage), []uint{alphaID, betaID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("category OR IDs = %v, want %v", got, want)
	}
	for _, item := range categoryPage.Items {
		if item.IsSubscribed {
			t.Fatalf("nonmember item %d has isSubscribed=true", item.ID)
		}
	}
}

func TestGORMGroupSearchStableSortingPaginationAndCount(t *testing.T) {
	database := newGroupServiceDB(t)
	service := servicegroups.NewGroupService(
		&testLogger{},
		servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: database}),
	)

	for _, userID := range []uint{1, 2, 3} {
		seedGroupServiceUser(t, database, userID)
	}
	memberRoleID := seedGroupServiceRole(t, database, groups.RoleMember)

	firstID := seedGroupServiceGroup(t, database, 1, false)
	secondID := seedGroupServiceGroup(t, database, 1, false)
	thirdID := seedGroupServiceGroup(t, database, 1, false)
	tiedCreatedAt := time.Date(2028, 2, 1, 12, 0, 0, 0, time.UTC)
	updateGroupSearchFields(t, database, firstID, "same", "First", "Moscow", false, tiedCreatedAt)
	updateGroupSearchFields(t, database, secondID, "SAME", "Second", "Moscow", false, tiedCreatedAt)
	updateGroupSearchFields(t, database, thirdID, "Zero members", "Third", "Moscow", false, tiedCreatedAt)
	seedGroupServiceMembership(t, database, firstID, 1, memberRoleID)
	seedGroupServiceMembership(t, database, firstID, 2, memberRoleID)
	seedGroupServiceMembership(t, database, secondID, 2, memberRoleID)
	seedGroupServiceMembership(t, database, secondID, 3, memberRoleID)

	firstPage, err := service.SearchGroups(context.Background(), 0, servicegroups.GroupSearchInput{
		SortBy: "memberCount", SortOrder: "desc", Page: 1, Limit: 1,
	})
	if err != nil {
		t.Fatalf("first page SearchGroups returned error: %v", err)
	}
	if firstPage.Total != 3 || firstPage.TotalPages != 3 || !firstPage.HasMore || len(firstPage.Items) != 1 || firstPage.Items[0].ID != secondID || firstPage.Items[0].MemberCount != 2 {
		t.Fatalf("first page = %#v, want higher-ID member-count tie with full total", firstPage)
	}
	if firstPage.Items[0].IsSubscribed {
		t.Fatal("anonymous group search item has isSubscribed=true")
	}

	secondPage, err := service.SearchGroups(context.Background(), 0, servicegroups.GroupSearchInput{
		SortBy: "memberCount", SortOrder: "desc", Page: 2, Limit: 1,
	})
	if err != nil {
		t.Fatalf("second page SearchGroups returned error: %v", err)
	}
	if secondPage.Total != 3 || len(secondPage.Items) != 1 || secondPage.Items[0].ID != firstID || !secondPage.HasMore {
		t.Fatalf("second page = %#v, want stable next member-count tie", secondPage)
	}

	lastPage, err := service.SearchGroups(context.Background(), 0, servicegroups.GroupSearchInput{
		SortBy: "memberCount", SortOrder: "desc", Page: 3, Limit: 1,
	})
	if err != nil {
		t.Fatalf("last page SearchGroups returned error: %v", err)
	}
	if lastPage.Total != 3 || len(lastPage.Items) != 1 || lastPage.Items[0].ID != thirdID || lastPage.Items[0].MemberCount != 0 || lastPage.HasMore {
		t.Fatalf("last page = %#v, want zero-member group and hasMore=false", lastPage)
	}

	nameAscending, err := service.SearchGroups(context.Background(), 0, servicegroups.GroupSearchInput{
		Query: "same", SortBy: "name", SortOrder: "asc", Page: 1, Limit: 20,
	})
	if err != nil {
		t.Fatalf("name SearchGroups returned error: %v", err)
	}
	if got, want := groupSearchItemIDs(nameAscending), []uint{firstID, secondID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("case-insensitive name tie IDs = %v, want stable ascending %v", got, want)
	}

	createdDescending, err := service.SearchGroups(context.Background(), 0, servicegroups.GroupSearchInput{
		SortBy: "createdAt", SortOrder: "desc", Page: 1, Limit: 20,
	})
	if err != nil {
		t.Fatalf("createdAt SearchGroups returned error: %v", err)
	}
	if got, want := groupSearchItemIDs(createdDescending), []uint{thirdID, secondID, firstID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("createdAt tie IDs = %v, want stable descending %v", got, want)
	}
}

func updateGroupSearchFields(t *testing.T, database *gorm.DB, groupID uint, name, smallDescription, city string, enterprise bool, createdAt time.Time) {
	t.Helper()

	if err := database.Model(&groups.Group{}).Where("id = ?", groupID).Updates(map[string]interface{}{
		"name": name, "small_description": smallDescription, "city": city, "enterprise": enterprise, "created_at": createdAt,
	}).Error; err != nil {
		t.Fatalf("update group search fields for %d: %v", groupID, err)
	}
}

func groupSearchItemIDs(result *dto.GroupSearchResponseDto) []uint {
	ids := make([]uint, 0, len(result.Items))
	for _, item := range result.Items {
		ids = append(ids, item.ID)
	}
	return ids
}
