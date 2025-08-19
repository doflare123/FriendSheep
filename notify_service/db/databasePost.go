package db

import (
	"fmt"
	"log"
	"notify_service/models/sessions"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDatabase() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	err = db.AutoMigrate(&sessions.Notification{}, &sessions.NotificationType{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}
	if err := SeedNotificationTypes(db); err != nil {
		return fmt.Errorf("failed to seed notification types: %v", err)
	}
	log.Println("Database initialized successfully")
	return nil
}

func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("Database not initialized")
	}
	return db
}
