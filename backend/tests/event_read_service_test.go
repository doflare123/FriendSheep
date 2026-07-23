package tests

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	servicesevents "friendship/services/events"
)

type eventReadServiceContextKey struct{}

type eventReadStoreStub struct {
	searchPage    servicesevents.EventSearchPageView
	searchErr     error
	groupView     servicesevents.EventGroupEventsView
	groupErr      error
	detailsView   servicesevents.EventDetailsView
	detailsErr    error
	searchCalls   int
	groupCalls    int
	detailsCalls  int
	searchContext context.Context
	groupContext  context.Context
	detailContext context.Context
	searchQuery   servicesevents.EventSearchQuery
	groupQuery    servicesevents.EventGroupEventsQuery
	detailsQuery  servicesevents.EventDetailsQuery
}

func (s *eventReadStoreStub) SearchEvents(ctx context.Context, query servicesevents.EventSearchQuery) (servicesevents.EventSearchPageView, error) {
	s.searchCalls++
	s.searchContext = ctx
	s.searchQuery = query
	return s.searchPage, s.searchErr
}

func (s *eventReadStoreStub) GetGroupEvents(ctx context.Context, query servicesevents.EventGroupEventsQuery) (servicesevents.EventGroupEventsView, error) {
	s.groupCalls++
	s.groupContext = ctx
	s.groupQuery = query
	return s.groupView, s.groupErr
}

func (s *eventReadStoreStub) GetEventDetails(ctx context.Context, query servicesevents.EventDetailsQuery) (servicesevents.EventDetailsView, error) {
	s.detailsCalls++
	s.detailContext = ctx
	s.detailsQuery = query
	return s.detailsView, s.detailsErr
}

