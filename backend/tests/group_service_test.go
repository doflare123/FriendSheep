package tests

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"friendship/models"
	groupmodels "friendship/models/groups"
	"friendship/repository"
	"friendship/services"
	servicegroups "friendship/services/groups"

	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestGroupServiceJoinGroupAddsMemberForPublicGroup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	result, err := service.JoinGroup(2, groupID)

	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
}

func TestGroupServiceJoinGroupRejectsDuplicatePublicMembership(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrAlreadyInGroup) {
		t.Fatalf("err = %v, want ErrAlreadyInGroup", err)
	}
	assertGroupMembershipCount(t, db, groupID, 2, 1)
}

func TestGroupServiceJoinGroupCreatesRequestForPrivateGroup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)

	result, err := service.JoinGroup(2, groupID)

	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || result.Joined {
		t.Fatalf("result = %#v, want pending request result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 1)
}

func TestGroupServiceJoinGroupRejectsBlacklistedUserWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupBlacklist(t, db, groupID, 2, 1)

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrUserInBlacklist) {
		t.Fatalf("err = %v, want ErrUserInBlacklist", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
}

func TestGroupServiceJoinGroupRejectsDuplicatePendingRequest(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupJoinRequest(t, db, groupID, 2, "pending")

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrRequestAlreadyExists) {
		t.Fatalf("err = %v, want ErrRequestAlreadyExists", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 1)
}

func TestGroupServiceJoinGroupAllowsNewPendingAfterRejectedRequest(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupJoinRequest(t, db, groupID, 2, "rejected")
	createPendingJoinRequestUniqueIndex(t, db)

	result, err := service.JoinGroup(2, groupID)

	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || result.Joined {
		t.Fatalf("result = %#v, want pending request result", result)
	}
	assertGroupJoinRequestCount(t, db, groupID, 2, "rejected", 1)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 1)
}

func TestGroupServiceJoinGroupMapsPendingUniqueViolationToDuplicateRequest(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &duplicatePendingRequestRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	createPendingJoinRequestUniqueIndex(t, db)

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrRequestAlreadyExists) {
		t.Fatalf("err = %v, want ErrRequestAlreadyExists", err)
	}
	if !repo.injected.Load() {
		t.Fatal("test unique violation injection was not reached")
	}
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
}

func TestGroupServiceJoinGroupMapsMembershipUniqueViolationToAlreadyInGroup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &duplicateMembershipRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	result, err := service.JoinGroup(2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrAlreadyInGroup) {
		t.Fatalf("err = %v, want ErrAlreadyInGroup", err)
	}
	if !repo.injected.Load() {
		t.Fatal("test membership unique violation injection was not reached")
	}
	assertGroupMembershipCount(t, db, groupID, 2, 0)
}

