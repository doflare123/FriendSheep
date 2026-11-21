package tests

import (
	"context"
	"reflect"
	"testing"
	"time"

	"friendship/models"
	"friendship/models/dto"
	eventmodels "friendship/models/events"
	servicesevents "friendship/services/events"
	servicegroups "friendship/services/groups"
	"gorm.io/gorm"
)

func TestEventSearchAppliesAgeLimitBeforeCountAndPagination(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	now := time.Date(2032, 3, 1, 8, 0, 0, 0, time.UTC)
	service := servicesevents.NewEventReadServiceWithClock(&testLogger{}, servicesevents.NewGORMEventReadStore(repo), eventReadClockStub{now: now})

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedDiscoveryAgeLimits(t, db)
	eventIDs := seedDiscoveryEvents(t, db, groupID, 1, []uint{10, 20, 1, 40, 50, 60})
	setDiscoveryBirthday(t, db, 2, time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name      string
		userID    uint
		birthday  *time.Time
		wantTotal int64
		wantIDs   []uint
	}{
		{name: "anonymous and no birthday stop at 12", wantTotal: 3, wantIDs: eventIDs[:3]},
		{name: "under twelve sees only 0 and 6", userID: 2, wantTotal: 2, wantIDs: eventIDs[:2]},
		{name: "on twelfth birthday includes 12", userID: 2, birthday: discoveryTimePointer(time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)), wantTotal: 3, wantIDs: eventIDs[:3]},
		{name: "on sixteenth birthday includes 16", userID: 2, birthday: discoveryTimePointer(time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC)), wantTotal: 4, wantIDs: eventIDs[:4]},
		{name: "on eighteenth birthday includes 18", userID: 2, birthday: discoveryTimePointer(time.Date(2014, 3, 1, 0, 0, 0, 0, time.UTC)), wantTotal: 5, wantIDs: eventIDs[:5]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.birthday != nil {
				setDiscoveryBirthday(t, db, 2, *tt.birthday)
			}
			result, err := service.SearchEvents(context.Background(), tt.userID, servicesevents.EventSearchInput{Page: 1, Limit: 1})
			if err != nil {
				t.Fatalf("SearchEvents returned error: %v", err)
			}
			if result.Total != tt.wantTotal || len(result.Items) != 1 || !containsDiscoveryID(tt.wantIDs, result.Items[0].ID) {
				t.Fatalf("search result = %#v, want total %d and page item from %#v", result, tt.wantTotal, tt.wantIDs)
			}
			all, err := service.SearchEvents(context.Background(), tt.userID, servicesevents.EventSearchInput{Page: 1, Limit: 100})
			if err != nil || int64(len(all.Items)) != tt.wantTotal {
				t.Fatalf("full result = %#v, err = %v", all, err)
			}
			for _, item := range all.Items {
				if !containsDiscoveryID(tt.wantIDs, item.ID) {
					t.Fatalf("выдача содержит недопустимое мероприятие: %#v", item)
				}
			}
		})
	}
}

func TestEventDiscoveryGroupListsApplyAgePolicy(t *testing.T) {
	db := newGroupServiceDB(t)
	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedDiscoveryAgeLimits(t, db)
	ids := seedDiscoveryEvents(t, db, groupID, 1, []uint{20, 1, 50})
	setDiscoveryBirthday(t, db, 2, time.Now().UTC().AddDate(-10, 0, 0))
	repo := &testPostgresRepository{db: db}
	reads := servicesevents.NewEventReadService(&testLogger{}, servicesevents.NewGORMEventReadStore(repo))
	items, err := reads.GetGroupEvents(context.Background(), 2, groupID)
	if err != nil || len(items) != 1 || items[0].ID != ids[0] {
		t.Fatalf("список группы = %#v, err = %v", items, err)
	}
	groups := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))
	group, err := groups.GetGroupDetails(context.Background(), 2, groupID)
	if err != nil || group == nil || len(group.ActiveEvents) != 1 || group.ActiveEvents[0].ID != ids[0] {
		t.Fatalf("карточки внутри группы = %#v, err = %v", group, err)
	}
}

