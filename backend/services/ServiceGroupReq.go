package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"

	"gorm.io/gorm"
)

// доп функция на проверку админки
func checkAdminPermissions(db *gorm.DB, email string, groupID uint) (*models.User, error) {
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("пользователь не найден")
	}

	var adminCheck groups.GroupUsers
	if err := db.Where("user_id = ? AND group_id = ? AND role_in_group = ?", user.ID, groupID, "admin").First(&adminCheck).Error; err != nil {
		return nil, errors.New("вы не администратор этой группы")
	}
	return &user, nil
}

func GetPendingJoinRequestsForAdmin(email string) ([]GroupJoinRequestRes, error) {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("пользователь не найден")
	}

	var requests []groups.GroupJoinRequest
	err := db.GetDB().
		Table("group_join_requests").
		Joins("JOIN groups ON group_join_requests.group_id = groups.id").
		Joins("JOIN group_users ON group_users.group_id = groups.id").
		Where("group_users.user_id = ? AND group_users.role_in_group = ? AND group_join_requests.status = ?", user.ID, "admin", "pending").
		Preload("User").
		Preload("Group").
		Find(&requests).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}

	var result []GroupJoinRequestRes
	for _, request := range requests {
		res := GroupJoinRequestRes{
			ID:     request.ID,
			UserID: request.UserID,
		}

		if request.User.Name != "" {
			res.Name = request.User.Name
		}
		if request.User.Us != "" {
			res.Us = request.User.Us
		}
		if request.User.Image != "" {
			res.Image = request.User.Image
		}

		result = append(result, res)
	}

	return result, nil
}

func ApproveJoinRequestByID(email string, requestID string) error {
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var request groups.GroupJoinRequest
	if err := tx.Preload("Group").First(&request, requestID).Error; err != nil {
		tx.Rollback()
		return errors.New("заявка не найдена")
	}

	if request.Status != "pending" {
		tx.Rollback()
		return errors.New("заявка уже обработана")
	}

	if _, err := checkAdminPermissions(tx, email, request.GroupID); err != nil {
		tx.Rollback()
		return err
	}

	// Обновить статус
	if err := tx.Model(&request).Update("status", "approved").Error; err != nil {
		tx.Rollback()
		return err
	}

	member := groups.GroupUsers{
		UserID:      request.UserID,
		GroupID:     request.GroupID,
		RoleInGroup: "member",
	}
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func RejectJoinRequestByID(email string, requestID string) error {
	db := db.GetDB()
	var request groups.GroupJoinRequest
	if err := db.Preload("Group").First(&request, requestID).Error; err != nil {
		return errors.New("заявка не найдена")
	}

	if request.Status != "pending" {
		return errors.New("заявка уже обработана")
	}

	if _, err := checkAdminPermissions(db, email, request.GroupID); err != nil {
		return err
	}

	return db.Model(&request).Update("status", "rejected").Error
}

// ApproveAllJoinRequests одобряет все ожидающие заявки для указанной группы.
func ApproveAllJoinRequests(email string, groupID uint) error {
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if _, err := checkAdminPermissions(tx, email, groupID); err != nil {
		tx.Rollback()
		return err
	}

	var requests []groups.GroupJoinRequest
	if err := tx.Where("group_id = ? AND status = ?", groupID, "pending").Find(&requests).Error; err != nil {
		tx.Rollback()
		return errors.New("ошибка при поиске заявок")
	}

	if len(requests) == 0 {
		tx.Rollback()
		return errors.New("нет ожидающих заявок для этой группы")
	}

	newMembers := make([]groups.GroupUsers, 0, len(requests))
	for _, req := range requests {
		newMembers = append(newMembers, groups.GroupUsers{
			UserID:      req.UserID,
			GroupID:     req.GroupID,
			RoleInGroup: "member",
		})
	}

	if err := tx.Create(&newMembers).Error; err != nil {
		tx.Rollback()
		return errors.New("ошибка при добавлении новых участников")
	}

	if err := tx.Model(&groups.GroupJoinRequest{}).Where("group_id = ? AND status = ?", groupID, "pending").Update("status", "approved").Error; err != nil {
		tx.Rollback()
		return errors.New("ошибка при обновлении статуса заявок")
	}

	return tx.Commit().Error
}

// RejectAllJoinRequests отклоняет все ожидающие заявки для указанной группы.
func RejectAllJoinRequests(email string, groupID uint) error {
	db := db.GetDB()
	if _, err := checkAdminPermissions(db, email, groupID); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&groups.GroupJoinRequest{}).Where("group_id = ? AND status = ?", groupID, "pending").Count(&count).Error; err != nil {
		return errors.New("ошибка при проверке заявок")
	}

	if count == 0 {
		return errors.New("нет ожидающих заявок для этой группы")
	}

	if err := db.Model(&groups.GroupJoinRequest{}).Where("group_id = ? AND status = ?", groupID, "pending").Update("status", "rejected").Error; err != nil {
		return errors.New("ошибка при отклонении заявок")
	}

	return nil
}
