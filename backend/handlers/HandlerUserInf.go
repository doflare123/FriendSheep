package handlers

import (
	"fmt"
	"friendship/models"
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetGroupsUserSub godoc
// @Summary Получить подписки пользователя на группы
// @Description Возвращает список групп, на которые подписан пользователь, исключая группы, созданные им.
// @Tags Users inf
// @Security BearerAuth
// @Produce json
// @Param id query int false "ID пользователя (опционально)"
// @Success 200 {array} services.GroupResponse "Список групп, на которые подписан пользователь"
// @Failure 400 {object} map[string]string "Не передан jwt"
// @Failure 403 {object} map[string]string "Нет прав"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users/subscriptions [get]
func GetGroupsUserSub(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)
	fmt.Println(email)

	userIDParam := c.Query("id")

	var targetUserID uint
	if userIDParam == "" {
		user := models.User{}
		targetUserID = user.ID
	} else {
		idUint, err := strconv.ParseUint(userIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат id"})
			return
		}
		targetUserID = uint(idUint)
	}

	groups, err := services.GetGroupsUserSubByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func SubscriptionsTelegramNotify(c *gin.Context) {
	var input services.TelegramNotifyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := input.SubscriptionsTelegramNotify()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "details": "Привязка телеграм аккаунта не удалась"})
	}

}

func UnSubscriptionsTelegramNotify(c *gin.Context) {
	var input services.TelegramNotifyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := input.UnSubscriptionsTelegramNotify()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "details": "Отвязка телеграм аккаунта не удалась"})
	}

}

// GetInfAboutUser godoc
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полную информацию о текущем авторизованном пользователе, включая его сессии и статистику.
// @Tags         Users inf
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  services.InformationAboutUser "Полная информация о пользователе"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      404  {object}  map[string]string "Пользователь не найден"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/users/inf [get]
func GetInfAboutUser(c *gin.Context) {
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

	userInf, err := services.GetInfAboutUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if userInf == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, userInf)
}

// GetInfAboutUser godoc
// @Summary      Получить информацию о другом пользователе
// @Description  Возвращает полную информацию о текущем авторизованном пользователе, включая его сессии и статистику.
// @Tags         Users inf
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "ID пользователя"
// @Success      200  {object}  services.InformationAboutUser "Полная информация о пользователе"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      404  {object}  map[string]string "Пользователь не найден"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/users/inf/{id} [get]
func GetInfAboutUserByID(c *gin.Context) {
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

	var id = c.Param("id")
	fmt.Println("id", id)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id не может быть пустым"})
		return
	}
	idMain, _ := strconv.ParseUint(id, 10, 32)
	fmt.Println("idMain", idMain)
	fmt.Println("idMain", uint(idMain))
	user := models.User{}
	if user.Email == email {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нононо мистер фиш, ты не будешь здесь получать информацию о себе"})
		return
	}
	userInf, err := services.GetInfAboutAnotherUser(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if userInf == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, userInf)
}

// UpdateUserProfile godoc
// @Summary      Обновить профиль пользователя
// @Description  Обновляет имя, us (юзернейм) или изображение текущего пользователя. Поля в теле запроса опциональны.
// @Tags         Users inf
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        profile body services.UpdateUserRequest true "Данные для обновления профиля"
// @Success      200 {object} object{message=string, user=object{name=string,us=string,image=string,email=string}} "Профиль успешно обновлен"
// @Failure      400 {object} map[string]string "Некорректные данные или ошибка (например, us уже занят)"
// @Failure      401 {object} map[string]string "Пользователь не авторизован"
// @Router       /api/users/user/profile [patch]
func UpdateUserProfile(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}

	email, ok := emailValue.(string)
	if !ok || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email"})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные"})
		return
	}

	updatedUser, err := services.UpdateUserProfile(email, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "профиль обновлён",
		"user": gin.H{
			"name":  updatedUser.Name,
			"us":    updatedUser.Us,
			"image": updatedUser.Image,
			"staus": updatedUser.Status,
		},
	})
}

// ChangeTilesPattern godoc
// @Summary      Изменить порядок отображения плиток статистики
// @Description  Позволяет пользователю настроить, какие плитки статистики будут отображаться в его профиле.
// @Tags         Users inf
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        settings body services.ChangeTilesPatternInput true "Новые настройки для плиток статистики"
// @Success      200  {object}  map[string]string "Порядок плиток изменен"
// @Failure      400  {object}  map[string]string "Некорректные данные"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Router       /api/users/tiles [patch]
func ChangeTilesPattern(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}

	email, ok := emailValue.(string)
	if !ok || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email"})
		return
	}

	var input services.ChangeTilesPatternInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось распарсить json", "details": err.Error()})
		return
	}

	if err := services.ChangePattern(email, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Порядок плиток изменен"})
}
