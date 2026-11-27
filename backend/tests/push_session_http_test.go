package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"

	"github.com/gin-gonic/gin"
)

type pushSessionReaderStub struct {
	active bool
	err    error
	calls  []pushSessionCheckCall
}

type pushSessionCheckCall struct {
	userID    uint64
	sessionID string
	ctx       context.Context
}

func (s *pushSessionReaderStub) HasActiveSessionForUser(ctx context.Context, userID uint64, sessionID string) (bool, error) {
	s.calls = append(s.calls, pushSessionCheckCall{userID: userID, sessionID: sessionID, ctx: ctx})
	return s.active, s.err
}

func TestInternalPushSessionCheckContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("только внутренний токен и корректный запрос достигают reader", func(t *testing.T) {
		reader := &pushSessionReaderStub{active: true}
		router := newPushSessionCheckRouter(reader)

		for _, token := range []string{"", "wrong"} {
			response := performPushSessionCheckRequest(router, token, `{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR"}`)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("token=%q status/body=%d/%s", token, response.Code, response.Body.String())
			}
		}
		if len(reader.calls) != 0 {
			t.Fatalf("неавторизованный запрос достиг reader: %#v", reader.calls)
		}

		requestContext := context.WithValue(context.Background(), pushSessionContextKey{}, "request-context")
		request := httptest.NewRequest(http.MethodPost, "/internal/v1/push-sessions/check", strings.NewReader(`{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR"}`)).WithContext(requestContext)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(middlewares.InternalTokenHeader, "push-session-secret")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status/body=%d/%s", response.Code, response.Body.String())
		}
		var body struct {
			SchemaVersion int  `json:"schemaVersion"`
			Active        bool `json:"active"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.SchemaVersion != 1 || !body.Active {
			t.Fatalf("response = %#v", body)
		}
		if len(reader.calls) != 1 || reader.calls[0].userID != 7 || reader.calls[0].sessionID != "AAECAwQFBgcICQoLDA0ODxAR" || reader.calls[0].ctx != requestContext {
			t.Fatalf("reader calls = %#v", reader.calls)
		}
	})

	t.Run("отсутствующая или чужая сессия не раскрывается", func(t *testing.T) {
		reader := &pushSessionReaderStub{active: false}
		response := performPushSessionCheckRequest(newPushSessionCheckRouter(reader), "push-session-secret", `{"schemaVersion":1,"userId":8,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR"}`)
		if response.Code != http.StatusOK {
			t.Fatalf("status/body=%d/%s", response.Code, response.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body["active"] != false || len(reader.calls) != 1 || reader.calls[0].userID != 8 {
			t.Fatalf("response/calls = %#v/%#v", body, reader.calls)
		}
	})

	t.Run("некорректный запрос не достигает reader", func(t *testing.T) {
		for _, body := range []string{
			``,
			`{}`,
			`{"schemaVersion":2,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR"}`,
			`{"schemaVersion":1,"userId":0,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR"}`,
			`{"schemaVersion":1,"userId":7,"sessionId":"invalid"}`,
		} {
			reader := &pushSessionReaderStub{active: true}
			response := performPushSessionCheckRequest(newPushSessionCheckRouter(reader), "push-session-secret", body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("body=%s status/body=%d/%s", body, response.Code, response.Body.String())
			}
			if len(reader.calls) != 0 {
				t.Fatalf("body=%s invalid request reached reader", body)
			}
		}
	})

	t.Run("ошибка redis не маскируется как неактивная сессия", func(t *testing.T) {
		reader := &pushSessionReaderStub{err: errors.New("redis unavailable")}
		response := performPushSessionCheckRequest(newPushSessionCheckRouter(reader), "push-session-secret", `{"schemaVersion":1,"userId":7,"sessionId":"AAECAwQFBgcICQoLDA0ODxAR"}`)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status/body=%d/%s", response.Code, response.Body.String())
		}
		if len(reader.calls) != 1 {
			t.Fatalf("reader calls = %#v", reader.calls)
		}
	})
}

type pushSessionContextKey struct{}

func newPushSessionCheckRouter(reader *pushSessionReaderStub) *gin.Engine {
	router := gin.New()
	routes.RegisterInternalPushSessionRoutes(
		router,
		handlers.NewPushSessionHandler(reader),
		middlewares.NewInternalTokenMiddleware("push-session-secret"),
	)
	return router
}

func performPushSessionCheckRequest(router http.Handler, token string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/push-sessions/check", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(middlewares.InternalTokenHeader, token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
