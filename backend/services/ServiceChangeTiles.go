package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
	statsusers "friendship/models/stats_users"
)

type ChangeTilesPatternInput struct {
	Count_films bool `json:"count_films"`
	Count_games bool `json:"count_games"`
	Count_table bool `json:"count_table"`
	Count_other bool `json:"count_other"`
	Count_all   bool `json:"count_all"`
	Spent_time  bool `json:"spent_time"`
}

func ChangePattern(email string, input ChangeTilesPatternInput) error {
	db := db.GetDB()

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	var settings statsusers.SettingTile

	if err := db.Where("user_id = ?", user.ID).First(&settings).Error; err != nil {
		// создаём с дефолтами
		settings = statsusers.SettingTile{
			UserID:      user.ID,
			Count_films: input.Count_films,
			Count_games: input.Count_games,
			Count_table: input.Count_table,
			Count_other: input.Count_other,
			Count_all:   input.Count_all,
			Spent_time:  input.Spent_time,
		}
		return db.Create(&settings).Error
	}

	return db.Model(&settings).Updates(map[string]interface{}{
		"count_films": input.Count_films,
		"count_games": input.Count_games,
		"count_table": input.Count_table,
		"count_other": input.Count_other,
		"count_all":   input.Count_all,
		"spent_time":  input.Spent_time,
	}).Error
}
