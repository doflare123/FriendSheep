package handlers

import (
	"net/http"
	"notify_service/services"

	"github.com/gin-gonic/gin"
)

func RegisterDevice(c *gin.Context) {
	email := c.MustGet("email").(string)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не передан jwt"})
		return
	}
	var input services.RegisterDeviceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный json"})
		return
	}
	if err := services.RegisterDevice(email, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "устройство успешно зарегистрировано"})
}