func TestGroupServiceApproveJoinRequestAddsMembershipAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(1, requestID)

	if err != nil {
		t.Fatalf("ApproveJoinRequest returned error: %v", err)
	}
	if !approved {
		t.Fatal("ApproveJoinRequest returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
	assertGroupJoinRequestStatus(t, db, requestID, "approved")
	assertGroupServiceActionLogCount(t, db, groupID, "approve_request", 1)
}

func TestGroupServiceApproveJoinRequestUsesTransactionForActorRoleLookup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &rootGroupAccessPoisonRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(1, requestID)

	if err != nil {
		t.Fatalf("ApproveJoinRequest returned error: %v", err)
	}
	if !approved {
		t.Fatal("ApproveJoinRequest returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
	assertGroupJoinRequestStatus(t, db, requestID, "approved")
}

func TestGroupServiceApproveJoinRequestRejectsMemberActorWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, memberRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(1, requestID)

	if approved {
		t.Fatal("ApproveJoinRequest returned true")
	}
	if !errors.Is(err, servicegroups.ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestStatus(t, db, requestID, "pending")
	assertGroupServiceActionLogCount(t, db, groupID, "approve_request", 0)
}

func TestGroupServiceApproveJoinRequestMapsMembershipUniqueViolationToAlreadyInGroup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &duplicateMembershipRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(1, requestID)

	if approved {
		t.Fatal("ApproveJoinRequest returned true")
	}
	if !errors.Is(err, servicegroups.ErrAlreadyInGroup) {
		t.Fatalf("err = %v, want ErrAlreadyInGroup", err)
	}
	if !repo.injected.Load() {
		t.Fatal("test membership unique violation injection was not reached")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestStatus(t, db, requestID, "pending")
	assertGroupServiceActionLogCount(t, db, groupID, "approve_request", 0)
}

func TestGroupServiceApproveJoinRequestRejectsBlacklistedUserWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupBlacklist(t, db, groupID, 2, 1)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(1, requestID)

	if approved {
		t.Fatal("ApproveJoinRequest returned true")
	}
	if !errors.Is(err, servicegroups.ErrUserInBlacklist) {
		t.Fatalf("err = %v, want ErrUserInBlacklist", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestStatus(t, db, requestID, "pending")
	assertGroupServiceActionLogCount(t, db, groupID, "approve_request", 0)
}

func TestGroupServiceRejectJoinRequestUpdatesStatusAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	operatorRoleID := seedGroupServiceRole(t, db, "Модератор")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinRequest(1, requestID)

	if err != nil {
		t.Fatalf("RejectJoinRequest returned error: %v", err)
	}
	if !rejected {
		t.Fatal("RejectJoinRequest returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestStatus(t, db, requestID, "rejected")
	assertGroupServiceActionLogCount(t, db, groupID, "reject_request", 1)
}

func TestGroupServiceRejectJoinRequestRejectsMemberActorWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, memberRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinRequest(1, requestID)

	if rejected {
		t.Fatal("RejectJoinRequest returned true")
	}
	if !errors.Is(err, servicegroups.ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestStatus(t, db, requestID, "pending")
	assertGroupServiceActionLogCount(t, db, groupID, "reject_request", 0)
}

func TestGroupServiceRejectJoinRequestUsesTransactionForActorRoleLookup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &rootGroupAccessPoisonRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	operatorRoleID := seedGroupServiceRole(t, db, groupmodels.RoleModerator)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinRequest(1, requestID)

	if err != nil {
		t.Fatalf("RejectJoinRequest returned error: %v", err)
	}
	if !rejected {
		t.Fatal("RejectJoinRequest returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestStatus(t, db, requestID, "rejected")
}

func TestGroupServiceCreateJoinInviteLoadsTargetUserAndRejectsMissingUser(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  404,
	})

	if created {
		t.Fatal("CreateJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
	assertGroupJoinInviteCount(t, db, groupID, 404, "pending", 0)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 0)
}

func TestGroupServiceCreateJoinInviteUsesTransactionForActorRoleLookup(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &rootGroupAccessPoisonRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	operatorRoleID := seedGroupServiceRole(t, db, groupmodels.RoleModerator)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if err != nil {
		t.Fatalf("CreateJoinInvite returned error: %v", err)
	}
	if !created {
		t.Fatal("CreateJoinInvite returned false")
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 1)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 1)
}

func TestGroupServiceCreateJoinInviteRejectsMemberActorWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, memberRoleID)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if created {
		t.Fatal("CreateJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 0)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 0)
}

func TestGroupServiceCreateJoinInviteRejectsMissingActorMembershipWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if created {
		t.Fatal("CreateJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrNotInGroup) {
		t.Fatalf("err = %v, want ErrNotInGroup", err)
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 0)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 0)
}

func TestGroupServiceCreateJoinInviteRejectsExistingMemberWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if created {
		t.Fatal("CreateJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrAlreadyInGroup) {
		t.Fatalf("err = %v, want ErrAlreadyInGroup", err)
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 0)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 0)
}

func TestGroupServiceCreateJoinInviteRejectsDuplicatePendingInviteWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if created {
		t.Fatal("CreateJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrInviteAlreadyExists) {
		t.Fatalf("err = %v, want ErrInviteAlreadyExists", err)
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 1)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 0)
}

