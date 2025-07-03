package db

import (
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"
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

	return db.AutoMigrate(&models.User{},
		&groups.Group{}, &groups.GroupCategory{}, &groups.GroupUsers{}, &groups.GroupJoinRequest{},
		&sessions.Session{}, &sessions.SessionGroupType{}, &sessions.SessionMetadata{}, &sessions.SessionUser{})
}

func GetDB() *gorm.DB {
	return db
}
