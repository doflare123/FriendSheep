package statistics

import (
	"context"
	"testing"
	"time"
)

type uniquePeopleStoreStub struct {
	claimedLimit int
	claimedNow   time.Time
	claimJobs    []UniquePeopleClaim

	calculation  UniquePeopleCalculation
	calculateCtx context.Context
	completed    bool
	completedAt  time.Time
	completedJob UniquePeopleClaim

	backfillAfter uint
	backfillLimit int
	backfillNow   time.Time
	metricsNow    time.Time
}

func (s *uniquePeopleStoreStub) Claim(_ context.Context, now time.Time, limit int, _ time.Duration, _ UniquePeopleConfig) ([]UniquePeopleClaim, error) {
	s.claimedNow, s.claimedLimit = now, limit
	return s.claimJobs, nil
}

func (s *uniquePeopleStoreStub) CompletedResult(context.Context, UniquePeopleClaim) (*UniquePeopleExecution, error) {
	return nil, nil
}

func (s *uniquePeopleStoreStub) Calculate(ctx context.Context, _ UniquePeopleClaim, _ int) (UniquePeopleCalculation, error) {
	s.calculateCtx = ctx
	return s.calculation, nil
}

func (s *uniquePeopleStoreStub) Complete(_ context.Context, claim UniquePeopleClaim, calculation UniquePeopleCalculation, clock func() time.Time, _ UniquePeopleConfig) (UniquePeopleExecution, error) {
	now := clock().UTC()
	s.completed = true
	s.completedJob = claim
	s.completedAt = now
	count := calculation.Count
	return UniquePeopleExecution{SchemaVersion: UniquePeopleSchemaVersion, UserID: claim.UserID, AppliedGeneration: calculation.Generation, Outcome: UniquePeopleOutcomeApplied, Count: &count, CalculatedAt: &now}, nil
}

func (s *uniquePeopleStoreStub) Retry(context.Context, UniquePeopleClaim, time.Time, bool, UniquePeopleConfig) (UniquePeopleExecution, error) {
	return UniquePeopleExecution{}, nil
}

func (s *uniquePeopleStoreStub) Backfill(_ context.Context, after uint, limit int, now time.Time, _ time.Duration) (UniquePeopleBackfillResult, error) {
	s.backfillAfter, s.backfillLimit, s.backfillNow = after, limit, now
	return UniquePeopleBackfillResult{NextAfterUserID: after + 1}, nil
}

func (s *uniquePeopleStoreStub) Metrics(_ context.Context, now time.Time, _ time.Duration) (UniquePeopleMetrics, error) {
	s.metricsNow = now
	return UniquePeopleMetrics{}, nil
}

func TestUniquePeopleClaimBoundsByBatchAndDeploymentCapacity(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	store := &uniquePeopleStoreStub{claimJobs: []UniquePeopleClaim{{UserID: 7, RequestedGeneration: 3}}}
	config := DefaultUniquePeopleConfig()
	config.ClaimBatch, config.Concurrency = 8, 2
	service := NewUniquePeopleService(store, config, func() time.Time { return now })

	claims, err := service.Claim(context.Background(), 99)
	if err != nil || store.claimedLimit != 2 || !store.claimedNow.Equal(now) {
		t.Fatalf("claims=%+v limit=%d now=%s err=%v", claims, store.claimedLimit, store.claimedNow, err)
	}
	if claims.SchemaVersion != UniquePeopleSchemaVersion || len(claims.Jobs) != 1 {
		t.Fatalf("claims=%+v", claims)
	}
}

func TestUniquePeopleExecuteUsesBoundedCalculationAndPersistsItsGeneration(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	claim := UniquePeopleClaim{UserID: 7, RequestedGeneration: 3, LeaseToken: "lease"}
	store := &uniquePeopleStoreStub{calculation: UniquePeopleCalculation{Generation: 3, Count: 11, Duration: 15 * time.Millisecond}}
	config := DefaultUniquePeopleConfig()
	config.QueryTimeout = time.Second
	service := NewUniquePeopleService(store, config, func() time.Time { return now })

	result, err := service.Execute(context.Background(), claim)
	if err != nil || !store.completed || store.completedJob != claim || !store.completedAt.Equal(now) {
		t.Fatalf("result=%+v completed=%t job=%+v at=%s err=%v", result, store.completed, store.completedJob, store.completedAt, err)
	}
	if result.Outcome != UniquePeopleOutcomeApplied || result.Count == nil || *result.Count != 11 || result.AppliedGeneration != 3 {
		t.Fatalf("result=%+v", result)
	}
	if deadline, ok := store.calculateCtx.Deadline(); !ok || time.Until(deadline) <= 0 || time.Until(deadline) > config.QueryTimeout+time.Millisecond {
		t.Fatalf("calculation context deadline=%s present=%t", deadline, ok)
	}
}

func TestUniquePeopleBackfillAndMetricsUseInjectedClock(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	store := &uniquePeopleStoreStub{}
	service := NewUniquePeopleService(store, DefaultUniquePeopleConfig(), func() time.Time { return now })

	if result, err := service.Backfill(context.Background(), 12, 5); err != nil || result.NextAfterUserID != 13 || store.backfillAfter != 12 || store.backfillLimit != 5 || !store.backfillNow.Equal(now) {
		t.Fatalf("backfill=%+v store=%+v err=%v", result, store, err)
	}
	if _, err := service.Metrics(context.Background()); err != nil || !store.metricsNow.Equal(now) {
		t.Fatalf("metrics now=%s err=%v", store.metricsNow, err)
	}
}
