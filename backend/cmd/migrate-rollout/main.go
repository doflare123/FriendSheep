package main

import (
	"fmt"
	"friendship/config"
	"friendship/db"
	"friendship/logger"
	"friendship/repository"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const (
	phaseAll        = "all"
	phasePreflight  = "preflight"
	phaseUp         = "up"
	phasePostflight = "postflight"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env not loaded, relying on environment variables")
	}

	phase := phaseAll
	if len(os.Args) > 1 {
		phase = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}
	if phase != phaseAll && phase != phasePreflight && phase != phaseUp && phase != phasePostflight {
		log.Fatalf("unknown phase %q. expected one of: %s, %s, %s, %s", phase, phaseAll, phasePreflight, phaseUp, phasePostflight)
	}
	if err := ensureMigrationSourceForPhase(phase, db.RequireMigrationSource); err != nil {
		log.Fatalf("strict rollout contract failed: %v", err)
	}

	cfg := config.NewConfig()
	logr, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("logger init error: %v", err)
	}
	pg := repository.NewSheepRepository(logr, cfg)
	defer func() {
		if err := pg.Close(); err != nil {
			logr.Warn("failed to close postgres connection", "error", err)
		}
	}()

	switch phase {
	case phasePreflight:
		runSQLFileOrDie(pg, "migration/000001_membership_uniqueness_preflight.sql")
	case phaseUp:
		if err := db.MigrationDBStrict(pg, logr); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
	case phasePostflight:
		runSQLFileOrDie(pg, "migration/000001_membership_uniqueness_postflight.sql")
	default:
		runSQLFileOrDie(pg, "migration/000001_membership_uniqueness_preflight.sql")
		if err := db.MigrationDBStrict(pg, logr); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
		runSQLFileOrDie(pg, "migration/000001_membership_uniqueness_postflight.sql")
	}

	logr.Info("migration rollout phase completed", "phase", phase)
}

func ensureMigrationSourceForPhase(phase string, requireSource func() (string, error)) error {
	if phase != phaseUp && phase != phaseAll {
		return nil
	}
	_, err := requireSource()
	return err
}

func runSQLFileOrDie(pg repository.PostgresRepository, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s failed: %v", path, err)
	}
	sqlText := strings.TrimSpace(string(content))
	if sqlText == "" {
		log.Fatalf("sql file %s is empty", path)
	}
	if err := pg.Exec(sqlText).Error; err != nil {
		log.Fatalf("execute %s failed: %v", filepath.ToSlash(path), err)
	}
	fmt.Printf("ok: %s\n", filepath.ToSlash(path))
}
