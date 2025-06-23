package db

import (
	"friendship/models"
	"log"
)

func SeedCategories() {
	categories := []models.GroupCategory{
		{Name: "Фильмы"},
		{Name: "Игры"},
		{Name: "Настольные игры"},
		{Name: "Другое"},
	}

	for _, category := range categories {
		var existing models.GroupCategory
		err := GetDB().Where("name = ?", category.Name).First(&existing).Error
		if err == nil {
			continue // Категория уже существует — пропускаем
		}

		if err := GetDB().Create(&category).Error; err != nil {
			log.Printf("Ошибка при создании категории %s: %v", category.Name, err)
		}
	}
}
