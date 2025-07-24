package statsusers

import "friendship/models"

type SessionsStatsGenres_users struct {
	ID      uint        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID  uint        `json:"userId"`
	User    models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	GenreID uint        `json:"genreId"`
	Genre   Genre       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
