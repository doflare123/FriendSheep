package tests

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"friendship/models"
	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
	statmodels "friendship/models/statistics"
	servicesevents "friendship/services/events"
	servicegroups "friendship/services/groups"
	stats "friendship/services/statistics"

	"gorm.io/gorm"
)

func statisticsPostgresDSN(t *testing.T) string {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("FRIENDSHEEP_TEST_POSTGRES_DSN обязателен: SQLite не подтверждает PostgreSQL locking и параллельные агрегаты")
	}
	return dsn
}

func TestParticipationStatisticsConcurrentOverlappingJobsPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, service := statisticsFixture(t, db)
	genre := seedEventGenre(t, db, "Параллельные встречи")
	start := time.Now().Add(-3 * time.Hour)
	events := []uint{
		seedStatisticsEvent(t, db, groupID, 1, 1, start, 35, []uint{3, 2}, []uint{genre}),
		seedStatisticsEvent(t, db, groupID, 1, 1, start, 45, []uint{4, 3, 2}, []uint{genre}),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	startWorkers := make(chan struct{})
	errs := make(chan error, 12)
	var wg sync.WaitGroup
	for range 6 {
		for _, id := range events {
			wg.Add(1)
			go func(id uint) {
				defer wg.Done()
				<-startWorkers
				_, err := service.Process(ctx, id)
				errs <- err
			}(id)
		}
	}
	close(startWorkers)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Process = %v", err)
		}
	}
	for _, id := range []uint{2, 3} {
		view, err := service.Read(ctx, id)
		if err != nil || view.CompletedEvents != 2 || view.PlannedMinutes != 80 || view.UniquePeopleCount != 2 || len(view.Genres) != 1 || view.Genres[0].Count != 2 || len(view.Types) != 1 || view.Types[0].Count != 2 {
			t.Fatalf("overlapping user %d view/error=%#v/%v", id, view, err)
		}
	}
	view, err := service.Read(ctx, 4)
	if err != nil || view.CompletedEvents != 1 || view.PlannedMinutes != 45 || view.UniquePeopleCount != 2 {
		t.Fatalf("user4 view/error=%#v/%v", view, err)
	}
	creator, err := service.Read(ctx, 1)
	if err != nil || creator.CompletedEvents != 0 || creator.CreatedCompletedEvents != 2 || creator.UniquePeopleCount != 0 {
		t.Fatalf("creator view/error=%#v/%v", creator, err)
	}
	var pairs, processed int64
	if err := db.Model(&statmodels.UserPair{}).Count(&pairs).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&statmodels.ProcessedEvent{}).Count(&processed).Error; err != nil {
		t.Fatal(err)
	}
	if pairs != 6 || processed != 2 {
		t.Fatalf("pairs/processed=%d/%d", pairs, processed)
	}
}

