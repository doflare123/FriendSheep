package tests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"friendship/models"
	groupmodels "friendship/models/groups"
	servicegroups "friendship/services/groups"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestGroupServiceJoinGroupConcurrentPrivateRequestsPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
	}

	db := newPostgresGroupServiceDB(t, dsn)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	createPendingJoinRequestUniqueIndexPostgres(t, db)

	var successCount int32
	var duplicateCount int32
	var unexpectedErr error
	var errMu sync.Mutex

	start := make(chan struct{})
	var wg sync.WaitGroup
	const workers = 2
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start

			result, err := service.JoinGroup(context.Background(), 2, groupID)
			switch {
			case err == nil:
				if result == nil || result.Joined {
					errMu.Lock()
					if unexpectedErr == nil {
						unexpectedErr = fmt.Errorf("unexpected join result: %#v", result)
					}
					errMu.Unlock()
					return
				}
				atomic.AddInt32(&successCount, 1)
			case errors.Is(err, servicegroups.ErrRequestAlreadyExists):
				atomic.AddInt32(&duplicateCount, 1)
			default:
				errMu.Lock()
				if unexpectedErr == nil {
					unexpectedErr = err
				}
				errMu.Unlock()
			}
		}()
	}

	close(start)
	wg.Wait()

	if unexpectedErr != nil {
		t.Fatalf("unexpected concurrent error: %v", unexpectedErr)
	}
	if successCount != 1 {
		t.Fatalf("success count = %d, want 1", successCount)
	}
	if duplicateCount != 1 {
		t.Fatalf("duplicate count = %d, want 1", duplicateCount)
	}

	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 1)
	assertGroupMembershipExists(t, db, groupID, 2, false)
}

func newPostgresGroupServiceDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	adminDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open postgres admin connection: %v", err)
	}
	adminSQL, err := adminDB.DB()
	if err != nil {
		t.Fatalf("get postgres admin sql db: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := adminSQL.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres admin connection: %v", err)
	}

	schemaName := fmt.Sprintf("it_group_%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	if err := adminDB.Exec(`CREATE SCHEMA "` + schemaName + `"`).Error; err != nil {
		t.Fatalf("create postgres test schema: %v", err)
	}

	scopedDSN := withSearchPath(dsn, schemaName)
	db, err := gorm.Open(postgres.Open(scopedDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open postgres schema connection: %v", err)
	}

	t.Cleanup(func() {
		_ = adminDB.Exec(`DROP SCHEMA IF EXISTS "` + schemaName + `" CASCADE`).Error
	})

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&groupmodels.Group{},
		&groupmodels.GroupGroupCategory{},
		&groupmodels.GroupContact{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&groupmodels.GroupJoinRequest{},
		&groupmodels.GroupJoinInvite{},
		&groupmodels.GroupBlacklist{},
		&groupmodels.GroupActionType{},
		&groupmodels.GroupActionLog{},
	); err != nil {
		t.Fatalf("auto migrate group models in postgres test schema: %v", err)
	}

	seedGroupActionTypes(t, db)

	return db
}

func withSearchPath(dsn, schema string) string {
	if strings.Contains(dsn, "://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return dsn
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		if query.Get("connect_timeout") == "" {
			query.Set("connect_timeout", "3")
		}
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}

	trimmed := strings.TrimSpace(dsn)
	if !strings.Contains(trimmed, "search_path=") {
		trimmed += " search_path=" + schema
	}
	if !strings.Contains(trimmed, "connect_timeout=") {
		trimmed += " connect_timeout=3"
	}
	return trimmed
}

func createPendingJoinRequestUniqueIndexPostgres(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec(
		"CREATE UNIQUE INDEX idx_group_join_request_pending_unique ON group_join_requests (user_id, group_id) WHERE status = 'pending'",
	).Error; err != nil {
		t.Fatalf("create postgres pending join request unique index: %v", err)
	}
}
