package statistics

import "time"

type UserAggregate struct {
	UserID                 uint      `gorm:"primaryKey;autoIncrement:false"`
	CompletedEvents        int64     `gorm:"not null;default:0;check:completed_events >= 0"`
	PlannedMinutes         int64     `gorm:"not null;default:0;check:planned_minutes >= 0"`
	CreatedCompletedEvents int64     `gorm:"not null;default:0;check:created_completed_events >= 0"`
	UniquePeopleCount      int64     `gorm:"not null;default:0;check:unique_people_count >= 0"`
	UpdatedAt              time.Time `gorm:"not null"`
}

func (UserAggregate) TableName() string { return "participation_statistics_users" }

type UserPair struct {
	UserID            uint `gorm:"primaryKey;autoIncrement:false;check:statistics_pair_not_self,user_id <> encountered_user_id"`
	EncounteredUserID uint `gorm:"primaryKey;autoIncrement:false"`
}

func (UserPair) TableName() string { return "participation_statistics_pairs" }

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
