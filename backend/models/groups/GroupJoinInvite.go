package groups

import (
	"friendship/models"
	"time"
)

type GroupJoinInvite struct {
	ID        uint        `gorm:"primaryKey;autoIncrement"`
	UserID    uint        `json:"userId"`
	User      models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	GroupID   uint        `json:"groupId"`
	Group     Group       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status    string      `json:"status"` // JoinStatusPending, JoinStatusAccepted, JoinStatusRejected
	CreatedAt time.Time
}
