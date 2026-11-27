package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DefaultRequestLogger() gin.HandlerFunc {
	standard := gin.Logger()
	protected := gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		return fmt.Sprintf("[GIN] %s %s %d %s\n", params.Method, pushRouteForLog(params.Request.URL.Path), params.StatusCode, params.Latency)
	})
	return func(c *gin.Context) {
		if pushRouteForLog(c.Request.URL.Path) != "" {
			protected(c)
			return
		}
		standard(c)
	}
}

func DefaultRecovery() gin.HandlerFunc {
	return recoveryWithPushRedaction(gin.Recovery())
}

func PushNoStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		if pushRouteForLog(c.Request.URL.Path) != "" {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	}
}

func recoveryWithPushRedaction(standard gin.HandlerFunc) gin.HandlerFunc {
	protected := gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})
	return func(c *gin.Context) {
		if pushRouteForLog(c.Request.URL.Path) != "" {
			protected(c)
			return
		}
		standard(c)
	}
}

func pushRouteForLog(path string) string {
	const devicesPrefix = "/api/v2/users/me/push-devices"
	const sessionsPath = "/internal/v1/push-sessions/check"
	switch {
	case path == devicesPrefix || strings.HasPrefix(path, devicesPrefix+"/"):
		return devicesPrefix + "/:installationId"
	case path == sessionsPath || strings.HasPrefix(path, sessionsPath+"/"):
		return sessionsPath
	default:
		return ""
	}
}
