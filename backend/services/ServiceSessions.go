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
	"gorm.io/gorm"
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

type SessionJoinInput struct {
	SessionID uint `json:"session_id" binding:"required"`
	GroupID   uint `json:"group_id" binding:"required"`
}

// Валидация времени начала сессии
func (input *SessionInput) ValidateStartTime() error {
	now := time.Now()

	if input.StartTime.Before(now) {
		return fmt.Errorf("время начала сессии не может быть в прошлом")
	}

	maxFutureTime := now.AddDate(1, 0, 0)
	if input.StartTime.After(maxFutureTime) {
		return fmt.Errorf("время начала сессии не может быть более чем через год")
	}

	return nil
}

// Валидация продолжительности
func (input *SessionInput) ValidateDuration() error {
	if input.Duration == 0 {
		return fmt.Errorf("продолжительность сессии обязательна")
	}

	if input.Duration == 0 {
		return fmt.Errorf("продолжительность сессии должна быть больше 0")
	}

	// Максимальная продолжительность 12 часов (720 минут)
	if input.Duration > 720 {
		return fmt.Errorf("продолжительность сессии не может превышать 12 часов")
	}

	return nil
}

// Расчет времени окончания
func (input SessionInput) CalculateEndTime() time.Time {
	if input.Duration <= 0 {
		return input.StartTime
	}
	return input.StartTime.Add(time.Duration(input.Duration) * time.Minute)
}

func CreateSession(email string, input SessionInput) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("не передан jwt")
	}

	// Валидация времени начала
	if err := input.ValidateStartTime(); err != nil {
		return false, fmt.Errorf("ошибка валидации времени: %v", err)
	}

	// Валидация продолжительности
	if err := input.ValidateDuration(); err != nil {
		return false, fmt.Errorf("ошибка валидации продолжительности: %v", err)
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

	// Расчет времени окончания
	endTime := input.CalculateEndTime()

	session := sessions.Session{
		Title:         input.Title,
		SessionType:   sessionType,
		GroupID:       input.GroupID,
		StartTime:     input.StartTime,
		EndTime:       endTime,
		Duration:      input.Duration,
		CurrentUsers:  1,
		CountUsersMax: input.CountUsers,
		ImageURL:      input.Image,
		UserID:        creator.ID,
	}

	if err := db.GetDB().Create(&session).Error; err != nil {
		return false, fmt.Errorf("ошибка создания сессии: %v", err)
	}

	// Обработка метаданных
	var genres []string
	if input.GenresRaw != "" {
		genres = strings.Split(input.GenresRaw, ",")
		// Очищаем от лишних пробелов
		for i, genre := range genres {
			genres[i] = strings.TrimSpace(genre)
		}
	}

	fields := make(map[string]interface{})
	if input.FieldsRaw != "" {
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
			} else if f, err := strconv.ParseFloat(value, 64); err == nil {
				fields[key] = f
			} else {
				fields[key] = value
			}
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
		// Откатываем создание сессии если не удалось сохранить метаданные
		db.GetDB().Delete(&session)
		return false, fmt.Errorf("не удалось сохранить метаданные Mongo: %v", err)
	}

	addNewUserToSession := sessions.SessionUser{
		SessionID: session.ID,
		UserID:    creator.ID,
	}

	if err := db.GetDB().Create(&addNewUserToSession).Error; err != nil {
		return false, fmt.Errorf("ошибка создания сессии: %v", err)
	}

	return true, nil
}

func JoinToSession(email string, input SessionJoinInput) error {
	dbTx := db.GetDB().Begin()

	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	var user models.User
	if err := dbTx.Where("email = ?", email).First(&user).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("пользователь не найден: %v", err)
	}
	var session sessions.Session
	if err := dbTx.Preload("Group").Where("id = ? AND group_id = ?", input.SessionID, input.GroupID).First(&session).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("сессия не найдена: %v", err)
	}
	if session.CurrentUsers >= session.CountUsersMax {
		dbTx.Rollback()
		return fmt.Errorf("сессия заполнена")
	}
	var exists sessions.SessionUser
	if err := dbTx.Where("session_id = ? AND user_id = ?", session.ID, user.ID).First(&exists).Error; err == nil {
		dbTx.Rollback()
		return fmt.Errorf("пользователь уже присоединился к сессии")
	}
	if err := dbTx.Model(&session).Update("current_users", gorm.Expr("current_users + ?", 1)).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("ошибка обновления сессии: %v", err)
	}
	if err := dbTx.Create(&sessions.SessionUser{
		SessionID: session.ID,
		UserID:    user.ID,
	}).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("ошибка добавления пользователя в сессию: %v", err)
	}

	return dbTx.Commit().Error
}

func LeaveSession(email string, sessionID uint) error {
	dbTx := db.GetDB().Begin()

	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	var user models.User
	if err := dbTx.Where("email = ?", email).First(&user).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("пользователь не найден")
	}

	var session sessions.Session
	if err := dbTx.First(&session, sessionID).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("сессия не найдена")
	}

	var sessionUser sessions.SessionUser
	if err := dbTx.Where("session_id = ? AND user_id = ?", sessionID, user.ID).First(&sessionUser).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("вы не состоите в этой сессии")
	}

	if err := dbTx.Delete(&sessionUser).Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("ошибка при выходе из сессии: %v", err)
	}

	if session.CurrentUsers > 0 {
		if err := dbTx.Model(&session).Update("current_users", gorm.Expr("current_users - ?", 1)).Error; err != nil {
			dbTx.Rollback()
			return fmt.Errorf("ошибка при обновлении счётчика сессии: %v", err)
		}
	}

	return dbTx.Commit().Error
}

//admin функционал

type SessionUpdateInput struct {
	Title         *string    `json:"title"`
	StartTime     *time.Time `json:"start_time"`
	Duration      *uint16    `json:"duration"`
	ImageURL      *string    `json:"image_url"`
	CountUsersMax *uint16    `json:"count_users_max"`
}

func DeleteSession(email string, sessionID uint) error {
	dbTx := db.GetDB().Begin()

	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	var user models.User
	if err := dbTx.Where("email = ?", email).First(&user).Error; err != nil {
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

func UpdateSession(email string, sessionID uint, input SessionUpdateInput) error {
	var session sessions.Session
	if err := db.GetDB().Preload("Group").First(&session, sessionID).Error; err != nil {
		return fmt.Errorf("сессия не найдена")
	}

	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}
	isAdmin, err := getUserRole(user.ID, session.GroupID)
	if isAdmin != "admin" {
		return err
	}

	if input.Title != nil {
		session.Title = *input.Title
	}
	if input.StartTime != nil {
		session.StartTime = *input.StartTime
	}
	if input.Duration != nil {
		session.Duration = *input.Duration
	}
	if input.ImageURL != nil {
		session.ImageURL = *input.ImageURL
	}
	if input.CountUsersMax != nil {
		session.CountUsersMax = *input.CountUsersMax
	}

	if err := db.GetDB().Save(&session).Error; err != nil {
		return fmt.Errorf("не удалось обновить сессию: %v", err)
	}

	return nil
}
