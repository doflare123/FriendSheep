package group

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type groupSearchStoreStub struct {
	page  groupSearchPage
	err   error
	calls int
	ctx   context.Context
	input groupSearchQuery
}

func (s *groupSearchStoreStub) SearchGroups(ctx context.Context, input groupSearchQuery) (groupSearchPage, error) {
	s.calls++
	s.ctx = ctx
	s.input = input
	return s.page, s.err
}

func TestGroupSearchServiceNormalizesDelegatesAndBuildsPage(t *testing.T) {
	store := &groupSearchStoreStub{page: groupSearchPage{
		Items: []groupSearchItemView{{ID: 17, Name: "Board Club", Enterprise: true, ViewerSubscribed: true}},
		Total: 5,
	}}
	service := &groupService{logger: &managedGroupsLoggerStub{}, search: store}
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("search"), "request-context")
	isPrivate := false

	result, err := service.SearchGroups(ctx, 73, GroupSearchInput{
		Query:       "  board  ",
		CategoryIDs: []uint{3, 7},
		IsPrivate:   &isPrivate,
		City:        "  Moscow  ",
		SortBy:      "memberCount",
		SortOrder:   "asc",
		Page:        2,
		Limit:       2,
	})

	if err != nil {
		t.Fatalf("SearchGroups returned error: %v", err)
	}
	if store.calls != 1 || store.ctx != ctx || store.ctx.Value(contextKey("search")) != "request-context" {
		t.Fatalf("store call = count:%d context:%v, want exact request context once", store.calls, store.ctx)
	}
	if store.input.ViewerID != 73 || store.input.Query != "board" || store.input.City != "Moscow" || store.input.IsPrivate != &isPrivate ||
		store.input.SortBy != "memberCount" || store.input.SortOrder != "asc" || store.input.Page != 2 || store.input.Limit != 2 ||
		!reflect.DeepEqual(store.input.CategoryIDs, []uint{3, 7}) {
		t.Fatalf("store input = %#v, want normalized exact filters", store.input)
	}
	if result == nil || len(result.Items) != 1 || result.Items[0].ID != 17 || !result.Items[0].Enterprise || !result.Items[0].IsSubscribed ||
		result.Total != 5 || result.Page != 2 || result.Limit != 2 || result.TotalPages != 3 || !result.HasMore {
		t.Fatalf("result = %#v, want mapped page metadata and item", result)
	}
}

func TestGroupSearchServiceReturnsNonNilEmptyItems(t *testing.T) {
	store := &groupSearchStoreStub{page: groupSearchPage{Total: 0}}
	service := &groupService{logger: &managedGroupsLoggerStub{}, search: store}

	result, err := service.SearchGroups(context.Background(), 0, GroupSearchInput{
		SortBy: "createdAt", SortOrder: "desc", Page: 1, Limit: 20,
	})

	if err != nil {
		t.Fatalf("SearchGroups returned error: %v", err)
	}
	if result.Items == nil || len(result.Items) != 0 || result.Total != 0 || result.TotalPages != 0 || result.HasMore {
		t.Fatalf("result = %#v, want non-nil empty page", result)
	}
}

func TestGroupSearchServiceRejectsInvalidInputBeforeStore(t *testing.T) {
	duplicateIDs := []uint{1, 1}
	tests := []struct {
		name  string
		input GroupSearchInput
	}{
		{name: "page negative", input: GroupSearchInput{SortBy: "createdAt", SortOrder: "desc", Page: -1, Limit: 20}},
		{name: "page above maximum", input: GroupSearchInput{SortBy: "createdAt", SortOrder: "desc", Page: MaxGroupSearchPage + 1, Limit: 20}},
		{name: "limit negative", input: GroupSearchInput{SortBy: "createdAt", SortOrder: "desc", Page: 1, Limit: -1}},
		{name: "limit above maximum", input: GroupSearchInput{SortBy: "createdAt", SortOrder: "desc", Page: 1, Limit: MaxGroupSearchLimit + 1}},
		{name: "unknown sort", input: GroupSearchInput{SortBy: "popular", SortOrder: "desc", Page: 1, Limit: 20}},
		{name: "unknown order", input: GroupSearchInput{SortBy: "createdAt", SortOrder: "sideways", Page: 1, Limit: 20}},
		{name: "zero category", input: GroupSearchInput{CategoryIDs: []uint{0}, SortBy: "createdAt", SortOrder: "desc", Page: 1, Limit: 20}},
		{name: "duplicate category", input: GroupSearchInput{CategoryIDs: duplicateIDs, SortBy: "createdAt", SortOrder: "desc", Page: 1, Limit: 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &groupSearchStoreStub{}
			service := &groupService{logger: &managedGroupsLoggerStub{}, search: store}

			result, err := service.SearchGroups(context.Background(), 0, tt.input)

			if result != nil || !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("result/error = %#v/%v, want nil/ErrInvalidInput", result, err)
			}
			if store.calls != 0 {
				t.Fatalf("store calls = %d, want 0", store.calls)
			}
		})
	}
}

func TestGroupSearchServicePropagatesStoreError(t *testing.T) {
	wantErr := errors.New("search unavailable")
	store := &groupSearchStoreStub{err: wantErr}
	service := &groupService{logger: &managedGroupsLoggerStub{}, search: store}

	result, err := service.SearchGroups(context.Background(), 0, GroupSearchInput{
		SortBy: "name", SortOrder: "asc", Page: 1, Limit: 20,
	})

	if result != nil || !errors.Is(err, wantErr) {
		t.Fatalf("result/error = %#v/%v, want nil/store error", result, err)
	}
}
