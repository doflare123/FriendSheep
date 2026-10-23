package statistics

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"time"

	statmodels "friendship/models/statistics"
	"friendship/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	uniquePeopleAdvisoryNamespace = 1718379631
	uniquePeopleClaimMutexKey     = 1
	uniquePeopleUserLockKey       = 2
	uniquePeopleSemaphoreKey      = 3
)

type gormUniquePeopleStore struct {
	repo repository.PostgresRepository
}

func NewGORMUniquePeopleStore(repo repository.PostgresRepository) UniquePeopleStore {
	return &gormUniquePeopleStore{repo: repo}
}

func (s *gormUniquePeopleStore) Claim(ctx context.Context, now time.Time, limit int, leaseDuration time.Duration, config UniquePeopleConfig) ([]UniquePeopleClaim, error) {
	db := s.repo.Model(nil).WithContext(ctx)
	claims := make([]UniquePeopleClaim, 0, limit)
	err := db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", uniquePeopleAdvisoryNamespace, uniquePeopleClaimMutexKey).Error; err != nil {
				return err
			}
		}

		var expired []statmodels.UniquePeopleJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("lease_token IS NOT NULL AND lease_expires_at <= ?", now).
			Order("lease_expires_at ASC, user_id ASC").Limit(limit).Find(&expired).Error; err != nil {
			return err
		}
		for i := range expired {
			next := now.Add(uniquePeopleRetryDelayFor(expired[i].UserID, expired[i].RequestedGeneration, expired[i].AttemptCount, config))
			if err := tx.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", expired[i].UserID).Updates(map[string]any{
				"lease_token": nil, "lease_expires_at": nil, "next_attempt_at": next,
				"retry_runs": gorm.Expr("retry_runs + 1"),
			}).Error; err != nil {
				return err
			}
		}
		var activeLeases int64
		if err := tx.Model(&statmodels.UniquePeopleJob{}).Where("lease_token IS NOT NULL AND lease_expires_at > ?", now).Count(&activeLeases).Error; err != nil {
			return err
		}
		available := config.Concurrency - int(activeLeases)
		if available <= 0 {
			return nil
		}
		limit = min(limit, available)

		var jobs []statmodels.UniquePeopleJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("requested_generation > applied_generation AND next_attempt_at <= ? AND lease_token IS NULL", now).
			Order("dirty_since ASC, next_attempt_at ASC, user_id ASC").Limit(limit).Find(&jobs).Error; err != nil {
			return err
		}
		for i := range jobs {
			token, err := newUniquePeopleLeaseToken()
			if err != nil {
				return err
			}
			expires := now.Add(leaseDuration)
			attempt := jobs[i].AttemptCount + 1
			if err := tx.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", jobs[i].UserID).Updates(map[string]any{
				"lease_token": token, "lease_expires_at": expires, "attempt_count": attempt,
				"lease_generation":     jobs[i].RequestedGeneration,
				"followup_dirty_since": nil,
			}).Error; err != nil {
				return err
			}
			dirtySince := now
			if jobs[i].DirtySince != nil {
				dirtySince = jobs[i].DirtySince.UTC()
			}
			claims = append(claims, UniquePeopleClaim{
				UserID: jobs[i].UserID, RequestedGeneration: jobs[i].RequestedGeneration,
				LeaseToken: token, DirtySince: dirtySince, DueAt: jobs[i].NextAttemptAt.UTC(),
				LeaseExpiresAt: expires,
			})
		}
		return nil
	})
	return claims, err
}

