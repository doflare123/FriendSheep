package groups

import "friendship/models"

type GroupUsers struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	UserID uint        `json:"userId" gorm:"not null;uniqueIndex:idx_group_user_membership"`
	User   models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	GroupID uint  `json:"groupId" gorm:"not null;uniqueIndex:idx_group_user_membership"`
	Group   Group `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	RoleInGroupID uint `json:"role"`
	RoleInGroup   Role_in_group
}
