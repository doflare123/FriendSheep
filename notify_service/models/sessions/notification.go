package sessions

import (
	"time"
)

type Notification struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	UserID             uint      `json:"user_id" gorm:"not null;index"`
	SessionID          uint      `json:"session_id" gorm:"not null;index"`
	NotificationTypeID uint      `json:"notification_type_id" gorm:"not null;index"`
	SendAt             time.Time `json:"send_at" gorm:"not null;index"`
	Sent               bool      `json:"sent" gorm:"default:false"`
	Text               string    `json:"text" gorm:"not null"`
	ImageURL           string    `json:"image_url" gorm:"type:text"`
	Title              string    `json:"title" gorm:"not null"`
	CreatedAt          time.Time
}
