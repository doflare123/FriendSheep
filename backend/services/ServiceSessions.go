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
	Title        string    `form:"title" binding:"required"`
	SessionType  string    `form:"session_type" binding:"required"`
	SessionPlace uint      `form:"session_place" binding:"required"`
	GroupID      uint      `form:"group_id" binding:"required"`
	StartTime    time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00" binding:"required"`
	Duration     uint16    `form:"duration"`
	CountUsers   uint16    `form:"count_users" binding:"required"`
	Image        string    `form:"-"`

	GenresRaw string `form:"genres"`
	FieldsRaw string `form:"fields"`
	Location  string `form:"location"`
	Year      *int   `form:"year"`
	Country   string `form:"country"`
	AgeLimit  string `form:"age_limit"`
	Notes     string `form:"notes"`
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
	fmt.Println(session.Group.IsPrivate, "dfkjsldkfjlsdkjflskdjflsdjflsdkjf")

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

func SearchSessions(email, query *string, categoryID *uint, sessionType *string, page *int) (*PaginatedSearchResponse, error) {
	if email == nil || *email == "" {
		return nil, fmt.Errorf("email не может быть пустым")
	}
	if query == nil || *query == "" {
		return nil, fmt.Errorf("query не может быть пустым")
	}
	if categoryID == nil {
		return nil, fmt.Errorf("categoryID обязателен")
	}

	const pageSize = 9
	currentPage := 1
	if page != nil && *page > 0 {
		currentPage = *page
	}

	var filteredSessions []sessions.Session
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

	var userGroups []groups.GroupUsers
	if err := dbTx.Where("user_id = ?", &user.ID).Find(&userGroups).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при получении групп пользователя: %v", err)
	}

	userGroupsMap := make(map[uint]bool)
	for _, ug := range userGroups {
		if ug.GroupID != 0 {
			userGroupsMap[ug.GroupID] = true
		}
	}

	query_db := dbTx.Table("sessions").
		Preload("SessionType").
		Preload("SessionPlace").
		Preload("Group").
		Preload("User").
		Where("session_type_id = ?", *categoryID).
		Where("title LIKE ?", "%"+*query+"%")

	if sessionType != nil && *sessionType != "" {
		var placeIDs []uint
		var places []sessions.SessionGroupPlace

		placeQuery := dbTx.Where("title = ?", *sessionType).Find(&places)
		if placeQuery.Error != nil {
			dbTx.Rollback()
			return nil, fmt.Errorf("ошибка при поиске мест проведения: %v", placeQuery.Error)
		}

		for _, place := range places {
			if place.ID != 0 {
				placeIDs = append(placeIDs, place.ID)
			}
		}

		if len(placeIDs) > 0 {
			query_db = query_db.Where("session_place_id IN ?", placeIDs)
		} else {
			// Если не найдено подходящих мест, возвращаем пустой результат
			dbTx.Commit()
			return &PaginatedSearchResponse{
				Sessions:    []SessionResponse{},
				Page:        currentPage,
				PageSize:    pageSize,
				Total:       0,
				TotalPages:  0,
				HasNext:     false,
				HasPrevious: currentPage > 1,
			}, nil
		}
	}

	var allSessions []sessions.Session
	if err := query_db.Find(&allSessions).Error; err != nil {
		dbTx.Rollback()
		return nil, fmt.Errorf("ошибка при поиске сессий: %v", err)
	}

	// Фильтруем сессии в зависимости от приватности группы
	for _, session := range allSessions {
		// Проверяем, что группа загружена
		if session.GroupID == 0 {
			continue
		}

		var group groups.Group
		if err := dbTx.Where("id = ?", &session.GroupID).First(&group).Error; err != nil {
			continue
		}

		if group.IsPrivate {
			if userGroupsMap[session.GroupID] {
				filteredSessions = append(filteredSessions, session)
			}
		} else {
			filteredSessions = append(filteredSessions, session)
		}
	}

	if err := dbTx.Commit().Error; err != nil {
		return nil, fmt.Errorf("ошибка при коммите транзакции: %v", err)
	}

	totalSessions := len(filteredSessions)
	totalPages := (totalSessions + pageSize - 1) / pageSize

	if totalPages == 0 {
		totalPages = 1
	}

	if currentPage > totalPages {
		currentPage = totalPages
	}

	startIndex := (currentPage - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= totalSessions {
		return &PaginatedSearchResponse{
			Sessions:    []SessionResponse{},
			Page:        currentPage,
			PageSize:    pageSize,
			Total:       totalSessions,
			TotalPages:  totalPages,
			HasNext:     false,
			HasPrevious: currentPage > 1,
		}, nil
	}

	if endIndex > totalSessions {
		endIndex = totalSessions
	}

	sessionsPaged := filteredSessions[startIndex:endIndex]

	sessionIDs := make([]uint, 0, len(sessionsPaged))
	for _, session := range sessionsPaged {
		if session.ID != 0 {
			sessionIDs = append(sessionIDs, session.ID)
		}
	}

	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
	if err != nil {
		// Логируем ошибку, но не прерываем выполнение - можем вернуть данные без метаданных
		fmt.Printf("Ошибка при получении метаданных из MongoDB: %v\n", err)
		metadataMap = make(map[uint]*sessions.SessionMetadata)
	}

	result := make([]SessionResponse, 0, len(sessionsPaged))
	for _, session := range sessionsPaged {
		response := SessionResponse{
			ID:            session.ID,
			Title:         session.Title,
			StartTime:     session.StartTime,
			ImageURL:      session.ImageURL,
			Duration:      session.Duration,
			CurrentUsers:  session.CurrentUsers,
			CountUsersMax: session.CountUsersMax,
		}

		if session.SessionType.Name != "" {
			response.SessionType = session.SessionType.Name
		}

		if session.SessionPlace.Title != "" {
			response.SessionPlace = session.SessionPlace.Title
		}

		if session.Group.Name != "" {
			response.GroupName = session.Group.Name
		}

		if session.ID != 0 {
			if metadata, exists := metadataMap[session.ID]; exists && metadata != nil {
				// Добавляем жанры из метаданных
				if metadata.Genres != nil {
					response.Genres = metadata.Genres
				} else {
					response.Genres = make([]string, 0)
				}
			} else {
				response.Genres = make([]string, 0)
			}
		} else {
			response.Genres = make([]string, 0)
		}

		result = append(result, response)
	}

	return &PaginatedSearchResponse{
		Sessions:    result,
		Page:        currentPage,
		PageSize:    pageSize,
		Total:       totalSessions,
		TotalPages:  totalPages,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}, nil
}

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

	// Проверяем доступ к сессии через группу
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
