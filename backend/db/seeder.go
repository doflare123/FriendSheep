package db

import (
	"friendship/models"
	"friendship/models/sessions"
	statsusers "friendship/models/stats_users"
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

func SeedStatusSessions() {
	statuses := []sessions.Status{
		{Status: "Набор"},
		{Status: "В процессе"},
		{Status: "Завершена"},
	}

	for _, status := range statuses {
		var existing sessions.Status
		err := GetDB().Where("Status = ?", status.Status).First(&existing).Error
		if err == nil {
			continue
		}
		if err := GetDB().Create(&status).Error; err != nil {
			log.Printf("Ошибка при создании типа сессии %s: %v", status.Status, err)
		}
	}
}

func SeedGenres() {
	genres := []statsusers.Genre{
		// Фильмы
		{Name: "Боевик"},
		{Name: "Приключения"},
		{Name: "Комедия"},
		{Name: "Драма"},
		{Name: "Фэнтези"},
		{Name: "Ужасы"},
		{Name: "Мистика"},
		{Name: "Романтика"},
		{Name: "Триллер"},
		{Name: "Научная фантастика"},
		{Name: "Анимация"},
		{Name: "Документальный"},
		// Игры
		{Name: "РПГ"},
		{Name: "Шутер"},
		{Name: "Стратегия"},
		{Name: "Симулятор"},
		{Name: "Спорт"},
		{Name: "Гонки"},
		{Name: "Файтинг"},
		{Name: "Платформер"},
		{Name: "Головоломка"},
		{Name: "Выживание"},
		{Name: "Песочница"},
	}

	for _, genre := range genres {
		var existing statsusers.Genre
		err := GetDB().Where("name = ?", genre.Name).First(&existing).Error
		if err == nil {
			continue
		}
		if err := GetDB().Create(&genre).Error; err != nil {
			log.Printf("Ошибка при создании жанра %s: %v", genre.Name, err)
		}
	}
}

func SeedDays() {
	days := []models.DaysWeek{
		{Name: "Понедельник"},
		{Name: "Вторник"},
		{Name: "Среда"},
		{Name: "Четверг"},
		{Name: "Пятница"},
		{Name: "Суббота"},
		{Name: "Воскресенье"},
	}

	for _, day := range days {
		var existing models.DaysWeek
		err := GetDB().Where("name = ?", day.Name).First(&existing).Error
		if err == nil {
			continue
		}
		if err := GetDB().Create(&day).Error; err != nil {
			log.Printf("Ошибка при создании жанра %s: %v", day.Name, err)
		}
	}
}
