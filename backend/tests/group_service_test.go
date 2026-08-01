package tests

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"friendship/models"
	eventmodels "friendship/models/events"
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

	if err != nil {
		t.Fatalf("JoinGroup returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipExists(t, db, groupID, 2, true)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
}

func TestGroupServiceJoinGroupCanceledContextSkipsTransactionWrites(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := service.JoinGroup(ctx, 2, groupID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinRequestCount(t, db, groupID, 2, "pending", 0)
	assertGroupServiceActionLogCount(t, db, groupID, groupmodels.ActionJoinGroup, 0)
}

func TestGroupServiceJoinGroupRejectsDuplicatePublicMembership(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupBlacklist(t, db, groupID, 2, 1)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupJoinRequest(t, db, groupID, 2, "pending")

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupJoinRequest(t, db, groupID, 2, "rejected")
	createPendingJoinRequestUniqueIndex(t, db)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	createPendingJoinRequestUniqueIndex(t, db)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	result, err := service.JoinGroup(context.Background(), 2, groupID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, memberRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupBlacklist(t, db, groupID, 2, 1)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	approved, err := service.ApproveJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	operatorRoleID := seedGroupServiceRole(t, db, "Модератор")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, memberRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	operatorRoleID := seedGroupServiceRole(t, db, groupmodels.RoleModerator)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, true)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)
	requestID := seedGroupJoinRequestWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinRequest(context.Background(), 1, requestID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	operatorRoleID := seedGroupServiceRole(t, db, groupmodels.RoleModerator)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, memberRoleID)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	assertGroupActionTargetUser(t, db, groupID, "send_invite", 2)
	assertGroupActionRawDescriptionNotContains(t, db, groupID, "send_invite", "group-user-2")
}

func TestGroupServiceCreateJoinInviteKeepsInviteWhenActionLogFails(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &actionLogFailureRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)

	created, err := service.CreateJoinInvite(context.Background(), 1, servicegroups.GroupUserInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	seedGroupCategory(t, db, 1, "Board Games")
	isPrivate := false
	categoryID := uint(1)
	missingCategoryID := uint(999)

	groupDTO, err := service.CreateGroup(context.Background(), 1, servicegroups.CreateGroupInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	setGroupServiceUserImage(t, db, 1, "https://cdn.example.com/avatar.png")
	seedGroupCategory(t, db, 1, "Board Games")
	isPrivate := true
	categoryID := uint(1)

	groupDTO, err := service.CreateGroup(context.Background(), 1, servicegroups.CreateGroupInput{
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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	setGroupServiceUserImage(t, db, 1, "https://cdn.example.com/avatar.png")
	seedGroupCategory(t, db, 1, "Board Games")
	isPrivate := true
	categoryID := uint(1)

	created, err := service.CreateGroup(context.Background(), 1, servicegroups.CreateGroupInput{
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
	updated, err := service.UpdateGroup(context.Background(), 1, servicegroups.GroupUpdateInput{
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

func TestGroupServiceCreateGroupRollsBackWhenMaterializationIsCanceled(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))
	services.InitValidator(validator.New())

	seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceUser(t, db, 1)
	seedGroupCategory(t, db, 1, "Board Games")
	registerCanceledGroupMaterialization(t, db, "create")
	isPrivate := false
	categoryID := uint(1)

	created, err := service.CreateGroup(context.Background(), 1, servicegroups.CreateGroupInput{
		Name:             "Board Game Club",
		Description:      "Group for board game fans",
		SmallDescription: "Play together",
		Image:            "https://example.com/group.png",
		IsPrivate:        &isPrivate,
		Categories:       []*uint{&categoryID},
		Contacts:         "tg:https://t.me/group",
	})

	if created != nil {
		t.Fatalf("created = %#v, want nil", created)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	assertGroupTotal(t, db, 0)
	assertGroupMembershipTotal(t, db, 0)
	assertGroupContactTotal(t, db, 0)
	assertGroupServiceActionLogTotal(t, db, 0)
}

func TestGroupServiceUpdateGroupRollsBackWhenMaterializationIsCanceled(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	seedGroupServiceUser(t, db, 1)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	registerCanceledGroupMaterialization(t, db, "update")
	newName := "Updated Group Name"

	updated, err := service.UpdateGroup(context.Background(), 1, servicegroups.GroupUpdateInput{
		GroupID: groupID,
		Name:    &newName,
	})

	if updated != nil {
		t.Fatalf("updated = %#v, want nil", updated)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}

	var persisted groupmodels.Group
	if err := db.First(&persisted, groupID).Error; err != nil {
		t.Fatalf("load persisted group: %v", err)
	}
	if persisted.Name != "Group Service Test" {
		t.Fatalf("persisted name = %q, want original %q", persisted.Name, "Group Service Test")
	}
	assertGroupServiceActionLogTotal(t, db, 0)
}

func TestGroupServiceGetGroupDetailsMarksActiveEventSubscription(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, groupmodels.RoleAdmin)
	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	seedGroupServiceUser(t, db, 3)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)
	seedGroupServiceMembership(t, db, groupID, 3, memberRoleID)
	eventID := seedEvent(t, db, groupID, 1, 1, 5)
	seedEventParticipant(t, db, eventID, 2)

	subscribedDetails, err := service.GetGroupDetails(context.Background(), 2, groupID)
	if err != nil {
		t.Fatalf("GetGroupDetails for subscribed user returned error: %v", err)
	}
	if len(subscribedDetails.ActiveEvents) != 1 {
		t.Fatalf("active events len = %d, want 1", len(subscribedDetails.ActiveEvents))
	}
	if !subscribedDetails.ActiveEvents[0].Subscribed {
		t.Fatal("active event subscribed = false, want true")
	}

	unsubscribedDetails, err := service.GetGroupDetails(context.Background(), 3, groupID)
	if err != nil {
		t.Fatalf("GetGroupDetails for unsubscribed user returned error: %v", err)
	}
	if len(unsubscribedDetails.ActiveEvents) != 1 {
		t.Fatalf("active events len = %d, want 1", len(unsubscribedDetails.ActiveEvents))
	}
	if unsubscribedDetails.ActiveEvents[0].Subscribed {
		t.Fatal("active event subscribed = true, want false")
	}
}

func TestGroupServiceGetGroupDetailsCanceledContextReturnsNoPartialResult(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	details, err := service.GetGroupDetails(ctx, 1, groupID)

	if details != nil {
		t.Fatalf("details = %#v, want nil", details)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestGroupServiceGetGroupDetailsPropagatesActiveEventsQueryError(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	if err := db.Callback().Query().Before("gorm:query").Register("test:cancel-active-events-query", func(tx *gorm.DB) {
		if tx.Statement.Table == "events" || tx.Statement.Schema != nil && tx.Statement.Schema.Table == "events" {
			tx.AddError(context.Canceled)
		}
	}); err != nil {
		t.Fatalf("register active-events query failure: %v", err)
	}

	details, err := service.GetGroupDetails(context.Background(), 1, groupID)

	if details != nil {
		t.Fatalf("details = %#v, want nil instead of partial success", details)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want active-events context.Canceled", err)
	}
}

func TestGroupServiceAddPermissionsPromotesMemberAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceRole(t, db, "Модератор")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	changed, err := service.AddPermissions(context.Background(), 1, servicegroups.GroupUserInput{GroupID: groupID, UserID: 2})

	if err != nil {
		t.Fatalf("AddPermissions returned error: %v", err)
	}
	if !changed {
		t.Fatal("AddPermissions returned false")
	}
	assertGroupMembershipRole(t, db, groupID, 2, "Модератор")
	assertGroupServiceActionLogCount(t, db, groupID, "add_operator", 1)
	assertGroupActionTargetUser(t, db, groupID, "add_operator", 2)
	assertGroupActionRawDescriptionNotContains(t, db, groupID, "add_operator", "Group User")

	renameGroupServiceUser(t, db, 1, "Renamed Admin", "renamed-admin")
	renameGroupServiceUser(t, db, 2, "Renamed Member", "renamed-member")
	actions, err := service.WatchRecentActions(context.Background(), 1, groupID, servicegroups.GroupActionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("WatchRecentActions returned error: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("actions len = %d, want 1", len(actions))
	}
	assertGroupActionDescription(t, actions[0], "'Renamed Admin' (@renamed-admin) назначил пользователя 'Renamed Member' (@renamed-member) оператором группы")
}

func TestGroupServiceAddPermissionsRejectsModeratorActorWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	operatorRoleID := seedGroupServiceRole(t, db, "Модератор")
	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, operatorRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	changed, err := service.AddPermissions(context.Background(), 1, servicegroups.GroupUserInput{GroupID: groupID, UserID: 2})

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	operatorRoleID := seedGroupServiceRole(t, db, "Модератор")
	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, operatorRoleID)

	changed, err := service.RemovePermissions(context.Background(), 1, servicegroups.GroupUserInput{GroupID: groupID, UserID: 2})

	if err != nil {
		t.Fatalf("RemovePermissions returned error: %v", err)
	}
	if !changed {
		t.Fatal("RemovePermissions returned false")
	}
	assertGroupMembershipRole(t, db, groupID, 2, "Участник")
	assertGroupServiceActionLogCount(t, db, groupID, "remove_operator", 1)
	assertGroupActionTargetUser(t, db, groupID, "remove_operator", 2)
	assertGroupActionRawDescriptionNotContains(t, db, groupID, "remove_operator", "Group User")

	renameGroupServiceUser(t, db, 1, "Renamed Admin", "renamed-admin")
	renameGroupServiceUser(t, db, 2, "Renamed Moderator", "renamed-moderator")
	actions, err := service.WatchRecentActions(context.Background(), 1, groupID, servicegroups.GroupActionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("WatchRecentActions returned error: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("actions len = %d, want 1", len(actions))
	}
	assertGroupActionDescription(t, actions[0], "'Renamed Admin' (@renamed-admin) снял с пользователя 'Renamed Moderator' (@renamed-moderator) права оператора")
}

func TestGroupServiceDeleteUserFromGroupMovesMemberToBlacklistAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	memberRoleID := seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)

	deleted, err := service.DeleteUserFromGroup(context.Background(), 1, groupID, 2)

	if err != nil {
		t.Fatalf("DeleteUserFromGroup returned error: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteUserFromGroup returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupBlacklistExists(t, db, groupID, 2, true)
	assertGroupServiceActionLogCount(t, db, groupID, "ban_user", 1)
	assertGroupActionTargetUser(t, db, groupID, "ban_user", 2)
	assertGroupActionRawDescriptionNotContains(t, db, groupID, "ban_user", "group-user-2")
}

func TestGroupServiceRemoveFromBlacklistDeletesEntryAndWritesActionLog(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	adminRoleID := seedGroupServiceRole(t, db, "Админ")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 1, adminRoleID)
	seedGroupBlacklist(t, db, groupID, 2, 1)

	removed, err := service.RemoveFromBlacklist(context.Background(), 1, groupID, 2)

	if err != nil {
		t.Fatalf("RemoveFromBlacklist returned error: %v", err)
	}
	if !removed {
		t.Fatal("RemoveFromBlacklist returned false")
	}
	assertGroupBlacklistExists(t, db, groupID, 2, false)
	assertGroupServiceActionLogCount(t, db, groupID, "unban_user", 1)
	assertGroupActionTargetUser(t, db, groupID, "unban_user", 2)
	assertGroupActionRawDescriptionNotContains(t, db, groupID, "unban_user", "group-user-2")
}

func TestGroupServiceAcceptJoinInviteAddsMembershipAndUpdatesStatus(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, "Участник")
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

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
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	memberRoleID := seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupServiceMembership(t, db, groupID, 2, memberRoleID)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "accepted")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if err != nil {
		t.Fatalf("AcceptJoinInvite returned error: %v", err)
	}
	if result == nil || !result.Joined {
		t.Fatalf("result = %#v, want joined result", result)
	}
	assertGroupMembershipCount(t, db, groupID, 2, 1)
	assertGroupJoinInviteStatus(t, db, inviteID, "accepted")
}

func TestGroupServiceAcceptJoinInviteRejectsMissingInvite(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	result, err := service.AcceptJoinInvite(context.Background(), 2, 404)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrInviteNotFound) {
		t.Fatalf("err = %v, want ErrInviteNotFound", err)
	}
}

func TestGroupServiceAcceptJoinInviteRejectsInviteOwnedByAnotherUser(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	seedGroupServiceUser(t, db, 3)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 3, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrInviteNotOwned) {
		t.Fatalf("err = %v, want ErrInviteNotOwned", err)
	}
	assertGroupMembershipExists(t, db, groupID, 3, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "pending")
}

func TestGroupServiceAcceptJoinInviteRejectsAcceptedInviteWithoutMembership(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "accepted")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrInviteAlreadyHandled) {
		t.Fatalf("err = %v, want ErrInviteAlreadyHandled", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "accepted")
}

func TestGroupServiceAcceptJoinInviteRejectsRejectedInvite(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "rejected")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrInviteAlreadyHandled) {
		t.Fatalf("err = %v, want ErrInviteAlreadyHandled", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "rejected")
}

