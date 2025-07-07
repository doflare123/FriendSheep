package handlers

import (
	"fmt"
	"friendship/middlewares"
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateSession godoc
// @Summary Создание новой сессии
// @Description Создает сессию в группе с возможностью загрузки изображения и добавления метаданных
// @Tags Сессии
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Название сессии"
// @Param session_type formData uint true "ID типа сессии"
// @Param group_id formData uint true "ID группы"
// @Param start_time formData string true "Время начала (в формате RFC3339, напр. 2025-07-10T19:00:00+02:00)"
// @Param duration formData uint false "Длительность в минутах"
// @Param count_users formData uint true "Максимальное количество участников"
// @Param meta_type formData string false "Тип метаданных (например: киновечер)"
// @Param genres formData string false "Жанры (через запятую, напр: драма,комедия)"
// @Param fields formData string false "Доп. поля (напр: ключ:значение,ключ2:знач2)"
// @Param location formData string false "Место проведения"
// @Param year formData int false "Год (например: 2023)"
// @Param country formData string false "Страна"
// @Param age_limit formData string false "Возрастное ограничение (напр: 16+)"
// @Param notes formData string false "Примечания"
// @Param image formData file false "Изображение"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось разобрать форму: " + err.Error()})
		return
	}

	header, err := c.FormFile("image")
	if err == nil { // Если файл есть
		file, err := header.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть изображение"})
			return
		}
		defer file.Close()

		filePath := fmt.Sprintf("uploads/%s", header.Filename)
		url, err := middlewares.UploadImage(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка загрузки изображения: " + err.Error()})
			return
		}
		input.Image = url
	}

	ok, err := services.CreateSession(email, input)
	if err != nil || !ok {
		if input.Image != "" {
			middlewares.DeleteImage(fmt.Sprintf("uploads/%s", header.Filename))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Сессия создана"})
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