func TestParticipationStatisticsRollbackAndRetryPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, service := statisticsFixture(t, db)
	id := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-2*time.Hour), 60, []uint{2, 3}, nil)
	if err := db.Exec(`CREATE FUNCTION fail_statistics_distribution() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected statistics rollback'; END; $$`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER fail_statistics_distribution BEFORE INSERT ON participation_statistics_distributions FOR EACH ROW EXECUTE FUNCTION fail_statistics_distribution()`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.Process(context.Background(), id); err == nil {
		t.Fatal("неуспешная транзакция принята")
	}
	for _, model := range []any{&statmodels.UserAggregate{}, &statmodels.UserPair{}, &statmodels.ProcessedEvent{}, &statmodels.Distribution{}} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("rollback %T count/error=%d/%v", model, count, err)
		}
	}
	if err := db.Exec(`DROP TRIGGER fail_statistics_distribution ON participation_statistics_distributions`).Error; err != nil {
		t.Fatal(err)
	}
	if result, err := service.Process(context.Background(), id); err != nil || result.Outcome != "processed" {
		t.Fatalf("retry result/error=%#v/%v", result, err)
	}
}

func TestStatisticsTriggerPublicationCommitOrderPostgres(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		name := "commit"
		if rollback {
			name = "rollback_gap"
		}
		t.Run(name, func(t *testing.T) {
			db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
			groupID, _ := statisticsFixture(t, db)
			ids := []uint{
				seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-time.Hour), 10, nil, nil),
				seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-time.Hour), 10, nil, nil),
			}
			repo := &testPostgresRepository{db: db}
			store := servicesevents.NewGORMEventLifecycleStore(repo)
			reader := servicesevents.NewStatisticsTriggerService(servicesevents.NewGORMStatisticsTriggerReader(repo), nil)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ready, release := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			defer releaseOnce.Do(func() { close(release) })
			doneA, doneB := make(chan error, 1), make(chan error, 1)
			rollbackErr := errors.New("controlled statistics rollback")
			go func() {
				doneA <- store.WithinLifecycleTransaction(ctx, ids[0], func(tx servicesevents.EventLifecycleTransaction) error {
					if _, err := tx.LoadEventForUpdate(ctx, ids[0]); err != nil {
						return err
					}
					if err := tx.PublishStatisticsTrigger(ctx, ids[0], time.Now()); err != nil {
						return err
					}
					close(ready)
					select {
					case <-release:
					case <-ctx.Done():
						return ctx.Err()
					}
					if rollback {
						return rollbackErr
					}
					return nil
				})
			}()
			select {
			case <-ready:
			case err := <-doneA:
				t.Fatalf("A before barrier=%v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			go func() {
				doneB <- store.WithinLifecycleTransaction(ctx, ids[1], func(tx servicesevents.EventLifecycleTransaction) error {
					if _, err := tx.LoadEventForUpdate(ctx, ids[1]); err != nil {
						return err
					}
					return tx.PublishStatisticsTrigger(ctx, ids[1], time.Now())
				})
			}()
			for {
				var waiting bool
				if err := db.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM pg_locks WHERE locktype='advisory' AND classid=1179862864 AND objid=1 AND NOT granted)").Scan(&waiting).Error; err != nil {
					t.Fatal(err)
				}
				if waiting {
					break
				}
				select {
				case err := <-doneB:
					t.Fatalf("B overtook A=%v", err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				default:
					runtime.Gosched()
				}
			}
			page, err := reader.List(ctx, 0, 10)
			if err != nil || len(page.Messages) != 0 || page.NextCursor != 0 {
				t.Fatalf("uncommitted page/error=%#v/%v", page, err)
			}
			// Waiter не должен успеть взять event row до общего publication lock.
			if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				return tx.Exec("SELECT id FROM events WHERE id = ? FOR UPDATE NOWAIT", ids[1]).Error
			}); err != nil {
				t.Fatalf("lifecycle took event row before publication lock: %v", err)
			}
			releaseOnce.Do(func() { close(release) })
			if err := <-doneA; (!rollback && err != nil) || (rollback && !errors.Is(err, rollbackErr)) {
				t.Fatalf("A=%v", err)
			}
			if err := <-doneB; err != nil {
				t.Fatalf("B=%v", err)
			}
			page, err = reader.List(ctx, 0, 10)
			want := ids
			if rollback {
				want = ids[1:]
			}
			if err != nil || len(page.Messages) != len(want) {
				t.Fatalf("page/error=%#v/%v", page, err)
			}
			for i, id := range want {
				if page.Messages[i].EventID != id {
					t.Fatalf("out-of-order publication=%#v", page)
				}
			}
			if rollback && page.Messages[0].Sequence <= 1 {
				t.Fatalf("rollback gap absent: %#v", page)
			}
			var cache int64
			if err := db.Raw("SELECT cache_size FROM pg_sequences WHERE schemaname=current_schema() AND sequencename='event_statistics_triggers_sequence_seq'").Scan(&cache).Error; err != nil || cache != 1 {
				t.Fatalf("sequence CACHE/error=%d/%v", cache, err)
			}
		})
	}
}

func TestParticipationStatisticsMaximumEventPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, service := statisticsFixture(t, db)
	const count = 1000
	users := make([]models.User, 0, count-4)
	participants := make([]uint, count)
	for i := 1; i <= count; i++ {
		participants[i-1] = uint(i)
		if i > 4 {
			id := uint(i)
			users = append(users, models.User{ID: id, Name: "Участник", Password: "Password123!", Us: "statistics-user-" + testUintString(id), Email: "statistics-user-" + testUintString(id) + "@example.com"})
		}
	}
	if err := db.CreateInBatches(users, 100).Error; err != nil {
		t.Fatal(err)
	}
	id := seedEvent(t, db, groupID, 1, count, count)
	statusID := seedLifecycleStatus(t, db, eventmodels.StatusCompleted)
	start := time.Now().Add(-2 * time.Hour)
	if err := db.Model(&eventmodels.Event{}).Where("id = ?", id).Updates(map[string]any{"status_id": statusID, "start_time": start, "end_time": start.Add(time.Hour), "duration": 60}).Error; err != nil {
		t.Fatal(err)
	}
	links := make([]eventmodels.EventsUser, 0, count)
	for _, userID := range participants {
		links = append(links, eventmodels.EventsUser{EventID: id, UserID: userID, JoinedAt: start})
	}
	if err := db.CreateInBatches(links, 100).Error; err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	result, err := service.Process(context.Background(), id)
	t.Logf("1000 participants, 999000 directed pairs: processing elapsed %s", time.Since(started))
	if err != nil || result.Outcome != "processed" {
		t.Fatalf("max process/error=%#v/%v", result, err)
	}
	var pairs int64
	if err := db.Model(&statmodels.UserPair{}).Count(&pairs).Error; err != nil || pairs != count*(count-1) {
		t.Fatalf("max pairs/error=%d/%v", pairs, err)
	}
	for _, userID := range []uint{1, 500, 1000} {
		view, err := service.Read(context.Background(), userID)
		if err != nil || view.UniquePeopleCount != count-1 || view.CompletedEvents != 1 || view.PlannedMinutes != 60 {
			t.Fatalf("max user %d view/error=%#v/%v", userID, view, err)
		}
	}
	started = time.Now()
	if result, err := service.Process(context.Background(), id); err != nil || result.Outcome != "already_processed" {
		t.Fatalf("max retry/error=%#v/%v", result, err)
	}
	t.Logf("1000-participant idempotent retry elapsed %s", time.Since(started))
}

func TestParticipationStatisticsHistorySurvivesGroupOperationsPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, service := statisticsFixture(t, db)
	adminRole := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	memberRole := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceMembership(t, db, groupID, 1, adminRole)
	seedGroupServiceMembership(t, db, groupID, 2, memberRole)
	seedGroupServiceMembership(t, db, groupID, 3, memberRole)
	id := seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-2*time.Hour), 60, []uint{2, 3}, nil)
	if _, err := service.Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	groups := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(&testPostgresRepository{db: db}))
	if left, err := groups.LeaveGroup(context.Background(), 2, groupID); err != nil || !left {
		t.Fatalf("LeaveGroup=%t/%v", left, err)
	}
	if banned, err := groups.DeleteUserFromGroup(context.Background(), 1, groupID, 3); err != nil || !banned {
		t.Fatalf("DeleteUserFromGroup=%t/%v", banned, err)
	}
	// Текущий FK запрещает удаление группы с мероприятиями; публичную семантику здесь не меняем.
	if deleted, err := groups.DeleteGroup(context.Background(), 1, groupID); err == nil || deleted {
		t.Fatalf("DeleteGroup with completed history=%t/%v", deleted, err)
	}
	var group groupmodels.Group
	if err := db.First(&group, groupID).Error; err != nil {
		t.Fatalf("group lost after rejected deletion: %v", err)
	}
	var event eventmodels.Event
	if err := db.First(&event, id).Error; err != nil || event.CurrentUsers != 2 {
		t.Fatalf("historical event=%#v/%v", event, err)
	}
	for _, userID := range []uint{2, 3} {
		assertEventParticipantExists(t, db, id, userID, true)
		view, err := service.Read(context.Background(), userID)
		if err != nil || view.CompletedEvents != 1 || view.PlannedMinutes != 60 || view.UniquePeopleCount != 1 {
			t.Fatalf("retained user %d view/error=%#v/%v", userID, view, err)
		}
	}
	var processed, pairs int64
	if err := db.Model(&statmodels.ProcessedEvent{}).Count(&processed).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&statmodels.UserPair{}).Count(&pairs).Error; err != nil {
		t.Fatal(err)
	}
	if processed != 1 || pairs != 2 {
		t.Fatalf("retained processed/pairs=%d/%d", processed, pairs)
	}
}

func TestParticipationStatisticsLastUpdatedDoesNotMoveBackwardsPostgres(t *testing.T) {
	db := newPostgresEventCommandDB(t, statisticsPostgresDSN(t))
	groupID, service := statisticsFixture(t, db)
	ids := []uint{
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-2*time.Hour), 15, []uint{2, 3}, nil),
		seedStatisticsEvent(t, db, groupID, 1, 1, time.Now().Add(-2*time.Hour), 20, []uint{3, 2}, nil),
	}
	if _, err := service.Process(context.Background(), ids[0]); err != nil {
		t.Fatal(err)
	}
	latest := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	if err := db.Model(&statmodels.UserAggregate{}).Where("user_id IN ?", []uint{1, 2, 3}).Update("updated_at", latest).Error; err != nil {
		t.Fatal(err)
	}
	// Время команды фиксируется до ожидания user locks и может отставать от другого commit.
	store := stats.NewGORMStore(&testPostgresRepository{db: db})
	err := store.WithinTransaction(context.Background(), func(tx stats.Transaction) error {
		snapshot, err := tx.LoadEvent(context.Background(), ids[1])
		if err != nil {
			return err
		}
		_, err = tx.Apply(context.Background(), stats.Contribution{Snapshot: snapshot, Month: snapshot.StartTime.UTC().Format("2006-01"), Weekday: 1}, latest.Add(-time.Minute))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []uint{1, 2, 3} {
		view, err := service.Read(context.Background(), id)
		if err != nil || view.LastUpdatedAt == nil || !view.LastUpdatedAt.Equal(latest) {
			t.Fatalf("monotonic user%d view/error=%#v/%v", id, view, err)
		}
	}
}
