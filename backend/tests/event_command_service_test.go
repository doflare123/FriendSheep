package tests

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
)

type eventCommandStoreStub struct {
	snapshot        servicesevents.EventCommandSnapshot
	findEventErr    error
	role            string
	findRoleErr     error
	eventTypeExists bool
	eventTypeErr    error
	locationExists  bool
	locationErr     error
	ageLimitExists  bool
	ageLimitErr     error
	genreCount      int
	genreCountErr   error
	createID        uint
	createErr       error
	updateErr       error
	deleteErr       error
	view            servicesevents.EventCommandView
	loadErr         error
	calls           []string
	roleActorID     uint
	roleGroupID     uint
	eventTypeID     uint
	locationID      uint
	ageLimitID      uint
	genreIDs        []uint
	createRecords   []servicesevents.EventCreateRecord
	updateRecords   []servicesevents.EventUpdateRecord
	deletedIDs      []uint
	loadedEventID   uint
	loadedActorID   uint
}

func (s *eventCommandStoreStub) record(call string) {
	s.calls = append(s.calls, call)
}

func (s *eventCommandStoreStub) FindEventForUpdate(eventID uint) (servicesevents.EventCommandSnapshot, error) {
	s.record("find_event")
	if s.snapshot.ID == 0 {
		s.snapshot.ID = eventID
	}
	return s.snapshot, s.findEventErr
}

func (s *eventCommandStoreStub) FindGroupRole(actorID uint, groupID uint) (string, error) {
	s.record("find_group_role")
	s.roleActorID = actorID
	s.roleGroupID = groupID
	return s.role, s.findRoleErr
}

func (s *eventCommandStoreStub) EventTypeExists(eventTypeID uint) (bool, error) {
	s.record("event_type_exists")
	s.eventTypeID = eventTypeID
	return s.eventTypeExists, s.eventTypeErr
}

func (s *eventCommandStoreStub) EventLocationExists(locationID uint) (bool, error) {
	s.record("event_location_exists")
	s.locationID = locationID
	return s.locationExists, s.locationErr
}

func (s *eventCommandStoreStub) AgeLimitExists(ageLimitID uint) (bool, error) {
	s.record("age_limit_exists")
	s.ageLimitID = ageLimitID
	return s.ageLimitExists, s.ageLimitErr
}

func (s *eventCommandStoreStub) CountGenres(genreIDs []uint) (int, error) {
	s.record("count_genres")
	s.genreIDs = append([]uint(nil), genreIDs...)
	return s.genreCount, s.genreCountErr
}

func (s *eventCommandStoreStub) CreateEvent(record servicesevents.EventCreateRecord) (uint, error) {
	s.record("create_event")
	s.createRecords = append(s.createRecords, record)
	return s.createID, s.createErr
}

func (s *eventCommandStoreStub) UpdateEvent(record servicesevents.EventUpdateRecord) error {
	s.record("update_event")
	s.updateRecords = append(s.updateRecords, record)
	return s.updateErr
}

func (s *eventCommandStoreStub) DeleteEventAggregate(eventID uint) error {
	s.record("delete_event")
	s.deletedIDs = append(s.deletedIDs, eventID)
	return s.deleteErr
}

func (s *eventCommandStoreStub) LoadEventResult(eventID uint, actorID uint) (servicesevents.EventCommandView, error) {
	s.record("load_event")
	s.loadedEventID = eventID
	s.loadedActorID = actorID
	return s.view, s.loadErr
}

func newEventCommandStoreStub() *eventCommandStoreStub {
	return &eventCommandStoreStub{
		snapshot: servicesevents.EventCommandSnapshot{
			ID:           77,
			GroupID:      42,
			Title:        "Existing event",
			StartTime:    time.Now().Add(24 * time.Hour),
			Duration:     120,
			CurrentUsers: 2,
		},
		role:            groupmodels.RoleAdmin,
		eventTypeExists: true,
		locationExists:  true,
		ageLimitExists:  true,
		genreCount:      2,
		createID:        77,
		view: servicesevents.EventCommandView{
			ID:           77,
			Title:        "Command event",
			CurrentUsers: 1,
			MaxUsers:     10,
			Group:        servicesevents.EventCommandGroupView{ID: 42, Name: "Command group"},
			Creator:      servicesevents.EventCommandCreatorView{ID: 5, Name: "Actor"},
		},
	}
}