func (s *gormUniquePeopleStore) CompletedResult(ctx context.Context, claim UniquePeopleClaim) (*UniquePeopleExecution, error) {
	var job statmodels.UniquePeopleJob
	err := s.repo.Model(nil).WithContext(ctx).First(&job, "user_id = ?", claim.UserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if job.LastLeaseToken == nil || *job.LastLeaseToken != claim.LeaseToken {
		return nil, nil
	}
	return completedUniquePeopleExecution(job), nil
}

func (s *gormUniquePeopleStore) Calculate(ctx context.Context, claim UniquePeopleClaim, concurrency int) (UniquePeopleCalculation, error) {
	started := time.Now()
	db := s.repo.Model(nil).WithContext(ctx)
	options := &sql.TxOptions{}
	if db.Dialector.Name() == "postgres" {
		options.Isolation = sql.LevelRepeatableRead
	}
	var calculation UniquePeopleCalculation
	err := db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			locked, err := tryUniquePeopleAdvisoryLock(tx, uniquePeopleUserLockKey, advisoryUserKey(claim.UserID))
			if err != nil {
				return err
			}
			if !locked {
				return ErrUniquePeopleInProgress
			}
			capacity := false
			for slot := 0; slot < concurrency; slot++ {
				locked, err = tryUniquePeopleAdvisoryLock(tx, uniquePeopleSemaphoreKey, int32(slot))
				if err != nil {
					return err
				}
				if locked {
					capacity = true
					break
				}
			}
			if !capacity {
				return ErrUniquePeopleCapacity
			}
		}

		var job statmodels.UniquePeopleJob
		err := tx.First(&job, "user_id = ? AND lease_token = ? AND lease_generation = ? AND lease_expires_at > CURRENT_TIMESTAMP", claim.UserID, claim.LeaseToken, claim.RequestedGeneration).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUniquePeopleStaleLease
		}
		if err != nil {
			return err
		}
		calculation.Generation = job.RequestedGeneration
		return tx.Raw(`SELECT COUNT(DISTINCT other.user_id)
			FROM events_users mine
			JOIN participation_statistics_processed_events processed ON processed.event_id = mine.event_id
			JOIN events_users other ON other.event_id = mine.event_id
			WHERE mine.user_id = ? AND other.user_id <> ?`, claim.UserID, claim.UserID).
			Scan(&calculation.Count).Error
	}, options)
	calculation.Duration = time.Since(started)
	return calculation, err
}

func (s *gormUniquePeopleStore) Complete(ctx context.Context, claim UniquePeopleClaim, calculation UniquePeopleCalculation, clock func() time.Time, config UniquePeopleConfig) (UniquePeopleExecution, error) {
	db := s.repo.Model(nil).WithContext(ctx)
	result := staleUniquePeopleExecution(claim)
	err := db.Transaction(func(tx *gorm.DB) error {
		initialNow := clock().UTC()
		aggregate := statmodels.UserAggregate{UserID: claim.UserID, UpdatedAt: initialNow}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&aggregate).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&aggregate, "user_id = ?", claim.UserID).Error; err != nil {
			return err
		}
		var job statmodels.UniquePeopleJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, "user_id = ?", claim.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if job.LastLeaseToken != nil && *job.LastLeaseToken == claim.LeaseToken {
			result = *completedUniquePeopleExecution(job)
			return nil
		}
		now := clock().UTC()
		if job.LeaseToken == nil || *job.LeaseToken != claim.LeaseToken || job.LeaseGeneration != claim.RequestedGeneration || job.LeaseExpiresAt == nil || !job.LeaseExpiresAt.After(now) {
			return nil
		}
		if calculation.Generation < job.LeaseGeneration || calculation.Generation < job.AppliedGeneration || calculation.Generation > job.RequestedGeneration {
			return nil
		}
		if err := tx.Model(&statmodels.UserAggregate{}).Where("user_id = ?", claim.UserID).UpdateColumns(map[string]any{
			"unique_people_count": calculation.Count, "unique_people_calculated_at": now,
		}).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"applied_generation": calculation.Generation, "lease_token": nil, "lease_expires_at": nil,
			"attempt_count": 0, "last_lease_token": claim.LeaseToken,
			"last_result_generation": calculation.Generation, "last_result_count": calculation.Count,
			"last_result_calculated_at": now, "successful_runs": gorm.Expr("successful_runs + 1"),
			"total_duration_millis": gorm.Expr("total_duration_millis + ?", max(calculation.Duration.Milliseconds(), int64(0))),
			"followup_dirty_since":  nil,
		}
		if job.RequestedGeneration > calculation.Generation {
			dirty := job.DirtySince
			if job.FollowupDirtySince != nil {
				dirty = job.FollowupDirtySince
			}
			if dirty == nil {
				d := now
				dirty = &d
			}
			updates["dirty_since"] = *dirty
			updates["next_attempt_at"] = dirty.Add(uniquePeopleSpreadOffset(claim.UserID, job.RequestedGeneration, config.InitialSpread))
		} else {
			updates["dirty_since"] = nil
			updates["next_attempt_at"] = now
		}
		if err := tx.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", claim.UserID).Updates(updates).Error; err != nil {
			return err
		}
		count := calculation.Count
		calculatedAt := now
		result = UniquePeopleExecution{SchemaVersion: UniquePeopleSchemaVersion, UserID: claim.UserID, AppliedGeneration: calculation.Generation, Outcome: UniquePeopleOutcomeApplied, Count: &count, CalculatedAt: &calculatedAt}
		return nil
	})
	return result, err
}

