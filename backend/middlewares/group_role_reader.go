package middlewares

import (
	"errors"
	eventmodels "friendship/models/events"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

var errGroupMembershipNotFound = errors.New("group membership not found")
var errGroupRoleReaderUnavailable = errors.New("group role reader unavailable")
var errEventNotFound = errors.New("event not found")

type groupRoleReader interface {
	FindUserGroupRole(userID, groupID uint) (string, error)
	FindUserEventGroupRole(userID, eventID uint) (uint, string, error)
}

type repositoryGroupRoleReader struct {
	repo repository.PostgresRepository
}

func newRepositoryGroupRoleReader(repo repository.PostgresRepository) groupRoleReader {
	return &repositoryGroupRoleReader{repo: repo}
}

func (r *repositoryGroupRoleReader) FindUserGroupRole(userID, groupID uint) (string, error) {
	if r == nil || r.repo == nil {
		return "", errGroupRoleReaderUnavailable
	}

	var groupUser groups.GroupUsers
	if err := r.repo.Where("user_id = ? AND group_id = ?", userID, groupID).First(&groupUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errGroupMembershipNotFound
		}
		return "", err
	}

	var role groups.Role_in_group
	if err := r.repo.First(&role, groupUser.RoleInGroupID).Error; err != nil {
		return "", err
	}

	return role.Name, nil
}

func (r *repositoryGroupRoleReader) FindUserEventGroupRole(userID, eventID uint) (uint, string, error) {
	if r == nil || r.repo == nil {
		return 0, "", errGroupRoleReaderUnavailable
	}

	var event eventmodels.Event
	if err := r.repo.Select("group_id").First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", errEventNotFound
		}
		return 0, "", err
	}

	role, err := r.FindUserGroupRole(userID, event.GroupID)
	if err != nil {
		return event.GroupID, "", err
	}

	return event.GroupID, role, nil
}