func TestGroupServiceAcceptJoinInviteRejectsBlacklistedUserWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	seedGroupBlacklist(t, db, groupID, 2, 1)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrUserInBlacklist) {
		t.Fatalf("err = %v, want ErrUserInBlacklist", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "pending")
}

func TestGroupServiceAcceptJoinInviteRejectsMissingMemberRoleWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrRoleMemberNotFound) {
		t.Fatalf("err = %v, want ErrRoleMemberNotFound", err)
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "pending")
}

func TestGroupServiceAcceptJoinInviteHandlesStatusRaceWithoutSideEffects(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &inviteStatusRaceRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !errors.Is(err, servicegroups.ErrInviteAlreadyHandled) {
		t.Fatalf("err = %v, want ErrInviteAlreadyHandled", err)
	}
	if !repo.injected.Load() {
		t.Fatal("test status race injection was not reached")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "pending")
}

func TestGroupServiceAcceptJoinInviteRollsBackMembershipWhenStatusUpdateFails(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &inviteStatusUpdateFailureRepository{testPostgresRepository: &testPostgresRepository{db: db}}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceRole(t, db, groupmodels.RoleMember)
	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	result, err := service.AcceptJoinInvite(context.Background(), 2, inviteID)

	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if err == nil || !strings.Contains(err.Error(), "forced invite status update failure") {
		t.Fatalf("err = %v, want forced invite status update failure", err)
	}
	if !repo.injected.Load() {
		t.Fatal("test status update failure injection was not reached")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "pending")
}