func validCreateEventCommandInput() servicesevents.CreateEventInput {
	year := 2026
	return servicesevents.CreateEventInput{
		Title:        "Command event",
		Description:  "Long command event description",
		GroupID:      42,
		EventTypeID:  3,
		LocationID:   4,
		ImageURL:     "https://example.com/command.png",
		StartTime:    time.Now().Add(48 * time.Hour).Truncate(time.Second),
		Duration:     90,
		MaxUsers:     10,
		Genres:       []uint{8, 9},
		Address:      "Command address",
		Country:      "RU",
		AgeLimitID:   6,
		Year:         &year,
		Notes:        "Command notes",
		CustomFields: map[string]interface{}{"mode": "table"},
	}
}

func eventCommandPointer[T any](value T) *T {
	return &value
}

func TestEventCommandServiceCreateEventPreservesContextAndDelegatesExactInput(t *testing.T) {
	store := newEventCommandStoreStub()
	audit := &eventAuditStoreStub{record: func(servicesevents.EventAuditInput) {
		store.record("audit")
	}}
	uow := &eventUnitOfWorkStub{commands: store, audit: audit}
	service := servicesevents.NewEventCommandService(&testLogger{}, uow)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "command-context")
	input := validCreateEventCommandInput()
	startedAt := time.Now()

	result, err := service.CreateEvent(ctx, 5, input)
	finishedAt := time.Now()

	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if result == nil || result.ID != 77 || result.Title != store.view.Title {
		t.Fatalf("result = %#v, want command view", result)
	}
	if uow.calls != 1 || uow.ctx != ctx {
		t.Fatalf("unit of work calls = %d, context preserved = %v", uow.calls, uow.ctx == ctx)
	}
	wantCalls := []string{"find_group_role", "age_limit_exists", "count_genres", "create_event", "audit", "load_event"}
	if !reflect.DeepEqual(store.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", store.calls, wantCalls)
	}
	if store.roleActorID != 5 || store.roleGroupID != input.GroupID || store.ageLimitID != input.AgeLimitID {
		t.Fatalf("lookup inputs = actor:%d group:%d age:%d", store.roleActorID, store.roleGroupID, store.ageLimitID)
	}
	if !reflect.DeepEqual(store.genreIDs, input.Genres) {
		t.Fatalf("genre IDs = %v, want %v", store.genreIDs, input.Genres)
	}
	if len(store.createRecords) != 1 {
		t.Fatalf("create records = %d, want 1", len(store.createRecords))
	}
	record := store.createRecords[0]
	if record.Title != input.Title || record.Description != input.Description || record.GroupID != input.GroupID ||
		record.EventTypeID != input.EventTypeID || record.LocationID != input.LocationID || record.CreatorID != 5 ||
		record.StartTime != input.StartTime || record.EndTime != input.StartTime.Add(90*time.Minute) ||
		record.Duration != input.Duration || record.MaxUsers != input.MaxUsers || record.CurrentUsers != 1 ||
		record.ImageURL != input.ImageURL || record.StatusID != 1 || record.Address != input.Address ||
		record.Country != input.Country || record.AgeLimitID != input.AgeLimitID || record.Year != input.Year ||
		record.Notes != input.Notes || !reflect.DeepEqual(record.CustomFields, input.CustomFields) ||
		!reflect.DeepEqual(record.GenreIDs, input.Genres) {
		t.Fatalf("create record = %#v, input = %#v", record, input)
	}
	if record.CreatorJoinedAt.Before(startedAt) || record.CreatorJoinedAt.After(finishedAt) {
		t.Fatalf("creator joined at = %s, want within command interval", record.CreatorJoinedAt)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.GroupID != input.GroupID || entry.ActorID != 5 || entry.Action != groupmodels.ActionCreateEvent ||
		entry.EntityID == nil || *entry.EntityID != 77 || entry.EntityName != input.Title ||
		!entry.CreatedAt.Equal(record.CreatorJoinedAt) || entry.TargetUserID != nil {
		t.Fatalf("audit entry = %#v", entry)
	}
	if store.loadedEventID != 77 || store.loadedActorID != 5 {
		t.Fatalf("load inputs = event:%d actor:%d", store.loadedEventID, store.loadedActorID)
	}
}

