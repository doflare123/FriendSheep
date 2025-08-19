package db

import (
	"fmt"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/news"
	"friendship/models/sessions"
	statsusers "friendship/models/stats_users"
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
	db.AutoMigrate(&news.News{}, &news.ContentNews{}, &news.Comments{})
	db.AutoMigrate(&statsusers.SideStats_users{}, &statsusers.SessionStats_users{}, &statsusers.SessionsStatsGenres_users{},
		&statsusers.Genre{}, statsusers.PopSessionType{}, statsusers.SettingTile{})
	db.AutoMigrate(&models.User{}, models.StatsProcessedEvent{},
		&groups.Group{}, &groups.GroupContact{}, &groups.GroupGroupCategory{}, &models.Category{}, &groups.GroupUsers{}, &groups.GroupJoinRequest{}, &groups.GroupJoinInvite{},
		&sessions.Session{}, &sessions.SessionGroupType{}, &sessions.SessionMetadata{}, sessions.Status{},
	)

	return db.AutoMigrate(&sessions.SessionUser{})
}

func GetDB() *gorm.DB {
	return db
}
