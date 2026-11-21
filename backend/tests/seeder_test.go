package tests

import (
	"os"
	"testing"
	"time"

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

	dsn := "file:seeder-" + testUintString(uint(time.Now().UnixNano())) + "?mode=memory&cache=shared"
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := gormDB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := gormDB.AutoMigrate(
		&models.City{},
		&models.User{},
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
	assertReferenceCount(t, gormDB, &models.City{}, 1117)

	var kaliningrad models.City
	if err := gormDB.Where("fias_id = ?", "df679694-d505-4dd3-b514-4ba48c8a97d8").First(&kaliningrad).Error; err != nil {
		t.Fatalf("find Kaliningrad: %v", err)
	}
	if kaliningrad.Name != "Калининград" || kaliningrad.Region != "Калининградская обл" {
		t.Fatalf("Kaliningrad = %#v, want its snapshot values", kaliningrad)
	}
	if kaliningrad.Latitude == nil || kaliningrad.Longitude == nil || *kaliningrad.Latitude != 54.7074702 || *kaliningrad.Longitude != 20.5073241 {
		t.Fatalf("координаты Калининграда = %#v, ожидались значения встроенного CSV", kaliningrad)
	}

	var moscow models.City
	if err := gormDB.Where("fias_id = ?", "0c5b2444-70a0-4932-980c-b4dc0d3f02b5").First(&moscow).Error; err != nil {
		t.Fatalf("find Moscow: %v", err)
	}
	if moscow.Name != "Москва" || moscow.Region != "Москва" {
		t.Fatalf("Moscow = %#v, want federal-city snapshot values", moscow)
	}

	var voskresensk models.City
	if err := gormDB.Where("fias_id = ?", "555e9f1b-627a-4444-b19e-27fb9d67a093").First(&voskresensk).Error; err != nil {
		t.Fatalf("find Voskresensk: %v", err)
	}
	if voskresensk.Name != "Воскресенск" || voskresensk.Region != "Московская обл" {
		t.Fatalf("Voskresensk = %#v, want area fallback and region", voskresensk)
	}

	var urusMartan models.City
	if err := gormDB.Where("fias_id = ?", "eb208b95-1f50-4290-a9c9-46909aea49a0").First(&urusMartan).Error; err != nil {
		t.Fatalf("find Urus-Martan: %v", err)
	}
	if urusMartan.Name != "Урус-Мартан" || urusMartan.Region != "Респ Чеченская" {
		t.Fatalf("Urus-Martan = %#v, want settlement fallback and region", urusMartan)
	}

	var belogorskCount int64
	if err := gormDB.Model(&models.City{}).Where("name = ?", "Белогорск").Count(&belogorskCount).Error; err != nil {
		t.Fatalf("count cities named Белогорск: %v", err)
	}
	if belogorskCount != 2 {
		t.Fatalf("cities named Белогорск = %d, want 2", belogorskCount)
	}

	stableKaliningradID := kaliningrad.ID
	user := models.User{Name: "City reference", Password: "Password123!", Us: "city-reference", Email: "city-reference@example.com", CityID: &kaliningrad.ID}
	if err := gormDB.Create(&user).Error; err != nil {
		t.Fatalf("create user with seeded city: %v", err)
	}
	if err := gormDB.Model(&models.City{}).Where("id = ?", kaliningrad.ID).Updates(map[string]interface{}{
		"name": "Неактуальное имя", "latitude": 0, "longitude": 0,
	}).Error; err != nil {
		t.Fatalf("modify city before repeat seed: %v", err)
	}
	if errs := db.Seeder(&testPostgresRepository{db: gormDB}); len(errs) > 0 {
		t.Fatalf("repeat Seeder returned errors: %v", errs)
	}
	assertReferenceCount(t, gormDB, &models.City{}, 1117)
	if err := gormDB.Where("fias_id = ?", kaliningrad.FIASID).First(&kaliningrad).Error; err != nil {
		t.Fatalf("find Kaliningrad after repeat seed: %v", err)
	}
	if kaliningrad.ID != stableKaliningradID || kaliningrad.Name != "Калининград" {
		t.Fatalf("Kaliningrad after repeat seed = %#v, want stable ID and restored snapshot name", kaliningrad)
	}
	if kaliningrad.Latitude == nil || kaliningrad.Longitude == nil || *kaliningrad.Latitude != 54.7074702 || *kaliningrad.Longitude != 20.5073241 {
		t.Fatalf("повторное заполнение не восстановило координаты: %#v", kaliningrad)
	}
	var readUser models.User
	if err := gormDB.First(&readUser, user.ID).Error; err != nil {
		t.Fatalf("read user after repeat seed: %v", err)
	}
	if readUser.CityID == nil || *readUser.CityID != stableKaliningradID {
		t.Fatalf("user city ID after repeat seed = %v, want %d", readUser.CityID, stableKaliningradID)
	}

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
