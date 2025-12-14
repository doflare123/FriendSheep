package services

// type CategorySessionsInput struct {
// 	CategoryID uint `form:"category_id" binding:"required" validate:"min=1"`
// 	Page       int  `form:"page" validate:"min=1"`
// }

// type CategorySessionsResponse struct {
// 	Sessions    []SessionResponse `json:"sessions"`
// 	TotalCount  int64             `json:"total_count"`
// 	CurrentPage int               `json:"current_page"`
// 	HasMore     bool              `json:"has_more"`
// }

// // func GetCategorySessions(input CategorySessionsInput) (*CategorySessionsResponse, error) {
// // 	database := db.GetDB()

// // 	var category models.Category
// // 	if err := database.Where("id = ?", input.CategoryID).First(&category).Error; err != nil {
// // 		return nil, fmt.Errorf("категория не найдена: %v", err)
// // 	}

// // 	var recruitmentStatus sessions.Status
// // 	if err := database.Where("status = ?", "Набор").First(&recruitmentStatus).Error; err != nil {
// // 		database.Rollback()
// // 		return nil, fmt.Errorf("статус 'Набор' не найден: %v", err)
// // 	}

// // 	limit := 10
// // 	offset := (input.Page - 1) * limit

// // 	var sessionsList []sessions.Session
// // 	query := database.
// // 		Preload("SessionType").
// // 		Preload("SessionPlace").
// // 		Preload("Group").
// // 		Preload("User").
// // 		Joins("JOIN groups ON sessions.group_id = groups.id").
// // 		Where("sessions.session_type_id = ? AND groups.is_private = ? AND sessions.start_time > ? AND sessions.status_id = ?",
// // 			input.CategoryID, false, time.Now(), recruitmentStatus.ID).
// // 		Order("sessions.current_users DESC, sessions.created_at DESC").
// // 		Limit(limit).
// // 		Offset(offset)

// // 	if err := query.Find(&sessionsList).Error; err != nil {
// // 		return nil, fmt.Errorf("ошибка получения сессий: %v", err)
// // 	}

// // 	// Получаем общее количество записей для пагинации
// // 	var totalCount int64
// // 	countQuery := database.
// // 		Model(&sessions.Session{}).
// // 		Joins("JOIN groups ON sessions.group_id = groups.id").
// // 		Where("sessions.session_type_id = ? AND groups.is_private = ? AND sessions.start_time > ?",
// // 			input.CategoryID, false, time.Now())

// // 	if err := countQuery.Count(&totalCount).Error; err != nil {
// // 		return nil, fmt.Errorf("ошибка подсчета сессий: %v", err)
// // 	}

// // 	sessionIDs := make([]uint, len(sessionsList))
// // 	for i, session := range sessionsList {
// // 		sessionIDs[i] = session.ID
// // 	}

// // 	metadataMap, err := db.GetSessionsMetadata(sessionIDs)
// // 	if err != nil {
// // 		log.Printf("Ошибка получения метаданных: %v", err)
// // 		metadataMap = make(map[uint]*sessions.SessionMetadata) // создаем пустую карту при ошибке
// // 	}

// // 	sessionResponses := make([]SessionResponse, 0, len(sessionsList))

// // 	for _, session := range sessionsList {
// // 		var genres []string
// // 		if metadata, exists := metadataMap[session.ID]; exists && metadata != nil {
// // 			genres = metadata.Genres
// // 		}

// // 		sessionResponse := SessionResponse{
// // 			ID:            session.ID,
// // 			Title:         session.Title,
// // 			StartTime:     session.StartTime,
// // 			Duration:      session.Duration,
// // 			SessionType:   session.SessionType.Name,
// // 			SessionPlace:  session.SessionPlace.Title,
// // 			ImageURL:      session.ImageURL,
// // 			Genres:        genres,
// // 			CurrentUsers:  session.CurrentUsers,
// // 			CountUsersMax: session.CountUsersMax,
// // 			GroupName:     session.Group.Name,
// // 		}

// // 		sessionResponses = append(sessionResponses, sessionResponse)
// // 	}

// // 	hasMore := int64(offset+limit) < totalCount

// // 	response := &CategorySessionsResponse{
// // 		Sessions:    sessionResponses,
// // 		TotalCount:  totalCount,
// // 		CurrentPage: input.Page,
// // 		HasMore:     hasMore,
// // 	}

// // 	return response, nil
// // }