func TestEventCommandServiceUpdateEventPreservesContextAndDelegatesExactInput(t *testing.T) {
	store := newEventCommandStoreStub()
	store.view.Title = "Updated command event"
	audit := &eventAuditStoreStub{record: func(servicesevents.EventAuditInput) {
		store.record("audit")
	}}
	uow := &eventUnitOfWorkStub{commands: store, audit: audit}
	service := servicesevents.NewEventCommandService(&testLogger{}, uow)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "update-context")
	title := "Updated command event"
	startTime := time.Now().Add(72 * time.Hour).Truncate(time.Second)
	duration := uint16(180)
	maxUsers := uint16(12)
	eventTypeID := uint(3)
	locationID := uint(4)
	ageLimitID := uint(7)
	input := servicesevents.UpdateEventInput{
		Title:       &title,
		EventTypeID: &eventTypeID,
		LocationID:  &locationID,
		StartTime:   &startTime,
		Duration:    &duration,
		MaxUsers:    &maxUsers,
		AgeLimit:    &ageLimitID,
		Genres:      []uint{10, 11},
	}

	result, err := service.UpdateEvent(ctx, 5, 77, input)

	if err != nil {
		t.Fatalf("UpdateEvent returned error: %v", err)
	}
	if result == nil || result.ID != 77 || result.Title != title {
		t.Fatalf("result = %#v, want updated command view", result)
	}
	if uow.calls != 1 || uow.ctx != ctx {
		t.Fatalf("unit of work calls = %d, context preserved = %v", uow.calls, uow.ctx == ctx)
	}
	wantCalls := []string{
		"find_event",
		"find_group_role",
		"event_type_exists",
		"event_location_exists",
		"age_limit_exists",
		"count_genres",
		"update_event",
		"audit",
		"load_event",
	}
	if !reflect.DeepEqual(store.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", store.calls, wantCalls)
	}
	if store.eventTypeID != eventTypeID || store.locationID != locationID {
		t.Fatalf("reference lookup IDs = event type:%d location:%d, want %d and %d", store.eventTypeID, store.locationID, eventTypeID, locationID)
	}
	if len(store.updateRecords) != 1 {
		t.Fatalf("update records = %d, want 1", len(store.updateRecords))
	}
	record := store.updateRecords[0]
	wantEndTime := startTime.Add(time.Duration(duration) * time.Minute)
	if record.EventID != 77 || record.Title != input.Title || record.EventTypeID != input.EventTypeID ||
		record.LocationID != input.LocationID || record.StartTime != input.StartTime ||
		record.Duration != input.Duration || record.MaxUsers != input.MaxUsers || record.AgeLimitID != input.AgeLimit ||
		!record.ReplaceGenres || !reflect.DeepEqual(record.GenreIDs, input.Genres) ||
		record.EndTime == nil || !record.EndTime.Equal(wantEndTime) {
		t.Fatalf("update record = %#v, input = %#v", record, input)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.GroupID != 42 || entry.ActorID != 5 || entry.Action != groupmodels.ActionUpdateEvent ||
		entry.EntityID == nil || *entry.EntityID != 77 || entry.EntityName != title ||
		entry.CreatedAt.IsZero() || entry.TargetUserID != nil {
		t.Fatalf("audit entry = %#v", entry)
	}
}

