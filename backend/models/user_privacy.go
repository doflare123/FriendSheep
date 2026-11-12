package models

import "gorm.io/gorm"

// UserPrivacy хранит принадлежащие монолиту настройки видимости профиля.
type UserPrivacy struct {
	UserID            uint     `gorm:"primaryKey;autoIncrement:false"`
	Mode              string   `gorm:"type:varchar(16);not null;default:open;check:user_privacy_mode_check,mode IN ('open','closed','custom')"`
	VisibleStatistics []string `gorm:"serializer:json;type:jsonb;not null"`
	User              User     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserPrivacy) TableName() string { return "user_privacies" }

func (privacy *UserPrivacy) BeforeCreate(_ *gorm.DB) error {
	if privacy.VisibleStatistics == nil {
		privacy.VisibleStatistics = []string{
			"films",
			"games",
			"tableGames",
			"other",
			"completedEvents",
			"plannedMinutes",
			"createdCompletedEvents",
			"uniquePeopleCount",
			"mostFrequentTypeId",
			"mostFrequentGenreId",
			"mostFrequentWeekday",
		}
	}
	return nil
}
