package events

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"unicode/utf8"
)

const eventLinkMaxLength = 2048

type eventLocationMode uint8

const (
	eventLocationModeUnknown eventLocationMode = iota
	eventLocationModeOnline
	eventLocationModeOffline
)

type eventLocationDetails struct {
	CityID    *uint
	Address   string
	Link      string
	Latitude  *float64
	Longitude *float64
	Country   string
}

func resolveCreateEventLocation(locationStore eventCommandLocationStore, input CreateEventInput) (eventLocationDetails, error) {
	locationName, found, err := locationStore.FindEventLocationName(input.LocationID)
	if err != nil {
		return eventLocationDetails{}, err
	}
	if !found {
		return eventLocationDetails{}, ErrEventLocationNotFound
	}

	details := eventLocationDetails{
		CityID:    copyUintPointer(input.CityID),
		Address:   strings.TrimSpace(input.Address),
		Link:      strings.TrimSpace(input.Link),
		Latitude:  copyFloatPointer(input.Latitude),
		Longitude: copyFloatPointer(input.Longitude),
		Country:   strings.TrimSpace(input.Country),
	}
	if err := validateEventLocationDetails(locationStore, locationName, details); err != nil {
		return eventLocationDetails{}, err
	}
	return details, nil
}

func resolveUpdatedEventLocation(locationStore eventCommandLocationStore, event EventCommandSnapshot, input UpdateEventInput) (eventLocationDetails, error) {
	locationID := event.LocationID
	if input.LocationID != nil {
		locationID = *input.LocationID
	}
	locationName, found, err := locationStore.FindEventLocationName(locationID)
	if err != nil {
		return eventLocationDetails{}, err
	}
	if !found {
		return eventLocationDetails{}, ErrEventLocationNotFound
	}

	currentMode := parseEventLocationMode(event.LocationName)
	targetMode := parseEventLocationMode(locationName)
	details := eventLocationDetails{
		CityID:    copyUintPointer(event.CityID),
		Address:   event.Address,
		Link:      event.Link,
		Latitude:  copyFloatPointer(event.Latitude),
		Longitude: copyFloatPointer(event.Longitude),
		Country:   event.Country,
	}
	if currentMode != targetMode {
		switch targetMode {
		case eventLocationModeOnline:
			details.CityID = nil
			details.Address = ""
			details.Latitude = nil
			details.Longitude = nil
			details.Country = ""
		case eventLocationModeOffline:
			details.Link = ""
		}
	}

	if input.CityID != nil || input.CityIDSet {
		details.CityID = copyUintPointer(input.CityID)
	}
	if input.Address != nil {
		details.Address = strings.TrimSpace(*input.Address)
	}
	if input.Link != nil {
		details.Link = strings.TrimSpace(*input.Link)
	}
	if input.Latitude != nil || input.LatitudeSet {
		details.Latitude = copyFloatPointer(input.Latitude)
	}
	if input.Longitude != nil || input.LongitudeSet {
		details.Longitude = copyFloatPointer(input.Longitude)
	}
	if input.Country != nil {
		details.Country = strings.TrimSpace(*input.Country)
	}

	if err := validateEventLocationDetails(locationStore, locationName, details); err != nil {
		return eventLocationDetails{}, err
	}
	return details, nil
}

func validateEventLocationDetails(store eventCommandLocationStore, locationName string, details eventLocationDetails) error {
	switch parseEventLocationMode(locationName) {
	case eventLocationModeOnline:
		if err := validateOnlineEventLink(details.Link); err != nil {
			return err
		}
		if details.CityID != nil || details.Address != "" || details.Latitude != nil || details.Longitude != nil || details.Country != "" {
			return fmt.Errorf("%w: для онлайн-мероприятия нельзя задавать cityId, address, country или координаты", ErrInvalidEventUpdate)
		}
	case eventLocationModeOffline:
		if details.Link != "" {
			return fmt.Errorf("%w: для офлайн-мероприятия нельзя задавать link", ErrInvalidEventUpdate)
		}
		if details.CityID == nil || *details.CityID == 0 {
			return fmt.Errorf("%w: для офлайн-мероприятия поле cityId обязательно", ErrInvalidEventUpdate)
		}
		if utf8.RuneCountInString(details.Address) == 0 || utf8.RuneCountInString(details.Address) > eventAddressMaxLength {
			return fmt.Errorf("%w: для офлайн-мероприятия поле address должно содержать от 1 до %d символов", ErrInvalidEventUpdate, eventAddressMaxLength)
		}
		if err := validateEventCoordinates(details.Latitude, details.Longitude); err != nil {
			return err
		}
		exists, err := store.CityExists(*details.CityID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%w: город не найден", ErrInvalidEventUpdate)
		}
	default:
		return fmt.Errorf("%w: неизвестный формат проведения %q", ErrInvalidEventUpdate, locationName)
	}
	return nil
}

func validateOnlineEventLink(value string) error {
	if value == "" || utf8.RuneCountInString(value) > eventLinkMaxLength {
		return fmt.Errorf("%w: для онлайн-мероприятия поле link обязательно и не должно превышать %d символов", ErrInvalidEventUpdate, eventLinkMaxLength)
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%w: поле link должно содержать абсолютный HTTP(S) URL", ErrInvalidEventUpdate)
	}
	return nil
}

func validateEventCoordinates(latitude, longitude *float64) error {
	if latitude == nil && longitude == nil {
		return nil
	}
	if latitude == nil || longitude == nil {
		return fmt.Errorf("%w: latitude и longitude должны быть заданы парой", ErrInvalidEventUpdate)
	}
	if math.IsNaN(*latitude) || math.IsInf(*latitude, 0) || *latitude < -90 || *latitude > 90 {
		return fmt.Errorf("%w: latitude должна находиться в диапазоне от -90 до 90", ErrInvalidEventUpdate)
	}
	if math.IsNaN(*longitude) || math.IsInf(*longitude, 0) || *longitude < -180 || *longitude > 180 {
		return fmt.Errorf("%w: longitude должна находиться в диапазоне от -180 до 180", ErrInvalidEventUpdate)
	}
	return nil
}

func parseEventLocationMode(value string) eventLocationMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "online", "онлайн":
		return eventLocationModeOnline
	case "offline", "off-line", "офлайн", "оффлайн":
		return eventLocationModeOffline
	default:
		return eventLocationModeUnknown
	}
}

func copyUintPointer(value *uint) *uint {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func copyFloatPointer(value *float64) *float64 {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func stringPointer(value string) *string {
	result := value
	return &result
}
