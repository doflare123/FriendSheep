package middlewares

import (
	"errors"
	"friendship/models/groups"
	"friendship/repository"

	"gorm.io/gorm"
)

var errGroupMembershipNotFound = errors.New("group membership not found")
var errGroupRoleReaderUnavailable = errors.New("group role reader unavailable")

type groupRoleReader interface {
	FindUserGroupRole(userID, groupID uint) (string, error)
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