func TestGroupServiceCreateJoinInviteCreatesPendingInviteAndActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if err != nil {
		t.Fatalf("CreateJoinInvite returned error: %v", err)
	}
	if !created {
		t.Fatal("CreateJoinInvite returned false")
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 1)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 1)
	assertGroupServiceActionLogContains(t, db, groupID, "send_invite", "group-user-2")
}

func TestGroupServiceCreateJoinInviteKeepsInviteWhenActionLogFails(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &actionLogFailureRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)

	created, err := service.CreateJoinInvite(1, servicegroups.JoinInviteInput{
		GroupID: groupID,
		UserID:  2,
	})

	if err != nil {
		t.Fatalf("CreateJoinInvite returned error: %v", err)
	}
	if !created {
		t.Fatal("CreateJoinInvite returned false")
	}
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 1)
	assertGroupServiceActionLogCount(t, db, groupID, "send_invite", 0)
}

func TestGroupServiceCreateGroupRejectsMissingCategoriesWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	seedGroupCategory(t, db, 1, "Board Games")
	isPrivate := false
	categoryID := uint(1)
	missingCategoryID := uint(999)

	groupDTO, err := service.CreateGroup(1, servicegroups.CreateGroupInput{
		Name:             "Board Game Club",
		Description:      "Group for board game fans",
		SmallDescription: "Play together",
		Image:            "https://example.com/group.png",
		IsPrivate:        &isPrivate,
		Categories:       []*uint{&categoryID, &missingCategoryID},
		Contacts:         "tg:https://t.me/group",
	})

	if groupDTO != nil {
		t.Fatalf("groupDTO = %#v, want nil", groupDTO)
	}
	if !errors.Is(err, servicegroups.ErrCategoriesNotFound) {
		t.Fatalf("err = %v, want ErrCategoriesNotFound", err)
	}
	assertGroupTotal(t, db, 0)
	assertGroupMembershipTotal(t, db, 0)
	assertGroupContactTotal(t, db, 0)
	assertGroupServiceActionLogTotal(t, db, 0)
}

func TestGroupServiceCreateGroupReturnsFullDetailsForCreator(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	setGroupServiceUserImage(t, db, 1, "https://cdn.example.com/avatar.png")
	seedGroupCategory(t, db, 1, "Board Games")
	isPrivate := true
	categoryID := uint(1)

	groupDTO, err := service.CreateGroup(1, servicegroups.CreateGroupInput{
		Name:             "Board Game Club",
		Description:      "Group for board game fans",
		SmallDescription: "Play together",
		Image:            "https://example.com/group.png",
		IsPrivate:        &isPrivate,
		Categories:       []*uint{&categoryID},
		Contacts:         "tg:https://t.me/group",
	})

	if err != nil {
		t.Fatalf("CreateGroup returned error: %v", err)
	}
	if groupDTO == nil {
		t.Fatal("groupDTO is nil")
	}
	if groupDTO.Creator.Image != "https://cdn.example.com/avatar.png" {
		t.Fatalf("creator image = %q, want avatar url", groupDTO.Creator.Image)
	}
	if groupDTO.MemberCount != 1 {
		t.Fatalf("member count = %d, want 1", groupDTO.MemberCount)
	}
	if len(groupDTO.Members) != 1 {
		t.Fatalf("members len = %d, want 1", len(groupDTO.Members))
	}
	if groupDTO.Members[0].ID != 1 {
		t.Fatalf("member id = %d, want 1", groupDTO.Members[0].ID)
	}
	if groupDTO.Members[0].Role != groupmodels.RoleAdmin {
		t.Fatalf("member role = %q, want %q", groupDTO.Members[0].Role, groupmodels.RoleAdmin)
	}
	if !groupDTO.IsSubscribed {
		t.Fatal("isSubscribed = false, want true")
	}
	if groupDTO.UserRole != groupmodels.RoleAdmin {
		t.Fatalf("user role = %q, want %q", groupDTO.UserRole, groupmodels.RoleAdmin)
	}
	assertGroupActionActorFields(t, db, groupDTO.ID, "create_group", "Group User", "group-user-1")
}

