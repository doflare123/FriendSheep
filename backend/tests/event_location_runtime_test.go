package tests

import (
	"context"
	"testing"

	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
)

func TestEventCommandStructuredLocationRoundTripAndModeSwitch(t *testing.T) {
	db := newEventsServiceDB(t)
	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembershipWithRole(t, db, 1, groupID, groupmodels.RoleAdmin)
	eventID := seedEvent(t, db, groupID, 1, 1, 10)

	city := models.City{Name: "Калининград", Region: "Калининградская обл", FIASID: "df679694-d505-4dd3-b514-4ba48c8a97d8"}
	if err := db.Create(&city).Error; err != nil {
		t.Fatalf("create city: %v", err)
	}
	latitude := 54.7104
	longitude := 20.4522
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"city_id":   city.ID,
		"address":   "ул. Театральная, 1",
		"latitude":  latitude,
		"longitude": longitude,
	}).Error; err != nil {
		t.Fatalf("configure offline location: %v", err)
	}

	onlineID := seedEventLocation(t, db, "Online")
	service := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)
	link := "https://example.com/live/event"
	online, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{
		LocationID: &onlineID,
		Link:       &link,
	})
	if err != nil {
		t.Fatalf("switch to online: %v", err)
	}
	if online.Link != link || online.CityID != nil || online.City != "" || online.Address != "" || online.Latitude != nil || online.Longitude != nil {
		t.Fatalf("online result = %#v", online)
	}

	offlineID := uint(1)
	address := "пл. Победы, 1"
	offline, err := service.UpdateEvent(context.Background(), 1, eventID, servicesevents.UpdateEventInput{
		LocationID:   &offlineID,
		CityID:       &city.ID,
		Address:      &address,
		LatitudeSet:  true,
		LongitudeSet: true,
	})
	if err != nil {
		t.Fatalf("switch back to offline: %v", err)
	}
	if offline.Link != "" || offline.CityID == nil || *offline.CityID != city.ID || offline.City != city.Name || offline.Address != address || offline.Latitude != nil || offline.Longitude != nil {
		t.Fatalf("offline result = %#v", offline)
	}

	var stored eventmodels.Event
	if err := db.First(&stored, eventID).Error; err != nil {
		t.Fatalf("load stored event: %v", err)
	}
	if stored.Link != "" || stored.CityID == nil || *stored.CityID != city.ID || stored.Address != address || stored.Latitude != nil || stored.Longitude != nil {
		t.Fatalf("stored event = %#v", stored)
	}
}
