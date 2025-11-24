package handlers

import (
	"friendship/middlewares"
	"friendship/services"
	"friendship/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SessionJoinInputDoc struct {
	SessionID uint `json:"session_id" binding:"required"`
	GroupID   uint `json:"group_id" binding:"required"`
}

// CreateSession godoc
// @Summary Создание новой сессии
// @Description Создает сессию в группе с возможностью загрузки изображения и добавления метаданных
// @Tags Сессии
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Название сессии"
// @Param session_type formData string true "Тип сессии"
// @Param session_place formData uint true "Типа проведения сессии"
// @Param group_id formData uint true "ID группы"
// @Param start_time formData string true "Время начала (в формате RFC3339, напр. 2025-07-10T19:00:00+02:00)"
// @Param duration formData uint false "Длительность в минутах"
// @Param count_users formData uint true "Максимальное количество участников"
// @Param genres formData string false "Жанры (через запятую, напр: драма,комедия)"
// @Param fields formData string false "Доп. поля (напр: ключ:значение,ключ2:знач2)"
// @Param location formData string false "Место проведения"
// @Param year formData int false "Год (например: 2023)"
// @Param country formData string false "Страна"
// @Param age_limit formData string false "Возрастное ограничение (напр: 16+)"
// @Param notes formData string false "Примечания"
// @Param image formData string true "Изображение"
// @Success 200 {object} map[string]string "Сессия создана"
// @Failure 400 {object} map[string]string "Ошибка запроса"
// @Failure 401 {object} map[string]string "Не передан JWT"
// @Failure 500 {object} map[string]string "Ошибка сервера или валидации"
// @Security BearerAuth
// @Router /api/sessions/createSession [post]
func CreateSession(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}

	var input services.SessionInput
	if err := c.ShouldBind(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	ok, err := services.CreateSession(email, input)
	if err != nil || !ok {
		if input.Image != "" {
			middlewares.DeleteImageFromS3(input.Image)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Сессия создана"})
}

// JoinToSession godoc
// @Summary Присоединение к сессии
// @Description Позволяет пользователю присоединиться к выбранной сессии, если она не заполнена
// @Tags Сессии
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param input body SessionJoinInputDoc true "Данные для присоединения к группе"
// @Success 200 {object} map[string]string "Вы успешно присоединились к сессии"
// @Failure 400 {object} map[string]string "Ошибка разбора формы"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Сессия или пользователь не найдены"
// @Failure 409 {object} map[string]string "Сессия заполнена или пользователь уже присоединился"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/sessions/join [post]
func JoinToSession(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}
	var input services.SessionJoinInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось разобрать форму: " + err.Error()})
		return
	}
	res := services.JoinToSession(&email, input)
	if res != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вы успешно присоединились к сессии"})
}

// DeleteSession.
//
// @Summary Удаление сессии
// @Description Удаляет сессию, если пользователь является admin или operator в группе
// @Tags Сессии
// @Security BearerAuth
// @Param id path int true "ID сессии для удаления"
// @Success 200 {object} map[string]string "Сессия удалена"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "JWT не передан"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 404 {object} map[string]string "Сессия не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера или БД"
// @Router /api/sessions/sessions/{id} [delete]
func DeleteSession(c *gin.Context) {
	email := c.MustGet("email").(string)
	idStr := c.Param("id")

	sessionID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID"})
		return
	}

	if err := services.DeleteSession(email, uint(sessionID)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Сессия удалена"})
}

// LeaveSessionHandler godoc
// @Summary Отписаться от сессии
// @Description Позволяет пользователю покинуть сессию, в которой он участвует
// @Tags Сессии
// @Security BearerAuth
// @Produce json
// @Param sessionId path int true "ID сессии"
// @Success 200 {object} map[string]string "Вы успешно покинули сессию"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 403 {object} map[string]string "Вы не состоите в сессии или нет доступа"
// @Failure 404 {object} map[string]string "Сессия не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /api/sessions/{sessionId}/leave [delete]
func LeaveSessionHandler(c *gin.Context) {
	email := c.MustGet("email").(string)

	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID сессии"})
		return
	}

	if err := services.LeaveSession(email, uint(sessionID)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вы успешно покинули сессию"})
}

//admin функционал

// UpdateSessionHandler godoc
// @Summary Обновить информацию о сессии
// @Description Позволяет администратору группы изменить данные сессии, принадлежащей этой группе
// @Tags Сессии_админ
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param sessionId path int true "ID сессии"
// @Param input body services.SessionUpdateInput true "Новые данные сессии"
// @Success 200 {object} map[string]string "Сессия успешно обновлена"
// @Failure 400 {object} map[string]string "Ошибка валидации или некорректный ID"
// @Failure 403 {object} map[string]string "Нет прав на редактирование"
// @Failure 404 {object} map[string]string "Сессия не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /api/admin/sessions/{sessionId} [patch]
func UpdateSessionHandler(c *gin.Context) {
	email := c.MustGet("email").(string)

	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "некорректный ID сессии"})
		return
	}

	var input services.SessionUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ошибка в теле запроса: " + err.Error()})
		return
	}

	if err := services.UpdateSession(email, uint(sessionID), input); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Сессия успешно обновлена"})
}

// GetGenres godoc
// @Summary      Получение списка жанров
// @Description  Возвращает список всех доступных жанров для сессий.
// @Tags         Сессии
// @Produce      json
// @Success      200  {array}   string "Список жанров"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /api/sessions/genres [get]
func GetGenres(c *gin.Context) {
	genres, err := services.GetGenres()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, genres)
}
