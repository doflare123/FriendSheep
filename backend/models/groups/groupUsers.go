package groups

import "friendship/models"

type GroupUsers struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	UserID uint        `json:"userId"`
	User   models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	GroupID uint  `json:"groupId"`
	Group   Group `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	RoleInGroup string `json:"role"`
}
