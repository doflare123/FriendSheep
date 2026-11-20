package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"friendship/services/users"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type UserEditApplication interface {
	GetOwn(context.Context, uint) (users.EditableProfile, error)
	Patch(context.Context, uint, users.ProfilePatch) (users.EditableProfile, error)
	UploadImage(context.Context, uint, io.Reader) (users.EditableProfile, error)
	DeleteImage(context.Context, uint) (users.EditableProfile, error)
}

type UserEditHandler struct{ service UserEditApplication }

func NewUserEditHandler(service UserEditApplication) *UserEditHandler {
	return &UserEditHandler{service: service}
}

// GetOwn возвращает собственные данные для формы редактирования.
// @Summary Получить данные для редактирования профиля
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} users.EditableProfile
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/me/profile [get]
func (h *UserEditHandler) GetOwn(c *gin.Context) {
	id, ok := profileUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetOwn(c.Request.Context(), id)
	writeUserEditResult(c, result, err)
}

// Patch изменяет только явно переданные поля собственного профиля.
// @Summary Изменить информацию своего профиля
// @Description Имя 1–100 символов, us 3–32 ASCII буквы/цифры/подчёркивания (нижний регистр, необязательный @), статус до 500 символов. birthday: YYYY-MM-DD от 1900-01-01 до сегодня UTC; null очищает. cityId: существующий положительный ID города; null очищает. Пропущенные поля сохраняются. Неизвестные и повторные поля, пустой объект и null для имени/us/статуса запрещены. Аватар изменяется отдельным запросом.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body users.ProfilePatch true "Изменяемые данные"
// @Success 200 {object} users.EditableProfile
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 413 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/me/profile [patch]
func (h *UserEditHandler) Patch(c *gin.Context) {
	id, ok := profileUserID(c)
	if !ok {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 8192))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			utils.JSONError(c, http.StatusRequestEntityTooLarge, "profile_body_too_large")
		} else {
			handleUserEditError(c, users.ErrInvalidProfile)
		}
		return
	}
	var patch users.ProfilePatch
	if err := json.Unmarshal(body, &patch); err != nil {
		handleUserEditError(c, users.ErrInvalidProfile)
		return
	}
	result, err := h.service.Patch(c.Request.Context(), id, patch)
	writeUserEditResult(c, result, err)
}

// UploadImage принимает единственный файл аватара и передаёт его сервису.
// @Summary Загрузить аватар своего профиля
// @Description Ровно один файл image в multipart/form-data. JPEG, PNG или WebP до 5 MiB и 4096×4096 пикселей; формат проверяется по содержимому. Остальные поля запрещены.
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param image formData file true "Аватар"
// @Success 200 {object} users.EditableProfile
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 413 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /api/v2/users/me/profile/image [put]
func (h *UserEditHandler) UploadImage(c *gin.Context) {
	id, ok := profileUserID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, (5<<20)+(64<<10))
	err := c.Request.ParseMultipartForm(5 << 20)
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			handleUserEditError(c, users.ErrImageTooLarge)
		} else {
			handleUserEditError(c, users.ErrInvalidImage)
		}
		return
	}
	form := c.Request.MultipartForm
	if form == nil || len(form.Value) != 0 || len(form.File) != 1 || len(form.File["image"]) != 1 {
		handleUserEditError(c, users.ErrInvalidImage)
		return
	}
	file, err := form.File["image"][0].Open()
	if err != nil {
		handleUserEditError(c, users.ErrInvalidImage)
		return
	}
	defer file.Close()
	result, err := h.service.UploadImage(c.Request.Context(), id, file)
	writeUserEditResult(c, result, err)
}

// DeleteImage возвращает стандартный аватар текущего пользователя.
// @Summary Удалить аватар своего профиля
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} users.EditableProfile
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v2/users/me/profile/image [delete]
func (h *UserEditHandler) DeleteImage(c *gin.Context) {
	id, ok := profileUserID(c)
	if !ok {
		return
	}
	result, err := h.service.DeleteImage(c.Request.Context(), id)
	writeUserEditResult(c, result, err)
}

func writeUserEditResult(c *gin.Context, result users.EditableProfile, err error) {
	if err != nil {
		handleUserEditError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func handleUserEditError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, users.ErrInvalidProfile):
		utils.BadRequest(c, "invalid_profile", utils.WithMessage("Некорректные данные профиля"))
	case errors.Is(err, users.ErrCityNotFound):
		utils.BadRequest(c, "city_not_found", utils.WithMessage("Город не найден"))
	case errors.Is(err, users.ErrUsTaken):
		utils.JSONError(c, http.StatusConflict, "us_taken", utils.WithMessage("Этот адрес пользователя уже занят"))
	case errors.Is(err, users.ErrInvalidImage):
		utils.BadRequest(c, "invalid_profile_image", utils.WithMessage("Требуется корректный JPEG, PNG или WebP до 4096×4096 пикселей"))
	case errors.Is(err, users.ErrImageTooLarge):
		utils.JSONError(c, http.StatusRequestEntityTooLarge, "profile_image_too_large", utils.WithMessage("Аватар превышает лимит 5 MiB или 4096×4096 пикселей"))
	case errors.Is(err, users.ErrImageStorageUnavailable):
		utils.JSONError(c, http.StatusServiceUnavailable, "profile_image_storage_unavailable", utils.WithMessage("Хранилище аватаров временно недоступно"))
	default:
		handleUserProfileError(c, err)
	}
}