func TestGroupServiceUpdateGroupReturnsFullDetailsForActor(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	setGroupServiceUserImage(t, db, 1, "https://cdn.example.com/avatar.png")
	seedGroupCategory(t, db, 1, "Board Games")
	isPrivate := true
	categoryID := uint(1)

	created, err := service.CreateGroup(1, servicegroups.CreateGroupInput{
		Name:             "Board Game Club",
		Description:      "Group for board game fans",
		SmallDescription: "Play together",
		Image:            "https://example.com/group.png",
		IsPrivate:        &isPrivate,
		Categories:       []*uint{&categoryID},
		Contacts:         "tg:https://t.me/group",
	})
	if err != nil {
		t.Fatalf("CreateGroup returned error: %v", err)
	}

	newName := "Board Game Club Updated"
	newContacts := "tg:https://t.me/newgroup"
	updated, err := service.UpdateGroup(1, servicegroups.GroupUpdateInput{
		GroupID:  created.ID,
		Name:     &newName,
		Contacts: &newContacts,
	})

	if err != nil {
		t.Fatalf("UpdateGroup returned error: %v", err)
	}
	if updated == nil {
		t.Fatal("updated dto is nil")
	}
	if updated.Name != newName {
		t.Fatalf("name = %q, want %q", updated.Name, newName)
	}
	if updated.Creator.Image != "https://cdn.example.com/avatar.png" {
		t.Fatalf("creator image = %q, want avatar url", updated.Creator.Image)
	}
	if updated.MemberCount != 1 {
		t.Fatalf("member count = %d, want 1", updated.MemberCount)
	}
	if len(updated.Members) != 1 {
		t.Fatalf("members len = %d, want 1", len(updated.Members))
	}
	if updated.Members[0].ID != 1 {
		t.Fatalf("member id = %d, want 1", updated.Members[0].ID)
	}
	if updated.Members[0].Role != groupmodels.RoleAdmin {
		t.Fatalf("member role = %q, want %q", updated.Members[0].Role, groupmodels.RoleAdmin)
	}
	if !updated.IsSubscribed {
		t.Fatal("isSubscribed = false, want true")
	}
	if updated.UserRole != groupmodels.RoleAdmin {
		t.Fatalf("user role = %q, want %q", updated.UserRole, groupmodels.RoleAdmin)
	}
}

func TestGroupServiceAddPermissionsPromotesMemberAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceRole(t, db, "Модератор")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	changed, err := service.AddPermissions(1, servicegroups.PermissionInput{GroupID: groupID, UserID: 2})

	if err != nil {
		t.Fatalf("AddPermissions returned error: %v", err)
	}
	if !changed {
		t.Fatal("AddPermissions returned false")
	}
	assertGroupMembershipRole(t, db, groupID, 2, "Модератор")
	assertGroupServiceActionLogCount(t, db, groupID, "add_operator", 1)
}

func TestGroupServiceAddPermissionsRejectsModeratorActorWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	operatorRoleID := seedGroupServiceRole(t, db, "Модератор")
	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	changed, err := service.AddPermissions(1, servicegroups.PermissionInput{GroupID: groupID, UserID: 2})

	if changed {
		t.Fatal("AddPermissions returned true")
	}
	if !errors.Is(err, servicegroups.ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
	assertGroupMembershipRole(t, db, groupID, 2, "Участник")
	assertGroupServiceActionLogCount(t, db, groupID, "add_operator", 0)
}

