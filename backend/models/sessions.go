package models

import "time"

type Session struct {
	ID            uint        `json:"id" gorm:"primaryKey;autoIncrement"`
	Title         string      `json:"title" gorm:"not null"`
	SessionTypeID uint16      `json:"-" gorm:"not null"`
	SessionType   SessionType `json:"type" gorm:"foreignKey:SessionTypeID"`
	StartTime     time.Time   `json:"start_time" gorm:"not null"`
	EndTime       time.Time   `json:"end_time" gorm:"not null"`
	UserID        uint16      `json:"-" gorm:"not null"`
	User          User        `json:"user" gorm:"foreignKey:UserID"`
	CurrentUsers  uint16      `json:"current_users" gorm:"not null"`
	CountUsersMax uint16      `json:"count_max" gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
