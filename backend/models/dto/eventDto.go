package dto

import "time"

type EventFullDto struct {
	// Базовая информация
	ID           uint          `json:"id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	ImageURL     string        `json:"imageUrl"`
	MaxUsers     uint16        `json:"maxUsers"`
	CurrentUsers uint16        `json:"currentUsers"`
	StartTime    time.Time     `json:"startTime"`
	EndTime      time.Time     `json:"endTime"`
	Duration     uint16        `json:"duration"`
	Group        EventGroupDto `json:"group"`

	// Типы и категории
	EventType    uint     `json:"eventType"`
	LocationType uint     `json:"location"`
	Status       uint     `json:"status"`
	Genres       []string `json:"genres"`

	// Создатель
	Creator EventCreatorDto `json:"creator"`

	// Дополнительная информация
	Address  string `json:"address"`
	Country  string `json:"country"`
	AgeLimit string `json:"ageLimit"`
	Year     *int   `json:"year,omitempty"`
	Notes    string `json:"notes"`

	// Произвольные поля
	CustomFields map[string]interface{} `json:"customFields"`

	Subscribed bool `json:"subscribed"`
	IsCreator  bool `json:"isCreator"`
	// Даты
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Вспомогательные DTO
type EventCreatorDto struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Us       string `json:"us"`
	Image    string `json:"image"`
	Verified bool   `json:"verified"`
}

type EventGroupDto struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	Enterprise bool   `json:"enterprise"`
}

// CachedPopularEvents - структура для кэша популярных событий
type CachedPopularEvents struct {
	Events    []EventSearchItemDto `json:"events"`
	UpdatedAt time.Time            `json:"updated_at"`
	Count     int                  `json:"count"`
}

type EventSearchGroupDto struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	Enterprise bool   `json:"enterprise"`
}

type EventSearchItemDto struct {
	ID           uint                `json:"id"`
	Title        string              `json:"title"`
	Group        EventSearchGroupDto `json:"group"`
	Image        string              `json:"image"`
	CurrentUsers uint16              `json:"currentUsers"`
	MaxUsers     uint16              `json:"maxUsers"`
	Duration     uint16              `json:"duration"`
	StartTime    time.Time           `json:"startTime"`
	EventType    string              `json:"eventType"`
	LocationType string              `json:"locationType"`
	AgeLimit     string              `json:"ageLimit"`
	Status       string              `json:"status"`
	City         string              `json:"city,omitempty"`
	Genres       []string            `json:"genres"`
	Subscribed   bool                `json:"subscribed"`
}

type EventSearchResponse struct {
	Items       []EventSearchItemDto `json:"items"`
	Total       int64                `json:"total"`
	Limit       int                  `json:"limit"`
	CurrentPage int                  `json:"currentPage"`
	TotalPages  int                  `json:"totalPages"`
	HasMore     bool                 `json:"hasMore"`
}
