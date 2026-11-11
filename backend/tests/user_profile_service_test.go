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
}

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
	if err := db.AutoMigrate(&models.User{}, &models.UserPrivacy{}); err != nil {
		t.Fatal(err)
	}
	return db
}

var _ users.ProfileStore = (*profileStoreStub)(nil)
var _ users.StatisticsReader = (*profileStatisticsStub)(nil)
