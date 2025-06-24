package handlers

import (
	"friendship/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateGroupInputDoc struct {
	Name             string `json:"name" example:"Моя группа"`
	Description      string `json:"description" example:"Полное описание"`
	SmallDescription string `json:"smallDescription" example:"Краткое описание"`
	Image            string `json:"image" example:"https://example.com/image.png"`
	IsPrivate        string `json:"isPrivate" example:"true"`
	City             string `json:"city,omitempty" example:"Москва"`
	Categories       []uint `json:"categories" example:"1,2,3"` // <- просто строка, swag так принимает
}

type JoinGroupInputDoc struct {
	GroupID uint `json:"groupId" binding:"required"`
}

// @Security BearerAuth
// CreateGroup godoc
// @Summary Создание группы
// @Description Создает новую группу
// @Tags groups
// @Accept  json
// @Produce  json
// @Param input body CreateGroupInputDoc true "Данные для создания группы"
// @Success 200 {object} map[string]string "Группа успешно создана"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/groups/createGroup [post]
func CreateGroup(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	var input services.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный json", "details": err.Error()})
		return
	}

	group, err := services.CreateGroup(email, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if group == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Группа не создана"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Группа успешно создана"})
}

// @Security BearerAuth
// joinToGroup godoc
// @Summary Присоединение к группе
// @Description Новый пользователь присоединяется к группе
// @Tags groups
// @Accept  json
// @Produce  json
// @Param input body JoinGroupInputDoc true "Данные для присоединения к группе"
// @Success 200 {object} map[string]string "Пользователь успешно присоеденен к группе"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/groups/joinToGroup [post]
func JoinGroup(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}
	email := emailValue.(string)

	var input services.JoinGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный json", "details": err.Error()})
		return
	}
	res, err := services.JoinGroup(email, input)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неудалось приисоединиться к группе"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно присоеденен к группе"})
}
