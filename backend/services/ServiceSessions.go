package services

import (
	"context"
	"fmt"
	"friendship/db"
	"friendship/models"
	"friendship/models/groups"
	"friendship/models/sessions"
	statsusers "friendship/models/stats_users"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gorm.io/gorm"
)

type SessionInput struct {
	Title        string    `form:"title" binding:"required,min=5,max=40"`
	SessionType  string    `form:"session_type" binding:"required"`
	SessionPlace uint      `form:"session_place" binding:"required"`
	GroupID      uint      `form:"group_id" binding:"required"`
	StartTime    time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00" binding:"required"`
	Duration     uint16    `form:"duration"`
	CountUsers   uint16    `form:"count_users" binding:"required"`
	Image        string    `form:"image" binding:"required"`

	GenresRaw string `form:"genres"`
	FieldsRaw string `form:"fields"`
	Location  string `form:"location"`
	Year      *int   `form:"year"`
	Country   string `form:"country"`
	AgeLimit  string `form:"age_limit"`
	Notes     string `form:"notes" binding:"min=0,max=300"`
}

type SearchSessionsRequest struct {
	Query       string  `form:"query" binding:"required" validate:"min=1,max=100"`
	CategoryID  uint    `form:"categoryID" binding:"required"`
	SessionType *string `form:"session_type,omitempty" validate:"omitempty,oneof=Оффлайн Онлайн"`
	Page        *int    `form:"Page" json:"page"`
}

type SessionJoinInput struct {
	SessionID uint `json:"session_id" binding:"required"`
	GroupID   uint `json:"group_id" binding:"required"`
}

type SessionDetailResponse struct {
	Session  SubSessionDetail          `json:"session"`
	Metadata *sessions.SessionMetadata `json:"metadata,omitempty"`
}

type SubSessionDetail struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	SessionType   string    `json:"session_type"`
	SessionPlace  string    `json:"session_place"`
	GroupID       uint      `json:"group_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Duration      uint16    `json:"duration"`
	CurrantUsers  uint16    `json:"current_users"`
	CountUsersMax uint16    `json:"count_users_max"`
	ImageURL      string    `json:"image_url"`
}

type PaginatedSearchResponse struct {
	Sessions    []SessionResponse `json:"sessions"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	Total       int               `json:"total"`
	TotalPages  int               `json:"total_pages"`
	HasNext     bool              `json:"has_next"`
	HasPrevious bool              `json:"has_previous"`
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

	var sessionPlace sessions.SessionGroupPlace
	if err := db.GetDB().Where("id = ?", input.SessionPlace).First(&sessionPlace).Error; err != nil {
		return false, fmt.Errorf("тип сессии(проведения) не найден (%v): %v", input.SessionPlace, err)
	}

	var sessionType models.Category
	if err := db.GetDB().Where("Name = ?", input.SessionType).First(&sessionType).Error; err != nil {
		return false, fmt.Errorf("тип сессии не найден (%v): %v", input.SessionType, err)
	}

	// Расчет времени окончания
	endTime := input.CalculateEndTime()

	session := sessions.Session{
		Title:          input.Title,
		SessionTypeID:  sessionType.ID,
		SessionPlaceID: sessionPlace.ID,
		GroupID:        input.GroupID,
		StartTime:      input.StartTime,
		EndTime:        endTime,
		Duration:       input.Duration,
		CurrentUsers:   1,
		CountUsersMax:  input.CountUsers,
		ImageURL:       input.Image,
		UserID:         creator.ID,
		StatusID:       1,
	}
	fmt.Print("session", "session", session)

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