func TestEventCommandServiceUpdateEventRejectsInvalidFieldsWithoutSideEffects(t *testing.T) {
	tests := []struct {
		name    string
		input   servicesevents.UpdateEventInput
		wantErr error
	}{
		{
			name:  "empty update",
			input: servicesevents.UpdateEventInput{},
		},
		{
			name:  "title shorter than five trimmed runes",
			input: servicesevents.UpdateEventInput{Title: eventCommandPointer("  four  ")},
		},
		{
			name:  "title longer than two hundred runes",
			input: servicesevents.UpdateEventInput{Title: eventCommandPointer(strings.Repeat("я", 201))},
		},
		{
			name:  "description shorter than ten trimmed runes",
			input: servicesevents.UpdateEventInput{Description: eventCommandPointer(" 123456789 ")},
		},
		{
			name:  "description longer than two thousand runes",
			input: servicesevents.UpdateEventInput{Description: eventCommandPointer(strings.Repeat("я", 2001))},
		},
		{
			name:  "zero event type",
			input: servicesevents.UpdateEventInput{EventTypeID: eventCommandPointer(uint(0))},
		},
		{
			name:  "zero location",
			input: servicesevents.UpdateEventInput{LocationID: eventCommandPointer(uint(0))},
		},
		{
			name:  "relative image URL",
			input: servicesevents.UpdateEventInput{ImageURL: eventCommandPointer("/events/image.png")},
		},
		{
			name:  "image URL without host",
			input: servicesevents.UpdateEventInput{ImageURL: eventCommandPointer("mailto:friend@example.com")},
		},
		{
			name:  "zero start time",
			input: servicesevents.UpdateEventInput{StartTime: eventCommandPointer(time.Time{})},
		},
		{
			name:  "past start time",
			input: servicesevents.UpdateEventInput{StartTime: eventCommandPointer(time.Now().Add(-time.Minute))},
		},
		{
			name:  "duration below fifteen minutes",
			input: servicesevents.UpdateEventInput{Duration: eventCommandPointer(uint16(14))},
		},
		{
			name:  "duration above one day",
			input: servicesevents.UpdateEventInput{Duration: eventCommandPointer(uint16(1441))},
		},
		{
			name:  "max users below two",
			input: servicesevents.UpdateEventInput{MaxUsers: eventCommandPointer(uint16(1))},
		},
		{
			name:  "max users above one thousand",
			input: servicesevents.UpdateEventInput{MaxUsers: eventCommandPointer(uint16(1001))},
		},
		{
			name:    "empty genres",
			input:   servicesevents.UpdateEventInput{Genres: []uint{}},
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name:    "more than nine genres",
			input:   servicesevents.UpdateEventInput{Genres: []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name:    "zero genre ID",
			input:   servicesevents.UpdateEventInput{Genres: []uint{0}},
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name:    "duplicate genre ID",
			input:   servicesevents.UpdateEventInput{Genres: []uint{1, 1}},
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name:  "zero age limit",
			input: servicesevents.UpdateEventInput{AgeLimit: eventCommandPointer(uint(0))},
		},
		{
			name:  "address longer than five hundred runes",
			input: servicesevents.UpdateEventInput{Address: eventCommandPointer(strings.Repeat("я", 501))},
		},
		{
			name:  "country longer than one hundred runes",
			input: servicesevents.UpdateEventInput{Country: eventCommandPointer(strings.Repeat("я", 101))},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventCommandStoreStub()
			audit := &eventAuditStoreStub{}
			service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
				commands: store,
				audit:    audit,
			})

			result, err := service.UpdateEvent(context.Background(), 5, 77, test.input)

			if result != nil {
				t.Fatalf("result = %#v, want nil", result)
			}
			wantErr := test.wantErr
			if wantErr == nil {
				wantErr = servicesevents.ErrInvalidEventUpdate
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("err = %v, want %v", err, wantErr)
			}
			if len(store.updateRecords) != 0 {
				t.Fatalf("update records = %#v, want none", store.updateRecords)
			}
			if len(audit.entries) != 0 {
				t.Fatalf("audit entries = %#v, want none", audit.entries)
			}
			for _, call := range store.calls {
				if call == "update_event" || call == "audit" || call == "load_event" {
					t.Fatalf("side-effect call after validation error: %v", store.calls)
				}
			}
		})
	}
}

