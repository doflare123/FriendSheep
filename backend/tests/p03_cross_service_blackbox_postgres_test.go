package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/middlewares"
	groupmodels "friendship/models/groups"
	"friendship/routes"
	servicesevents "friendship/services/events"
	"friendship/services/notifications"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestP03EventReminderCrossServiceBlackBoxPostgres(t *testing.T) {
	backendDSN := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	notifyDSN := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if backendDSN == "" || notifyDSN == "" {
		t.Skip("FRIENDSHEEP_TEST_POSTGRES_DSN и NOTIFY_SERVICE_TEST_POSTGRES_DSN обязательны для P0.3 cross-service black-box; SQLite не заменяет PostgreSQL")
	}

	const (
		userID        = uint(1)
		internalToken = "p03-blackbox-internal-token"
		jwtKey        = "p03-blackbox-jwt-key-at-least-32-bytes"
	)

	db := newPostgresEventCommandDB(t, backendDSN)
	seedEventReferences(t, db)
	seedEventUser(t, db, userID)
	groupID := seedEventGroup(t, db, userID, false)
	seedEventGroupMembershipWithRole(t, db, userID, groupID, groupmodels.RoleAdmin)
	genreID := seedEventGenre(t, db, "P0.3 black-box")

	input := validCreateEventCommandInput()
	input.GroupID = groupID
	input.EventTypeID = 1
	input.LocationID = 1
	input.AgeLimitID = 1
	input.Genres = []uint{genreID}
	input.StartTime = time.Now().UTC().Add(30 * time.Minute)

	eventService := servicesevents.NewEventCommandService(
		&testLogger{},
		servicesevents.NewGORMEventUnitOfWork(&testPostgresRepository{db: db}),
	)
	created, err := eventService.CreateEvent(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("CreateEvent(): %v", err)
	}

	gin.SetMode(gin.TestMode)
	internalRouter := gin.New()
	repo := &testPostgresRepository{db: db}
	intentHandler := handlers.NewEventReminderIntentOutboxHandler(
		servicesevents.NewEventReminderIntentService(servicesevents.NewGORMEventReminderIntentOutboxReader(repo)),
	)
	recipientService := notifications.NewEventReminderRecipientService(
		notifications.NewGORMEventReminderRecipientStore(repo),
		notifications.NewDefaultEventReminderPreferenceReader(),
	)
	recipientHandler := handlers.NewEventReminderRecipientsHandler(recipientService)
	internalAuth := middlewares.NewInternalTokenMiddleware(internalToken)
	routes.RegisterInternalEventReminderIntentRoutes(internalRouter, intentHandler, internalAuth)
	routes.RegisterInternalEventReminderRecipientRoutes(internalRouter, recipientHandler, internalAuth)
	monolithServer := httptest.NewServer(internalRouter)
	defer monolithServer.Close()

	notifyURL, stopNotify := startNotifyServiceForBlackBox(t, notifyDSN, monolithServer.URL, internalToken)
	defer stopNotify()
	waitForBlackBoxReadiness(t, notifyURL)

	inboxClient, err := notifications.NewHTTPNotificationInboxClient(notifyURL, internalToken, 2*time.Second)
	if err != nil {
		t.Fatalf("NewHTTPNotificationInboxClient(): %v", err)
	}
	publicRouter := newBlackBoxNotificationRouter(t, inboxClient, userID, jwtKey)
	jwtToken := signBlackBoxJWT(t, userID, jwtKey)

	var page notifications.InboxPage
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		response := performBlackBoxRequest(publicRouter, http.MethodGet, "/api/v2/users/me/notifications?userId=999&limit=10&unread=true", jwtToken)
		if response.Code == http.StatusOK {
			if err := json.Unmarshal(response.Body.Bytes(), &page); err == nil && len(page.Items) == 1 {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if len(page.Items) != 1 {
		t.Fatalf("inbox не получил reminder event=%d до deadline: %#v", created.ID, page)
	}
	notification := page.Items[0]
	if notification.ResourceID != created.ID || notification.ReminderOffsetMinutes != 60 ||
		notification.Payload.SchemaVersion != 1 || notification.Payload.EventID != created.ID || notification.ReadAt != nil {
		t.Fatalf("notification = %#v", notification)
	}

	countResponse := performBlackBoxRequest(publicRouter, http.MethodGet, "/api/v2/users/me/notifications/unread-count?userId=999", jwtToken)
	if countResponse.Code != http.StatusOK {
		t.Fatalf("unread-count status/body = %d/%s", countResponse.Code, countResponse.Body.String())
	}
	var unread notifications.UnreadCount
	if err := json.Unmarshal(countResponse.Body.Bytes(), &unread); err != nil || unread.UnreadCount != 1 {
		t.Fatalf("unread-count = %#v error:%v", unread, err)
	}

	markPath := "/api/v2/users/me/notifications/" + notification.ID + "/read?userId=999"
	firstMark := performBlackBoxRequest(publicRouter, http.MethodPatch, markPath, jwtToken)
	secondMark := performBlackBoxRequest(publicRouter, http.MethodPatch, markPath, jwtToken)
	if firstMark.Code != http.StatusOK || secondMark.Code != http.StatusOK {
		t.Fatalf("mark-read statuses = %d/%d; bodies = %s / %s", firstMark.Code, secondMark.Code, firstMark.Body.String(), secondMark.Body.String())
	}
	var firstResult, secondResult notifications.MarkReadResult
	if err := json.Unmarshal(firstMark.Body.Bytes(), &firstResult); err != nil {
		t.Fatalf("decode first mark-read: %v", err)
	}
	if err := json.Unmarshal(secondMark.Body.Bytes(), &secondResult); err != nil {
		t.Fatalf("decode second mark-read: %v", err)
	}
	if firstResult.ID != notification.ID || !firstResult.ReadAt.Equal(secondResult.ReadAt) {
		t.Fatalf("idempotent mark-read = first:%#v second:%#v", firstResult, secondResult)
	}
}

func startNotifyServiceForBlackBox(t *testing.T, dsn string, monolithURL string, token string) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve notify port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	notifyRoot := filepath.Clean(filepath.Join("..", "..", "notify_service"))
	executableName := "notify_service_blackbox"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executable := filepath.Join(t.TempDir(), executableName)
	build := exec.Command("go", "build", "-buildvcs=false", "-o", executable, ".")
	build.Dir = notifyRoot
	build.Stdout = io.Discard
	build.Stderr = io.Discard
	if err := build.Run(); err != nil {
		t.Fatalf("build notify_service black-box binary: %v", err)
	}

	runContext, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(runContext, executable)
	command.Dir = notifyRoot
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	command.Env = append(os.Environ(),
		"NOTIFY_DATABASE_URL="+dsn,
		"NOTIFY_MONOLITH_BASE_URL="+monolithURL,
		"NOTIFY_SERVICE_TOKEN="+token,
		"PORT="+strconv.Itoa(port),
		"NOTIFY_SOURCE_POLL_INTERVAL=20ms",
		"NOTIFY_SOURCE_RETRY_MIN_BACKOFF=20ms",
		"NOTIFY_SOURCE_RETRY_MAX_BACKOFF=200ms",
		"NOTIFY_WORKER_SCAN_INTERVAL=20ms",
		"NOTIFY_WORKER_RETRY_MIN_BACKOFF=20ms",
		"NOTIFY_WORKER_RETRY_MAX_BACKOFF=200ms",
		"NOTIFY_MONOLITH_HTTP_TIMEOUT=2s",
		"NOTIFY_JOB_LEASE_DURATION=5s",
	)
	if err := command.Start(); err != nil {
		cancel()
		t.Fatalf("start notify_service black-box binary: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	stop := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			if command.Process != nil {
				_ = command.Process.Kill()
			}
			<-done
		}
	}
	return fmt.Sprintf("http://127.0.0.1:%d", port), stop
}

func waitForBlackBoxReadiness(t *testing.T, notifyURL string) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(notifyURL + "/ready")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("notify_service не стал ready до deadline")
}

func newBlackBoxNotificationRouter(t *testing.T, inbox notifications.NotificationInbox, userID uint, jwtKey string) *gin.Engine {
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
	routes.RegisterNotificationRoutes(router, handlers.NewNotificationsHandler(inbox), auth)
	return router
}

func signBlackBoxJWT(t *testing.T, userID uint, key string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(userID), 10),
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	signed, err := token.SignedString([]byte(key))
	if err != nil {
		t.Fatalf("sign black-box JWT: %v", err)
	}
	return signed
}

func performBlackBoxRequest(router http.Handler, method string, path string, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
