package events

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	eventTitleMinLength       = 5
	eventTitleMaxLength       = 200
	eventDescriptionMinLength = 10
	eventDescriptionMaxLength = 2000
	eventAddressMaxLength     = 500
	eventCountryMaxLength     = 100
	eventDurationMinMinutes   = 15
	eventDurationMaxMinutes   = 1440
	eventMinUsers             = 2
	eventMaxUsers             = 1000
)

func validateUpdateEventInput(input UpdateEventInput, now time.Time) error {
	if !hasUpdateEventChanges(input) {
		return fmt.Errorf("%w: необходимо передать хотя бы одно поле для обновления", ErrInvalidEventUpdate)
	}
	if err := validateUpdatedEventTitle(input.Title); err != nil {
		return err
	}
	if err := validateUpdatedEventDescription(input.Description); err != nil {
		return err
	}
	if err := validateUpdatedPositiveUint("eventTypeId", input.EventTypeID); err != nil {
		return err
	}
	if err := validateUpdatedPositiveUint("locationId", input.LocationID); err != nil {
		return err
	}
	if err := validateUpdatedEventImageURL(input.ImageURL); err != nil {
		return err
	}
	if err := validateUpdatedEventStartTime(input.StartTime, now); err != nil {
		return err
	}
	if err := validateUpdatedEventDuration(input.Duration); err != nil {
		return err
	}
	if err := validateUpdatedEventMaxUsers(input.MaxUsers); err != nil {
		return err
	}
	if err := validateUpdatedPositiveUint("ageLimit", input.AgeLimit); err != nil {
		return err
	}
	if err := validateUpdatedMaxLength("address", input.Address, eventAddressMaxLength); err != nil {
		return err
	}
	if err := validateUpdatedMaxLength("country", input.Country, eventCountryMaxLength); err != nil {
		return err
	}
	if err := validateUpdatedGenres(input.Genres); err != nil {
		return err
	}

	return nil
}

func hasUpdateEventChanges(input UpdateEventInput) bool {
	return input.Title != nil ||
		input.Description != nil ||
		input.EventTypeID != nil ||
		input.LocationID != nil ||
		input.ImageURL != nil ||
		input.StartTime != nil ||
		input.Duration != nil ||
		input.MaxUsers != nil ||
		input.Genres != nil ||
		input.Address != nil ||
		input.Country != nil ||
		input.AgeLimit != nil ||
		input.Year != nil ||
		input.Notes != nil ||
		input.CustomFields != nil
}

func validateUpdatedEventTitle(value *string) error {
	if value == nil {
		return nil
	}

	length := utf8.RuneCountInString(*value)
	if length > eventTitleMaxLength || utf8.RuneCountInString(strings.TrimSpace(*value)) < eventTitleMinLength {
		return fmt.Errorf("%w: поле title должно содержать минимум %d непустых символов и максимум %d символов", ErrInvalidEventUpdate, eventTitleMinLength, eventTitleMaxLength)
	}

	return nil
}

func validateUpdatedEventDescription(value *string) error {
	if value == nil {
		return nil
	}

	length := utf8.RuneCountInString(*value)
	if length > eventDescriptionMaxLength || utf8.RuneCountInString(strings.TrimSpace(*value)) < eventDescriptionMinLength {
		return fmt.Errorf("%w: поле description должно содержать минимум %d непустых символов и максимум %d символов", ErrInvalidEventUpdate, eventDescriptionMinLength, eventDescriptionMaxLength)
	}

	return nil
}

func validateUpdatedPositiveUint(field string, value *uint) error {
	if value == nil {
		return nil
	}
	if *value == 0 {
		return fmt.Errorf("%w: поле %s должно быть больше 0", ErrInvalidEventUpdate, field)
	}

	return nil
}

func validateUpdatedEventImageURL(value *string) error {
	if value == nil {
		return nil
	}

	parsed, err := url.ParseRequestURI(*value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%w: поле imageUrl должно содержать абсолютный URL", ErrInvalidEventUpdate)
	}

	return nil
}

func validateUpdatedEventStartTime(value *time.Time, now time.Time) error {
	if value == nil {
		return nil
	}
	if !value.After(now) {
		return fmt.Errorf("%w: поле startTime должно быть строго в будущем", ErrInvalidEventUpdate)
	}

	return nil
}

func validateUpdatedEventDuration(value *uint16) error {
	if value == nil {
		return nil
	}
	if *value < eventDurationMinMinutes || *value > eventDurationMaxMinutes {
		return fmt.Errorf("%w: поле duration должно быть от %d до %d минут", ErrInvalidEventUpdate, eventDurationMinMinutes, eventDurationMaxMinutes)
	}

	return nil
}

func validateUpdatedEventMaxUsers(value *uint16) error {
	if value == nil {
		return nil
	}
	if *value < eventMinUsers || *value > eventMaxUsers {
		return fmt.Errorf("%w: поле maxUsers должно быть от %d до %d", ErrInvalidEventUpdate, eventMinUsers, eventMaxUsers)
	}

	return nil
}

func validateUpdatedMaxLength(field string, value *string, maxLength int) error {
	if value == nil {
		return nil
	}
	if utf8.RuneCountInString(*value) > maxLength {
		return fmt.Errorf("%w: поле %s не должно превышать %d символов", ErrInvalidEventUpdate, field, maxLength)
	}

	return nil
}

func validateUpdatedGenres(genreIDs []uint) error {
	if genreIDs == nil {
		return nil
	}
	if err := validateEventGenreCount(genreIDs); err != nil {
		return err
	}

	seen := make(map[uint]struct{}, len(genreIDs))
	for _, genreID := range genreIDs {
		if genreID == 0 {
			return fmt.Errorf("%w: идентификаторы жанров должны быть больше 0", ErrInvalidGenres)
		}
		if _, exists := seen[genreID]; exists {
			return fmt.Errorf("%w: жанры не должны повторяться", ErrInvalidGenres)
		}
		seen[genreID] = struct{}{}
	}

	return nil
}