func TestEventReadServiceSearchEventsNormalizesDelegatesAndBuildsPage(t *testing.T) {
	start := time.Date(2028, 1, 2, 15, 4, 5, 0, time.UTC)
	groupID := uint(42)
	dateFrom := start.Add(-time.Hour)
	dateTo := start.Add(time.Hour)
	hasFreeSlots := true
	store := &eventReadStoreStub{
		searchPage: servicesevents.EventSearchPageView{
			Total: 5,
			Items: []servicesevents.EventSearchItemView{
				{
					ID:                17,
					Title:             "Настольные игры",
					Group:             servicesevents.EventReadGroupView{ID: 42, Name: "Клуб"},
					ImageURL:          "https://example.com/event.png",
					ParticipantsCount: 3,
					MaxUsers:          8,
					Duration:          90,
					StartTime:         start,
					EventType:         "Игра",
					LocationType:      "Offline",
					City:              "Москва",
					Genres:            []string{"Strategy", "Party"},
					ViewerSubscribed:  true,
				},
			},
		},
	}
	service := servicesevents.NewEventReadService(&testLogger{}, store)
	ctx := context.WithValue(context.Background(), eventReadServiceContextKey{}, "search")

	result, err := service.SearchEvents(ctx, 9, servicesevents.EventSearchInput{
		Query:                "  board  ",
		GroupID:              &groupID,
		CategoryIDs:          []uint{1, 2},
		ExcludeCategoryIDs:   []uint{3},
		GenreIDs:             []uint{4},
		ExcludeGenreIDs:      []uint{5},
		EventTypeIDs:         []uint{6},
		ExcludeEventTypeIDs:  []uint{7},
		LocationTypes:        []string{" Online "},
		ExcludeLocationTypes: []string{" OFF-LINE "},
		City:                 "  Moscow  ",
		DateFrom:             &dateFrom,
		DateTo:               &dateTo,
		HasFreeSlots:         &hasFreeSlots,
		Page:                 2,
		Limit:                2,
	})

	if err != nil {
		t.Fatalf("SearchEvents returned error: %v", err)
	}
	if store.searchCalls != 1 {
		t.Fatalf("search calls = %d, want 1", store.searchCalls)
	}
	if store.searchContext.Value(eventReadServiceContextKey{}) != "search" {
		t.Fatal("контекст не передан в хранилище")
	}
	if store.searchQuery.ViewerID != 9 ||
		store.searchQuery.Query != "board" ||
		store.searchQuery.GroupID == nil ||
		*store.searchQuery.GroupID != 42 ||
		store.searchQuery.City != "Moscow" ||
		store.searchQuery.Page != 2 ||
		store.searchQuery.Limit != 2 {
		t.Fatalf("search query = %#v, want normalized exact query", store.searchQuery)
	}
	if !reflect.DeepEqual(store.searchQuery.CategoryIDs, []uint{1, 2}) ||
		!reflect.DeepEqual(store.searchQuery.ExcludeCategoryIDs, []uint{3}) ||
		!reflect.DeepEqual(store.searchQuery.GenreIDs, []uint{4}) ||
		!reflect.DeepEqual(store.searchQuery.ExcludeGenreIDs, []uint{5}) ||
		!reflect.DeepEqual(store.searchQuery.EventTypeIDs, []uint{6}) ||
		!reflect.DeepEqual(store.searchQuery.ExcludeEventTypeIDs, []uint{7}) {
		t.Fatalf("ID filters = %#v, want exact delegated filters", store.searchQuery)
	}
	if !reflect.DeepEqual(store.searchQuery.LocationTypes, []string{"online", "онлайн"}) {
		t.Fatalf("LocationTypes = %#v, want normalized online aliases", store.searchQuery.LocationTypes)
	}
	if !reflect.DeepEqual(store.searchQuery.ExcludeLocationTypes, []string{"offline", "off-line", "офлайн", "оффлайн"}) {
		t.Fatalf("ExcludeLocationTypes = %#v, want normalized offline aliases", store.searchQuery.ExcludeLocationTypes)
	}
	if store.searchQuery.DateFrom != &dateFrom ||
		store.searchQuery.DateTo != &dateTo ||
		store.searchQuery.HasFreeSlots != &hasFreeSlots {
		t.Fatalf("pointer filters = %#v, want original values", store.searchQuery)
	}
	if result == nil ||
		result.Total != 5 ||
		result.CurrentPage != 2 ||
		result.Limit != 2 ||
		result.TotalPages != 3 ||
		!result.HasMore {
		t.Fatalf("result metadata = %#v, want total=5 page=2 limit=2 totalPages=3 hasMore=true", result)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %#v, want one mapped item", result.Items)
	}
	item := result.Items[0]
	if item.ID != 17 ||
		item.Title != "Настольные игры" ||
		item.Group.ID != 42 ||
		item.Group.Name != "Клуб" ||
		item.Image != "https://example.com/event.png" ||
		item.ParticipantsCount != 3 ||
		item.MaxUsers != 8 ||
		item.Duration != 90 ||
		item.StartDate != "2028-01-02" ||
		item.EventType != "Игра" ||
		item.LocationType != "Offline" ||
		item.City != "Москва" ||
		!reflect.DeepEqual(item.Genres, []string{"Strategy", "Party"}) ||
		!item.Subscribed {
		t.Fatalf("mapped item = %#v, want all view fields", item)
	}
}

func TestEventReadServiceSearchEventsClampsDirectInput(t *testing.T) {
	store := &eventReadStoreStub{searchPage: servicesevents.EventSearchPageView{}}
	service := servicesevents.NewEventReadService(&testLogger{}, store)

	result, err := service.SearchEvents(context.Background(), 3, servicesevents.EventSearchInput{
		Page:  20000,
		Limit: 1000,
	})

	if err != nil {
		t.Fatalf("SearchEvents returned error: %v", err)
	}
	if store.searchQuery.Page != 10000 || store.searchQuery.Limit != 100 {
		t.Fatalf("query page/limit = %d/%d, want 10000/100", store.searchQuery.Page, store.searchQuery.Limit)
	}
	if result.CurrentPage != 10000 || result.Limit != 100 {
		t.Fatalf("result page/limit = %d/%d, want 10000/100", result.CurrentPage, result.Limit)
	}
	if result.Items == nil {
		t.Fatal("Items = nil, want non-nil empty slice")
	}
}

