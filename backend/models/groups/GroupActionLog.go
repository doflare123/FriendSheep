package groups

import (
	"friendship/models"
	"time"
)

type GroupActionLog struct {
	ID          uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID     uint        `json:"groupId"`
	Group       Group       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserID      uint        `json:"userId"`
	User        models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Username    string      `json:"username"`
	Us          string      `json:"us"`
	Role        string      `json:"role"`
	Action      string      `json:"action"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
}
