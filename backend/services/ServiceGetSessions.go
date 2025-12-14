package services

// import (
// 	"fmt"
// 	"friendship/db"
// 	"friendship/models"
// 	"friendship/models/sessions"
// 	"log"
// 	"sort"
// 	"time"
// )

// type GetNewSessionsResponse struct {
// 	Sessions []SessionResponse `json:"sessions"`
// 	HasMore  bool              `json:"has_more"`
// 	Page     int               `json:"page"`
// 	Total    int64             `json:"total"`
// }

// type SessionResponse struct {
// 	ID            uint      `json:"id"`
// 	Title         string    `json:"title"`
// 	StartTime     time.Time `json:"start_time"`
// 	SessionType   string    `json:"session_type"`
// 	ImageURL      string    `json:"image_url"`
// 	SessionPlace  string    `json:"session_place"`
// 	Genres        []string  `json:"genres"`
// 	Duration      uint16    `json:"duration"`
// 	CurrentUsers  uint16    `json:"current_users"`
// 	CountUsersMax uint16    `json:"count_users_max"`
// 	GroupName     string    `json:"group_name"`
// 	GroupID       uint      `json:"group_id"`
// 	City          *string   `json:"city" default:"prikol"`
// }

// // func getSessionGenres(sessionID uint) ([]string, error) {
// // 	mongoDB := db.GetMongoDB()
// // 	if mongoDB == nil {
// // 		return []string{}, fmt.Errorf("MongoDB не инициализирована")
// // 	}

// // 	collection := mongoDB.Collection("session_metadata")
// // 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// // 	defer cancel()

// // 	var metadata sessions.SessionMetadata
// // 	err := collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&metadata)
// // 	if err != nil {
// // 		if err == mongo.ErrNoDocuments {
// // 			return []string{}, nil
// // 		}
// // 		return nil, fmt.Errorf("ошибка получения метаданных из MongoDB для сессии %d: %v", sessionID, err)
// // 	}

// // 	if metadata.Genres == nil {
// // 		return []string{}, nil
// // 	}

// // 	genres := make([]string, 0, len(metadata.Genres))
// // 	for _, genre := range metadata.Genres {
// // 		if genre != "" {
// // 			genres = append(genres, genre)
// // 		}
// // 	}

// // 	return genres, nil
// // }

// // func GetNewSessions(email string, page int) (*GetNewSessionsResponse, error) {
// // 	if email == "" {
// // 		return nil, fmt.Errorf("email не может быть пустым")
// // 	}
// // 	if page < 1 {
// // 		return nil, fmt.Errorf("номер страницы должен быть больше 0")
// // 	}

// // 	dbTx := db.GetDB().Begin()
// // 	defer func() {
// // 		if r := recover(); r != nil {
// // 			dbTx.Rollback()
// // 		}
// // 	}()

// // 	// Проверка существования пользователя
// // 	var user models.User
// // 	if err := dbTx.Where("email = ?", email).First(&user).Error; err != nil {
// // 		dbTx.Rollback()
// // 		return nil, fmt.Errorf("пользователь не найден: %v", err)
// // 	}

// // 	var recruitmentStatus sessions.Status
// // 	if err := dbTx.Where("status = ?", "Набор").First(&recruitmentStatus).Error; err != nil {
// // 		dbTx.Rollback()
// // 		return nil, fmt.Errorf("статус 'Набор' не найден: %v", err)
// // 	}

// // 	// Получение начала и конца текущего дня в UTC
// // 	now := time.Now().UTC()
// // 	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
// // 	endOfDay := startOfDay.Add(24 * time.Hour)

// // 	const limit = 6
// // 	offset := (page - 1) * limit

// // 	// Основной запрос с учетом приватности групп
// // 	baseQuery := dbTx.Table("sessions").
// // 		Joins("LEFT JOIN groups ON sessions.group_id = groups.id").
// // 		Joins("LEFT JOIN group_users ON groups.id = group_users.group_id AND group_users.user_id = ?", user.ID).
// // 		Where("sessions.created_at >= ? AND sessions.created_at < ? AND sessions.status_id = ?", startOfDay, endOfDay, recruitmentStatus.ID).
// // 		Where("(groups.is_private = ? OR (groups.is_private = ? AND group_users.user_id IS NOT NULL))", false, true)

// // 	// Подсчет общего количества сессий
// // 	var total int64
// // 	if err := baseQuery.Count(&total).Error; err != nil {
// // 		dbTx.Rollback()
// // 		return nil, fmt.Errorf("ошибка подсчета сессий: %v", err)
// // 	}

