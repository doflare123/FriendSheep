package tests

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"friendship/models"
	"friendship/services/references"
	"friendship/services/statistics"
	"friendship/services/users"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type profileStoreStub struct {
	user       users.UserRecord
	findTarget users.Target
	findCalls  int
	privacy    users.Privacy
	patch      users.PrivacyPatch
	privacyErr error
	patchErr   error
	tiles      users.TileSettings
	references []references.ReferenceItem
	refCalls   int
}

func (s *profileStoreStub) FindUser(_ context.Context, target users.Target) (users.UserRecord, error) {
	s.findTarget = target
	s.findCalls++
	return s.user, s.privacyErr
}

func (s *profileStoreStub) GetPrivacy(_ context.Context, _ uint) (users.Privacy, error) {
	return s.privacy, s.privacyErr
}

func (s *profileStoreStub) PatchPrivacy(_ context.Context, _ uint, patch users.PrivacyPatch) (users.Privacy, error) {
	s.patch = patch
	return s.privacy, s.patchErr
}

func (s *profileStoreStub) GetTileSettings(_ context.Context, _ uint) (users.TileSettings, error) {
	return s.tiles, s.privacyErr
}

func (s *profileStoreStub) ReplaceTileSettings(_ context.Context, _ uint, settings users.TileSettings) (users.TileSettings, error) {
	s.tiles = settings
	return settings, s.patchErr
}

func (s *profileStoreStub) FindEventTypesByNames(_ context.Context, _ []string) ([]references.ReferenceItem, error) {
	s.refCalls++
	return s.references, s.privacyErr
}

type profileStatisticsStub struct {
	view  statistics.View
	calls int
	user  uint
}

func (s *profileStatisticsStub) Read(_ context.Context, userID uint) (statistics.View, error) {
	s.calls++
	s.user = userID
	return s.view, nil
}

func TestProfileParseIdentifierSeparatesIDAndAddress(t *testing.T) {
	for _, test := range []struct {
		raw  string
		want users.Target
		ok   bool
	}{
		{raw: "123", want: users.IDTarget(123), ok: true},
		{raw: "@123", want: users.UsTarget("123"), ok: true},
		{raw: "@alice", want: users.UsTarget("alice"), ok: true},
		{raw: "0"}, {raw: "-1"}, {raw: "alice"}, {raw: "@"}, {raw: " 12"},
	} {
		t.Run(test.raw, func(t *testing.T) {
			got, err := users.ParseIdentifier(test.raw)
			if test.ok {
				if err != nil || got != test.want {
					t.Fatalf("ParseIdentifier(%q) = %#v, %v; want %#v, nil", test.raw, got, err, test.want)
				}
				return
			}
			if !errors.Is(err, users.ErrInvalidTarget) {
				t.Fatalf("ParseIdentifier(%q) error = %v, want ErrInvalidTarget", test.raw, err)
			}
		})
	}
}

