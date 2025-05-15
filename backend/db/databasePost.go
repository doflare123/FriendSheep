package db

import (
	"fmt"
	"friendship/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Объявление глобальной переменной для хранения экземпляра базы данных
var db *gorm.DB

// Функция инициализации базы данных
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

	return db.AutoMigrate(&models.User{}, &models.SessionType{}, &models.SessionUser{}, &models.Session{})
}

// Функция для получения экземпляра базы данных
func GetDB() *gorm.DB {
	// Возвращение глобальной переменной db, содержащей подключение к базе данных
	return db
}
