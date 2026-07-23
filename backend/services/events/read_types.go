package events

import (
	"strings"
	"time"

	"friendship/models/dto"
)

type EventSearchInput struct {
	Query                string
	GroupID              *uint
	CategoryIDs          []uint
	ExcludeCategoryIDs   []uint
	GenreIDs             []uint
	ExcludeGenreIDs      []uint
	EventTypeIDs         []uint
	ExcludeEventTypeIDs  []uint
	LocationTypes        []string
	ExcludeLocationTypes []string
	City                 string
	DateFrom             *time.Time
	DateTo               *time.Time
	HasFreeSlots         *bool
	OnlySubscriptionNews bool
	Page                 int
	Limit                int
}

type EventSearchQuery struct {
	ViewerID             uint
	Query                string
	GroupID              *uint
	CategoryIDs          []uint
	ExcludeCategoryIDs   []uint
	GenreIDs             []uint
	ExcludeGenreIDs      []uint
	EventTypeIDs         []uint
	ExcludeEventTypeIDs  []uint
	LocationTypes        []string
	ExcludeLocationTypes []string
	City                 string
	DateFrom             *time.Time
	DateTo               *time.Time
	HasFreeSlots         *bool
	OnlySubscriptionNews bool
	Page                 int
	Limit                int
}

type EventGroupEventsQuery struct {
	ViewerID uint
	GroupID  uint
}

type EventDetailsQuery struct {
	ViewerID uint
	EventID  uint
}

type EventSearchPageView struct {
	Total int64
	Items []EventSearchItemView
}

type EventSearchItemView struct {
	ID                uint
	Title             string
	Group             EventReadGroupView
	ImageURL          string
	ParticipantsCount uint16
	MaxUsers          uint16
	Duration          uint16
	StartTime         time.Time
	EventType         string
	LocationType      string
	City              string
	Genres            []string
	ViewerSubscribed  bool
}

type EventGroupEventsView struct {
	GroupFound          bool
	GroupPrivate        bool
	ViewerIsGroupMember bool
	Items               []EventShortView
}

type EventShortView struct {
	ID               uint
	Title            string
	ImageURL         string
	MaxUsers         uint16
	CurrentUsers     uint16
	EventTypeID      uint
	LocationTypeID   uint
	AgeLimit         string
	Genres           []string
	StartTime        time.Time
	Duration         uint16
	GroupID          uint
	Status           string
	ViewerSubscribed bool
}

type EventDetailsView struct {
	Found               bool
	GroupPrivate        bool
	ViewerIsGroupMember bool
	Event               EventFullView
}

type EventReadGroupView struct {
	ID         uint
	Name       string
	Image      string
	Enterprise bool
	City       string
}

type EventReadCreatorView struct {
	ID       uint
	Name     string
	Us       string
	Image    string
	Verified bool
}

type EventFullView struct {
	ID               uint
	Title            string
	Description      string
	ImageURL         string
	MaxUsers         uint16
	CurrentUsers     uint16
	StartTime        time.Time
	EndTime          time.Time
	Duration         uint16
	Group            EventReadGroupView
	EventTypeID      uint
	LocationTypeID   uint
	StatusID         uint
	Genres           []string
	Creator          EventReadCreatorView
	Address          string
	Country          string
	AgeLimit         string
	Year             *int
	Notes            string
	CustomFields     map[string]interface{}
	ViewerSubscribed bool
	ViewerIsCreator  bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func normalizeEventSearchInput(input EventSearchInput) EventSearchInput {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Page > 10000 {
		input.Page = 10000
	}
	if input.Limit < 1 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	input.Query = strings.TrimSpace(input.Query)
	input.City = strings.TrimSpace(input.City)
	input.LocationTypes = normalizeLocationTypes(input.LocationTypes)
	input.ExcludeLocationTypes = normalizeLocationTypes(input.ExcludeLocationTypes)
	return input
}

func emptyEventSearchResponse(page, limit int) *dto.EventSearchResponse {
	return &dto.EventSearchResponse{
		Items:       []dto.EventSearchItemDto{},
		Total:       0,
		Limit:       limit,
		CurrentPage: page,
		TotalPages:  0,
		HasMore:     false,
	}
}

func calculateTotalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

func maxIntValue() int {
	return int(^uint(0) >> 1)
}

func normalizeLocationTypes(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "online", "онлайн":
			normalized = append(normalized, "online", "онлайн")
		case "offline", "off-line", "офлайн", "оффлайн":
			normalized = append(normalized, "offline", "off-line", "офлайн", "оффлайн")
		case "":
		default:
			normalized = append(normalized, strings.ToLower(strings.TrimSpace(value)))
		}
	}

	return normalized
}
