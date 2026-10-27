package tests

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	statmodels "friendship/models/statistics"
	stats "friendship/services/statistics"

	"gorm.io/gorm"
)

func uniquePeopleServiceAt(db *gorm.DB, config stats.UniquePeopleConfig, now time.Time) *stats.UniquePeopleService {
	return stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(&testPostgresRepository{db: db}), config, func() time.Time { return now })
}

func TestP04AUniquePeopleLeaseRecoveryStaleAckAndCommitReplayPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, aggregate := statisticsFixture(t, db)
	eventID := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{2, 3}, nil)
	if _, err := aggregate.Process(context.Background(), eventID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(19 * time.Hour)
	// Изолируем исходную работу: второй участник не должен подменить её при recovery.
	if err := db.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", 3).Update("next_attempt_at", now.Add(48*time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	config := stats.DefaultUniquePeopleConfig()
	config.ClaimBatch, config.Concurrency = 1, 1
	serviceAt := func(at time.Time) *stats.UniquePeopleService {
		return stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(&testPostgresRepository{db: db}), config, func() time.Time { return at })
	}
	first, err := serviceAt(now).Claim(context.Background(), 1)
	if err != nil || len(first.Jobs) != 1 || first.Jobs[0].UserID != 2 {
		t.Fatalf("first claim=%+v err=%v", first, err)
	}
	if stale, err := serviceAt(now.Add(config.LeaseDuration)).Execute(context.Background(), first.Jobs[0]); err != nil || stale.Outcome != stats.UniquePeopleOutcomeStaleLease {
		t.Fatalf("expired execute=%+v err=%v", stale, err)
	}
	expiredAt := now.Add(config.LeaseDuration)
	expiryClaims, err := serviceAt(expiredAt).Claim(context.Background(), 1)
	if err != nil || len(expiryClaims.Jobs) != 0 {
		t.Fatalf("expiry must schedule retry, claims=%+v err=%v", expiryClaims, err)
	}
	var expiredJob statmodels.UniquePeopleJob
	if err := db.First(&expiredJob, "user_id = ?", first.Jobs[0].UserID).Error; err != nil || expiredJob.LeaseToken != nil || !expiredJob.NextAttemptAt.After(expiredAt) {
		t.Fatalf("expired job=%+v err=%v", expiredJob, err)
	}
	recoveredAt := expiredJob.NextAttemptAt
	recovered, err := serviceAt(recoveredAt).Claim(context.Background(), 1)
	if err != nil || len(recovered.Jobs) != 1 || recovered.Jobs[0].UserID != first.Jobs[0].UserID || recovered.Jobs[0].LeaseToken == first.Jobs[0].LeaseToken {
		t.Fatalf("recovered=%+v err=%v", recovered, err)
	}
	if stale, err := serviceAt(recoveredAt).Execute(context.Background(), first.Jobs[0]); err != nil || stale.Outcome != stats.UniquePeopleOutcomeStaleLease {
		t.Fatalf("old token after new lease=%+v err=%v", stale, err)
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

func TestP04AUniquePeopleExpiredLeaseSchedulesRetryBeforeRecoveryPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	migrateParticipationStatistics(t, db)
	base := time.Now().UTC().Add(time.Hour)
	for _, id := range []uint{2, 3} {
		if err := db.Create(&statmodels.UserAggregate{UserID: id, UpdatedAt: base}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&statmodels.UniquePeopleJob{UserID: 2, RequestedGeneration: 1, DirtySince: &base, NextAttemptAt: base}).Error; err != nil {
		t.Fatal(err)
	}
	// Третья строка не должна случайно заменить проверяемого пользователя в claim.
	far := base.Add(48 * time.Hour)
	if err := db.Create(&statmodels.UniquePeopleJob{UserID: 3, RequestedGeneration: 1, DirtySince: &base, NextAttemptAt: far}).Error; err != nil {
		t.Fatal(err)
	}
	config := stats.DefaultUniquePeopleConfig()
	config.ClaimBatch, config.Concurrency = 1, 1
	first, err := uniquePeopleServiceAt(db, config, base).Claim(context.Background(), 1)
	if err != nil || len(first.Jobs) != 1 || first.Jobs[0].UserID != 2 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	expiredAt := base.Add(config.LeaseDuration)
	second, err := uniquePeopleServiceAt(db, config, expiredAt).Claim(context.Background(), 1)
	if err != nil || len(second.Jobs) != 0 {
		t.Fatalf("expiry must only schedule retry: %+v/%v", second, err)
	}
	var job statmodels.UniquePeopleJob
	if err := db.First(&job, "user_id = ?", 2).Error; err != nil || job.LeaseToken != nil || !job.NextAttemptAt.After(expiredAt) || job.RetryRuns != 1 {
		t.Fatalf("expired job=%+v err=%v", job, err)
	}
	recovered, err := uniquePeopleServiceAt(db, config, job.NextAttemptAt).Claim(context.Background(), 1)
	if err != nil || len(recovered.Jobs) != 1 || recovered.Jobs[0].UserID != 2 || recovered.Jobs[0].LeaseToken == first.Jobs[0].LeaseToken {
		t.Fatalf("recovered=%+v err=%v", recovered, err)
	}
	stale, err := uniquePeopleServiceAt(db, config, job.NextAttemptAt).Execute(context.Background(), first.Jobs[0])
	if err != nil || stale.Outcome != stats.UniquePeopleOutcomeStaleLease {
		t.Fatalf("stale old lease=%+v err=%v", stale, err)
	}
}

func TestP04AUniquePeopleTimeoutPersistsBackoffPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, aggregate := statisticsFixture(t, db)
	eventID := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{2}, nil)
	if _, err := aggregate.Process(context.Background(), eventID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(19 * time.Hour)
	config := stats.DefaultUniquePeopleConfig()
	config.QueryTimeout, config.LeaseDuration = 50*time.Millisecond, time.Minute
	service := uniquePeopleServiceAt(db, config, now)
	claims, err := service.Claim(context.Background(), 1)
	if err != nil || len(claims.Jobs) != 1 {
		t.Fatalf("claim=%+v err=%v", claims, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	lock, err := sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.Exec(`LOCK TABLE events_users IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatal(err)
	}
	result, err := service.Execute(context.Background(), claims.Jobs[0])
	if err != nil || result.Outcome != stats.UniquePeopleOutcomeRetryScheduled {
		t.Fatalf("timeout result=%+v err=%v", result, err)
	}
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	var job statmodels.UniquePeopleJob
	if err := db.First(&job, "user_id = ?", 2).Error; err != nil || job.TimeoutRuns != 1 || job.RetryRuns != 1 || job.LeaseToken != nil || !job.NextAttemptAt.After(now) {
		t.Fatalf("timeout job=%+v err=%v", job, err)
	}
}

func TestP04AUniquePeopleGenerationFollowupDuringRepeatableReadCountPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, _ := statisticsFixture(t, db)
	repo := &testPostgresRepository{db: db}
	config := stats.DefaultUniquePeopleConfig()
	config.InitialSpread, config.ClaimBatch, config.Concurrency = time.Nanosecond, 1, 1
	aggregate := stats.NewService(stats.NewGORMStoreWithUniquePeopleConfig(repo, config))
	start := time.Now().UTC().Add(-2 * time.Hour)
	firstEvent := seedStatisticsEvent(t, db, groupID, 1, 1, start, 30, []uint{2, 3}, nil)
	if _, err := aggregate.Process(context.Background(), firstEvent); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", 3).Update("next_attempt_at", time.Now().UTC().Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	unique := uniquePeopleServiceAt(db, config, now)
	claims, err := unique.Claim(context.Background(), 1)
	if err != nil || len(claims.Jobs) != 1 || claims.Jobs[0].UserID != 2 {
		t.Fatalf("claim=%+v err=%v", claims, err)
	}
	countReached := make(chan struct{})
	releaseCount := make(chan struct{})
	callbackName := "p04a_block_unique_count"
	callback := func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "COUNT(DISTINCT other.user_id)") {
			select {
			case <-countReached:
			default:
				close(countReached)
			}
			<-releaseCount
		}
	}
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, callback); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Row().Before("gorm:row").Register(callbackName, callback); err != nil {
		t.Fatal(err)
	}
	defer db.Callback().Query().Remove(callbackName)
	defer db.Callback().Row().Remove(callbackName)
	defer func() {
		select {
		case <-releaseCount:
		default:
			close(releaseCount)
		}
	}()
	done := make(chan struct {
		result stats.UniquePeopleExecution
		err    error
	}, 1)
	go func() {
		result, executeErr := unique.Execute(context.Background(), claims.Jobs[0])
		done <- struct {
			result stats.UniquePeopleExecution
			err    error
		}{result, executeErr}
	}()
	select {
	case <-countReached:
	case <-time.After(5 * time.Second):
		t.Fatal("RR calculation did not reach COUNT barrier")
	}
	secondEvent := seedStatisticsEvent(t, db, groupID, 1, 1, start, 30, []uint{2, 4}, nil)
	if _, err := aggregate.Process(context.Background(), secondEvent); err != nil {
		t.Fatal(err)
	}
	ordinaryView, err := aggregate.Read(context.Background(), 2)
	if err != nil || ordinaryView.LastUpdatedAt == nil {
		t.Fatalf("ordinary updated timestamp=%+v err=%v", ordinaryView, err)
	}
	if err := db.Model(&statmodels.UniquePeopleJob{}).Where("user_id IN ?", []uint{3, 4}).Update("next_attempt_at", time.Now().UTC().Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	close(releaseCount)
	executed := <-done
	if executed.err != nil || executed.result.Outcome != stats.UniquePeopleOutcomeApplied || executed.result.AppliedGeneration != 1 || executed.result.Count == nil || *executed.result.Count != 1 {
		t.Fatalf("first execution=%+v err=%v", executed.result, executed.err)
	}
	view, err := aggregate.Read(context.Background(), 2)
	if err != nil || view.UniquePeopleCount == nil || *view.UniquePeopleCount != 1 || view.UniquePeopleStatus != stats.UniquePeopleStatusPending {
		t.Fatalf("G+1 view=%+v err=%v", view, err)
	}
	if view.LastUpdatedAt == nil || !view.LastUpdatedAt.Equal(*ordinaryView.LastUpdatedAt) {
		t.Fatalf("unique save changed ordinary updatedAt: before=%v after=%v", ordinaryView.LastUpdatedAt, view.LastUpdatedAt)
	}
	var job statmodels.UniquePeopleJob
	if err := db.First(&job, "user_id = ?", 2).Error; err != nil || job.RequestedGeneration != 2 || job.AppliedGeneration != 1 || job.LeaseToken != nil {
		t.Fatalf("G+1 job=%+v err=%v", job, err)
	}
	nextClaims, err := uniquePeopleServiceAt(db, config, time.Now().UTC()).Claim(context.Background(), 1)
	if err != nil || len(nextClaims.Jobs) != 1 || nextClaims.Jobs[0].RequestedGeneration != 2 {
		t.Fatalf("followup claim=%+v err=%v", nextClaims, err)
	}
	final, err := uniquePeopleServiceAt(db, config, time.Now().UTC()).Execute(context.Background(), nextClaims.Jobs[0])
	if err != nil || final.Outcome != stats.UniquePeopleOutcomeApplied || final.Count == nil || *final.Count != 2 {
		t.Fatalf("followup execute=%+v err=%v", final, err)
	}
}

func TestP04AUniquePeopleSQLSemaphoreSurvivesExpiredLeasePostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, _ := statisticsFixture(t, db)
	repo := &testPostgresRepository{db: db}
	config := stats.DefaultUniquePeopleConfig()
	config.InitialSpread, config.ClaimBatch, config.Concurrency = time.Nanosecond, 2, 1
	aggregate := stats.NewService(stats.NewGORMStoreWithUniquePeopleConfig(repo, config))
	eventID := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{2, 3}, nil)
	if _, err := aggregate.Process(context.Background(), eventID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	firstService := uniquePeopleServiceAt(db, config, now)
	firstClaims, err := firstService.Claim(context.Background(), 1)
	if err != nil || len(firstClaims.Jobs) != 1 || firstClaims.Jobs[0].UserID != 2 {
		t.Fatalf("first claim=%+v err=%v", firstClaims, err)
	}
	countReached := make(chan struct{})
	releaseCount := make(chan struct{})
	callbackName := "p04a_hold_semaphore_count"
	callback := func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "COUNT(DISTINCT other.user_id)") {
			select {
			case <-countReached:
			default:
				close(countReached)
			}
			<-releaseCount
		}
	}
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, callback); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Row().Before("gorm:row").Register(callbackName, callback); err != nil {
		t.Fatal(err)
	}
	defer db.Callback().Query().Remove(callbackName)
	defer db.Callback().Row().Remove(callbackName)
	defer func() {
		select {
		case <-releaseCount:
		default:
			close(releaseCount)
		}
	}()
	done := make(chan struct {
		result stats.UniquePeopleExecution
		err    error
	}, 1)
	go func() {
		result, executeErr := firstService.Execute(context.Background(), firstClaims.Jobs[0])
		done <- struct {
			result stats.UniquePeopleExecution
			err    error
		}{result, executeErr}
	}()
	select {
	case <-countReached:
	case <-time.After(5 * time.Second):
		t.Fatal("first calculation did not acquire SQL semaphore")
	}
	expiredAt := now.Add(config.LeaseDuration + time.Second)
	secondClaims, err := uniquePeopleServiceAt(db, config, expiredAt).Claim(context.Background(), 2)
	if err != nil || len(secondClaims.Jobs) != 1 || secondClaims.Jobs[0].UserID != 3 {
		t.Fatalf("claim after expired lease=%+v err=%v", secondClaims, err)
	}
	store := stats.NewGORMUniquePeopleStore(repo)
	if _, err := store.Calculate(context.Background(), secondClaims.Jobs[0], config.Concurrency, expiredAt); !errors.Is(err, stats.ErrUniquePeopleCapacity) {
		t.Fatalf("second calculation bypassed global semaphore: %v", err)
	}
	close(releaseCount)
	completed := <-done
	if completed.err != nil || completed.result.Outcome != stats.UniquePeopleOutcomeStaleLease {
		t.Fatalf("expired worker committed result=%+v err=%v", completed.result, completed.err)
	}
	var abandoned statmodels.UniquePeopleJob
	if err := db.First(&abandoned, "user_id = ?", 2).Error; err != nil || abandoned.AppliedGeneration != 0 || abandoned.SuccessfulRuns != 0 {
		t.Fatalf("stale worker changed job=%+v err=%v", abandoned, err)
	}
	// После освобождения SQL slot другой уже выданный lease действительно рассчитывается.
	calculation, err := store.Calculate(context.Background(), secondClaims.Jobs[0], config.Concurrency, expiredAt)
	if err != nil || calculation.Count != 1 {
		t.Fatalf("slot was not released: %+v err=%v", calculation, err)
	}
	if result, err := store.Complete(context.Background(), secondClaims.Jobs[0], calculation, func() time.Time { return expiredAt }, config); err != nil || result.Outcome != stats.UniquePeopleOutcomeApplied {
		t.Fatalf("second complete=%+v err=%v", result, err)
	}
	recovered, err := uniquePeopleServiceAt(db, config, abandoned.NextAttemptAt).Claim(context.Background(), 1)
	if err != nil || len(recovered.Jobs) != 1 || recovered.Jobs[0].UserID != 2 || recovered.Jobs[0].LeaseToken == firstClaims.Jobs[0].LeaseToken {
		t.Fatalf("original recovery=%+v err=%v", recovered, err)
	}
	if result, err := uniquePeopleServiceAt(db, config, abandoned.NextAttemptAt).Execute(context.Background(), recovered.Jobs[0]); err != nil || result.Outcome != stats.UniquePeopleOutcomeApplied || result.Count == nil || *result.Count != 1 {
		t.Fatalf("original recovery execute=%+v err=%v", result, err)
	}
}

func TestP04AUniquePeopleDeadlineAndCoalescingPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, ordinary := statisticsFixture(t, db)
	start := time.Now().UTC().Add(-3 * time.Hour)
	firstEvent := seedStatisticsEvent(t, db, groupID, 1, 1, start, 30, []uint{2}, nil)
	if _, err := ordinary.Process(context.Background(), firstEvent); err != nil {
		t.Fatal(err)
	}
	var before statmodels.UniquePeopleJob
	if err := db.First(&before, "user_id = ?", 2).Error; err != nil || before.DirtySince == nil {
		t.Fatalf("first job=%+v err=%v", before, err)
	}
	for range 4 {
		eventID := seedStatisticsEvent(t, db, groupID, 1, 1, start, 30, []uint{2}, nil)
		if _, err := ordinary.Process(context.Background(), eventID); err != nil {
			t.Fatal(err)
		}
	}
	var after statmodels.UniquePeopleJob
	if err := db.First(&after, "user_id = ?", 2).Error; err != nil || after.RequestedGeneration != 5 || after.AppliedGeneration != 0 || after.DirtySince == nil || !after.DirtySince.Equal(*before.DirtySince) || !after.NextAttemptAt.Equal(before.NextAttemptAt) {
		t.Fatalf("coalescing moved initial deadline: before=%+v after=%+v err=%v", before, after, err)
	}
	config := stats.DefaultUniquePeopleConfig()
	deadline := before.DirtySince.Add(config.TargetWindow)
	for _, boundary := range []struct {
		name    string
		at      time.Time
		overdue int64
	}{
		{"before", deadline.Add(-time.Microsecond), 0}, {"exact", deadline, 1}, {"after", deadline.Add(time.Microsecond), 1},
	} {
		t.Run(boundary.name, func(t *testing.T) {
			metrics, err := uniquePeopleServiceAt(db, config, boundary.at).Metrics(context.Background())
			if err != nil || metrics.PendingUsers != 1 || metrics.OverdueUsers != boundary.overdue || metrics.OldestDirtySince == nil || !metrics.OldestDirtySince.Equal(*before.DirtySince) {
				t.Fatalf("boundary=%s metrics=%+v err=%v", boundary.name, metrics, err)
			}
		})
	}
}