func TestProfileServiceOwnerByResolvedIDBypassesPrivacyAndClosedAvoidsStatistics(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	store := &profileStoreStub{user: users.UserRecord{ID: 7, Name: "Alice", Us: "alice", CreatedAt: now, Privacy: users.Privacy{Mode: users.PrivacyModeClosed}}}
	stats := &profileStatisticsStub{view: statistics.View{CompletedEvents: 3, Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	service := users.NewProfileService(store, stats)

	owner, err := service.GetProfile(context.Background(), 7, users.UsTarget("alice"))
	if err != nil || !owner.IsOwner || stats.calls != 1 || store.findTarget != users.UsTarget("alice") || owner.Statistics == nil || owner.Statistics.CompletedEvents == nil || *owner.Statistics.CompletedEvents != 3 {
		t.Fatalf("owner result/error = %#v/%v, store=%#v stats=%#v", owner, err, store, stats)
	}

	_, err = service.GetProfile(context.Background(), 8, users.IDTarget(7))
	if !errors.Is(err, users.ErrProfileClosed) || stats.calls != 1 {
		t.Fatalf("closed result/error/statistics calls = %v/%d, want closed/1", err, stats.calls)
	}
}

func TestProfileServiceCustomSelectionMasksCompactStatistics(t *testing.T) {
	unique := int64(4)
	store := &profileStoreStub{user: users.UserRecord{ID: 9, Privacy: users.Privacy{Mode: users.PrivacyModeCustom, VisibleStatistics: []users.StatisticField{users.StatisticUniquePeopleCount, users.StatisticCompletedEvents}}}}
	stats := &profileStatisticsStub{view: statistics.View{CompletedEvents: 2, PlannedMinutes: 120, UniquePeopleCount: &unique, UniquePeopleStatus: "fresh", Types: []statistics.CountItem{{ID: 4, Count: 1}}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	profile, err := users.NewProfileService(store, stats).GetProfile(context.Background(), 7, users.IDTarget(9))
	if err != nil || profile.Statistics == nil || profile.Statistics.CompletedEvents == nil || *profile.Statistics.CompletedEvents != 2 || profile.Statistics.UniquePeopleCount == nil || profile.Statistics.UniquePeopleCount.Value == nil || *profile.Statistics.UniquePeopleCount.Value != 4 || profile.Statistics.PlannedMinutes != nil || profile.Statistics.MostFrequentGenreID != nil {
		t.Fatalf("profile/error = %#v/%v", profile, err)
	}
}

func TestProfileServiceRejectsDuplicateAndUnknownPrivacyFields(t *testing.T) {
	store := &profileStoreStub{}
	service := users.NewProfileService(store, nil)
	duplicate := []users.StatisticField{users.StatisticCompletedEvents, users.StatisticCompletedEvents}
	unknown := []users.StatisticField{"secret"}
	for _, patch := range []users.PrivacyPatch{{VisibleStatistics: &duplicate}, {VisibleStatistics: &unknown}, {}} {
		if _, err := service.PatchPrivacy(context.Background(), 7, patch); !errors.Is(err, users.ErrInvalidPrivacy) {
			t.Fatalf("PatchPrivacy(%#v) error = %v, want ErrInvalidPrivacy", patch, err)
		}
	}
}

func TestProfileServiceDoesNotLeakHiddenStatisticsAndSkipsEmptyCustomSelection(t *testing.T) {
	store := &profileStoreStub{user: users.UserRecord{ID: 7, Name: "Alice", Us: "alice", Privacy: users.Privacy{Mode: users.PrivacyModeCustom, VisibleStatistics: []users.StatisticField{users.StatisticCompletedEvents}}}}
	stats := &profileStatisticsStub{view: statistics.View{CompletedEvents: 3, PlannedMinutes: 60, UniquePeopleStatus: "fresh", Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	profile, err := users.NewProfileService(store, stats).GetProfile(context.Background(), 8, users.IDTarget(7))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"uniquePeopleCount", "uniquePeopleStatus", "uniquePeopleCalculatedAt", "plannedMinutes", "email", "password", "plans", "history", "types", "genres"} {
		if strings.Contains(string(encoded), `"`+forbidden+`"`) {
			t.Fatalf("profile leaked %q: %s", forbidden, encoded)
		}
	}
	if !strings.Contains(string(encoded), `"completedEvents":3`) {
		t.Fatalf("selected field absent: %s", encoded)
	}

	store.user.Privacy.VisibleStatistics = []users.StatisticField{}
	stats.calls = 0
	profile, err = users.NewProfileService(store, stats).GetProfile(context.Background(), 8, users.IDTarget(7))
	if err != nil || stats.calls != 0 || profile.Statistics == nil {
		t.Fatalf("empty custom result/calls/error = %#v/%d/%v", profile, stats.calls, err)
	}
}

func TestProfileServiceOwnerGetsFullCompactStatisticsForEveryTarget(t *testing.T) {
	store := &profileStoreStub{user: users.UserRecord{ID: 7, Us: "alice", Privacy: users.Privacy{Mode: users.PrivacyModeClosed}}}
	stats := &profileStatisticsStub{view: statistics.View{CompletedEvents: 2, PlannedMinutes: 30, Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	service := users.NewProfileService(store, stats)
	for _, target := range []users.Target{users.OwnTarget(), users.IDTarget(7), users.UsTarget("alice")} {
		profile, err := service.GetProfile(context.Background(), 7, target)
		if err != nil || !profile.IsOwner || profile.Statistics == nil || profile.Statistics.CompletedEvents == nil || profile.Statistics.PlannedMinutes == nil {
			t.Fatalf("target %#v profile/error = %#v/%v", target, profile, err)
		}
	}
}

func TestGORMProfileStoreDefaultsAndPartialPrivacyPatch(t *testing.T) {
	db := newProfileStoreDB(t)
	user := models.User{ID: 7, Name: "Alice", Password: "Password123!", Us: "alice", Email: "alice@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	store := users.NewGORMProfileStore(&testPostgresRepository{db: db})
	defaultPrivacy, err := store.GetPrivacy(context.Background(), user.ID)
	if err != nil || !reflect.DeepEqual(defaultPrivacy, users.DefaultPrivacy()) {
		t.Fatalf("default privacy/error = %#v/%v", defaultPrivacy, err)
	}
	custom := users.PrivacyModeCustom
	visible := []users.StatisticField{users.StatisticCompletedEvents}
	privacy, err := store.PatchPrivacy(context.Background(), user.ID, users.PrivacyPatch{Mode: &custom, VisibleStatistics: &visible})
	if err != nil || privacy.Mode != custom || !reflect.DeepEqual(privacy.VisibleStatistics, visible) {
		t.Fatalf("first patch/error = %#v/%v", privacy, err)
	}
	closed := users.PrivacyModeClosed
	privacy, err = store.PatchPrivacy(context.Background(), user.ID, users.PrivacyPatch{Mode: &closed})
	if err != nil || privacy.Mode != closed || !reflect.DeepEqual(privacy.VisibleStatistics, visible) {
		t.Fatalf("partial patch/error = %#v/%v", privacy, err)
	}
	if _, err := store.GetPrivacy(context.Background(), 999); !errors.Is(err, users.ErrUserNotFound) {
		t.Fatalf("unknown user error = %v, want ErrUserNotFound", err)
	}
	defaults, err := store.GetTileSettings(context.Background(), user.ID)
	if err != nil || !reflect.DeepEqual(defaults, users.DefaultTileSettings()) {
		t.Fatalf("default tiles/error = %#v/%v", defaults, err)
	}
	chosen := users.TileSettings{Tiles: []users.TileKey{users.TileGames, users.TileFilms, users.TileCompletedEvents}}
	persisted, err := store.ReplaceTileSettings(context.Background(), user.ID, chosen)
	if err != nil || !reflect.DeepEqual(persisted, chosen) {
		t.Fatalf("replace tiles/error = %#v/%v", persisted, err)
	}
	persisted, err = store.GetTileSettings(context.Background(), user.ID)
	if err != nil || !reflect.DeepEqual(persisted, chosen) {
		t.Fatalf("read tiles/error = %#v/%v", persisted, err)
	}
	empty := users.TileSettings{Tiles: []users.TileKey{}}
	if _, err := store.ReplaceTileSettings(context.Background(), user.ID, empty); err != nil {
		t.Fatalf("replace empty tiles: %v", err)
	}
	persisted, err = store.GetTileSettings(context.Background(), user.ID)
	if err != nil || persisted.Tiles == nil || len(persisted.Tiles) != 0 {
		t.Fatalf("empty tile roundtrip/error = %#v/%v", persisted, err)
	}
}

func TestProfileServiceTilesPreserveSelectionOrderAndMapAllValues(t *testing.T) {
	unique := int64(9)
	typeID, genreID, weekday := uint(42), uint(77), 3
	keys := []users.TileKey{users.TileGames, users.TileFilms, users.TileTableGames, users.TileOther, users.TileCompletedEvents, users.TilePlannedMinutes, users.TileCreatedCompletedEvents, users.TileUniquePeopleCount, users.TileMostFrequentTypeID, users.TileMostFrequentGenreID, users.TileMostFrequentWeekday}
	store := &profileStoreStub{
		user:       users.UserRecord{ID: 7, Privacy: users.DefaultPrivacy()},
		tiles:      users.TileSettings{Tiles: keys},
		references: []references.ReferenceItem{{ID: 101, Name: "Игры"}, {ID: 202, Name: "Фильмы"}, {ID: 303, Name: "Настольные игры"}, {ID: 404, Name: "Другое"}},
	}
	stats := &profileStatisticsStub{view: statistics.View{CompletedEvents: 5, PlannedMinutes: 120, CreatedCompletedEvents: 2, UniquePeopleCount: &unique, UniquePeopleStatus: "fresh", Types: []statistics.CountItem{{ID: 101, Count: 11}, {ID: 202, Count: 12}, {ID: 303, Count: 13}, {ID: 404, Count: 14}}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}, MostFrequentTypeID: &typeID, MostFrequentGenreID: &genreID, MostFrequentWeekday: &weekday}}
	wantValues := []*int64{int64Pointer(11), int64Pointer(12), int64Pointer(13), int64Pointer(14), int64Pointer(5), int64Pointer(120), int64Pointer(2), int64Pointer(9), int64Pointer(42), int64Pointer(77), int64Pointer(3)}
	service := users.NewProfileService(store, stats)
	for first := 0; first < len(keys); first += users.MaxStatisticTiles {
		last := first + users.MaxStatisticTiles
		if last > len(keys) {
			last = len(keys)
		}
		store.tiles = users.TileSettings{Tiles: keys[first:last]}
		profile, err := service.GetProfile(context.Background(), 7, users.OwnTarget())
		if err != nil || len(profile.Tiles) != last-first {
			t.Fatalf("batch %d profile/error = %#v/%v", first, profile, err)
		}
		for i, tile := range profile.Tiles {
			wantIndex := first + i
			if tile.Key != keys[wantIndex] || tile.Value == nil || *tile.Value != *wantValues[wantIndex] {
				t.Fatalf("tile %d = %#v, want key/value %s/%d", wantIndex, tile, keys[wantIndex], *wantValues[wantIndex])
			}
		}
	}
	store.tiles = users.TileSettings{Tiles: []users.TileKey{users.TilePlannedMinutes, users.TileUniquePeopleCount}}
	profile, err := service.GetProfile(context.Background(), 7, users.OwnTarget())
	if err != nil || profile.Tiles[0].Unit != "minutes" || profile.Tiles[1].Status == nil || *profile.Tiles[1].Status != "fresh" {
		t.Fatalf("tile metadata/error = %#v/%v", profile.Tiles, err)
	}
}

func TestProfileServiceTileSettingsValidateOrderAndCustomPrivacyFiltersTiles(t *testing.T) {
	store := &profileStoreStub{user: users.UserRecord{ID: 7, Privacy: users.Privacy{Mode: users.PrivacyModeCustom, VisibleStatistics: []users.StatisticField{users.StatisticGames, users.StatisticCompletedEvents}}}, tiles: users.TileSettings{Tiles: []users.TileKey{users.TileFilms, users.TileGames, users.TileCompletedEvents, users.TilePlannedMinutes}}, references: []references.ReferenceItem{{ID: 1, Name: "Игры"}}}
	stats := &profileStatisticsStub{view: statistics.View{Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	profile, err := users.NewProfileService(store, stats).GetProfile(context.Background(), 8, users.IDTarget(7))
	if err != nil || len(profile.Tiles) != 2 || profile.Tiles[0].Key != users.TileGames || profile.Tiles[1].Key != users.TileCompletedEvents || stats.calls != 1 {
		t.Fatalf("custom profile/error = %#v/%v", profile, err)
	}
	for _, settings := range []users.TileSettings{{}, {Tiles: []users.TileKey{users.TileFilms, users.TileFilms}}, {Tiles: []users.TileKey{"unknown"}}, {Tiles: []users.TileKey{users.TileFilms, users.TileGames, users.TileTableGames, users.TileOther, users.TileCompletedEvents}}} {
		if _, err := users.NewProfileService(store, stats).ReplaceTileSettings(context.Background(), 7, settings); !errors.Is(err, users.ErrInvalidTileSettings) {
			t.Fatalf("settings %#v error = %v", settings, err)
		}
	}
	want := users.TileSettings{Tiles: []users.TileKey{users.TileTableGames, users.TileFilms}}
	got, err := users.NewProfileService(store, stats).ReplaceTileSettings(context.Background(), 7, want)
	if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(store.tiles, want) {
		t.Fatalf("replacement = %#v/%#v/%v", got, store.tiles, err)
	}
}

func TestProfileServiceReturnsEffectiveTileSettings(t *testing.T) {
	keys := []users.TileKey{users.TilePlannedMinutes, users.TileGames, users.TileCompletedEvents}
	store := &profileStoreStub{
		user:       users.UserRecord{ID: 7, Privacy: users.DefaultPrivacy()},
		tiles:      users.TileSettings{Tiles: keys},
		references: []references.ReferenceItem{{ID: 1, Name: "Игры"}},
	}
	stats := &profileStatisticsStub{view: statistics.View{Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	service := users.NewProfileService(store, stats)

	owner, err := service.GetProfile(context.Background(), 7, users.OwnTarget())
	if err != nil || !reflect.DeepEqual(owner.TileSettings.Tiles, keys) {
		t.Fatalf("owner tile settings/error = %#v/%v", owner.TileSettings, err)
	}

	openProfile, err := service.GetProfile(context.Background(), 8, users.IDTarget(7))
	if err != nil || !reflect.DeepEqual(openProfile.TileSettings.Tiles, keys) {
		t.Fatalf("open tile settings/error = %#v/%v", openProfile.TileSettings, err)
	}

	store.user.Privacy = users.Privacy{Mode: users.PrivacyModeCustom, VisibleStatistics: []users.StatisticField{users.StatisticCompletedEvents, users.StatisticGames}}
	customProfile, err := service.GetProfile(context.Background(), 8, users.IDTarget(7))
	wantCustom := []users.TileKey{users.TileGames, users.TileCompletedEvents}
	if err != nil || !reflect.DeepEqual(customProfile.TileSettings.Tiles, wantCustom) || len(customProfile.Tiles) != len(wantCustom) {
		t.Fatalf("custom tile settings/tiles/error = %#v/%#v/%v", customProfile.TileSettings, customProfile.Tiles, err)
	}
	for i, tile := range customProfile.Tiles {
		if tile.Key != customProfile.TileSettings.Tiles[i] {
			t.Fatalf("visible tile key = %q, want %q", tile.Key, customProfile.TileSettings.Tiles[i])
		}
	}

	store.user.Privacy.VisibleStatistics = []users.StatisticField{}
	emptyProfile, err := service.GetProfile(context.Background(), 8, users.IDTarget(7))
	encoded, marshalErr := json.Marshal(emptyProfile)
	if err != nil || marshalErr != nil || !strings.Contains(string(encoded), `"tileSettings":{"tiles":[]}`) {
		t.Fatalf("empty tile settings/error/JSON = %#v/%v/%v/%s", emptyProfile.TileSettings, err, marshalErr, encoded)
	}

	store.user.Privacy = users.Privacy{Mode: users.PrivacyModeClosed}
	if _, err := service.GetProfile(context.Background(), 8, users.IDTarget(7)); !errors.Is(err, users.ErrProfileClosed) {
		t.Fatalf("closed profile error = %v, want ErrProfileClosed", err)
	}
}

func TestProfileServiceTilesPreserveNullUniqueMetadataAndAvoidUnusedReferenceLookup(t *testing.T) {
	store := &profileStoreStub{user: users.UserRecord{ID: 7, Privacy: users.DefaultPrivacy()}, tiles: users.TileSettings{Tiles: []users.TileKey{users.TileUniquePeopleCount, users.TileCompletedEvents}}}
	stats := &profileStatisticsStub{view: statistics.View{CompletedEvents: 2, UniquePeopleStatus: "pending", Types: []statistics.CountItem{}, Genres: []statistics.CountItem{}, Months: []statistics.MonthCount{}, Weekdays: []statistics.WeekdayCount{}}}
	profile, err := users.NewProfileService(store, stats).GetProfile(context.Background(), 7, users.OwnTarget())
	if err != nil || store.refCalls != 0 || len(profile.Tiles) != 2 || profile.Tiles[0].Value != nil || profile.Tiles[0].Status == nil || *profile.Tiles[0].Status != "pending" {
		t.Fatalf("unique tile/error/store = %#v/%v/%#v", profile.Tiles, err, store)
	}
	encoded, err := json.Marshal(profile.Tiles[0])
	if err != nil || !strings.Contains(string(encoded), `"value":null`) || !strings.Contains(string(encoded), `"status":"pending"`) {
		t.Fatalf("unique tile JSON/error = %s/%v", encoded, err)
	}
	store.user.Privacy = users.Privacy{Mode: users.PrivacyModeCustom, VisibleStatistics: []users.StatisticField{users.StatisticCompletedEvents}}
	profile, err = users.NewProfileService(store, stats).GetProfile(context.Background(), 8, users.IDTarget(7))
	if err != nil || len(profile.Tiles) != 1 || profile.Tiles[0].Key != users.TileCompletedEvents {
		t.Fatalf("filtered tile/error = %#v/%v", profile.Tiles, err)
	}
}

func int64Pointer(value int64) *int64 { return &value }

func newProfileStoreDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:profile-" + testUintString(uint(time.Now().UnixNano())) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserPrivacy{}, &models.UserStatisticTiles{}); err != nil {
		t.Fatal(err)
	}
	return db
}

var _ users.ProfileStore = (*profileStoreStub)(nil)
var _ users.StatisticsReader = (*profileStatisticsStub)(nil)