func TestEventReadServiceSearchEventsAnonymousSubscriptionNewsSkipsStore(t *testing.T) {
	store := &eventReadStoreStub{searchErr: errors.New("хранилище не должно вызываться")}
	service := servicesevents.NewEventReadService(&testLogger{}, store)

	result, err := service.SearchEvents(context.Background(), 0, servicesevents.EventSearchInput{
		OnlySubscriptionNews: true,
	})

	if err != nil {
		t.Fatalf("SearchEvents returned error: %v", err)
	}
	if store.searchCalls != 0 {
		t.Fatalf("search calls = %d, want 0", store.searchCalls)
	}
	if result == nil ||
		result.Items == nil ||
		len(result.Items) != 0 ||
		result.Total != 0 ||
		result.CurrentPage != 1 ||
		result.Limit != 20 ||
		result.TotalPages != 0 ||
		result.HasMore {
		t.Fatalf("result = %#v, want normalized empty response", result)
	}
}

func TestEventReadServiceSearchEventsPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("ошибка поиска в хранилище")
	store := &eventReadStoreStub{searchErr: storeErr}
	service := servicesevents.NewEventReadService(&testLogger{}, store)

	result, err := service.SearchEvents(context.Background(), 5, servicesevents.EventSearchInput{Page: 1, Limit: 20})

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want store error", err)
	}
}

func TestEventReadServiceGetGroupEventsDelegatesAndMapsDTO(t *testing.T) {
	start := time.Date(2028, 2, 3, 12, 0, 0, 0, time.UTC)
	store := &eventReadStoreStub{
		groupView: servicesevents.EventGroupEventsView{
			GroupFound:          true,
			GroupPrivate:        false,
			ViewerIsGroupMember: false,
			Items: []servicesevents.EventShortView{
				{
					ID:               21,
					Title:            "Короткое событие",
					ImageURL:         "https://example.com/short.png",
					MaxUsers:         12,
					CurrentUsers:     4,
					EventTypeID:      6,
					LocationTypeID:   7,
					AgeLimit:         "12+",
					Genres:           []string{"Quest"},
					StartTime:        start,
					Duration:         75,
					GroupID:          42,
					Status:           "Набор",
					ViewerSubscribed: true,
				},
			},
		},
	}
	service := servicesevents.NewEventReadService(&testLogger{}, store)
	ctx := context.WithValue(context.Background(), eventReadServiceContextKey{}, "group")

	result, err := service.GetGroupEvents(ctx, 9, 42)

	if err != nil {
		t.Fatalf("GetGroupEvents returned error: %v", err)
	}
	if store.groupCalls != 1 ||
		store.groupQuery.ViewerID != 9 ||
		store.groupQuery.GroupID != 42 ||
		store.groupContext.Value(eventReadServiceContextKey{}) != "group" {
		t.Fatalf("delegation = calls:%d query:%#v context:%v", store.groupCalls, store.groupQuery, store.groupContext)
	}
	if len(result) != 1 {
		t.Fatalf("result = %#v, want one event", result)
	}
	item := result[0]
	if item.ID != 21 ||
		item.EventID != 21 ||
		item.GroupID != 42 ||
		item.Title != "Короткое событие" ||
		item.ImageURL != "https://example.com/short.png" ||
		item.MaxUsers != 12 ||
		item.CurrentUsers != 4 ||
		item.EventType != 6 ||
		item.LocationType != 7 ||
		item.AgeLimit != "12+" ||
		!reflect.DeepEqual(item.Genres, []string{"Quest"}) ||
		!item.StartTime.Equal(start) ||
		item.Duration != 75 ||
		item.Status != "Набор" ||
		!item.Subscribed {
		t.Fatalf("mapped item = %#v, want all short view fields", item)
	}
}

