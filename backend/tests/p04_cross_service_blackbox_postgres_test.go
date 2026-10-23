package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/middlewares"
	eventmodels "friendship/models/events"
	"friendship/routes"
	servicesevents "friendship/services/events"
	"friendship/services/statistics"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestP04StatisticsDowntimeLostResponseBackfillCrossServiceBlackBoxPostgres(t *testing.T) {
	backendDSN := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	notifyDSN := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if backendDSN == "" || notifyDSN == "" {
		t.Skip("оба PostgreSQL test DSN обязательны для P0.4 cross-process recovery; SQLite не заменяет эту проверку")
	}
	const internalToken = "p04-blackbox-internal-token"
	const jwtKey = "p04-blackbox-jwt-at-least-32-bytes"
	db := newPostgresEventCommandDB(t, backendDSN)
	groupID, application := statisticsFixture(t, db)
	start := time.Now().UTC().Add(-3 * time.Hour)
	ids := []uint{
		seedStatisticsEvent(t, db, groupID, 1, 1, start, 30, []uint{2, 3}, nil),
		seedStatisticsEvent(t, db, groupID, 1, 1, start, 45, []uint{2, 3, 4}, nil),
		seedStatisticsEvent(t, db, groupID, 1, 1, start, 60, []uint{2}, nil),
	}
	activeID := seedLifecycleStatus(t, db, eventmodels.StatusActive)
	for _, id := range ids[:2] {
		if err := db.Model(&eventmodels.Event{}).Where("id = ?", id).Update("status_id", activeID).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := &testPostgresRepository{db: db}
	lifecycle := servicesevents.NewEventLifecycleService(servicesevents.NewGORMEventLifecycleStore(repo), nil)
	triggerService := servicesevents.NewStatisticsTriggerService(servicesevents.NewGORMStatisticsTriggerReader(repo), servicesevents.NewGORMStatisticsTriggerBackfiller(repo))
	statsHandler := handlers.NewStatisticsHandler(application, triggerService)
	gin.SetMode(gin.TestMode)
	internalRouter := gin.New()
	routes.RegisterInternalStatisticsRoutes(internalRouter, statsHandler, middlewares.NewInternalTokenMiddleware(internalToken))
	var lostResponses, processCalls atomic.Int32
	monolithServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/statistics/process") {
			processCalls.Add(1)
			if lostResponses.CompareAndSwap(0, 1) {
				// Команда коммитит факты, затем рвём соединение до отправки успешного HTTP-ответа.
				recorder := httptest.NewRecorder()
				internalRouter.ServeHTTP(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Errorf("first processing status/body=%d/%s", recorder.Code, recorder.Body.String())
				}
				connection, _, err := writer.(http.Hijacker).Hijack()
				if err != nil {
					t.Errorf("hijack response: %v", err)
					return
				}
				_ = connection.Close()
				return
			}
		}
		internalRouter.ServeHTTP(writer, request)
	}))
	defer monolithServer.Close()
	if result, err := lifecycle.Advance(context.Background(), ids[0]); err != nil || !result.Applied {
		t.Fatalf("complete while notify down=%#v/%v", result, err)
	}
	before, err := application.Read(context.Background(), 2)
	if err != nil || before.CompletedEvents != 0 {
		t.Fatalf("statistics before notify start=%#v/%v", before, err)
	}
	publicRouter := newBlackBoxStatisticsRouter(t, statsHandler, 2, jwtKey)
	jwtToken := signBlackBoxJWT(t, 2, jwtKey)
	unauthorized := performBlackBoxRequest(publicRouter, http.MethodGet, "/api/v2/users/me/statistics?userId=1", "invalid")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("JWT authorization status=%d", unauthorized.Code)
	}
	notifyURL, stopFirst := startNotifyServiceWithStateForBlackBox(t, notifyDSN, monolithServer.URL, internalToken, true)
	firstStopped := false
	defer func() {
		if !firstStopped {
			stopFirst()
		}
	}()
	waitForBlackBoxReadiness(t, notifyURL)
	first := waitForBlackBoxStatistics(t, publicRouter, jwtToken, 1)
	if first.PlannedMinutes != 30 || first.UniquePeopleCount != nil || first.UniquePeopleStatus != statistics.UniquePeopleStatusNotCalculated {
		t.Fatalf("first contribution=%#v", first)
	}
	assertBlackBoxStatisticsJobs(t, notifyDSN, 1)
	if lostResponses.Load() != 1 || processCalls.Load() < 2 {
		t.Fatalf("lost response/retries=%d/%d", lostResponses.Load(), processCalls.Load())
	}
	stopFirst()
	firstStopped = true
	if result, err := lifecycle.Advance(context.Background(), ids[1]); err != nil || !result.Applied {
		t.Fatalf("complete during downtime=%#v/%v", result, err)
	}
	firstBackfill, err := triggerService.Backfill(context.Background(), 0, 2)
	if err != nil || firstBackfill.Scanned != 2 || firstBackfill.Published != 0 || !firstBackfill.HasMore {
		t.Fatalf("bounded backfill first page=%#v/%v", firstBackfill, err)
	}
	secondBackfill, err := triggerService.Backfill(context.Background(), firstBackfill.NextEventID, 2)
	if err != nil || secondBackfill.Scanned != 1 || secondBackfill.Published != 1 || secondBackfill.HasMore {
		t.Fatalf("resumed backfill=%#v/%v", secondBackfill, err)
	}
	beforeRestart, err := application.Read(context.Background(), 2)
	if err != nil || beforeRestart.CompletedEvents != 1 {
		t.Fatalf("downtime unexpectedly processed=%#v/%v", beforeRestart, err)
	}
	notifyURL, stopSecond := startNotifyServiceWithStateForBlackBox(t, notifyDSN, monolithServer.URL, internalToken, false)
	defer stopSecond()
	waitForBlackBoxReadiness(t, notifyURL)
	final := waitForBlackBoxStatistics(t, publicRouter, jwtToken, 3)
	if final.PlannedMinutes != 135 || final.UniquePeopleCount != nil || final.UniquePeopleStatus != statistics.UniquePeopleStatusNotCalculated || final.CreatedCompletedEvents != 0 || final.Timezone != "UTC" {
		t.Fatalf("recovered aggregates=%#v", final)
	}
	assertBlackBoxStatisticsJobs(t, notifyDSN, 3)
	replay, err := triggerService.Backfill(context.Background(), 0, 500)
	if err != nil || replay.Published != 0 {
		t.Fatalf("repeated backfill=%#v/%v", replay, err)
	}
	view, err := application.Read(context.Background(), 1)
	if err != nil || view.CreatedCompletedEvents != 3 || view.CompletedEvents != 0 {
		t.Fatalf("organizer statistics=%#v/%v", view, err)
	}
}

