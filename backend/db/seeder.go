package db

import (
	"friendship/models/groups"
	"friendship/models/sessions"
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

func SeedCategoriesSessions() {
	categories := []sessions.SessionGroupType{
		{Title: "Онлайн"},
		{Title: "Оффлайн"},
	}

	for _, category := range categories {
		var existing sessions.SessionGroupType
		err := GetDB().Where("title = ?", category.Title).First(&existing).Error
		if err == nil {
			continue // уже существует
		}
		// ❗️ ты забывал это:
		if err := GetDB().Create(&category).Error; err != nil {
			log.Printf("Ошибка при создании типа сессии %s: %v", category.Title, err)
		}
	}
}
