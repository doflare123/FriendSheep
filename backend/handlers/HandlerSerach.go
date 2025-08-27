package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SearchUsers godoc
// @Summary Поиск пользователей
// @Description Поиск пользователей по имени с пагинацией.
// @Tags search
// @Accept  json
// @Produce  json
// @Param name query string false "Имя для поиска"
// @Param page query int false "Номер страницы" default(1)
// @Success 200 {object} services.GetUsersResponse "Успешный поиск"
// @Failure 400 {object} map[string]string "Некорректный номер страницы"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/search [get]
func SearchUsers(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}

	email, ok := emailValue.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email в контексте"})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email не может быть пустым"})
		return
	}

	name := c.Query("name")
	pageStr := c.DefaultQuery("page", "1")

	user, err := services.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка поиска исходного пользователя"})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный номер страницы"})
		return
	}

	users, err := services.SearchUsers(name, page, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка поиска пользователей"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// SearchGroups godoc
// @Summary Поиск групп
// @Description Поиск групп по названию, категориям и с сортировкой (с пагинацией).
// @Tags search
// @Accept  json
// @Produce  json
// @Param name query string false "Название группы для поиска"
// @Param category query string false "Фильтр по категории"
// @Param sort_by query string false "Поле сортировки: members (по числу участников), date (по дате регистрации), category (по имени категории)"
// @Param order query string false "Порядок сортировки: asc или desc" default(desc)
// @Param page query int false "Номер страницы (>=1)" default(1)
// @Success 200 {object} services.GetGroupsResponse "Успешный поиск"
// @Failure 400 {object} map[string]string "Некорректный номер страницы"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/groups/search [get]
func SearchGroups(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}

	name := c.Query("name")
	pageStr := c.DefaultQuery("page", "1")
	sortBy := c.DefaultQuery("sort_by", "")
	order := c.DefaultQuery("order", "desc")
	category := c.DefaultQuery("category", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный номер страницы"})
		return
	}

	groups, err := services.SearchGroups(name, page, sortBy, order, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка поиска групп"})
		return
	}

	c.JSON(http.StatusOK, groups)
}