// // 	// Получение сессий с предзагрузкой связанных данных
// // 	var sessions []sessions.Session
// // 	query := dbTx.Preload("SessionType").
// // 		Preload("SessionPlace").
// // 		Preload("Group").
// // 		Preload("User").
// // 		Joins("LEFT JOIN groups ON sessions.group_id = groups.id").
// // 		Joins("LEFT JOIN group_users ON groups.id = group_users.group_id AND group_users.user_id = ?", user.ID).
// // 		Where("sessions.created_at >= ? AND sessions.created_at < ?", startOfDay, endOfDay).
// // 		Where("(groups.is_private = ? OR (groups.is_private = ? AND group_users.user_id IS NOT NULL))", false, true).
// // 		Order("sessions.created_at DESC").
// // 		Limit(limit).
// // 		Offset(offset)

// // 	if err := query.Find(&sessions).Error; err != nil {
// // 		dbTx.Rollback()
// // 		return nil, fmt.Errorf("ошибка получения сессий: %v", err)
// // 	}

// // 	if err := dbTx.Commit().Error; err != nil {
// // 		return nil, fmt.Errorf("ошибка коммита транзакции: %v", err)
// // 	}

// // 	// Получение метаданных из MongoDB для каждой сессии
// // 	sessionResponses := make([]SessionResponse, 0, len(sessions))
// // 	for _, session := range sessions {
// // 		genres, err := getSessionGenres(session.ID)
// // 		if err != nil {
// // 			// Логируем ошибку, но не останавливаем обработку
// // 			fmt.Printf("Предупреждение: не удалось получить жанры для сессии %d: %v\n", session.ID, err)
// // 			genres = []string{}
// // 		}

// // 		sessionResponse := SessionResponse{
// // 			ID:            session.ID,
// // 			Title:         session.Title,
// // 			StartTime:     session.StartTime,
// // 			SessionType:   session.SessionType.Name,
// // 			ImageURL:      session.ImageURL,
// // 			SessionPlace:  session.SessionPlace.Title,
// // 			Genres:        genres,
// // 			Duration:      session.Duration,
// // 			CurrentUsers:  session.CurrentUsers,
// // 			CountUsersMax: session.CountUsersMax,
// // 			GroupName:     session.Group.Name,
// // 		}
// // 		sessionResponses = append(sessionResponses, sessionResponse)
// // 	}

// // 	hasMore := int64(page*limit) < total

// // 	return &GetNewSessionsResponse{
// // 		Sessions: sessionResponses,
// // 		HasMore:  hasMore,
// // 		Page:     page,
// // 		Total:    total,
// // 	}, nil
// // }

// type GetSessionsInput struct {
// 	Query       *string `form:"query"`
// 	CategoryID  *uint   `form:"category_id"`
// 	SessionType *string `form:"session_type"`
// 	City        *string `form:"city"`
// 	Page        int     `form:"page" binding:"required,min=1"`
// 	SortBy      *string `form:"sort_by"`  // "date" или "users", "city"
// 	Order       *string `form:"order"`    // "asc" или "desc"
// 	NewOnly     bool    `form:"new_only"` // true — только новые
// }

// func GetSessions(email string, input GetSessionsInput) (*PaginatedSearchResponse, error) {
// 	if email == "" {
// 		return nil, fmt.Errorf("email не может быть пустым")
// 	}

// 	const pageSize = 10
// 	currentPage := input.Page
// 	if currentPage < 1 {
// 		currentPage = 1
// 	}

// 	dbTx := db.GetDB().Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			dbTx.Rollback()
// 		}
// 	}()

// 	var user models.User
// 	if err := dbTx.Where("email = ?", email).First(&user).Error; err != nil {
// 		dbTx.Rollback()
// 		return nil, fmt.Errorf("пользователь не найден: %v", err)
// 	}

// 	query := dbTx.Model(&sessions.Session{}).
// 		Preload("SessionType").
// 		Preload("SessionPlace").
// 		Preload("Group").
// 		Preload("User").
// 		Joins("JOIN groups ON sessions.group_id = groups.id").
// 		Joins("LEFT JOIN group_users ON groups.id = group_users.group_id AND group_users.user_id = ?", user.ID).
// 		Where("(groups.is_private = ? OR (groups.is_private = ? AND group_users.user_id IS NOT NULL))", false, true)

// 	if input.Query != nil && *input.Query != "" {
// 		query = query.Where("sessions.title ILIKE ?", "%"+*input.Query+"%")
// 	}

// 	if input.CategoryID != nil && *input.CategoryID > 0 {
// 		query = query.Where("sessions.session_type_id = ?", *input.CategoryID)
// 	}

