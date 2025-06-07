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
