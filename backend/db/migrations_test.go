package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"friendship/models"
	"friendship/models/events"
	"friendship/models/groups"
	statsusers "friendship/models/stats_users"
	"friendship/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository.PostgresRepository = (*recordingRepository)(nil)
var _ repository.PostgresRepository = (*gormRepository)(nil)

type recordingRepository struct {
	autoMigrated   []reflect.Type
	getSQLDBCalled bool
}

func (r *recordingRepository) Model(value interface{}) *gorm.DB { panic("unexpected call to Model") }
func (r *recordingRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	panic("unexpected call to Select")
}
func (r *recordingRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	panic("unexpected call to Find")
}
func (r *recordingRepository) Exec(sql string, values ...interface{}) *gorm.DB {
	panic("unexpected call to Exec")
}
func (r *recordingRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	panic("unexpected call to First")
}
func (r *recordingRepository) Raw(sql string, values ...interface{}) *gorm.DB {
	panic("unexpected call to Raw")
}
func (r *recordingRepository) Create(value interface{}) *gorm.DB { panic("unexpected call to Create") }
func (r *recordingRepository) Save(value interface{}) *gorm.DB   { panic("unexpected call to Save") }
func (r *recordingRepository) Updates(value interface{}) *gorm.DB {
	panic("unexpected call to Updates")
}
func (r *recordingRepository) Delete(value interface{}) *gorm.DB { panic("unexpected call to Delete") }
func (r *recordingRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	panic("unexpected call to Where")
}
func (r *recordingRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	panic("unexpected call to Preload")
}
func (r *recordingRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	panic("unexpected call to Scopes")
}
func (r *recordingRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	panic("unexpected call to ScanRows")
}
func (r *recordingRepository) Transaction(fc func(tx repository.PostgresRepository) error) error {
	panic("unexpected call to Transaction")
}
func (r *recordingRepository) Close() error { panic("unexpected call to Close") }
func (r *recordingRepository) DropTableIfExists(value interface{}) error {
	panic("unexpected call to DropTableIfExists")
}
func (r *recordingRepository) GetSQLDB() (*sql.DB, error) {
	r.getSQLDBCalled = true
	return nil, nil
}
func (r *recordingRepository) Clauses(conds ...clause.Expression) *gorm.DB {
	panic("unexpected call to Clauses")
}
func (r *recordingRepository) AutoMigrate(value interface{}) error {
	r.autoMigrated = append(r.autoMigrated, reflect.TypeOf(value))
	return nil
}
func (r *recordingRepository) Order(value interface{}) *gorm.DB { panic("unexpected call to Order") }
func (r *recordingRepository) Limit(limit int) *gorm.DB         { panic("unexpected call to Limit") }
func (r *recordingRepository) Count(count *int64) *gorm.DB      { panic("unexpected call to Count") }
func (r *recordingRepository) Association(column string) *gorm.Association {
	panic("unexpected call to Association")
}

type noopLogger struct{}

func (noopLogger) Info(string, ...interface{})  {}
func (noopLogger) Error(string, ...interface{}) {}
func (noopLogger) Debug(string, ...interface{}) {}
func (noopLogger) Warn(string, ...interface{})  {}
func (noopLogger) Fatal(string, ...interface{}) {}
func (noopLogger) Panic(string, ...interface{}) {}

type gormRepository struct {
	db *gorm.DB
}

