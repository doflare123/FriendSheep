package models

import "time"

type SessionType struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
