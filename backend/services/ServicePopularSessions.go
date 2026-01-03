package services

import (
	"context"
	"encoding/json"
	"fmt"
	"friendship/db"
	"friendship/models/events"
	"log"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	POPULAR_SESSIONS_CACHE_KEY = "popular_sessions:top10"
	CACHE_EXPIRATION           = 6 * time.Hour
)

var (
	popularSessionsCron *cron.Cron
	ctx                 = context.Background()
)

type PopularSessionResponse struct {
	ID             uint      `json:"id"`
	Title          string    `json:"title"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Duration       uint16    `json:"duration"`
	SessionType    string    `json:"session_type"`
	SessionPlace   string    `json:"session_place"`
	ImageURL       string    `json:"image_url"`
	Genres         []string  `json:"genres,omitempty"`
	CurrentUsers   uint16    `json:"current_users"`
	CountUsersMax  uint16    `json:"count_users_max"`
	PopularityRate float64   `json:"popularity_rate"`
	GroupName      string    `json:"group_name"`
}

type CachedPopularSessions struct {
	Sessions  []PopularSessionResponse `json:"sessions"`
	UpdatedAt time.Time                `json:"updated_at"`
	Count     int                      `json:"count"`
}

// InitPopularSessionsCache инициализирует cron задачу для обновления кэша
func InitPopularSessionsCache() error {
	popularSessionsCron = cron.New()

	// Запускаем каждые 4 часа
	_, err := popularSessionsCron.AddFunc("0 */4 * * *", func() {
		log.Println("Обновление кэша популярных сессий...")
		if err := updatePopularSessionsCache(); err != nil {
			log.Printf("Ошибка обновления кэша популярных сессий: %v", err)
		} else {
			log.Println("Кэш популярных сессий успешно обновлен")
		}
	})

	if err != nil {
		return fmt.Errorf("ошибка создания cron задачи: %v", err)
	}

	popularSessionsCron.Start()

	go func() {
		log.Println("Первоначальное заполнение кэша популярных сессий...")
		if err := updatePopularSessionsCache(); err != nil {
			log.Printf("Ошибка первоначального заполнения кэша: %v", err)
		} else {
			log.Println("Кэш популярных сессий изначально заполнен")
		}
	}()

	return nil
}

// StopPopularSessionsCache останавливает cron задачу
func StopPopularSessionsCache() {
	if popularSessionsCron != nil {
		popularSessionsCron.Stop()
		log.Println("Cron задача для популярных сессий остановлена")
	}
}

// GetPopularSessionsFromCache получает популярные сессии из Redis кэша
func GetPopularSessionsFromCache() (*CachedPopularSessions, error) {
	redisClient := db.GetRedis()
	if redisClient == nil {
		return nil, fmt.Errorf("Redis клиент недоступен")
	}

	cachedData, err := redisClient.Get(ctx, POPULAR_SESSIONS_CACHE_KEY).Result()
	if err == redis.Nil {
		// Кэш пуст, принудительно обновляем
		log.Println("Кэш популярных сессий пуст, обновляем...")
		if updateErr := updatePopularSessionsCache(); updateErr != nil {
			return nil, fmt.Errorf("кэш пуст и не удалось обновить: %v", updateErr)
		}

		cachedData, err = redisClient.Get(ctx, POPULAR_SESSIONS_CACHE_KEY).Result()
		if err != nil {
			return nil, fmt.Errorf("ошибка получения данных из кэша после обновления: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения данных из Redis: %v", err)
	}

	var result CachedPopularSessions
	if err := json.Unmarshal([]byte(cachedData), &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации данных из кэша: %v", err)
	}

	// Проверяем актуальность кэша
	if time.Since(result.UpdatedAt) > CACHE_EXPIRATION {
		go func() {
			log.Println("Кэш устарел, асинхронное обновление...")
			if err := updatePopularSessionsCache(); err != nil {
				log.Printf("Ошибка асинхронного обновления кэша: %v", err)
			}
		}()
	}

	return &result, nil
}

// updatePopularSessionsCache обновляет кэш популярных сессий
func updatePopularSessionsCache() error {
	redisClient := db.GetRedis()
	if redisClient == nil {
		return fmt.Errorf("Redis клиент недоступен")
	}

	sessions, err := fetchPopularSessionsFromDB()
	if err != nil {
		return fmt.Errorf("ошибка получения данных из БД: %v", err)
	}

	cachedData := CachedPopularSessions{
		Sessions:  sessions,
		UpdatedAt: time.Now(),
		Count:     len(sessions),
	}

	jsonData, err := json.Marshal(cachedData)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %v", err)
	}

	err = redisClient.Set(ctx, POPULAR_SESSIONS_CACHE_KEY, jsonData, CACHE_EXPIRATION+time.Hour).Err()
	if err != nil {
		return fmt.Errorf("ошибка сохранения в Redis: %v", err)
	}

	log.Printf("Кэш обновлен: %d популярных сессий", len(sessions))
	return nil
}

// fetchPopularSessionsFromDB получает популярные сессии из базы данных
func fetchPopularSessionsFromDB() ([]PopularSessionResponse, error) {
	dbConn := db.GetDB()

	var sessionData []struct {
		ID                uint      `gorm:"column:id"`
		Title             string    `gorm:"column:title"`
		StartTime         time.Time `gorm:"column:start_time"`
		EndTime           time.Time `gorm:"column:end_time"`
		Duration          uint16    `gorm:"column:duration"`
		CurrentUsers      uint16    `gorm:"column:current_users"`
		CountUsersMax     uint16    `gorm:"column:count_users_max"`
		ImageURL          string    `gorm:"column:image_url"`
		SessionTypeName   string    `gorm:"column:session_type_name"`
		SessionPlaceTitle string    `gorm:"column:session_place_title"`
		GroupName         string    `gorm:"column:group_name"`
	}

	var recruitmentStatus events.Status
	if err := dbConn.Where("status = ?", "Набор").First(&recruitmentStatus).Error; err != nil {
		dbConn.Rollback()
		return nil, fmt.Errorf("статус 'Набор' не найден: %v", err)
	}

	query := `
		SELECT 
			s.id, s.title, s.start_time, s.end_time, s.duration,
			s.current_users, s.count_users_max, s.image_url,
			c.name as session_type_name,
			sgp.title as session_place_title,
			g.name as group_name
		FROM sessions s
		JOIN categories c ON s.session_type_id = c.id
		JOIN session_group_places sgp ON s.session_place_id = sgp.id
		JOIN groups g ON s.group_id = g.id
		WHERE g.is_private = false 
		  AND s.start_time > NOW()
		  AND s.count_users_max > 0
		  AND s.current_users > 1
		  AND s.status_id = ?
		ORDER BY 
			(CAST(s.current_users AS FLOAT) / CAST(s.count_users_max AS FLOAT)) DESC,
			s.current_users DESC
		LIMIT 10
	`

	if err := dbConn.Raw(query, recruitmentStatus.ID).Scan(&sessionData).Error; err != nil {
		return nil, fmt.Errorf("ошибка выполнения SQL запроса: %v", err)
	}

	if len(sessionData) == 0 {
		return []PopularSessionResponse{}, nil
	}

	sessionIDs := make([]uint, len(sessionData))
	for i, session := range sessionData {
		sessionIDs[i] = session.ID
	}

	metadataMap, err := getSessionsMetadata(sessionIDs)
	if err != nil {
		log.Printf("Предупреждение: ошибка получения метаданных из MongoDB: %v", err)
		metadataMap = make(map[uint]*events.SessionMetadata)
	}

	var result []PopularSessionResponse
	for _, session := range sessionData {
		popularityRate := 0.0
		if session.CountUsersMax > 0 {
			popularityRate = float64(session.CurrentUsers) / float64(session.CountUsersMax)
		}

		response := PopularSessionResponse{
			ID:             session.ID,
			Title:          session.Title,
			StartTime:      session.StartTime,
			EndTime:        session.EndTime,
			Duration:       session.Duration,
			SessionType:    session.SessionTypeName,
			SessionPlace:   session.SessionPlaceTitle,
			ImageURL:       session.ImageURL,
			CurrentUsers:   session.CurrentUsers,
			CountUsersMax:  session.CountUsersMax,
			PopularityRate: popularityRate,
			GroupName:      session.GroupName,
			Genres:         []string{},
		}

		if metadata, exists := metadataMap[session.ID]; exists && metadata != nil {
			response.Genres = metadata.Genres
		}

		result = append(result, response)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].PopularityRate != result[j].PopularityRate {
			return result[i].PopularityRate > result[j].PopularityRate
		}
		return result[i].CurrentUsers > result[j].CurrentUsers
	})

	return result, nil
}

func getSessionsMetadata(sessionIDs []uint) (map[uint]*events.SessionMetadata, error) {
	mongoClient := db.GetMongoDB()
	if mongoClient == nil {
		return make(map[uint]*events.SessionMetadata), nil
	}

	collection := db.Database().Collection("session_metadata")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionIDsInterface := make([]interface{}, len(sessionIDs))
	for i, id := range sessionIDs {
		sessionIDsInterface[i] = id
	}

	filter := bson.M{
		"session_id": bson.M{"$in": sessionIDsInterface},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return make(map[uint]*events.SessionMetadata), nil
		}
		return nil, err
	}
	defer cursor.Close(ctx)

	metadataMap := make(map[uint]*events.SessionMetadata)
	for cursor.Next(ctx) {
		var metadata events.SessionMetadata
		if err := cursor.Decode(&metadata); err != nil {
			continue
		}
		metadataMap[metadata.SessionID] = &metadata
	}

	return metadataMap, cursor.Err()
}
