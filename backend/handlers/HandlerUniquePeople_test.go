package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services/statistics"

	"github.com/gin-gonic/gin"
)

type uniquePeopleApplicationStub struct {
	calls int
	limit int
	ctx   context.Context
	claim statistics.UniquePeopleClaim
	err   error
}

func (s *uniquePeopleApplicationStub) Claim(ctx context.Context, limit int) (statistics.UniquePeopleClaims, error) {
	s.calls++
	s.ctx = ctx
	s.limit = limit
	return statistics.UniquePeopleClaims{}, s.err
}
func (s *uniquePeopleApplicationStub) Execute(ctx context.Context, claim statistics.UniquePeopleClaim) (statistics.UniquePeopleExecution, error) {
	s.calls++
	s.ctx = ctx
	s.claim = claim
	return statistics.UniquePeopleExecution{Outcome: "applied"}, s.err
}
func (s *uniquePeopleApplicationStub) Backfill(ctx context.Context, after uint, limit int) (statistics.UniquePeopleBackfillResult, error) {
	s.calls++
	s.ctx = ctx
	s.limit = limit
	return statistics.UniquePeopleBackfillResult{}, s.err
}
func (s *uniquePeopleApplicationStub) Metrics(ctx context.Context) (statistics.UniquePeopleMetrics, error) {
	s.calls++
	s.ctx = ctx
	return statistics.UniquePeopleMetrics{}, s.err
}

func uniquePeopleTestRouter(s *uniquePeopleApplicationStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	routes.RegisterInternalUniquePeopleRoutes(r, handlers.NewUniquePeopleHandler(s), middlewares.NewInternalTokenMiddleware("test-secret"))
	return r
}

func TestUniquePeopleRoutesAuthenticateEveryCommand(t *testing.T) {
	s := &uniquePeopleApplicationStub{}
	r := uniquePeopleTestRouter(s)
	for _, endpoint := range []string{"claim", "execute", "backfill", "metrics"} {
		method := http.MethodPost
		if endpoint == "metrics" {
			method = http.MethodGet
		}
		for _, token := range []string{"", "wrong"} {
			req := httptest.NewRequest(method, "/internal/v1/statistics/unique-people/"+endpoint, nil)
			req.Header.Set("X-Internal-Token", token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("%s=%d", endpoint, w.Code)
			}
		}
	}
	if s.calls != 0 {
		t.Fatalf("вызовы без токена: %d", s.calls)
	}
}

func TestUniquePeopleHandlerBoundsAndContext(t *testing.T) {
	s := &uniquePeopleApplicationStub{}
	r := uniquePeopleTestRouter(s)
	for _, path := range []string{"claim?limit=0", "claim?limit=9", "claim?limit=bad", "backfill?afterUserId=-1", "backfill?limit=501"} {
		req := httptest.NewRequest(http.MethodPost, "/internal/v1/statistics/unique-people/"+path, nil)
		req.Header.Set("X-Internal-Token", "test-secret")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s=%d", path, w.Code)
		}
	}
	if s.calls != 0 {
		t.Fatal("невалидный запрос достиг приложения")
	}
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "unique")
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/statistics/unique-people/claim?limit=2", nil).WithContext(ctx)
	req.Header.Set("X-Internal-Token", "test-secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || s.limit != 2 || s.ctx.Value(contextKey{}) != "unique" {
		t.Fatalf("status=%d stub=%+v", w.Code, s)
	}
	for _, body := range []string{"{", `{} {}`, `{"schemaVersion":2}`, strings.Repeat("x", 4097)} {
		req = httptest.NewRequest(http.MethodPost, "/internal/v1/statistics/unique-people/execute", strings.NewReader(body))
		req.Header.Set("X-Internal-Token", "test-secret")
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("body accepted: %d", w.Code)
		}
	}
	s.err = statistics.ErrInvalidInput
	req = httptest.NewRequest(http.MethodPost, "/internal/v1/statistics/unique-people/execute", strings.NewReader(`{"userId":1,"generation":1,"leaseToken":"token"}`))
	req.Header.Set("X-Internal-Token", "test-secret")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("application validation=%d", w.Code)
	}
}

func TestStatisticsViewUnknownUniquePeopleIsNull(t *testing.T) {
	view := statistics.View{UniquePeopleStatus: "not_calculated"}
	body, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["uniquePeopleCount"] != nil || fields["uniquePeopleCalculatedAt"] != nil || fields["uniquePeopleStatus"] != "not_calculated" {
		t.Fatalf("unknown=%s", body)
	}
}
