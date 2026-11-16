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
	cityPage      references.CitySearchPage
	cityErr       error
	loadCalls     int
	searchCalls   int
	cityCalls     int
	loadContext   context.Context
	searchContext context.Context
	searchQuery   references.GenreSearchQuery
	cityContext   context.Context
	cityQuery     references.CitySearchQuery
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

func (s *referenceStoreStub) SearchCities(ctx context.Context, query references.CitySearchQuery) (references.CitySearchPage, error) {
	s.cityCalls++
	s.cityContext = ctx
	s.cityQuery = query
	return s.cityPage, s.cityErr
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

func TestReferenceServiceSearchCitiesNormalizesDelegatesAndMapsPage(t *testing.T) {
	store := &referenceStoreStub{cityPage: references.CitySearchPage{
		Items: []references.CityReferenceItem{
			{ID: 11, Name: "Советск", Region: "Калининградская обл"},
			{ID: 12, Name: "Советск", Region: "Кировская обл"},
		},
		Total: 5,
	}}
	service := references.NewReferenceService(store)
	ctx := context.WithValue(context.Background(), referenceServiceContextKey{}, "cities")

	result, err := service.SearchCities(ctx, references.CitySearchInput{Query: "  Совет  ", Page: 2, Limit: 2})

	if err != nil {
		t.Fatalf("SearchCities returned error: %v", err)
	}
	if store.cityCalls != 1 || store.cityContext.Value(referenceServiceContextKey{}) != "cities" {
		t.Fatalf("city store call = %d, context = %v; want one call with request context", store.cityCalls, store.cityContext)
	}
	if store.cityQuery != (references.CitySearchQuery{Query: "Совет", Page: 2, Limit: 2}) {
		t.Fatalf("city query = %#v, want trimmed query and pagination", store.cityQuery)
	}
	if result == nil || result.Total != 5 || result.Page != 2 || result.Limit != 2 || !result.HasMore || len(result.Items) != 2 {
		t.Fatalf("city result = %#v, want complete paginated page", result)
	}
	if result.Items[0].ID != 11 || result.Items[0].Name != "Советск" || result.Items[0].Region != "Калининградская обл" ||
		result.Items[1].ID != 12 || result.Items[1].Region != "Кировская обл" {
		t.Fatalf("city items = %#v, want mapped cities including regions", result.Items)
	}
}

func TestReferenceServiceSearchCitiesDefaultsAndValidation(t *testing.T) {
	t.Run("defaults and empty page", func(t *testing.T) {
		store := &referenceStoreStub{}
		result, err := references.NewReferenceService(store).SearchCities(context.Background(), references.CitySearchInput{})
		if err != nil {
			t.Fatalf("SearchCities returned error: %v", err)
		}
		if store.cityQuery != (references.CitySearchQuery{Page: references.DefaultCityPage, Limit: references.DefaultCityLimit}) ||
			result == nil || result.Items == nil || len(result.Items) != 0 || result.Total != 0 || result.Page != references.DefaultCityPage || result.Limit != references.DefaultCityLimit || result.HasMore {
			t.Fatalf("city defaults = query %#v result %#v", store.cityQuery, result)
		}
	})

	for _, tt := range []struct {
		name  string
		input references.CitySearchInput
		want  error
	}{
		{name: "invalid page", input: references.CitySearchInput{Page: -1, Limit: 10}, want: references.ErrInvalidCityPage},
		{name: "invalid limit", input: references.CitySearchInput{Page: 1, Limit: references.MaxCityLimit + 1}, want: references.ErrInvalidCityLimit},
		{name: "offset overflow", input: references.CitySearchInput{Page: math.MaxInt, Limit: references.MaxCityLimit}, want: references.ErrInvalidCityPage},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &referenceStoreStub{}
			result, err := references.NewReferenceService(store).SearchCities(context.Background(), tt.input)
			if result != nil || !errors.Is(err, tt.want) || store.cityCalls != 0 {
				t.Fatalf("result = %#v, err = %v, calls = %d; want nil, %v, 0", result, err, store.cityCalls, tt.want)
			}
		})
	}
}

func TestReferenceServiceSearchCitiesPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("city storage failed")
	result, err := references.NewReferenceService(&referenceStoreStub{cityErr: storeErr}).SearchCities(context.Background(), references.CitySearchInput{Page: 1, Limit: 10})
	if result != nil || !errors.Is(err, storeErr) {
		t.Fatalf("result = %#v, err = %v; want nil and store error", result, err)
	}
}

var _ references.ReferenceStore = (*referenceStoreStub)(nil)