func TestEventCommandServiceUpdateEventAcceptsValidationBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		input servicesevents.UpdateEventInput
	}{
		{
			name:  "title minimum",
			input: servicesevents.UpdateEventInput{Title: eventCommandPointer(strings.Repeat("я", 5))},
		},
		{
			name:  "description minimum",
			input: servicesevents.UpdateEventInput{Description: eventCommandPointer(strings.Repeat("я", 10))},
		},
		{
			name:  "absolute image URL",
			input: servicesevents.UpdateEventInput{ImageURL: eventCommandPointer("https://example.com/events/image.png")},
		},
		{
			name:  "duration minimum",
			input: servicesevents.UpdateEventInput{Duration: eventCommandPointer(uint16(15))},
		},
		{
			name:  "duration maximum",
			input: servicesevents.UpdateEventInput{Duration: eventCommandPointer(uint16(1440))},
		},
		{
			name:  "max users minimum",
			input: servicesevents.UpdateEventInput{MaxUsers: eventCommandPointer(uint16(2))},
		},
		{
			name:  "max users maximum",
			input: servicesevents.UpdateEventInput{MaxUsers: eventCommandPointer(uint16(1000))},
		},
		{
			name:  "address maximum",
			input: servicesevents.UpdateEventInput{Address: eventCommandPointer(strings.Repeat("я", 500))},
		},
		{
			name:  "country maximum",
			input: servicesevents.UpdateEventInput{Country: eventCommandPointer(strings.Repeat("я", 100))},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventCommandStoreStub()
			audit := &eventAuditStoreStub{record: func(servicesevents.EventAuditInput) {
				store.record("audit")
			}}
			service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
				commands: store,
				audit:    audit,
			})

			result, err := service.UpdateEvent(context.Background(), 5, 77, test.input)

			if err != nil {
				t.Fatalf("UpdateEvent returned error: %v", err)
			}
			if result == nil || result.ID != store.view.ID {
				t.Fatalf("result = %#v, want event %d", result, store.view.ID)
			}
			if len(store.updateRecords) != 1 {
				t.Fatalf("update records = %#v, want one", store.updateRecords)
			}
			if len(audit.entries) != 1 {
				t.Fatalf("audit entries = %#v, want one", audit.entries)
			}
		})
	}
}

func TestEventCommandServiceDeleteEventPreservesContextAndDelegatesExactInput(t *testing.T) {
	store := newEventCommandStoreStub()
	audit := &eventAuditStoreStub{record: func(servicesevents.EventAuditInput) {
		store.record("audit")
	}}
	uow := &eventUnitOfWorkStub{commands: store, audit: audit}
	service := servicesevents.NewEventCommandService(&testLogger{}, uow)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "delete-context")

	deleted, err := service.DeleteEvent(ctx, 5, 77)

	if err != nil {
		t.Fatalf("DeleteEvent returned error: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteEvent returned false")
	}
	if uow.calls != 1 || uow.ctx != ctx {
		t.Fatalf("unit of work calls = %d, context preserved = %v", uow.calls, uow.ctx == ctx)
	}
	wantCalls := []string{"find_event", "find_group_role", "delete_event", "audit"}
	if !reflect.DeepEqual(store.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", store.calls, wantCalls)
	}
	if !reflect.DeepEqual(store.deletedIDs, []uint{77}) {
		t.Fatalf("deleted IDs = %v, want [77]", store.deletedIDs)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.GroupID != 42 || entry.ActorID != 5 || entry.Action != groupmodels.ActionDeleteEvent ||
		entry.EntityID == nil || *entry.EntityID != 77 || entry.EntityName != store.snapshot.Title ||
		entry.CreatedAt.IsZero() || entry.TargetUserID != nil {
		t.Fatalf("audit entry = %#v", entry)
	}
}

