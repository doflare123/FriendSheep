package groups

import (
	"friendship/models"
	"time"
)

type GroupBlacklist struct {
	ID        uint        `gorm:"primaryKey;autoIncrement"`
	GroupID   uint        `json:"groupId"`
	Group     Group       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID    uint        `json:"userId"`
	User      models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BannedBy  uint        `json:"bannedBy"`
	Banner    models.User `gorm:"foreignKey:BannedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Reason    string      `json:"reason"`
	CreatedAt time.Time   `json:"createdAt"`
}
