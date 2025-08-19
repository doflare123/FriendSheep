package statsusers

import "friendship/models"

type SessionsStatsGenres_users struct {
	ID      *uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID  *uint       `json:"userId" gorm:"uniqueIndex:idx_user_genre"`
	User    models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	GenreID *uint       `json:"genreId" gorm:"uniqueIndex:idx_user_genre"`
	Count   *uint16     `json:"count" gorm:"default:0"`
	Genre   Genre       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
