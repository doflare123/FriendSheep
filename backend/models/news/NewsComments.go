package news

import "friendship/models"

type Comments struct {
	ID     int         `json:"id" gorm:"primaryKey;autoIncrement"`
	NewsID uint        `json:"news_id" gorm:"not null;index"`
	News   News        `json:"-" gorm:"foreignKey:NewsID;constraint:OnDelete:CASCADE"`
	Text   string      `json:"text" gorm:"not null"`
	UserID uint        `json:"user_id" gorm:"not null"`
	User   models.User `json:"user" gorm:"foreignKey:UserID"`
}
