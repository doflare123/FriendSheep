package tests

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	eventmodels "friendship/models/events"
	"friendship/services/notifications"
)

type reminderRecipientStoreStub struct {
	snapshot notifications.EventReminderSnapshot
	err      error
	calls    int
	eventID  uint
}

func (s *reminderRecipientStoreStub) LoadEventReminderSnapshot(_ context.Context, eventID uint) (notifications.EventReminderSnapshot, error) {
	s.calls++
	s.eventID = eventID
	return s.snapshot, s.err
}

type reminderPreferenceStub struct {
	enabledOffset int
	calls         []reminderPreferenceCall
	err           error
}

type reminderPreferenceCall struct {
	userID  uint
	eventID uint
	offset  int
}

func (s *reminderPreferenceStub) ReadEventReminderPreference(_ context.Context, userID uint, eventID uint, offset int) (notifications.EventReminderPreferenceDecision, error) {
	s.calls = append(s.calls, reminderPreferenceCall{userID: userID, eventID: eventID, offset: offset})
	if s.err != nil {
		return notifications.EventReminderPreferenceDecision{}, s.err
	}
	return notifications.EventReminderPreferenceDecision{
		Enabled:  offset == s.enabledOffset,
		Channels: []string{notifications.ChannelInApp},
	}, nil
}

func TestDefaultEventReminderPreferenceEnablesEveryCanonicalOffsetForInAppOnly(t *testing.T) {
	reader := notifications.NewDefaultEventReminderPreferenceReader()
	for _, offset := range []int{1440, 360, 60} {
		decision, err := reader.ReadEventReminderPreference(context.Background(), 7, 42, offset)
		if err != nil {
			t.Fatalf("offset %d returned error: %v", offset, err)
		}
		if !decision.Enabled || !reflect.DeepEqual(decision.Channels, []string{notifications.ChannelInApp}) {
			t.Fatalf("offset %d decision = %#v, want enabled in_app only", offset, decision)
		}
	}
	custom, err := reader.ReadEventReminderPreference(context.Background(), 7, 42, 30)
	if err != nil || custom.Enabled || len(custom.Channels) != 0 {
		t.Fatalf("non-canonical positive offset decision = %#v err:%v, want disabled/no channels", custom, err)
	}
	if _, err := reader.ReadEventReminderPreference(context.Background(), 7, 42, 0); !errors.Is(err, notifications.ErrInvalidReminderOffset) {
		t.Fatalf("non-positive offset error = %v, want ErrInvalidReminderOffset", err)
	}
}

func TestEventReminderRecipientServiceDeduplicatesAndAppliesOffsetPreference(t *testing.T) {
	start := time.Date(2036, 8, 30, 18, 0, 0, 0, time.UTC)
	store := &reminderRecipientStoreStub{snapshot: notifications.EventReminderSnapshot{
		EventID: 42, Title: "Reminder event", StartTime: start, UserIDs: []uint{7, 7, 0, 9, 9},
	}}
	preferences := &reminderPreferenceStub{enabledOffset: 360}
	service := notifications.NewEventReminderRecipientService(store, preferences)

	enabled, err := service.Resolve(context.Background(), 42, 360)
	if err != nil {
		t.Fatalf("Resolve(360) returned error: %v", err)
	}
	if enabled.EventID != 42 || enabled.Title != "Reminder event" || !enabled.StartTime.Equal(start) {
		t.Fatalf("event snapshot = %#v", enabled)
	}
	wantRecipients := []notifications.EventReminderRecipient{
		{UserID: 7, Channels: []string{notifications.ChannelInApp}},
		{UserID: 9, Channels: []string{notifications.ChannelInApp}},
	}
	if !reflect.DeepEqual(enabled.Recipients, wantRecipients) {
		t.Fatalf("6h recipients = %#v, want %#v", enabled.Recipients, wantRecipients)
	}
	if len(preferences.calls) != 2 {
		t.Fatalf("preference calls = %#v, want one per unique non-zero user", preferences.calls)
	}

	preferences.calls = nil
	disabled, err := service.Resolve(context.Background(), 42, 1440)
	if err != nil {
		t.Fatalf("Resolve(1440) returned error: %v", err)
	}
	if disabled.Recipients == nil || len(disabled.Recipients) != 0 {
		t.Fatalf("disabled offset recipients = %#v, want stable empty slice", disabled.Recipients)
	}
}

func TestGORMEventReminderRecipientResolutionUsesCurrentParticipantsAtTriggerTime(t *testing.T) {
	db := newEventsServiceDB(t)
	seedEventUser(t, db, 1)
	seedEventUser(t, db, 2)
	seedEventUser(t, db, 3)
	groupID := seedEventGroup(t, db, 1, false)
	eventID := seedEvent(t, db, groupID, 1, 2, 5)
	seedEventParticipant(t, db, eventID, 1)
	seedEventParticipant(t, db, eventID, 2)
	service := notifications.NewEventReminderRecipientService(
		notifications.NewGORMEventReminderRecipientStore(&testPostgresRepository{db: db}),
		notifications.NewDefaultEventReminderPreferenceReader(),
	)

	first, err := service.Resolve(context.Background(), eventID, 360)
	if err != nil {
		t.Fatalf("first Resolve() returned error: %v", err)
	}
	assertReminderRecipientIDs(t, first.Recipients, []uint{1, 2})

	if err := db.Where("event_id = ? AND user_id = ?", eventID, 2).Delete(&eventmodels.EventsUser{}).Error; err != nil {
		t.Fatalf("remove participant before reminder trigger: %v", err)
	}
	seedEventParticipant(t, db, eventID, 3)

	second, err := service.Resolve(context.Background(), eventID, 360)
	if err != nil {
		t.Fatalf("second Resolve() returned error: %v", err)
	}
	assertReminderRecipientIDs(t, second.Recipients, []uint{1, 3})
}

func TestGORMEventReminderRecipientResolutionReturnsTypedNotFound(t *testing.T) {
	db := newEventsServiceDB(t)
	service := notifications.NewEventReminderRecipientService(
		notifications.NewGORMEventReminderRecipientStore(&testPostgresRepository{db: db}),
		notifications.NewDefaultEventReminderPreferenceReader(),
	)
	if _, err := service.Resolve(context.Background(), 999, 60); !errors.Is(err, notifications.ErrEventNotFound) {
		t.Fatalf("Resolve() err = %v, want ErrEventNotFound", err)
	}
}

func assertReminderRecipientIDs(t *testing.T, recipients []notifications.EventReminderRecipient, want []uint) {
	t.Helper()
	got := make([]uint, 0, len(recipients))
	for _, recipient := range recipients {
		got = append(got, recipient.UserID)
		if !reflect.DeepEqual(recipient.Channels, []string{notifications.ChannelInApp}) {
			t.Fatalf("recipient %d channels = %v, want in_app only", recipient.UserID, recipient.Channels)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("recipient IDs = %v, want %v", got, want)
	}
}

var _ notifications.EventReminderRecipientStore = (*reminderRecipientStoreStub)(nil)
var _ notifications.EventReminderPreferenceReader = (*reminderPreferenceStub)(nil)
