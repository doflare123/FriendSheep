package statsusers

import "friendship/models"

type SessionStats_users struct {
	ID              uint        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID          uint        `json:"userId"`
	User            models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CountFilms      uint16      `json:"count_films" gorm:"default:0"`
	CountGames      uint16      `json:"count_games" gorm:"default:0"`
	CountTableGames uint16      `json:"count_table_games" gorm:"default:0"`
	CountAnother    uint16      `json:"count_another" gorm:"default:0"`
	CountAll        uint16      `json:"count_all" gorm:"default:0"`
	SpentTime       uint64      `json:"spent_time" gorm:"default:0"`
}
