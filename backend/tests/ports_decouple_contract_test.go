package tests

import (
	"database/sql"
	"reflect"
	"testing"

	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
	servicegroups "friendship/services/groups"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// groupServiceLocalPort mirrors the local groups storage contract expected by service constructors.
type groupServiceLocalPort interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Find(out interface{}, where ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Updates(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Clauses(conds ...clause.Expression) *gorm.DB
	Order(value interface{}) *gorm.DB
	Limit(limit int) *gorm.DB
	Count(count *int64) *gorm.DB
	ScanRows(rows *sql.Rows, result interface{}) error
}

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

var _ groupServiceLocalPort = (*testPostgresRepository)(nil)
var _ eventsServiceLocalPort = (*testPostgresRepository)(nil)

func TestNewGroupServiceAcceptsLocalPortAndPreservesBehavior(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	var localPort groupServiceLocalPort = repo

	assertConstructorAcceptsRepo(t, servicegroups.NewGroupService, repo)
	service := invokeGroupServiceConstructor(t, &testLogger{}, localPort)

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

	ctorType := reflect.TypeOf(constructor)
	if ctorType.Kind() != reflect.Func {
		t.Fatalf("constructor kind = %v, want func", ctorType.Kind())
	}
	if ctorType.NumIn() != 2 {
		t.Fatalf("constructor args = %d, want 2", ctorType.NumIn())
	}

	repoType := reflect.TypeOf(repo)
	ctorRepoArg := ctorType.In(1)
	accepts := repoType.AssignableTo(ctorRepoArg)
	if !accepts && ctorRepoArg.Kind() == reflect.Interface {
		accepts = repoType.Implements(ctorRepoArg)
	}
	if !accepts {
		t.Fatalf("constructor repo arg type %v does not accept %v", ctorRepoArg, repoType)
	}
}

func invokeGroupServiceConstructor(t *testing.T, l *testLogger, repo groupServiceLocalPort) servicegroups.GroupsService {
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
