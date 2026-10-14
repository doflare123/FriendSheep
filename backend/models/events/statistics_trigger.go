package events

import "time"

// StatisticsTrigger сохраняется вместе с первым завершением мероприятия.
// Sequence выдаётся под общей блокировкой публикации event outbox; CACHE 1 обязателен.
type StatisticsTrigger struct {
	Sequence      int64     `gorm:"primaryKey;autoIncrement" json:"sequence"`
	SchemaVersion int       `gorm:"not null" json:"schemaVersion"`
	EventID       uint      `gorm:"not null;uniqueIndex" json:"eventId"`
	CompletedAt   time.Time `gorm:"not null" json:"completedAt"`
}

func (StatisticsTrigger) TableName() string { return "event_statistics_triggers" }
