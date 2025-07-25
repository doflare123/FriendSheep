package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
)

type CreateGroupInput struct {
	Name             string `form:"name" binding:"required" example:"Моя группа"`
	Description      string `form:"description" binding:"required" example:"Полное описание группы"`
	SmallDescription string `form:"smallDescription" binding:"required" example:"Короткое описание"`
	Image            string `form:"-"`
	IsPrivate        bool   `form:"isPrivate" binding:"required" example:"true/1/0/false"` // true / false
	City             string `form:"city,omitempty" example:"Москва"`
	Categories       []uint `form:"categories" binding:"required" example:"[1, 2, 3]"`
}

type JoinGroupResult struct {
	Message string
	Joined  bool
}

// func (input *CreateGroupInput) IsPrivateBool() (bool, error) {
// 	switch input.IsPrivate {
// 	case "1", "true", "True", "TRUE":
// 		return true, nil
// 	case "0", "false", "False", "FALSE":
// 		return false, nil
// 	default:
// 		return false, fmt.Errorf("некорректное значение isPrivate: %s", input.IsPrivate)
// 	}
// }

type JoinGroupInput struct {
	GroupID uint `json:"groupId" binding:"required"`
}

func CreateGroup(email string, input CreateGroupInput) (*groups.Group, error) {
	if err := ValidateInput(input); err != nil {
		return nil, fmt.Errorf("невалидная структура данных: %v", err)
	}

	var creator models.User
	if err := db.GetDB().Where("email = ?", email).First(&creator).Error; err != nil {
		return nil, fmt.Errorf("создатель не найден по email (%s): %v", email, err)
	}

	// Загрузка категорий
	var categories []models.Category
	if len(input.Categories) > 0 {
		if err := db.GetDB().Where("id IN ?", input.Categories).Find(&categories).Error; err != nil {
			return nil, fmt.Errorf("ошибка загрузки категорий: %v", err)
		}
	}

	group := groups.Group{
		Name:             input.Name,
		Description:      input.Description,
		SmallDescription: input.SmallDescription,
		Image:            input.Image,
		Creater:          creator,
		IsPrivate:        input.IsPrivate,
		City:             input.City,
		Categories:       categories,
	}

	if err := db.GetDB().Create(&group).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания группы: %v", err)
	}

	groupUser := groups.GroupUsers{
		UserID:      creator.ID,
		GroupID:     group.ID,
		RoleInGroup: "admin",
	}

	if err := db.GetDB().Create(&groupUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
	}

	return &group, nil
}

func JoinGroup(email string, input JoinGroupInput) (*JoinGroupResult, error) {
	// Найти пользователя
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	// Проверить, есть ли уже заявка или участие
	var existing groups.GroupUsers
	if err := db.GetDB().
		Where("user_id = ? AND group_id = ?", user.ID, input.GroupID).
		First(&existing).Error; err == nil {
		return nil, fmt.Errorf("пользователь уже в группе")
	}

	// Проверка на уже поданную заявку
	var existingRequest groups.GroupJoinRequest
	if err := db.GetDB().
		Where("user_id = ? AND group_id = ?", user.ID, input.GroupID).
		First(&existingRequest).Error; err == nil {
		if existingRequest.Status == "pending" {
			return nil, fmt.Errorf("заявка уже отправлена и ожидает подтверждения")
		}
	}

	// Получить группу
	var group groups.Group
	if err := db.GetDB().Where("id = ?", input.GroupID).First(&group).Error; err != nil {
		return nil, fmt.Errorf("группа не найдена")
	}

	if group.IsPrivate {
		// Группа закрыта — создаём заявку
		request := groups.GroupJoinRequest{
			UserID:  user.ID,
			GroupID: group.ID,
			Status:  "pending",
		}
		if err := db.GetDB().Create(&request).Error; err != nil {
			return nil, fmt.Errorf("ошибка создания заявки: %v", err)
		}
		return &JoinGroupResult{
			Message: "Заявка на вступление отправлена, ожидайте подтверждения от администратора группы",
			Joined:  false,
		}, nil // возвращаем группу с инфой, но не добавляем в participants
	} else {
		// Открытая группа — добавляем напрямую
		member := groups.GroupUsers{
			UserID:      user.ID,
			GroupID:     group.ID,
			RoleInGroup: "member",
		}
		if err := db.GetDB().Create(&member).Error; err != nil {
			return nil, fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
		}
		return &JoinGroupResult{
			Message: "Пользователь успешно присоединился к группе",
			Joined:  true,
		}, nil
	}
}

func DeleteGroup(requesterEmail string, groupID uint) error {
	var requester models.User
	if err := db.GetDB().Where("email = ?", requesterEmail).First(&requester).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	role, err := getUserRole(requester.ID, groupID)
	if err != nil || role != "admin" {
		return errors.New("только админ может удалить группу")
	}

	var group groups.Group
	if err := db.GetDB().First(&group, groupID).Error; err != nil {
		return errors.New("группа не найдена")
	}

	tx := db.GetDB().Begin()

	if err := tx.Where("group_id = ?", groupID).Delete(&groups.GroupUsers{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("group_id = ?", groupID).Delete(&groups.GroupJoinRequest{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("group_id = ?", groupID).Delete(&groups.GroupGroupCategory{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&group).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

//Admin часть групп

type GroupUpdateInput struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	SmallDescription *string `json:"small_description"`
	Image            *string `json:"image"`
	IsPrivate        *bool   `json:"is_private"`
	City             *string `json:"city"`
}

func UpdateGroup(email string, groupID uint, input GroupUpdateInput) error {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}
	isAdmin, err := getUserRole(user.ID, groupID)
	if isAdmin != "admin" {
		return err
	}

	var group groups.Group
	if err := db.GetDB().First(&group, groupID).Error; err != nil {
		return fmt.Errorf("группа не найдена")
	}
	if input.Name != nil {
		group.Name = *input.Name
	}
	if input.Description != nil {
		group.Description = *input.Description
	}
	if input.SmallDescription != nil {
		group.SmallDescription = *input.SmallDescription
	}
	if input.Image != nil {
		group.Image = *input.Image
	}
	if input.IsPrivate != nil {
		group.IsPrivate = *input.IsPrivate
	}
	if input.City != nil {
		group.City = *input.City
	}

	if err := db.GetDB().Save(&group).Error; err != nil {
		return fmt.Errorf("не удалось сохранить изменения: %v", err)
	}

	return nil
}
