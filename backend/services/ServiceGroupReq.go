package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
)

func GetPendingJoinRequestsForAdmin(email string) ([]models.GroupJoinRequest, error) {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("пользователь не найден")
	}

	var requests []models.GroupJoinRequest
	err := db.GetDB().
		Table("group_join_requests").
		Joins("JOIN groups ON group_join_requests.group_id = groups.id").
		Joins("JOIN group_users ON group_users.group_id = groups.id").
		Where("group_users.user_id = ? AND group_users.role_in_group = ? AND group_join_requests.status = ?", user.ID, "admin", "pending").
		Preload("User").
		Preload("Group").
		Find(&requests).Error

	return requests, err
}

func ApproveJoinRequestByID(email string, requestID string) error {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	var request models.GroupJoinRequest
	if err := db.GetDB().Preload("Group").First(&request, requestID).Error; err != nil {
		return errors.New("заявка не найдена")
	}

	var adminCheck models.GroupUsers
	if err := db.GetDB().Where("user_id = ? AND group_id = ? AND role_in_group = ?", user.ID, request.GroupID, "admin").First(&adminCheck).Error; err != nil {
		return errors.New("вы не администратор этой группы")
	}

	// Обновить статус
	if err := db.GetDB().Model(&request).Update("status", "approved").Error; err != nil {
		return err
	}

	member := models.GroupUsers{
		UserID:      request.UserID,
		GroupID:     request.GroupID,
		RoleInGroup: "member",
	}
	return db.GetDB().Create(&member).Error
}

func RejectJoinRequestByID(email string, requestID string) error {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	var request models.GroupJoinRequest
	if err := db.GetDB().Preload("Group").First(&request, requestID).Error; err != nil {
		return errors.New("заявка не найдена")
	}

	var adminCheck models.GroupUsers
	if err := db.GetDB().Where("user_id = ? AND group_id = ? AND role_in_group = ?", user.ID, request.GroupID, "admin").First(&adminCheck).Error; err != nil {
		return errors.New("вы не администратор этой группы")
	}

	return db.GetDB().Model(&request).Update("status", "rejected").Error
}
