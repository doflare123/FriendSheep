package events

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
)

func TestUpdateEventInputTracksExplicitLocationNulls(t *testing.T) {
	var input UpdateEventInput
	if err := json.Unmarshal([]byte(`{"latitude":null,"longitude":null,"cityId":null}`), &input); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if !input.LatitudeSet || !input.LongitudeSet || !input.CityIDSet {
		t.Fatalf("presence flags = city:%v latitude:%v longitude:%v", input.CityIDSet, input.LatitudeSet, input.LongitudeSet)
	}
	if input.Latitude != nil || input.Longitude != nil || input.CityID != nil {
		t.Fatalf("explicit null values were not preserved: %#v", input)
	}
}

type eventLocationStoreStub struct {
	locationNames map[uint]string
	cityExists    bool
	cityErr       error
}

func (s eventLocationStoreStub) FindEventLocationName(locationID uint) (string, bool, error) {
	name, found := s.locationNames[locationID]
	return name, found, nil
}

func TestResolveUpdatedEventLocationCleansFieldsWhenModeChanges(t *testing.T) {
	cityID := uint(42)
	latitude := 54.7104
	longitude := 20.4522
	link := "https://example.com/live"
	onlineID := uint(2)
	store := eventLocationStoreStub{
		locationNames: map[uint]string{2: "Онлайн"},
		cityExists:    true,
	}

	details, err := resolveUpdatedEventLocation(store, EventCommandSnapshot{
		LocationID:   1,
		LocationName: "Оффлайн",
		CityID:       &cityID,
		Address:      "ул. Театральная, 1",
		Latitude:     &latitude,
		Longitude:    &longitude,
		Country:      "RU",
	}, UpdateEventInput{LocationID: &onlineID, Link: &link})
	if err != nil {
		t.Fatalf("resolveUpdatedEventLocation returned error: %v", err)
	}
	if details.Link != link || details.CityID != nil || details.Address != "" || details.Latitude != nil || details.Longitude != nil || details.Country != "" {
		t.Fatalf("online details after switch = %#v", details)
	}
}

func TestResolveUpdatedEventLocationAllowsExplicitCoordinateReset(t *testing.T) {
	cityID := uint(42)
	latitude := 54.7104
	longitude := 20.4522
	store := eventLocationStoreStub{
		locationNames: map[uint]string{1: "Оффлайн"},
		cityExists:    true,
	}

	details, err := resolveUpdatedEventLocation(store, EventCommandSnapshot{
		LocationID:   1,
		LocationName: "Оффлайн",
		CityID:       &cityID,
		Address:      "ул. Театральная, 1",
		Latitude:     &latitude,
		Longitude:    &longitude,
	}, UpdateEventInput{LatitudeSet: true, LongitudeSet: true})
	if err != nil {
		t.Fatalf("resolveUpdatedEventLocation returned error: %v", err)
	}
	if details.Latitude != nil || details.Longitude != nil {
		t.Fatalf("coordinates after explicit reset = %v/%v", details.Latitude, details.Longitude)
	}
}

func (s eventLocationStoreStub) CityExists(uint) (bool, error) {
	return s.cityExists, s.cityErr
}

func TestValidateEventLocationDetailsAcceptsStructuredModes(t *testing.T) {
	cityID := uint(42)
	latitude := 54.7104
	longitude := 20.4522
	store := eventLocationStoreStub{cityExists: true}

	tests := []struct {
		name     string
		location string
		details  eventLocationDetails
	}{
		{
			name:     "онлайн",
			location: "Онлайн",
			details:  eventLocationDetails{Link: "https://example.com/events/42"},
		},
		{
			name:     "офлайн с точкой",
			location: "Оффлайн",
			details: eventLocationDetails{
				CityID:    &cityID,
				Address:   "ул. Театральная, 1",
				Latitude:  &latitude,
				Longitude: &longitude,
			},
		},
		{
			name:     "офлайн без точки",
			location: "offline",
			details:  eventLocationDetails{CityID: &cityID, Address: "Главная площадь"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateEventLocationDetails(store, test.location, test.details); err != nil {
				t.Fatalf("validateEventLocationDetails returned error: %v", err)
			}
		})
	}
}

func TestValidateEventLocationDetailsRejectsInconsistentDataBeforePersistence(t *testing.T) {
	cityID := uint(42)
	nan := math.NaN()
	storeErr := errors.New("city lookup failed")

	tests := []struct {
		name     string
		store    eventLocationStoreStub
		location string
		details  eventLocationDetails
		wantErr  error
	}{
		{
			name:     "онлайн без ссылки",
			store:    eventLocationStoreStub{},
			location: "online",
			wantErr:  ErrInvalidEventUpdate,
		},
		{
			name:     "онлайн с городом",
			store:    eventLocationStoreStub{},
			location: "online",
			details:  eventLocationDetails{Link: "https://example.com", CityID: &cityID},
			wantErr:  ErrInvalidEventUpdate,
		},
		{
			name:     "офлайн без адреса",
			store:    eventLocationStoreStub{cityExists: true},
			location: "offline",
			details:  eventLocationDetails{CityID: &cityID},
			wantErr:  ErrInvalidEventUpdate,
		},
		{
			name:     "координаты не парой",
			store:    eventLocationStoreStub{cityExists: true},
			location: "offline",
			details:  eventLocationDetails{CityID: &cityID, Address: "Адрес", Latitude: &nan},
			wantErr:  ErrInvalidEventUpdate,
		},
		{
			name:     "ошибка проверки города",
			store:    eventLocationStoreStub{cityErr: storeErr},
			location: "offline",
			details:  eventLocationDetails{CityID: &cityID, Address: "Адрес"},
			wantErr:  storeErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateEventLocationDetails(test.store, test.location, test.details)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}
		})
	}
}
