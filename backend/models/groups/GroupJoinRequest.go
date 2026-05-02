package groups

import (
	"friendship/models"
	"time"
)

type GroupJoinRequest struct {
	ID        uint        `gorm:"primaryKey;autoIncrement"`
	UserID    uint        `json:"userId" gorm:"not null"`
	User      models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	GroupID   uint        `json:"groupId" gorm:"not null"`
	Group     Group       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status    string      `json:"status" gorm:"not null"` // "pending", "approved", "rejected"
	CreatedAt time.Time
}
