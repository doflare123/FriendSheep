package tests

import (
	"context"
	"testing"
	"time"

	statmodels "friendship/models/statistics"
	stats "friendship/services/statistics"
)

func TestP04AUniquePeopleLeaseRecoveryStaleAckAndCommitReplayPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, aggregate := statisticsFixture(t, db)
	eventID := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{2, 3}, nil)
	if _, err := aggregate.Process(context.Background(), eventID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(19 * time.Hour)
	config := stats.DefaultUniquePeopleConfig()
	config.ClaimBatch, config.Concurrency = 1, 1
	serviceAt := func(at time.Time) *stats.UniquePeopleService {
		return stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(&testPostgresRepository{db: db}), config, func() time.Time { return at })
	}
	first, err := serviceAt(now).Claim(context.Background(), 1)
	if err != nil || len(first.Jobs) != 1 {
		t.Fatalf("first claim=%+v err=%v", first, err)
	}
	if stale, err := serviceAt(now.Add(config.LeaseDuration)).Execute(context.Background(), first.Jobs[0]); err != nil || stale.Outcome != stats.UniquePeopleOutcomeStaleLease {
		t.Fatalf("expired execute=%+v err=%v", stale, err)
	}
	recoveredAt := now.Add(config.LeaseDuration + config.RetryBase + time.Second)
	recovered, err := serviceAt(recoveredAt).Claim(context.Background(), 1)
	if err != nil || len(recovered.Jobs) != 1 || recovered.Jobs[0].LeaseToken == first.Jobs[0].LeaseToken {
		t.Fatalf("recovered=%+v err=%v", recovered, err)
	}
	result, err := serviceAt(recoveredAt).Execute(context.Background(), recovered.Jobs[0])
	if err != nil || result.Outcome != stats.UniquePeopleOutcomeApplied || result.Count == nil || *result.Count != 1 {
		t.Fatalf("recovered execute=%+v err=%v", result, err)
	}
	replay, err := serviceAt(recoveredAt).Execute(context.Background(), recovered.Jobs[0])
	if err != nil || replay.Outcome != stats.UniquePeopleOutcomeAlreadyApplied || replay.Count == nil || *replay.Count != 1 {
		t.Fatalf("commit-loss replay=%+v err=%v", replay, err)
	}
	var job statmodels.UniquePeopleJob
	if err := db.First(&job, "user_id = ?", 2).Error; err != nil || job.SuccessfulRuns != 1 || job.TimeoutRuns != 0 {
		t.Fatalf("job=%+v err=%v", job, err)
	}
}

func TestP04AUniquePeopleBackfillAndTwentyFourHourMetricsPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, aggregate := statisticsFixture(t, db)
	for _, eventID := range []uint{
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{2}, nil),
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{3}, nil),
	} {
		if _, err := aggregate.Process(context.Background(), eventID); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Add(25 * time.Hour)
	config := stats.DefaultUniquePeopleConfig()
	service := stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(&testPostgresRepository{db: db}), config, func() time.Time { return now })
	metrics, err := service.Metrics(context.Background())
	if err != nil || metrics.PendingUsers < 2 || metrics.OverdueUsers < 2 || metrics.OldestDirtySince == nil || metrics.OldestDirtyAgeSeconds < int64((24*time.Hour).Seconds()) {
		t.Fatalf("metrics=%+v err=%v", metrics, err)
	}
	backfill, err := service.Backfill(context.Background(), 0, 1)
	if err != nil || !backfill.HasMore || backfill.NextAfterUserID == 0 {
		t.Fatalf("backfill=%+v err=%v", backfill, err)
	}
}
