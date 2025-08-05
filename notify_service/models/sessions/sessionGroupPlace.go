package sessions

import "time"

type SessionGroupPlace struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
