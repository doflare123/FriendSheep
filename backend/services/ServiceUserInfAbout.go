package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/middlewares"
	"friendship/models"
	"friendship/models/sessions"
	statsusers "friendship/models/stats_users"
	"time"

	"gorm.io/gorm"
)

type InformationAboutUser struct {
	Name         string    `json:"name"`
	Us           string    `json:"us"`
	Image        string    `json:"image"`
	DataRegister time.Time `json:"data_register"`
	Enterprise   bool      `json:"enterprise"`
	TelegramLink bool      `json:"telegram_link"`
	Status       string    `json:"status"`

	UpcomingSessions []SessionInfo `json:"upcoming_sessions"`
	RecentSessions   []SessionInfo `json:"recent_sessions"`
	PopularGenres    []GenreStats  `json:"popular_genres"`
	UserStats        UserStatsInfo `json:"user_stats"`
	Tiles            []string      `json:"tiles"`
}

type SessionInfo struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	CurrentUsers     uint16    `json:"current_users"`
	MaxUsers         uint16    `json:"max_users"`
	ImageURL         string    `json:"image_url"`
	Status           string    `json:"status"`
	Type_session     string    `json:"type_session"`
	Category_session string    `json:"category_session"`
	Location         string    `json:"location,omitempty"`
	Genres           []string  `json:"genres,omitempty"`
	City             *string   `json:"city"`
}

type GenreStats struct {
	Name  string `json:"name"`
	Count uint16 `json:"count"`
}

type UserStatsInfo struct {
	CountCreateSession uint16 `json:"count_create_session"`
	SeriesSessionCount uint16 `json:"series_session_count"`
	MostBigSession     uint16 `json:"most_big_session"`
	MostPopDay         string `json:"most_pop_day,omitempty"`
	CountFilms         uint16 `json:"count_films"`
	CountGames         uint16 `json:"count_games"`
	CountTableGames    uint16 `json:"count_table_games"`
	CountAnother       uint16 `json:"count_another"`
	CountAll           uint16 `json:"count_all"`
	SpentTime          uint64 `json:"spent_time,omitempty"`
}

type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty" binding:"omitempty,min=5,max=40"`
	Us     *string `json:"us,omitempty" binding:"omitempty,min=5,max=40"`
	Image  *string `json:"image,omitempty"`
	Status *string `json:"status,omitempty" binding:"omitempty,min=1,max=50"`
}

func GetInfAboutUser(email string) (*InformationAboutUser, error) {
	database := db.GetDB()
	if database == nil {
		return nil, errors.New("база данных недоступна")
	}

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var tiles statsusers.SettingTile
	if err := database.Where("user_id = ?", user.ID).First(&tiles).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var enabledTiles []string
	if tiles.Count_films {
		enabledTiles = append(enabledTiles, "count_films")
	}
	if tiles.Count_games {
		enabledTiles = append(enabledTiles, "count_games")
	}
	if tiles.Count_table {
		enabledTiles = append(enabledTiles, "count_table")
	}
	if tiles.Count_other {
		enabledTiles = append(enabledTiles, "count_other")
	}
	if tiles.Count_all {
		enabledTiles = append(enabledTiles, "count_all")
	}
	if tiles.Spent_time {
		enabledTiles = append(enabledTiles, "spent_time")
	}

	upcomingSessions, err := getUpcomingSessions(database, user.ID)
	if err != nil {
		return nil, err
	}

	recentSessions, err := getRecentSessions(database, user.ID)
	if err != nil {
		return nil, err
	}

	popularGenres, err := getPopularGenres(database, user.ID)
	if err != nil {
		return nil, err
	}

	userStats, err := getUserStats(database, user.ID)
	if err != nil {
		return nil, err
	}

	result := &InformationAboutUser{
		Name:             user.Name,
		Us:               user.Us,
		Image:            user.Image,
		DataRegister:     user.DataRegister,
		Enterprise:       user.Enterprise,
		TelegramLink:     user.TelegramID != nil,
		Status:           user.Status,
		UpcomingSessions: upcomingSessions,
		RecentSessions:   recentSessions,
		PopularGenres:    popularGenres,
		UserStats:        userStats,
		Tiles:            enabledTiles,
	}

	return result, nil
}

