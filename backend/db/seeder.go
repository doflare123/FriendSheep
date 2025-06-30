package db

import (
	"friendship/models/groups"
	"log"
)

func SeedCategories() {
	categories := []groups.GroupCategory{
		{Name: "Фильмы"},
		{Name: "Игры"},
		{Name: "Настольные игры"},
		{Name: "Другое"},
	}

	for _, category := range categories {
		var existing groups.GroupCategory
		err := GetDB().Where("name = ?", category.Name).First(&existing).Error
		if err == nil {
			continue // Категория уже существует — пропускаем
		}

		if err := GetDB().Create(&category).Error; err != nil {
			log.Printf("Ошибка при создании категории %s: %v", category.Name, err)
		}
	}
}
