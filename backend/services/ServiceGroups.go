package services

import (
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/utils"
)

type CreateGroupInput struct {
	Name             string `json:"name" binding:"required"`
	Discription      string `json:"discription" binding:"required"`
	SmallDiscription string `json:"smallDiscription" binding:"required"`
	Image            string `json:"image" binding:"required"`
	CreaterName      uint   `json:"createrName" binding:"required"`
	IsPrivate        bool   `json:"isPrivate" binding:"required"`
	City             string `json:"city,omitempty"`
	Categories       []uint `json:"categories" binding:"required"`
}

func CreateGroup(input CreateGroupInput) (*models.Group, error) {

	if err := utils.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	var creator models.User

	if err := db.GetDB().First(&creator, input.CreaterName).Error; err != nil {
		return nil, fmt.Errorf("создатель не найден: %v", err)
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
		Discription:      input.Discription,
		SmallDiscription: input.SmallDiscription,
		Image:            input.Image,
		Creater:          creator,
		IsPrivate:        input.IsPrivate,
		City:             input.City,
		Categories:       categories,
	}

	if err := db.GetDB().Create(&group).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания группы: %v", err)
	}

	// Добавление создателя как участника группы
	groupUser := models.GroupUsers{
		UserID:  creator.ID,
		GroupID: group.ID,
		Role:    "admin",
	}

	if err := db.GetDB().Create(&groupUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
	}

	return &group, nil
}
