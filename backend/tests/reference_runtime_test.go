package tests

import (
	"context"
	"testing"

	"friendship/models"
	eventmodels "friendship/models/events"
	"friendship/services/references"

	"gorm.io/gorm"
)

func TestGORMReferenceStoreSearchGenresIsCaseInsensitivePaginatedAndStable(t *testing.T) {
	db := newEventsServiceDB(t)
	seedReferenceGenres(t, db,
		eventmodels.Genre{ID: 30, Name: "board Tactics"},
		eventmodels.Genre{ID: 10, Name: "Adventure"},
		eventmodels.Genre{ID: 20, Name: "Board Games"},
		eventmodels.Genre{ID: 40, Name: "Strategy"},
	)
	service := references.NewReferenceService(references.NewGORMReferenceStore(&testPostgresRepository{db: db}))

	first, err := service.SearchGenres(context.Background(), references.GenreSearchInput{
		Query: "BOARD",
		Page:  1,
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("first SearchGenres returned error: %v", err)
	}
	if first.Total != 2 || first.Page != 1 || first.Limit != 1 || !first.HasMore {
		t.Fatalf("first page metadata = %#v, want total=2 page=1 limit=1 hasMore=true", first)
	}
	if len(first.Items) != 1 || first.Items[0].ID != 20 || first.Items[0].Name != "Board Games" {
		t.Fatalf("first page items = %#v, want stable first item Board Games", first.Items)
	}

	second, err := service.SearchGenres(context.Background(), references.GenreSearchInput{
		Query: "board",
		Page:  2,
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("second SearchGenres returned error: %v", err)
	}
	if second.Total != 2 || second.Page != 2 || second.Limit != 1 || second.HasMore {
		t.Fatalf("second page metadata = %#v, want total=2 page=2 limit=1 hasMore=false", second)
	}
	if len(second.Items) != 1 || second.Items[0].ID != 30 || second.Items[0].Name != "board Tactics" {
		t.Fatalf("second page items = %#v, want stable second item board Tactics", second.Items)
	}
}

func TestGORMReferenceStoreSearchGenresReturnsEmptyPageWithTotal(t *testing.T) {
	db := newEventsServiceDB(t)
	seedReferenceGenres(t, db,
		eventmodels.Genre{ID: 10, Name: "Adventure"},
		eventmodels.Genre{ID: 20, Name: "Board Games"},
	)
	service := references.NewReferenceService(references.NewGORMReferenceStore(&testPostgresRepository{db: db}))

	result, err := service.SearchGenres(context.Background(), references.GenreSearchInput{
		Query: "missing",
		Page:  1,
		Limit: 10,
	})

	if err != nil {
		t.Fatalf("SearchGenres returned error: %v", err)
	}
	if result == nil ||
		result.Items == nil ||
		len(result.Items) != 0 ||
		result.Total != 0 ||
		result.Page != 1 ||
		result.Limit != 10 ||
		result.HasMore {
		t.Fatalf("result = %#v, want non-nil empty page", result)
	}
}

func TestGORMReferenceStoreSearchCitiesPaginatesOrdersAndEscapesWildcards(t *testing.T) {
	db := newEventsServiceDB(t)
	for _, city := range []models.City{
		{ID: 40, Name: "Alfa_1", Region: "Север", FIASID: "00000000-0000-0000-0000-000000000040"},
		{ID: 50, Name: "Alfa_1", Region: "Север", FIASID: "00000000-0000-0000-0000-000000000050"},
		{ID: 30, Name: "Alfa_1", Region: "Юг", FIASID: "00000000-0000-0000-0000-000000000030"},
		{ID: 20, Name: "AlfaX1", Region: "Запад", FIASID: "00000000-0000-0000-0000-000000000020"},
		{ID: 10, Name: "100% Город", Region: "Восток", FIASID: "00000000-0000-0000-0000-000000000010"},
		{ID: 15, Name: "Bang! Город", Region: "Центр", FIASID: "00000000-0000-0000-0000-000000000015"},
	} {
		if err := db.Create(&city).Error; err != nil {
			t.Fatalf("seed city %q: %v", city.Name, err)
		}
	}
	service := references.NewReferenceService(references.NewGORMReferenceStore(&testPostgresRepository{db: db}))

	underscore, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "_", Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("SearchCities underscore returned error: %v", err)
	}
	if underscore.Total != 3 || len(underscore.Items) != 1 || !underscore.HasMore || underscore.Items[0].ID != 40 || underscore.Items[0].Region != "Север" {
		t.Fatalf("underscore first page = %#v, want first literal underscore match", underscore)
	}
	second, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "_", Page: 2, Limit: 1})
	if err != nil || second.Total != 3 || len(second.Items) != 1 || second.Items[0].ID != 50 || !second.HasMore {
		t.Fatalf("underscore second page/error = %#v/%v, want ID tiebreaker", second, err)
	}
	third, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "_", Page: 3, Limit: 1})
	if err != nil || third.Total != 3 || len(third.Items) != 1 || third.Items[0].ID != 30 || third.HasMore {
		t.Fatalf("underscore third page/error = %#v/%v, want final region-ordered match", third, err)
	}
	outOfRange, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "_", Page: 4, Limit: 1})
	if err != nil || outOfRange.Total != 3 || outOfRange.Items == nil || len(outOfRange.Items) != 0 || outOfRange.HasMore {
		t.Fatalf("underscore out-of-range page/error = %#v/%v, want empty page with total", outOfRange, err)
	}

	percent, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "%", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("SearchCities percent returned error: %v", err)
	}
	if percent.Total != 1 || len(percent.Items) != 1 || percent.Items[0].ID != 10 {
		t.Fatalf("percent page = %#v, want only literal percent match", percent)
	}

	escaped, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "!", Page: 1, Limit: 10})
	if err != nil || escaped.Total != 1 || len(escaped.Items) != 1 || escaped.Items[0].ID != 15 {
		t.Fatalf("escaped bang page/error = %#v/%v, want literal bang match", escaped, err)
	}

	empty, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "missing", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("SearchCities empty page returned error: %v", err)
	}
	if empty.Items == nil || len(empty.Items) != 0 || empty.Total != 0 || empty.HasMore {
		t.Fatalf("empty page = %#v, want non-nil empty page", empty)
	}
}

