package middlewares

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"friendship/utils"

	"github.com/gin-gonic/gin"
)

const InternalTokenHeader = "X-Internal-Token"

func NewInternalTokenMiddleware(configuredToken string) gin.HandlerFunc {
	expected := sha256.Sum256([]byte(configuredToken))
	configured := strings.TrimSpace(configuredToken) != ""

	return func(c *gin.Context) {
		provided := c.GetHeader(InternalTokenHeader)
		actual := sha256.Sum256([]byte(provided))
		if !configured || provided == "" || subtle.ConstantTimeCompare(actual[:], expected[:]) != 1 {
			utils.AbortJSONError(c, http.StatusUnauthorized, "invalid_internal_token", utils.WithMessage("Неверный токен внутреннего сервиса"))
			return
		}
		c.Next()
	}
}
