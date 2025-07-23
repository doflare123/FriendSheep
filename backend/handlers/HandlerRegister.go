package handlers

import (
	"friendship/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionEmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

// CreateSessionRegisterHandler создает сессию регистрации по email
// @Summary Создать сессию регистрации
// @Description Создает сессию для подтверждения email пользователя при регистрации
// @Tags sessions
// @Accept json
// @Produce json
// @Param input body SessionEmailInput true "Email пользователя"
// @Success 200 {object} models.SessionRegResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sessions/register [post]
func CreateSessionRegisterHandler(c *gin.Context) {
	var input SessionEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный формат почты"})
		return
	}

	session, err := services.CreateSessionRegister(input.Email)
	if err != nil {
		log.Printf("Ошибка при создании сессии: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create session"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// VerifySessionHandler проверяет код сессии
// @Summary Проверить сессию
// @Description Проверяет код сессии, отправленный на email
// @Tags sessions
// @Accept json
// @Produce json
// @Param input body services.VerifySessionInput true "Данные сессии для проверки"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Router /api/sessions/verify [patch]
func VerifySessionHandler(c *gin.Context) {
	var input services.VerifySessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный формат кода"})
		return
	}

	verified, err := services.VerifySession(input)
	if err != nil {
		if err.Error() == "сессия не найдена или удалена" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "неизвестная ошибка"})
		}
		return
	}

	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"messege": false, "error": "код или типа некорректные"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": true})
}

// CreateUserHandler регистрирует нового пользователя
// @Summary Создать пользователя
// @Description Регистрирует нового пользователя по данным из запроса
// @Tags users
// @Accept json
// @Produce json
// @Param input body services.CreateUserInput true "Данные для создания пользователя"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/users [post]
func CreateUserHandler(c *gin.Context) {
	var input services.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный json"})
		return
	}

	user, err := services.CreateUser(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user != nil {
		log.Printf("Что-то не так")
	}

	c.JSON(http.StatusCreated, gin.H{"messege": "Пользователь зарегистрирован"})
}
