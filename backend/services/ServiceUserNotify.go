package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"
	"time"
)

type NotificationDTO struct {
	ID     uint      `json:"id"`
	Type   string    `json:"type"`
	SendAt time.Time `json:"sendAt"`
	Sent   bool      `json:"sent"`
	Viewed bool      `json:"viewed"`
}

type InviteDTO struct {
	ID        uint      `json:"id"`
	GroupID   uint      `json:"groupId"`
	GroupName string    `json:"groupName"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type GetNotifyResponse struct {
	Notifications []NotificationDTO `json:"notifications"`
	Invites       []InviteDTO       `json:"invites"`
}

func GetNotify(email string) (*GetNotifyResponse, error) {
	database := db.GetDB()

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("пользователь не найден")
	}

	var notifications []sessions.Notification
	if err := database.
		Preload("NotificationType").
		Where("user_id = ? AND viewed = ?", user.ID, false).
		Find(&notifications).Error; err != nil {
		return nil, err
	}

	var invites []groups.GroupJoinInvite
	if err := database.
		Preload("Group").
		Where("user_id = ? AND status = ?", user.ID, "pending").
		Find(&invites).Error; err != nil {
		return nil, err
	}

	var notifDTOs []NotificationDTO
	for _, n := range notifications {
		notifDTOs = append(notifDTOs, NotificationDTO{
			ID:     n.ID,
			Type:   n.NotificationType.Label,
			SendAt: n.SendAt,
			Sent:   n.Sent,
			Viewed: n.Viewed,
		})
	}

	var inviteDTOs []InviteDTO
	for _, i := range invites {
		inviteDTOs = append(inviteDTOs, InviteDTO{
			ID:        i.ID,
			GroupID:   i.GroupID,
			GroupName: i.Group.Name,
			Status:    i.Status,
			CreatedAt: i.CreatedAt,
		})
	}

	return &GetNotifyResponse{
		Notifications: notifDTOs,
		Invites:       inviteDTOs,
	}, nil
}

func GetNotifyInf(email string) (bool, error) {
	database := db.GetDB()

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		return false, errors.New("пользователь не найден")
	}

	var notifications []sessions.Notification
	if err := database.
		Preload("NotificationType").
		Where("user_id = ? AND viewed = ?", user.ID, false).
		Find(&notifications).Error; err != nil {
		return false, err
	}

	var invites []groups.GroupJoinInvite
	if err := database.
		Preload("Group").
		Where("user_id = ? AND status = ?", user.ID, "pending").
		Find(&invites).Error; err != nil {
		return false, err
	}

	hasNotifications := len(notifications) > 0 || len(invites) > 0
	return hasNotifications, nil
}

func MarkNotificationViewed(notificationID uint, userID uint) error {
	database := db.GetDB()

	var notif sessions.Notification
	if err := database.Where("id = ? AND user_id = ?", notificationID, userID).First(&notif).Error; err != nil {
		return errors.New("уведомление не найдено")
	}

	notif.Viewed = true
	return database.Save(&notif).Error
}

func ApproveInviteByID(email string, inviteID uint) error {
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
		tx.Rollback()
		return errors.New("пользователь не найден")
	}

	var invite groups.GroupJoinInvite
	if err := tx.Preload("Group").First(&invite, inviteID).Error; err != nil {
		tx.Rollback()
		return errors.New("приглашение не найдено")
	}

	if invite.UserID != user.ID {
		tx.Rollback()
		return errors.New("у вас нет доступа к этому приглашению")
	}

	if invite.Status != "pending" {
		tx.Rollback()
		return errors.New("приглашение уже обработано")
	}

	if err := tx.Model(&invite).Update("status", "approved").Error; err != nil {
		tx.Rollback()
		return err
	}

	member := groups.GroupUsers{
		UserID:      invite.UserID,
		GroupID:     invite.GroupID,
		RoleInGroup: "member",
	}
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func RejectInviteByID(email string, inviteID uint) error {
	database := db.GetDB()

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}

	var invite groups.GroupJoinInvite
	if err := database.Preload("Group").First(&invite, inviteID).Error; err != nil {
		return errors.New("приглашение не найдено")
	}

	if invite.UserID != user.ID {
		return errors.New("у вас нет доступа к этому приглашению")
	}

	if invite.Status != "pending" {
		return errors.New("приглашение уже обработано")
	}

	return database.Model(&invite).Update("status", "rejected").Error
}
