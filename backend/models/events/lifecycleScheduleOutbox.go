package events

import "time"

// EventLifecycleScheduleOutbox — запись PostgreSQL-адаптера для источника
// lifecycle-расписания, который читает notify_service. Прикладные и HTTP
// контракты используют storage-независимые типы из services/events.
type EventLifecycleScheduleOutbox struct {
	Sequence      int64  `gorm:"primaryKey;autoIncrement"`
	MessageID     string `gorm:"type:uuid;not null;uniqueIndex"`
	EventID       uint   `gorm:"not null;index"`
	Operation     string `gorm:"type:varchar(32);not null"`
	StartTime     *time.Time
	EndTime       *time.Time
	OccurredAt    time.Time `gorm:"not null;index"`
	SchemaVersion uint16    `gorm:"not null;default:1"`
}

func (EventLifecycleScheduleOutbox) TableName() string {
	return "event_lifecycle_schedule_outbox"
}
