package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureMigrationSourceForPhaseRequiresSourceForUpAndAll(t *testing.T) {
	phases := []string{phaseUp, phaseAll}

	for _, phase := range phases {
		t.Run(phase, func(t *testing.T) {
			calls := 0
			wantErr := errors.New("missing source")
			err := ensureMigrationSourceForPhase(phase, func() (string, error) {
				calls++
				return "", wantErr
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
			if calls != 1 {
				t.Fatalf("requireSource calls = %d, want 1", calls)
			}
		})
	}
}

func TestEnsureMigrationSourceForPhaseSkipsCheckForPreAndPostflight(t *testing.T) {
	phases := []string{phasePreflight, phasePostflight}

	for _, phase := range phases {
		t.Run(phase, func(t *testing.T) {
			calls := 0
			err := ensureMigrationSourceForPhase(phase, func() (string, error) {
				calls++
				return "", errors.New("should not be called")
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if calls != 0 {
				t.Fatalf("requireSource calls = %d, want 0", calls)
			}
		})
	}
}

func TestResolveRolloutFilesUsesExplicitBase(t *testing.T) {
	migrationDir := t.TempDir()
	base := "000002_group_join_request_pending_uniqueness"
	writeRolloutFile(t, migrationDir, base+".up.sql")
	writeRolloutFile(t, migrationDir, base+"_preflight.sql")
	writeRolloutFile(t, migrationDir, base+"_postflight.sql")

	rollout, err := resolveRolloutFiles(base, func() (string, error) {
		return migrationDir, nil
	})
	if err != nil {
		t.Fatalf("resolveRolloutFiles returned error: %v", err)
	}
	if rollout.base != base {
		t.Fatalf("rollout.base = %q, want %q", rollout.base, base)
	}
	if rollout.upPath != filepath.Join(migrationDir, base+".up.sql") {
		t.Fatalf("unexpected up path: %s", rollout.upPath)
	}
}

func TestResolveRolloutFilesUsesLatestBaseWhenExplicitBaseMissing(t *testing.T) {
	migrationDir := t.TempDir()
	writeRolloutFile(t, migrationDir, "000001_membership_uniqueness.up.sql")
	writeRolloutFile(t, migrationDir, "000002_group_join_request_pending_uniqueness.up.sql")

	rollout, err := resolveRolloutFiles("", func() (string, error) {
		return migrationDir, nil
	})
	if err != nil {
		t.Fatalf("resolveRolloutFiles returned error: %v", err)
	}
	if rollout.base != "000002_group_join_request_pending_uniqueness" {
		t.Fatalf("rollout.base = %q, want latest base", rollout.base)
	}
}

func TestResolveRolloutFilesFailsWhenNoUpMigrationsExist(t *testing.T) {
	migrationDir := t.TempDir()
	writeRolloutFile(t, migrationDir, "notes.txt")

	_, err := resolveRolloutFiles("", func() (string, error) {
		return migrationDir, nil
	})
	if err == nil {
		t.Fatal("resolveRolloutFiles returned nil error, want no migrations error")
	}
}

func TestEnsureRolloutFilesForPhaseRequiresPhaseCompanions(t *testing.T) {
	migrationDir := t.TempDir()
	base := "000002_group_join_request_pending_uniqueness"
	writeRolloutFile(t, migrationDir, base+".up.sql")

	err := ensureRolloutFilesForPhase(phaseAll, rolloutFiles{
		base:           base,
		migrationDir:   migrationDir,
		upPath:         filepath.Join(migrationDir, base+".up.sql"),
		preflightPath:  filepath.Join(migrationDir, base+"_preflight.sql"),
		postflightPath: filepath.Join(migrationDir, base+"_postflight.sql"),
	})
	if err == nil {
		t.Fatal("ensureRolloutFilesForPhase returned nil error, want missing pre/postflight error")
	}
}

func TestMigrationVersionFromBaseParsesVersionPrefix(t *testing.T) {
	version, err := migrationVersionFromBase("000002_group_join_request_pending_uniqueness")
	if err != nil {
		t.Fatalf("migrationVersionFromBase returned error: %v", err)
	}
	if version != 2 {
		t.Fatalf("version = %d, want 2", version)
	}
}

func TestMigrationVersionFromBaseRejectsInvalidValue(t *testing.T) {
	_, err := migrationVersionFromBase("not-a-rollout")
	if err == nil {
		t.Fatal("migrationVersionFromBase returned nil error, want invalid prefix error")
	}
}

func TestRunMigrationUpForRolloutUsesParsedTargetVersion(t *testing.T) {
	called := 0
	var gotVersion uint

	err := runMigrationUpForRollout("000002_group_join_request_pending_uniqueness", func(version uint) error {
		called++
		gotVersion = version
		return nil
	})
	if err != nil {
		t.Fatalf("runMigrationUpForRollout returned error: %v", err)
	}
	if called != 1 {
		t.Fatalf("migrateToVersion calls = %d, want 1", called)
	}
	if gotVersion != 2 {
		t.Fatalf("target version = %d, want 2", gotVersion)
	}
}

func writeRolloutFile(t *testing.T, dir string, name string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write rollout file %s: %v", name, err)
	}
}
