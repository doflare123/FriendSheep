package models

import "time"

type StatsProcessedEvent struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	SessionID uint      `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
