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
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func AutoMigDB(db repository.PostgresRepository, models ...interface{}) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}

func BootstrapRegistrationSchema(db repository.PostgresRepository) error {
	return AutoMigDB(
		db,
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
	)
}

func HasMigrationSource() (bool, error) {
	migrationsPath, err := resolveMigrationSource()
	if err != nil {
		return false, err
	}
	return migrationsPath != "", nil
}

func MigrationDB(db repository.PostgresRepository, logger logger.Logger) error {
	sqlDB, err := db.GetSQLDB()
	if err != nil {
		return err
	}
	migrationsPath, err := resolveMigrationSource()
	if err != nil {
		return err
	}
	if migrationsPath == "" {
		logger.Info("No SQL migrations found, skipping migrate bootstrap")
		return nil
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

func resolveMigrationSource() (string, error) {
	candidates := []string{
		filepath.Join(".", "migration"),
		filepath.Join(".", "backend", "migration"),
		filepath.Join(".", "migrations"),
		filepath.Join(".", "backend", "migrations"),
	}

	for _, candidate := range candidates {
		entries, err := os.ReadDir(candidate)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("read migration dir %s: %w", candidate, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
				continue
			}

			absPath, err := filepath.Abs(candidate)
			if err != nil {
				return "", fmt.Errorf("resolve migration dir %s: %w", candidate, err)
			}
			return "file://" + filepath.ToSlash(absPath), nil
		}
	}

	return "", nil
}
