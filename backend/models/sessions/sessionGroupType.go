package sessions

import "time"

type sessionGroupType struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
