package handlers

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"friendship/services/notifications"

	"github.com/gin-gonic/gin"
)

const pushDeviceBodyLimit = 16_384

type RegisterPushDeviceRequest struct {
	SchemaVersion        int     `json:"schemaVersion" binding:"required" enums:"1" example:"1"`
	OperationID          string  `json:"operationId" binding:"required" example:"c0a00000-0000-4000-8000-000000000001"`
	ExpectedVersion      *string `json:"expectedVersion" binding:"required" extensions:"x-nullable" example:"d0a00000-0000-4000-8000-000000000001"`
	Platform             string  `json:"platform" binding:"required" enums:"android,ios" example:"android"`
	ProviderToken        string  `json:"providerToken" binding:"required" maxLength:"4096" example:"example-fcm-android-1"`
	NotificationsEnabled bool    `json:"notificationsEnabled" binding:"required" example:"true"`
}

type PushDevicesHandler struct {
	registry notifications.PushDeviceRegistry
}

func NewPushDevicesHandler(registry notifications.PushDeviceRegistry) *PushDevicesHandler {
	return &PushDevicesHandler{registry: registry}
}

// Register регистрирует или обновляет push-адрес текущей установки.
// @Summary      Зарегистрировать push-устройство
// @Description  Идемпотентно регистрирует или обновляет FCM-адрес установки для текущего пользователя и его действующей сессии.
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        installationId       path      string                     true  "ID установки"
// @Param        X-Installation-Secret header    string                     true  "Секрет установки"
// @Param        request               body      RegisterPushDeviceRequest  true  "Команда регистрации"
// @Success      200                   {object}  notifications.PushDeviceMutationResult
// @Failure      400                   {object}  dto.ErrorResponse
// @Failure      401                   {object}  dto.ErrorResponse
// @Failure      404                   {object}  dto.ErrorResponse
// @Failure      409                   {object}  dto.ErrorResponse
// @Failure      413                   {object}  dto.ErrorResponse
// @Failure      415                   {object}  dto.ErrorResponse
// @Failure      429                   {object}  dto.ErrorResponse
// @Failure      503                   {object}  dto.ErrorResponse
// @Router       /api/v2/users/me/push-devices/{installationId} [put]
func (h *PushDevicesHandler) Register(c *gin.Context) {
	preparePushResponse(c)
	userID, sessionID, ok := authenticatedPushIdentity(c)
	if !ok {
		writePushError(c, http.StatusUnauthorized, "missing_authenticated_user", nil)
		return
	}
	if h == nil || h.registry == nil {
		writePushError(c, http.StatusServiceUnavailable, "push_registry_unavailable", nil)
		return
	}
	if c.Request.URL.RawQuery != "" {
		writePushError(c, http.StatusBadRequest, notifications.PushErrorInvalidRequest, nil)
		return
	}
	secret, ok := installationSecret(c)
	if !ok {
		return
	}
	if !hasJSONContentType(c.Request) {
		writePushError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", nil)
		return
	}

	var request RegisterPushDeviceRequest
	err := DecodeStrictJSONObject(c.Writer, c.Request, pushDeviceBodyLimit, &request,
		"schemaVersion", "operationId", "expectedVersion", "platform", "providerToken", "notificationsEnabled")
	if errors.Is(err, ErrStrictJSONRequestTooLarge) {
		writePushError(c, http.StatusRequestEntityTooLarge, "request_too_large", nil)
		return
	}
	if err != nil {
		writePushError(c, http.StatusBadRequest, notifications.PushErrorInvalidRequest, nil)
		return
	}

	input := notifications.RegisterPushDeviceInput{
		InstallationID: c.Param("installationId"), InstallationSecret: secret,
		SchemaVersion: request.SchemaVersion, UserID: uint64(userID), SessionID: sessionID,
		OperationID: request.OperationID, ExpectedVersion: request.ExpectedVersion, Platform: request.Platform,
		ProviderToken: request.ProviderToken, NotificationsEnabled: request.NotificationsEnabled,
	}
	if err := notifications.ValidateRegisterPushDeviceInput(input); err != nil {
		handlePushRegistryError(c, err)
		return
	}
	result, err := h.registry.Register(c.Request.Context(), input)
	if err != nil {
		handlePushRegistryError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Detach отвязывает push-адрес текущей установки.
// @Summary      Отключить push-устройство
// @Description  Идемпотентно отвязывает установку от текущего пользователя и его действующей сессии.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        installationId          path      string  true  "ID установки"
// @Param        X-Installation-Secret   header    string  true  "Секрет установки"
// @Param        X-Push-Operation-ID     header    string  true  "UUID v4 операции"
// @Param        X-Push-Expected-Version header    string  true  "Ожидаемая версия UUID v4 или none"
// @Success      200                     {object}  notifications.PushDeviceMutationResult
// @Failure      400                     {object}  dto.ErrorResponse
// @Failure      401                     {object}  dto.ErrorResponse
// @Failure      404                     {object}  dto.ErrorResponse
// @Failure      409                     {object}  dto.ErrorResponse
// @Failure      429                     {object}  dto.ErrorResponse
// @Failure      503                     {object}  dto.ErrorResponse
// @Router       /api/v2/users/me/push-devices/{installationId} [delete]
func (h *PushDevicesHandler) Detach(c *gin.Context) {
	preparePushResponse(c)
	userID, sessionID, ok := authenticatedPushIdentity(c)
	if !ok {
		writePushError(c, http.StatusUnauthorized, "missing_authenticated_user", nil)
		return
	}
	if h == nil || h.registry == nil {
		writePushError(c, http.StatusServiceUnavailable, "push_registry_unavailable", nil)
		return
	}
	if c.Request.URL.RawQuery != "" {
		writePushError(c, http.StatusBadRequest, notifications.PushErrorInvalidRequest, nil)
		return
	}
	secret, ok := installationSecret(c)
	if !ok {
		return
	}
	operationID, ok := singleRequiredPushHeader(c, "X-Push-Operation-ID")
	if !ok {
		return
	}
	expectedHeader, ok := singleRequiredPushHeader(c, "X-Push-Expected-Version")
	if !ok {
		return
	}
	if !hasEmptyBody(c.Request) {
		writePushError(c, http.StatusBadRequest, notifications.PushErrorInvalidRequest, nil)
		return
	}
	var expectedVersion *string
	if expectedHeader != "none" {
		expectedVersion = &expectedHeader
	}
	input := notifications.DetachPushDeviceInput{
		InstallationID: c.Param("installationId"), InstallationSecret: secret,
		SchemaVersion: notifications.PushDeviceSchemaVersion, UserID: uint64(userID), SessionID: sessionID,
		OperationID: operationID, ExpectedVersion: expectedVersion,
	}
	if err := notifications.ValidateDetachPushDeviceInput(input); err != nil {
		handlePushRegistryError(c, err)
		return
	}
	result, err := h.registry.Detach(c.Request.Context(), input)
	if err != nil {
		handlePushRegistryError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func preparePushResponse(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
}

func authenticatedPushIdentity(c *gin.Context) (uint, string, bool) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return 0, "", false
	}
	sessionID, exists := c.Get("authSessionID")
	typed, ok := sessionID.(string)
	if !exists || !ok || typed == "" {
		return 0, "", false
	}
	return userID, typed, true
}

func installationSecret(c *gin.Context) (string, bool) {
	values := c.Request.Header.Values("X-Installation-Secret")
	if len(values) == 0 || len(values) == 1 && values[0] == "" {
		writePushError(c, http.StatusNotFound, notifications.PushErrorDeviceNotFound, nil)
		return "", false
	}
	if len(values) != 1 {
		writePushError(c, http.StatusBadRequest, notifications.PushErrorInvalidRequest, nil)
		return "", false
	}
	return values[0], true
}

func singleRequiredPushHeader(c *gin.Context, name string) (string, bool) {
	values := c.Request.Header.Values(name)
	if len(values) != 1 || values[0] == "" {
		writePushError(c, http.StatusBadRequest, notifications.PushErrorInvalidRequest, nil)
		return "", false
	}
	return values[0], true
}

func hasJSONContentType(request *http.Request) bool {
	values := request.Header.Values("Content-Type")
	if len(values) != 1 {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(values[0])
	return err == nil && mediaType == "application/json"
}

func hasEmptyBody(request *http.Request) bool {
	if request.Body == nil {
		return true
	}
	defer request.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(request.Body, 1))
	return err == nil && len(raw) == 0
}

func handlePushRegistryError(c *gin.Context, err error) {
	var registryError *notifications.PushRegistryError
	if !errors.As(err, &registryError) {
		writePushError(c, http.StatusServiceUnavailable, "push_registry_unavailable", nil)
		return
	}
	switch registryError.Code {
	case notifications.PushErrorInvalidRequest, notifications.PushErrorUnsupportedSchemaVersion, notifications.PushErrorOptInRequired:
		writePushError(c, http.StatusBadRequest, registryError.Code, nil)
	case notifications.PushErrorDeviceNotFound:
		writePushError(c, http.StatusNotFound, registryError.Code, nil)
	case notifications.PushErrorVersionConflict:
		writePushError(c, http.StatusConflict, registryError.Code, map[string]*string{"currentVersion": registryError.CurrentVersion})
	case notifications.PushErrorOperationConflict, notifications.PushErrorInstallationConflict,
		notifications.PushErrorTokenConflict, notifications.PushErrorTokenInvalidated:
		writePushError(c, http.StatusConflict, registryError.Code, nil)
	default:
		writePushError(c, http.StatusServiceUnavailable, "push_registry_unavailable", nil)
	}
}

func writePushError(c *gin.Context, status int, code string, details any) {
	if details == nil {
		c.JSON(status, gin.H{"error": code})
		return
	}
	c.JSON(status, gin.H{"error": code, "details": details})
}
