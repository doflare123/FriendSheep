package sessions

import (
	"friendship/models"
	"friendship/models/groups"
	"time"
)

type Session struct {
	ID            uint             `gorm:"primaryKey;autoIncrement"`
	Title         string           `gorm:"not null"`
	SessionTypeID uint             `gorm:"not null"`
	SessionType   SessionGroupType `gorm:"foreignKey:SessionTypeID"`

	GroupID uint         `gorm:"not null"`
	Group   groups.Group `gorm:"foreignKey:GroupID"`

	StartTime time.Time `gorm:"not null"`
	Duration  uint16    `gorm:"null"`

	UserID uint        `gorm:"not null"` // создатель
	User   models.User `gorm:"foreignKey:UserID"`

	CurrentUsers  uint16 `gorm:"not null;default:0"`
	CountUsersMax uint16 `gorm:"not null"`

	ImageURL string `gorm:"type:text"` // путь к картинке

	CreatedAt time.Time
	UpdatedAt time.Time
}
