package db

import (
	"friendship/logger"
	"friendship/repository"

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

func MigrationDB(db repository.PostgresRepository, logger logger.Logger) error {
	sqlDB, err := db.GetSQLDB()
	if err != nil {
		return err
	}
	migrationsPath := "./backend/migrations"
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
