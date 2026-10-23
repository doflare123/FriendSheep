package statistics

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	UniquePeopleSchemaVersion       = 1
	UniquePeopleStatusNotCalculated = "not_calculated"
	UniquePeopleStatusPending       = "pending"
	UniquePeopleStatusFresh         = "fresh"

	UniquePeopleOutcomeApplied        = "applied"
	UniquePeopleOutcomeAlreadyApplied = "already_applied"
	UniquePeopleOutcomeStaleLease     = "stale_lease"
	UniquePeopleOutcomeRetryScheduled = "retry_scheduled"
)

var ErrUniquePeopleCapacity = errors.New("лимит одновременных расчётов исчерпан")
var ErrUniquePeopleStaleLease = errors.New("lease расчёта больше не действует")
var ErrUniquePeopleInProgress = errors.New("расчёт пользователя уже выполняется")

type UniquePeopleConfig struct {
	ClaimBatch    int
	Concurrency   int
	QueryTimeout  time.Duration
	LeaseDuration time.Duration
	InitialSpread time.Duration
	TargetWindow  time.Duration
	RetryBase     time.Duration
	RetryMax      time.Duration
	RetryJitter   float64
}

func DefaultUniquePeopleConfig() UniquePeopleConfig {
	return UniquePeopleConfig{
		ClaimBatch: 1, Concurrency: 1, QueryTimeout: 20 * time.Second,
		LeaseDuration: 2 * time.Minute, InitialSpread: 18 * time.Hour,
		TargetWindow: 24 * time.Hour, RetryBase: time.Minute,
		RetryMax: time.Hour, RetryJitter: 0.2,
	}
}

func (c UniquePeopleConfig) Validate() error {
	if c.ClaimBatch < 1 || c.ClaimBatch > 8 || c.Concurrency < 1 || c.Concurrency > 8 {
		return fmt.Errorf("%w: некорректные ограничения batch/concurrency", ErrInvalidInput)
	}
	if c.QueryTimeout <= 0 || c.LeaseDuration <= c.QueryTimeout || c.InitialSpread <= 0 || c.TargetWindow <= c.InitialSpread {
		return fmt.Errorf("%w: некорректные временные ограничения", ErrInvalidInput)
	}
	if c.RetryBase <= 0 || c.RetryMax < c.RetryBase || c.RetryJitter < 0 || c.RetryJitter > 1 {
		return fmt.Errorf("%w: некорректные параметры retry", ErrInvalidInput)
	}
	return nil
}

type UniquePeopleClaim struct {
	UserID              uint      `json:"userId"`
	RequestedGeneration uint64    `json:"generation"`
	LeaseToken          string    `json:"leaseToken"`
	DirtySince          time.Time `json:"dirtySince,omitempty"`
	DueAt               time.Time `json:"dueAt,omitempty"`
	LeaseExpiresAt      time.Time `json:"leaseUntil"`
}

type UniquePeopleClaims struct {
	SchemaVersion int                 `json:"schemaVersion"`
	Jobs          []UniquePeopleClaim `json:"jobs"`
}

type UniquePeopleExecution struct {
	SchemaVersion     int        `json:"schemaVersion"`
	UserID            uint       `json:"userId"`
	AppliedGeneration uint64     `json:"generation"`
	Outcome           string     `json:"outcome"`
	Count             *int64     `json:"count,omitempty"`
	CalculatedAt      *time.Time `json:"calculatedAt,omitempty"`
}

type UniquePeopleBackfillResult struct {
	Enqueued        int64 `json:"enqueued"`
	NextAfterUserID uint  `json:"nextAfterUserId"`
	HasMore         bool  `json:"hasMore"`
}

type UniquePeopleMetrics struct {
	PendingUsers          int64      `json:"pendingUsers"`
	OverdueUsers          int64      `json:"overdueUsers"`
	ActiveLeases          int64      `json:"activeLeases"`
	OldestDirtySince      *time.Time `json:"oldestDirtySince,omitempty"`
	OldestDirtyAgeSeconds int64      `json:"oldestDirtyAgeSeconds"`
	SuccessfulRuns        int64      `json:"successfulRuns"`
	RetryRuns             int64      `json:"retryRuns"`
	TimeoutRuns           int64      `json:"timeoutRuns"`
	TotalDurationMillis   int64      `json:"totalDurationMillis"`
}

type UniquePeopleCalculation struct {
	Generation uint64
	Count      int64
	Duration   time.Duration
}