func (r *gormRepository) Model(value interface{}) *gorm.DB { return r.db.Model(value) }
func (r *gormRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Select(query, args...)
}
func (r *gormRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.Find(out, where...)
}
func (r *gormRepository) Exec(query string, values ...interface{}) *gorm.DB {
	return r.db.Exec(query, values...)
}
func (r *gormRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.First(out, where...)
}
func (r *gormRepository) Raw(query string, values ...interface{}) *gorm.DB {
	return r.db.Raw(query, values...)
}
func (r *gormRepository) Create(value interface{}) *gorm.DB  { return r.db.Create(value) }
func (r *gormRepository) Save(value interface{}) *gorm.DB    { return r.db.Save(value) }
func (r *gormRepository) Updates(value interface{}) *gorm.DB { return r.db.Updates(value) }
func (r *gormRepository) Delete(value interface{}) *gorm.DB  { return r.db.Delete(value) }
func (r *gormRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Where(query, args...)
}
func (r *gormRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return r.db.Preload(column, conditions...)
}
func (r *gormRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return r.db.Scopes(funcs...)
}
func (r *gormRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	return r.db.ScanRows(rows, result)
}
func (r *gormRepository) Transaction(fc func(tx repository.PostgresRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&gormRepository{db: tx})
	})
}
func (r *gormRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (r *gormRepository) DropTableIfExists(value interface{}) error {
	return r.db.Migrator().DropTable(value)
}
func (r *gormRepository) GetSQLDB() (*sql.DB, error) { return r.db.DB() }
func (r *gormRepository) Clauses(conds ...clause.Expression) *gorm.DB {
	return r.db.Clauses(conds...)
}
func (r *gormRepository) AutoMigrate(value interface{}) error { return r.db.AutoMigrate(value) }
func (r *gormRepository) Order(value interface{}) *gorm.DB    { return r.db.Order(value) }
func (r *gormRepository) Limit(limit int) *gorm.DB            { return r.db.Limit(limit) }
func (r *gormRepository) Count(count *int64) *gorm.DB         { return r.db.Count(count) }
func (r *gormRepository) Association(column string) *gorm.Association {
	return r.db.Association(column)
}

func TestBootstrapRegistrationSchemaMigratesExpectedModels(t *testing.T) {
	repo := &recordingRepository{}

	if err := BootstrapRegistrationSchema(repo); err != nil {
		t.Fatalf("BootstrapRegistrationSchema returned error: %v", err)
	}

	want := []reflect.Type{
		reflect.TypeOf(&events.Event{}),
		reflect.TypeOf(&events.AgeLimit{}),
		reflect.TypeOf(&events.EventLocation{}),
		reflect.TypeOf(&events.Status{}),
		reflect.TypeOf(&events.EventsUser{}),
		reflect.TypeOf(&events.Genre{}),
		reflect.TypeOf(&events.EventGenre{}),
		reflect.TypeOf(&statsusers.PopSessionType{}),
		reflect.TypeOf(&models.User{}),
		reflect.TypeOf(&models.StatsProcessedEvent{}),
		reflect.TypeOf(&models.DaysWeek{}),
		reflect.TypeOf(&models.Category{}),
		reflect.TypeOf(&groups.Role_in_group{}),
		reflect.TypeOf(&groups.Group{}),
		reflect.TypeOf(&groups.GroupContact{}),
		reflect.TypeOf(&groups.GroupGroupCategory{}),
		reflect.TypeOf(&groups.GroupUsers{}),
		reflect.TypeOf(&groups.GroupJoinRequest{}),
		reflect.TypeOf(&groups.GroupJoinInvite{}),
		reflect.TypeOf(&groups.GroupBlacklist{}),
		reflect.TypeOf(&groups.GroupActionLog{}),
		reflect.TypeOf(&statsusers.Genre{}),
		reflect.TypeOf(&statsusers.SettingTile{}),
		reflect.TypeOf(&statsusers.SessionStats_users{}),
		reflect.TypeOf(&statsusers.SideStats_users{}),
	}

	if !reflect.DeepEqual(repo.autoMigrated, want) {
		t.Fatalf("unexpected models migrated:\nwant: %#v\ngot:  %#v", want, repo.autoMigrated)
	}
}

func TestBootstrapRegistrationSchemaCreatesRegistrationTables(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	repo := &gormRepository{db: gormDB}
	if err := BootstrapRegistrationSchema(repo); err != nil {
		t.Fatalf("BootstrapRegistrationSchema returned error: %v", err)
	}

	for _, table := range []interface{}{
		&events.Event{},
		&events.AgeLimit{},
		&events.EventLocation{},
		&events.Status{},
		&events.EventsUser{},
		&events.Genre{},
		&events.EventGenre{},
		&statsusers.PopSessionType{},
		&models.User{},
		&models.StatsProcessedEvent{},
		&models.DaysWeek{},
		&models.Category{},
		&groups.Role_in_group{},
		&groups.Group{},
		&groups.GroupContact{},
		&groups.GroupGroupCategory{},
		&groups.GroupUsers{},
		&groups.GroupJoinRequest{},
		&groups.GroupJoinInvite{},
		&groups.GroupBlacklist{},
		&groups.GroupActionLog{},
		&statsusers.Genre{},
		&statsusers.SettingTile{},
		&statsusers.SessionStats_users{},
		&statsusers.SideStats_users{},
	} {
		if !gormDB.Migrator().HasTable(table) {
			t.Fatalf("expected table for %T to exist after bootstrap", table)
		}
	}
}

func TestBootstrapRegistrationSchemaSupportsSeederOnFreshDB(t *testing.T) {
	restoreWorkingDirDBTest(t, "..")

	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	repo := &gormRepository{db: gormDB}
	if err := BootstrapRegistrationSchema(repo); err != nil {
		t.Fatalf("BootstrapRegistrationSchema returned error: %v", err)
	}

	errs := Seeder(repo)
	if len(errs) > 0 {
		t.Fatalf("Seeder returned errors after bootstrap: %v", errs)
	}
}

func TestResolveMigrationSourcePrefersExistingSQLDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "migration"), 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "0001_registration.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write migration file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "README.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write non-sql marker: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := resolveMigrationSource()
	if err != nil {
		t.Fatalf("resolveMigrationSource returned error: %v", err)
	}

	want := "file://" + filepath.ToSlash(filepath.Join(tempDir, "migration"))
	if got != want {
		t.Fatalf("unexpected migration source:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestHasMigrationSourceReturnsFalseWhenSQLMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := HasMigrationSource()
	if err != nil {
		t.Fatalf("HasMigrationSource returned error: %v", err)
	}
	if got {
		t.Fatal("HasMigrationSource returned true, want false")
	}
}

func TestResolveMigrationSourceUsesBackendMigrationFallback(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "backend", "migration"), 0o755); err != nil {
		t.Fatalf("mkdir backend/migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "backend", "migration", "0001_init.SQL"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write migration file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := resolveMigrationSource()
	if err != nil {
		t.Fatalf("resolveMigrationSource returned error: %v", err)
	}

	want := "file://" + filepath.ToSlash(filepath.Join(tempDir, "backend", "migration"))
	if got != want {
		t.Fatalf("unexpected migration source:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestResolveMigrationSourceUsesBackendMigrationsAsLastFallback(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "backend", "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir backend/migrations: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "backend", "migrations", "0002_seed.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write migration file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := resolveMigrationSource()
	if err != nil {
		t.Fatalf("resolveMigrationSource returned error: %v", err)
	}

	want := "file://" + filepath.ToSlash(filepath.Join(tempDir, "backend", "migrations"))
	if got != want {
		t.Fatalf("unexpected migration source:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestHasMigrationSourceReturnsTrueWhenFallbackDirectoryContainsSQL(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "backend", "migration"), 0o755); err != nil {
		t.Fatalf("mkdir backend/migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "backend", "migration", "0001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write migration file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := HasMigrationSource()
	if err != nil {
		t.Fatalf("HasMigrationSource returned error: %v", err)
	}
	if !got {
		t.Fatal("HasMigrationSource returned false, want true")
	}
}

func TestHasMigrationSourceReturnsFalseWhenOnlyNonSQLFilesExist(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "migration"), 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "README.md"), []byte("no sql"), 0o644); err != nil {
		t.Fatalf("write marker file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := HasMigrationSource()
	if err != nil {
		t.Fatalf("HasMigrationSource returned error: %v", err)
	}
	if got {
		t.Fatal("HasMigrationSource returned true, want false")
	}
}

func TestHasMigrationSourceReturnsFalseWhenCanonicalHasOnlyRunbook(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "migration"), 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "membership_uniqueness_runbook.ru.md"), []byte("runbook"), 0o644); err != nil {
		t.Fatalf("write runbook file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := HasMigrationSource()
	if err != nil {
		t.Fatalf("HasMigrationSource returned error: %v", err)
	}
	if got {
		t.Fatal("HasMigrationSource returned true with runbook-only canonical migration dir, want false")
	}
}

func TestRequireMigrationSourceFailsWhenSQLMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	_, err = RequireMigrationSource()
	if err == nil {
		t.Fatal("RequireMigrationSource returned nil error, want missing source error")
	}
	if !strings.Contains(err.Error(), "migration source is required for rollout phase") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireMigrationSourceReturnsSourceWhenSQLPresent(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "migration"), 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "0001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write sql file: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	got, err := RequireMigrationSource()
	if err != nil {
		t.Fatalf("RequireMigrationSource returned error: %v", err)
	}
	want := "file://" + filepath.ToSlash(filepath.Join(tempDir, "migration"))
	if got != want {
		t.Fatalf("unexpected migration source:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestRequireMigrationSourceFailsWhenResolvedDirectoryHasNoSQLFiles(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "migration"), 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "membership_uniqueness_runbook.ru.md"), []byte("runbook"), 0o644); err != nil {
		t.Fatalf("write runbook file: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	_, err = RequireMigrationSource()
	if err == nil {
		t.Fatal("RequireMigrationSource returned nil error, want strict missing-sql error")
	}
	if !strings.Contains(err.Error(), "does not contain SQL migration files") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrationDBStrictFailsBeforeDBLookupWhenSourceMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	repo := &recordingRepository{}
	err = MigrationDBStrict(repo, noopLogger{})
	if err == nil {
		t.Fatal("MigrationDBStrict returned nil error, want strict missing source error")
	}
	if !strings.Contains(err.Error(), "migration source is required for rollout phase") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.getSQLDBCalled {
		t.Fatal("MigrationDBStrict called GetSQLDB before validating migration source")
	}
}

func TestMigrationDBStrictToVersionFailsBeforeDBLookupWhenSourceMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	repo := &recordingRepository{}
	err = MigrationDBStrictToVersion(repo, noopLogger{}, 2)
	if err == nil {
		t.Fatal("MigrationDBStrictToVersion returned nil error, want strict missing source error")
	}
	if !strings.Contains(err.Error(), "migration source is required for rollout phase") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.getSQLDBCalled {
		t.Fatal("MigrationDBStrictToVersion called GetSQLDB before validating migration source")
	}
}

func TestPendingJoinRequestUniquenessMigrationArtifacts(t *testing.T) {
	repoRoot := repoRootFromDBTest(t)
	files := map[string][]string{
		filepath.Join(repoRoot, "migration", "000002_group_join_request_pending_uniqueness.up.sql"): {
			"CREATE UNIQUE INDEX idx_group_join_request_pending_unique",
			"ON public.group_join_requests (user_id, group_id)",
			"WHERE status = 'pending'",
		},
		filepath.Join(repoRoot, "migration", "000002_group_join_request_pending_uniqueness_preflight.sql"): {
			"group_join_request_pending_uniqueness_preflight_checks",
			"data.pending_join_request_duplicates",
			"WHERE status = ''pending''",
			"RAISE EXCEPTION",
		},
		filepath.Join(repoRoot, "migration", "000002_group_join_request_pending_uniqueness_postflight.sql"): {
			"group_join_request_pending_uniqueness_postflight_checks",
			"index_contract.idx_group_join_request_pending_unique",
			"data.pending_join_request_duplicates_removed",
			"RAISE EXCEPTION",
		},
	}

	for path, snippets := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			sqlText := string(content)
			for _, snippet := range snippets {
				if !strings.Contains(sqlText, snippet) {
					t.Fatalf("%s does not contain %q", path, snippet)
				}
			}
		})
	}
}

func TestMembershipUniquenessMigrationArtifacts(t *testing.T) {
	repoRoot := repoRootFromDBTest(t)
	files := map[string][]string{
		filepath.Join(repoRoot, "migration", "000001_membership_uniqueness.up.sql"): {
			"membership_dedupe_audit",
			"DELETE FROM public.group_users",
			"DELETE FROM public.events_users",
			"UPDATE public.events e",
			"CREATE UNIQUE INDEX idx_group_user_membership",
			"CREATE UNIQUE INDEX idx_event_user_membership",
		},
		filepath.Join(repoRoot, "migration", "000001_membership_uniqueness_preflight.sql"): {
			"membership_uniqueness_preflight_checks",
			"readiness.group_users_duplicates",
			"readiness.events_users_duplicates",
			"readiness.events_current_users_drift",
			"RAISE EXCEPTION",
		},
		filepath.Join(repoRoot, "migration", "000001_membership_uniqueness_postflight.sql"): {
			"membership_uniqueness_postflight_checks",
			"data.group_users_duplicates_removed",
			"data.events_users_duplicates_removed",
			"data.events_current_users_reconciled",
			"RAISE EXCEPTION",
		},
		filepath.Join(repoRoot, "migration", "000001_membership_uniqueness_rollout_checks.sql"): {
			"duplicate group_users with chosen keeper",
			"duplicate events_users",
			"current_users drift",
			"postflight: duplicate keys must be zero",
		},
	}

	for path, snippets := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			sqlText := string(content)
			for _, snippet := range snippets {
				if !strings.Contains(sqlText, snippet) {
					t.Fatalf("%s does not contain %q", path, snippet)
				}
			}
		})
	}
}

