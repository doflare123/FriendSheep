package statsusers

import "friendship/models"

type SettingTile struct {
	ID          uint        `gorm:"primaryKey;autoIncrement"`
	UserID      uint        `gorm:"not null"`
	User        models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Count_films bool        `json:"count_films" gorm:"default:true"`
	Count_games bool        `json:"count_games" gorm:"default:true"`
	Count_table bool        `json:"count_table" gorm:"default:true"`
	Count_other bool        `json:"count_other" gorm:"default:false"`
	Count_all   bool        `json:"count_all" gorm:"default:true"`
	Spent_time  bool        `json:"spent_time" gorm:"default:false"`
}
