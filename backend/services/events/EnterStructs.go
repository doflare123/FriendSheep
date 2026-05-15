package events

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalidStartTimeFormat = errors.New("поле startTime должно быть в формате RFC3339 (например, 2027-01-02T15:04:05Z) или YYYY-MM-DDTHH:MM:SS")

type CreateEventInput struct {
	Title       string    `json:"title" binding:"required,min=5,max=200"`
	Description string    `json:"description" binding:"required,min=10,max=2000"`
	GroupID     uint      `json:"groupId" binding:"required"`
	EventTypeID uint      `json:"eventTypeId" binding:"required"`
	LocationID  uint      `json:"locationId" binding:"required"`
	ImageURL    string    `json:"imageUrl" binding:"required,url"`
	StartTime   time.Time `json:"startTime" binding:"required"`
	Duration    uint16    `json:"duration" binding:"required,min=15,max=1440"`
	MaxUsers    uint16    `json:"maxUsers" binding:"required,min=2,max=1000"`
	Genres      []uint    `json:"genres" binding:"required,min=1,max=9"`

	// Опциональные поля
	Address      string                 `json:"address,omitempty"`
	Country      string                 `json:"country,omitempty"`
	AgeLimitID   uint                   `json:"ageLimit" binding:"required"`
	Year         *int                   `json:"year,omitempty"`
	Notes        string                 `json:"notes,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
}

type UpdateEventInput struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	EventTypeID *uint      `json:"eventTypeId,omitempty"`
	LocationID  *uint      `json:"locationId,omitempty"`
	ImageURL    *string    `json:"imageUrl,omitempty"`
	StartTime   *time.Time `json:"startTime,omitempty"`
	Duration    *uint16    `json:"duration,omitempty"`
	MaxUsers    *uint16    `json:"maxUsers,omitempty"`
	Genres      []uint     `json:"genres,omitempty"` // Если передано, то обновляем

	// Опциональные поля
	Address      *string                 `json:"address,omitempty"`
	Country      *string                 `json:"country,omitempty"`
	AgeLimit     *uint                   `json:"ageLimit,omitempty"`
	Year         *int                    `json:"year,omitempty"`
	Notes        *string                 `json:"notes,omitempty"`
	CustomFields *map[string]interface{} `json:"customFields,omitempty"`
}

func (input *CreateEventInput) UnmarshalJSON(data []byte) error {
	type createEventInputAlias struct {
		Title        string                 `json:"title"`
		Description  string                 `json:"description"`
		GroupID      uint                   `json:"groupId"`
		EventTypeID  uint                   `json:"eventTypeId"`
		LocationID   uint                   `json:"locationId"`
		ImageURL     string                 `json:"imageUrl"`
		StartTimeRaw string                 `json:"startTime"`
		Duration     uint16                 `json:"duration"`
		MaxUsers     uint16                 `json:"maxUsers"`
		Genres       []uint                 `json:"genres"`
		Address      string                 `json:"address,omitempty"`
		Country      string                 `json:"country,omitempty"`
		AgeLimitID   uint                   `json:"ageLimit"`
		Year         *int                   `json:"year,omitempty"`
		Notes        string                 `json:"notes,omitempty"`
		CustomFields map[string]interface{} `json:"customFields,omitempty"`
	}

	var raw createEventInputAlias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsedStartTime, err := parseEventStartTime(raw.StartTimeRaw)
	if err != nil {
		return err
	}

	input.Title = raw.Title
	input.Description = raw.Description
	input.GroupID = raw.GroupID
	input.EventTypeID = raw.EventTypeID
	input.LocationID = raw.LocationID
	input.ImageURL = raw.ImageURL
	input.StartTime = parsedStartTime
	input.Duration = raw.Duration
	input.MaxUsers = raw.MaxUsers
	input.Genres = raw.Genres
	input.Address = raw.Address
	input.Country = raw.Country
	input.AgeLimitID = raw.AgeLimitID
	input.Year = raw.Year
	input.Notes = raw.Notes
	input.CustomFields = raw.CustomFields

	return nil
}

func (input *UpdateEventInput) UnmarshalJSON(data []byte) error {
	type updateEventInputAlias struct {
		Title        *string                 `json:"title,omitempty"`
		Description  *string                 `json:"description,omitempty"`
		EventTypeID  *uint                   `json:"eventTypeId,omitempty"`
		LocationID   *uint                   `json:"locationId,omitempty"`
		ImageURL     *string                 `json:"imageUrl,omitempty"`
		StartTimeRaw *string                 `json:"startTime,omitempty"`
		Duration     *uint16                 `json:"duration,omitempty"`
		MaxUsers     *uint16                 `json:"maxUsers,omitempty"`
		Genres       []uint                  `json:"genres,omitempty"`
		Address      *string                 `json:"address,omitempty"`
		Country      *string                 `json:"country,omitempty"`
		AgeLimit     *uint                   `json:"ageLimit,omitempty"`
		Year         *int                    `json:"year,omitempty"`
		Notes        *string                 `json:"notes,omitempty"`
		CustomFields *map[string]interface{} `json:"customFields,omitempty"`
	}

	var raw updateEventInputAlias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var parsedStartTime *time.Time
	if raw.StartTimeRaw != nil {
		parsed, err := parseEventStartTime(*raw.StartTimeRaw)
		if err != nil {
			return err
		}
		parsedStartTime = &parsed
	}

	input.Title = raw.Title
	input.Description = raw.Description
	input.EventTypeID = raw.EventTypeID
	input.LocationID = raw.LocationID
	input.ImageURL = raw.ImageURL
	input.StartTime = parsedStartTime
	input.Duration = raw.Duration
	input.MaxUsers = raw.MaxUsers
	input.Genres = raw.Genres
	input.Address = raw.Address
	input.Country = raw.Country
	input.AgeLimit = raw.AgeLimit
	input.Year = raw.Year
	input.Notes = raw.Notes
	input.CustomFields = raw.CustomFields

	return nil
}

func parseEventStartTime(raw string) (time.Time, error) {
	const noTimezoneLayout = "2006-01-02T15:04:05"

	if raw == "" {
		return time.Time{}, ErrInvalidStartTimeFormat
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return parsed, nil
		}
	}

	parsedWithoutTimezone, err := time.ParseInLocation(noTimezoneLayout, raw, time.UTC)
	if err == nil {
		return parsedWithoutTimezone, nil
	}

	return time.Time{}, ErrInvalidStartTimeFormat
}
