package handlers

import (
	"friendship/middlewares"
	"friendship/services"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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

type JoinGroupResponseDoc struct {
	Message string `json:"message" example:"Заявка на вступление отправлена"`
	Joined  bool   `json:"joined" example:"false"`
}

// @Security BearerAuth
// CreateGroup godoc
// @Summary Создание группы
// @Description Создает новую группу
// @Tags groups
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Название сессии"
// @Param description formData string true "Описание группы"
// @Param smallDescription formData string true "Короткое описание"
// @Param city formData string false "Город локации группы (опционально)"
// @Param categories formData []int true "Категории группы (записываются в виде массива целых чисел)"
// @Param isPrivate formData bool true "Приватная ли группа"
// @Param image formData file true "Изображение"
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
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы", "details": err.Error()})
		return
	}

	var imageURL string
	header, err := c.FormFile("image")
	if err == nil {
		if err := middlewares.ValidateImageMIME(header); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неподдерживаемый тип файла", "details": err.Error()})
			return
		}

		if header.Size > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Файл слишком большой, максимум 10MB"})
			return
		}

		file, err := header.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть изображение"})
			return
		}
		defer file.Close()

		imageURL, err = middlewares.UploadImage(file, header.Filename, "groups")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка загрузки изображения: " + err.Error()})
			return
		}
		input.Image = imageURL
	}

	group, err := services.CreateGroup(email, input)
	if err != nil {
		// Удаляем загруженное изображение при ошибке создания группы
		if imageURL != "" {
			// Извлекаем имя файла из URL
			parts := strings.Split(imageURL, "/")
			if len(parts) > 0 {
				filename := parts[len(parts)-1]
				middlewares.DeleteImage(filepath.Join("uploads", "groups", filename))
			}
		}
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
// @Success 200 {object} JoinGroupResponseDoc "Успешный ответ: вступил или заявка отправлена"
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

	c.JSON(http.StatusOK, gin.H{
		"message": res.Message,
		"joined":  res.Joined,
	})
}

// DeleteGroups godoc
// @Summary Удаление группы
// @Description Удаляет группу. Только администратор группы может удалить её. Удаляются также участники и заявки на вступление.
// @Tags groups
// @Security BearerAuth
// @Param groupId path int true "ID группы"
// @Produce json
// @Success 200 {object} map[string]string "Группа успешно удалена"
// @Failure 400 {object} map[string]string "Некорректный ID группы"
// @Failure 403 {object} map[string]string "Нет прав или группа не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/groups/{groupId} [delete]
func DeleteGroups(c *gin.Context) {
	email := c.MustGet("email").(string)
	groupIDParam := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID группы"})
		return
	}

	if err := services.DeleteGroup(email, uint(groupID)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вы удалили группу"})
}