func TestGroupServiceRemovePermissionsDemotesModeratorAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	operatorRoleID := seedGroupServiceRole(t, db, "Модератор")
	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, operatorRoleID)

	changed, err := service.RemovePermissions(1, servicegroups.PermissionInput{GroupID: groupID, UserID: 2})

	if err != nil {
		t.Fatalf("RemovePermissions returned error: %v", err)
	}
	if !changed {
		t.Fatal("RemovePermissions returned false")
	}
	assertGroupMembershipRole(t, db, groupID, 2, "Участник")
	assertGroupServiceActionLogCount(t, db, groupID, "remove_operator", 1)
}

func TestGroupServiceDeleteUserFromGroupMovesMemberToBlacklistAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	deleted, err := service.DeleteUserFromGroup(1, groupID, 2)

	if err != nil {
		t.Fatalf("DeleteUserFromGroup returned error: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteUserFromGroup returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupBlacklistExists(t, db, groupID, 2, true)
	assertGroupServiceActionLogCount(t, db, groupID, "ban_user", 1)
	assertGroupServiceActionLogContains(t, db, groupID, "ban_user", "group-user-2")
}

func TestGroupServiceRemoveFromBlacklistDeletesEntryAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupBlacklist(t, db, groupID, 2, 1)

	removed, err := service.RemoveFromBlacklist(1, groupID, 2)

	if err != nil {
		t.Fatalf("RemoveFromBlacklist returned error: %v", err)
	}
	if !removed {
		t.Fatal("RemoveFromBlacklist returned false")
	}
	assertGroupBlacklistExists(t, db, groupID, 2, false)
	assertGroupServiceActionLogCount(t, db, groupID, "unban_user", 1)
	assertGroupServiceActionLogContains(t, db, groupID, "unban_user", "group-user-2")
}

