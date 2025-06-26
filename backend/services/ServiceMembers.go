package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
)

// Проверка роли пользователя в группе
func getUserRole(userID, groupID uint) (string, error) {
	var groupUser models.GroupUsers
	err := db.GetDB().Where("user_id = ? AND group_id = ?", userID, groupID).First(&groupUser).Error
	if err != nil {
		return "", errors.New("пользователь не состоит в группе")
	}
	return groupUser.RoleInGroup, nil
}

func RemoveUserFromGroup(requesterEmail string, groupID, targetUserID uint) error {
	var requester models.User
	if err := db.GetDB().Where("email = ?", requesterEmail).First(&requester).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	var targetUser models.User
	if err := db.GetDB().First(&targetUser, targetUserID).Error; err != nil {
		return errors.New("удаляемый пользователь не найден")
	}

	role, err := getUserRole(requester.ID, groupID)
	if err != nil || (role != "admin" && role != "operator") {
		return errors.New("нет прав на удаление")
	}

	if requester.ID == targetUserID {
		return errors.New("используйте /leave для выхода")
	}

	return db.GetDB().Where("user_id = ? AND group_id = ?", targetUserID, groupID).Delete(&models.GroupUsers{}).Error
}

func LeaveGroup(email string, groupID uint) error {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	role, err := getUserRole(user.ID, groupID)
	if err != nil {
		return err
	}

	if role == "admin" {
		var adminCount int64
		db.GetDB().Model(&models.GroupUsers{}).
			Where("group_id = ? AND role_in_group = ?", groupID, "admin").
			Count(&adminCount)

		if adminCount <= 1 {
			return errors.New("нельзя выйти — вы единственный админ")
		}
	}

	return db.GetDB().Where("user_id = ? AND group_id = ?", user.ID, groupID).Delete(&models.GroupUsers{}).Error
}
