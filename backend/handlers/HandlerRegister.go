package handlers

import (
	"errors"
	"friendship/services/register"
	"friendship/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionEmailInput struct {
	Email   string `json:"email" binding:"required,email"`
	TypeSes string `json:"type_ses" binding:"required"`
}

type RegHandler interface {
	CreateSessionRegister(c *gin.Context)
	VerifySession(c *gin.Context)
	CreateUser(c *gin.Context)
	ChangePassword(c *gin.Context)
}

type regHandler struct {
	srv register.RegService
}

func NewRegisterHandler(srv register.RegService) RegHandler {
	return &regHandler{
		srv: srv,
	}
}

// CreateSessionRegister создает сессию регистрации по email
// @Summary Создать сессию регистрации
// @Description Создает сессию для подтверждения email пользователя при регистрации
// @Tags sessions
// @Accept json
// @Produce json
// @Param input body SessionEmailInput true "Email пользователя"
// @Success 200 {object} models.SessionRegResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v2/register/session/register [post]
func (h *regHandler) CreateSessionRegister(c *gin.Context) {
	var input SessionEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	session, err := h.srv.CreateSessionRegister(c.Request.Context(), input.Email, input.TypeSes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// VerifySession проверяет код сессии
// @Summary Проверить сессию
// @Description Проверяет код сессии, отправленный на email
// @Tags sessions
// @Accept json
// @Produce json
// @Param input body register.VerifySessionInput true "Данные сессии для проверки"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Router /api/v2/register/session/verify [patch]
func (h *regHandler) VerifySession(c *gin.Context) {
	var input register.VerifySessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный формат данных"})
		return
	}

	verified, err := h.srv.VerifySession(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, register.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена или истекла"})
		case errors.Is(err, register.ErrTooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Превышено количество попыток"})
		case errors.Is(err, register.ErrSessionTypeMismatch):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный тип сессии"})
		case errors.Is(err, register.ErrInvalidCode):
			c.JSON(http.StatusUnauthorized, gin.H{"verified": false, "error": "Неверный код"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		}
		return
	}

	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"verified": false, "error": "Код или тип некорректные"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": true})
}

// CreateUser регистрирует нового пользователя
// @Summary Создать пользователя
// @Description Регистрирует нового пользователя по данным из запроса
// @Tags users
// @Accept json
// @Produce json
// @Param input body register.CreateUserInput true "Данные для создания пользователя"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v2/register/ [post]
func (h *regHandler) CreateUser(c *gin.Context) {
	var input register.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	authResponse, err := h.srv.CreateUser(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, register.ErrSessionNotVerified):
			c.JSON(http.StatusForbidden, gin.H{"error": "Сессия не подтверждена"})
		case errors.Is(err, register.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		case errors.Is(err, register.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email уже существует"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Пользователь успешно зарегистрирован",
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
		"admin_groups":  authResponse.AdminGroups,
	})
}

// ChangePassword изменяет пароль пользователя
// @Summary Изменить пароль
// @Description Изменяет пароль пользователя после подтверждения сессии восстановления
// @Tags users
// @Accept json
// @Produce json
// @Param input body register.ChangePasswordInput true "Данные для смены пароля"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v2/register/password/change [post]
func (h *regHandler) ChangePassword(c *gin.Context) {
	var input register.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	err := h.srv.ChangePassword(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, register.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось изменить пароль"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменен"})
}
