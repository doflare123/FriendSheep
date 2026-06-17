package db

import (
	"errors"
	"fmt"
	"friendship/logger"
	"friendship/models"
	"friendship/models/events"
	"friendship/models/groups"
	statsusers "friendship/models/stats_users"
	"friendship/repository"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm/schema"
)

const canonicalMigrationDir = "migration"

var legacyMigrationDirs = []string{
	filepath.Join(".", "backend", "migration"),
	filepath.Join(".", "migrations"),
	filepath.Join(".", "backend", "migrations"),
}

func AutoMigDB(db repository.PostgresRepository, models ...interface{}) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}

func bootstrapRegistrationModels() []interface{} {
	return []interface{}{
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
		&groups.GroupActionType{},
		&groups.GroupActionLog{},
		&statsusers.Genre{},
		&statsusers.SettingTile{},
		&statsusers.SessionStats_users{},
		&statsusers.SideStats_users{},
	}
}

func BootstrapRegistrationSchema(db repository.PostgresRepository) error {
	return AutoMigDB(db, bootstrapRegistrationModels()...)
}

func HasMigrationSource() (bool, error) {
	migrationDir, err := ResolveMigrationDirectory()
	if err != nil {
		return false, err
	}
	if migrationDir == "" {
		return false, nil
	}

	hasSQL, err := hasSQLMigrationFiles(migrationDir)
	if err != nil {
		return false, fmt.Errorf("read migration dir %s: %w", migrationDir, err)
	}
	return hasSQL, nil
}

func RequireMigrationSource() (string, error) {
	migrationSource, err := resolveMigrationSourceStrict()
	if err != nil {
		return "", err
	}
	return migrationSource, nil
}

func MigrationDB(db repository.PostgresRepository, logger logger.Logger) error {
	return migrationDB(db, logger, false)
}

func MigrationDBStrict(db repository.PostgresRepository, logger logger.Logger) error {
	return migrationDB(db, logger, true)
}

func MigrationDBStrictToVersion(db repository.PostgresRepository, logger logger.Logger, targetVersion uint) error {
	return migrationDBToVersion(db, logger, targetVersion)
}

func migrationDB(db repository.PostgresRepository, logger logger.Logger, strict bool) error {
	migrationsPath, err := resolveMigrationSourceWithMode(strict)
	if err != nil {
		return err
	}
	if migrationsPath == "" {
		logger.Info("No SQL migrations found, skipping migrate bootstrap")
		return nil
	}

	sqlDB, err := db.GetSQLDB()
	if err != nil {
		return err
	}
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return err
	}
	migrat, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}
	if err := migrat.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("No new migrations to apply")
		} else {
			return err
		}
	}
	logger.Info("Migrations done")
	return nil
}

func migrationDBToVersion(db repository.PostgresRepository, logger logger.Logger, targetVersion uint) error {
	migrationsPath, err := resolveMigrationSourceWithMode(true)
	if err != nil {
		return err
	}
	if migrationsPath == "" {
		return errors.New("migration source is required for rollout target version")
	}

	sqlDB, err := db.GetSQLDB()
	if err != nil {
		return err
	}
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return err
	}
	migrat, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}

	currentVersion, dirty, versionErr := migrat.Version()
	if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
		return versionErr
	}
	if dirty {
		return fmt.Errorf("database is in dirty migration state at version %d", currentVersion)
	}
	if versionErr == nil && currentVersion > uint(targetVersion) {
		return fmt.Errorf(
			"current migration version %d is above requested target %d; refusing to migrate down",
			currentVersion,
			targetVersion,
		)
	}

	if err := migrat.Migrate(uint(targetVersion)); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No new migrations to apply for target", "targetVersion", targetVersion)
			return nil
		}
		return err
	}

	logger.Info("Migrations done to target version", "targetVersion", targetVersion)
	return nil
}

func HasCoreSchemaTables(db repository.PostgresRepository) (bool, error) {
	tableNames, err := bootstrapRegistrationTableNames()
	if err != nil {
		return false, err
	}

	var present int64
	err = db.Raw(
		`SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_name IN ?`,
		tableNames,
	).Scan(&present).Error
	if err != nil {
		return false, err
	}
	return present == int64(len(tableNames)), nil
}

func HasPendingJoinRequestUniqueIndex(db repository.PostgresRepository) (bool, error) {
	var indexes []pendingJoinRequestIndexContract
	err := db.Raw(
		`SELECT
			idx.indisunique AS is_unique,
			COALESCE(pg_get_expr(idx.indpred, idx.indrelid), '') AS predicate,
			COALESCE(string_agg(att.attname, ',' ORDER BY ord.ordinality), '') AS columns
		FROM pg_index idx
		JOIN pg_class tbl ON tbl.oid = idx.indrelid
		JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
		JOIN LATERAL unnest(idx.indkey) WITH ORDINALITY AS ord(attnum, ordinality) ON TRUE
		JOIN pg_attribute att ON att.attrelid = idx.indrelid AND att.attnum = ord.attnum
		WHERE ns.nspname = 'public'
		  AND tbl.relname = 'group_join_requests'
		GROUP BY idx.indexrelid, idx.indisunique, idx.indpred, idx.indrelid`,
	).Scan(&indexes).Error
	if err != nil {
		return false, err
	}

	for _, index := range indexes {
		if matchesPendingJoinRequestUniqueIndexContract(index) {
			return true, nil
		}
	}

	return false, nil
}

type pendingJoinRequestIndexContract struct {
	IsUnique  bool   `gorm:"column:is_unique"`
	Predicate string `gorm:"column:predicate"`
	Columns   string `gorm:"column:columns"`
}

