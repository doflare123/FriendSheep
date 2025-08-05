package sessions

import "time"

type Notification struct {
	ID                 uint      `gorm:"primaryKey"`
	UserID             uint      `gorm:"not null;index"`
	SessionID          uint      `gorm:"not null;index"`
	NotificationTypeID uint      `gorm:"not null;index"`
	SendAt             time.Time `gorm:"not null;index"`
	Sent               bool      `gorm:"default:false"`
	Text               string    `gorm:"not null"`
	ImageURL           string    `gorm:"type:text"`
	Title              string    `gorm:"not null"`
	CreatedAt          time.Time
}
