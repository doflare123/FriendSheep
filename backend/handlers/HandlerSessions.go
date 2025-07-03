package handlers

import (
	"fmt"
	"friendship/middlewares"
	"friendship/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

	// Получаем файл
	header, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось получить изображение"})
		return
	}

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

	// Создание сессии
	ok, err := services.CreateSession(email, input)
	if err != nil || !ok {
		middlewares.DeleteImage(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Сессия создана"})
}