func GetInfAboutAnotherUser(email string) (*InformationAboutUser, error) {
	database := db.GetDB()
	if database == nil {
		return nil, errors.New("база данных недоступна")
	}

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var tiles statsusers.SettingTile
	if err := database.Where("user_id = ?", user.ID).First(&tiles).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var enabledTiles []string
	if tiles.Count_films {
		enabledTiles = append(enabledTiles, "count_films")
	}
	if tiles.Count_games {
		enabledTiles = append(enabledTiles, "count_games")
	}
	if tiles.Count_table {
		enabledTiles = append(enabledTiles, "count_table")
	}
	if tiles.Count_other {
		enabledTiles = append(enabledTiles, "count_other")
	}
	if tiles.Count_all {
		enabledTiles = append(enabledTiles, "count_all")
	}
	if tiles.Spent_time {
		enabledTiles = append(enabledTiles, "spent_time")
	}

	upcomingSessions, err := getUpcomingSessions(database, user.ID)
	if err != nil {
		return nil, err
	}

	// recentSessions, err := getRecentSessions(database, user.ID)
	// if err != nil {
	// 	return nil, err
	// }

	popularGenres, err := getPopularGenres(database, user.ID)
	if err != nil {
		return nil, err
	}

	userStats, err := getUserStats(database, user.ID)
	if err != nil {
		return nil, err
	}

	result := &InformationAboutUser{
		Name:         user.Name,
		Us:           user.Us,
		Image:        user.Image,
		DataRegister: user.DataRegister,
		Enterprise:   user.Enterprise,
		// TelegramLink:     user.TelegramID != nil,
		Status:           user.Status,
		UpcomingSessions: upcomingSessions,
		// RecentSessions:   recentSessions,
		PopularGenres: popularGenres,
		UserStats:     userStats,
		Tiles:         enabledTiles,
	}

	return result, nil
}

func getUpcomingSessions(dbCon *gorm.DB, userID uint) ([]SessionInfo, error) {
	var sessionUsers []sessions.SessionUser

	err := dbCon.Preload("Session").
		Preload("Session.Status").
		Preload("Session.SessionPlace").
		Preload("Session.SessionType").
		Joins("JOIN sessions ON session_users.session_id = sessions.id").
		Joins("JOIN statuses ON sessions.status_id = statuses.id").
		Where("session_users.user_id = ? AND statuses.status = ?", userID, "Набор").
		Find(&sessionUsers).Error

	if err != nil {
		return nil, err
	}

	sessionIDs := make([]uint, len(sessionUsers))
	for i, su := range sessionUsers {
		sessionIDs[i] = su.Session.ID
	}

	metadata, err := db.GetSessionsMetadata(sessionIDs)
	if err != nil {
		metadata = make(map[uint]*sessions.SessionMetadata)
	}

	result := make([]SessionInfo, len(sessionUsers))
	for i, su := range sessionUsers {
		sessionInfo := SessionInfo{
			ID:               su.Session.ID,
			Title:            su.Session.Title,
			StartTime:        su.Session.StartTime,
			EndTime:          su.Session.EndTime,
			CurrentUsers:     su.Session.CurrentUsers,
			MaxUsers:         su.Session.CountUsersMax,
			Type_session:     su.Session.SessionPlace.Title,
			Category_session: su.Session.SessionType.Name,
			ImageURL:         safeStringValue(&su.Session.ImageURL),
			Status:           su.Session.Status.Status,
		}

		if meta, exists := metadata[su.Session.ID]; exists && meta != nil {
			sessionInfo.Location = safeStringValue(&meta.Location)
			sessionInfo.Genres = safeStringSlice(meta.Genres)
			var city *string
			if meta.Fields != nil {
				if v, ok := meta.Fields["city"]; ok {
					if strCity, ok := v.(string); ok {
						city = &strCity
					}
				}
			}
			sessionInfo.City = city
		}

		result[i] = sessionInfo
	}

	return result, nil
}

