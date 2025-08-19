package statsusers

import "friendship/models"

type PopSessionType struct {
	ID            uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        uint            `gorm:"not null;index:ux_user_type,unique"`
	User          models.User     `gorm:"foreignKey:UserID"`
	SessionTypeID uint            `gorm:"not null;index:ux_user_type,unique"`
	SessionType   models.Category `gorm:"foreignKey:SessionTypeID"`
}