func (s *gormUniquePeopleStore) Retry(ctx context.Context, claim UniquePeopleClaim, now time.Time, timedOut bool, config UniquePeopleConfig) (UniquePeopleExecution, error) {
	db := s.repo.Model(nil).WithContext(ctx)
	result := staleUniquePeopleExecution(claim)
	err := db.Transaction(func(tx *gorm.DB) error {
		var job statmodels.UniquePeopleJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, "user_id = ?", claim.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if job.LastLeaseToken != nil && *job.LastLeaseToken == claim.LeaseToken {
			result = *completedUniquePeopleExecution(job)
			return nil
		}
		if job.LeaseToken == nil || *job.LeaseToken != claim.LeaseToken || job.LeaseGeneration != claim.RequestedGeneration || job.LeaseExpiresAt == nil || !job.LeaseExpiresAt.After(now) {
			return nil
		}
		next := now.Add(uniquePeopleRetryDelayFor(job.UserID, job.RequestedGeneration, job.AttemptCount, config))
		updates := map[string]any{
			"lease_token": nil, "lease_expires_at": nil, "next_attempt_at": next,
			"retry_runs": gorm.Expr("retry_runs + 1"),
		}
		if timedOut {
			updates["timeout_runs"] = gorm.Expr("timeout_runs + 1")
		}
		if err := tx.Model(&statmodels.UniquePeopleJob{}).Where("user_id = ?", claim.UserID).Updates(updates).Error; err != nil {
			return err
		}
		result = UniquePeopleExecution{SchemaVersion: UniquePeopleSchemaVersion, UserID: claim.UserID, AppliedGeneration: claim.RequestedGeneration, Outcome: UniquePeopleOutcomeRetryScheduled}
		return nil
	})
	return result, err
}

func (s *gormUniquePeopleStore) Backfill(ctx context.Context, afterUserID uint, limit int, now time.Time, spread time.Duration) (UniquePeopleBackfillResult, error) {
	db := s.repo.Model(nil).WithContext(ctx)
	var userIDs []uint
	if err := db.Model(&statmodels.UserAggregate{}).Where("user_id > ?", afterUserID).Order("user_id ASC").Limit(limit+1).Pluck("user_id", &userIDs).Error; err != nil {
		return UniquePeopleBackfillResult{}, err
	}
	hasMore := len(userIDs) > limit
	if hasMore {
		userIDs = userIDs[:limit]
	}
	result := UniquePeopleBackfillResult{HasMore: hasMore, NextAfterUserID: afterUserID}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, userID := range userIDs {
			job := statmodels.UniquePeopleJob{
				UserID: userID, RequestedGeneration: 1, DirtySince: &now,
				NextAttemptAt: now.Add(uniquePeopleSpreadOffset(userID, 1, spread)),
			}
			created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&job)
			if created.Error != nil {
				return created.Error
			}
			result.Enqueued += created.RowsAffected
			result.NextAfterUserID = userID
		}
		return nil
	})
	return result, err
}

