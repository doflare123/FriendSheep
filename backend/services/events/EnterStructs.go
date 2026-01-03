package events

import "time"

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
	AgeLimitID   uint                   `json:"ageLimit,omitempty"`
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
