package statsusers

import "friendship/models"

type SideStats_users struct {
	ID                 *uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID             *uint            `json:"userId"`
	User               models.User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CountCreateSession *uint16          `json:"count_create_session" gorm:"default:0"`
	SeriesSesionCount  *uint16          `json:"series_session_count" gorm:"default:0"`
	MostPopDay         *uint16          `json:"dayWeekId"`                                                                         // Это должно быть nullable
	DaysWeek           *models.DaysWeek `gorm:"foreignKey:MostPopDay;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"` // Добавляем указатель
	MostBigSession     *uint16          `json:"most_big_session" gorm:"default:0"`
}
