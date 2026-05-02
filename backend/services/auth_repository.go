package services

import (
	"fmt"
	"friendship/models/dto"
	groupmodels "friendship/models/groups"

	"gorm.io/gorm"
)

const (
	authPrivateGroupType = "приватная группа"
	authPublicGroupType  = "открытая группа"
	authAdminRole        = groupmodels.RoleAdmin
	authModeratorRole    = groupmodels.RoleModerator
)

type AuthUser struct {
	ID       uint
	Name     string
	Password string
	Us       string
	Image    string
}

type AuthUserAdminGroupRepository interface {
	FindAuthUserByEmail(email string) (AuthUser, error)
	FindAuthUserByID(id uint) (AuthUser, error)
	GetAuthAdminGroups(userID uint) ([]dto.AdminGroupResponse, error)
}

type AuthGORMStore interface {
	Model(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
}

type gormAuthRepository struct {
	db AuthGORMStore
}

type authUserRecord struct {
	ID       uint
	Name     string
	Password string
	Us       string
	Image    string
}

func (authUserRecord) TableName() string {
	return "users"
}

type authGroupRecord struct{}

func (authGroupRecord) TableName() string {
	return "groups"
}

type authCategoryRecord struct{}

func (authCategoryRecord) TableName() string {
	return "categories"
}

func NewGORMAuthRepository(db AuthGORMStore) AuthUserAdminGroupRepository {
	return &gormAuthRepository{db: db}
}

func (r *gormAuthRepository) FindAuthUserByEmail(email string) (AuthUser, error) {
	var user authUserRecord
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return AuthUser{}, err
	}
	return authUserFromRecord(user), nil
}

func (r *gormAuthRepository) FindAuthUserByID(id uint) (AuthUser, error) {
	var user authUserRecord
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return AuthUser{}, err
	}
	return authUserFromRecord(user), nil
}

func (r *gormAuthRepository) GetAuthAdminGroups(userID uint) ([]dto.AdminGroupResponse, error) {
	if userID == 0 {
		return nil, fmt.Errorf("некорректный ID пользователя")
	}

	var adminGroups []dto.AdminGroupResponse

	err := r.db.Model(&authGroupRecord{}).
		Select(
			"groups.id, groups.name, groups.image, groups.small_description, CASE WHEN groups.is_private THEN ? ELSE ? END as type, COUNT(DISTINCT gu2.user_id) as member_count, rig.name as role",
			authPrivateGroupType,
			authPublicGroupType,
		).
		Joins("JOIN group_users gu ON gu.group_id = groups.id").
		Joins("JOIN role_in_groups rig ON gu.role_in_group_id = rig.id").
		Joins("LEFT JOIN group_users gu2 ON gu2.group_id = groups.id").
		Where("gu.user_id = ? AND (rig.name = ? OR rig.name = ?)", userID, authAdminRole, authModeratorRole).
		Group("groups.id, groups.name, groups.image, groups.small_description, groups.is_private, rig.name").
		Order("groups.id DESC").
		Scan(&adminGroups).Error

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении групп: %w", err)
	}

	for i := range adminGroups {
		if adminGroups[i].ID == nil {
			continue
		}

		categories, err := r.loadGroupCategories(*adminGroups[i].ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении категорий для группы %d: %w", *adminGroups[i].ID, err)
		}
		adminGroups[i].Category = categories
	}

	return adminGroups, nil
}

func (r *gormAuthRepository) loadGroupCategories(groupID uint) ([]*string, error) {
	var categoryNames []string

	err := r.db.Model(&authCategoryRecord{}).
		Select("categories.name").
		Joins("JOIN group_group_categories ON group_group_categories.group_category_id = categories.id").
		Where("group_group_categories.group_id = ?", groupID).
		Pluck("name", &categoryNames).Error
	if err != nil {
		return nil, err
	}

	categories := make([]*string, len(categoryNames))
	for i, name := range categoryNames {
		categoryName := name
		categories[i] = &categoryName
	}

	return categories, nil
}

func authUserFromRecord(user authUserRecord) AuthUser {
	return AuthUser{
		ID:       user.ID,
		Name:     user.Name,
		Password: user.Password,
		Us:       user.Us,
		Image:    user.Image,
	}
}
