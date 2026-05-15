package group

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsPendingJoinRequestUniqueViolationMatchesPostgresConstraint(t *testing.T) {
	if !isPendingJoinRequestUniqueViolation(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "idx_group_join_request_pending_unique",
	}) {
		t.Fatal("expected helper to match pending join request unique violation")
	}
}

func TestIsPendingJoinRequestUniqueViolationMatchesSQLiteFallback(t *testing.T) {
	err := errors.New("UNIQUE constraint failed: group_join_requests.user_id, group_join_requests.group_id")
	if !isPendingJoinRequestUniqueViolation(err) {
		t.Fatal("expected helper to match sqlite unique violation text")
	}
}

func TestIsPendingJoinRequestUniqueViolationRejectsOtherErrors(t *testing.T) {
	if isPendingJoinRequestUniqueViolation(errors.New("boom")) {
		t.Fatal("expected helper to reject unrelated error")
	}
}
