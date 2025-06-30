package sessions

import "time"

type Notification struct {
	ID                 uint             `gorm:"primaryKey"`
	UserID             uint             `gorm:"not null;index"`
	SessionID          uint             `gorm:"not null;index"`
	NotificationTypeID uint             `gorm:"not null;index"`
	NotificationType   NotificationType `gorm:"foreignKey:NotificationTypeID"`
	SendAt             time.Time        `gorm:"not null;index"`
	Sent               bool             `gorm:"default:false"`
	CreatedAt          time.Time
}
