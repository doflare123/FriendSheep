//go:build p04a_load

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/routes"
	events "friendship/services/events"
	stats "friendship/services/statistics"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Полная матрица выполняется только явно; отдельный case выбирается обычным -run.
func TestP04aUniquePeopleLoad(t *testing.T) {
	for _, n := range []int{500, 1000} {
		for _, history := range []int{100, 500} {
			for _, overlap := range []string{"full", "partial", "disjoint"} {
				t.Run(fmt.Sprintf("N%d/H%d/%s", n, history, overlap), func(t *testing.T) { runP04aLoad(t, n, history, overlap) })
			}
		}
	}
}

type p04aLoadSamples struct {
	mu             sync.Mutex
	profile, event []float64
	errors         int
}

func (s *p04aLoadSamples) snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	quantiles := func(values []float64) map[string]any {
		v := append([]float64{}, values...)
		sort.Float64s(v)
		if len(v) == 0 {
			return map[string]any{"requests": 0}
		}
		return map[string]any{"requests": len(v), "p95_ms": v[(len(v)-1)*95/100], "p99_ms": v[(len(v)-1)*99/100]}
	}
	return map[string]any{"profile": quantiles(s.profile), "event": quantiles(s.event), "errors": s.errors}
}

func runP04aLoad(t *testing.T, n, history int, overlap string) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, _ := statisticsFixture(t, db)
	repo := &testPostgresRepository{db: db}
	config := stats.DefaultUniquePeopleConfig()
	config.InitialSpread = time.Nanosecond
	ordinary := stats.NewService(stats.NewGORMStoreWithUniquePeopleConfig(repo, config))
	unique := stats.NewUniquePeopleService(stats.NewGORMUniquePeopleStore(repo), config, nil)
	cohorts := 6
	if overlap == "full" {
		cohorts = 1
	}
	people := cohorts * n
	if overlap == "partial" {
		people = n/2 + cohorts*n/2
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if err := db.Exec(query, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(id,name,password,us,email) SELECT i,'Нагрузочный участник','test','load-'||i,'load-'||i||'@example.test' FROM generate_series(100,?) i`, 99+people)
	completedID := seedLifecycleStatus(t, db, "Завершена")
	exec(`INSERT INTO events(id,title,event_type_id,event_location_id,group_id,start_time,end_time,duration,creator_id,current_users,max_users,status_id,age_limit_id)
		SELECT 1000+c*?+h,'История нагрузки',1,1,?,CURRENT_TIMESTAMP-interval '3 years'+h*interval '1 day',CURRENT_TIMESTAMP-interval '3 years'+h*interval '1 day'+interval '1 hour',60,1,?,?,?,1
		FROM generate_series(0,?) c CROSS JOIN generate_series(1,?) h`, history, groupID, n, n, completedID, cohorts-1, history)
	memberExpr := "100+m"
	if overlap == "partial" {
		memberExpr = fmt.Sprintf("CASE WHEN m < %d THEN 100+m ELSE 100+c*%d+m END", n/2, n/2)
	}
	if overlap == "disjoint" {
		memberExpr = fmt.Sprintf("100+c*%d+m", n)
	}
	exec(fmt.Sprintf(`INSERT INTO events_users(event_id,user_id,joined_at) SELECT 1000+c*?+h,%s,CURRENT_TIMESTAMP FROM generate_series(0,?) c CROSS JOIN generate_series(1,?) h CROSS JOIN generate_series(0,?) m`, memberExpr), history, cohorts-1, history, n-1)
	exec(`INSERT INTO participation_statistics_processed_events(event_id,processed_at) SELECT id,CURRENT_TIMESTAMP FROM events`)
	exec(`INSERT INTO participation_statistics_users(user_id,completed_events,planned_minutes,updated_at) SELECT user_id,COUNT(*),60*COUNT(*),CURRENT_TIMESTAMP FROM events_users GROUP BY user_id`)
	for event := 0; event < 6; event++ {
		c := event
		if overlap == "full" {
			c = 0
		}
		exec(`INSERT INTO events(id,title,event_type_id,event_location_id,group_id,start_time,end_time,duration,creator_id,current_users,max_users,status_id,age_limit_id)
			VALUES(?,'Новое мероприятие нагрузки',1,1,?,CURRENT_TIMESTAMP-interval '3 hour',CURRENT_TIMESTAMP-interval '2 hour',60,1,?,?,?,1)`, 10+event, groupID, n, n, completedID)
		exec(fmt.Sprintf(`INSERT INTO events_users(event_id,user_id,joined_at) SELECT ?,%s,CURRENT_TIMESTAMP FROM generate_series(0,?) m CROSS JOIN (SELECT ?::int c) cohort`, memberExpr), 10+event, n-1, c)
	}
	exec("ANALYZE events_users")
	exec("ANALYZE participation_statistics_processed_events")
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(16)
	sqlDB.SetMaxIdleConns(8)
	var fsync, synchronous string
	if err = db.Raw("SHOW fsync").Scan(&fsync).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Raw("SHOW synchronous_commit").Scan(&synchronous).Error; err != nil {
		t.Fatal(err)
	}
	if fsync != "on" || synchronous != "on" {
		t.Fatalf("нагрузочная проверка требует fsync=on и synchronous_commit=on: %s/%s", fsync, synchronous)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	identity := func(c *gin.Context) { c.Set("userID", uint(100)); c.Next() }
	routes.RegisterStatisticsRoutes(router, handlers.NewStatisticsHandler(ordinary, nil), identity)
	eventHandler := handlers.NewEventsHandler(handlers.EventsHandlerDependencies{Reads: events.NewEventReadService(&testLogger{}, events.NewGORMEventReadStore(repo))})
	router.GET("/api/v2/events/:eventId", identity, eventHandler.GetEventDetails)
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	measure := func(ctx context.Context, samples *p04aLoadSamples) *sync.WaitGroup {
		wg := &sync.WaitGroup{}
		for _, kind := range []string{"profile", "event"} {
			wg.Add(1)
			go func(kind string) {
				defer wg.Done()
				path := "/api/v2/users/me/statistics"
				if kind == "event" {
					path = "/api/v2/events/10"
				}
				for ctx.Err() == nil {
					started := time.Now()
					request, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+path, nil)
					response, e := client.Do(request)
					code := 0
					if response != nil {
						code = response.StatusCode
						_, _ = io.Copy(io.Discard, response.Body)
						_ = response.Body.Close()
					}
					elapsed := float64(time.Since(started).Microseconds()) / 1000
					if ctx.Err() != nil {
						return
					}
					samples.mu.Lock()
					if e != nil || code != 200 {
						samples.errors++
					}
					if kind == "profile" {
						samples.profile = append(samples.profile, elapsed)
					} else {
						samples.event = append(samples.event, elapsed)
					}
					samples.mu.Unlock()
					timer := time.NewTimer(10 * time.Millisecond)
					select {
					case <-ctx.Done():
						timer.Stop()
						return
					case <-timer.C:
					}
				}
			}(kind)
		}
		return wg
	}
	baseline := &p04aLoadSamples{}
	baselineCtx, baselineCancel := context.WithTimeout(context.Background(), time.Second)
	measure(baselineCtx, baseline).Wait()
	baselineCancel()
	loaded := &p04aLoadSamples{}
	loadedCtx, loadedCancel := context.WithCancel(context.Background())
	loadedWG := measure(loadedCtx, loaded)
	defer func() { loadedCancel(); loadedWG.Wait() }()
	before := p04aLoadResource(db)
	startProcess := time.Now()
	var processWG sync.WaitGroup
	errs := make(chan error, 6)
	for id := uint(10); id < 16; id++ {
		processWG.Add(1)
		go func(id uint) {
			defer processWG.Done()
			_, err := ordinary.Process(context.Background(), id)
			errs <- err
		}(id)
	}
	processWG.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	processDuration := time.Since(startProcess)
	startDrain := time.Now()
	done := 0
	for {
		claims, err := unique.Claim(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(claims.Jobs) == 0 {
			metrics, err := unique.Metrics(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if metrics.PendingUsers == 0 {
				break
			}
			t.Fatalf("catch-up остановился с pending: %+v", metrics)
		}
		for _, claim := range claims.Jobs {
			result, err := unique.Execute(context.Background(), claim)
			if err != nil || result.Outcome != stats.UniquePeopleOutcomeApplied {
				t.Fatalf("execute=%+v/%v", result, err)
			}
			done++
		}
	}
	drainDuration := time.Since(startDrain)
	after := p04aLoadResource(db)
	loadedCancel()
	loadedWG.Wait()
	if baseline.errors+loaded.errors != 0 {
		t.Fatalf("ошибки HTTP baseline=%d loaded=%d", baseline.errors, loaded.errors)
	}
	if done != people {
		t.Fatalf("calculations=%d want=%d", done, people)
	}
	var wrong int64
	expect := fmt.Sprintf("unique_people_count <> %d", n-1)
	if overlap == "partial" {
		expect = fmt.Sprintf("unique_people_count <> CASE WHEN user_id < %d THEN %d ELSE %d END", 100+n/2, people-1, n-1)
	}
	if err = db.Raw("SELECT COUNT(*) FROM participation_statistics_users WHERE user_id >=100 AND (unique_people_count IS NULL OR " + expect + ")").Scan(&wrong).Error; err != nil || wrong != 0 {
		t.Fatalf("incorrect counts=%d/%v", wrong, err)
	}
	var single, batch []string
	query := `SELECT COUNT(DISTINCT other.user_id) FROM events_users mine JOIN participation_statistics_processed_events processed ON processed.event_id=mine.event_id JOIN events_users other ON other.event_id=mine.event_id WHERE mine.user_id=100 AND other.user_id<>100`
	if err = db.Raw("EXPLAIN (ANALYZE, BUFFERS) " + query).Scan(&single).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Raw(`EXPLAIN (ANALYZE, BUFFERS) SELECT mine.user_id,COUNT(DISTINCT other.user_id) FROM events_users mine JOIN participation_statistics_processed_events processed ON processed.event_id=mine.event_id JOIN events_users other ON other.event_id=mine.event_id WHERE mine.user_id BETWEEN 100 AND 107 AND other.user_id<>mine.user_id GROUP BY mine.user_id`).Scan(&batch).Error; err != nil {
		t.Fatal(err)
	}
	var claimPlan []string
	if err = db.Raw(`EXPLAIN (ANALYZE, BUFFERS) SELECT user_id FROM participation_statistics_unique_people_jobs WHERE requested_generation>applied_generation AND next_attempt_at<=CURRENT_TIMESTAMP AND lease_token IS NULL ORDER BY dirty_since,next_attempt_at,user_id LIMIT 1`).Scan(&claimPlan).Error; err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"case": t.Name(), "go": runtime.Version(), "cpus": runtime.NumCPU(), "os": runtime.GOOS, "fsync": fsync, "synchronous_commit": synchronous, "participants_per_event": n, "history_per_cohort": history, "cohorts": cohorts, "users": people, "new_events": 6, "process_6_events_ms": processDuration.Milliseconds(), "drain_ms": drainDuration.Milliseconds(), "calculations": done, "calculations_per_second": float64(done) / drainDuration.Seconds(), "baseline_http": baseline.snapshot(), "loaded_http": loaded.snapshot(), "db_before": before, "db_after": after, "pool": sqlDB.Stats(), "single_plan": single, "batch8_plan": batch, "claim_plan": claimPlan, "limitations": "История загружена SQL с processed markers; шесть новых мероприятий проходят реальный P0.4 service. HTTP использует реальные Gin handlers/services/БД, но подменяет проверку JWT. Catch-up насыщенный без 5s polling, concurrency=1; не SLA."}
	body, _ := json.MarshalIndent(result, "", "  ")
	t.Log(string(body))
	if directory := os.Getenv("P04A_LOAD_OUTPUT_DIR"); directory != "" {
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		name := fmt.Sprintf("n%d_h%d_%s.json", n, history, overlap)
		if err = os.WriteFile(filepath.Join(directory, name), body, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func p04aLoadResource(db *gorm.DB) map[string]any {
	result := map[string]any{}
	var rows []map[string]any
	if err := db.Raw(`SELECT pg_total_relation_size('events_users') AS membership_bytes,pg_total_relation_size('participation_statistics_unique_people_jobs') AS queue_bytes,pg_total_relation_size('participation_statistics_users') AS aggregate_bytes,(SELECT COUNT(*) FROM pg_stat_activity WHERE datname=current_database()) AS connections,(SELECT COUNT(*) FROM pg_locks WHERE NOT granted) AS waiting_locks`).Scan(&rows).Error; err == nil && len(rows) > 0 {
		result = rows[0]
	}
	var resources []map[string]any
	if err := db.Raw(`SELECT temp_files,temp_bytes,blks_read,blks_hit FROM pg_stat_database WHERE datname=current_database()`).Scan(&resources).Error; err == nil && len(resources) > 0 {
		for key, value := range resources[0] {
			result[key] = value
		}
	}
	result["scope"] = "Снимок общей test database; при параллельных suites ресурсы могут включать их работу."
	return result
}

var _ = strings.Builder{}