func TestGroupServiceAcceptJoinInviteAddsMembershipAndUpdatesStatus(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(2, inviteID)

	if err != nil {
		t.Fatalf("AcceptJoinInvite returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
	assertGroupJoinInviteStatus(t, db, inviteID, "accepted")
}

func TestGroupServiceAcceptJoinInviteIdempotentForExistingMembershipAndPendingInvite(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(2, inviteID)

	if err != nil {
		t.Fatalf("AcceptJoinInvite returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipCount(t, db, groupID, 2, 1)
	assertGroupJoinInviteCount(t, db, groupID, 2, "pending", 0)
	assertGroupJoinInviteCount(t, db, groupID, 2, "accepted", 1)
	assertGroupJoinInviteStatus(t, db, inviteID, "accepted")
}

func TestGroupServiceAcceptJoinInviteIdempotentForAcceptedInvite(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "accepted")

	result, err := service.AcceptJoinInvite(2, inviteID)

	if err != nil {
		t.Fatalf("AcceptJoinInvite returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipCount(t, db, groupID, 2, 1)
	assertGroupJoinInviteStatus(t, db, inviteID, "accepted")
}

func TestGroupServiceRejectJoinInviteUpdatesStatusWithoutMembership(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, repo)

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinInvite(2, inviteID)

	if err != nil {
		t.Fatalf("RejectJoinInvite returned error: %v", err)
	}
	if !rejected {
		t.Fatal("RejectJoinInvite returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "rejected")
}

func newGroupServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "-" + testUintString(uint(time.Now().UnixNano())) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&groupmodels.Group{},
		&groupmodels.GroupGroupCategory{},
		&groupmodels.GroupContact{},
		&groupmodels.Role_in_group{},
		&groupmodels.GroupUsers{},
		&groupmodels.GroupJoinRequest{},
		&groupmodels.GroupJoinInvite{},
		&groupmodels.GroupBlacklist{},
		&groupmodels.GroupActionLog{},
	); err != nil {
		t.Fatalf("auto migrate group service models: %v", err)
	}

	return db
}

func seedGroupServiceUser(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()

	user := models.User{
		ID:       userID,
		Name:     "Group User",
		Password: "Password123!",
		Us:       "group-user-" + testUintString(userID),
		Email:    "group-user-" + testUintString(userID) + "@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create group user %d: %v", userID, err)
	}
}

func setGroupServiceUserImage(t *testing.T, db *gorm.DB, userID uint, image string) {
	t.Helper()

	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("image", image).Error; err != nil {
		t.Fatalf("update user image for %d: %v", userID, err)
	}
}

func seedGroupServiceRole(t *testing.T, db *gorm.DB, roleName string) uint {
	t.Helper()

	role := groupmodels.Role_in_group{Name: roleName}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create group role %q: %v", roleName, err)
	}
	return role.Id
}

func seedGroupServiceGroup(t *testing.T, db *gorm.DB, creatorID uint, isPrivate bool) uint {
	t.Helper()

	group := groupmodels.Group{
		Name:             "Group Service Test",
		Description:      "Group Service Test Description",
		SmallDescription: "Group Service",
		Image:            "https://example.com/group.png",
		CreaterID:        creatorID,
		IsPrivate:        isPrivate,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group service group: %v", err)
	}
	return group.ID
}

func seedGroupCategory(t *testing.T, db *gorm.DB, categoryID uint, name string) {
	t.Helper()

	category := models.Category{ID: categoryID, Name: name}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create group category: %v", err)
	}
}

func seedGroupServiceMembership(t *testing.T, db *gorm.DB, groupID, userID, roleID uint) {
	t.Helper()

	membership := groupmodels.GroupUsers{
		GroupID:       groupID,
		UserID:        userID,
		RoleInGroupID: roleID,
	}
	if err := db.Create(&membership).Error; err != nil {
		t.Fatalf("create group membership: %v", err)
	}
}

func seedGroupBlacklist(t *testing.T, db *gorm.DB, groupID, userID, bannedBy uint) {
	t.Helper()

	entry := groupmodels.GroupBlacklist{
		GroupID:  groupID,
		UserID:   userID,
		BannedBy: bannedBy,
		Reason:   "test",
	}
	if err := db.Create(&entry).Error; err != nil {
		t.Fatalf("create group blacklist entry: %v", err)
	}
}

func seedGroupJoinRequest(t *testing.T, db *gorm.DB, groupID, userID uint, status string) {
	t.Helper()

	_ = seedGroupJoinRequestWithID(t, db, groupID, userID, status)
}

func seedGroupJoinRequestWithID(t *testing.T, db *gorm.DB, groupID, userID uint, status string) uint {
	t.Helper()

	request := groupmodels.GroupJoinRequest{
		GroupID: groupID,
		UserID:  userID,
		Status:  status,
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatalf("create group join request: %v", err)
	}
	return request.ID
}

func createPendingJoinRequestUniqueIndex(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec(
		"CREATE UNIQUE INDEX idx_group_join_request_pending_unique ON group_join_requests (user_id, group_id) WHERE status = 'pending'",
	).Error; err != nil {
		t.Fatalf("create pending join request unique index: %v", err)
	}
}

type duplicatePendingRequestRepository struct {
	*testPostgresRepository
	injected atomic.Bool
}

func (r *duplicatePendingRequestRepository) Transaction(fc func(tx repository.PostgresRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&duplicatePendingRequestTxRepository{
			testPostgresRepository: &testPostgresRepository{db: tx},
			injected:               &r.injected,
		})
	})
}

type duplicatePendingRequestTxRepository struct {
	*testPostgresRepository
	injected *atomic.Bool
}

func (r *duplicatePendingRequestTxRepository) Create(value interface{}) *gorm.DB {
	if _, ok := value.(*groupmodels.GroupJoinRequest); ok && r.injected.CompareAndSwap(false, true) {
		result := r.db.Session(&gorm.Session{})
		result.Error = errors.New("UNIQUE constraint failed: group_join_requests.user_id, group_join_requests.group_id")
		return result
	}
	return r.testPostgresRepository.Create(value)
}

