package handlers

import (
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateNews godoc
// @Summary      Создание новости
// @Description  Создает новую новость вместе с текстом. Доступно только администраторам.
// @Tags         News
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        news body services.CreateNewsInput true "Данные для создания новости"
// @Success      201  {object} news.News "Новость успешно создана"
// @Failure      400  {object} map[string]string "Некорректные данные или ошибка валидации"
// @Failure      401  {object} map[string]string "Пользователь не авторизован или не является администратором"
// @Router       /api/news [post]
func CreateNews(c *gin.Context) {
	var input services.CreateNewsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные: " + err.Error()})
		return
	}

	res, err := services.CreateNews(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetNews godoc
// @Summary      Получить список новостей
// @Description  Возвращает постраничный список новостей, отсортированных по дате создания.
// @Tags         News
// @Produce      json
// @Param        page query int false "Номер страницы" default(1)
// @Success      200  {object}  services.NewsPage "Постраничный список новостей"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/news [get]
func GetNews(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	res, err := services.GetNews(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     res.Page,
		"total":    res.Total,
		"has_more": res.HasMore,
		"items":    res.Items,
	})
}

// GetNewsByID godoc
// @Summary      Получение новости
// @Description  Возвращает новость по ID с текстом и комментариями (с ником и картинкой юзера)
// @Tags         News
// @Accept       json
// @Produce      json
// @Param        id path int true "ID новости"
// @Success      200  {object} services.NewsDTO "Новость с текстом и комментариями"
// @Failure      404  {object} map[string]string "Новость не найдена"
// @Router       /api/news/{id} [get]
func GetNewsByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	item, err := services.GetNewsByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "новость не найдена"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// AddComment godoc
// @Summary      Добавить комментарий
// @Description  Создаёт комментарий для новости (только авторизованные пользователи)
// @Tags         News
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID новости"
// @Param        comment body services.CreateCommentInput true "Комментарий"
// @Success      201 {object} news.Comments
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /api/news/{id}/comments [post]
func AddComment(c *gin.Context) {
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

	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	var input services.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные: " + err.Error()})
		return
	}

	comment, err := services.CreateComment(uint(newsID), email, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// DeleteComment godoc
// @Summary      Удалить комментарий
// @Description  Удаляет комментарий (только для администраторов)
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        newsId path int true "ID новости"
// @Param        commentId path int true "ID комментария"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Router       /api/news/{newsId}/comments/{commentId} [delete]
func DeleteComment(c *gin.Context) {
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

	user, err := services.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "недостаточно прав"})
		return
	}

	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id комментария"})
		return
	}

	if err := services.DeleteComment(commentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "комментарий удалён"})
}