func repoRootFromDBTest(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

func TestResolveMigrationDirectoryFailsOnCanonicalAndLegacyAmbiguity(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "migration"), 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migration", "0001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write canonical sql file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "backend", "migration"), 0o755); err != nil {
		t.Fatalf("mkdir backend/migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "backend", "migration", "0001_legacy.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write legacy sql file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	_, err = ResolveMigrationDirectory()
	if err == nil {
		t.Fatal("ResolveMigrationDirectory returned nil error, want ambiguity error")
	}
	if !strings.Contains(err.Error(), "ambiguous migration discovery: canonical") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveMigrationDirectoryFailsOnMultipleLegacyDirsAmbiguity(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "backend", "migration"), 0o755); err != nil {
		t.Fatalf("mkdir backend/migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "backend", "migration", "0001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write backend/migration sql file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "migrations", "0001_alt.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write migrations sql file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd: %v", chdirErr)
		}
	})

	_, err = ResolveMigrationDirectory()
	if err == nil {
		t.Fatal("ResolveMigrationDirectory returned nil error, want ambiguity error")
	}
	if !strings.Contains(err.Error(), "ambiguous migration discovery across legacy directories") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHasCoreSchemaTablesReturnsFalseForPartialSchema(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := gormDB.Exec("ATTACH DATABASE ':memory:' AS information_schema").Error; err != nil {
		t.Fatalf("attach information_schema db: %v", err)
	}
	if err := gormDB.Exec(`
		CREATE TABLE information_schema.tables (
			table_schema TEXT,
			table_name TEXT
		)
	`).Error; err != nil {
		t.Fatalf("create information_schema.tables: %v", err)
	}

	requiredTables, err := bootstrapRegistrationTableNames()
	if err != nil {
		t.Fatalf("bootstrapRegistrationTableNames returned error: %v", err)
	}
	requiredTables = requiredTables[:len(requiredTables)-1]
	for _, name := range requiredTables {
		if err := gormDB.Exec(
			"INSERT INTO information_schema.tables(table_schema, table_name) VALUES (?, ?)",
			"public",
			name,
		).Error; err != nil {
			t.Fatalf("insert table %q: %v", name, err)
		}
	}

	ready, err := HasCoreSchemaTables(&gormRepository{db: gormDB})
	if err != nil {
		t.Fatalf("HasCoreSchemaTables returned error: %v", err)
	}
	if ready {
		t.Fatal("HasCoreSchemaTables returned true for partial schema, want false")
	}
}

