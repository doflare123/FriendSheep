package tests

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	applicationlogger "friendship/logger"
	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
	servicegroups "friendship/services/groups"
)

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

	result, err := service.JoinGroup(context.Background(), 2, groupID)
	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
}

func TestGroupSharedRepositoryOnlyAppearsInAdapters(t *testing.T) {
	groupsDir := filepath.Join("..", "services", "groups")
	allowedAdapters := groupPersistenceAdapterFiles()

	err := filepath.WalkDir(groupsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") || allowedAdapters[filepath.Base(path)] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(content)
		if strings.Contains(text, `"friendship/repository"`) || strings.Contains(text, "repository.PostgresRepository") {
			t.Fatalf("%s imports or references shared Postgres repository outside the explicit group adapter allowlist", path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk group services: %v", err)
	}
}

func TestGroupApplicationContractsDoNotLeakPersistenceTypes(t *testing.T) {
	groupsDir := filepath.Join("..", "services", "groups")
	applicationFiles := []string{filepath.Join(groupsDir, "store.go")}

	serviceFiles, err := filepath.Glob(filepath.Join(groupsDir, "Service*.go"))
	if err != nil {
		t.Fatalf("glob group service files: %v", err)
	}
	if len(serviceFiles) == 0 {
		t.Fatal("no group application service files found")
	}
	applicationFiles = append(applicationFiles, serviceFiles...)

	for _, path := range applicationFiles {
		assertGoSourceDoesNotUsePersistence(t, path)
	}
}

func TestGroupPersistenceLibrariesAreIsolatedInAdapters(t *testing.T) {
	groupsDir := filepath.Join("..", "services", "groups")
	allowedAdapters := groupPersistenceAdapterFiles()

	err := filepath.WalkDir(groupsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if allowedAdapters[filepath.Base(path)] {
			return nil
		}

		assertGoSourceDoesNotUsePersistence(t, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk group service sources: %v", err)
	}
}

func TestGroupApplicationInterfacesDoNotLeakPersistenceTypes(t *testing.T) {
	groupsDir := filepath.Join("..", "services", "groups")
	allowedAdapterInterfaces := map[string]bool{
		"groupRepositoryTransactor": true,
	}
	adapterPersistenceTypes := map[string]bool{
		"gormGroupLookupStore": true,
		"gormGroupStore":       true,
	}

	err := filepath.WalkDir(groupsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		persistenceImports := persistenceImportAliases(parsed)
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || allowedAdapterInterfaces[typeSpec.Name.Name] {
					continue
				}
				contract, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				ast.Inspect(contract, func(node ast.Node) bool {
					identifier, ok := node.(*ast.Ident)
					if ok && adapterPersistenceTypes[identifier.Name] {
						t.Fatalf("%s interface %s leaks adapter persistence type %s", path, typeSpec.Name.Name, identifier.Name)
					}

					selector, ok := node.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					qualifier, ok := selector.X.(*ast.Ident)
					if ok && persistenceImports[qualifier.Name] {
						t.Fatalf("%s interface %s leaks persistence type %s.%s", path, typeSpec.Name.Name, qualifier.Name, selector.Sel.Name)
					}
					return true
				})
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("inspect group application interfaces: %v", err)
	}
}

func TestGroupSharedGORMStoreAbstractionIsRemoved(t *testing.T) {
	groupsDir := filepath.Join("..", "services", "groups")
	removedSupportFile := filepath.Join(groupsDir, "gorm_store_support.go")
	if _, err := os.Stat(removedSupportFile); err == nil {
		t.Fatalf("%s still exists; group adapters must depend directly on their explicit repository ports", removedSupportFile)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat removed group GORM support file %s: %v", removedSupportFile, err)
	}

	err := filepath.WalkDir(groupsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, removedType := range []string{"gormGroupStore", "gormGroupLookupStore"} {
			if strings.Contains(string(content), removedType) {
				t.Fatalf("%s still references removed shared adapter abstraction %s", path, removedType)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify removed shared group GORM abstraction: %v", err)
	}
}

func TestGroupActionTypeModelDoesNotExportGORMShapedLookup(t *testing.T) {
	path := filepath.Join("..", "models", "groups", "GroupActionType.go")
	assertGoSourceDoesNotUsePersistence(t, path)

	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse group action type model %s: %v", path, err)
	}

	for _, declaration := range parsed.Decls {
		switch declaration := declaration.(type) {
		case *ast.GenDecl:
			for _, spec := range declaration.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok && typeSpec.Name.Name == "ActionTypeLookup" {
					t.Fatalf("%s exports ActionTypeLookup; persistence lookup belongs in a GORM adapter", path)
				}
			}
		case *ast.FuncDecl:
			if declaration.Recv == nil && declaration.Name.Name == "FindGroupActionTypeID" {
				t.Fatalf("%s exports FindGroupActionTypeID; action-type lookup belongs in a GORM adapter", path)
			}
		}
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

func TestNewEventAdminServiceAcceptsOnlyCleanAdminDependencies(t *testing.T) {
	constructorType := reflect.TypeOf(servicesevents.NewEventAdminService)
	if constructorType.NumIn() != 3 {
		t.Fatalf("NewEventAdminService has %d arguments, want 3", constructorType.NumIn())
	}

	loggerType := reflect.TypeOf((*applicationlogger.Logger)(nil)).Elem()
	readerType := reflect.TypeOf((*servicesevents.EventAdminReader)(nil)).Elem()
	uowType := reflect.TypeOf((*servicesevents.EventUnitOfWork)(nil)).Elem()
	serviceType := reflect.TypeOf((*servicesevents.EventAdminService)(nil)).Elem()
	for index, wantType := range []reflect.Type{loggerType, readerType, uowType} {
		if constructorType.In(index) != wantType {
			t.Fatalf("NewEventAdminService argument %d = %v, want %v", index, constructorType.In(index), wantType)
		}
	}
	if constructorType.NumOut() != 1 || constructorType.Out(0) != serviceType {
		t.Fatalf("NewEventAdminService result = %v, want %v", constructorType.Out(0), serviceType)
	}

	repoType := reflect.TypeOf(&testPostgresRepository{})
	for index := 0; index < constructorType.NumIn(); index++ {
		argumentType := constructorType.In(index)
		if repoType.AssignableTo(argumentType) ||
			(argumentType.Kind() == reflect.Interface && repoType.Implements(argumentType)) {
			t.Fatalf("NewEventAdminService argument %d accepts broad GORM-shaped repository", index)
		}
	}
}

func TestEventAdminContractsDoNotLeakStorageTypes(t *testing.T) {
	checked := make(map[reflect.Type]bool)
	for _, contract := range []reflect.Type{
		reflect.TypeOf((*servicesevents.EventAdminService)(nil)).Elem(),
		reflect.TypeOf((*servicesevents.EventAdminReader)(nil)).Elem(),
		reflect.TypeOf((*servicesevents.EventAdminStore)(nil)).Elem(),
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

func TestEventAdminContractsAreDeclaredInSource(t *testing.T) {
	eventsDir := filepath.Join("..", "services", "events")
	combined := readCombinedGoSource(t, eventsDir)

	for _, snippet := range []string{
		"type EventAdminService interface",
		"type EventAdminReader interface",
		"type EventAdminStore interface",
		"NewEventAdminService(",
		"NewGORMEventAdminReader(",
		"func (tx EventTransaction) Admin() EventAdminStore",
		"AdminStore      EventAdminStore",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("event admin contract is missing source snippet %q", snippet)
		}
	}
}

func TestEventAdminCleanSourceDoesNotImportStorageLibraries(t *testing.T) {
	eventsDir := filepath.Join("..", "services", "events")
	matches, err := filepath.Glob(filepath.Join(eventsDir, "*admin*.go"))
	if err != nil {
		t.Fatalf("glob admin event sources: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no admin event source files found")
	}

	var checkedCleanFile bool
	for _, path := range matches {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read admin event source %s: %v", path, err)
		}

		text := string(content)
		fileName := strings.ToLower(filepath.Base(path))
		if strings.Contains(fileName, "gorm") {
			continue
		}

		checkedCleanFile = true
		for _, forbidden := range []string{
			`"friendship/repository"`,
			`"gorm.io/`,
			`"friendship/models/events"`,
			"repository.PostgresRepository",
			"gorm.DB",
			"eventmodels.",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s contains storage dependency %q; keep it isolated in GORM admin adapters", path, forbidden)
			}
		}
	}
	if !checkedCleanFile {
		t.Fatal("no clean admin event service source file found outside GORM adapters")
	}
}

func readCombinedGoSource(t *testing.T, dir string) string {
	t.Helper()

	var builder strings.Builder
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		builder.Write(content)
		builder.WriteByte('\n')
		return nil
	})
	if err != nil {
		t.Fatalf("walk go source %s: %v", dir, err)
	}

	return builder.String()
}

func assertGoSourceDoesNotUsePersistence(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Go source %s: %v", path, err)
	}

	parsed, err := parser.ParseFile(token.NewFileSet(), path, content, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse Go source %s: %v", path, err)
	}
	for _, sourceImport := range parsed.Imports {
		importPath := strings.Trim(sourceImport.Path.Value, `"`)
		if importPath == "friendship/repository" || strings.HasPrefix(importPath, "gorm.io/") {
			t.Fatalf("%s imports persistence implementation %q; keep it in an explicitly allowed group adapter", path, importPath)
		}
	}

	text := string(content)
	for _, forbidden := range []string{
		"repository.PostgresRepository",
		"*gorm.DB",
		"gorm.DB",
		"clause.Expression",
		"gormGroupStore",
		"gormGroupLookupStore",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("%s contains persistence type %q; application contracts must remain storage-agnostic", path, forbidden)
		}
	}
}

func groupPersistenceAdapterFiles() map[string]bool {
	return map[string]bool{
		"gorm_helpers.go":                   true,
		"gorm_repository.go":                true,
		"group_admin_store.go":              true,
		"group_management_read_store.go":    true,
		"gorm_group_subscriptions_store.go": true,
		"group_shared_store.go":             true,
		"join_invite_store.go":              true,
		"join_request_store.go":             true,
		"membership_store.go":               true,
	}
}

func persistenceImportAliases(file *ast.File) map[string]bool {
	aliases := make(map[string]bool)
	for _, sourceImport := range file.Imports {
		importPath := strings.Trim(sourceImport.Path.Value, `"`)
		if importPath != "friendship/repository" && !strings.HasPrefix(importPath, "gorm.io/") {
			continue
		}

		alias := ""
		if sourceImport.Name != nil {
			alias = sourceImport.Name.Name
		} else if separator := strings.LastIndex(importPath, "/"); separator >= 0 {
			alias = importPath[separator+1:]
		} else {
			alias = importPath
		}
		aliases[alias] = true
	}
	return aliases
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
