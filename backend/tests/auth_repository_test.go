package tests

import (
	"context"
	"errors"
	"testing"

	"friendship/models"
	"friendship/models/groups"
	"friendship/services"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGORMAuthRepositoryFindsUserAndAdminGroups(t *testing.T) {
	db := newAuthRepositoryDB(t)

	user := models.User{
		Name:     "Valid User",
		Password: "hash",
		Us:       "valid_user",
		Email:    "user@example.com",
		Image:    "avatar.png",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	role := groups.Role_in_group{Name: "Админ"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	category := models.Category{Name: "Sport"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	group := groups.Group{
		Name:             "Running Club",
		Description:      "Long description",
		SmallDescription: "Short description",
		Image:            "group.png",
		CreaterID:        user.ID,
		IsPrivate:        false,
		City:             "Kaliningrad",
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}

	if err := db.Create(&groups.GroupUsers{
		UserID:        user.ID,
		GroupID:       group.ID,
		RoleInGroupID: role.Id,
	}).Error; err != nil {
		t.Fatalf("create group user: %v", err)
	}
	if err := db.Create(&groups.GroupGroupCategory{
		GroupID:         group.ID,
		GroupCategoryID: category.ID,
	}).Error; err != nil {
		t.Fatalf("create group category: %v", err)
	}

	repo := services.NewGORMAuthRepository(&testPostgresRepository{db: db})

	authUser, err := repo.FindAuthUserByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("FindAuthUserByEmail returned error: %v", err)
	}
	if authUser.ID != user.ID || authUser.Name != user.Name || authUser.Password != user.Password {
		t.Fatalf("FindAuthUserByEmail returned %#v, want user id/name/password", authUser)
	}

	authUser, err = repo.FindAuthUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("FindAuthUserByID returned error: %v", err)
	}
	if authUser.Us != user.Us || authUser.Image != user.Image {
		t.Fatalf("FindAuthUserByID returned %#v, want us/image", authUser)
	}

	adminGroups, err := repo.GetAuthAdminGroups(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetAuthAdminGroups returned error: %v", err)
	}
	if len(adminGroups) != 1 {
		t.Fatalf("admin groups len = %d, want 1", len(adminGroups))
	}

	gotGroup := adminGroups[0]
	if gotGroup.ID == nil || *gotGroup.ID != group.ID {
		t.Fatalf("admin group id = %v, want %d", gotGroup.ID, group.ID)
	}
	if gotGroup.Name == nil || *gotGroup.Name != group.Name {
		t.Fatalf("admin group name = %v, want %s", gotGroup.Name, group.Name)
	}
	if gotGroup.Role != role.Name {
		t.Fatalf("admin group role = %q, want %q", gotGroup.Role, role.Name)
	}
	if gotGroup.MemberCount == nil || *gotGroup.MemberCount != 1 {
		t.Fatalf("admin group member count = %v, want 1", gotGroup.MemberCount)
	}
	if len(gotGroup.Category) != 1 || gotGroup.Category[0] == nil || *gotGroup.Category[0] != category.Name {
		t.Fatalf("admin group category = %#v, want %s", gotGroup.Category, category.Name)
	}
}

func TestGORMAuthRepositoryMapsMissingUsers(t *testing.T) {
	repo := services.NewGORMAuthRepository(&testPostgresRepository{db: newAuthRepositoryDB(t)})
	if _, err := repo.FindAuthUserByEmail(context.Background(), "missing@example.com"); !errors.Is(err, services.ErrAuthUserNotFound) {
		t.Fatalf("FindAuthUserByEmail error = %v, want ErrAuthUserNotFound", err)
	}
	if _, err := repo.FindAuthUserByID(context.Background(), 999); !errors.Is(err, services.ErrAuthUserNotFound) {
		t.Fatalf("FindAuthUserByID error = %v, want ErrAuthUserNotFound", err)
	}
}

func newAuthRepositoryDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&groups.Role_in_group{},
		&groups.Group{},
		&groups.GroupUsers{},
		&groups.GroupGroupCategory{},
	); err != nil {
		t.Fatalf("auto migrate auth repository models: %v", err)
	}

	return db
}
