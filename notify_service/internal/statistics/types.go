package statistics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

const SourceName = "monolith_event_statistics"

const (
	StateScheduled  = "scheduled"
	StateProcessing = "processing"
	StateRetryWait  = "retry_wait"
	StateCompleted  = "completed"
	StateMissing    = "missing"
)

type Trigger struct {
	Sequence      int64     `json:"sequence"`
	SchemaVersion uint16    `json:"schemaVersion"`
	EventID       uint64    `json:"eventId"`
	CompletedAt   time.Time `json:"completedAt"`
}

type Page struct {
	Messages   []Trigger `json:"messages"`
	NextCursor int64     `json:"nextCursor"`
	HasMore    bool      `json:"hasMore"`
}

type ProcessResult struct {
	EventID     uint64     `json:"eventId"`
	Outcome     string     `json:"outcome"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
}

type Job struct {
	ID             int64
	EventID        uint64
	SourceSequence int64
	AttemptCount   int
	LeaseToken     string
}

type ApplyResult struct {
	LastSequence int64
	AppliedCount int
}

type ClientError struct {
	Code  string
	Cause error
}

func (e *ClientError) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Cause.Error()
	}
	return e.Code
}
func (e *ClientError) Unwrap() error { return e.Cause }

func malformed(err error) error { return &ClientError{Code: "malformed_response", Cause: err} }

func validateBatch(messages []Trigger) error {
	var previous int64
	for _, message := range messages {
		if message.Sequence <= previous || message.SchemaVersion != 1 || message.EventID == 0 || message.EventID > math.MaxInt64 || message.CompletedAt.IsZero() {
			return malformed(errors.New("некорректный statistics trigger или порядок sequence"))
		}
		previous = message.Sequence
	}
	return nil
}

func validatePage(page Page, after int64, limit int) error {
	if page.Messages == nil || len(page.Messages) > limit || (page.HasMore && len(page.Messages) == 0) {
		return malformed(errors.New("некорректный размер statistics batch"))
	}
	if err := validateBatch(page.Messages); err != nil {
		return err
	}
	if len(page.Messages) == 0 {
		if page.NextCursor != after {
			return malformed(errors.New("пустой batch изменил cursor"))
		}
		return nil
	}
	if page.Messages[0].Sequence <= after || page.NextCursor != page.Messages[len(page.Messages)-1].Sequence {
		return malformed(errors.New("некорректный statistics cursor"))
	}
	return nil
}

type Clock interface{ Now() time.Time }
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

func wait(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func code(err error) string {
	var typed *ClientError
	if errors.As(err, &typed) {
		return typed.Code
	}
	return fmt.Sprintf("%T", err)
}
