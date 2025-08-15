package transport

import (
	"context"
	"friendship/services"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type finishedSessionPayload struct {
	SessionID uint `json:"session_id" binding:"required"`
}

func UpdateStatisticsHandler(c *gin.Context) {
	// Авторизация по внутреннему токену
	reqToken := strings.TrimSpace(c.GetHeader("X-Internal-Token"))
	if reqToken == "" || reqToken != os.Getenv("NOTIFY_SERVICE_TOKEN") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var p finishedSessionPayload
	if err := c.ShouldBindJSON(&p); err != nil || p.SessionID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Таймаут на весь процесс
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := services.UpdateStatisticsForFinishedSession(ctx, p.SessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
