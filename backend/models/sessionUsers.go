package models

import "time"

type SessionUser struct {
	SessionID uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"primaryKey"`
	Role      string `gorm:"type:text"`

	Session   Session `gorm:"foreignKey:SessionID;references:ID;constraint:OnDelete:CASCADE"`
	User      User    `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
