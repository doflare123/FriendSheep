package services

import (
	"fmt"
	"friendship/db"
)

type AdminGroupResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Type        string `json:"type"`
	MemberCount int64  `json:"member_count"`
}

// GetAdminGroups получает все группы, где пользователь является создателем.
// Возвращает срез структур AdminGroupResponse, содержащих ID, имя, изображение и количество участников.
func GetAdminGroups(email string) ([]AdminGroupResponse, error) {
	db := db.GetDB()

	user, err := FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %v", err)
	}

	var adminGroups []AdminGroupResponse
	if err := db.Table("groups").
		Select("groups.id, groups.name, groups.image, CASE WHEN groups.is_private THEN 'приватная группа' ELSE 'открытая группа' END as type, COUNT(group_users.user_id) as member_count").
		Joins("LEFT JOIN group_users ON group_users.group_id = groups.id").
		Where("groups.creater_id = ?", user.ID).
		Group("groups.id, groups.is_private").
		Order("groups.id DESC").
		Scan(&adminGroups).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении групп: %v", err)
	}

	return adminGroups, nil
}