func TestGORMReferenceStoreSearchCitiesPreservesCanceledContext(t *testing.T) {
	db := newEventsServiceDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := references.NewReferenceService(references.NewGORMReferenceStore(&testPostgresRepository{db: db}))

	result, err := service.SearchCities(ctx, references.CitySearchInput{Page: 1, Limit: 10})
	if result != nil || err == nil {
		t.Fatalf("result = %#v, err = %v; want canceled-context error", result, err)
	}
}

func TestGORMReferenceStoreGetReferencesReturnsGeneralCollections(t *testing.T) {
	db := newEventsServiceDB(t)
	if err := db.Create(&models.Category{ID: 101, Name: "Games"}).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if err := db.Create(&eventmodels.EventLocation{ID: 102, Name: "Online"}).Error; err != nil {
		t.Fatalf("seed location: %v", err)
	}
	if err := db.Create(&eventmodels.AgeLimit{ID: 103, Name: "18+"}).Error; err != nil {
		t.Fatalf("seed age limit: %v", err)
	}
	if err := db.Create(&eventmodels.Status{ID: 104, Name: "Open"}).Error; err != nil {
		t.Fatalf("seed status: %v", err)
	}

	service := references.NewReferenceService(references.NewGORMReferenceStore(&testPostgresRepository{db: db}))
	result, err := service.GetReferences(context.Background())

	if err != nil {
		t.Fatalf("GetReferences returned error: %v", err)
	}
	if result == nil {
		t.Fatal("result = nil")
	}
	if len(result.EventTypes) != 1 || len(result.GroupCategories) != 1 {
		t.Fatalf("categories = event:%#v group:%#v, want both populated", result.EventTypes, result.GroupCategories)
	}
	if len(result.Locations) != 1 || len(result.AgeLimits) != 1 || len(result.Statuses) != 1 {
		t.Fatalf("references missing required collections: %#v", result)
	}
	if len(result.GroupActionTypes) == 0 {
		t.Fatal("group action types = empty, want seeded references")
	}
}

func seedReferenceGenres(t *testing.T, db *gorm.DB, genres ...eventmodels.Genre) {
	t.Helper()

	if err := db.Create(&genres).Error; err != nil {
		t.Fatalf("seed genres: %v", err)
	}
}
