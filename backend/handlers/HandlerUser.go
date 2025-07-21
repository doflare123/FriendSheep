package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetNewSessionsResponse struct {
	Sessions []services.SessionResponse `json:"sessions"`
	HasMore  bool                       `json:"has_more"`
	Page     int                        `json:"page"`
	Total    int64                      `json:"total"`
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
// @Router /api/users/new [get]
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

// func GetPopularSessions(c *gin.Context) {
// 	email := c.MustGet("email").(string)

// }

// func GetCategorySessions(c *gin.Context) {
// 	email := c.MustGet("email").(string)

// }

// func GetCategorySessions(c *gin.Context) {
// 	email := c.MustGet("email").(string)

// }
