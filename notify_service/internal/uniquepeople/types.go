package uniquepeople

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	SchemaVersion = 1
	maxClaimLimit = 8
)

const (
	OutcomeApplied        = "applied"
	OutcomeAlreadyApplied = "already_applied"
	OutcomeStaleLease     = "stale_lease"
	OutcomeRetryScheduled = "retry_scheduled"
)

type Claim struct {
	UserID        uint64     `json:"userId"`
	Generation    uint64     `json:"generation"`
	LeaseToken    string     `json:"leaseToken"`
	LeaseUntil    time.Time  `json:"leaseUntil"`
	DirtySince    *time.Time `json:"dirtySince,omitempty"`
	DueAt         *time.Time `json:"dueAt,omitempty"`
	SchemaVersion uint16     `json:"schemaVersion,omitempty"`
}

type ClaimResponse struct {
	SchemaVersion uint16  `json:"schemaVersion"`
	Jobs          []Claim `json:"jobs"`
}

type ExecuteResult struct {
	SchemaVersion uint16     `json:"schemaVersion"`
	UserID        uint64     `json:"userId"`
	Generation    uint64     `json:"generation"`
	Outcome       string     `json:"outcome"`
	Count         *int64     `json:"count,omitempty"`
	CalculatedAt  *time.Time `json:"calculatedAt,omitempty"`
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

func malformed(err error) error {
	return &ClientError{Code: "malformed_response", Cause: err}
}

func errorCode(err error) string {
	var typed *ClientError
	if errors.As(err, &typed) {
		return typed.Code
	}
	return fmt.Sprintf("%T", err)
}

func validateClaimResponse(response ClaimResponse, limit int, now time.Time) error {
	if response.SchemaVersion != SchemaVersion || response.Jobs == nil || len(response.Jobs) > limit {
		return malformed(errors.New("некорректная версия или размер claim batch"))
	}
	seen := make(map[uint64]struct{}, len(response.Jobs))
	for _, job := range response.Jobs {
		if job.UserID == 0 || job.Generation == 0 || strings.TrimSpace(job.LeaseToken) == "" || job.LeaseUntil.IsZero() || !job.LeaseUntil.After(now) {
			return malformed(errors.New("некорректный unique people claim"))
		}
		if job.SchemaVersion != 0 && job.SchemaVersion != SchemaVersion {
			return malformed(errors.New("неподдерживаемая версия unique people claim"))
		}
		if job.DirtySince != nil && job.DirtySince.IsZero() {
			return malformed(errors.New("некорректный dirtySince"))
		}
		if job.DueAt != nil && job.DueAt.IsZero() {
			return malformed(errors.New("некорректный dueAt"))
		}
		if _, exists := seen[job.UserID]; exists {
			return malformed(errors.New("claim batch содержит дубликат"))
		}
		seen[job.UserID] = struct{}{}
	}
	return nil
}

func validateExecuteResult(result ExecuteResult, claim Claim) error {
	if result.SchemaVersion != SchemaVersion || result.UserID != claim.UserID {
		return malformed(errors.New("ответ execute не соответствует claim"))
	}
	if result.Count != nil && *result.Count < 0 {
		return malformed(errors.New("count не может быть отрицательным"))
	}
	if result.CalculatedAt != nil && result.CalculatedAt.IsZero() {
		return malformed(errors.New("некорректный calculatedAt"))
	}
	switch result.Outcome {
	case OutcomeApplied, OutcomeAlreadyApplied:
		if result.Generation < claim.Generation {
			return malformed(errors.New("generation ответа меньше claim generation"))
		}
		if result.Count == nil || result.CalculatedAt == nil {
			return malformed(errors.New("успешный execute не содержит count или calculatedAt"))
		}
		return nil
	case OutcomeStaleLease:
		if result.Generation != 0 && result.Generation < claim.Generation {
			return malformed(errors.New("generation ответа меньше claim generation"))
		}
		return nil
	case OutcomeRetryScheduled:
		if result.Generation < claim.Generation {
			return malformed(errors.New("generation ответа меньше claim generation"))
		}
		return nil
	default:
		return malformed(errors.New("неизвестный unique people outcome"))
	}
}