func getRecentSessions(dbCon *gorm.DB, userID uint) ([]SessionInfo, error) {
	var sessionUsers []sessions.SessionUser

	err := dbCon.Preload("Session").
		Preload("Session.Status").
		Preload("Session.SessionPlace").
		Joins("JOIN sessions ON session_users.session_id = sessions.id").
		Joins("JOIN statuses ON sessions.status_id = statuses.id").
		Where("session_users.user_id = ? AND statuses.status = ?", userID, "Завершена").
		Order("sessions.end_time DESC").
		Limit(10).
		Find(&sessionUsers).Error

	if err != nil {
		return nil, err
	}

	sessionIDs := make([]uint, len(sessionUsers))
	for i, su := range sessionUsers {
		sessionIDs[i] = su.Session.ID
	}

	metadata, err := db.GetSessionsMetadata(sessionIDs)
	if err != nil {
		metadata = make(map[uint]*sessions.SessionMetadata)
	}

	result := make([]SessionInfo, len(sessionUsers))
	for i, su := range sessionUsers {
		sessionInfo := SessionInfo{
			ID:           su.Session.ID,
			Title:        su.Session.Title,
			StartTime:    su.Session.StartTime,
			EndTime:      su.Session.EndTime,
			CurrentUsers: su.Session.CurrentUsers,
			MaxUsers:     su.Session.CountUsersMax,
			Type_session: su.Session.SessionPlace.Title,
			ImageURL:     safeStringValue(&su.Session.ImageURL),
			Status:       su.Session.Status.Status,
		}

		if meta, exists := metadata[su.Session.ID]; exists && meta != nil {
			sessionInfo.Location = safeStringValue(&meta.Location)
			sessionInfo.Genres = safeStringSlice(meta.Genres)
			var city *string
			if meta.Fields != nil {
				if v, ok := meta.Fields["city"]; ok {
					if strCity, ok := v.(string); ok {
						city = &strCity
					}
				}
			}
			sessionInfo.City = city
		}
		result[i] = sessionInfo
	}
	return result, nil
}

func getPopularGenres(dbCon *gorm.DB, userID uint) ([]GenreStats, error) {
	var genreStats []struct {
		Name  string
		Count uint16
	}

	err := dbCon.Table("sessions_stats_genres_users").
		Select("genres.name, sessions_stats_genres_users.count").
		Joins("JOIN genres ON sessions_stats_genres_users.genre_id = genres.id").
		Where("sessions_stats_genres_users.user_id = ?", userID).
		Order("sessions_stats_genres_users.count DESC").
		Limit(10).
		Scan(&genreStats).Error

	if err != nil {
		return nil, err
	}

	result := make([]GenreStats, len(genreStats))
	for i, gs := range genreStats {
		result[i] = GenreStats{
			Name:  gs.Name,
			Count: gs.Count,
		}
	}

	return result, nil
}

func getUserStats(db *gorm.DB, userID uint) (UserStatsInfo, error) {
	var sideStats statsusers.SideStats_users
	var sessionStats statsusers.SessionStats_users

	err := db.Where("user_id = ?", userID).Preload("DaysWeek").First(&sideStats).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return UserStatsInfo{}, err
	}

	err = db.Where("user_id = ?", userID).First(&sessionStats).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return UserStatsInfo{}, err
	}

	return UserStatsInfo{
		CountCreateSession: safeUint16Value(sideStats.CountCreateSession),
		SeriesSessionCount: safeUint16Value(sideStats.SeriesSesionCount),
		MostBigSession:     safeUint16Value(sideStats.MostBigSession),
		MostPopDay:         safeStringValue(&sideStats.DaysWeek.Name),
		CountFilms:         safeUint16Value(&sessionStats.CountFilms),
		CountGames:         safeUint16Value(&sessionStats.CountGames),
		CountTableGames:    safeUint16Value(&sessionStats.CountTableGames),
		CountAnother:       safeUint16Value(&sessionStats.CountAnother),
		CountAll:           safeUint16Value(&sessionStats.CountAll),
		SpentTime:          sessionStats.SpentTime,
	}, nil
}

func safeStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeUint16Value(u *uint16) uint16 {
	if u == nil {
		return 0
	}
	return *u
}

func safeStringSlice(slice []string) []string {
	if slice == nil {
		return []string{}
	}
	return slice
}

func UpdateUserProfile(email string, req UpdateUserRequest) (*models.User, error) {
	database := db.GetDB()

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Сохраняем старый URL изображения
	oldImageURL := user.Image

	updates := make(map[string]interface{})

	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}
	if req.Us != nil && *req.Us != "" {
		var existing models.User
		if err := database.Where("us = ? AND email <> ?", *req.Us, email).First(&existing).Error; err == nil {
			return nil, errors.New("US уже используется")
		}
		updates["us"] = *req.Us
	}
	if req.Image != nil && *req.Image != "" {
		updates["image"] = *req.Image
	}
	if req.Status != nil && *req.Status != "" {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return &user, nil
	}

	if err := database.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// После успешного обновления удаляем старое изображение из S3
	if req.Image != nil && oldImageURL != "" && oldImageURL != *req.Image {
		go func(url string) {
			if err := middlewares.DeleteImageFromS3(url); err != nil {
				fmt.Printf("Предупреждение: не удалось удалить старое изображение пользователя из S3: %v\n", err)
			}
		}(oldImageURL)
	}

	return &user, nil
}
