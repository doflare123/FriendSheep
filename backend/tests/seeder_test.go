package tests

import (
	"os"
	"testing"

	"friendship/db"
	"friendship/models"
	"friendship/models/events"
	"friendship/models/groups"
	statsusers "friendship/models/stats_users"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSeederFillsReferenceTables(t *testing.T) {
	restoreWorkingDir(t, "..")

	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := gormDB.AutoMigrate(
		&models.Category{},
		&models.DaysWeek{},
		&events.EventLocation{},
		&events.Status{},
		&events.AgeLimit{},
		&groups.Role_in_group{},
		&groups.GroupActionType{},
		&statsusers.Genre{},
	); err != nil {
		t.Fatalf("auto migrate seeded models: %v", err)
	}

	errs := db.Seeder(&testPostgresRepository{db: gormDB})
	if len(errs) > 0 {
		t.Fatalf("Seeder returned errors: %v", errs)
	}

	assertReferenceCount(t, gormDB, &models.Category{}, 4)
	assertReferenceCount(t, gormDB, &models.DaysWeek{}, 7)
	assertReferenceCount(t, gormDB, &events.EventLocation{}, 2)
	assertReferenceCount(t, gormDB, &events.Status{}, 3)
	assertReferenceCount(t, gormDB, &events.AgeLimit{}, 5)
	assertReferenceCount(t, gormDB, &groups.Role_in_group{}, 3)
	assertReferenceCount(t, gormDB, &groups.GroupActionType{}, int64(len(groups.DefaultGroupActionTypes())))

	var genresCount int64
	if err := gormDB.Model(&statsusers.Genre{}).Count(&genresCount).Error; err != nil {
		t.Fatalf("count genres: %v", err)
	}
	if genresCount == 0 {
		t.Fatal("seeded genres count = 0, want > 0")
	}
}

func restoreWorkingDir(t *testing.T, target string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	if err := os.Chdir(target); err != nil {
		t.Fatalf("change working dir to %s: %v", target, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
}

func assertReferenceCount(t *testing.T, gormDB *gorm.DB, model interface{}, want int64) {
	t.Helper()

	var got int64
	if err := gormDB.Model(model).Count(&got).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if got != want {
		t.Fatalf("count %T = %d, want %d", model, got, want)
	}
}