func TestEventSearchUsesEventCityForFilterAndDiscoveryOrder(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	now := time.Date(2032, 3, 1, 8, 0, 0, 0, time.UTC)
	service := servicesevents.NewEventReadServiceWithClock(&testLogger{}, servicesevents.NewGORMEventReadStore(repo), eventReadClockStub{now: now})

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedDiscoveryAgeLimits(t, db)
	setDiscoveryBirthday(t, db, 2, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
	ownCityID := seedDiscoveryCity(t, db, "Свой", 54.7, 20.5)
	nearCityID := seedDiscoveryCity(t, db, "Nearby", 55.0, 21.0)
	farCityID := seedDiscoveryCity(t, db, "Далеко", 55.8, 37.6)
	setDiscoveryUserCity(t, db, 2, ownCityID)

	eventIDs := seedDiscoveryEvents(t, db, groupID, 1, []uint{1, 1, 1, 1})
	onlineLocationID := seedEventLocation(t, db, "Online")
	setDiscoveryEvent(t, db, eventIDs[0], ownCityID, 1, 2, 20)
	setDiscoveryEvent(t, db, eventIDs[1], nearCityID, 1, 99, 100)
	setDiscoveryEvent(t, db, eventIDs[2], farCityID, 1, 80, 100)
	setDiscoveryEvent(t, db, eventIDs[3], 0, onlineLocationID, 100, 100)

	result, err := service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("SearchEvents returned error: %v", err)
	}
	assertDiscoveryIDs(t, result, []uint{eventIDs[0], eventIDs[1], eventIDs[2], eventIDs[3]})

	explicitCity := nearCityID
	filtered, err := service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{CityID: &explicitCity, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("SearchEvents with cityId returned error: %v", err)
	}
	assertDiscoveryIDs(t, filtered, []uint{eventIDs[1]})

	byName, err := service.SearchEvents(context.Background(), 2, servicesevents.EventSearchInput{City: "NEARBY", Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("SearchEvents with city name returned error: %v", err)
	}
	assertDiscoveryIDs(t, byName, []uint{eventIDs[1]})
}

func TestEventSearchWithoutPreferredCityOrdersOfflineByPopularity(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicesevents.NewEventReadServiceWithClock(&testLogger{}, servicesevents.NewGORMEventReadStore(repo), eventReadClockStub{now: time.Date(2032, 3, 1, 0, 0, 0, 0, time.UTC)})

	seedEventUser(t, db, 1)
	groupID := seedEventGroup(t, db, 1, false)
	seedDiscoveryAgeLimits(t, db)
	eventIDs := seedDiscoveryEvents(t, db, groupID, 1, []uint{1, 1, 1})
	setDiscoveryEvent(t, db, eventIDs[0], 0, 1, 4, 5)
	setDiscoveryEvent(t, db, eventIDs[1], 0, 1, 4, 20)
	setDiscoveryEvent(t, db, eventIDs[2], 0, 1, 7, 8)

	result, err := service.SearchEvents(context.Background(), 0, servicesevents.EventSearchInput{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("SearchEvents returned error: %v", err)
	}
	assertDiscoveryIDs(t, result, []uint{eventIDs[2], eventIDs[1], eventIDs[0]})
}

func seedDiscoveryAgeLimits(t *testing.T, db *gorm.DB) {
	t.Helper()
	seedEventReferences(t, db)
	limits := []eventmodels.AgeLimit{
		{ID: 10, Name: "0+"},
		{ID: 20, Name: "6+"},
		{ID: 1, Name: "12+"},
		{ID: 40, Name: "16+"},
		{ID: 50, Name: "18+"},
		{ID: 60, Name: "21+"},
	}
	for _, limit := range limits {
		if err := db.FirstOrCreate(&limit, eventmodels.AgeLimit{ID: limit.ID}).Error; err != nil {
			t.Fatalf("create age limit %q: %v", limit.Name, err)
		}
	}
}

func seedDiscoveryEvents(t *testing.T, db *gorm.DB, groupID, creatorID uint, ageLimitIDs []uint) []uint {
	t.Helper()
	result := make([]uint, 0, len(ageLimitIDs))
	for _, ageLimitID := range ageLimitIDs {
		eventID := seedEvent(t, db, groupID, creatorID, 1, 10)
		if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Update("age_limit_id", ageLimitID).Error; err != nil {
			t.Fatalf("update event %d age limit: %v", eventID, err)
		}
		result = append(result, eventID)
	}
	return result
}

func setDiscoveryBirthday(t *testing.T, db *gorm.DB, userID uint, birthday time.Time) {
	t.Helper()
	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("birthday", birthday).Error; err != nil {
		t.Fatalf("set user birthday: %v", err)
	}
}

func seedDiscoveryCity(t *testing.T, db *gorm.DB, name string, latitude, longitude float64) uint {
	t.Helper()
	city := models.City{Name: name, Region: "Регион", FIASID: "discovery-" + name, Latitude: &latitude, Longitude: &longitude}
	if err := db.Create(&city).Error; err != nil {
		t.Fatalf("create city %q: %v", name, err)
	}
	return city.ID
}

func setDiscoveryUserCity(t *testing.T, db *gorm.DB, userID, cityID uint) {
	t.Helper()
	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("city_id", cityID).Error; err != nil {
		t.Fatalf("set user city: %v", err)
	}
}

func setDiscoveryEvent(t *testing.T, db *gorm.DB, eventID, cityID, locationID uint, currentUsers, maxUsers uint16) {
	t.Helper()
	updates := map[string]interface{}{
		"event_location_id": locationID,
		"current_users":     currentUsers,
		"max_users":         maxUsers,
		"city_id":           nil,
	}
	if cityID > 0 {
		updates["city_id"] = cityID
	}
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", eventID).Updates(updates).Error; err != nil {
		t.Fatalf("update event %d discovery fields: %v", eventID, err)
	}
}

func assertDiscoveryIDs(t *testing.T, result *dto.EventSearchResponse, want []uint) {
	t.Helper()
	if result == nil || len(result.Items) != len(want) {
		t.Fatalf("result = %#v, want %d items", result, len(want))
	}
	actual := make([]uint, 0, len(result.Items))
	for _, item := range result.Items {
		actual = append(actual, item.ID)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("event IDs = %#v, want %#v", actual, want)
	}
}

func containsDiscoveryID(ids []uint, target uint) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func discoveryTimePointer(value time.Time) *time.Time { return &value }
