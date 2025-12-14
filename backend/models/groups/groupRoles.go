package groups

import (
	"errors"
	"friendship/repository"

	"gorm.io/gorm"
)

type Role_in_group struct {
	Id   uint
	Name string
}

func (r *Role_in_group) GetIdRole(str string, post repository.PostgresRepository) uint {
	var role Role_in_group

	err := post.Where("name = ?", str).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0
		}
		return 0
	}

	return role.Id
}
