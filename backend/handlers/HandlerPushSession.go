package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"friendship/utils"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PushSessionReader interface {
	HasActiveSessionForUser(ctx context.Context, userID uint64, sessionID string) (bool, error)
}

type PushSessionHandler struct {
	sessions PushSessionReader
}

func NewPushSessionHandler(sessions PushSessionReader) *PushSessionHandler {
	return &PushSessionHandler{sessions: sessions}
}

func (h *PushSessionHandler) Check(c *gin.Context) {
	if c.Request.URL.RawQuery != "" {
		utils.BadRequest(c, "invalid_request")
		return
	}
	if !hasJSONContentType(c.Request) {
		utils.JSONError(c, http.StatusUnsupportedMediaType, "unsupported_media_type")
		return
	}
	var request struct {
		SchemaVersion int    `json:"schemaVersion"`
		UserID        uint64 `json:"userId"`
		SessionID     string `json:"sessionId"`
	}
	if err := DecodeStrictJSONObject(c.Writer, c.Request, pushDeviceBodyLimit, &request, "schemaVersion", "userId", "sessionId"); err != nil {
		if errors.Is(err, ErrStrictJSONRequestTooLarge) {
			utils.JSONError(c, http.StatusRequestEntityTooLarge, "request_too_large")
			return
		}
		utils.BadRequest(c, "invalid_request")
		return
	}
	if request.SchemaVersion != 1 {
		utils.BadRequest(c, "unsupported_schema_version")
		return
	}
	sessionBytes, err := base64.RawURLEncoding.Strict().DecodeString(request.SessionID)
	if request.UserID == 0 || request.UserID > math.MaxInt64 || err != nil || len(sessionBytes) != 18 || base64.RawURLEncoding.EncodeToString(sessionBytes) != request.SessionID {
		utils.BadRequest(c, "invalid_request")
		return
	}
	if h.sessions == nil {
		utils.JSONError(c, http.StatusServiceUnavailable, "authentication_unavailable")
		return
	}
	active, err := h.sessions.HasActiveSessionForUser(c.Request.Context(), request.UserID, request.SessionID)
	if err != nil {
		utils.JSONError(c, http.StatusServiceUnavailable, "authentication_unavailable")
		return
	}
	c.JSON(http.StatusOK, struct {
		SchemaVersion int  `json:"schemaVersion"`
		Active        bool `json:"active"`
	}{SchemaVersion: 1, Active: active})
}
