package handlers

import (
	"fmt"
	"friendship/services"
	"net/http"
	"strconv"
	"strings"
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
// @Tags Получение данных о сессиях
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
// @Tags Получение данных о сессиях
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

// GetCategorySessions godoc
// @Summary Получение сессий по категории
// @Description Возвращает список из 10 сессий по указанной категории из открытых групп. Вводить вам нужно будет иметь мапу с категориями, где ключ - название категории, а значение - id категории. Сейчас такие значнеия: 1 - Фильмы, 2 - Игры. 3 - Настолки, 4 - Другое
// @Tags Получение данных о сессиях
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param category_id query int true "ID категории сессии"
// @Param page query int false "Номер страницы (по умолчанию 1)" default(1)
// @Success 200 {object} services.CategorySessionsResponse "Список сессий"
// @Failure 400 {object} map[string]string "Некорректные параметры запроса"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/sessions/category [get]
func GetCategorySessions(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}

	var input services.CategorySessionsInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные параметры запроса: " + err.Error()})
		return
	}

	if input.CategoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID категории"})
		return
	}

	if input.Page <= 0 {
		input.Page = 1
	}

	response, err := services.GetCategorySessions(email, input)
	if err != nil {
		if strings.Contains(err.Error(), "категория не найдена") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "пользователь не найден") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)

}

// SearchSessions godoc
// @Summary Поиск сессий
// @Description Поиск сессий по заданным критериям с учетом приватности групп
// @Tags Получение данных о сессиях
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query string true "Поисковый запрос для поиска по названию сессии"
// @Param categoryID query uint true "ID категории сессии (session_type_id)"
// @Param sessionType query string false "Тип места проведения сессии (опционально)"
// @Param Page query string true "Страница, так как отдает от 9 штук"
// @Success 200 {array} services.SessionResponse "Список найденных сессий"
// @Failure 400 {object} map[string]string "Некорректные параметры запроса"
// @Failure 401 {object} map[string]string "Не передан JWT токен"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/sessions/search [get]
func SearchSessions(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}
	var input services.SearchSessionsRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные параметры запроса: " + err.Error()})
		return
	}
	result, err := services.SearchSessions(&email, &input.Query, &input.CategoryID, input.SessionType, input.Page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetDetailedInfo получает детальную информацию о сессии
// @Summary Получить детальную информацию о сессии
// @Description Возвращает подробную информацию о сессии, включая метаданные
// @Tags Получение данных о сессиях
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path int true "Уникальный идентификатор сессии" minimum(1)
// @Success 200 {object} services.SessionDetailResponse "Успешное получение информации о сессии"
// @Failure 400 {object} map[string]string "Некорректные параметры запроса"
// @Failure 401 {object} map[string]string "Ошибки аутентификации"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Сессия не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/sessions/{sessionId} [get]
func GetDetailedInfo(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}

	sessionIDParam := c.Param("sessionId")
	if sessionIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId не указан"})
		return
	}

	sessionIDInt, err := strconv.Atoi(sessionIDParam)
	if err != nil || sessionIDInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный sessionId"})
		return
	}
	sessionID := uint(sessionIDInt)

	session, err := services.GetInfoAboutSession(&email, &sessionID)
	if err != nil {
		fmt.Printf("Ошибка при получении сессии: %v\n", err)

		if err.Error() == "сессия не найдена" {
			c.JSON(http.StatusNotFound, gin.H{"error": "сессия не найдена"})
			return
		}
		if err.Error() == "доступ запрещен: пользователь не является членом приватной группы" {
			c.JSON(http.StatusForbidden, gin.H{"error": "доступ запрещен"})
			return
		}
		if err.Error() == "пользователь не найден" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// @Summary Получить сессии пользователя из групп
// @Description Возвращает список сессий со статусом "Набор" из групп, в которых состоит пользователь, с пагинацией (9 сессий на страницу)
// @Tags Получение данных о сессиях
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Success 200 {array} services.SessionResponse "Сессии успешно получены"
// @Failure 400 {object} object{error=string} "Неверный запрос"
// @Failure 401 {object} object{error=string} "Не авторизован"
// @Failure 404 {object} object{error=string} "Пользователь не найден"
// @Failure 500 {object} object{error=string} "Внутренняя ошибка сервера"
// @Router /api/users/sessions/user-groups [get]
func GetSessionsUserGroups(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный номер страницы"})
		return
	}
	sessions, err := services.GetSessionsUserGroups(&email, page)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "пользователь не найден"):
			c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
			return

		case strings.Contains(err.Error(), "email не может быть пустым"):
			c.JSON(http.StatusBadRequest, gin.H{"error": "email не может быть пустым"})
			return

		case strings.Contains(err.Error(), "статус 'Набор' не найден"):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка конфигурации системы"})
			return

		case strings.Contains(err.Error(), "ошибка при получении групп"):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при получении групп пользователя"})
			return

		case strings.Contains(err.Error(), "ошибка при получении сессий"):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при получении сессий"})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
			return
		}
	}
	c.JSON(http.StatusOK, sessions)
}
