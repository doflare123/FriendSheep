package handlers

import (
	storage "friendship/db"
	"friendship/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	users := storage.GetAllUsers()
	c.JSON(http.StatusAccepted, users)
}

func GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	user := storage.GetUserByID(id)
	if user == nil {
		// Возвращение ошибки 404 (Not Found), если заметка не найдена
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, user)
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

	c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		// Возвращение ошибки 400 (Bad Request), если ID некорректен
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}
	// Структура для хранения входных данных
	var input struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required"`
	}
	// Привязка входных данных в формате JSON к структуре input
	if err := c.ShouldBindJSON(&input); err != nil {
		// Возвращение ошибки 400 (Bad Request), если входные данные некорректны
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Обновление заметки в хранилище
	User := storage.UpdateUser(id, input.Name, input.Email)
	if User == nil {
		// Возвращение ошибки 404 (Not Found), если заметка не найдена
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	// Возвращение обновленной заметки в формате JSON с кодом 200 (OK)
	c.JSON(http.StatusOK, gin.H{
		"message": "Новый пользователь зарегистрирован",
	})
}
