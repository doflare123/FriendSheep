package events

import "time"

type PopularEventGroupView struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	Enterprise bool   `json:"enterprise"`
}

type PopularEventView struct {
	ID           uint                  `json:"id"`
	Title        string                `json:"title"`
	Group        PopularEventGroupView `json:"group"`
	Image        string                `json:"image"`
	CurrentUsers uint16                `json:"currentUsers"`
	MaxUsers     uint16                `json:"maxUsers"`
	Duration     uint16                `json:"duration"`
	StartTime    time.Time             `json:"startTime"`
	EventType    string                `json:"eventType"`
	LocationType string                `json:"locationType"`
	AgeLimit     string                `json:"ageLimit"`
	Status       string                `json:"status"`
	City         string                `json:"city,omitempty"`
	Genres       []string              `json:"genres"`
	Subscribed   bool                  `json:"subscribed"`
}

type PopularEventsSnapshot struct {
	Events    []PopularEventView `json:"events"`
	UpdatedAt time.Time          `json:"updated_at"`
	Count     int                `json:"count"`
}

type PopularEventRecord struct {
	PopularEventView
	GroupName      string
	OwnerEmail     string
	OwnerUserID    uint
	PopularityRate float64
	StartTime      time.Time
}

type PopularEventNotification struct {
	EventID        uint
	EventName      string
	GroupName      string
	RecipientEmail string
	StartTime      time.Time
}
