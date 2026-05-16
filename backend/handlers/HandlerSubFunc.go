package handlers

import (
	service "friendship/services/sub"
	"friendship/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubHandler interface {
	ChangePhoto(c *gin.Context)
}

type subHandler struct {
	imgService service.ImgService
}

func NewSubHandler(imgService service.ImgService) SubHandler {
	return &subHandler{imgService: imgService}
}

// ChangePhoto godoc
// @Summary      Загрузка фотографии
// @Description  Загружает фотографию в хранилище и возвращает ссылку. Эту ссылку затем можно использовать для создания или обновления данных сущности (например, группы).
// @Tags         sub_func
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Изображение для загрузки (JPG, PNG, максимум 10MB)"
// @Success      200  {object}  map[string]string "URL загруженного изображения"
// @Failure      400  {object}  dto.ErrorResponse "Файл не предоставлен, имеет неверный тип или размер"
// @Failure      401  {object}  dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object}  dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /api/v2/sub/UploadImg [post]
func (h *subHandler) ChangePhoto(c *gin.Context) {
	header, err := c.FormFile("image")
	if err != nil {
		utils.BadRequest(c, "Image upload failed", utils.WithDetails(err.Error()))
		return
	}

	url, err := h.imgService.UploadImg(c.Request.Context(), header, "photos")
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": url})
}
