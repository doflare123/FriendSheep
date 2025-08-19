package news

import "time"

type News struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string `json:"title" gorm:"not null" validate:"required"`
	Description string `json:"description" gorm:"not null" validate:"required"`
	Image       string `json:"image" gorm:"not null" validate:"required"`
	CreatedTime string `json:"created_time"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Content     []ContentNews `json:"content" gorm:"foreignKey:NewsID;constraint:OnDelete:CASCADE"`
	Comments    []Comments    `json:"comments" gorm:"foreignKey:NewsID;constraint:OnDelete:CASCADE"`
}
