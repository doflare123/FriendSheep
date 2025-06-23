package services

import (
	"fmt"
	"friendship/db"
	"friendship/models"
)

type CreateGroupInput struct {
	Name             string `json:"name" binding:"required" example:"Моя группа"`
	Description      string `json:"description" binding:"required" example:"Полное описание группы"`
	SmallDescription string `json:"smallDescription" binding:"required" example:"Короткое описание"`
	Image            string `json:"image" binding:"required" example:"https://example.com/image.png"`
	IsPrivate        string `json:"isPrivate" binding:"required" example:"true/1/0/false"` // true / false
	City             string `json:"city,omitempty" example:"Москва"`
	Categories       []uint `json:"categories" binding:"required" example:"[1, 2, 3]"`
}

func (input *CreateGroupInput) IsPrivateBool() (bool, error) {
	switch input.IsPrivate {
	case "1", "true", "True", "TRUE":
		return true, nil
	case "0", "false", "False", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("некорректное значение isPrivate: %s", input.IsPrivate)
	}
}

type JoinGroupInput struct {
	GroupID uint `json:"groupId" binding:"required"`
	UserID  uint `json:"userId" binding:"required"`
}

func CreateGroup(email string, input CreateGroupInput) (*models.Group, error) {
	if err := ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	var creator models.User
	if err := db.GetDB().Where("email = ?", email).First(&creator).Error; err != nil {
		return nil, fmt.Errorf("создатель не найден по email (%s): %v", email, err)
	}

	isPrivate, err := input.IsPrivateBool()
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить isPrivate: %v", err)
	}

	// Загрузка категорий
	var categories []models.GroupCategory
	if len(input.Categories) > 0 {
		if err := db.GetDB().Where("id IN ?", input.Categories).Find(&categories).Error; err != nil {
			return nil, fmt.Errorf("ошибка загрузки категорий: %v", err)
		}
	}

	group := models.Group{
		Name:             input.Name,
		Discription:      input.Description,
		SmallDiscription: input.SmallDescription,
		Image:            input.Image,
		Creater:          creator,
		IsPrivate:        isPrivate,
		City:             input.City,
		Categories:       categories,
	}

	if err := db.GetDB().Create(&group).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания группы: %v", err)
	}

	groupUser := models.GroupUsers{
		UserID:      creator.ID,
		GroupID:     group.ID,
		RoleInGroup: "admin",
	}

	if err := db.GetDB().Create(&groupUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
	}

	return &group, nil
}

func JoinGroup(input JoinGroupInput) (*models.Group, error) {
	if err := ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	newMember := models.GroupUsers{
		UserID:      input.UserID,
		GroupID:     input.GroupID,
		RoleInGroup: "member",
	}
	if err := db.GetDB().Create(&newMember).Error; err != nil {
		return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
	}
	return nil, nil
}
