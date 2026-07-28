package events

import "time"

type PopularEventView struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	ImageURL     string    `json:"imageUrl"`
	MaxUsers     uint16    `json:"maxUsers"`
	CurrentUsers uint16    `json:"currentUsers"`
	EventType    uint      `json:"eventType"`
	LocationType uint      `json:"location"`
	AgeLimit     string    `json:"ageLimit"`
	Genres       []string  `json:"genres"`
	StartTime    time.Time `json:"startTime"`
	Duration     uint16    `json:"duration"`
	EventID      uint      `json:"eventId"`
	GroupID      uint      `json:"groupId"`
	Status       string    `json:"status"`
	Subscribed   bool      `json:"subscribed"`
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
}

type PopularEventNotification struct {
	EventID        uint
	EventName      string
	GroupName      string
	RecipientEmail string
	StartTime      time.Time
}
