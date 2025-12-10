package groups

import (
	"fmt"
	"friendship/models"
	"friendship/models/dto"
	"friendship/repository"
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

func (g *Group) GetAdminGroups(userID uint, rep repository.PostgresRepository) ([]dto.AdminGroupResponse, error) {
	if userID == 0 {
		return nil, fmt.Errorf("некорректный ID пользователя")
	}

	var adminGroups []dto.AdminGroupResponse

	err := rep.Model(&Group{}).
		Select(
			"groups.id",
			"groups.name",
			"groups.image",
			"groups.small_description",
			"CASE WHEN groups.is_private THEN 'приватная группа' ELSE 'открытая группа' END as type",
			"COUNT(DISTINCT gu2.user_id) as member_count",
		).
		Joins("JOIN group_users gu ON gu.group_id = groups.id").
		Joins("LEFT JOIN group_users gu2 ON gu2.group_id = groups.id").
		Where("gu.user_id = ? AND gu.role_in_group = ?", userID, "admin").
		Group("groups.id, groups.name, groups.image, groups.small_description, groups.is_private").
		Order("groups.id DESC").
		Scan(&adminGroups).Error

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении групп: %w", err)
	}

	for i := range adminGroups {
		if adminGroups[i].ID != nil {
			categories, err := g.loadGroupCategories(*adminGroups[i].ID, rep)
			if err != nil {
				return nil, fmt.Errorf("ошибка при получении категорий для группы %d: %w", *adminGroups[i].ID, err)
			}
			adminGroups[i].Category = categories
		}
	}

	return adminGroups, nil
}

func (g *Group) loadGroupCategories(groupID uint, rep repository.PostgresRepository) ([]*string, error) {
	var categoryNames []string

	err := rep.Model(&models.Category{}).
		Select("categories.name").
		Joins("JOIN group_group_categories ON group_group_categories.group_category_id = categories.id").
		Where("group_group_categories.group_id = ?", groupID).
		Pluck("name", &categoryNames).Error

	if err != nil {
		return nil, err
	}

	categories := make([]*string, len(categoryNames))
	for j, name := range categoryNames {
		categoryName := name
		categories[j] = &categoryName
	}

	return categories, nil
}