var indexPredicateWhitespace = regexp.MustCompile(`\s+`)

func matchesPendingJoinRequestUniqueIndexContract(index pendingJoinRequestIndexContract) bool {
	return index.IsUnique &&
		hasPendingJoinRequestIndexColumns(index.Columns) &&
		hasPendingJoinRequestIndexPredicate(index.Predicate)
}

func hasPendingJoinRequestIndexColumns(columns string) bool {
	parts := strings.Split(columns, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return slices.Equal(parts, []string{"user_id", "group_id"})
}

func hasPendingJoinRequestIndexPredicate(predicate string) bool {
	normalized := normalizeIndexPredicate(predicate)
	return normalized == "status = 'pending'"
}

func normalizeIndexPredicate(predicate string) string {
	normalized := strings.ToLower(strings.TrimSpace(predicate))
	normalized = strings.ReplaceAll(normalized, `"`, "")
	normalized = strings.ReplaceAll(normalized, "::text", "")
	normalized = strings.ReplaceAll(normalized, "::character varying", "")
	normalized = strings.ReplaceAll(normalized, "=", " = ")
	normalized = indexPredicateWhitespace.ReplaceAllString(normalized, " ")

	for strings.HasPrefix(normalized, "(") && strings.HasSuffix(normalized, ")") {
		normalized = strings.TrimSpace(normalized[1 : len(normalized)-1])
	}

	return normalized
}

func bootstrapRegistrationTableNames() ([]string, error) {
	models := bootstrapRegistrationModels()
	tableNames := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))

	for _, model := range models {
		parsedSchema, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			return nil, fmt.Errorf("parse bootstrap schema for %T: %w", model, err)
		}
		if _, ok := seen[parsedSchema.Table]; ok {
			continue
		}
		seen[parsedSchema.Table] = struct{}{}
		tableNames = append(tableNames, parsedSchema.Table)
	}

	return tableNames, nil
}

func resolveMigrationSource() (string, error) {
	return resolveMigrationSourceWithMode(false)
}

func resolveMigrationSourceStrict() (string, error) {
	return resolveMigrationSourceWithMode(true)
}

func resolveMigrationSourceWithMode(strict bool) (string, error) {
	migrationDir, err := ResolveMigrationDirectory()
	if err != nil {
		return "", err
	}
	if migrationDir == "" {
		if strict {
			return "", errors.New("migration source is required for rollout phase but no SQL migration files were found")
		}
		return "", nil
	}

	hasSQL, err := hasSQLMigrationFiles(migrationDir)
	if err != nil {
		return "", fmt.Errorf("read migration dir %s: %w", migrationDir, err)
	}
	if !hasSQL {
		if strict {
			return "", fmt.Errorf("migration source is required for rollout phase but resolved directory %s does not contain SQL migration files", migrationDir)
		}
		return "", nil
	}

	return "file://" + filepath.ToSlash(migrationDir), nil
}

func ResolveMigrationDirectory() (string, error) {
	canonicalDir := filepath.Join(".", canonicalMigrationDir)
	canonicalHasAssets, err := hasMigrationAssets(canonicalDir)
	if err != nil {
		return "", fmt.Errorf("read migration dir %s: %w", canonicalDir, err)
	}

	legacyDirs, err := findLegacyMigrationDirectories()
	if err != nil {
		return "", err
	}

	if canonicalHasAssets {
		if len(legacyDirs) > 0 {
			return "", fmt.Errorf(
				"ambiguous migration discovery: canonical %s and legacy %s both contain migration assets; keep only %s",
				canonicalDir,
				strings.Join(legacyDirs, ", "),
				canonicalDir,
			)
		}
		absPath, err := filepath.Abs(canonicalDir)
		if err != nil {
			return "", fmt.Errorf("resolve migration dir %s: %w", canonicalDir, err)
		}
		return absPath, nil
	}

	if len(legacyDirs) > 1 {
		return "", fmt.Errorf(
			"ambiguous migration discovery across legacy directories %s; move SQL migrations and runbooks to %s",
			strings.Join(legacyDirs, ", "),
			canonicalDir,
		)
	}

	if len(legacyDirs) == 1 {
		absPath, err := filepath.Abs(legacyDirs[0])
		if err != nil {
			return "", fmt.Errorf("resolve migration dir %s: %w", legacyDirs[0], err)
		}
		return absPath, nil
	}

	return "", nil
}

func findLegacyMigrationDirectories() ([]string, error) {
	var dirs []string
	for _, dir := range legacyMigrationDirs {
		hasAssets, err := hasMigrationAssets(dir)
		if err != nil {
			return nil, fmt.Errorf("read migration dir %s: %w", dir, err)
		}
		if hasAssets {
			dirs = append(dirs, dir)
		}
	}

	return dirs, nil
}

func hasSQLMigrationFiles(dir string) (bool, error) {
	return directoryHasMatchingFiles(dir, func(name string) bool {
		return strings.HasSuffix(strings.ToLower(name), ".sql")
	})
}

func hasMigrationAssets(dir string) (bool, error) {
	return directoryHasMatchingFiles(dir, func(name string) bool {
		lowerName := strings.ToLower(name)
		return strings.HasSuffix(lowerName, ".sql") ||
			(strings.HasSuffix(lowerName, ".md") && strings.Contains(lowerName, "runbook"))
	})
}

func directoryHasMatchingFiles(dir string, match func(name string) bool) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if match(entry.Name()) {
			return true, nil
		}
	}

	return false, nil
}
