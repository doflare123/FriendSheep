package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models/groups"

	"gorm.io/gorm"
)

type GroupResponse struct {
	ID               *uint     `json:"id"`
	Name             *string   `json:"name"`
	Image            *string   `json:"image"`
	Type             *string   `json:"type"`
	SmallDescription *string   `json:"small_description"`
	Category         []*string `json:"category"`
	MemberCount      *int64    `json:"member_count"`
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
		Where("user_id = ? AND role_in_group = 'member'", &user.ID).
		Find(&groupUsers).Error

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении групп: %v", err)
	}

	if len(groupUsers) == 0 {
		return []GroupResponse{}, nil
	}

	// Фильтруем группы, где пользователь НЕ является создателем
	filteredGroups := make([]groups.GroupUsers, 0)
	groupIDs := make([]uint, 0)

	for _, gu := range groupUsers {
		if gu.Group.ID != 0 && gu.Group.CreaterID != 0 &&
			user.ID != 0 && &gu.Group.CreaterID != &user.ID {
			filteredGroups = append(filteredGroups, gu)
			groupIDs = append(groupIDs, gu.Group.ID)
		}
	}

	if len(filteredGroups) == 0 {
		return []GroupResponse{}, nil
	}

	// Получаем количество участников для групп
	memberCounts, err := getGroupMemberCounts(groupIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подсчете участников: %v", err)
	}

	// Преобразуем в GroupResponse
	groupResponses := make([]GroupResponse, 0, len(filteredGroups))
	for _, groupUser := range filteredGroups {
		groupResponse := createSafeGroupResponse(groupUser.Group, memberCounts)
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

	response := GroupResponse{
		ID: &group.ID,
	}

	if group.Name != "" {
		response.Name = &group.Name
	}

	if group.SmallDescription != "" {
		response.SmallDescription = &group.SmallDescription
	}

	if group.Image != "" {
		response.Image = &group.Image
	}

	// Обрабатываем категории
	if len(group.Categories) > 0 {
		categories := make([]*string, 0, len(group.Categories))
		for _, category := range group.Categories {
			if category.Name != "" {
				categories = append(categories, &category.Name)
			}
		}
		response.Category = categories
	}

	// Устанавливаем количество участников
	if count, exists := memberCounts[group.ID]; exists {
		response.MemberCount = &count
	} else {
		zero := int64(0)
		response.MemberCount = &zero
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
