package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	servicesevents "friendship/services/events"
)

func TestPopularEventsViewerAgeAndCacheIsolation(t *testing.T) {
	now := time.Date(2032, 6, 15, 0, 0, 0, 0, time.UTC)
	childBirthday := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
	cache := &popularEventsCacheStub{snapshot: &servicesevents.PopularEventsSnapshot{
		Events: []servicesevents.PopularEventView{{ID: 1, AgeLimit: "6+"}, {ID: 2, AgeLimit: "12+"}},
		Count:  2, UpdatedAt: now,
	}}
	store := &popularEventsStoreStub{viewer: servicesevents.EventViewerProfile{Birthday: &childBirthday}}
	service := servicesevents.NewPopularEventsService(&testLogger{}, store, cache, nil, nil, func() time.Time { return now })
	defer service.Stop()

	child, err := service.GetPopularEventsForViewer(context.Background(), 7)
	if err != nil || child.Count != 1 || len(child.Events) != 1 || child.Events[0].ID != 1 {
		t.Fatalf("child snapshot = %#v, err = %v", child, err)
	}
	anonymous, err := service.GetPopularEventsForViewer(context.Background(), 0)
	if err != nil || anonymous.Count != 2 || len(anonymous.Events) != 2 {
		t.Fatalf("anonymous snapshot = %#v, err = %v", anonymous, err)
	}
	if cache.snapshot.Count != 2 || len(cache.snapshot.Events) != 2 {
		t.Fatal("возрастной фильтр изменил общий кэш")
	}
	store.viewer = servicesevents.EventViewerProfile{}
	withoutBirthday, err := service.GetPopularEventsForViewer(context.Background(), 7)
	if err != nil || withoutBirthday.Count != 2 {
		t.Fatalf("null birthday snapshot = %#v, err = %v", withoutBirthday, err)
	}
	store.viewerErr = errors.New("viewer lookup failed")
	if _, err := service.GetPopularEventsForViewer(context.Background(), 7); err == nil {
		t.Fatal("ошибка чтения возраста должна прерывать выдачу")
	}
}