// 	if input.SessionType != nil && *input.SessionType != "" {
// 		query = query.Joins("JOIN session_group_places ON sessions.session_place_id = session_group_places.id").
// 			Where("session_group_places.title = ?", *input.SessionType)
// 	}

// 	if input.NewOnly {
// 		now := time.Now().UTC()
// 		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
// 		endOfDay := startOfDay.Add(24 * time.Hour)
// 		query = query.Where("sessions.created_at BETWEEN ? AND ?", startOfDay, endOfDay)
// 	}

// 	var total int64
// 	if err := query.Count(&total).Error; err != nil {
// 		dbTx.Rollback()
// 		return nil, fmt.Errorf("ошибка подсчета: %v", err)
// 	}

// 	var allSessions []sessions.Session

// 	sortColumn := "sessions.start_time"
// 	if input.SortBy != nil {
// 		switch *input.SortBy {
// 		case "users":
// 			sortColumn = "sessions.current_users"
// 		case "date":
// 			sortColumn = "sessions.start_time"
// 		}
// 	}
// 	sortOrder := "ASC"
// 	if input.Order != nil && (*input.Order == "desc" || *input.Order == "DESC") {
// 		sortOrder = "DESC"
// 	}
// 	if input.SortBy == nil || *input.SortBy != "city" {
// 		query = query.Order(fmt.Sprintf("%s %s", sortColumn, sortOrder))
// 	}

// 	if err := query.Find(&allSessions).Error; err != nil {
// 		dbTx.Rollback()
// 		return nil, fmt.Errorf("ошибка получения: %v", err)
// 	}

// 	if err := dbTx.Commit().Error; err != nil {
// 		return nil, fmt.Errorf("ошибка коммита: %v", err)
// 	}

// 	sessionIDs := make([]uint, len(allSessions))
// 	for i, s := range allSessions {
// 		sessionIDs[i] = s.ID
// 	}

// 	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
// 	if err != nil {
// 		log.Printf("Ошибка получения метаданных: %v", err)
// 		metadataMap = make(map[uint]*sessions.SessionMetadata)
// 	}

// 	var filteredResult []SessionResponse
// 	for _, s := range allSessions {
// 		var genres []string
// 		var city string

// 		if metadata, exists := metadataMap[s.ID]; exists && metadata != nil {
// 			genres = metadata.Genres
// 			if metadata.Fields != nil {
// 				if v, ok := metadata.Fields["city"]; ok {
// 					if strCity, ok := v.(string); ok {
// 						city = strCity
// 					}
// 				}
// 			}
// 		}

// 		if input.City != nil && *input.City != "" {
// 			if city != *input.City {
// 				continue
// 			}
// 		}

// 		resp := SessionResponse{
// 			ID:            s.ID,
// 			Title:         s.Title,
// 			StartTime:     s.StartTime,
// 			Duration:      s.Duration,
// 			SessionType:   s.SessionType.Name,
// 			SessionPlace:  s.SessionPlace.Title,
// 			ImageURL:      s.ImageURL,
// 			Genres:        genres,
// 			CurrentUsers:  s.CurrentUsers,
// 			CountUsersMax: s.CountUsersMax,
// 			GroupName:     s.Group.Name,
// 			GroupID:       s.Group.ID,
// 			City:          &city,
// 		}
// 		filteredResult = append(filteredResult, resp)
// 	}

// 	if input.SortBy != nil && *input.SortBy == "city" {
// 		sort.Slice(filteredResult, func(i, j int) bool {
// 			cityI := ""
// 			cityJ := ""
// 			if filteredResult[i].City != nil {
// 				cityI = *filteredResult[i].City
// 			}
// 			if filteredResult[j].City != nil {
// 				cityJ = *filteredResult[j].City
// 			}

// 			if input.Order != nil && (*input.Order == "desc" || *input.Order == "DESC") {
// 				return cityI > cityJ
// 			}
// 			return cityI < cityJ
// 		})
// 	}

// 	totalFiltered := len(filteredResult)
// 	totalPages := int((int64(totalFiltered) + int64(pageSize) - 1) / int64(pageSize))

// 	offset := (currentPage - 1) * pageSize
// 	end := offset + pageSize
// 	if end > totalFiltered {
// 		end = totalFiltered
// 	}

// 	var paginatedResult []SessionResponse
// 	if offset < totalFiltered {
// 		paginatedResult = filteredResult[offset:end]
// 	}

// 	return &PaginatedSearchResponse{
// 		Sessions:    paginatedResult,
// 		Page:        currentPage,
// 		PageSize:    pageSize,
// 		Total:       totalFiltered,
// 		TotalPages:  totalPages,
// 		HasNext:     currentPage < totalPages,
// 		HasPrevious: currentPage > 1,
// 	}, nil
// }
