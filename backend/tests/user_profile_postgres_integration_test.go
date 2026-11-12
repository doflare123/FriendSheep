package tests

import (
	"context"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"friendship/models"
	"friendship/services/users"
)

func TestGORMProfileStorePostgresPrivacyJSONBConstraintsAndConcurrentFirstPatch(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FRIENDSHEEP_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set FRIENDSHEEP_TEST_POSTGRES_DSN to run postgres integration tests")
	}
	db := newPostgresEventCommandDB(t, dsn)
	if err := db.AutoMigrate(&models.UserPrivacy{}, &models.UserStatisticTiles{}); err != nil {
		t.Fatalf("auto migrate user privacy: %v", err)
	}
	var dataType string
	if err := db.Raw(`SELECT data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'user_privacies' AND column_name = 'visible_statistics'`).Scan(&dataType).Error; err != nil {
		t.Fatalf("read jsonb type: %v", err)
	}
	if dataType != "jsonb" {
		t.Fatalf("visible_statistics type = %q, want jsonb", dataType)
	}
	var tileDataType string
	if err := db.Raw(`SELECT data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'user_statistic_tiles' AND column_name = 'selected_keys'`).Scan(&tileDataType).Error; err != nil || tileDataType != "jsonb" {
		t.Fatalf("selected_keys jsonb type/error = %q/%v", tileDataType, err)
	}
	user := models.User{Name: "Privacy User", Password: "Password123!", Us: "privacy-user", Email: "privacy-user@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.UserPrivacy{UserID: user.ID, Mode: "invalid"}).Error; err == nil {
		t.Fatal("mode check accepted invalid value")
	}
	if err := db.Create(&models.UserPrivacy{UserID: user.ID + 1000, Mode: string(users.PrivacyModeOpen)}).Error; err == nil {
		t.Fatal("foreign key accepted a missing user")
	}
	defaultUser := models.User{Name: "Default Privacy", Password: "Password123!", Us: "default-privacy", Email: "default-privacy@example.com"}
	if err := db.Create(&defaultUser).Error; err != nil {
		t.Fatalf("create default user: %v", err)
	}
	if err := db.Create(&models.UserPrivacy{UserID: defaultUser.ID, VisibleStatistics: []string{}}).Error; err != nil {
		t.Fatalf("create default privacy row: %v", err)
	}
	var defaultRow models.UserPrivacy
	if err := db.First(&defaultRow, "user_id = ?", defaultUser.ID).Error; err != nil || defaultRow.Mode != string(users.PrivacyModeOpen) {
		t.Fatalf("mode default/error = %q/%v", defaultRow.Mode, err)
	}
	store := users.NewGORMProfileStore(&testPostgresRepository{db: db})
	defaults, err := store.GetTileSettings(context.Background(), user.ID)
	if err != nil || !reflect.DeepEqual(defaults, users.DefaultTileSettings()) {
		t.Fatalf("default tiles/error = %#v/%v", defaults, err)
	}
	defaultPrivacy, err := store.GetPrivacy(context.Background(), user.ID)
	if err != nil || defaultPrivacy.Mode != users.PrivacyModeOpen || len(defaultPrivacy.VisibleStatistics) != len(users.AllStatisticFields()) {
		t.Fatalf("default privacy/error = %#v/%v", defaultPrivacy, err)
	}
	mode := users.PrivacyModeCustom
	visible := []users.StatisticField{users.StatisticCompletedEvents}
	var group sync.WaitGroup
	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	errs := make(chan error, 2)
	group.Add(2)
	go func() {
		defer group.Done()
		ready <- struct{}{}
		<-start
		_, err := store.PatchPrivacy(context.Background(), user.ID, users.PrivacyPatch{Mode: &mode})
		errs <- err
	}()
	go func() {
		defer group.Done()
		ready <- struct{}{}
		<-start
		_, err := store.PatchPrivacy(context.Background(), user.ID, users.PrivacyPatch{VisibleStatistics: &visible})
		errs <- err
	}()
	<-ready
	<-ready
	close(start)
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent patch: %v", err)
		}
	}
	privacy, err := store.GetPrivacy(context.Background(), user.ID)
	if err != nil || privacy.Mode != users.PrivacyModeCustom || len(privacy.VisibleStatistics) != 1 || privacy.VisibleStatistics[0] != users.StatisticCompletedEvents {
		t.Fatalf("concurrent result/error = %#v/%v", privacy, err)
	}
	first := users.TileSettings{Tiles: []users.TileKey{users.TileFilms, users.TileGames, users.TileCompletedEvents}}
	second := users.TileSettings{Tiles: []users.TileKey{users.TileMostFrequentWeekday, users.TilePlannedMinutes}}
	readyTiles := make(chan struct{}, 2)
	startTiles := make(chan struct{})
	tileErrors := make(chan error, 2)
	group.Add(2)
	go func() {
		defer group.Done()
		readyTiles <- struct{}{}
		<-startTiles
		_, err := store.ReplaceTileSettings(context.Background(), user.ID, first)
		tileErrors <- err
	}()
	go func() {
		defer group.Done()
		readyTiles <- struct{}{}
		<-startTiles
		_, err := store.ReplaceTileSettings(context.Background(), user.ID, second)
		tileErrors <- err
	}()
	<-readyTiles
	<-readyTiles
	close(startTiles)
	group.Wait()
	close(tileErrors)
	for err := range tileErrors {
		if err != nil {
			t.Fatalf("concurrent tile replacement: %v", err)
		}
	}
	persistedTiles, err := store.GetTileSettings(context.Background(), user.ID)
	if err != nil || (!reflect.DeepEqual(persistedTiles, first) && !reflect.DeepEqual(persistedTiles, second)) {
		t.Fatalf("atomic tile result/error = %#v/%v", persistedTiles, err)
	}
	if err := db.Delete(&user).Error; err != nil {
		t.Fatalf("delete user for cascade: %v", err)
	}
	var privacyRows, tileRows int64
	if err := db.Model(&models.UserPrivacy{}).Where("user_id = ?", user.ID).Count(&privacyRows).Error; err != nil || privacyRows != 0 {
		t.Fatalf("cascade rows/error = %d/%v", privacyRows, err)
	}
	if err := db.Model(&models.UserStatisticTiles{}).Where("user_id = ?", user.ID).Count(&tileRows).Error; err != nil || tileRows != 0 {
		t.Fatalf("tile cascade rows/error = %d/%v", tileRows, err)
	}
}
