package services

import (
	"context"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SessionInput struct {
	Title       string    `form:"title" binding:"required"`
	SessionType uint      `form:"session_type" binding:"required"`
	GroupID     uint      `form:"group_id" binding:"required"`
	StartTime   time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00" binding:"required"`
	Duration    uint16    `form:"duration"`
	CountUsers  uint16    `form:"count_users" binding:"required"`
	Image       string    `form:"-"`

	MetaType  string `form:"meta_type"`
	GenresRaw string `form:"genres"`
	FieldsRaw string `form:"fields"`
	Location  string `form:"location"`
	Year      *int   `form:"year"`
	Country   string `form:"country"`
	AgeLimit  string `form:"age_limit"`
	Notes     string `form:"notes"`
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

	genres := strings.Split(input.GenresRaw, ",")
	fields := make(map[string]interface{})

	for _, pair := range strings.Split(input.FieldsRaw, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Попробуем определить тип
		if value == "true" || value == "false" {
			fields[key] = (value == "true")
		} else if i, err := strconv.Atoi(value); err == nil {
			fields[key] = i
		} else {
			fields[key] = value
		}
	}

	meta := sessions.SessionMetadata{
		SessionID: session.ID,
		Type:      input.MetaType,
		Fields:    fields,
		Location:  input.Location,
		Genres:    genres,
		Year:      input.Year,
		Country:   input.Country,
		AgeLimit:  input.AgeLimit,
		Notes:     input.Notes,
	}

	collection := db.GetMongoDB().Collection("session_metadata")
	_, err := collection.InsertOne(context.TODO(), meta)
	if err != nil {
		return false, fmt.Errorf("не удалось сохранить метаданные Mongo: %v", err)
	}

	return true, nil
}

func DeleteSession(userEmail string, sessionID uint) error {
	dbTx := db.GetDB().Begin()

	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	var user models.User
	if err := dbTx.Where("email = ?", userEmail).First(&user).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("пользователь не найден: %v", err)
	}

	var session sessions.Session
	if err := dbTx.Preload("Group").Where("id = ?", sessionID).First(&session).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("сессия не найдена: %v", err)
	}

	var member groups.GroupUsers
	if err := dbTx.Where("user_id = ? AND group_id = ?", user.ID, session.GroupID).First(&member).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("вы не состоите в группе или нет доступа")
	}
	if member.RoleInGroup != "admin" && member.RoleInGroup != "operator" {
		dbTx.Rollback()
		return fmt.Errorf("у вас нет прав на удаление этой сессии")
	}

	if err := dbTx.Where("session_id = ?", session.ID).Delete(&sessions.SessionUser{}).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("не удалось удалить пользователей сессии: %v", err)
	}

	if err := dbTx.Where("session_id = ?", session.ID).Delete(&sessions.Notification{}).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("не удалось удалить уведомления: %v", err)
	}

	if err := dbTx.Delete(&session).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("не удалось удалить сессию: %v", err)
	}

	mongoColl := db.GetMongoDB().Collection("session_metadata")
	_, err := mongoColl.DeleteOne(context.TODO(), bson.M{"session_id": session.ID})
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("не удалось удалить метаданные Mongo: %v", err)
	}

	if session.ImageURL != "" {
		filename := filepath.Base(session.ImageURL)
		imagePath := filepath.Join("uploads", filename)
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			dbTx.Rollback()
			return fmt.Errorf("не удалось удалить изображение: %v", err)
		}
	}

	return dbTx.Commit().Error
}
