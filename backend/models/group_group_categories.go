package models

type GroupGroupCategory struct {
	GroupID         uint `gorm:"column:group_id"`
	GroupCategoryID uint `gorm:"column:group_category_id"`
}

func (GroupGroupCategory) TableName() string {
	return "group_group_categories"
}
