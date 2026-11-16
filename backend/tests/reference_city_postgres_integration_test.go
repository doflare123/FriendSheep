package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"friendship/models"
	"friendship/services/references"
)

func TestGORMReferenceStoreSearchCitiesPostgresUsesCyrillicCaseInsensitiveLiteralSearch(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to verify Cyrillic city search in PostgreSQL")
	}

	db := newPostgresEventCommandDB(t, dsn)
	if err := db.AutoMigrate(&models.City{}); err != nil {
		t.Fatalf("auto migrate cities: %v", err)
	}
	for _, city := range []models.City{
		{Name: "Калининград", Region: "Калининградская обл", FIASID: "10000000-0000-0000-0000-000000000001"},
		{Name: "КАЛИНИН", Region: "Тверская обл", FIASID: "10000000-0000-0000-0000-000000000002"},
		{Name: "Калин!град", Region: "Тестовая обл", FIASID: "10000000-0000-0000-0000-000000000003"},
		{Name: "Калин_град", Region: "Тестовая обл", FIASID: "10000000-0000-0000-0000-000000000004"},
		{Name: "Калин%град", Region: "Тестовая обл", FIASID: "10000000-0000-0000-0000-000000000005"},
	} {
		if err := db.Create(&city).Error; err != nil {
			t.Fatalf("seed city %q: %v", city.Name, err)
		}
	}
	service := references.NewReferenceService(references.NewGORMReferenceStore(&testPostgresRepository{db: db}))

	caseFolded, err := service.SearchCities(context.Background(), references.CitySearchInput{Query: "кАлиН", Page: 1, Limit: 10})
	if err != nil || caseFolded.Total != 5 || len(caseFolded.Items) != 5 {
		t.Fatalf("Cyrillic case-insensitive search/error = %#v/%v, want all five cities", caseFolded, err)
	}
	for _, tt := range []struct {
		query string
		want  string
	}{
		{query: "!", want: "Калин!град"},
		{query: "_", want: "Калин_град"},
		{query: "%", want: "Калин%град"},
	} {
		t.Run(tt.query, func(t *testing.T) {
			result, searchErr := service.SearchCities(context.Background(), references.CitySearchInput{Query: tt.query, Page: 1, Limit: 10})
			if searchErr != nil || result.Total != 1 || len(result.Items) != 1 || result.Items[0].Name != tt.want {
				t.Fatalf("literal %q search/error = %#v/%v, want %q", tt.query, result, searchErr, tt.want)
			}
		})
	}
}