func TestEventReadServiceGetGroupEventsAccessAndMissingGroup(t *testing.T) {
	tests := []struct {
		name       string
		view       servicesevents.EventGroupEventsView
		wantErr    error
		wantNil    bool
		wantLength int
	}{
		{
			name: "публичная группа доступна постороннему",
			view: servicesevents.EventGroupEventsView{
				GroupFound:          true,
				GroupPrivate:        false,
				ViewerIsGroupMember: false,
				Items:               []servicesevents.EventShortView{},
			},
			wantLength: 0,
		},
		{
			name: "приватная группа доступна участнику",
			view: servicesevents.EventGroupEventsView{
				GroupFound:          true,
				GroupPrivate:        true,
				ViewerIsGroupMember: true,
				Items:               []servicesevents.EventShortView{{ID: 1}},
			},
			wantLength: 1,
		},
		{
			name: "приватная группа запрещена постороннему",
			view: servicesevents.EventGroupEventsView{
				GroupFound:          true,
				GroupPrivate:        true,
				ViewerIsGroupMember: false,
				Items:               []servicesevents.EventShortView{{ID: 1}},
			},
			wantErr: servicesevents.ErrNotGroupMember,
			wantNil: true,
		},
		{
			name: "отсутствующая группа возвращает пустой список",
			view: servicesevents.EventGroupEventsView{
				GroupFound: false,
				Items:      []servicesevents.EventShortView{{ID: 1}},
			},
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &eventReadStoreStub{groupView: tt.view}
			service := servicesevents.NewEventReadService(&testLogger{}, store)

			result, err := service.GetGroupEvents(context.Background(), 9, 42)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("GetGroupEvents returned error: %v", err)
			}
			if tt.wantNil {
				if result != nil {
					t.Fatalf("result = %#v, want nil", result)
				}
				return
			}
			if result == nil {
				t.Fatal("result = nil, want non-nil slice")
			}
			if len(result) != tt.wantLength {
				t.Fatalf("len(result) = %d, want %d; result = %#v", len(result), tt.wantLength, result)
			}
		})
	}
}

func TestEventReadServiceGetGroupEventsPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("ошибка чтения группы")
	store := &eventReadStoreStub{groupErr: storeErr}
	service := servicesevents.NewEventReadService(&testLogger{}, store)

	result, err := service.GetGroupEvents(context.Background(), 9, 42)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want store error", err)
	}
}

