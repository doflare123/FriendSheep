package handlers

import (
	"friendship/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChangePhoto godoc
// @Summary      Загрузка фотографии
// @Description  Загружает фотографию в S3 хранилище и возвращает URL. Этот URL затем можно использовать для создания или обновления данных сущности (например, группы).
// @Tags         groups_admin
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Изображение для загрузки (JPG, PNG, максимум 10MB)"
// @Success      200  {object}  map[string]string "URL загруженного изображения"
// @Failure      400  {object}  map[string]string "Файл не предоставлен, имеет неверный тип или размер"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /api/admin/groups/UploadPhoto [post]
func ChangePhoto(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
		return
	}

	header, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось получить файл: " + err.Error()})
		return
	}

	// Валидация MIME типа
	if err := middlewares.ValidateImageMIME(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неподдерживаемый тип файла", "details": err.Error()})
		return
	}

	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл слишком большой, максимум 10MB"})
		return
	}

	// Открываем файл заново после валидации
	file, err := header.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть изображение"})
		return
	}
	defer file.Close()

	// Используем новую функцию для загрузки в S3
	imageURL, err := middlewares.UploadImageToS3(file, header.Filename, "groups")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка загрузки изображения: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": imageURL})
}
