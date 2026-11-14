package tests

import (
	"sync"
	"testing"
	"time"

	"friendship/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestUserOptionalBirthdayAndCityPersist(t *testing.T) {
	dsn := "file:user-fields-" + testUintString(uint(time.Now().UnixNano())) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	if err := db.AutoMigrate(&models.City{}, &models.User{}); err != nil {
		t.Fatalf("migrate cities and users: %v", err)
	}

	parsedSchema, err := schema.Parse(&models.User{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse user schema: %v", err)
	}
	birthdaySchemaField := parsedSchema.LookUpField("Birthday")
	if birthdaySchemaField == nil {
		t.Fatal("Birthday column is missing")
	}
	if dataType := postgres.New(postgres.Config{}).DataTypeOf(birthdaySchemaField); dataType != "date" {
		t.Fatalf("PostgreSQL Birthday type = %q, want date", dataType)
	}

	withoutOptionalFields := models.User{Name: "Without optional fields", Password: "Password123!", Us: "no-optional-fields", Email: "no-optional-fields@example.com"}
	if err := db.Create(&withoutOptionalFields).Error; err != nil {
		t.Fatalf("create user without optional fields: %v", err)
	}
	var readWithoutOptionalFields models.User
	if err := db.First(&readWithoutOptionalFields, withoutOptionalFields.ID).Error; err != nil {
		t.Fatalf("read user without optional fields: %v", err)
	}
	if readWithoutOptionalFields.Birthday != nil || readWithoutOptionalFields.CityID != nil {
		t.Fatalf("optional defaults = birthday %v, city ID %v; want NULL values", readWithoutOptionalFields.Birthday, readWithoutOptionalFields.CityID)
	}

	city := models.City{Name: "Калининград", Region: "Калининградская обл", FIASID: "df679824-5d6a-4a2d-a2d4-5b1e8a13fd57"}
	if err := db.Create(&city).Error; err != nil {
		t.Fatalf("create city: %v", err)
	}
	birthday := time.Date(2001, time.February, 3, 0, 0, 0, 0, time.UTC)
	withOptionalFields := models.User{Name: "With optional fields", Password: "Password123!", Us: "with-optional-fields", Email: "with-optional-fields@example.com", Birthday: &birthday, CityID: &city.ID}
	if err := db.Create(&withOptionalFields).Error; err != nil {
		t.Fatalf("create user with optional fields: %v", err)
	}
	var readWithOptionalFields models.User
	if err := db.Preload("City").First(&readWithOptionalFields, withOptionalFields.ID).Error; err != nil {
		t.Fatalf("read user with city: %v", err)
	}
	if readWithOptionalFields.Birthday == nil || readWithOptionalFields.Birthday.Format(time.DateOnly) != birthday.Format(time.DateOnly) || readWithOptionalFields.CityID == nil || *readWithOptionalFields.CityID != city.ID || readWithOptionalFields.City == nil || readWithOptionalFields.City.FIASID != city.FIASID {
		t.Fatalf("optional roundtrip = birthday %v, city ID %v, city %#v; want %s and %d", readWithOptionalFields.Birthday, readWithOptionalFields.CityID, readWithOptionalFields.City, birthday.Format(time.DateOnly), city.ID)
	}

	unknownCityID := city.ID + 1000
	invalidCityUser := models.User{Name: "Invalid city", Password: "Password123!", Us: "invalid-city", Email: "invalid-city@example.com", CityID: &unknownCityID}
	if err := db.Create(&invalidCityUser).Error; err == nil {
		t.Fatal("create user with unknown city succeeded, want foreign-key error")
	}

	if err := db.Delete(&city).Error; err != nil {
		t.Fatalf("delete city: %v", err)
	}
	readWithOptionalFields = models.User{}
	if err := db.First(&readWithOptionalFields, withOptionalFields.ID).Error; err != nil {
		t.Fatalf("read user after city deletion: %v", err)
	}
	if readWithOptionalFields.CityID != nil {
		t.Fatalf("city ID after city deletion = %v, want NULL", readWithOptionalFields.CityID)
	}
}
