package group

type groupUnitOfWork interface {
	WithinTransaction(func(groupTx) error) error
	Access() groupActorRoleFinder
	Reads() groupManagementReadStore
	Subscriptions() groupSubscriptionsStore
}

type groupTx interface {
	Admin() groupAdminStore
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

func (s *groupService) runInTx(fn func(groupTx) error) error {
	return s.uow.WithinTransaction(fn)
}
