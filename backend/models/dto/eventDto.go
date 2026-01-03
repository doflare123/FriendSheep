package dto

import "time"

// краткая информация о событии
type EventShortDto struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	ImageURL     string    `json:"imageUrl"`
	MaxUsers     uint16    `json:"maxUsers"`
	CurrentUsers uint16    `json:"currentUsers"`
	EventType    uint      `json:"eventType"` // Название типа события
	LocationType uint      `json:"location"`  // Название места проведения (онлайн/оффлайн)
	AgeLimit     string    `json:"ageLimit"`  // Возрастное ограничение
	Genres       []string  `json:"genres"`    // Список названий жанров
	StartTime    time.Time `json:"startTime"` // Дата и время начала
	Duration     uint16    `json:"duration"`  // Длительность в минутах
	EventID      uint      `json:"eventId"`   // ID события (дублирует ID для удобства)
	GroupID      uint      `json:"groupId"`   // ID группы
	Status       string    `json:"status"`    // Статус события
}

type EventFullDto struct {
	// Базовая информация
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"imageUrl"`
	MaxUsers     uint16    `json:"maxUsers"`
	CurrentUsers uint16    `json:"currentUsers"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Duration     uint16    `json:"duration"`
	GroupID      uint      `json:"groupId"`

	// Типы и категории
	EventType    uint     `json:"eventType"`
	LocationType uint     `json:"location"`
	Status       uint     `json:"status"`
	Genres       []string `json:"genres"`

	// Создатель
	Creator EventCreatorDto `json:"creator"`

	// Дополнительная информация (из MongoDB)
	Address  string `json:"address"`
	Country  string `json:"country"`
	AgeLimit string `json:"ageLimit"`
	Year     *int   `json:"year,omitempty"`
	Notes    string `json:"notes"`

	// Произвольные поля (из CustomFields)
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
	Username string `json:"username"`
	Image    string `json:"image"`
}
