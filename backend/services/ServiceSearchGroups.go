package services

import (
	"friendship/db"
	"friendship/models/groups"
	"time"
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
	IsPrivate        *bool     `json:"isPrivate"`
	CreatedAt        time.Time `json:"createdAt"`
}

const GroupsPerPage = 4

func SearchGroups(name string, page int, sortBy, order, category string) (*GetGroupsResponse, error) {
	var groupsData []groups.Group
	var total int64

	dbConn := db.GetDB()

	query := dbConn.Model(&groups.Group{}).
		Preload("Categories").
		Preload("Creater")

	if name != "" {
		searchPattern := "%" + name + "%"
		query = query.Where("groups.name ILIKE ?", searchPattern)
	}

	if category != "" {
		query = query.
			Joins("JOIN group_group_categories ggc ON ggc.group_id = groups.id").
			Joins("JOIN categories c ON c.id = ggc.group_category_id").
			Where("c.name ILIKE ?", "%"+category+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	switch sortBy {
	case "members":
		query = query.
			Joins("LEFT JOIN group_users gu ON gu.group_id = groups.id").
			Group("groups.id").
			Order("COUNT(gu.user_id) " + order)
	case "date":
		query = query.Order("groups.created_at " + order)
	case "category":
		query = query.
			Joins("JOIN group_group_categories ggc2 ON ggc2.group_id = groups.id").
			Joins("JOIN categories c2 ON c2.id = ggc2.group_category_id").
			Group("groups.id").
			Order("MIN(c2.name) " + order)
	default:
		query = query.Order("groups.id desc")
	}

	offset := (page - 1) * GroupsPerPage
	if err := query.Limit(GroupsPerPage).Offset(offset).Find(&groupsData).Error; err != nil {
		return nil, err
	}

	var groupsDTO []GetGroups
	for _, group := range groupsData {
		var categories []*string
		for _, cat := range group.Categories {
			categories = append(categories, &cat.Name)
		}

		var memberCount int64
		dbConn.Table("group_users").Where("group_id = ?", group.ID).Count(&memberCount)
		memberCountUint := uint(memberCount)

		groupsDTO = append(groupsDTO, GetGroups{
			Id:               &group.ID,
			Name:             &group.Name,
			SmallDescription: &group.SmallDescription,
			CountMembers:     &memberCountUint,
			Image:            &group.Image,
			Category:         categories,
			IsPrivate:        &group.IsPrivate,
			CreatedAt:        group.CreatedAt,
		})
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