func TestEventReadServiceGetEventDetailsDelegatesAndMapsDTO(t *testing.T) {
	start := time.Date(2028, 3, 4, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	created := start.Add(-24 * time.Hour)
	updated := created.Add(time.Hour)
	year := 2028
	store := &eventReadStoreStub{
		detailsView: servicesevents.EventDetailsView{
			Found:               true,
			GroupPrivate:        false,
			ViewerIsGroupMember: false,
			Event: servicesevents.EventFullView{
				ID:           31,
				Title:        "Полное событие",
				Description:  "Описание",
				ImageURL:     "https://example.com/full.png",
				MaxUsers:     20,
				CurrentUsers: 6,
				StartTime:    start,
				EndTime:      end,
				Duration:     120,
				Group: servicesevents.EventReadGroupView{
					ID:         42,
					Name:       "Группа",
					Image:      "https://example.com/group.png",
					Enterprise: true,
				},
				EventTypeID:    3,
				LocationTypeID: 4,
				StatusID:       5,
				Genres:         []string{"Strategy", "Social"},
				Creator: servicesevents.EventReadCreatorView{
					ID:       9,
					Name:     "Автор",
					Us:       "author",
					Image:    "https://example.com/user.png",
					Verified: true,
				},
				Address:          "Улица 1",
				Country:          "RU",
				AgeLimit:         "18+",
				Year:             &year,
				Notes:            "Заметка",
				CustomFields:     map[string]interface{}{"format": "table"},
				ViewerSubscribed: true,
				ViewerIsCreator:  true,
				CreatedAt:        created,
				UpdatedAt:        updated,
			},
		},
	}
	service := servicesevents.NewEventReadService(&testLogger{}, store)
	ctx := context.WithValue(context.Background(), eventReadServiceContextKey{}, "details")

	result, err := service.GetEventDetails(ctx, 9, 31)

	if err != nil {
		t.Fatalf("GetEventDetails returned error: %v", err)
	}
	if store.detailsCalls != 1 ||
		store.detailsQuery.ViewerID != 9 ||
		store.detailsQuery.EventID != 31 ||
		store.detailContext.Value(eventReadServiceContextKey{}) != "details" {
		t.Fatalf("delegation = calls:%d query:%#v context:%v", store.detailsCalls, store.detailsQuery, store.detailContext)
	}
	if result == nil {
		t.Fatal("result = nil")
	}
	if result.ID != 31 ||
		result.Title != "Полное событие" ||
		result.Description != "Описание" ||
		result.ImageURL != "https://example.com/full.png" ||
		result.MaxUsers != 20 ||
		result.CurrentUsers != 6 ||
		!result.StartTime.Equal(start) ||
		!result.EndTime.Equal(end) ||
		result.Duration != 120 ||
		result.Group.ID != 42 ||
		result.Group.Name != "Группа" ||
		result.Group.Image != "https://example.com/group.png" ||
		!result.Group.Enterprise ||
		result.EventType != 3 ||
		result.LocationType != 4 ||
		result.Status != 5 ||
		!reflect.DeepEqual(result.Genres, []string{"Strategy", "Social"}) ||
		result.Creator.ID != 9 ||
		result.Creator.Name != "Автор" ||
		result.Creator.Us != "author" ||
		result.Creator.Image != "https://example.com/user.png" ||
		!result.Creator.Verified ||
		result.Address != "Улица 1" ||
		result.Country != "RU" ||
		result.AgeLimit != "18+" ||
		result.Year == nil ||
		*result.Year != 2028 ||
		result.Notes != "Заметка" ||
		result.CustomFields["format"] != "table" ||
		!result.Subscribed ||
		!result.IsCreator ||
		!result.CreatedAt.Equal(created) ||
		!result.UpdatedAt.Equal(updated) {
		t.Fatalf("mapped result = %#v, want all full view fields", result)
	}
}

func TestEventReadServiceGetEventDetailsAccessAndMissingEvent(t *testing.T) {
	tests := []struct {
		name    string
		view    servicesevents.EventDetailsView
		wantErr error
	}{
		{
			name: "публичное событие доступно постороннему",
			view: servicesevents.EventDetailsView{
				Found:               true,
				GroupPrivate:        false,
				ViewerIsGroupMember: false,
				Event:               servicesevents.EventFullView{ID: 1, AgeLimit: "12+"},
			},
		},
		{
			name: "приватное событие доступно участнику",
			view: servicesevents.EventDetailsView{
				Found:               true,
				GroupPrivate:        true,
				ViewerIsGroupMember: true,
				Event:               servicesevents.EventFullView{ID: 1, AgeLimit: "12+"},
			},
		},
		{
			name: "приватное событие запрещено постороннему",
			view: servicesevents.EventDetailsView{
				Found:               true,
				GroupPrivate:        true,
				ViewerIsGroupMember: false,
				Event:               servicesevents.EventFullView{ID: 1},
			},
			wantErr: servicesevents.ErrNotGroupMember,
		},
		{
			name:    "событие не найдено",
			view:    servicesevents.EventDetailsView{Found: false},
			wantErr: servicesevents.ErrEventNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &eventReadStoreStub{detailsView: tt.view}
			service := servicesevents.NewEventReadService(&testLogger{}, store)

			result, err := service.GetEventDetails(context.Background(), 9, 31)

			if tt.wantErr != nil {
				if result != nil {
					t.Fatalf("result = %#v, want nil", result)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetEventDetails returned error: %v", err)
			}
			if result == nil || result.ID != 1 || result.AgeLimit != "12+" {
				t.Fatalf("result = %#v, want mapped details", result)
			}
		})
	}
}

func TestEventReadServiceGetEventDetailsPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("ошибка чтения события")
	store := &eventReadStoreStub{detailsErr: storeErr}
	service := servicesevents.NewEventReadService(&testLogger{}, store)

	result, err := service.GetEventDetails(context.Background(), 9, 31)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want store error", err)
	}
}

var _ servicesevents.EventReadStore = (*eventReadStoreStub)(nil)
