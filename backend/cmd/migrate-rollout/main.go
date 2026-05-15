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
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	phaseAll        = "all"
	phasePreflight  = "preflight"
	phaseUp         = "up"
	phasePostflight = "postflight"
)

type rolloutFiles struct {
	base           string
	migrationDir   string
	upPath         string
	preflightPath  string
	postflightPath string
}

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

	rollout, err := resolveRolloutFiles(rolloutBaseArg(os.Args), db.ResolveMigrationDirectory)
	if err != nil {
		log.Fatalf("resolve rollout files failed: %v", err)
	}
	if err := ensureRolloutFilesForPhase(phase, rollout); err != nil {
		log.Fatalf("rollout file contract failed: %v", err)
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
		runSQLFileOrDie(pg, rollout.preflightPath)
	case phaseUp:
		if err := runMigrationUpForRollout(rollout.base, func(version uint) error {
			return db.MigrationDBStrictToVersion(pg, logr, version)
		}); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
	case phasePostflight:
		runSQLFileOrDie(pg, rollout.postflightPath)
	default:
		runSQLFileOrDie(pg, rollout.preflightPath)
		if err := runMigrationUpForRollout(rollout.base, func(version uint) error {
			return db.MigrationDBStrictToVersion(pg, logr, version)
		}); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
		runSQLFileOrDie(pg, rollout.postflightPath)
	}

	logr.Info("migration rollout phase completed", "phase", phase, "rolloutBase", rollout.base)
}

func ensureMigrationSourceForPhase(phase string, requireSource func() (string, error)) error {
	if phase != phaseUp && phase != phaseAll {
		return nil
	}
	_, err := requireSource()
	return err
}

func rolloutBaseArg(args []string) string {
	if len(args) < 3 {
		return ""
	}
	return strings.TrimSpace(args[2])
}

func runMigrationUpForRollout(rolloutBase string, migrateToVersion func(version uint) error) error {
	targetVersion, err := migrationVersionFromBase(rolloutBase)
	if err != nil {
		return err
	}
	return migrateToVersion(targetVersion)
}

func migrationVersionFromBase(rolloutBase string) (uint, error) {
	base := strings.TrimSpace(rolloutBase)
	if base == "" {
		return 0, fmt.Errorf("rollout base is empty")
	}

	prefix := base
	if underscore := strings.Index(prefix, "_"); underscore >= 0 {
		prefix = prefix[:underscore]
	}
	if prefix == "" {
		return 0, fmt.Errorf("rollout base %q has empty version prefix", rolloutBase)
	}

	version, err := strconv.ParseUint(prefix, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("rollout base %q has invalid numeric version prefix: %w", rolloutBase, err)
	}
	if version == 0 {
		return 0, fmt.Errorf("rollout base %q has invalid zero version prefix", rolloutBase)
	}

	return uint(version), nil
}

func resolveRolloutFiles(base string, resolveMigrationDir func() (string, error)) (rolloutFiles, error) {
	migrationDir, err := resolveMigrationDir()
	if err != nil {
		return rolloutFiles{}, err
	}
	if migrationDir == "" {
		return rolloutFiles{}, fmt.Errorf("migration directory is not available")
	}

	rolloutBase := strings.TrimSpace(base)
	if rolloutBase == "" {
		rolloutBase, err = latestMigrationBase(migrationDir)
		if err != nil {
			return rolloutFiles{}, err
		}
	}

	return rolloutFiles{
		base:           rolloutBase,
		migrationDir:   migrationDir,
		upPath:         filepath.Join(migrationDir, rolloutBase+".up.sql"),
		preflightPath:  filepath.Join(migrationDir, rolloutBase+"_preflight.sql"),
		postflightPath: filepath.Join(migrationDir, rolloutBase+"_postflight.sql"),
	}, nil
}

func latestMigrationBase(migrationDir string) (string, error) {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return "", err
	}

	var bases []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		bases = append(bases, strings.TrimSuffix(name, ".up.sql"))
	}

	if len(bases) == 0 {
		return "", fmt.Errorf("no .up.sql migrations found in %s", migrationDir)
	}

	sort.Strings(bases)
	return bases[len(bases)-1], nil
}

func ensureRolloutFilesForPhase(phase string, rollout rolloutFiles) error {
	if err := requireRolloutFile(rollout.upPath, "up"); err != nil {
		return err
	}

	switch phase {
	case phasePreflight, phaseAll:
		if err := requireRolloutFile(rollout.preflightPath, "preflight"); err != nil {
			return err
		}
	}

	switch phase {
	case phasePostflight, phaseAll:
		if err := requireRolloutFile(rollout.postflightPath, "postflight"); err != nil {
			return err
		}
	}

	return nil
}

func requireRolloutFile(path string, kind string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s rollout file %s is not available: %w", kind, filepath.ToSlash(path), err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s rollout path %s points to a directory", kind, filepath.ToSlash(path))
	}
	return nil
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
