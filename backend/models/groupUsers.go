package models

type GroupUsers struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	UserID uint `json:"userId"`
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	GroupID uint  `json:"groupId"`
	Group   Group `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Role string `json:"role"`
}
