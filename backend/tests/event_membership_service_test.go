package tests

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	groupmodels "friendship/models/groups"
	servicesevents "friendship/services/events"
)

type eventMembershipStoreStub struct {
	event             servicesevents.EventMembershipSnapshot
	findEventErr      error
	groupMember       bool
	groupMemberErr    error
	participant       bool
	participantErr    error
	addErr            error
	incremented       bool
	incrementErr      error
	removeErr         error
	decrementErr      error
	calls             []string
	addedUserID       uint
	addedEventID      uint
	removedUserID     uint
	removedEventID    uint
	decrementedEvent  uint
	incrementedEvent  uint
	participantUserID uint
	participantEvent  uint
}

func (s *eventMembershipStoreStub) record(call string) {
	s.calls = append(s.calls, call)
}

func (s *eventMembershipStoreStub) FindEvent(_ uint) (servicesevents.EventMembershipSnapshot, error) {
	s.record("find_event")
	return s.event, s.findEventErr
}

func (s *eventMembershipStoreStub) IsGroupMember(_ uint, _ uint) (bool, error) {
	s.record("is_group_member")
	return s.groupMember, s.groupMemberErr
}

func (s *eventMembershipStoreStub) IsParticipant(userID uint, eventID uint) (bool, error) {
	s.record("is_participant")
	s.participantUserID = userID
	s.participantEvent = eventID
	return s.participant, s.participantErr
}

func (s *eventMembershipStoreStub) AddParticipant(userID uint, eventID uint, _ time.Time) error {
	s.record("add_participant")
	s.addedUserID = userID
	s.addedEventID = eventID
	return s.addErr
}

func (s *eventMembershipStoreStub) IncrementParticipantsIfSpace(eventID uint) (bool, error) {
	s.record("increment_participants")
	s.incrementedEvent = eventID
	return s.incremented, s.incrementErr
}

func (s *eventMembershipStoreStub) RemoveParticipant(userID uint, eventID uint) error {
	s.record("remove_participant")
	s.removedUserID = userID
	s.removedEventID = eventID
	return s.removeErr
}

func (s *eventMembershipStoreStub) DecrementParticipants(eventID uint) error {
	s.record("decrement_participants")
	s.decrementedEvent = eventID
	return s.decrementErr
}

type eventAuditStoreStub struct {
	recordErr error
	entries   []servicesevents.EventAuditInput
	record    func(servicesevents.EventAuditInput)
}

func (s *eventAuditStoreStub) RecordBestEffort(entry servicesevents.EventAuditInput) error {
	s.entries = append(s.entries, entry)
	if s.record != nil {
		s.record(entry)
	}
	return s.recordErr
}

type eventUnitOfWorkStub struct {
	membership *eventMembershipStoreStub
	commands   servicesevents.EventCommandStore
	admin      servicesevents.EventAdminStore
	audit      *eventAuditStoreStub
	err        error
	ctx        context.Context
	calls      int
}

func (u *eventUnitOfWorkStub) WithinTransaction(ctx context.Context, fn func(servicesevents.EventTransaction) error) error {
	u.calls++
	u.ctx = ctx
	if u.err != nil {
		return u.err
	}
	return fn(servicesevents.NewEventTransaction(servicesevents.EventTransactionStores{
		MembershipStore: u.membership,
		CommandStore:    u.commands,
		AdminStore:      u.admin,
		AuditStore:      u.audit,
	}))
}

func TestEventMembershipServiceJoinEventDelegatesToCleanUnitOfWork(t *testing.T) {
	membership := &eventMembershipStoreStub{
		event: servicesevents.EventMembershipSnapshot{
			ID:        77,
			GroupID:   42,
			CreatorID: 1,
			Title:     "Clean membership event",
			StartTime: time.Now().Add(24 * time.Hour),
		},
		groupMember: true,
		incremented: true,
	}
	audit := &eventAuditStoreStub{}
	uow := &eventUnitOfWorkStub{membership: membership, audit: audit}
	service := servicesevents.NewEventMembershipService(&testLogger{}, uow)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "membership-context")

	joined, err := service.JoinEvent(ctx, 2, 77)

	if err != nil {
		t.Fatalf("JoinEvent returned error: %v", err)
	}
	if !joined {
		t.Fatal("JoinEvent returned false")
	}
	if uow.calls != 1 || uow.ctx != ctx {
		t.Fatalf("unit of work calls = %d, context preserved = %v", uow.calls, uow.ctx == ctx)
	}
	wantCalls := []string{"find_event", "is_group_member", "is_participant", "add_participant", "increment_participants"}
	if !reflect.DeepEqual(membership.calls, wantCalls) {
		t.Fatalf("membership calls = %v, want %v", membership.calls, wantCalls)
	}
	if membership.addedUserID != 2 || membership.addedEventID != 77 || membership.incrementedEvent != 77 {
		t.Fatalf("membership arguments: add=(%d,%d), increment=%d", membership.addedUserID, membership.addedEventID, membership.incrementedEvent)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.Action != groupmodels.ActionJoinEvent || entry.GroupID != 42 || entry.ActorID != 2 || entry.EntityID == nil || *entry.EntityID != 77 {
		t.Fatalf("audit entry = %#v", entry)
	}
}

