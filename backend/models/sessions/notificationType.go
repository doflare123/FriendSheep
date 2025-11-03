package sessions

import "time"

type NotificationType struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"unique;not null"` // "24_hours", "6_hours", "1_hour"
	Description string `json:"description" gorm:"not null"` // "За 24 часа до начала"
	HoursBefore int    `json:"hours_before" gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
