package tests

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	applicationlogger "friendship/logger"
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

type eventReadStoreContractStub struct{}

func (eventReadStoreContractStub) SearchEvents(context.Context, servicesevents.EventSearchQuery) (servicesevents.EventSearchPageView, error) {
	return servicesevents.EventSearchPageView{}, nil
}

func (eventReadStoreContractStub) GetGroupEvents(context.Context, servicesevents.EventGroupEventsQuery) (servicesevents.EventGroupEventsView, error) {
	return servicesevents.EventGroupEventsView{}, nil
}

func (eventReadStoreContractStub) GetEventDetails(context.Context, servicesevents.EventDetailsQuery) (servicesevents.EventDetailsView, error) {
	return servicesevents.EventDetailsView{}, nil
}

var _ servicesevents.EventReadStore = eventReadStoreContractStub{}

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

func TestEventApplicationServicesDoNotImportStorageLibraries(t *testing.T) {
	eventsDir := filepath.Join("..", "services", "events")
	applicationFiles := []string{
		filepath.Join(eventsDir, "command_service.go"),
		filepath.Join(eventsDir, "membership_uow.go"),
		filepath.Join(eventsDir, "membership_service.go"),
	}

	for _, path := range applicationFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read event membership application file %s: %v", path, err)
		}

		text := string(content)
		for _, forbidden := range []string{
			`"friendship/repository"`,
			`"friendship/models/events"`,
			`"gorm.io/`,
			`"github.com/jackc/pgx/`,
			"repository.PostgresRepository",
			"eventmodels.",
			"gorm.DB",
			"pgconn.PgError",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s references storage implementation %q; keep it isolated in GORM adapter files", path, forbidden)
			}
		}
	}
}

func TestNewEventCommandServiceAcceptsOnlyLoggerAndUnitOfWork(t *testing.T) {
	constructorType := reflect.TypeOf(servicesevents.NewEventCommandService)
	if constructorType.NumIn() != 2 {
		t.Fatalf("NewEventCommandService has %d arguments, want 2", constructorType.NumIn())
	}

	loggerType := reflect.TypeOf((*applicationlogger.Logger)(nil)).Elem()
	if constructorType.In(0) != loggerType {
		t.Fatalf("NewEventCommandService first argument = %v, want %v", constructorType.In(0), loggerType)
	}
	uowType := reflect.TypeOf((*servicesevents.EventUnitOfWork)(nil)).Elem()
	if constructorType.In(1) != uowType {
		t.Fatalf("NewEventCommandService second argument = %v, want %v", constructorType.In(1), uowType)
	}

	uow := &eventUnitOfWorkStub{}
	out := reflect.ValueOf(servicesevents.NewEventCommandService).Call([]reflect.Value{
		reflect.ValueOf(&testLogger{}),
		reflect.ValueOf(uow),
	})
	service, ok := out[0].Interface().(servicesevents.EventCommandService)
	if !ok || service == nil {
		t.Fatalf("NewEventCommandService returned %T, want EventCommandService", out[0].Interface())
	}

	repo := &testPostgresRepository{}
	if constructorAcceptsRepo(servicesevents.NewEventCommandService, repo) {
		t.Fatal("NewEventCommandService accepts broad GORM-shaped repository; want EventUnitOfWork")
	}
}

func TestNewEventsServiceAcceptsLocalPort(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}
	var localPort eventsServiceLocalPort = repo

	assertConstructorAcceptsRepo(t, servicesevents.NewEventsService, repo)
	service := invokeEventsServiceConstructor(t, &testLogger{}, localPort)
	if service == nil {
		t.Fatal("NewEventsService returned nil")
	}
}

