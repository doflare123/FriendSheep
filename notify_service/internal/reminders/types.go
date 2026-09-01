package reminders

import "time"

const (
	SourceName = "monolith_event_reminders"

	IntentTypeEventReminder = "event_reminder"

	OperationScheduleUpsert = "schedule_upsert"
	OperationScheduleCancel = "schedule_cancel"

	StateScheduled           = "scheduled"
	StateProcessing          = "processing"
	StateRetryWait           = "retry_wait"
	StateCompleted           = "completed"
	StateCancelled           = "cancelled"
	StateSuperseded          = "superseded"
	StateSkippedMissedWindow = "skipped_missed_window"
	StateExpired             = "expired"
	StateTerminalFailed      = "terminal_failed"

	NotificationKindEventReminder = "event_reminder"
	ResourceTypeEvent             = "event"

	DeliveryStatePending        = "pending"
	DeliveryStateProcessing     = "processing"
	DeliveryStateRetryWait      = "retry_wait"
	DeliveryStateDelivered      = "delivered"
	DeliveryStateTerminalFailed = "terminal_failed"

	ChannelCodeInApp = "in_app"

	ErrorCodeInvalidInternalToken = "invalid_internal_token"
	ErrorCodeEventNotFound        = "event_not_found"
	ErrorCodeInvalidEventID       = "invalid_event_id"
	ErrorCodeInvalidReminder      = "invalid_reminder_offset"
	ErrorCodeInvalidChannel       = "invalid_recipient_channel"
	ErrorCodeStaleSchedule        = "stale_schedule"
	ErrorCodeMalformedResponse    = "malformed_response"
	ErrorCodeUnexpectedStatus     = "unexpected_status"
	ErrorCodeNetwork              = "network_error"
	ErrorCodeTimeout              = "timeout"
	ErrorCodeNotificationNotFound = "notification_not_found"
	ErrorCodeInvalidCursor        = "invalid_cursor"
	ErrorCodeInvalidUserID        = "invalid_user_id"
	ErrorCodeInvalidLimit         = "invalid_limit"
	ErrorCodeInvalidUnreadFilter  = "invalid_unread_filter"
)

type ReminderIntent struct {
	Sequence              uint64     `json:"sequence"`
	MessageID             string     `json:"messageId"`
	SchemaVersion         uint16     `json:"schemaVersion"`
	IntentType            string     `json:"intentType"`
	Operation             string     `json:"operation"`
	EventID               uint64     `json:"eventId"`
	StartTime             *time.Time `json:"startTime,omitempty"`
	ReminderOffsetMinutes []int      `json:"reminderOffsetMinutes,omitempty"`
	OccurredAt            time.Time  `json:"occurredAt"`
}

type IntentPage struct {
	Items      []ReminderIntent `json:"items"`
	NextCursor uint64           `json:"nextCursor"`
	HasMore    bool             `json:"hasMore"`
}

type Recipient struct {
	UserID   uint64   `json:"userId"`
	Channels []string `json:"channels"`
}

type RecipientSnapshot struct {
	EventID               uint64      `json:"eventId"`
	Title                 string      `json:"title"`
	StartTime             time.Time   `json:"startTime"`
	ReminderOffsetMinutes int         `json:"reminderOffsetMinutes"`
	Recipients            []Recipient `json:"recipients"`
}

type Job struct {
	ID                    int64
	EventID               uint64
	SourceRevision        uint64
	SourceMessageID       string
	ReminderOffsetMinutes int
	StartTime             *time.Time
	DueAt                 *time.Time
	WindowEndAt           *time.Time
	State                 string
	AttemptCount          int
	NextAttemptAt         *time.Time
	LeaseUntil            *time.Time
	LeaseToken            string
	LastErrorCode         string
}

type ApplyResult struct {
	LastSequence uint64
	AppliedCount int
}

type NotificationPayload struct {
	SchemaVersion uint16    `json:"schemaVersion"`
	EventID       uint64    `json:"eventId"`
	Title         string    `json:"title"`
	StartTime     time.Time `json:"startTime"`
}

type NotificationRecord struct {
	ID                    string              `json:"id"`
	UserID                uint64              `json:"userId"`
	Kind                  string              `json:"kind"`
	ResourceType          string              `json:"resourceType"`
	ResourceID            uint64              `json:"resourceId"`
	ReminderOffsetMinutes int                 `json:"reminderOffsetMinutes"`
	ScheduleRevision      uint64              `json:"scheduleRevision"`
	Payload               NotificationPayload `json:"payload"`
	CreatedAt             time.Time           `json:"createdAt"`
	ReadAt                *time.Time          `json:"readAt,omitempty"`
}

type DeliveryTarget struct {
	ID                int64
	NotificationID    string
	ChannelCode       string
	State             string
	AttemptCount      int
	NextAttemptAt     *time.Time
	LeaseUntil        *time.Time
	LeaseToken        string
	LastErrorCode     string
	ProviderMessageID string
	IdempotencyKey    string
	Notification      NotificationRecord
}

type NotificationPage struct {
	Items      []NotificationRecord `json:"items"`
	NextCursor string               `json:"nextCursor"`
	HasMore    bool                 `json:"hasMore"`
}

type ClientError struct {
	Code      string
	Retryable bool
	Terminal  bool
	Cause     error
}

func (e *ClientError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Code + ": " + e.Cause.Error()
	}
	return e.Code
}

func (e *ClientError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
