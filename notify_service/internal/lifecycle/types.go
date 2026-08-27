package lifecycle

import "time"

const (
	SourceName = "monolith_event_lifecycle"

	OperationScheduleUpsert = "schedule_upsert"
	OperationScheduleCancel = "schedule_cancel"

	StateScheduled      = "scheduled"
	StateActiveWait     = "active_wait"
	StateProcessing     = "processing"
	StateCompleted      = "completed"
	StateCancelled      = "cancelled"
	StateRetryWait      = "retry_wait"
	StateTerminalFailed = "terminal_failed"

	OutcomeNotDue           = "not_due"
	OutcomeStarted          = "started"
	OutcomeCompleted        = "completed"
	OutcomeCompletedCatchUp = "completed_catch_up"
	OutcomeAlreadyActive    = "already_active"
	OutcomeAlreadyCompleted = "already_completed"

	ErrorCodeInvalidInternalToken = "invalid_internal_token"
	ErrorCodeEventNotFound        = "event_not_found"
	ErrorCodeInvalidEventID       = "invalid_event_id"
	ErrorCodeInvalidLifecycle     = "invalid_event_lifecycle_state"
	ErrorCodeMalformedResponse    = "malformed_response"
	ErrorCodeUnexpectedStatus     = "unexpected_status"
	ErrorCodeNetwork              = "network_error"
	ErrorCodeTimeout              = "timeout"
)

type ScheduleEvent struct {
	Sequence      uint64     `json:"sequence"`
	MessageID     string     `json:"messageId"`
	SchemaVersion uint16     `json:"schemaVersion"`
	Operation     string     `json:"operation"`
	EventID       uint64     `json:"eventId"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	OccurredAt    time.Time  `json:"occurredAt"`
}

type SchedulePage struct {
	Items      []ScheduleEvent `json:"items"`
	NextCursor uint64          `json:"nextCursor"`
	HasMore    bool            `json:"hasMore"`
}

type AdvanceResult struct {
	EventID        uint64    `json:"eventId"`
	PreviousStatus string    `json:"previousStatus"`
	CurrentStatus  string    `json:"currentStatus"`
	Outcome        string    `json:"outcome"`
	Applied        bool      `json:"applied"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	ProcessedAt    time.Time `json:"processedAt"`
}

type Job struct {
	ID              int64
	EventID         uint64
	SourceSequence  uint64
	SourceMessageID string
	StartTime       *time.Time
	EndTime         *time.Time
	State           string
	NextActionAt    *time.Time
	AttemptCount    int
	NextAttemptAt   *time.Time
	LeaseUntil      *time.Time
	LeaseToken      string
	LastErrorCode   string
}

type ApplyResult struct {
	LastSequence uint64
	AppliedCount int
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
