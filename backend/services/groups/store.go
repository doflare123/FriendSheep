package group

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type groupUnitOfWork interface {
	WithinTransaction(func(groupTx) error) error
	Access() groupActorRoleFinder
	Reads() groupManagementReadStore
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

type groupPersistence interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Find(out interface{}, where ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Clauses(conds ...clause.Expression) *gorm.DB
	Order(value interface{}) *gorm.DB
	Limit(limit int) *gorm.DB
}

type groupRoleStore interface {
	Where(query interface{}, args ...interface{}) *gorm.DB
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

func isGroupRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
