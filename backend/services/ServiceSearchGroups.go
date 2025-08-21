package services

import (
	"friendship/db"
	"friendship/models/groups"
)

type GetGroupsResponse struct {
	Groups  []GetGroups `json:"groups"`
	HasMore *bool       `json:"has_more"`
	Page    *int        `json:"page"`
	Total   *int64      `json:"total"`
}

type GetGroups struct {
	Id               *uint     `json:"id"`
	Name             *string   `json:"name"`
	SmallDescription *string   `json:"description"`
	CountMembers     *uint     `json:"count"`
	Image            *string   `json:"image"`
	Category         []*string `json:"category"`
}

const GroupsPerPage = 4

func SearchGroups(name string, page int) (*GetGroupsResponse, error) {
	var groupsData []groups.Group
	var total int64

	query := db.GetDB().Model(&groups.Group{}).Preload("Categories").Where("is_private = ? OR is_private IS NULL", false)

	if name != "" {
		searchPattern := "%" + name + "%"
		query = query.Where("name ILIKE ?", searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * GroupsPerPage
	if err := query.Limit(GroupsPerPage).Offset(offset).Find(&groupsData).Error; err != nil {
		return nil, err
	}

	var groupsDTO []GetGroups
	for _, group := range groupsData {
		// Получаем категории
		var categories []*string
		for _, cat := range group.Categories {
			categories = append(categories, &cat.Name)
		}

		var memberCount int64
		db.GetDB().Table("group_users").Where("group_id = ?", group.ID).Count(&memberCount)
		memberCountUint := uint(memberCount)

		groupDTO := GetGroups{
			Id:               &group.ID,
			Name:             &group.Name,
			SmallDescription: &group.SmallDescription,
			CountMembers:     &memberCountUint,
			Image:            &group.Image,
			Category:         categories,
		}
		groupsDTO = append(groupsDTO, groupDTO)
	}

	hasMore := int64(offset+len(groupsData)) < total

	response := &GetGroupsResponse{
		Groups:  groupsDTO,
		HasMore: &hasMore,
		Page:    &page,
		Total:   &total,
	}

	return response, nil
}