func (s *gormUniquePeopleStore) Metrics(ctx context.Context, now time.Time, targetWindow time.Duration) (UniquePeopleMetrics, error) {
	var row struct {
		PendingUsers        int64
		OverdueUsers        int64
		ActiveLeases        int64
		OldestDirtySince    *time.Time
		SuccessfulRuns      int64
		RetryRuns           int64
		TimeoutRuns         int64
		TotalDurationMillis int64
	}
	err := s.repo.Model(nil).WithContext(ctx).Raw(`SELECT
		COALESCE(SUM(CASE WHEN requested_generation > applied_generation THEN 1 ELSE 0 END),0) AS pending_users,
		COALESCE(SUM(CASE WHEN requested_generation > applied_generation AND dirty_since <= ? THEN 1 ELSE 0 END),0) AS overdue_users,
		COALESCE(SUM(CASE WHEN lease_token IS NOT NULL AND lease_expires_at > ? THEN 1 ELSE 0 END),0) AS active_leases,
		MIN(CASE WHEN requested_generation > applied_generation THEN dirty_since END) AS oldest_dirty_since,
		COALESCE(SUM(successful_runs),0) AS successful_runs,
		COALESCE(SUM(retry_runs),0) AS retry_runs,
		COALESCE(SUM(timeout_runs),0) AS timeout_runs,
		COALESCE(SUM(total_duration_millis),0) AS total_duration_millis
		FROM participation_statistics_unique_people_jobs`, now.Add(-targetWindow), now).Scan(&row).Error
	if err != nil {
		return UniquePeopleMetrics{}, err
	}
	result := UniquePeopleMetrics{
		PendingUsers: row.PendingUsers, OverdueUsers: row.OverdueUsers, ActiveLeases: row.ActiveLeases,
		OldestDirtySince: row.OldestDirtySince, SuccessfulRuns: row.SuccessfulRuns,
		RetryRuns: row.RetryRuns, TimeoutRuns: row.TimeoutRuns, TotalDurationMillis: row.TotalDurationMillis,
	}
	if row.OldestDirtySince != nil && now.After(*row.OldestDirtySince) {
		result.OldestDirtyAgeSeconds = int64(now.Sub(*row.OldestDirtySince).Seconds())
	}
	return result, nil
}

func completedUniquePeopleExecution(job statmodels.UniquePeopleJob) *UniquePeopleExecution {
	return &UniquePeopleExecution{
		SchemaVersion: UniquePeopleSchemaVersion, UserID: job.UserID,
		AppliedGeneration: job.LastResultGeneration, Outcome: UniquePeopleOutcomeAlreadyApplied,
		Count: job.LastResultCount, CalculatedAt: job.LastResultCalculatedAt,
	}
}

func tryUniquePeopleAdvisoryLock(tx *gorm.DB, kind, key int32) (bool, error) {
	var locked bool
	err := tx.Raw("SELECT pg_try_advisory_xact_lock(?, ?)", uniquePeopleAdvisoryNamespace+kind, key).Scan(&locked).Error
	return locked, err
}

func advisoryUserKey(userID uint) int32 {
	return int32(uint64(userID) ^ (uint64(userID) >> 32))
}

func newUniquePeopleLeaseToken() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("создать lease token: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func uniquePeopleSpreadOffset(userID uint, generation uint64, spread time.Duration) time.Duration {
	if spread <= 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d", userID, generation)))
	return time.Duration(h.Sum64() % uint64(spread))
}

func uniquePeopleRetryDelayFor(userID uint, generation, attempt uint64, config UniquePeopleConfig) time.Duration {
	if attempt == 0 {
		attempt = 1
	}
	exponent := min(attempt-1, uint64(30))
	base := float64(config.RetryBase) * math.Pow(2, float64(exponent))
	if base > float64(config.RetryMax) {
		base = float64(config.RetryMax)
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d:%d", userID, generation, attempt)))
	fraction := float64(h.Sum64()%10001) / 10000
	factor := 1 - config.RetryJitter + 2*config.RetryJitter*fraction
	delay := time.Duration(base * factor)
	if delay > config.RetryMax {
		delay = config.RetryMax
	}
	if delay < time.Millisecond {
		return time.Millisecond
	}
	return delay
}