func TestEventMembershipServiceJoinEventRejectsStartedEventWithoutSideEffects(t *testing.T) {
	membership := &eventMembershipStoreStub{
		event: servicesevents.EventMembershipSnapshot{
			ID:        77,
			GroupID:   42,
			CreatorID: 1,
			Title:     "Already started event",
			StartTime: time.Now().Add(-time.Hour),
		},
		groupMember: true,
		incremented: true,
	}
	audit := &eventAuditStoreStub{}
	uow := &eventUnitOfWorkStub{membership: membership, audit: audit}
	service := servicesevents.NewEventMembershipService(&testLogger{}, uow)

	joined, err := service.JoinEvent(context.Background(), 2, 77)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrEventAlreadyStarted) {
		t.Fatalf("JoinEvent err = %v, want ErrEventAlreadyStarted", err)
	}
	if wantCalls := []string{"find_event", "is_group_member", "is_participant"}; !reflect.DeepEqual(membership.calls, wantCalls) {
		t.Fatalf("membership calls = %v, want %v", membership.calls, wantCalls)
	}
	if len(audit.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(audit.entries))
	}
}

func TestEventMembershipServiceJoinEventPreservesMembershipErrorPrecedenceForStartedEvent(t *testing.T) {
	tests := []struct {
		name       string
		membership *eventMembershipStoreStub
		wantErr    error
		wantCalls  []string
	}{
		{
			name: "non group member",
			membership: &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					StartTime: time.Now().Add(-time.Hour),
				},
			},
			wantErr:   servicesevents.ErrNotGroupMember,
			wantCalls: []string{"find_event", "is_group_member"},
		},
		{
			name: "existing participant",
			membership: &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					StartTime: time.Now().Add(-time.Hour),
				},
				groupMember: true,
				participant: true,
			},
			wantErr:   servicesevents.ErrAlreadyJoined,
			wantCalls: []string{"find_event", "is_group_member", "is_participant"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := servicesevents.NewEventMembershipService(
				&testLogger{},
				&eventUnitOfWorkStub{membership: test.membership},
			)

			joined, err := service.JoinEvent(context.Background(), 2, 77)

			if joined {
				t.Fatal("JoinEvent returned true")
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("JoinEvent err = %v, want %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(test.membership.calls, test.wantCalls) {
				t.Fatalf("membership calls = %v, want %v", test.membership.calls, test.wantCalls)
			}
		})
	}
}

