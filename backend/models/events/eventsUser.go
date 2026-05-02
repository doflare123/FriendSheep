package events

import (
	"friendship/models"
	"time"
)

type EventsUser struct {
	ID      uint  `gorm:"primaryKey;autoIncrement"`
	EventID uint  `gorm:"not null;index;uniqueIndex:idx_event_user_membership" json:"eventId"`
	Event   Event `gorm:"foreignKey:EventID" json:"-"`

	UserID uint        `gorm:"not null;index;uniqueIndex:idx_event_user_membership" json:"userId"`
	User   models.User `gorm:"foreignKey:UserID" json:"user"`

	JoinedAt time.Time `gorm:"autoCreateTime" json:"joinedAt"`
}
