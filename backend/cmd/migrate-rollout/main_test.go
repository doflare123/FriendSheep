package main

import (
	"errors"
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