func JoinToSession(email *string, input SessionJoinInput) error {
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

	if session.Group.IsPrivate == true {
		var groupUser groups.GroupUsers
		if err := dbTx.Where("group_id = ? AND user_id = ?", session.GroupID, user.ID).First(&groupUser).Error; err != nil {
			dbTx.Rollback()
			return fmt.Errorf("пользователь не является участником приватной группы")
		}
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
	Title          *string    `json:"title"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Duration       *uint16    `json:"duration"`
	ImageURL       *string    `json:"image_url"`
	CountUsersMax  *uint16    `json:"count_users_max"`
	Notes          *string    `json:"notes"`
	Location       *string    `json:"location"`
	Genres         *[]string  `json:"genres"`
	Year           *int       `json:"year"`
	Country        *string    `json:"country"`
	AgeLimit       *string    `json:"age_limit"`
	SessionTypeID  *uint      `json:"session_type_id"`
	SessionPlaceID *uint      `json:"session_place_id"`
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
	var ses sessions.Session
	if err := db.GetDB().Preload("Group").First(&ses, sessionID).Error; err != nil {
		return fmt.Errorf("сессия не найдена: %v", err)
	}

	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return fmt.Errorf("пользователь не найден")
	}

	role, err := getUserRole(user.ID, ses.GroupID)
	if err != nil {
		return fmt.Errorf("ошибка проверки роли: %v", err)
	}
	if role != "admin" {
		return fmt.Errorf("доступ запрещён: только администратор может редактировать сессию")
	}

	if input.Title != nil {
		ses.Title = *input.Title
	}
	if input.StartTime != nil {
		ses.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		ses.EndTime = *input.EndTime
	}
	if input.Duration != nil {
		ses.Duration = *input.Duration
	}
	if input.ImageURL != nil {
		ses.ImageURL = *input.ImageURL
	}
	if input.CountUsersMax != nil {
		ses.CountUsersMax = *input.CountUsersMax
	}
	if input.SessionTypeID != nil {
		ses.SessionTypeID = *input.SessionTypeID
	}
	if input.SessionPlaceID != nil {
		ses.SessionPlaceID = *input.SessionPlaceID
	}

	if err := db.GetDB().Save(&ses).Error; err != nil {
		return fmt.Errorf("не удалось обновить сессию: %v", err)
	}

	coll := db.GetMongoDB().Collection("session_metadata")

	update := bson.M{}

	if input.Notes != nil {
		update["notes"] = *input.Notes
	}
	if input.Location != nil {
		update["location"] = *input.Location
	}
	if input.Genres != nil {
		update["genres"] = *input.Genres
	}
	if input.Year != nil {
		update["year"] = *input.Year
	}
	if input.Country != nil {
		update["country"] = *input.Country
	}
	if input.AgeLimit != nil {
		update["age_limit"] = *input.AgeLimit
	}

	if len(update) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"session_id": uint(sessionID)}

		opts := options.Update().SetUpsert(false)

		res, err := coll.UpdateOne(ctx, filter, bson.M{"$set": update}, opts)
		if err != nil {
			return fmt.Errorf("не удалось обновить метаданные в MongoDB: %v", err)
		}

		if res.MatchedCount == 0 {
			return fmt.Errorf("метаданные для сессии %d не найдены", sessionID)
		}
	}

	return nil
}

// func SearchSessions(email, query *string, categoryID *uint, sessionType *string, page *int) (*PaginatedSearchResponse, error) {
// 	if email == nil || *email == "" {
// 		return nil, fmt.Errorf("email не может быть пустым")
// 	}
// 	if query == nil || *query == "" {
// 		return nil, fmt.Errorf("query не может быть пустым")
// 	}
// 	if categoryID == nil {
// 		return nil, fmt.Errorf("categoryID обязателен")
// 	}

// 	const pageSize = 9
// 	currentPage := 1
// 	if page != nil && *page > 0 {
// 		currentPage = *page
// 	}

// 	var filteredSessions []sessions.Session
// 	dbTx := db.GetDB().Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			dbTx.Rollback()
// 		}
// 	}()

// 	var user models.User
// 	if err := dbTx.Where("email = ?", *email).First(&user).Error; err != nil {
// 		dbTx.Rollback()
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, fmt.Errorf("пользователь не найден")
// 		}
// 		return nil, fmt.Errorf("ошибка при поиске пользователя: %v", err)
// 	}

// 	if user.ID == 0 {
// 		dbTx.Rollback()
// 		return nil, fmt.Errorf("некорректный ID пользователя")
// 	}

// 	var userGroups []groups.GroupUsers
// 	if err := dbTx.Where("user_id = ?", &user.ID).Find(&userGroups).Error; err != nil {
// 		dbTx.Rollback()
// 		return nil, fmt.Errorf("ошибка при получении групп пользователя: %v", err)
// 	}

// 	userGroupsMap := make(map[uint]bool)
// 	for _, ug := range userGroups {
// 		if ug.GroupID != 0 {
// 			userGroupsMap[ug.GroupID] = true
// 		}
// 	}

// 	query_db := dbTx.Table("sessions").
// 		Preload("SessionType").
// 		Preload("SessionPlace").
// 		Preload("Group").
// 		Preload("User").
// 		Where("session_type_id = ?", *categoryID).
// 		Where("title LIKE ?", "%"+*query+"%")

// 	if sessionType != nil && *sessionType != "" {
// 		var placeIDs []uint
// 		var places []sessions.SessionGroupPlace

// 		placeQuery := dbTx.Where("title = ?", *sessionType).Find(&places)
// 		if placeQuery.Error != nil {
// 			dbTx.Rollback()
// 			return nil, fmt.Errorf("ошибка при поиске мест проведения: %v", placeQuery.Error)
// 		}

// 		for _, place := range places {
// 			if place.ID != 0 {
// 				placeIDs = append(placeIDs, place.ID)
// 			}
// 		}

// 		if len(placeIDs) > 0 {
// 			query_db = query_db.Where("session_place_id IN ?", placeIDs)
// 		} else {
// 			// Если не найдено подходящих мест, возвращаем пустой результат
// 			dbTx.Commit()
// 			return &PaginatedSearchResponse{
// 				Sessions:    []SessionResponse{},
// 				Page:        currentPage,
// 				PageSize:    pageSize,
// 				Total:       0,
// 				TotalPages:  0,
// 				HasNext:     false,
// 				HasPrevious: currentPage > 1,
// 			}, nil
// 		}
// 	}

// 	var allSessions []sessions.Session
// 	if err := query_db.Find(&allSessions).Error; err != nil {
// 		dbTx.Rollback()
// 		return nil, fmt.Errorf("ошибка при поиске сессий: %v", err)
// 	}

// 	// Фильтруем сессии в зависимости от приватности группы
// 	for _, session := range allSessions {
// 		// Проверяем, что группа загружена
// 		if session.GroupID == 0 {
// 			continue
// 		}

// 		var group groups.Group
// 		if err := dbTx.Where("id = ?", &session.GroupID).First(&group).Error; err != nil {
// 			continue
// 		}

// 		if group.IsPrivate {
// 			if userGroupsMap[session.GroupID] {
// 				filteredSessions = append(filteredSessions, session)
// 			}
// 		} else {
// 			filteredSessions = append(filteredSessions, session)
// 		}
// 	}

// 	if err := dbTx.Commit().Error; err != nil {
// 		return nil, fmt.Errorf("ошибка при коммите транзакции: %v", err)
// 	}

// 	totalSessions := len(filteredSessions)
// 	totalPages := (totalSessions + pageSize - 1) / pageSize

// 	if totalPages == 0 {
// 		totalPages = 1
// 	}

// 	if currentPage > totalPages {
// 		currentPage = totalPages
// 	}

// 	startIndex := (currentPage - 1) * pageSize
// 	endIndex := startIndex + pageSize

// 	if startIndex >= totalSessions {
// 		return &PaginatedSearchResponse{
// 			Sessions:    []SessionResponse{},
// 			Page:        currentPage,
// 			PageSize:    pageSize,
// 			Total:       totalSessions,
// 			TotalPages:  totalPages,
// 			HasNext:     false,
// 			HasPrevious: currentPage > 1,
// 		}, nil
// 	}

// 	if endIndex > totalSessions {
// 		endIndex = totalSessions
// 	}

// 	sessionsPaged := filteredSessions[startIndex:endIndex]

// 	sessionIDs := make([]uint, 0, len(sessionsPaged))
// 	for _, session := range sessionsPaged {
// 		if session.ID != 0 {
// 			sessionIDs = append(sessionIDs, session.ID)
// 		}
// 	}

// 	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
// 	if err != nil {
// 		// Логируем ошибку, но не прерываем выполнение - можем вернуть данные без метаданных
// 		fmt.Printf("Ошибка при получении метаданных из MongoDB: %v\n", err)
// 		metadataMap = make(map[uint]*sessions.SessionMetadata)
// 	}

// 	result := make([]SessionResponse, 0, len(sessionsPaged))
// 	for _, session := range sessionsPaged {
// 		response := SessionResponse{
// 			ID:            session.ID,
// 			Title:         session.Title,
// 			StartTime:     session.StartTime,
// 			ImageURL:      session.ImageURL,
// 			Duration:      session.Duration,
// 			CurrentUsers:  session.CurrentUsers,
// 			CountUsersMax: session.CountUsersMax,
// 		}

// 		if session.SessionType.Name != "" {
// 			response.SessionType = session.SessionType.Name
// 		}

// 		if session.SessionPlace.Title != "" {
// 			response.SessionPlace = session.SessionPlace.Title
// 		}

// 		if session.Group.Name != "" {
// 			response.GroupName = session.Group.Name
// 		}

// 		if session.ID != 0 {
// 			if metadata, exists := metadataMap[session.ID]; exists && metadata != nil {
// 				// Добавляем жанры из метаданных
// 				if metadata.Genres != nil {
// 					response.Genres = metadata.Genres
// 				} else {
// 					response.Genres = make([]string, 0)
// 				}
// 			} else {
// 				response.Genres = make([]string, 0)
// 			}
// 		} else {
// 			response.Genres = make([]string, 0)
// 		}

// 		result = append(result, response)
// 	}

// 	return &PaginatedSearchResponse{
// 		Sessions:    result,
// 		Page:        currentPage,
// 		PageSize:    pageSize,
// 		Total:       totalSessions,
// 		TotalPages:  totalPages,
// 		HasNext:     currentPage < totalPages,
// 		HasPrevious: currentPage > 1,
// 	}, nil
// }

func GetInfoAboutSession(email *string, sessionID *uint) (*SessionDetailResponse, error) {
	if email == nil || *email == "" {
		return nil, fmt.Errorf("email не может быть пустым")
	}
	if sessionID == nil || *sessionID == 0 {
		return nil, fmt.Errorf("sessionID должен быть больше 0")
	}

	var sessionInf SessionDetailResponse
	dbTx := db.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	var user models.User
	if err := dbTx.Where("email = ?", *email).First(&user).Error; err != nil {
		dbTx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка при поиске пользователя: %v", err)
	}

	if user.ID == 0 {
		dbTx.Rollback()
		return nil, fmt.Errorf("некорректный ID пользователя")
	}

	var session sessions.Session
	if err := dbTx.Preload("SessionType").
		Preload("SessionPlace").
		First(&session, *sessionID).Error; err != nil {
		dbTx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("сессия не найдена")
		}
		return nil, fmt.Errorf("ошибка при поиске сессии: %v", err)
	}

	var group struct {
		IsPrivate bool `gorm:"column:is_private"`
	}
	if err := dbTx.Table("groups").Select("is_private").Where("id = ?", session.GroupID).First(&group).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при получении информации о группе: %v", err)
	}

	if group.IsPrivate {
		var membership struct {
			UserID  *uint `gorm:"column:user_id"`
			GroupID *uint `gorm:"column:group_id"`
		}
		err := dbTx.Table("group_users").
			Select("user_id, group_id").
			Where("user_id = ? AND group_id = ?", &user.ID, session.GroupID).
			First(&membership).Error
		if err != nil {
			dbTx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("доступ запрещен: пользователь не является членом приватной группы")
			}
			return nil, fmt.Errorf("ошибка при проверке членства в группе: %v", err)
		}
	}

	subIng := SubSessionDetail{
		ID:            session.ID,
		Title:         session.Title,
		SessionType:   session.SessionType.Name,
		SessionPlace:  session.SessionPlace.Title,
		GroupID:       session.GroupID,
		StartTime:     session.StartTime,
		EndTime:       session.EndTime,
		Duration:      session.Duration,
		CurrantUsers:  session.CurrentUsers,
		CountUsersMax: session.CountUsersMax,
		ImageURL:      session.ImageURL,
	}

	sessionInf.Session = subIng

	metadata, err := db.GetSessionMetadataId(sessionID)
	if err != nil {
		fmt.Printf("Ошибка при получении метаданных: %v\n", err)
	}
	sessionInf.Metadata = metadata

	if err := dbTx.Commit().Error; err != nil {
		return nil, fmt.Errorf("ошибка при сохранении транзакции: %v", err)
	}

	return &sessionInf, nil
}

func GetSessionsUserGroups(email *string, page int) (*GetNewSessionsResponse, error) {
	if *email == "" {
		return nil, fmt.Errorf("email не может быть пустым")
	}
	if page < 1 {
		page = 1
	}
	const limit = 9
	offset := (page - 1) * limit
	var sessionsUser []SessionResponse

	dbTx := db.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	var user models.User
	if err := dbTx.Where("email = ?", *email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			dbTx.Rollback()
			return nil, fmt.Errorf("пользователь не найден")
		}
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при поиске пользователя: %v", err)
	}

	var groupMemberships []struct {
		GroupID *uint `gorm:"column:group_id"`
	}
	if err := dbTx.Table("group_users").
		Select("group_id").
		Where("user_id = ?", user.ID).
		Find(&groupMemberships).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при получении групп пользователя: %v", err)
	}

	if len(groupMemberships) == 0 {
		dbTx.Commit()
		return &GetNewSessionsResponse{
			Sessions: sessionsUser,
			HasMore:  false,
			Page:     page,
			Total:    0,
		}, nil
	}

	var groupIDs []*uint
	for _, membership := range groupMemberships {
		groupIDs = append(groupIDs, membership.GroupID)
	}

	var recruitmentStatus sessions.Status
	if err := dbTx.Where("status = ?", "Набор").First(&recruitmentStatus).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("статус 'Набор' не найден: %v", err)
	}

	var totalCount int64
	if err := dbTx.Model(&sessions.Session{}).
		Where("group_id IN ? AND status_id = ?", groupIDs, recruitmentStatus.ID).
		Count(&totalCount).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при подсчете общего количества сессий: %v", err)
	}

	var sessionModels []sessions.Session
	query := dbTx.Preload("SessionType").
		Preload("SessionPlace").
		Preload("Status").
		Preload("User").
		Preload("Group").
		Where("group_id IN ? AND status_id = ?", groupIDs, recruitmentStatus.ID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&sessionModels).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при получении сессий: %v", err)
	}

	var sessionIDs []uint
	for _, session := range sessionModels {
		sessionIDs = append(sessionIDs, session.ID)
	}

	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
	if err != nil {
		fmt.Printf("Предупреждение: не удалось загрузить метаданные сессий: %v\n", err)
		metadataMap = make(map[uint]*sessions.SessionMetadata)
	}

	for _, session := range sessionModels {
		var genres []string
		var groupName *string
		var city string

		if metadata, exists := metadataMap[session.ID]; exists {
			genres = metadata.Genres
			if metadata.Fields != nil {
				if v, ok := metadata.Fields["city"]; ok {
					if strCity, ok := v.(string); ok {
						city = strCity
					}
				}
			}
		}

		groupName = &session.Group.Name

		sessionResponse := SessionResponse{
			ID:            session.ID,
			Title:         session.Title,
			SessionType:   session.SessionType.Name,
			SessionPlace:  session.SessionPlace.Title,
			StartTime:     session.StartTime,
			Duration:      session.Duration,
			CurrentUsers:  session.CurrentUsers,
			CountUsersMax: session.CountUsersMax,
			ImageURL:      session.ImageURL,
			Genres:        genres,
			GroupName:     *groupName,
			City:          &city,
		}
		sessionsUser = append(sessionsUser, sessionResponse)
	}

	if err := dbTx.Commit().Error; err != nil {
		return nil, fmt.Errorf("ошибка при коммите транзакции: %v", err)
	}

	hasMore := int64(offset+len(sessionModels)) < totalCount

	return &GetNewSessionsResponse{
		Sessions: sessionsUser,
		HasMore:  hasMore,
		Page:     page,
		Total:    int64(totalCount),
	}, nil
}

func GetGenres() ([]string, error) {
	var genres []statsusers.Genre
	result := db.GetDB().Find(&genres)
	if result.Error != nil {
		return nil, result.Error
	}

	names := make([]string, len(genres))
	for i, g := range genres {
		names[i] = g.Name
	}

	return names, nil
}