func TestNewEventReadServiceAcceptsOnlyCleanReadStore(t *testing.T) {
	constructorType := reflect.TypeOf(servicesevents.NewEventReadService)
	if constructorType.NumIn() != 2 {
		t.Fatalf("NewEventReadService has %d arguments, want 2", constructorType.NumIn())
	}

	loggerType := reflect.TypeOf((*applicationlogger.Logger)(nil)).Elem()
	if constructorType.In(0) != loggerType {
		t.Fatalf("NewEventReadService first argument = %v, want %v", constructorType.In(0), loggerType)
	}
	storeType := reflect.TypeOf((*servicesevents.EventReadStore)(nil)).Elem()
	if constructorType.In(1) != storeType {
		t.Fatalf("NewEventReadService second argument = %v, want %v", constructorType.In(1), storeType)
	}
	serviceType := reflect.TypeOf((*servicesevents.EventReadService)(nil)).Elem()
	if constructorType.NumOut() != 1 || constructorType.Out(0) != serviceType {
		t.Fatalf("NewEventReadService result = %v, want %v", constructorType.Out(0), serviceType)
	}

	if constructorAcceptsRepo(servicesevents.NewEventReadService, &testPostgresRepository{}) {
		t.Fatal("NewEventReadService accepts broad GORM-shaped repository; want EventReadStore")
	}

	service := servicesevents.NewEventReadService(&testLogger{}, eventReadStoreContractStub{})
	if service == nil {
		t.Fatal("NewEventReadService returned nil")
	}
}

func TestNewGORMEventReadStoreAdaptsSharedRepositoryAtBoundary(t *testing.T) {
	db := newEventsServiceDB(t)
	repo := &testPostgresRepository{db: db}

	var store servicesevents.EventReadStore = servicesevents.NewGORMEventReadStore(repo)
	if store == nil {
		t.Fatal("NewGORMEventReadStore returned nil")
	}
}

func TestEventReadContractsDoNotLeakStorageTypes(t *testing.T) {
	checked := make(map[reflect.Type]bool)
	for _, contract := range []reflect.Type{
		reflect.TypeOf((*servicesevents.EventReadStore)(nil)).Elem(),
		reflect.TypeOf((*servicesevents.EventReadService)(nil)).Elem(),
	} {
		for i := 0; i < contract.NumMethod(); i++ {
			method := contract.Method(i)
			for argument := 0; argument < method.Type.NumIn(); argument++ {
				assertEventReadTypeDoesNotLeakStorage(t, method.Type.In(argument), checked)
			}
			for result := 0; result < method.Type.NumOut(); result++ {
				assertEventReadTypeDoesNotLeakStorage(t, method.Type.Out(result), checked)
			}
		}
	}
}

func TestEventReadCleanSourceDoesNotImportStorageLibraries(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "services", "events", "read_service.go"),
		filepath.Join("..", "services", "events", "read_types.go"),
	} {
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read event read source %s: %v", sourcePath, err)
		}

		text := string(content)
		for _, forbidden := range []string{
			`"gorm.io/gorm"`,
			`"friendship/repository"`,
			`"friendship/models/events"`,
			`"friendship/models/groups"`,
			"gorm.DB",
			"repository.PostgresRepository",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s contains storage dependency %q", sourcePath, forbidden)
			}
		}
	}
}

func TestLegacyEventsServiceDoesNotExposeUserReadMethods(t *testing.T) {
	serviceInterface := reflect.TypeOf((*servicesevents.EventsService)(nil)).Elem()
	for _, methodName := range []string{"SearchEvents", "GetGroupEvents", "GetEventDetails"} {
		if _, exists := serviceInterface.MethodByName(methodName); exists {
			t.Fatalf("EventsService still exposes %s; keep user reads in EventReadService", methodName)
		}
	}
}

