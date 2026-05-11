package server

import (
	"errors"
	"strings"
	"testing"
)

func TestDiscoverStartupMigrationAssetsSkipsDiscoveryInDEV(t *testing.T) {
	calls := 0

	got, err := discoverStartupMigrationAssets("DEV", func() (bool, error) {
		calls++
		return false, errors.New("should not be called")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatal("discoverStartupMigrationAssets returned true in DEV, want false")
	}
	if calls != 0 {
		t.Fatalf("discover function calls = %d, want 0", calls)
	}
}

func TestDiscoverStartupMigrationAssetsCallsDiscoveryOutsideDEV(t *testing.T) {
	calls := 0

	got, err := discoverStartupMigrationAssets("PROD", func() (bool, error) {
		calls++
		return true, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("discoverStartupMigrationAssets returned false outside DEV, want true")
	}
	if calls != 1 {
		t.Fatalf("discover function calls = %d, want 1", calls)
	}
}

func TestValidateNonDevStartupMigrationAssets(t *testing.T) {
	tests := []struct {
		name             string
		hasSQLMigrations bool
		wantErr          string
	}{
		{
			name:             "migration assets present",
			hasSQLMigrations: true,
			wantErr:          "",
		},
		{
			name:             "migration assets missing must fail fast",
			hasSQLMigrations: false,
			wantErr:          "startup aborted: SQL migration assets not found in non-DEV environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNonDevStartupMigrationAssets(tt.hasSQLMigrations)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want contains %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestValidateNonDevStartupWithSQLMigrations(t *testing.T) {
	tests := []struct {
		name                     string
		startupMigrationsEnabled bool
		hasCoreSchema            bool
		wantErr                  string
	}{
		{
			name:                     "migration enabled and schema complete",
			startupMigrationsEnabled: true,
			hasCoreSchema:            true,
			wantErr:                  "",
		},
		{
			name:                     "migration enabled and schema incomplete fail fast",
			startupMigrationsEnabled: true,
			hasCoreSchema:            false,
			wantErr:                  "core schema is incomplete after SQL migrations",
		},
		{
			name:                     "migration disabled and schema complete",
			startupMigrationsEnabled: false,
			hasCoreSchema:            true,
			wantErr:                  "",
		},
		{
			name:                     "migration disabled and schema incomplete must fail",
			startupMigrationsEnabled: false,
			hasCoreSchema:            false,
			wantErr:                  "startup SQL migrations are disabled and core schema is incomplete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNonDevStartupWithSQLMigrations(tt.startupMigrationsEnabled, tt.hasCoreSchema)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want contains %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}