func TestEventMembershipServiceLeaveEventDelegatesToCleanUnitOfWork(t *testing.T) {
	membership := &eventMembershipStoreStub{event: servicesevents.EventMembershipSnapshot{
		ID:        77,
		GroupID:   42,
		CreatorID: 1,
		Title:     "Clean membership event",
	}}
	audit := &eventAuditStoreStub{}
	uow := &eventUnitOfWorkStub{membership: membership, audit: audit}
	service := servicesevents.NewEventMembershipService(&testLogger{}, uow)
	ctx := context.Background()

	left, err := service.LeaveEvent(ctx, 2, 77)

	if err != nil {
		t.Fatalf("LeaveEvent returned error: %v", err)
	}
	if !left {
		t.Fatal("LeaveEvent returned false")
	}
	wantCalls := []string{"find_event", "remove_participant", "decrement_participants"}
	if !reflect.DeepEqual(membership.calls, wantCalls) {
		t.Fatalf("membership calls = %v, want %v", membership.calls, wantCalls)
	}
	if membership.removedUserID != 2 || membership.removedEventID != 77 || membership.decrementedEvent != 77 {
		t.Fatalf("membership arguments: remove=(%d,%d), decrement=%d", membership.removedUserID, membership.removedEventID, membership.decrementedEvent)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != groupmodels.ActionLeaveEvent {
		t.Fatalf("audit entries = %#v", audit.entries)
	}
}

func TestEventMembershipServiceIgnoresAuditFailures(t *testing.T) {
	injectedErr := errors.New("injected audit failure")
	tests := []struct {
		name  string
		join  bool
		audit *eventAuditStoreStub
	}{
		{
			name:  "join audit record",
			join:  true,
			audit: &eventAuditStoreStub{recordErr: injectedErr},
		},
		{
			name:  "leave audit record",
			audit: &eventAuditStoreStub{recordErr: injectedErr},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			membership := &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					StartTime: time.Now().Add(24 * time.Hour),
				},
				groupMember: true,
				incremented: true,
			}
			uow := &eventUnitOfWorkStub{membership: membership, audit: test.audit}
			service := servicesevents.NewEventMembershipService(&testLogger{}, uow)

			var success bool
			var err error
			if test.join {
				success, err = service.JoinEvent(context.Background(), 2, 77)
			} else {
				success, err = service.LeaveEvent(context.Background(), 2, 77)
			}
			if err != nil {
				t.Fatalf("membership operation returned audit error: %v", err)
			}
			if !success {
				t.Fatal("membership operation returned false")
			}
		})
	}
}

func TestEventMembershipServiceDomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		membership *eventMembershipStoreStub
		join       bool
		userID     uint
		wantErr    error
	}{
		{
			name:       "join missing event",
			membership: &eventMembershipStoreStub{findEventErr: servicesevents.ErrEventNotFound},
			join:       true,
			userID:     2,
			wantErr:    servicesevents.ErrEventNotFound,
		},
		{
			name: "join non group member",
			membership: &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					StartTime: time.Now().Add(24 * time.Hour),
				},
				groupMember: false,
			},
			join:    true,
			userID:  2,
			wantErr: servicesevents.ErrNotGroupMember,
		},
		{
			name: "join existing participant",
			membership: &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					StartTime: time.Now().Add(24 * time.Hour),
				},
				groupMember: true,
				participant: true,
			},
			join:    true,
			userID:  2,
			wantErr: servicesevents.ErrAlreadyJoined,
		},
		{
			name: "join full event",
			membership: &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					StartTime: time.Now().Add(24 * time.Hour),
				},
				groupMember: true,
				incremented: false,
			},
			join:    true,
			userID:  2,
			wantErr: servicesevents.ErrEventFull,
		},
		{
			name: "creator leave",
			membership: &eventMembershipStoreStub{
				event: servicesevents.EventMembershipSnapshot{ID: 77, GroupID: 42, CreatorID: 1},
			},
			userID:  1,
			wantErr: servicesevents.ErrCreatorCantLeave,
		},
		{
			name: "not joined leave",
			membership: &eventMembershipStoreStub{
				event:     servicesevents.EventMembershipSnapshot{ID: 77, GroupID: 42, CreatorID: 1},
				removeErr: servicesevents.ErrNotJoined,
			},
			userID:  2,
			wantErr: servicesevents.ErrNotJoined,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uow := &eventUnitOfWorkStub{membership: test.membership}
			service := servicesevents.NewEventMembershipService(&testLogger{}, uow)

			var success bool
			var err error
			if test.join {
				success, err = service.JoinEvent(context.Background(), test.userID, 77)
			} else {
				success, err = service.LeaveEvent(context.Background(), test.userID, 77)
			}
			if success {
				t.Fatal("membership operation returned true")
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestEventMembershipServiceReturnsUnitOfWorkError(t *testing.T) {
	injectedErr := errors.New("injected unit of work failure")
	uow := &eventUnitOfWorkStub{err: injectedErr}
	service := servicesevents.NewEventMembershipService(&testLogger{}, uow)

	joined, err := service.JoinEvent(context.Background(), 2, 77)

	if joined {
		t.Fatal("JoinEvent returned true")
	}
	if !errors.Is(err, injectedErr) {
		t.Fatalf("JoinEvent err = %v, want injected failure", err)
	}
}
