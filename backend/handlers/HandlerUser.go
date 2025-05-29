package handlers

import (
	"friendship/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionEmailInput struct {
	Email string `json:"email" validate:"required,email"`
}

func CreateSessionRegisterHandler(c *gin.Context) {
	var input SessionEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
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

func CreateUserHandler(c *gin.Context) {
	var input services.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	user, err := services.CreateUser(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"validation_error": err.Error()})
		return
	}
	if user != nil {
		log.Printf("Что-то не так")
	}

	c.JSON(http.StatusCreated, gin.H{"messege": "Пользователь зарегистрирован"})
}

func VerifySessionHandler(c *gin.Context) {
	var input services.VerifySessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	verified, err := services.VerifySession(input)
	if err != nil {
		if err.Error() == "session not found or incomplete" {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"verified": false, "error": "code or type incorrect"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": true})
}