func TestGroupServiceRejectJoinInviteUpdatesStatusWithoutMembership(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinInvite(context.Background(), 2, inviteID)

	if err != nil {
		t.Fatalf("RejectJoinInvite returned error: %v", err)
	}
	if !rejected {
		t.Fatal("RejectJoinInvite returned false")
	}
	assertGroupMembershipExists(t, db, groupID, 2, false)
	assertGroupJoinInviteStatus(t, db, inviteID, "rejected")
}

func TestGroupServiceRejectJoinInviteRejectsMissingInvite(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	rejected, err := service.RejectJoinInvite(context.Background(), 2, 404)

	if rejected {
		t.Fatal("RejectJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrInviteNotFound) {
		t.Fatalf("err = %v, want ErrInviteNotFound", err)
	}
}

func TestGroupServiceRejectJoinInviteRejectsInviteOwnedByAnotherUser(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	seedGroupServiceUser(t, db, 3)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "pending")

	rejected, err := service.RejectJoinInvite(context.Background(), 3, inviteID)

	if rejected {
		t.Fatal("RejectJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrInviteNotOwned) {
		t.Fatalf("err = %v, want ErrInviteNotOwned", err)
	}
	assertGroupJoinInviteStatus(t, db, inviteID, "pending")
}

func TestGroupServiceRejectJoinInviteRejectsHandledInvite(t *testing.T) {
	db := newGroupServiceDB(t)
	repo := &testPostgresRepository{db: db}
	service := servicegroups.NewGroupService(&testLogger{}, servicegroups.NewGORMGroupRepository(repo))

	seedGroupServiceUser(t, db, 1)
	seedGroupServiceUser(t, db, 2)
	groupID := seedGroupServiceGroup(t, db, 1, false)
	inviteID := seedGroupJoinInviteWithID(t, db, groupID, 2, "accepted")

	rejected, err := service.RejectJoinInvite(context.Background(), 2, inviteID)

	if rejected {
		t.Fatal("RejectJoinInvite returned true")
	}
	if !errors.Is(err, servicegroups.ErrInviteAlreadyHandled) {
		t.Fatalf("err = %v, want ErrInviteAlreadyHandled", err)
	}
	assertGroupJoinInviteStatus(t, db, inviteID, "accepted")
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
		&groupmodels.GroupActionType{},
		&groupmodels.GroupActionLog{},
		&eventmodels.EventLocation{},
		&eventmodels.Status{},
		&eventmodels.AgeLimit{},
		&eventmodels.Genre{},
		&eventmodels.Event{},
		&eventmodels.EventGenre{},
		&eventmodels.EventsUser{},
	); err != nil {
		t.Fatalf("auto migrate group service models: %v", err)
	}

	seedGroupActionTypes(t, db)

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

func seedGroupActionTypes(t *testing.T, db *gorm.DB) {
	t.Helper()

	for _, actionType := range groupmodels.DefaultGroupActionTypes() {
		if err := db.Where("code = ?", actionType.Code).FirstOrCreate(&groupmodels.GroupActionType{}, actionType).Error; err != nil {
			t.Fatalf("seed group action type %q: %v", actionType.Code, err)
		}
	}
}

func setGroupServiceUserImage(t *testing.T, db *gorm.DB, userID uint, image string) {
	t.Helper()

	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("image", image).Error; err != nil {
		t.Fatalf("update user image for %d: %v", userID, err)
	}
}

func renameGroupServiceUser(t *testing.T, db *gorm.DB, userID uint, name string, us string) {
	t.Helper()

	if err := db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"name": name,
			"us":   us,
		}).Error; err != nil {
		t.Fatalf("rename user %d: %v", userID, err)
	}
}