type UniquePeopleStore interface {
	Claim(context.Context, time.Time, int, time.Duration, UniquePeopleConfig) ([]UniquePeopleClaim, error)
	CompletedResult(context.Context, UniquePeopleClaim) (*UniquePeopleExecution, error)
	Calculate(context.Context, UniquePeopleClaim, int) (UniquePeopleCalculation, error)
	Complete(context.Context, UniquePeopleClaim, UniquePeopleCalculation, func() time.Time, UniquePeopleConfig) (UniquePeopleExecution, error)
	Retry(context.Context, UniquePeopleClaim, time.Time, bool, UniquePeopleConfig) (UniquePeopleExecution, error)
	Backfill(context.Context, uint, int, time.Time, time.Duration) (UniquePeopleBackfillResult, error)
	Metrics(context.Context, time.Time, time.Duration) (UniquePeopleMetrics, error)
}

type UniquePeopleService struct {
	store  UniquePeopleStore
	config UniquePeopleConfig
	clock  func() time.Time
}

func NewUniquePeopleService(store UniquePeopleStore, config UniquePeopleConfig, clock func() time.Time) *UniquePeopleService {
	if clock == nil {
		clock = time.Now
	}
	return &UniquePeopleService{store: store, config: config, clock: clock}
}

func (s *UniquePeopleService) Claim(ctx context.Context, limit int) (UniquePeopleClaims, error) {
	if err := s.config.Validate(); err != nil {
		return UniquePeopleClaims{}, err
	}
	if limit <= 0 {
		return UniquePeopleClaims{}, ErrInvalidInput
	}
	limit = min(limit, s.config.ClaimBatch, s.config.Concurrency)
	jobs, err := s.store.Claim(ctx, s.clock().UTC(), limit, s.config.LeaseDuration, s.config)
	if err != nil {
		return UniquePeopleClaims{}, err
	}
	return UniquePeopleClaims{SchemaVersion: UniquePeopleSchemaVersion, Jobs: jobs}, nil
}

func (s *UniquePeopleService) Execute(ctx context.Context, claim UniquePeopleClaim) (UniquePeopleExecution, error) {
	if err := s.config.Validate(); err != nil {
		return UniquePeopleExecution{}, err
	}
	if claim.UserID == 0 || claim.RequestedGeneration == 0 || claim.LeaseToken == "" {
		return UniquePeopleExecution{}, ErrInvalidInput
	}
	if completed, err := s.store.CompletedResult(ctx, claim); err != nil {
		return UniquePeopleExecution{}, err
	} else if completed != nil {
		return *completed, nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, s.config.QueryTimeout)
	calculation, err := s.store.Calculate(queryCtx, claim, s.config.Concurrency)
	queryContextErr := queryCtx.Err()
	cancel()
	if errors.Is(err, ErrUniquePeopleStaleLease) {
		return staleUniquePeopleExecution(claim), nil
	}
	if errors.Is(err, ErrUniquePeopleInProgress) {
		return UniquePeopleExecution{}, err
	}
	if err != nil {
		if ctx.Err() != nil {
			return UniquePeopleExecution{}, ctx.Err()
		}
		now := s.clock().UTC()
		timedOut := errors.Is(err, context.DeadlineExceeded) || errors.Is(queryContextErr, context.DeadlineExceeded)
		return s.store.Retry(ctx, claim, now, timedOut, s.config)
	}
	return s.store.Complete(ctx, claim, calculation, s.clock, s.config)
}

func staleUniquePeopleExecution(claim UniquePeopleClaim) UniquePeopleExecution {
	return UniquePeopleExecution{SchemaVersion: UniquePeopleSchemaVersion, UserID: claim.UserID, AppliedGeneration: claim.RequestedGeneration, Outcome: UniquePeopleOutcomeStaleLease}
}

func (s *UniquePeopleService) Backfill(ctx context.Context, afterUserID uint, limit int) (UniquePeopleBackfillResult, error) {
	if err := s.config.Validate(); err != nil {
		return UniquePeopleBackfillResult{}, err
	}
	if limit <= 0 || limit > 500 {
		return UniquePeopleBackfillResult{}, ErrInvalidInput
	}
	return s.store.Backfill(ctx, afterUserID, limit, s.clock().UTC(), s.config.InitialSpread)
}

func (s *UniquePeopleService) Metrics(ctx context.Context) (UniquePeopleMetrics, error) {
	if err := s.config.Validate(); err != nil {
		return UniquePeopleMetrics{}, err
	}
	return s.store.Metrics(ctx, s.clock().UTC(), s.config.TargetWindow)
}
