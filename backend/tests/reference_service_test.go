package tests

import (
	"context"
	"errors"
	"math"
	"testing"

	"friendship/services/references"
)

type referenceServiceContextKey struct{}

type referenceStoreStub struct {
	snapshot      references.ReferenceSnapshot
	loadErr       error
	searchPage    references.GenreSearchPage
	searchErr     error
	loadCalls     int
	searchCalls   int
	loadContext   context.Context
	searchContext context.Context
	searchQuery   references.GenreSearchQuery
}

func (s *referenceStoreStub) LoadReferences(ctx context.Context) (references.ReferenceSnapshot, error) {
	s.loadCalls++
	s.loadContext = ctx
	return s.snapshot, s.loadErr
}

func (s *referenceStoreStub) SearchGenres(ctx context.Context, query references.GenreSearchQuery) (references.GenreSearchPage, error) {
	s.searchCalls++
	s.searchContext = ctx
	s.searchQuery = query
	return s.searchPage, s.searchErr
}

func TestReferenceServiceGetReferencesDelegatesContextAndMapsGeneralSnapshot(t *testing.T) {
	store := &referenceStoreStub{
		snapshot: references.ReferenceSnapshot{
			EventTypes:      []references.ReferenceItem{{ID: 1, Name: "Event type"}},
			Locations:       []references.ReferenceItem{{ID: 2, Name: "Location"}},
			AgeLimits:       []references.ReferenceItem{{ID: 3, Name: "18+"}},
			Statuses:        []references.ReferenceItem{{ID: 4, Name: "Open"}},
			GroupCategories: []references.ReferenceItem{{ID: 6, Name: "Category"}},
			GroupActionTypes: []references.ActionReferenceItem{
				{ID: 7, Code: "create_event", Name: "Create event"},
			},
		},
	}
	service := references.NewReferenceService(store)
	ctx := context.WithValue(context.Background(), referenceServiceContextKey{}, "references")

	result, err := service.GetReferences(ctx)

	if err != nil {
		t.Fatalf("GetReferences returned error: %v", err)
	}
	if store.loadCalls != 1 {
		t.Fatalf("load calls = %d, want 1", store.loadCalls)
	}
	if store.loadContext.Value(referenceServiceContextKey{}) != "references" {
		t.Fatal("request context was not passed to the reference store")
	}
	if result == nil {
		t.Fatal("result = nil")
	}
	if len(result.EventTypes) != 1 || result.EventTypes[0].ID != 1 || result.EventTypes[0].Name != "Event type" {
		t.Fatalf("event types = %#v, want mapped item", result.EventTypes)
	}
	if len(result.Locations) != 1 || result.Locations[0].ID != 2 {
		t.Fatalf("locations = %#v, want mapped item", result.Locations)
	}
	if len(result.AgeLimits) != 1 || result.AgeLimits[0].ID != 3 {
		t.Fatalf("age limits = %#v, want mapped item", result.AgeLimits)
	}
	if len(result.Statuses) != 1 || result.Statuses[0].ID != 4 {
		t.Fatalf("statuses = %#v, want mapped item", result.Statuses)
	}
	if len(result.GroupCategories) != 1 || result.GroupCategories[0].ID != 6 {
		t.Fatalf("group categories = %#v, want mapped item", result.GroupCategories)
	}
	if len(result.GroupActionTypes) != 1 ||
		result.GroupActionTypes[0].ID != 7 ||
		result.GroupActionTypes[0].Code != "create_event" ||
		result.GroupActionTypes[0].Name != "Create event" {
		t.Fatalf("group action types = %#v, want mapped item", result.GroupActionTypes)
	}
}

func TestReferenceServiceGetReferencesPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("reference storage failed")
	store := &referenceStoreStub{loadErr: storeErr}
	service := references.NewReferenceService(store)

	result, err := service.GetReferences(context.Background())

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want store error", err)
	}
}

