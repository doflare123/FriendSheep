package services

import (
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"
	"time"
)

type SessionInput struct {
	Title       string    `form:"title" binding:"required"`
	SessionType uint      `form:"session_type" binding:"required"` // ID!
	GroupID     uint      `form:"group_id" binding:"required"`
	StartTime   time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00" binding:"required"`
	Duration    uint16    `form:"duration"`
	CountUsers  uint16    `form:"count_users" binding:"required"`
	Image       string    `form:"-"` // обрабатывается вручную
}

func CreateSession(email string, input SessionInput) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("не передан jwt")
	}
	if err := ValidateInput(input); err != nil {
		return false, fmt.Errorf("невалидная структура данных: %v", err)
	}

	var creator models.User
	if err := db.GetDB().Where("email = ?", email).First(&creator).Error; err != nil {
		return false, fmt.Errorf("пользователь не найден (%s): %v", email, err)
	}

	var group groups.Group
	if err := db.GetDB().Where("id = ?", input.GroupID).First(&group).Error; err != nil {
		return false, fmt.Errorf("группа не найдена (%d): %v", input.GroupID, err)
	}

	var sessionType sessions.SessionGroupType
	if err := db.GetDB().Where("id = ?", input.SessionType).First(&sessionType).Error; err != nil {
		return false, fmt.Errorf("тип сессии не найден (%v): %v", input.SessionType, err)
	}

	session := sessions.Session{
		Title:         input.Title,
		SessionType:   sessionType,
		GroupID:       input.GroupID,
		StartTime:     input.StartTime,
		Duration:      input.Duration,
		CurrentUsers:  1,
		CountUsersMax: input.CountUsers,
		ImageURL:      input.Image,
		UserID:        creator.ID,
	}
	if err := db.GetDB().Create(&session).Error; err != nil {
		return false, fmt.Errorf("ошибка создания сессии: %v", err)
	}

	return true, nil
}
