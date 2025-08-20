package handlers

import (
	"friendship/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteAccount godoc
// @Summary      Удаление аккаунта пользователя
// @Description  Позволяет текущему авторизованному пользователю удалить свой аккаунт. Это действие необратимо.
// @Tags         Users inf
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string "Аккаунт успешно удалён"
// @Failure      400  {object}  map[string]string "Ошибка (например, пользователь не найден)"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Router       /api/users/delete [delete]
func DeleteAccount(c *gin.Context) {
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
	err := services.DeleteAccount(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Аккаунт успешно удалён"})
}
