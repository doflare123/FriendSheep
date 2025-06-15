package models

type Group struct {
	ID               uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string `json:"nameGroup" gorm:"not null"`
	Discription      string `json:"discription" gorm:"not null"`
	SmallDiscription string `json:"smallDiscription" gorm:"not null"`
	Image            string `json:"image" gorm:"not null"`

	CreaterID uint `json:"createrId"`
	Creater   User `json:"creater" gorm:"foreignKey:CreaterID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	IsPrivate  bool            `json:"isPrivate" gorm:"default:false"`
	City       string          `json:"city"`
	Categories []GroupCategory `json:"categories" gorm:"many2many:group_categories;"`
}
