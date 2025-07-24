package statsusers

import "friendship/models"

type PopSessionType struct {
	ID            uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        uint            `json:"user_id" gorm:"not null"`
	User          models.User     `gorm:"foreignKey:UserID"`
	SessionTypeID uint            `json:"session_type_id" gorm:"not null"`
	SessionType   models.Category `gorm:"foreignKey:SessionTypeID"`
}
