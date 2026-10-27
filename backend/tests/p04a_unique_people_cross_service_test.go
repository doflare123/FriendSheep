package tests

import (
	"bytes"
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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/middlewares"
	statmodels "friendship/models/statistics"
	"friendship/routes"
	stats "friendship/services/statistics"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	p04aProxyPass int32 = iota
	p04aProxyLoseExecute
	p04aProxyLoseClaim
)

type p04aControlledClock struct {
	mu     sync.RWMutex
	offset time.Duration
}

func (c *p04aControlledClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Now().UTC().Add(c.offset)
}

func (c *p04aControlledClock) Advance(delta time.Duration) {
	c.mu.Lock()
	c.offset += delta
	c.mu.Unlock()
}

type p04aClaimRecord struct {
	UserID     uint
	Generation uint64
	LeaseToken string
}

type p04aMonolithProxy struct {
	t             *testing.T
	router        http.Handler
	token         string
	mode          atomic.Int32
	authenticated atomic.Int32
	lostExecute   chan struct{}
	lostClaim     chan p04aClaimRecord
	mu            sync.Mutex
	claims        []p04aClaimRecord
	executeTokens []string
	executeTimes  []time.Time
}

func (p *p04aMonolithProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	isUniquePath := strings.HasPrefix(request.URL.Path, "/internal/v1/statistics/unique-people/")
	if isUniquePath && request.Header.Get("X-Internal-Token") == p.token {
		p.authenticated.Add(1)
	}

	var executeClaim p04aClaimRecord
	if strings.HasSuffix(request.URL.Path, "/execute") {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			p.t.Errorf("read unique people execute request: %v", err)
		} else {
			request.Body = io.NopCloser(bytes.NewReader(body))
			var payload struct {
				UserID     uint   `json:"userId"`
				Generation uint64 `json:"generation"`
				LeaseToken string `json:"leaseToken"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				p.t.Errorf("decode unique people execute request: %v", err)
			} else {
				executeClaim = p04aClaimRecord{UserID: payload.UserID, Generation: payload.Generation, LeaseToken: payload.LeaseToken}
				p.mu.Lock()
				p.executeTokens = append(p.executeTokens, payload.LeaseToken)
				p.executeTimes = append(p.executeTimes, time.Now())
				p.mu.Unlock()
			}
		}
	}

	recorder := httptest.NewRecorder()
	p.router.ServeHTTP(recorder, request)

	var claimed []p04aClaimRecord
	if strings.HasSuffix(request.URL.Path, "/claim") && recorder.Code == http.StatusOK {
		var response stats.UniquePeopleClaims
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			p.t.Errorf("decode unique people claim response: %v", err)
		} else {
			for _, job := range response.Jobs {
				claimed = append(claimed, p04aClaimRecord{UserID: job.UserID, Generation: job.RequestedGeneration, LeaseToken: job.LeaseToken})
			}
			p.mu.Lock()
			p.claims = append(p.claims, claimed...)
			p.mu.Unlock()
		}
	}

	if p.mode.Load() == p04aProxyLoseExecute && executeClaim.UserID == 2 && recorder.Code == http.StatusOK {
		var result stats.UniquePeopleExecution
		if json.Unmarshal(recorder.Body.Bytes(), &result) == nil && result.Outcome == stats.UniquePeopleOutcomeApplied && p.mode.CompareAndSwap(p04aProxyLoseExecute, p04aProxyPass) {
			close(p.lostExecute)
			p.dropResponse(writer)
			return
		}
	}
	if p.mode.Load() == p04aProxyLoseClaim && len(claimed) == 1 && claimed[0].UserID == 2 && p.mode.CompareAndSwap(p04aProxyLoseClaim, p04aProxyPass) {
		p.lostClaim <- claimed[0]
		p.dropResponse(writer)
		return
	}

	for key, values := range recorder.Header() {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}
	writer.WriteHeader(recorder.Code)
	_, _ = writer.Write(recorder.Body.Bytes())
}

func (p *p04aMonolithProxy) dropResponse(writer http.ResponseWriter) {
	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		p.t.Error("HTTP writer не поддерживает hijack для controlled lost response")
		return
	}
	connection, _, err := hijacker.Hijack()
	if err != nil {
		p.t.Errorf("hijack controlled lost response: %v", err)
		return
	}
	_ = connection.Close()
}

func TestP04AUniquePeopleCrossServiceRecoveryPostgres(t *testing.T) {
	backendDSN, notifyDSN := p04aCrossServiceDSNs(t)
	const internalToken = "p04a-cross-service-internal-token"
	const jwtKey = "p04a-cross-service-jwt-at-least-32-bytes"

	db := newPostgresEventCommandDB(t, backendDSN)
	groupID, _ := statisticsFixture(t, db)
	repo := &testPostgresRepository{db: db}
	config := p04aFastUniquePeopleConfig()
	ordinary := stats.NewService(stats.NewGORMStoreWithUniquePeopleConfig(repo, config))
	clock := &p04aControlledClock{offset: 2 * time.Second}
	unique := stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(repo), config, clock.Now)

	statisticsHandler := handlers.NewStatisticsHandler(ordinary, nil)
	uniqueHandler := handlers.NewUniquePeopleHandler(unique)
	gin.SetMode(gin.TestMode)
	internalRouter := gin.New()
	routes.RegisterInternalUniquePeopleRoutes(internalRouter, uniqueHandler, middlewares.NewInternalTokenMiddleware(internalToken))
	proxy := &p04aMonolithProxy{
		t: t, router: internalRouter, token: internalToken,
		lostExecute: make(chan struct{}), lostClaim: make(chan p04aClaimRecord, 1),
	}
	monolithServer := httptest.NewServer(proxy)
	defer monolithServer.Close()

	publicRouter := newBlackBoxStatisticsRouter(t, statisticsHandler, 2, jwtKey)
	jwtToken := signBlackBoxJWT(t, 2, jwtKey)
	executable := buildP04aNotifyService(t)

	firstEvent := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-2*time.Hour), 30, []uint{2, 3}, nil)
	processP04aStatisticsEvent(t, ordinary, firstEvent)
	delayP04aJobs(t, db, clock.Now().Add(time.Hour), 3)

	notifyURL, stopFirst := startP04aNotifyService(t, executable, notifyDSN, monolithServer.URL, internalToken, true, true)
	firstRunning := true
	defer func() {
		if firstRunning {
			stopFirst()
		}
	}()
	waitForBlackBoxReadiness(t, notifyURL)
	waitForP04aUniqueView(t, publicRouter, jwtToken, 1, stats.UniquePeopleStatusFresh)
	assertP04aJob(t, db, 2, 1, 1, 1, 0, false)

	proxy.mode.Store(p04aProxyLoseExecute)
	secondEvent := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-90*time.Minute), 30, []uint{2, 4}, nil)
	processP04aStatisticsEvent(t, ordinary, secondEvent)
	delayP04aJobs(t, db, clock.Now().Add(time.Hour), 4)
	waitP04aSignal(t, proxy.lostExecute, "execute commit с потерянным HTTP-ответом")
	waitForP04aUniqueView(t, publicRouter, jwtToken, 2, stats.UniquePeopleStatusFresh)
	assertP04aJob(t, db, 2, 2, 2, 2, 0, false)
	stopFirst()
	firstRunning = false

	notifyURL, stopSecond := startP04aNotifyService(t, executable, notifyDSN, monolithServer.URL, internalToken, false, true)
	secondRunning := true
	defer func() {
		if secondRunning {
			stopSecond()
		}
	}()
	waitForBlackBoxReadiness(t, notifyURL)
	time.Sleep(150 * time.Millisecond)
	assertP04aJob(t, db, 2, 2, 2, 2, 0, false)

	proxy.mode.Store(p04aProxyLoseClaim)
	thirdEvent := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-time.Hour), 30, []uint{2, 3}, nil)
	processP04aStatisticsEvent(t, ordinary, thirdEvent)
	lostClaim := waitP04aClaim(t, proxy.lostClaim)
	if lostClaim.UserID != 2 || lostClaim.Generation != 3 {
		t.Fatalf("lost claim = %#v, ожидался user 2 generation 3", lostClaim)
	}
	stopSecond()
	secondRunning = false
	assertP04aJob(t, db, 2, 3, 2, 2, 0, true)

	clock.Advance(config.LeaseDuration + time.Millisecond)
	notifyURL, stopThird := startP04aNotifyService(t, executable, notifyDSN, monolithServer.URL, internalToken, false, true)
	defer stopThird()
	waitForBlackBoxReadiness(t, notifyURL)
	waitForP04aRetryRecovery(t, db, 2)
	clock.Advance(config.RetryMax + time.Millisecond)
	waitForP04aUniqueView(t, publicRouter, jwtToken, 2, stats.UniquePeopleStatusFresh)
	assertP04aJob(t, db, 2, 3, 3, 3, 1, false)
	time.Sleep(150 * time.Millisecond)
	assertP04aJob(t, db, 2, 3, 3, 3, 1, false)

	proxy.mu.Lock()
	claims := append([]p04aClaimRecord(nil), proxy.claims...)
	executedTokens := append([]string(nil), proxy.executeTokens...)
	proxy.mu.Unlock()
	var recovered bool
	for _, claim := range claims {
		if claim.UserID == lostClaim.UserID && claim.Generation == lostClaim.Generation && claim.LeaseToken != lostClaim.LeaseToken {
			recovered = true
			break
		}
	}
	if !recovered {
		t.Fatalf("same-user lease recovery отсутствует; lost=%#v claims=%#v", lostClaim, claims)
	}
	for _, token := range executedTokens {
		if token == lostClaim.LeaseToken {
			t.Fatalf("claim с потерянным ответом был неожиданно выполнен: %s", token)
		}
	}
	if proxy.authenticated.Load() == 0 {
		t.Fatal("реальный notify_service не выполнил authenticated unique people запросы")
	}
}

func TestP04AUniquePeopleDefaultPollingCrossServicePostgres(t *testing.T) {
	backendDSN, notifyDSN := p04aCrossServiceDSNs(t)
	const internalToken = "p04a-default-poll-internal-token"
	db := newPostgresEventCommandDB(t, backendDSN)
	groupID, _ := statisticsFixture(t, db)
	repo := &testPostgresRepository{db: db}
	config := p04aFastUniquePeopleConfig()
	ordinary := stats.NewService(stats.NewGORMStoreWithUniquePeopleConfig(repo, config))
	clock := &p04aControlledClock{offset: 2 * time.Second}
	unique := stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(repo), config, clock.Now)
	uniqueHandler := handlers.NewUniquePeopleHandler(unique)
	gin.SetMode(gin.TestMode)
	internalRouter := gin.New()
	routes.RegisterInternalUniquePeopleRoutes(internalRouter, uniqueHandler, middlewares.NewInternalTokenMiddleware(internalToken))
	proxy := &p04aMonolithProxy{t: t, router: internalRouter, token: internalToken, lostExecute: make(chan struct{}), lostClaim: make(chan p04aClaimRecord, 1)}
	monolithServer := httptest.NewServer(proxy)
	defer monolithServer.Close()

	eventID := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().UTC().Add(-time.Hour), 30, []uint{2, 3, 4}, nil)
	processP04aStatisticsEvent(t, ordinary, eventID)
	executable := buildP04aNotifyService(t)
	notifyURL, stopNotify := startP04aNotifyService(t, executable, notifyDSN, monolithServer.URL, internalToken, true, false)
	defer stopNotify()
	waitForBlackBoxReadiness(t, notifyURL)

	deadline := time.Now().Add(20 * time.Second)
	for {
		proxy.mu.Lock()
		times := append([]time.Time(nil), proxy.executeTimes...)
		proxy.mu.Unlock()
		if len(times) >= 3 {
			var completed int64
			if err := db.Model(&statmodels.UniquePeopleJob{}).
				Where("successful_runs = 1 AND applied_generation = requested_generation AND lease_token IS NULL").
				Count(&completed).Error; err != nil {
				t.Fatalf("read completed default-polling jobs: %v", err)
			}
			if completed != 3 {
				if time.Now().After(deadline) {
					t.Fatalf("default polling сохранил %d результатов вместо трёх", completed)
				}
				time.Sleep(25 * time.Millisecond)
				continue
			}
			firstGap := times[1].Sub(times[0])
			secondGap := times[2].Sub(times[1])
			elapsed := times[2].Sub(times[0])
			t.Logf("default unique people polling: first_gap=%s second_gap=%s elapsed_for_3=%s", firstGap, secondGap, elapsed)
			if firstGap < 4500*time.Millisecond || secondGap < 4500*time.Millisecond || elapsed < 9*time.Second {
				t.Fatalf("default polling выполняет batch слишком часто: gaps=%s/%s elapsed=%s", firstGap, secondGap, elapsed)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("default polling не обработал три ready jobs; execute timestamps=%v", times)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func p04aCrossServiceDSNs(t *testing.T) (string, string) {
	t.Helper()
	backendDSN := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	notifyDSN := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if backendDSN == "" || notifyDSN == "" {
		t.Skip("оба PostgreSQL test DSN обязательны для P0.4a cross-process recovery; SQLite не заменяет lease/retry проверку")
	}
	return backendDSN, notifyDSN
}

func p04aFastUniquePeopleConfig() stats.UniquePeopleConfig {
	config := stats.DefaultUniquePeopleConfig()
	config.QueryTimeout = 100 * time.Millisecond
	config.LeaseDuration = 500 * time.Millisecond
	config.InitialSpread = time.Nanosecond
	config.RetryBase = 50 * time.Millisecond
	config.RetryMax = 50 * time.Millisecond
	config.RetryJitter = 0
	return config
}

func processP04aStatisticsEvent(t *testing.T, service *stats.Service, eventID uint) {
	t.Helper()
	result, err := service.Process(context.Background(), eventID)
	if err != nil || result.Outcome != "processed" {
		t.Fatalf("process P0.4 event %d = %#v, %v", eventID, result, err)
	}
}

func delayP04aJobs(t *testing.T, db *gorm.DB, until time.Time, userIDs ...uint) {
	t.Helper()
	if err := db.Model(&statmodels.UniquePeopleJob{}).Where("user_id IN ?", userIDs).Update("next_attempt_at", until).Error; err != nil {
		t.Fatalf("delay non-target unique people jobs: %v", err)
	}
}

func waitForP04aUniqueView(t *testing.T, router http.Handler, token string, count int64, status string) stats.View {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	var view stats.View
	for time.Now().Before(deadline) {
		response := performBlackBoxRequest(router, http.MethodGet, "/api/v2/users/me/statistics", token)
		if response.Code == http.StatusOK && json.Unmarshal(response.Body.Bytes(), &view) == nil && view.UniquePeopleCount != nil && *view.UniquePeopleCount == count && view.UniquePeopleStatus == status {
			return view
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("unique people view не достиг count=%d status=%s: %#v", count, status, view)
	return stats.View{}
}

func assertP04aJob(t *testing.T, db *gorm.DB, userID uint, requested, applied, successful, retries uint64, leased bool) {
	t.Helper()
	var job statmodels.UniquePeopleJob
	if err := db.First(&job, "user_id = ?", userID).Error; err != nil {
		t.Fatalf("load unique people job %d: %v", userID, err)
	}
	if job.RequestedGeneration != requested || job.AppliedGeneration != applied || job.SuccessfulRuns != successful || job.RetryRuns != retries || (job.LeaseToken != nil) != leased {
		t.Fatalf("unique people job %d = %#v; want requested/applied/success/retry/leased=%d/%d/%d/%d/%t", userID, job, requested, applied, successful, retries, leased)
	}
}

func waitForP04aRetryRecovery(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var job statmodels.UniquePeopleJob
		if err := db.First(&job, "user_id = ?", userID).Error; err == nil && job.RetryRuns == 1 && job.LeaseToken == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expired lease пользователя %d не перешёл в persisted retry", userID)
}

func waitP04aSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(10 * time.Second):
		t.Fatalf("timeout: %s", description)
	}
}

func waitP04aClaim(t *testing.T, signal <-chan p04aClaimRecord) p04aClaimRecord {
	t.Helper()
	select {
	case claim := <-signal:
		return claim
	case <-time.After(10 * time.Second):
		t.Fatal("timeout: claim response loss")
		return p04aClaimRecord{}
	}
}

func buildP04aNotifyService(t *testing.T) string {
	t.Helper()
	notifyRoot := filepath.Clean(filepath.Join("..", "..", "notify_service"))
	name := "notify_service_p04a_blackbox"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable := filepath.Join(t.TempDir(), name)
	command := exec.Command("go", "build", "-buildvcs=false", "-o", executable, ".")
	command.Dir = notifyRoot
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build notify_service P0.4a black-box binary: %v\n%s", err, output)
	}
	return executable
}

func startP04aNotifyService(t *testing.T, executable, dsn, monolithURL, token string, reset, fastPoll bool) (string, func()) {
	t.Helper()
	if reset {
		resetNotifyServiceBlackBoxState(t, dsn)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve notify P0.4a port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	runContext, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(runContext, executable)
	command.Dir = filepath.Clean(filepath.Join("..", "..", "notify_service"))
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	var environment []string
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if strings.EqualFold(name, "NOTIFY_UNIQUE_PEOPLE_POLL_INTERVAL") || strings.EqualFold(name, "NOTIFY_UNIQUE_PEOPLE_CONCURRENCY") {
			continue
		}
		environment = append(environment, variable)
	}
	command.Env = append(environment,
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
		"NOTIFY_UNIQUE_PEOPLE_HTTP_TIMEOUT=2s",
	)
	if fastPoll {
		command.Env = append(command.Env, "NOTIFY_UNIQUE_PEOPLE_POLL_INTERVAL=20ms")
	}
	if err := command.Start(); err != nil {
		cancel()
		t.Fatalf("start notify_service P0.4a black-box binary: %v", err)
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
