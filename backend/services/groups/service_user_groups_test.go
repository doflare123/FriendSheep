package group

import (
	"errors"
	"friendship/models/groups"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestGroupAdminStoreUpdateGroupRejectsRoleWithoutModerateCapability(t *testing.T) {
	store := gormGroupAdminStore{}

	_, err := store.UpdateGroup(GroupUpdateInput{GroupID: 1}, joinRequestActor{ID: 2}, groups.RoleMember)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
}

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

func TestIsGroupMembershipUniqueViolationMatchesPostgresConstraint(t *testing.T) {
	if !isGroupMembershipUniqueViolation(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "idx_group_user_membership",
	}) {
		t.Fatal("expected helper to match group membership unique violation")
	}
}

func TestIsGroupMembershipUniqueViolationMatchesSQLiteFallback(t *testing.T) {
	err := errors.New("UNIQUE constraint failed: group_users.user_id, group_users.group_id")
	if !isGroupMembershipUniqueViolation(err) {
		t.Fatal("expected helper to match sqlite group membership unique violation text")
	}
}

func TestIsGroupMembershipUniqueViolationRejectsOtherErrors(t *testing.T) {
	if isGroupMembershipUniqueViolation(errors.New("boom")) {
		t.Fatal("expected helper to reject unrelated error")
	}
}
