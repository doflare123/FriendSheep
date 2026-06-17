package groups

import (
	"friendship/models"
	"time"
)

type GroupActionLog struct {
	ID           uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID      uint        `json:"groupId"`
	Group        Group       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID       uint        `json:"userId"`
	User         models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Username     string      `json:"username"`
	Us           string      `json:"us"`
	Role         string      `json:"role"`
	ActionTypeID uint        `json:"actionTypeId"`
	ActionType   GroupActionType
	Description  string      `json:"description"`
	TargetUserID *uint       `json:"targetUserId"`
	TargetUser   models.User `gorm:"foreignKey:TargetUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	EntityID     *uint       `json:"entityId"`
	EntityName   string      `json:"entityName"`
	CreatedAt    time.Time   `json:"createdAt"`
}
