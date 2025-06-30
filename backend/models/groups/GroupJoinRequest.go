package groups

import (
	"friendship/models"
	"time"
)

type GroupJoinRequest struct {
	ID        uint        `gorm:"primaryKey;autoIncrement"`
	UserID    uint        `json:"userId"`
	User      models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	GroupID   uint        `json:"groupId"`
	Group     Group       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status    string      `json:"status"` // "pending", "approved", "rejected"
	CreatedAt time.Time
}
