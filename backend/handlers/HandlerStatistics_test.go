package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services/events"
	"friendship/services/statistics"

	"github.com/gin-gonic/gin"
)

type statisticsApplicationStub struct {
	userID, eventID  uint
	reads, processes int
	ctx              context.Context
	err              error
}

func (s *statisticsApplicationStub) Read(ctx context.Context, id uint) (statistics.View, error) {
	s.userID, s.ctx = id, ctx
	s.reads++
	return statistics.View{CompletedEvents: 7, Timezone: "UTC", Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}, s.err
}
func (s *statisticsApplicationStub) Process(ctx context.Context, id uint) (statistics.ProcessResult, error) {
	s.eventID, s.ctx = id, ctx
	s.processes++
	return statistics.ProcessResult{EventID: id, Outcome: "already_processed"}, s.err
}

type statisticsTriggersStub struct {
	after                   int64
	afterEvent              uint
	limit, lists, backfills int
}

func (s *statisticsTriggersStub) List(_ context.Context, after int64, limit int) (events.StatisticsTriggersPage, error) {
	s.after, s.limit = after, limit
	s.lists++
	return events.StatisticsTriggersPage{Messages: []events.StatisticsTriggerMessage{}, NextCursor: after}, nil
}
func (s *statisticsTriggersStub) Backfill(_ context.Context, after uint, limit int) (events.StatisticsBackfillResult, error) {
	s.afterEvent, s.limit = after, limit
	s.backfills++
	return events.StatisticsBackfillResult{NextEventID: after}, nil
}

func TestStatisticsHandlerUsesAuthenticatedIdentityAndRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := &statisticsApplicationStub{}
	router := gin.New()
	routes.RegisterStatisticsRoutes(router, handlers.NewStatisticsHandler(application, nil), func(c *gin.Context) { c.Set("userID", uint(7)); c.Next() })
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "statistics")
	request := httptest.NewRequest(http.MethodGet, "/api/v2/users/me/statistics?userId=999", nil).WithContext(ctx)
	request.Header.Set("X-User-ID", "999")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || application.userID != 7 || application.ctx != ctx || application.reads != 1 {
		t.Fatalf("response/call = %d/%#v, body=%s", response.Code, application, response.Body.String())
	}
	var view statistics.View
	if err := json.Unmarshal(response.Body.Bytes(), &view); err != nil || view.CompletedEvents != 7 || view.Timezone != "UTC" || view.Types == nil {
		t.Fatalf("view/error = %#v/%v", view, err)
	}
	application.err = errors.New("storage failed")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("failure status = %d", response.Code)
	}
}

func TestStatisticsHandlerRejectsMissingOrInvalidIdentity(t *testing.T) {
	for _, identity := range []any{nil, uint(0), "7", -1} {
		application := &statisticsApplicationStub{}
		router := gin.New()
		routes.RegisterStatisticsRoutes(router, handlers.NewStatisticsHandler(application, nil), func(c *gin.Context) {
			if identity != nil {
				c.Set("userID", identity)
			}
			c.Next()
		})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/users/me/statistics?userId=7", nil))
		if response.Code != http.StatusUnauthorized || application.reads != 0 {
			t.Fatalf("identity %v response/calls=%d/%d", identity, response.Code, application.reads)
		}
	}
}

func TestStatisticsInternalRoutesRequireTokenAndValidateInput(t *testing.T) {
	application := &statisticsApplicationStub{}
	triggers := &statisticsTriggersStub{}
	router := gin.New()
	routes.RegisterInternalStatisticsRoutes(router, handlers.NewStatisticsHandler(application, triggers), middlewares.NewInternalTokenMiddleware("statistics-secret"))
	for _, item := range []struct{ method, path string }{
		{http.MethodGet, "/internal/v1/statistics/triggers"},
		{http.MethodPost, "/internal/v1/statistics/backfill"},
		{http.MethodPost, "/internal/v1/events/5/statistics/process"},
	} {
		for _, token := range []string{"", "wrong"} {
			request := httptest.NewRequest(item.method, item.path, nil)
			request.Header.Set("X-Internal-Token", token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("%s token=%q status=%d", item.path, token, response.Code)
			}
		}
	}
	if application.processes != 0 || triggers.lists != 0 || triggers.backfills != 0 {
		t.Fatal("неавторизованный запрос достиг приложения")
	}
	for _, path := range []string{
		"/internal/v1/statistics/triggers?afterSequence=-1", "/internal/v1/statistics/triggers?afterSequence=abc",
		"/internal/v1/statistics/triggers?limit=0", "/internal/v1/statistics/triggers?limit=501",
		"/internal/v1/statistics/triggers?limit=bad", "/internal/v1/statistics/backfill?afterEventId=-1",
		"/internal/v1/statistics/backfill?limit=501", "/internal/v1/events/0/statistics/process", "/internal/v1/events/x/statistics/process",
	} {
		method := http.MethodPost
		if len(path) >= len("/internal/v1/statistics/triggers") && path[:len("/internal/v1/statistics/triggers")] == "/internal/v1/statistics/triggers" {
			method = http.MethodGet
		}
		request := httptest.NewRequest(method, path, nil)
		request.Header.Set("X-Internal-Token", "statistics-secret")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("invalid %s status/body=%d/%s", path, response.Code, response.Body.String())
		}
	}
	for _, item := range []struct{ method, path string }{
		{http.MethodGet, "/internal/v1/statistics/triggers?afterSequence=31&limit=500"},
		{http.MethodPost, "/internal/v1/statistics/backfill?afterEventId=12&limit=4"},
		{http.MethodPost, "/internal/v1/events/5/statistics/process"},
	} {
		request := httptest.NewRequest(item.method, item.path, nil)
		request.Header.Set("X-Internal-Token", "statistics-secret")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("valid %s status/body=%d/%s", item.path, response.Code, response.Body.String())
		}
	}
	if triggers.after != 31 || triggers.afterEvent != 12 || triggers.limit != 4 || triggers.lists != 1 || triggers.backfills != 1 || application.eventID != 5 || application.processes != 1 {
		t.Fatalf("arguments/counts = %#v / %#v", triggers, application)
	}
}
