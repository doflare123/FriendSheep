package events

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"friendship/models"
	"friendship/models/groups"
	"time"
)

type Event struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string `gorm:"not null;size:200" json:"title"`
	Description string `gorm:"type:text" json:"description"`

	// Тип события
	EventTypeID uint            `gorm:"not null" json:"eventTypeId"`
	EventType   models.Category `gorm:"foreignKey:EventTypeID" json:"eventType"`

	// Тип проведения
	EventLocationID uint          `gorm:"not null" json:"eventLocationId"`
	EventLocation   EventLocation `gorm:"foreignKey:EventLocationID" json:"eventLocation"`

	// Привязка к группе
	GroupID uint         `gorm:"not null;index" json:"groupId"`
	Group   groups.Group `gorm:"foreignKey:GroupID" json:"-"`

	// Время проведения
	StartTime time.Time `gorm:"not null;index" json:"startTime"`
	EndTime   time.Time `gorm:"not null" json:"endTime"`
	Duration  uint16    `gorm:"not null" json:"duration"` // в минутах

	// Создатель события
	CreatorID uint        `gorm:"not null;index" json:"creatorId"`
	Creator   models.User `gorm:"foreignKey:CreatorID" json:"creator"`

	// Количество участников
	CurrentUsers uint16 `gorm:"not null;default:0" json:"currentUsers"`
	MaxUsers     uint16 `gorm:"not null" json:"maxUsers"`

	// Изображение
	ImageURL string `gorm:"type:text" json:"imageUrl"`

	// Статус события
	StatusID uint   `gorm:"not null;default:1" json:"statusId"`
	Status   Status `gorm:"foreignKey:StatusID" json:"status"`

	// Дополнительная информация (из MongoDB)
	Adress     string   `gorm:"type:varchar(500)" json:"adress"` // Конкретный адрес/место
	Country    string   `gorm:"type:varchar(100)" json:"country"`
	AgeLimitID uint     `gorm:"index" json:"ageLimitId"`
	AgeLimit   AgeLimit `gorm:"foreignKey:AgeLimitID"` // "18+", "12+", "6+" и т.д.
	Year       *int     `gorm:"type:int" json:"year,omitempty"`
	Notes      string   `gorm:"type:text" json:"notes"` // Дополнительные заметки

	// JSON поле для произвольных данных (замена fields из MongoDB)
	CustomFields CustomFields `gorm:"type:jsonb" json:"customFields"`

	// Связи
	Genres []EventGenre `gorm:"foreignKey:EventID" json:"genres"`
	Users  []EventsUser `gorm:"foreignKey:EventID" json:"-"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CustomFields map[string]interface{}

func (cf CustomFields) Value() (driver.Value, error) {
	if cf == nil {
		return nil, nil
	}
	return json.Marshal(cf)
}

func (cf *CustomFields) Scan(value interface{}) error {
	if value == nil {
		*cf = make(CustomFields)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	result := make(CustomFields)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*cf = result
	return nil
}