func newBlackBoxStatisticsRouter(t *testing.T, handler *handlers.StatisticsHandler, userID uint, jwtKey string) *gin.Engine {
	t.Helper()
	router := gin.New()
	auth := func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtKey), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		subject, err := token.Claims.GetSubject()
		if err != nil || subject != strconv.FormatUint(uint64(userID), 10) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
	routes.RegisterStatisticsRoutes(router, handler, auth)
	return router
}

func waitForBlackBoxStatistics(t *testing.T, router http.Handler, token string, count int64) statistics.View {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var view statistics.View
	for time.Now().Before(deadline) {
		response := performBlackBoxRequest(router, http.MethodGet, "/api/v2/users/me/statistics?userId=1", token)
		if response.Code == http.StatusOK && json.Unmarshal(response.Body.Bytes(), &view) == nil && view.CompletedEvents == count {
			return view
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("statistics count %d not reached: %#v", count, view)
	return statistics.View{}
}

func assertBlackBoxStatisticsJobs(t *testing.T, dsn string, count int64) {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	deadline := time.Now().Add(10 * time.Second)
	var completed, receipts int64
	for time.Now().Before(deadline) {
		err = db.QueryRow(`SELECT COUNT(*) FROM notify_service.event_statistics_jobs WHERE state='completed'`).Scan(&completed)
		if err == nil && completed == count {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil || completed != count {
		t.Fatalf("completed durable jobs=%d/%v, want %d", completed, err, count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM notify_service.event_statistics_source_receipts`).Scan(&receipts); err != nil || receipts != count {
		t.Fatalf("durable receipts=%d/%v, want %d", receipts, err, count)
	}
	var attempts int64
	if err := db.QueryRow(`SELECT MIN(attempt_count) FROM notify_service.event_statistics_jobs WHERE source_sequence=1`).Scan(&attempts); err != nil || attempts < 1 {
		t.Fatalf("lost-response persisted attempts=%d/%v", attempts, err)
	}
}
