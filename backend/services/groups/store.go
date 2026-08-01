package group

import (
	"context"
	"errors"
)

var (
	errGroupTransactionContextMissing  = errors.New("контекст транзакции группы не задан")
	errGroupTransactionCallbackMissing = errors.New("обработчик транзакции группы не задан")
	errGroupOperationContextMissing    = errors.New("контекст операции группы не задан")
)

type groupUnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(groupTx) error) error
	Access() rootGroupActorRoleFinder
	Reads() groupManagementReadStore
	Subscriptions() groupSubscriptionsStore
}

type groupTx interface {
	Admin() groupAdminStore
	Reads() groupManagementReadStore
	JoinGroup() joinGroupStore
	LeaveGroup() leaveGroupStore
	JoinInviteCreation() joinInviteCreationStore
	JoinInviteResponse() joinInviteResponseStore
	JoinRequestReview() joinRequestReviewStore
	BulkJoinRequest() bulkJoinRequestStore
	LogActorAction(input groupActorActionLogInput) error
}

type groupActorActionLogInput struct {
	GroupID      uint
	UserID       uint
	Username     string
	Us           string
	Role         string
	Action       string
	Description  string
	TargetUserID *uint
}

func (s *groupService) runInTx(ctx context.Context, fn func(groupTx) error) error {
	return s.uow.WithinTransaction(ctx, fn)
}
