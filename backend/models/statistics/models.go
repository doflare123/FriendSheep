package statistics

import "time"

type UserAggregate struct {
	UserID                   uint   `gorm:"primaryKey;autoIncrement:false"`
	CompletedEvents          int64  `gorm:"not null;default:0;check:completed_events >= 0"`
	PlannedMinutes           int64  `gorm:"not null;default:0;check:planned_minutes >= 0"`
	CreatedCompletedEvents   int64  `gorm:"not null;default:0;check:created_completed_events >= 0"`
	UniquePeopleCount        *int64 `gorm:"check:unique_people_count >= 0"`
	UniquePeopleCalculatedAt *time.Time
	UpdatedAt                time.Time `gorm:"not null"`
}

func (UserAggregate) TableName() string { return "participation_statistics_users" }

type UniquePeopleJob struct {
	UserID                 uint       `gorm:"primaryKey;autoIncrement:false;index:idx_unique_people_due,priority:3"`
	RequestedGeneration    uint64     `gorm:"not null;default:0;check:unique_people_generation_and_dirty,requested_generation >= applied_generation AND ((requested_generation > applied_generation) = (dirty_since IS NOT NULL))"`
	AppliedGeneration      uint64     `gorm:"not null;default:0"`
	DirtySince             *time.Time `gorm:"index:idx_unique_people_due,priority:2"`
	FollowupDirtySince     *time.Time
	NextAttemptAt          time.Time  `gorm:"not null;index:idx_unique_people_due,priority:1"`
	LeaseToken             *string    `gorm:"size:64;check:unique_people_lease_pair,(lease_token IS NULL) = (lease_expires_at IS NULL)"`
	LeaseExpiresAt         *time.Time `gorm:"index"`
	LeaseGeneration        uint64     `gorm:"not null;default:0"`
	AttemptCount           uint64     `gorm:"not null;default:0"`
	LastLeaseToken         *string    `gorm:"size:64"`
	LastResultGeneration   uint64     `gorm:"not null;default:0"`
	LastResultCount        *int64     `gorm:"check:last_result_count >= 0"`
	LastResultCalculatedAt *time.Time
	SuccessfulRuns         uint64 `gorm:"not null;default:0"`
	RetryRuns              uint64 `gorm:"not null;default:0"`
	TimeoutRuns            uint64 `gorm:"not null;default:0"`
	TotalDurationMillis    uint64 `gorm:"not null;default:0"`
}

func (UniquePeopleJob) TableName() string { return "participation_statistics_unique_people_jobs" }

type ProcessedEvent struct {
	EventID     uint      `gorm:"primaryKey;autoIncrement:false"`
	ProcessedAt time.Time `gorm:"not null"`
}

func (ProcessedEvent) TableName() string { return "participation_statistics_processed_events" }

type Distribution struct {
	UserID uint   `gorm:"primaryKey;autoIncrement:false"`
	Kind   string `gorm:"primaryKey;size:16"`
	Key    string `gorm:"primaryKey;size:32"`
	Count  int64  `gorm:"not null;check:count > 0"`
}

func (Distribution) TableName() string { return "participation_statistics_distributions" }