func TestEventCommandServiceDomainErrorsDoNotMutateOrAudit(t *testing.T) {
	eventTypeLookupErr := errors.New("event type lookup failed")
	locationLookupErr := errors.New("event location lookup failed")

	type commandRun func(servicesevents.EventCommandService) (bool, error)
	createRun := func(service servicesevents.EventCommandService) (bool, error) {
		result, err := service.CreateEvent(context.Background(), 5, validCreateEventCommandInput())
		return result != nil, err
	}
	updateRun := func(input servicesevents.UpdateEventInput) commandRun {
		return func(service servicesevents.EventCommandService) (bool, error) {
			result, err := service.UpdateEvent(context.Background(), 5, 77, input)
			return result != nil, err
		}
	}
	deleteRun := func(service servicesevents.EventCommandService) (bool, error) {
		return service.DeleteEvent(context.Background(), 5, 77)
	}

	tests := []struct {
		name    string
		prepare func(*eventCommandStoreStub)
		run     commandRun
		wantErr error
	}{
		{
			name: "create non-member",
			prepare: func(store *eventCommandStoreStub) {
				store.findRoleErr = servicesevents.ErrNotGroupMember
			},
			run:     createRun,
			wantErr: servicesevents.ErrNotGroupMember,
		},
		{
			name: "create permission denied",
			prepare: func(store *eventCommandStoreStub) {
				store.role = groupmodels.RoleMember
			},
			run:     createRun,
			wantErr: servicesevents.ErrPermissionDenied,
		},
		{
			name: "create empty genres",
			prepare: func(store *eventCommandStoreStub) {
				store.genreCount = 0
			},
			run: func(service servicesevents.EventCommandService) (bool, error) {
				input := validCreateEventCommandInput()
				input.Genres = []uint{}
				result, err := service.CreateEvent(context.Background(), 5, input)
				return result != nil, err
			},
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name: "create too many genres",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				input := validCreateEventCommandInput()
				input.Genres = []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
				result, err := service.CreateEvent(context.Background(), 5, input)
				return result != nil, err
			},
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name: "create missing age limit",
			prepare: func(store *eventCommandStoreStub) {
				store.ageLimitExists = false
			},
			run:     createRun,
			wantErr: servicesevents.ErrAgeLimitNotFound,
		},
		{
			name: "create missing genre",
			prepare: func(store *eventCommandStoreStub) {
				store.genreCount = 1
			},
			run:     createRun,
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name: "update event not found",
			prepare: func(store *eventCommandStoreStub) {
				store.findEventErr = servicesevents.ErrEventNotFound
			},
			run:     updateRun(servicesevents.UpdateEventInput{}),
			wantErr: servicesevents.ErrEventNotFound,
		},
		{
			name: "update non-member",
			prepare: func(store *eventCommandStoreStub) {
				store.findRoleErr = servicesevents.ErrNotGroupMember
			},
			run:     updateRun(servicesevents.UpdateEventInput{}),
			wantErr: servicesevents.ErrNotGroupMember,
		},
		{
			name: "update permission denied",
			prepare: func(store *eventCommandStoreStub) {
				store.role = groupmodels.RoleMember
			},
			run:     updateRun(servicesevents.UpdateEventInput{}),
			wantErr: servicesevents.ErrPermissionDenied,
		},
		{
			name: "update started event",
			prepare: func(store *eventCommandStoreStub) {
				store.snapshot.StartTime = time.Now().Add(-time.Minute)
			},
			run:     updateRun(servicesevents.UpdateEventInput{}),
			wantErr: servicesevents.ErrEventAlreadyStarted,
		},
		{
			name: "update max users below current",
			prepare: func(store *eventCommandStoreStub) {
				store.snapshot.CurrentUsers = 4
			},
			run: func(service servicesevents.EventCommandService) (bool, error) {
				maxUsers := uint16(3)
				result, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{MaxUsers: &maxUsers})
				return result != nil, err
			},
			wantErr: servicesevents.ErrMaxUsersBelowCurrent,
		},
		{
			name:    "update empty genres",
			run:     updateRun(servicesevents.UpdateEventInput{Genres: []uint{}}),
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name:    "update too many genres",
			run:     updateRun(servicesevents.UpdateEventInput{Genres: []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}),
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name: "update missing age limit",
			prepare: func(store *eventCommandStoreStub) {
				store.ageLimitExists = false
			},
			run: func(service servicesevents.EventCommandService) (bool, error) {
				ageLimitID := uint(99)
				result, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{AgeLimit: &ageLimitID})
				return result != nil, err
			},
			wantErr: servicesevents.ErrAgeLimitNotFound,
		},
		{
			name: "update missing event type",
			prepare: func(store *eventCommandStoreStub) {
				store.eventTypeExists = false
			},
			run:     updateRun(servicesevents.UpdateEventInput{EventTypeID: eventCommandPointer(uint(99))}),
			wantErr: servicesevents.ErrEventTypeNotFound,
		},
		{
			name: "update event type lookup error",
			prepare: func(store *eventCommandStoreStub) {
				store.eventTypeErr = eventTypeLookupErr
			},
			run:     updateRun(servicesevents.UpdateEventInput{EventTypeID: eventCommandPointer(uint(3))}),
			wantErr: eventTypeLookupErr,
		},
		{
			name: "update missing location",
			prepare: func(store *eventCommandStoreStub) {
				store.locationExists = false
			},
			run:     updateRun(servicesevents.UpdateEventInput{LocationID: eventCommandPointer(uint(99))}),
			wantErr: servicesevents.ErrEventLocationNotFound,
		},
		{
			name: "update location lookup error",
			prepare: func(store *eventCommandStoreStub) {
				store.locationErr = locationLookupErr
			},
			run:     updateRun(servicesevents.UpdateEventInput{LocationID: eventCommandPointer(uint(4))}),
			wantErr: locationLookupErr,
		},
		{
			name: "update missing genre",
			prepare: func(store *eventCommandStoreStub) {
				store.genreCount = 1
			},
			run:     updateRun(servicesevents.UpdateEventInput{Genres: []uint{10, 11}}),
			wantErr: servicesevents.ErrInvalidGenres,
		},
		{
			name: "delete event not found",
			prepare: func(store *eventCommandStoreStub) {
				store.findEventErr = servicesevents.ErrEventNotFound
			},
			run:     deleteRun,
			wantErr: servicesevents.ErrEventNotFound,
		},
		{
			name: "delete non-member",
			prepare: func(store *eventCommandStoreStub) {
				store.findRoleErr = servicesevents.ErrNotGroupMember
			},
			run:     deleteRun,
			wantErr: servicesevents.ErrNotGroupMember,
		},
		{
			name: "delete permission denied",
			prepare: func(store *eventCommandStoreStub) {
				store.role = groupmodels.RoleMember
			},
			run:     deleteRun,
			wantErr: servicesevents.ErrPermissionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventCommandStoreStub()
			if test.prepare != nil {
				test.prepare(store)
			}
			audit := &eventAuditStoreStub{}
			service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
				commands: store,
				audit:    audit,
			})

			succeeded, err := test.run(service)

			if succeeded {
				t.Fatal("command unexpectedly succeeded")
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}
			if len(store.createRecords) != 0 || len(store.updateRecords) != 0 || len(store.deletedIDs) != 0 {
				t.Fatalf("mutations occurred: create=%d update=%d delete=%d", len(store.createRecords), len(store.updateRecords), len(store.deletedIDs))
			}
			if len(audit.entries) != 0 {
				t.Fatalf("audit entries = %#v, want none", audit.entries)
			}
			for _, call := range store.calls {
				if call == "load_event" {
					t.Fatalf("load called after domain error: %v", store.calls)
				}
			}
		})
	}
}

