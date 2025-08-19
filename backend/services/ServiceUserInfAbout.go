package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
	"friendship/models/sessions"
	statsusers "friendship/models/stats_users"
	"time"

	"gorm.io/gorm"
)

type InformationAboutUser struct {
	Name             string        `json:"name"`
	Image            string        `json:"image"`
	DataRegister     time.Time     `json:"data_register"`
	Enterprise       bool          `json:"enterprise"`
	TelegramLink     bool          `json:"telegram_link"`
	UpcomingSessions []SessionInfo `json:"upcoming_sessions"`
	RecentSessions   []SessionInfo `json:"recent_sessions"`
	PopularGenres    []GenreStats  `json:"popular_genres"`
	UserStats        UserStatsInfo `json:"user_stats"`
}

type SessionInfo struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	CurrentUsers uint16    `json:"current_users"`
	MaxUsers     uint16    `json:"max_users"`
	ImageURL     string    `json:"image_url"`
	Status       string    `json:"status"`
	Location     string    `json:"location,omitempty"`
	Genres       []string  `json:"genres,omitempty"`
}

type GenreStats struct {
	Name  string `json:"name"`
	Count uint16 `json:"count"`
}

type UserStatsInfo struct {
	CountCreateSession uint16 `json:"count_create_session"`
	SeriesSessionCount uint16 `json:"series_session_count"`
	MostBigSession     uint16 `json:"most_big_session"`
	CountFilms         uint16 `json:"count_films"`
	CountGames         uint16 `json:"count_games"`
	CountTableGames    uint16 `json:"count_table_games"`
	CountAnother       uint16 `json:"count_another"`
	CountAll           uint16 `json:"count_all"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Us    *string `json:"us,omitempty"`
	Image *string `json:"image,omitempty"`
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
		Image:            user.Image,
		DataRegister:     user.DataRegister,
		Enterprise:       user.Enterprise,
		TelegramLink:     user.TelegramID != nil,
		UpcomingSessions: upcomingSessions,
		RecentSessions:   recentSessions,
		PopularGenres:    popularGenres,
		UserStats:        userStats,
	}

	return result, nil
}

func getUpcomingSessions(dbCon *gorm.DB, userID uint) ([]SessionInfo, error) {
	var sessionUsers []sessions.SessionUser

	err := dbCon.Preload("Session").
		Preload("Session.Status").
		Preload("Session.SessionPlace").
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
			ID:           su.Session.ID,
			Title:        su.Session.Title,
			StartTime:    su.Session.StartTime,
			EndTime:      su.Session.EndTime,
			CurrentUsers: su.Session.CurrentUsers,
			MaxUsers:     su.Session.CountUsersMax,
			ImageURL:     safeStringValue(&su.Session.ImageURL),
			Status:       su.Session.Status.Status,
		}

		if meta, exists := metadata[su.Session.ID]; exists && meta != nil {
			sessionInfo.Location = safeStringValue(&meta.Location)
			sessionInfo.Genres = safeStringSlice(meta.Genres)
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
			ImageURL:     safeStringValue(&su.Session.ImageURL),
			Status:       su.Session.Status.Status,
		}

		if meta, exists := metadata[su.Session.ID]; exists && meta != nil {
			sessionInfo.Location = safeStringValue(&meta.Location)
			sessionInfo.Genres = safeStringSlice(meta.Genres)
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

	err := db.Where("user_id = ?", userID).First(&sideStats).Error
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
		CountFilms:         safeUint16Value(&sessionStats.CountFilms),
		CountGames:         safeUint16Value(&sessionStats.CountGames),
		CountTableGames:    safeUint16Value(&sessionStats.CountTableGames),
		CountAnother:       safeUint16Value(&sessionStats.CountAnother),
		CountAll:           safeUint16Value(&sessionStats.CountAll),
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

	if len(updates) == 0 {
		return &user, nil
	}

	if err := database.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
