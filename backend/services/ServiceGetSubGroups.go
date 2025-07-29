package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models/groups"

	"gorm.io/gorm"
)

type GroupResponse struct {
	ID               uint     `json:"id"`
	Name             string   `json:"name"`
	Category         []string `json:"category"`
	SmallDescription string   `json:"description"`
	Image            string   `json:"image"`
	CountMembers     int64    `json:"count_members"`
}

func GetGroupsUserSub(email string) ([]GroupResponse, error) {
	if email == "" {
		return nil, fmt.Errorf("не передан jwt")
	}
	user, err := FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %v", err)
	}

	var groupUsers []groups.GroupUsers
	err = db.GetDB().Preload("Group").
		Preload("Group.Categories").
		Where("user_id = ? AND role_in_group != ?", user.ID, "creator").
		Find(&groupUsers).Error

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении групп: %v", err)
	}

	if len(groupUsers) == 0 {
		return []GroupResponse{}, nil
	}

	// Получаем ID групп для подсчета участников
	groupIDs := make([]uint, 0, len(groupUsers))
	for _, gu := range groupUsers {
		if gu.Group.ID != 0 {
			groupIDs = append(groupIDs, gu.Group.ID)
		}
	}

	memberCounts, err := getGroupMemberCounts(groupIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подсчете участников: %v", err)
	}

	// Преобразуем в GroupResponse с дополнительной фильтрацией
	groupResponses := make([]GroupResponse, 0, len(groupUsers))
	for _, groupUser := range groupUsers {
		group := groupUser.Group

		if group.CreaterID != 0 && user.ID != 0 && group.CreaterID == user.ID {
			continue
		}

		groupResponse := createSafeGroupResponse(group, memberCounts)
		if groupResponse != nil {
			groupResponses = append(groupResponses, *groupResponse)
		}
	}

	return groupResponses, nil
}

func createSafeGroupResponse(group groups.Group, memberCounts map[uint]int64) *GroupResponse {
	if group.ID == 0 {
		return nil
	}

	var response GroupResponse

	groupID := &group.ID
	response.ID = *groupID

	if group.Name != "" {
		groupName := &group.Name
		response.Name = *groupName
	}

	if group.SmallDescription != "" {
		smallDesc := &group.SmallDescription
		response.SmallDescription = *smallDesc
	}

	if group.Image != "" {
		groupImage := &group.Image
		response.Image = *groupImage
	}

	// Безопасно извлекаем категории
	categories := make([]string, 0, len(group.Categories))
	for _, category := range group.Categories {
		if category.Name != "" {
			categoryName := &category.Name
			categories = append(categories, *categoryName)
		}
	}
	response.Category = categories

	if count, exists := memberCounts[group.ID]; exists {
		response.CountMembers = count
	} else {
		zero := int64(0)
		response.CountMembers = zero
	}

	return &response
}

// getGroupMemberCounts получает количество участников для каждой группы
func getGroupMemberCounts(groupIDs []uint) (map[uint]int64, error) {
	if len(groupIDs) == 0 {
		return make(map[uint]int64), nil
	}

	type GroupMemberCount struct {
		GroupID uint  `json:"group_id"`
		Count   int64 `json:"count"`
	}

	var counts []GroupMemberCount
	err := db.GetDB().Model(&groups.GroupUsers{}).
		Select("group_id, COUNT(*) as count").
		Where("group_id IN ?", groupIDs).
		Group("group_id").
		Find(&counts).Error

	if err != nil {
		return nil, err
	}

	// Создаем карту для быстрого доступа
	memberCounts := make(map[uint]int64)
	for _, count := range counts {
		memberCounts[count.GroupID] = count.Count
	}

	// Устанавливаем 0 для групп без участников
	for _, groupID := range groupIDs {
		if _, exists := memberCounts[groupID]; !exists {
			memberCounts[groupID] = 0
		}
	}

	return memberCounts, nil
}

func ValidateUserGroupAccess(userID uint, groupID uint) (bool, error) {
	var count int64
	err := db.GetDB().Model(&groups.GroupUsers{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetUserGroupRole(userID uint, groupID uint) (string, error) {
	var groupUser groups.GroupUsers
	err := db.GetDB().Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&groupUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("пользователь не состоит в группе")
		}
		return "", err
	}

	if groupUser.RoleInGroup == "" {
		return "", errors.New("роль не определена")
	}

	return groupUser.RoleInGroup, nil
}