func TestEventCommandServiceReturnsUnitOfWorkErrors(t *testing.T) {
	injectedErr := errors.New("injected command unit of work failure")
	tests := []struct {
		name string
		run  func(servicesevents.EventCommandService) (bool, error)
	}{
		{
			name: "create",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				result, err := service.CreateEvent(context.Background(), 5, validCreateEventCommandInput())
				return result != nil, err
			},
		},
		{
			name: "update",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				result, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{})
				return result != nil, err
			},
		},
		{
			name: "delete",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				return service.DeleteEvent(context.Background(), 5, 77)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uow := &eventUnitOfWorkStub{err: injectedErr}
			service := servicesevents.NewEventCommandService(&testLogger{}, uow)

			succeeded, err := test.run(service)

			if succeeded {
				t.Fatal("command unexpectedly succeeded")
			}
			if !errors.Is(err, injectedErr) {
				t.Fatalf("err = %v, want injected failure", err)
			}
			if uow.calls != 1 {
				t.Fatalf("unit of work calls = %d, want 1", uow.calls)
			}
		})
	}
}

func TestEventCommandServiceIgnoresAuditFailures(t *testing.T) {
	injectedErr := errors.New("injected command audit failure")
	tests := []struct {
		name string
		run  func(servicesevents.EventCommandService) (bool, error)
	}{
		{
			name: "create",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				result, err := service.CreateEvent(context.Background(), 5, validCreateEventCommandInput())
				return result != nil, err
			},
		},
		{
			name: "update",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				title := "Updated despite audit failure"
				result, err := service.UpdateEvent(context.Background(), 5, 77, servicesevents.UpdateEventInput{Title: &title})
				return result != nil, err
			},
		},
		{
			name: "delete",
			run: func(service servicesevents.EventCommandService) (bool, error) {
				return service.DeleteEvent(context.Background(), 5, 77)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventCommandStoreStub()
			audit := &eventAuditStoreStub{recordErr: injectedErr}
			service := servicesevents.NewEventCommandService(&testLogger{}, &eventUnitOfWorkStub{
				commands: store,
				audit:    audit,
			})

			succeeded, err := test.run(service)

			if err != nil {
				t.Fatalf("command returned audit error: %v", err)
			}
			if !succeeded {
				t.Fatal("command returned unsuccessful result")
			}
			if len(audit.entries) != 1 {
				t.Fatalf("audit entries = %d, want 1", len(audit.entries))
			}
		})
	}
}