func registerCanceledGroupMaterialization(t *testing.T, db *gorm.DB, mutation string) {
	t.Helper()

	var armed atomic.Bool
	callbackPrefix := "test:cancel-group-materialization:" + strings.NewReplacer("/", "-", " ", "-").Replace(t.Name())
	arm := func(tx *gorm.DB) {
		if gormStatementTargetsTable(tx, "groups") {
			armed.Store(true)
		}
	}

	var err error
	switch mutation {
	case "create":
		err = db.Callback().Create().After("gorm:create").Register(callbackPrefix+":arm", arm)
	case "update":
		err = db.Callback().Update().After("gorm:update").Register(callbackPrefix+":arm", arm)
	default:
		t.Fatalf("unsupported mutation callback %q", mutation)
	}
	if err != nil {
		t.Fatalf("register %s materialization arm callback: %v", mutation, err)
	}

	if err := db.Callback().Query().Before("gorm:query").Register(callbackPrefix+":cancel", func(tx *gorm.DB) {
		if gormStatementTargetsTable(tx, "groups") && armed.CompareAndSwap(true, false) {
			tx.AddError(context.Canceled)
		}
	}); err != nil {
		t.Fatalf("register group materialization cancellation callback: %v", err)
	}
}

func gormStatementTargetsTable(tx *gorm.DB, table string) bool {
	return tx.Statement.Table == table || tx.Statement.Schema != nil && tx.Statement.Schema.Table == table
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

func (r *duplicatePendingRequestRepository) TransactionWithContext(ctx context.Context, fc func(tx repository.PostgresRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (r *duplicateMembershipRepository) TransactionWithContext(ctx context.Context, fc func(tx repository.PostgresRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (r *actionLogFailureRepository) TransactionWithContext(ctx context.Context, fc func(tx repository.PostgresRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

type inviteStatusRaceRepository struct {
	*testPostgresRepository
	injected atomic.Bool
}

func (r *inviteStatusRaceRepository) TransactionWithContext(ctx context.Context, fc func(tx repository.PostgresRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(&inviteStatusRaceTxRepository{
			testPostgresRepository: &testPostgresRepository{db: tx},
			injected:               &r.injected,
		})
	})
}

type inviteStatusRaceTxRepository struct {
	*testPostgresRepository
	injected *atomic.Bool
}

func (r *inviteStatusRaceTxRepository) Model(value interface{}) *gorm.DB {
	if _, ok := value.(*groupmodels.GroupJoinInvite); ok && r.injected.CompareAndSwap(false, true) {
		if err := r.db.Model(&groupmodels.GroupJoinInvite{}).
			Where("status = ?", "pending").
			Update("status", "rejected").Error; err != nil {
			result := r.db.Session(&gorm.Session{})
			result.Error = err
			return result
		}
	}
	return r.testPostgresRepository.Model(value)
}

type inviteStatusUpdateFailureRepository struct {
	*testPostgresRepository
	injected atomic.Bool
}

func (r *inviteStatusUpdateFailureRepository) TransactionWithContext(ctx context.Context, fc func(tx repository.PostgresRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(&inviteStatusUpdateFailureTxRepository{
			testPostgresRepository: &testPostgresRepository{db: tx},
			injected:               &r.injected,
		})
	})
}

type inviteStatusUpdateFailureTxRepository struct {
	*testPostgresRepository
	injected *atomic.Bool
}

func (r *inviteStatusUpdateFailureTxRepository) Model(value interface{}) *gorm.DB {
	if _, ok := value.(*groupmodels.GroupJoinInvite); ok && r.injected.CompareAndSwap(false, true) {
		result := r.db.Session(&gorm.Session{})
		result.Error = errors.New("forced invite status update failure")
		return result
	}
	return r.testPostgresRepository.Model(value)
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
		Joins("JOIN group_action_types ON group_action_types.id = group_action_logs.action_type_id").
		Where("group_action_logs.group_id = ? AND group_action_types.code = ?", groupID, action).
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

func assertGroupActionRawDescriptionNotContains(t *testing.T, db *gorm.DB, groupID uint, action, unwantedSubstring string) {
	t.Helper()

	var log groupmodels.GroupActionLog
	if err := db.Joins("JOIN group_action_types ON group_action_types.id = group_action_logs.action_type_id").
		Where("group_action_logs.group_id = ? AND group_action_types.code = ?", groupID, action).
		First(&log).Error; err != nil {
		t.Fatalf("find group action log: %v", err)
	}
	if strings.Contains(log.Description, unwantedSubstring) {
		t.Fatalf("group action log description = %q, must not contain %q", log.Description, unwantedSubstring)
	}
}

func assertGroupActionTargetUser(t *testing.T, db *gorm.DB, groupID uint, action string, wantUserID uint) {
	t.Helper()

	var log groupmodels.GroupActionLog
	if err := db.Joins("JOIN group_action_types ON group_action_types.id = group_action_logs.action_type_id").
		Where("group_action_logs.group_id = ? AND group_action_types.code = ?", groupID, action).
		First(&log).Error; err != nil {
		t.Fatalf("find group action log: %v", err)
	}
	if log.TargetUserID == nil {
		t.Fatalf("target user id is nil, want %d", wantUserID)
	}
	if *log.TargetUserID != wantUserID {
		t.Fatalf("target user id = %d, want %d", *log.TargetUserID, wantUserID)
	}
}

func assertGroupActionDescription(t *testing.T, action servicegroups.GroupAction, want string) {
	t.Helper()

	if action.Description != want {
		t.Fatalf("action description = %q, want %q", action.Description, want)
	}
}

func assertGroupActionActorFields(t *testing.T, db *gorm.DB, groupID uint, action, wantName, wantUs string) {
	t.Helper()

	var log groupmodels.GroupActionLog
	if err := db.Joins("JOIN group_action_types ON group_action_types.id = group_action_logs.action_type_id").
		Where("group_action_logs.group_id = ? AND group_action_types.code = ?", groupID, action).
		First(&log).Error; err != nil {
		t.Fatalf("find group action log: %v", err)
	}
	if log.Username != wantName {
		t.Fatalf("action username = %q, want %q", log.Username, wantName)
	}
	if log.Us != wantUs {
		t.Fatalf("action us = %q, want %q", log.Us, wantUs)
	}
}
