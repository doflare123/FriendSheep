package events

import "time"

// EventReminderIntentOutbox — запись PostgreSQL-адаптера для источника
// расписания уведомлений, который читает notify_service. Внешний ключ на Event
// намеренно отсутствует, чтобы tombstone отмены сохранялся после удаления события.
type EventReminderIntentOutbox struct {
	Sequence              int64  `gorm:"primaryKey;autoIncrement;index:idx_event_reminder_intent_outbox_event_sequence,priority:2,sort:desc"`
	MessageID             string `gorm:"type:uuid;not null;uniqueIndex"`
	SchemaVersion         uint16 `gorm:"not null;default:1;check:event_reminder_schema_version_check,schema_version = 1"`
	IntentType            string `gorm:"type:varchar(64);not null;check:event_reminder_intent_type_check,intent_type = 'event_reminder'"`
	Operation             string `gorm:"type:varchar(32);not null;check:event_reminder_operation_check,operation IN ('schedule_upsert','schedule_cancel')"`
	EventID               uint   `gorm:"not null;index:idx_event_reminder_intent_outbox_event_sequence,priority:1"`
	StartTime             *time.Time
	ReminderOffsetMinutes []int     `gorm:"serializer:json;type:jsonb"`
	OccurredAt            time.Time `gorm:"not null;index:idx_event_reminder_intent_outbox_occurred_at"`
}

func (EventReminderIntentOutbox) TableName() string {
	return "event_reminder_intent_outbox"
}
