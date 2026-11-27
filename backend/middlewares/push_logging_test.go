package middlewares

import (
	"bytes"
	"errors"
	"friendship/utils"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPushLoggingRedactsRequestAndPanic(t *testing.T) {
	const credential = "push-secret-for-redaction-test"
	const pattern = "/api/v2/users/me/push-devices/:installationId"
	var output bytes.Buffer
	previousWriter, previousErrorWriter, previousLogger := gin.DefaultWriter, gin.DefaultErrorWriter, utils.Log
	gin.DefaultWriter, gin.DefaultErrorWriter = &output, &output
	utils.Log = &utils.Logger{Logger: log.New(&output, "", 0)}
	t.Cleanup(func() {
		gin.DefaultWriter, gin.DefaultErrorWriter, utils.Log = previousWriter, previousErrorWriter, previousLogger
	})
	t.Setenv("LOG_BODIES", "true")
	t.Setenv("LOG_HEADERS", "true")

	for _, logger := range []gin.HandlerFunc{DefaultRequestLogger(), RequestResponseLogger(), DetailedRequestLogger()} {
		router := gin.New()
		router.Use(logger, DefaultRecovery())
		router.PUT(pattern, func(c *gin.Context) {
			_ = c.Error(errors.New(credential))
			panic(credential)
		})
		output.Reset()
		request := httptest.NewRequest(http.MethodPut, "/api/v2/users/me/push-devices/installation?secret="+credential,
			strings.NewReader(`{"providerToken":"`+credential+`"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+credential)
		request.Header.Set("X-Installation-Secret", credential)
		request.Header.Set("X-Internal-Token", credential)
		result := httptest.NewRecorder()
		router.ServeHTTP(result, request)
		if result.Code != http.StatusInternalServerError || !strings.Contains(output.String(), pattern) {
			t.Fatal("журнал должен содержать маршрут и результат запроса")
		}
		if strings.Contains(output.String(), credential) || strings.Contains(result.Body.String(), credential) {
			t.Fatal("журнал или ответ содержит секрет запроса")
		}
	}
}
