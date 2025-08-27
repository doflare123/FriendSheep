package groups

import (
	"friendship/models"
	"time"
)

type Group struct {
	ID               uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string `json:"nameGroup" gorm:"not null"`
	Description      string `json:"Description" gorm:"not null"`
	SmallDescription string `json:"smallDescription" gorm:"not null"`
	Image            string `json:"image" gorm:"not null"`

	CreaterID uint        `json:"createrId"`
	Creater   models.User `json:"creater" gorm:"foreignKey:CreaterID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	IsPrivate  bool              `json:"isPrivate" gorm:"default:false"`
	City       string            `json:"city"`
	Categories []models.Category `gorm:"many2many:group_group_categories;joinForeignKey:GroupID;JoinReferences:GroupCategoryID"`
	Contacts   []GroupContact    `json:"contacts" gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}
