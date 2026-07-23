package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"friendship/models"
	"friendship/models/events"
	"friendship/models/groups"
	statsusers "friendship/models/stats_users"
	"friendship/repository"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seeder(db repository.PostgresRepository) []error {
	seeds := []struct {
		name string
		data []interface{}
	}{
		{
			name: "категории",
			data: []interface{}{
				&models.Category{Name: "Фильмы"},
				&models.Category{Name: "Игры"},
				&models.Category{Name: "Настольные игры"},
				&models.Category{Name: "Другое"},
			},
		},
		{
			name: "типы мероприятий",
			data: []interface{}{
				&events.EventLocation{Name: "Онлайн"},
				&events.EventLocation{Name: "Оффлайн"},
			},
		},
		{
			name: "статусы",
			data: []interface{}{
				&events.Status{Name: "Набор"},
				&events.Status{Name: "В процессе"},
				&events.Status{Name: "Завершена"},
			},
		},
		{
			name: "Дни недели",
			data: []interface{}{
				&models.DaysWeek{Name: "Понедельник"},
				&models.DaysWeek{Name: "Вторник"},
				&models.DaysWeek{Name: "Среда"},
				&models.DaysWeek{Name: "Четверг"},
				&models.DaysWeek{Name: "Пятница"},
				&models.DaysWeek{Name: "Суббота"},
				&models.DaysWeek{Name: "Воскресенье"},
			},
		},
		{
			name: "типы участников групп",
			data: []interface{}{
				&groups.Role_in_group{Name: groups.RoleAdmin},
				&groups.Role_in_group{Name: groups.RoleModerator},
				&groups.Role_in_group{Name: groups.RoleMember},
			},
		},
		{
			name: "типы действий группы",
			data: groupActionTypeSeeds(),
		},
		{
			name: "типы ограничений по возрасту",
			data: []interface{}{
				&events.AgeLimit{Name: "18+"},
				&events.AgeLimit{Name: "16+"},
				&events.AgeLimit{Name: "12+"},
				&events.AgeLimit{Name: "6+"},
				&events.AgeLimit{Name: "0+"},
			},
		},
	}

	var errs []error

	for _, seed := range seeds {
		if err := genericSeed(db, seed.data); err != nil {
			errs = append(errs, fmt.Errorf("ошибка при заполнении %s: %w", seed.name, err))
		}
	}

	if err := seedGenresFromJSON(db, "db/data/genres.json"); err != nil {
		errs = append(errs, fmt.Errorf("ошибка при заполнении жанров: %w", err))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func groupActionTypeSeeds() []interface{} {
	actionTypes := groups.DefaultGroupActionTypes()
	items := make([]interface{}, 0, len(actionTypes))
	for i := range actionTypes {
		items = append(items, &actionTypes[i])
	}

	return items
}

func genericSeed(db repository.PostgresRepository, items []interface{}) error {
	for _, item := range items {
		if status, ok := item.(*events.Status); ok {
			var existing events.Status
			err := db.Where("name = ?", status.Name).First(&existing).Error
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedGenresFromJSON(db repository.PostgresRepository, filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл %s: %w", filepath, err)
	}

	var jsonData struct {
		Genres []struct {
			Name string `json:"name"`
		} `json:"genres"`
	}

	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("не удалось распарсить JSON: %w", err)
	}

	for _, g := range jsonData.Genres {
		genre := &statsusers.Genre{Name: g.Name}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(genre).Error; err != nil {
			return fmt.Errorf("ошибка при создании жанра %s: %w", g.Name, err)
		}
	}

	return nil
}
