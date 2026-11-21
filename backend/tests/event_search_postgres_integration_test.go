package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"friendship/models"
	eventmodels "friendship/models/events"
	servicesevents "friendship/services/events"
)

func TestEventDiscoveryAgeCountAndCityOrderPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
	}
	db := newPostgresEventCommandDB(t, dsn)
	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventReferences(t, db)
	if err := db.Model(&eventmodels.AgeLimit{}).Where("id = ?", 1).Update("name", "12+").Error; err != nil {
		t.Fatal(err)
	}
	adult := eventmodels.AgeLimit{Name: "18+"}
	if err := db.Create(&adult).Error; err != nil {
		t.Fatal(err)
	}
	lat, lon, nearLon := 54.7, 20.5, 21.0
	city := models.City{Name: "Свой", Region: "Регион", FIASID: "own", Latitude: &lat, Longitude: &lon}
	nearby := models.City{Name: "Рядом", Region: "Регион", FIASID: "near", Latitude: &lat, Longitude: &nearLon}
	if err := db.Create(&city).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&nearby).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.User{}).Where("id = ?", 2).Update("city_id", city.ID).Error; err != nil {
		t.Fatal(err)
	}
	ownID := seedEvent(t, db, groupID, 1, 1, 20)
	nearID := seedEvent(t, db, groupID, 1, 19, 20)
	adultID := seedEvent(t, db, groupID, 1, 20, 20)
	for id, cityID := range map[uint]uint{ownID: city.ID, nearID: nearby.ID, adultID: city.ID} {
		if err := db.Model(&eventmodels.Event{}).Where("id = ?", id).Update("city_id", cityID).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", adultID).Update("age_limit_id", adult.ID).Error; err != nil {
		t.Fatal(err)
	}
	service := servicesevents.NewEventReadService(&testLogger{}, servicesevents.NewGORMEventReadStore(&testPostgresRepository{db: db}))
	result, err := service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{Page: 1, Limit: 1})
	if err != nil || result.Total != 2 || len(result.Items) != 1 || result.Items[0].ID != ownID {
		t.Fatalf("preferred city page = %#v, err = %v", result, err)
	}
	result, err = service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{CityID: &nearby.ID, Page: 1, Limit: 1})
	if err != nil || result.Total != 1 || len(result.Items) != 1 || result.Items[0].ID != nearID {
		t.Fatalf("explicit city page = %#v, err = %v", result, err)
	}
	result, err = service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{City: "РЯДОМ", Page: 1, Limit: 1})
	if err != nil || result.Total != 1 || len(result.Items) != 1 || result.Items[0].ID != nearID {
		t.Fatalf("case insensitive city page = %#v, err = %v", result, err)
	}
	childBirthday := time.Now().UTC().AddDate(-10, 0, 0)
	if err := db.Model(&models.User{}).Where("id = ?", 2).Update("birthday", childBirthday).Error; err != nil {
		t.Fatal(err)
	}
	result, err = service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{Page: 1, Limit: 1})
	if err != nil || result.Total != 0 || len(result.Items) != 0 {
		t.Fatalf("child page = %#v, err = %v", result, err)
	}
}

func TestEventDiscoveryNumericCityRankPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
	}
	db := newPostgresEventCommandDB(t, dsn)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	var want []uint
	var originID uint
	for i := 0; i < 12; i++ {
		lat, lon := 50.0, float64(i)
		city := models.City{Name: fmt.Sprintf("Город %d", i), Region: "Регион", FIASID: fmt.Sprintf("rank-%d", i), Latitude: &lat, Longitude: &lon}
		if err := db.Create(&city).Error; err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			originID = city.ID
		}
		id := seedEvent(t, db, groupID, 1, uint16(12-i), 20)
		if err := db.Model(&eventmodels.Event{}).Where("id = ?", id).Update("city_id", city.ID).Error; err != nil {
			t.Fatal(err)
		}
		want = append(want, id)
	}
	if err := db.Model(&models.User{}).Where("id = ?", 1).Update("city_id", originID).Error; err != nil {
		t.Fatal(err)
	}
	service := servicesevents.NewEventReadService(&testLogger{}, servicesevents.NewGORMEventReadStore(&testPostgresRepository{db: db}))
	result, err := service.SearchEvents(context.Background(), 1, servicesevents.EventSearchInput{Page: 1, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	assertDiscoveryIDs(t, result, want)
}
