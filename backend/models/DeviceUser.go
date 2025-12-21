package models

import "time"

type DeviceUser struct {
	ID          *uint      `gorm:"primaryKey;autoIncrement"`
	UserID      *uint      `gorm:"not null;index"`
	User        User       `gorm:"foreignKey:UserID;references:ID"`
	DeviceToken *string    `json:"device_token" gorm:"not null;index;size:255"`
	Platform    *string    `gorm:"not null;index"` // "web", "android", "ios"
	DeviceInfo  *string    `gorm:"type:text"`      // JSON с доп. информацией об устройстве
	IsActive    *bool      `gorm:"default:true"`   // Активно ли устройство
	LastUsed    *time.Time `gorm:"autoUpdateTime"` // Последнее использование
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