func TestReferenceServiceSearchGenresNormalizesDelegatesAndMapsPage(t *testing.T) {
	store := &referenceStoreStub{
		searchPage: references.GenreSearchPage{
			Items: []references.ReferenceItem{
				{ID: 11, Name: "Board Games"},
				{ID: 12, Name: "Board Tactics"},
			},
			Total: 5,
		},
	}
	service := references.NewReferenceService(store)
	ctx := context.WithValue(context.Background(), referenceServiceContextKey{}, "search")

	result, err := service.SearchGenres(ctx, references.GenreSearchInput{
		Query: "  Board  ",
		Page:  2,
		Limit: 2,
	})

	if err != nil {
		t.Fatalf("SearchGenres returned error: %v", err)
	}
	if store.searchCalls != 1 {
		t.Fatalf("search calls = %d, want 1", store.searchCalls)
	}
	if store.searchContext.Value(referenceServiceContextKey{}) != "search" {
		t.Fatal("request context was not passed to the genre store")
	}
	if store.searchQuery != (references.GenreSearchQuery{Query: "Board", Page: 2, Limit: 2}) {
		t.Fatalf("search query = %#v, want trimmed query and exact pagination", store.searchQuery)
	}
	if result == nil ||
		result.Total != 5 ||
		result.Page != 2 ||
		result.Limit != 2 ||
		!result.HasMore {
		t.Fatalf("result metadata = %#v, want total=5 page=2 limit=2 hasMore=true", result)
	}
	wantItems := []struct {
		ID   uint
		Name string
	}{
		{ID: 11, Name: "Board Games"},
		{ID: 12, Name: "Board Tactics"},
	}
	if len(result.Items) != len(wantItems) {
		t.Fatalf("items = %#v, want %d mapped items", result.Items, len(wantItems))
	}
	for index, want := range wantItems {
		if result.Items[index].ID != want.ID || result.Items[index].Name != want.Name {
			t.Fatalf("item %d = %#v, want %#v", index, result.Items[index], want)
		}
	}
}

func TestReferenceServiceSearchGenresAppliesDefaultsAndReturnsNonNilEmptyItems(t *testing.T) {
	store := &referenceStoreStub{searchPage: references.GenreSearchPage{}}
	service := references.NewReferenceService(store)

	result, err := service.SearchGenres(context.Background(), references.GenreSearchInput{})

	if err != nil {
		t.Fatalf("SearchGenres returned error: %v", err)
	}
	wantQuery := references.GenreSearchQuery{
		Page:  references.DefaultGenrePage,
		Limit: references.DefaultGenreLimit,
	}
	if store.searchQuery != wantQuery {
		t.Fatalf("search query = %#v, want %#v", store.searchQuery, wantQuery)
	}
	if result == nil ||
		result.Items == nil ||
		len(result.Items) != 0 ||
		result.Total != 0 ||
		result.Page != references.DefaultGenrePage ||
		result.Limit != references.DefaultGenreLimit ||
		result.HasMore {
		t.Fatalf("result = %#v, want normalized empty page", result)
	}
}

func TestReferenceServiceSearchGenresRejectsInvalidPaginationWithoutStoreCall(t *testing.T) {
	tests := []struct {
		name    string
		input   references.GenreSearchInput
		wantErr error
	}{
		{
			name:    "negative page",
			input:   references.GenreSearchInput{Page: -1, Limit: 10},
			wantErr: references.ErrInvalidGenrePage,
		},
		{
			name:    "negative limit",
			input:   references.GenreSearchInput{Page: 1, Limit: -1},
			wantErr: references.ErrInvalidGenreLimit,
		},
		{
			name:    "limit above maximum",
			input:   references.GenreSearchInput{Page: 1, Limit: references.MaxGenreLimit + 1},
			wantErr: references.ErrInvalidGenreLimit,
		},
		{
			name:    "offset overflow",
			input:   references.GenreSearchInput{Page: math.MaxInt, Limit: references.MaxGenreLimit},
			wantErr: references.ErrInvalidGenrePage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &referenceStoreStub{}
			service := references.NewReferenceService(store)

			result, err := service.SearchGenres(context.Background(), tt.input)

			if result != nil {
				t.Fatalf("result = %#v, want nil", result)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if store.searchCalls != 0 {
				t.Fatalf("search calls = %d, want 0", store.searchCalls)
			}
		})
	}
}

func TestReferenceServiceSearchGenresPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("genre storage failed")
	store := &referenceStoreStub{searchErr: storeErr}
	service := references.NewReferenceService(store)

	result, err := service.SearchGenres(context.Background(), references.GenreSearchInput{Page: 1, Limit: 10})

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want store error", err)
	}
}

var _ references.ReferenceStore = (*referenceStoreStub)(nil)