func TestNewEventMembershipServiceAcceptsOnlyLoggerAndUnitOfWork(t *testing.T) {
	constructorType := reflect.TypeOf(servicesevents.NewEventMembershipService)
	if constructorType.NumIn() != 2 {
		t.Fatalf("NewEventMembershipService has %d arguments, want 2", constructorType.NumIn())
	}

	loggerType := reflect.TypeOf((*applicationlogger.Logger)(nil)).Elem()
	if constructorType.In(0) != loggerType {
		t.Fatalf("NewEventMembershipService first argument = %v, want %v", constructorType.In(0), loggerType)
	}
	uowType := reflect.TypeOf((*servicesevents.EventUnitOfWork)(nil)).Elem()
	if constructorType.In(1) != uowType {
		t.Fatalf("NewEventMembershipService second argument = %v, want %v", constructorType.In(1), uowType)
	}

	uow := &eventUnitOfWorkStub{}
	out := reflect.ValueOf(servicesevents.NewEventMembershipService).Call([]reflect.Value{
		reflect.ValueOf(&testLogger{}),
		reflect.ValueOf(uow),
	})
	service, ok := out[0].Interface().(servicesevents.EventMembershipService)
	if !ok || service == nil {
		t.Fatalf("NewEventMembershipService returned %T, want EventMembershipService", out[0].Interface())
	}

	repo := &testPostgresRepository{}
	if constructorAcceptsRepo(servicesevents.NewEventMembershipService, repo) {
		t.Fatal("NewEventMembershipService accepts broad GORM-shaped repository; want EventUnitOfWork")
	}
}

func TestEventsServiceDoesNotExposeCommandsMembershipOrStoreUnitOfWork(t *testing.T) {
	serviceInterface := reflect.TypeOf((*servicesevents.EventsService)(nil)).Elem()
	for _, methodName := range []string{
		"CreateEvent",
		"UpdateEvent",
		"DeleteEvent",
		"JoinEvent",
		"LeaveEvent",
		"SearchEvents",
		"GetGroupEvents",
		"GetEventDetails",
	} {
		if _, exists := serviceInterface.MethodByName(methodName); exists {
			t.Fatalf("EventsService still exposes %s; keep commands, membership and reads in focused services", methodName)
		}
	}

	db := newEventsServiceDB(t)
	service := servicesevents.NewEventsService(&testLogger{}, &testPostgresRepository{db: db})
	implementationType := reflect.TypeOf(service)
	if implementationType.Kind() == reflect.Pointer {
		implementationType = implementationType.Elem()
	}
	uowType := reflect.TypeOf((*servicesevents.EventUnitOfWork)(nil)).Elem()
	for i := 0; i < implementationType.NumField(); i++ {
		field := implementationType.Field(i)
		if field.Type.Implements(uowType) || strings.Contains(strings.ToLower(field.Name), "unitofwork") {
			t.Fatalf("EventsService implementation still stores membership unit of work in field %s", field.Name)
		}
	}

	serviceSourcePath := filepath.Join("..", "services", "events", "ServiceEvents.go")
	content, err := os.ReadFile(serviceSourcePath)
	if err != nil {
		t.Fatalf("read events service source: %v", err)
	}
	if strings.Contains(string(content), "NewEventsServiceWithUnitOfWork") {
		t.Fatal("EventsService still exposes a membership-aware constructor")
	}
}

func assertEventReadTypeDoesNotLeakStorage(t *testing.T, typ reflect.Type, checked map[reflect.Type]bool) {
	t.Helper()

	if typ == nil || checked[typ] {
		return
	}
	checked[typ] = true

	packagePath := typ.PkgPath()
	if strings.HasPrefix(packagePath, "gorm.io/gorm") ||
		packagePath == "friendship/repository" ||
		packagePath == "friendship/models/events" ||
		packagePath == "friendship/models/groups" {
		t.Fatalf("event read contract leaks storage type %v from %s", typ, packagePath)
	}

	switch typ.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		assertEventReadTypeDoesNotLeakStorage(t, typ.Elem(), checked)
	case reflect.Map:
		assertEventReadTypeDoesNotLeakStorage(t, typ.Key(), checked)
		assertEventReadTypeDoesNotLeakStorage(t, typ.Elem(), checked)
	case reflect.Struct:
		if packagePath != "friendship/services/events" {
			return
		}
		for i := 0; i < typ.NumField(); i++ {
			assertEventReadTypeDoesNotLeakStorage(t, typ.Field(i).Type, checked)
		}
	}
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
