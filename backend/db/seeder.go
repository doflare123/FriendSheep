package db

import (
	"friendship/models"
	"friendship/models/sessions"
	"log"
)

func SeedCategories() {
	categories := []models.Category{
		{Name: "Фильмы"},
		{Name: "Игры"},
		{Name: "Настольные игры"},
		{Name: "Другое"},
	}

	for _, category := range categories {
		var existing models.Category
		err := GetDB().Where("name = ?", category.Name).First(&existing).Error
		if err == nil {
			continue
		}

		if err := GetDB().Create(&category).Error; err != nil {
			log.Printf("Ошибка при создании категории %s: %v", category.Name, err)
		}
	}
}

func SeedCategoriesSessionsVisibility() {
	categories := []sessions.SessionGroupPlace{
		{Title: "Онлайн"},
		{Title: "Оффлайн"},
	}

	for _, category := range categories {
		var existing sessions.SessionGroupPlace
		err := GetDB().Where("title = ?", category.Title).First(&existing).Error
		if err == nil {
			continue
		}
		if err := GetDB().Create(&category).Error; err != nil {
			log.Printf("Ошибка при создании типа сессии %s: %v", category.Title, err)
		}
	}
}
