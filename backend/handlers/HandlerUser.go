package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type GetNewSessionsResponse struct {
	Sessions []services.SessionResponse `json:"sessions"`
	HasMore  bool                       `json:"has_more"`
	Page     int                        `json:"page"`
	Total    int64                      `json:"total"`
}

type PopularSessionResponseDoc struct {
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

type CachedPopularSessionsDoc struct {
	Sessions  []PopularSessionResponseDoc `json:"sessions"`
	UpdatedAt time.Time                   `json:"updated_at"`
	Count     int                         `json:"count"`
}

// GetNewSessions godoc
// @Summary Получение новых сессий
// @Description Получает новые сессии, созданные сегодня, с пагинацией по 6 штук
// @Tags Сессии
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param page query int false "Номер страницы (по умолчанию 1)" minimum(1)
// @Success 200 {object} GetNewSessionsResponse "Список новых сессий"
// @Failure 400 {object} map[string]string "Неверные параметры запроса"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/sessions/new [get]
func GetNewSessions(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}

	// Получение и валидация параметра page
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный номер страницы"})
		return
	}

	response, err := services.GetNewSessions(email, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPopularSessions godoc
// @Summary Получение популярных сессий
// @Description Возвращает 10 самых популярных сессий из кэша Redis (обновляется каждые 4 часа)
// @Tags Сессии
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Success 200 {object} CachedPopularSessionsDoc "Список популярных сессий из кэша"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/sessions/popular [get]
func GetPopularSessions(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}

	cachedSessions, err := services.GetPopularSessionsFromCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cachedSessions)
}

// func GetCategorySessions(c *gin.Context) {
// 	email := c.MustGet("email").(string)

// }

// func GetCategorySessions(c *gin.Context) {
// 	email := c.MustGet("email").(string)

// }
