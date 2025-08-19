package db

import (
	"notify_service/models/sessions"

	"gorm.io/gorm"
)

func SeedNotificationTypes(db *gorm.DB) error {
	types := []sessions.NotificationType{
		{Name: "24_hours", Description: "За 24 часа до начала", HoursBefore: 24},
		{Name: "6_hours", Description: "За 6 часов до начала", HoursBefore: 6},
		{Name: "1_hour", Description: "За 1 час до начала", HoursBefore: 1},
	}

	for _, t := range types {
		var existing sessions.NotificationType
		if err := db.Where("name = ?", t.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&t).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
