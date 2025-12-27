package events

import (
	"friendship/models"
	"time"
)

type EventsUser struct {
	ID      uint  `gorm:"primaryKey;autoIncrement"`
	EventID uint  `gorm:"not null;index" json:"eventId"`
	Event   Event `gorm:"foreignKey:EventID" json:"-"`

	UserID uint        `gorm:"not null;index" json:"userId"`
	User   models.User `gorm:"foreignKey:UserID" json:"user"`

	JoinedAt time.Time `gorm:"autoCreateTime" json:"joinedAt"`
}
