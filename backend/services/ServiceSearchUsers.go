package services

import (
	"friendship/db"
	"friendship/models"
)

type GetUsersResponse struct {
	Users   []GetUsers `json:"users"`
	HasMore *bool      `json:"has_more"`
	Page    *int       `json:"page"`
	Total   *int64     `json:"total"`
}

type GetUsers struct {
	Id     *uint   `json:"id"`
	Name   *string `json:"name"`
	Us     *string `json:"us"`
	Status *string `json:"status"`
	Image  *string `json:"image"`
}

const UsersPerPage = 5

func SearchUsers(name string, page int, userId uint) (*GetUsersResponse, error) {
	var users []models.User
	var total int64

	query := db.GetDB().Model(&models.User{}).Where("id != ?", userId)

	if name != "" {
		searchPattern := "%" + name + "%"
		query = query.Where("name ILIKE ?", searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * UsersPerPage

	if err := query.Limit(UsersPerPage).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}

	var usersDTO []GetUsers
	for _, user := range users {
		userDTO := GetUsers{
			Id:     &user.ID,
			Name:   &user.Name,
			Us:     &user.Us,
			Status: &user.Status,
			Image:  &user.Image,
		}
		usersDTO = append(usersDTO, userDTO)
	}

	hasMore := int64(offset+len(users)) < total

	response := &GetUsersResponse{
		Users:   usersDTO,
		HasMore: &hasMore,
		Page:    &page,
		Total:   &total,
	}

	return response, nil
}
