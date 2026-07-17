package tests

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
	servicegroups "friendship/services/groups"

	"gorm.io/gorm"
)

// eventsServiceLocalPort mirrors the local events storage contract expected by service constructors.
type eventsServiceLocalPort interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Find(out interface{}, where ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Count(count *int64) *gorm.DB
}

var _ eventsServiceLocalPort = (*testPostgresRepository)(nil)

func TestNewGroupServiceUsesGORMAdapterAndPreservesBehavior(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}

	if constructorAcceptsRepo(servicegroups.NewGroupService, repo) {
		t.Fatal("NewGroupService accepts raw GORM-shaped repository; want only group adapter")
	}

	adapter := servicegroups.NewGORMGroupRepository(repo)
	assertConstructorAcceptsRepo(t, servicegroups.NewGroupService, adapter)
	service := invokeGroupServiceConstructor(t, &testLogger{}, adapter)

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	result, err := service.JoinGroup(2, groupID)
	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
}

func TestGroupStoresDoNotImportSharedPostgresRepository(t *testing.T) {
	groupsDir := filepath.Join("..", "services", "groups")
	allowedFile := filepath.Join(groupsDir, "gorm_repository.go")

	err := filepath.WalkDir(groupsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || path == allowedFile {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(content)
		if strings.Contains(text, `"friendship/repository"`) || strings.Contains(text, "repository.PostgresRepository") {
			t.Fatalf("%s imports or references shared Postgres repository; keep it isolated in %s", path, allowedFile)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk group services: %v", err)
	}
}

func TestRegistrationServiceDoesNotImportStorageLibraries(t *testing.T) {
	registerDir := filepath.Join("..", "services", "register")
	allowedAdapter := filepath.Join(registerDir, "gorm_registration_store.go")

	err := filepath.WalkDir(registerDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || path == allowedAdapter {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		for _, forbidden := range []string{
			`"friendship/repository"`,
			`"gorm.io/`,
			`"github.com/jackc/pgx/`,
			"repository.PostgresRepository",
			"gorm.DB",
			"pgconn.PgError",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s references storage implementation %q; keep it isolated in %s", path, forbidden, allowedAdapter)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk registration services: %v", err)
	}
}

func TestNewEventsServiceAcceptsLocalPortAndPreservesBehavior(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	var localPort eventsServiceLocalPort = repo

	assertConstructorAcceptsRepo(t, servicesevents.NewEventsService, repo)
	service := invokeEventsServiceConstructor(t, &testLogger{}, localPort)

	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	groupID := seedEventGroup(t, db, 1, false)
	seedEventGroupMembership(t, db, 2, groupID)
	eventID := seedEvent(t, db, groupID, 1, 1, 2)

	joined, err := service.JoinEvent(2, eventID)
	if err != nil {
		t.Fatalf("JoinEvent returned error: %v", err)
	}
	if !joined {
		t.Fatal("JoinEvent returned false")
	}
	assertEventParticipantExists(t, db, eventID, 2, true)
	assertEventCurrentUsers(t, db, eventID, 2)
}

func assertConstructorAcceptsRepo(t *testing.T, constructor interface{}, repo interface{}) {
	t.Helper()

	if !constructorAcceptsRepo(constructor, repo) {
		t.Fatalf("constructor repo arg does not accept %T", repo)
	}
}

func constructorAcceptsRepo(constructor interface{}, repo interface{}) bool {
	ctorType := reflect.TypeOf(constructor)
	if ctorType.Kind() != reflect.Func {
		return false
	}
	if ctorType.NumIn() != 2 {
		return false
	}

	repoType := reflect.TypeOf(repo)
	ctorRepoArg := ctorType.In(1)
	accepts := repoType.AssignableTo(ctorRepoArg)
	if !accepts && ctorRepoArg.Kind() == reflect.Interface {
		accepts = repoType.Implements(ctorRepoArg)
	}
	return accepts
}

func invokeGroupServiceConstructor(t *testing.T, l *testLogger, repo interface{}) servicegroups.GroupsService {
	t.Helper()

	out := reflect.ValueOf(servicegroups.NewGroupService).Call([]reflect.Value{
		reflect.ValueOf(l),
		reflect.ValueOf(repo),
	})
	svc, ok := out[0].Interface().(servicegroups.GroupsService)
	if !ok || svc == nil {
		t.Fatalf("NewGroupService returned %T, want GroupsService", out[0].Interface())
	}
	return svc
}

func invokeEventsServiceConstructor(t *testing.T, l *testLogger, repo eventsServiceLocalPort) servicesevents.EventsService {
	t.Helper()

	out := reflect.ValueOf(servicesevents.NewEventsService).Call([]reflect.Value{
		reflect.ValueOf(l),
		reflect.ValueOf(repo),
	})
	svc, ok := out[0].Interface().(servicesevents.EventsService)
	if !ok || svc == nil {
		t.Fatalf("NewEventsService returned %T, want EventsService", out[0].Interface())
	}
	return svc
}