func TestHasCoreSchemaTablesReturnsTrueForBootstrapSchemaSet(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := gormDB.Exec("ATTACH DATABASE ':memory:' AS information_schema").Error; err != nil {
		t.Fatalf("attach information_schema db: %v", err)
	}
	if err := gormDB.Exec(`
		CREATE TABLE information_schema.tables (
			table_schema TEXT,
			table_name TEXT
		)
	`).Error; err != nil {
		t.Fatalf("create information_schema.tables: %v", err)
	}

	requiredTables, err := bootstrapRegistrationTableNames()
	if err != nil {
		t.Fatalf("bootstrapRegistrationTableNames returned error: %v", err)
	}
	for _, name := range requiredTables {
		if err := gormDB.Exec(
			"INSERT INTO information_schema.tables(table_schema, table_name) VALUES (?, ?)",
			"public",
			name,
		).Error; err != nil {
			t.Fatalf("insert table %q: %v", name, err)
		}
	}

	ready, err := HasCoreSchemaTables(&gormRepository{db: gormDB})
	if err != nil {
		t.Fatalf("HasCoreSchemaTables returned error: %v", err)
	}
	if !ready {
		t.Fatal("HasCoreSchemaTables returned false for full bootstrap schema, want true")
	}
}

func restoreWorkingDirDBTest(t *testing.T, target string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	if err := os.Chdir(target); err != nil {
		t.Fatalf("change working dir to %s: %v", target, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
}