type duplicateMembershipRepository struct {
	*testPostgresRepository
	injected atomic.Bool
}

func (r *duplicateMembershipRepository) Transaction(fc func(tx repository.PostgresRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&duplicateMembershipTxRepository{
			testPostgresRepository: &testPostgresRepository{db: tx},
			injected:               &r.injected,
		})
	})
}

type duplicateMembershipTxRepository struct {
	*testPostgresRepository
	injected *atomic.Bool
}

func (r *duplicateMembershipTxRepository) Create(value interface{}) *gorm.DB {
	if _, ok := value.(*groupmodels.GroupUsers); ok && r.injected.CompareAndSwap(false, true) {
		result := r.db.Session(&gorm.Session{})
		result.Error = errors.New("UNIQUE constraint failed: group_users.user_id, group_users.group_id")
		return result
	}
	return r.testPostgresRepository.Create(value)
}

type rootGroupAccessPoisonRepository struct {
	*testPostgresRepository
}

func (r *rootGroupAccessPoisonRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	result := r.db.Session(&gorm.Session{})
	result.Error = errors.New("root repository role lookup should not be used inside request review transaction")
	return result
}

type actionLogFailureRepository struct {
	*testPostgresRepository
	injected atomic.Bool
}

func (r *actionLogFailureRepository) Transaction(fc func(tx repository.PostgresRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&actionLogFailureTxRepository{
			testPostgresRepository: &testPostgresRepository{db: tx},
			injected:               &r.injected,
		})
	})
}

type actionLogFailureTxRepository struct {
	*testPostgresRepository
	injected *atomic.Bool
}

func (r *actionLogFailureTxRepository) Create(value interface{}) *gorm.DB {
	if _, ok := value.(*groupmodels.GroupActionLog); ok && r.injected.CompareAndSwap(false, true) {
		result := r.db.Session(&gorm.Session{})
		result.Error = errors.New("forced group action log failure")
		return result
	}
	return r.testPostgresRepository.Create(value)
}

func seedGroupJoinInviteWithID(t *testing.T, db *gorm.DB, groupID, userID uint, status string) uint {
	t.Helper()

	invite := groupmodels.GroupJoinInvite{
		GroupID: groupID,
		UserID:  userID,
		Status:  status,
	}
	if err := db.Create(&invite).Error; err != nil {
		t.Fatalf("create group join invite: %v", err)
	}
	return invite.ID
}

func assertGroupMembershipExists(t *testing.T, db *gorm.DB, groupID, userID uint, want bool) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupUsers{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		t.Fatalf("count group membership: %v", err)
	}
	if (count > 0) != want {
		t.Fatalf("membership exists = %v, want %v", count > 0, want)
	}
}

func assertGroupMembershipCount(t *testing.T, db *gorm.DB, groupID, userID uint, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupUsers{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		t.Fatalf("count group membership: %v", err)
	}
	if count != want {
		t.Fatalf("membership count = %d, want %d", count, want)
	}
}

func assertGroupTotal(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.Group{}).Count(&count).Error; err != nil {
		t.Fatalf("count groups: %v", err)
	}
	if count != want {
		t.Fatalf("group count = %d, want %d", count, want)
	}
}

func assertGroupMembershipTotal(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupUsers{}).Count(&count).Error; err != nil {
		t.Fatalf("count group memberships: %v", err)
	}
	if count != want {
		t.Fatalf("group membership count = %d, want %d", count, want)
	}
}

func assertGroupContactTotal(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupContact{}).Count(&count).Error; err != nil {
		t.Fatalf("count group contacts: %v", err)
	}
	if count != want {
		t.Fatalf("group contact count = %d, want %d", count, want)
	}
}

