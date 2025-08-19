package sessions

import (
	"notify_service/models"
	"notify_service/models/groups"
	"time"
)

type Session struct {
	ID             uint              `gorm:"primaryKey;autoIncrement"`
	Title          string            `gorm:"not null"`
	SessionTypeID  uint              `gorm:"not null"`
	SessionType    models.Category   `gorm:"foreignKey:SessionTypeID"`
	SessionPlaceID uint              `gorm:"not null"`
	SessionPlace   SessionGroupPlace `gorm:"foreignKey:SessionPlaceID"`

	GroupID uint         `gorm:"not null"`
	Group   groups.Group `json:"-" gorm:"foreignKey:GroupID"`

	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`
	Duration  uint16    `gorm:"null"`

	UserID uint        `gorm:"not null"` // создатель
	User   models.User `json:"-" gorm:"foreignKey:UserID"`

	CurrentUsers  uint16 `gorm:"not null;default:0"`
	CountUsersMax uint16 `gorm:"not null"`

	ImageURL string `gorm:"type:text"` // путь к картинке
	StatusID uint   `gorm:"not null"`
	Status   Status `gorm:"foreignKey:StatusID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
