package tests

import (
	"context"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	servicesevents "friendship/services/events"

	"gorm.io/gorm"
)

func TestGORMPopularEventsStoreFiltersRanksLimitsAndLoadsOwnerMetadata(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	store := servicesevents.NewGORMPopularEventsStore(repo)
	now := time.Now().UTC()

	seedEventUser(t, db, 701)
	seedEventUser(t, db, 702)
	publicGroupID := seedEventGroup(t, db, 701, false)
	privateGroupID := seedEventGroup(t, db, 702, true)
	seedEventReferences(t, db)
	if err := db.Exec(
		"UPDATE groups SET name = ? WHERE id = ?",
		"Public Popular Group",
		publicGroupID,
	).Error; err != nil {
		t.Fatalf("set public group name: %v", err)
	}

	highestID := seedNamedPopularEvent(t, db, publicGroupID, 701, "Highest occupancy", 9, 10)
	tiedHigherUsersID := seedNamedPopularEvent(t, db, publicGroupID, 701, "Tied with more users", 8, 10)
	tiedLowerUsersID := seedNamedPopularEvent(t, db, publicGroupID, 701, "Tied with fewer users", 4, 5)
	seedNamedPopularEvent(t, db, publicGroupID, 701, "Below limit", 2, 10)

	seedNamedPopularEvent(t, db, privateGroupID, 702, "Private", 10, 10)
	pastID := seedNamedPopularEvent(t, db, publicGroupID, 701, "Past", 10, 10)
	if err := db.Model(&eventmodels.Event{}).
		Where("id = ?", pastID).
		Updates(map[string]interface{}{
			"start_time": now.Add(-2 * time.Hour),
			"end_time":   now.Add(-time.Hour),
		}).Error; err != nil {
		t.Fatalf("move event into past: %v", err)
	}
	seedNamedPopularEvent(t, db, publicGroupID, 701, "Only creator", 1, 10)
	seedNamedPopularEvent(t, db, publicGroupID, 701, "No capacity", 2, 0)

	finishedStatus := eventmodels.Status{Name: "Завершена"}
	if err := db.Create(&finishedStatus).Error; err != nil {
		t.Fatalf("create finished status: %v", err)
	}
	finishedID := seedNamedPopularEvent(t, db, publicGroupID, 701, "Finished", 10, 10)
	if err := db.Model(&eventmodels.Event{}).
		Where("id = ?", finishedID).
		Update("status_id", finishedStatus.ID).Error; err != nil {
		t.Fatalf("finish event: %v", err)
	}
	if err := db.Model(&eventmodels.Status{}).
		Where("id = ?", 1).
		Update("name", "Набор").Error; err != nil {
		t.Fatalf("set recruitment status: %v", err)
	}

	records, err := store.ListTopPopularEvents(context.Background(), now, 3)

	if err != nil {
		t.Fatalf("ListTopPopularEvents returned error: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("records = %#v, want three limited popular events", records)
	}
	wantIDs := []uint{highestID, tiedHigherUsersID, tiedLowerUsersID}
	for index, wantID := range wantIDs {
		if records[index].ID != wantID {
			t.Fatalf("record %d ID = %d, want %d; records = %#v", index, records[index].ID, wantID, records)
		}
		if records[index].GroupName != "Public Popular Group" ||
			records[index].OwnerEmail != "event-user-701@example.com" ||
			records[index].OwnerUserID != 701 {
			t.Fatalf("record %d owner metadata = group:%q email:%q user:%d", index, records[index].GroupName, records[index].OwnerEmail, records[index].OwnerUserID)
		}
		if records[index].Group.ID != publicGroupID || records[index].Group.Name != "Public Popular Group" ||
			records[index].Group.Image != "https://example.com/group.png" ||
			records[index].EventType != "Game" || records[index].LocationType != "Offline" ||
			records[index].AgeLimit != "18+" || records[index].Status != "Набор" ||
			records[index].StartTime.IsZero() || records[index].Subscribed {
			t.Fatalf("record %d shared search fields = %#v", index, records[index].PopularEventView)
		}
	}
	if records[0].PopularityRate != 0.9 ||
		records[1].PopularityRate != 0.8 ||
		records[2].PopularityRate != 0.8 {
		t.Fatalf("popularity rates = %v, %v, %v", records[0].PopularityRate, records[1].PopularityRate, records[2].PopularityRate)
	}
}

func seedNamedPopularEvent(
	t *testing.T,
	db *gorm.DB,
	groupID uint,
	creatorID uint,
	title string,
	currentUsers uint16,
	maxUsers uint16,
) uint {
	t.Helper()

	eventID := seedEvent(t, db, groupID, creatorID, currentUsers, maxUsers)
	if err := db.Model(&eventmodels.Event{}).
		Where("id = ?", eventID).
		Update("title", title).Error; err != nil {
		t.Fatalf("name popular event %q: %v", title, err)
	}
	return eventID
}