func assertGroupJoinRequestCount(t *testing.T, db *gorm.DB, groupID, userID uint, status string, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupJoinRequest{}).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, status).
		Count(&count).Error; err != nil {
		t.Fatalf("count group join requests: %v", err)
	}
	if count != want {
		t.Fatalf("join request count = %d, want %d", count, want)
	}
}

func assertGroupJoinRequestStatus(t *testing.T, db *gorm.DB, requestID uint, want string) {
	t.Helper()

	var request groupmodels.GroupJoinRequest
	if err := db.First(&request, requestID).Error; err != nil {
		t.Fatalf("find group join request: %v", err)
	}
	if request.Status != want {
		t.Fatalf("join request status = %q, want %q", request.Status, want)
	}
}

func assertGroupJoinInviteCount(t *testing.T, db *gorm.DB, groupID, userID uint, status string, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupJoinInvite{}).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, status).
		Count(&count).Error; err != nil {
		t.Fatalf("count group join invites: %v", err)
	}
	if count != want {
		t.Fatalf("join invite count = %d, want %d", count, want)
	}
}

func assertGroupJoinInviteStatus(t *testing.T, db *gorm.DB, inviteID uint, want string) {
	t.Helper()

	var invite groupmodels.GroupJoinInvite
	if err := db.First(&invite, inviteID).Error; err != nil {
		t.Fatalf("find group join invite: %v", err)
	}
	if invite.Status != want {
		t.Fatalf("join invite status = %q, want %q", invite.Status, want)
	}
}

func assertGroupMembershipRole(t *testing.T, db *gorm.DB, groupID, userID uint, wantRole string) {
	t.Helper()

	var membership groupmodels.GroupUsers
	if err := db.Preload("RoleInGroup").
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&membership).Error; err != nil {
		t.Fatalf("find group membership: %v", err)
	}
	if membership.RoleInGroup.Name != wantRole {
		t.Fatalf("membership role = %q, want %q", membership.RoleInGroup.Name, wantRole)
	}
}

func assertGroupBlacklistExists(t *testing.T, db *gorm.DB, groupID, userID uint, want bool) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupBlacklist{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		t.Fatalf("count group blacklist entries: %v", err)
	}
	if (count > 0) != want {
		t.Fatalf("blacklist entry exists = %v, want %v", count > 0, want)
	}
}

func assertGroupServiceActionLogCount(t *testing.T, db *gorm.DB, groupID uint, action string, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupActionLog{}).
		Where("group_id = ? AND action = ?", groupID, action).
		Count(&count).Error; err != nil {
		t.Fatalf("count group action logs: %v", err)
	}
	if count != want {
		t.Fatalf("group action log count = %d, want %d", count, want)
	}
}

func assertGroupServiceActionLogTotal(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&groupmodels.GroupActionLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count group action logs: %v", err)
	}
	if count != want {
		t.Fatalf("group action log count = %d, want %d", count, want)
	}
}

func assertGroupServiceActionLogContains(t *testing.T, db *gorm.DB, groupID uint, action, wantSubstring string) {
	t.Helper()

	var log groupmodels.GroupActionLog
	if err := db.Where("group_id = ? AND action = ?", groupID, action).First(&log).Error; err != nil {
		t.Fatalf("find group action log: %v", err)
	}
	if !strings.Contains(log.Description, wantSubstring) {
		t.Fatalf("group action log description = %q, want substring %q", log.Description, wantSubstring)
	}
}

func assertGroupActionActorFields(t *testing.T, db *gorm.DB, groupID uint, action, wantName, wantUs string) {
	t.Helper()

	var log groupmodels.GroupActionLog
	if err := db.Where("group_id = ? AND action = ?", groupID, action).First(&log).Error; err != nil {
		t.Fatalf("find group action log: %v", err)
	}
	if log.Username != wantName {
		t.Fatalf("action username = %q, want %q", log.Username, wantName)
	}
	if log.Us != wantUs {
		t.Fatalf("action us = %q, want %q", log.Us, wantUs)
	}
}
