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

type eventAdminReaderStub struct {
	access       servicesevents.EventAdminAccessView
	accessErr    error
	details      servicesevents.EventAdminDetailsView
	detailsErr   error
	calls        []string
	inspectCtx   context.Context
	loadCtx      context.Context
	inspectActor uint
	inspectEvent uint
	loadActor    uint
	loadEvent    uint
}

func (s *eventAdminReaderStub) InspectAccess(ctx context.Context, actorID uint, eventID uint) (servicesevents.EventAdminAccessView, error) {
	s.calls = append(s.calls, "inspect_access")
	s.inspectCtx = ctx
	s.inspectActor = actorID
	s.inspectEvent = eventID
	return s.access, s.accessErr
}

func (s *eventAdminReaderStub) LoadDetails(ctx context.Context, actorID uint, eventID uint) (servicesevents.EventAdminDetailsView, error) {
	s.calls = append(s.calls, "load_details")
	s.loadCtx = ctx
	s.loadActor = actorID
	s.loadEvent = eventID
	return s.details, s.detailsErr
}

type eventAdminStoreStub struct {
	event              servicesevents.EventAdminEventSnapshot
	findEventErr       error
	role               string
	findRoleErr        error
	targetExists       bool
	userExistsErr      error
	removeErr          error
	decrementErr       error
	calls              []string
	findEventID        uint
	findRoleActorID    uint
	findRoleGroupID    uint
	userExistsID       uint
	removedUserID      uint
	removedEventID     uint
	decrementedEventID uint
}

func (s *eventAdminStoreStub) FindEvent(eventID uint) (servicesevents.EventAdminEventSnapshot, error) {
	s.calls = append(s.calls, "find_event")
	s.findEventID = eventID
	return s.event, s.findEventErr
}

func (s *eventAdminStoreStub) FindGroupRole(actorID uint, groupID uint) (string, error) {
	s.calls = append(s.calls, "find_group_role")
	s.findRoleActorID = actorID
	s.findRoleGroupID = groupID
	return s.role, s.findRoleErr
}

func (s *eventAdminStoreStub) UserExists(userID uint) (bool, error) {
	s.calls = append(s.calls, "user_exists")
	s.userExistsID = userID
	return s.targetExists, s.userExistsErr
}

func (s *eventAdminStoreStub) RemoveParticipant(userID uint, eventID uint) error {
	s.calls = append(s.calls, "remove_participant")
	s.removedUserID = userID
	s.removedEventID = eventID
	return s.removeErr
}

func (s *eventAdminStoreStub) DecrementParticipants(eventID uint) error {
	s.calls = append(s.calls, "decrement_participants")
	s.decrementedEventID = eventID
	return s.decrementErr
}

func TestEventAdminServiceGetEventDetailsChecksAccessBeforeLoadingDetails(t *testing.T) {
	reader := &eventAdminReaderStub{
		access: servicesevents.EventAdminAccessView{
			EventFound:         true,
			ActorIsGroupMember: true,
			ActorRole:          groupmodels.RoleMember,
		},
	}
	service := servicesevents.NewEventAdminService(&testLogger{}, reader, nil)

	eventDTO, err := service.GetEventDetailsForAdmin(context.Background(), 7, 77)

	if eventDTO != nil {
		t.Fatalf("eventDTO = %#v, want nil", eventDTO)
	}
	if !errors.Is(err, servicesevents.ErrPermissionDenied) {
		t.Fatalf("err = %v, want ErrPermissionDenied", err)
	}
	if !reflect.DeepEqual(reader.calls, []string{"inspect_access"}) {
		t.Fatalf("reader calls = %v, want only inspect_access", reader.calls)
	}
	if reader.inspectActor != 7 || reader.inspectEvent != 77 {
		t.Fatalf("inspect access args = (%d,%d), want (7,77)", reader.inspectActor, reader.inspectEvent)
	}
}

func TestEventAdminServiceGetEventDetailsMapsCleanView(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "admin-details")
	reader := &eventAdminReaderStub{
		access: servicesevents.EventAdminAccessView{
			EventFound:         true,
			ActorIsGroupMember: true,
			ActorRole:          groupmodels.RoleAdmin,
		},
		details: servicesevents.EventAdminDetailsView{
			Found: true,
			Access: servicesevents.EventAdminAccessView{
				EventFound:         true,
				ActorIsGroupMember: true,
				ActorRole:          groupmodels.RoleAdmin,
			},
			Event: servicesevents.EventFullView{
				ID:      77,
				Title:   "Admin details",
				Creator: servicesevents.EventReadCreatorView{ID: 5},
			},
			Participants: []servicesevents.EventAdminParticipantView{
				{ID: 11, UserID: 5, Name: "Creator", Username: "creator"},
				{ID: 12, UserID: 9, Name: "Member", Username: "member"},
			},
		},
	}
	service := servicesevents.NewEventAdminService(&testLogger{}, reader, nil)

	eventDTO, err := service.GetEventDetailsForAdmin(ctx, 5, 77)

	if err != nil {
		t.Fatalf("GetEventDetailsForAdmin returned error: %v", err)
	}
	if eventDTO == nil || eventDTO.ID != 77 || eventDTO.Title != "Admin details" {
		t.Fatalf("eventDTO = %#v, want mapped event", eventDTO)
	}
	if len(eventDTO.AllParticipants) != 2 {
		t.Fatalf("participants = %d, want 2", len(eventDTO.AllParticipants))
	}
	if !eventDTO.AllParticipants[0].IsCreator || eventDTO.AllParticipants[1].IsCreator {
		t.Fatalf("participants = %#v, want creator marker only on user 5", eventDTO.AllParticipants)
	}
	if !reflect.DeepEqual(reader.calls, []string{"inspect_access", "load_details"}) {
		t.Fatalf("reader calls = %v, want access then details", reader.calls)
	}
	if reader.inspectCtx != ctx || reader.loadCtx != ctx ||
		reader.inspectActor != 5 || reader.loadActor != 5 ||
		reader.inspectEvent != 77 || reader.loadEvent != 77 {
		t.Fatal("reader did not receive the original context and identifiers")
	}
}

func TestEventAdminServiceGetEventDetailsStopsOnAccessFailures(t *testing.T) {
	storageErr := errors.New("access storage failure")
	tests := []struct {
		name    string
		access  servicesevents.EventAdminAccessView
		readErr error
		wantErr error
	}{
		{
			name:    "storage error",
			readErr: storageErr,
			wantErr: storageErr,
		},
		{
			name:    "event not found",
			access:  servicesevents.EventAdminAccessView{},
			wantErr: servicesevents.ErrEventNotFound,
		},
		{
			name: "not group member",
			access: servicesevents.EventAdminAccessView{
				EventFound: true,
			},
			wantErr: servicesevents.ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &eventAdminReaderStub{access: tt.access, accessErr: tt.readErr}
			service := servicesevents.NewEventAdminService(&testLogger{}, reader, nil)

			eventDTO, err := service.GetEventDetailsForAdmin(context.Background(), 5, 77)

			if eventDTO != nil {
				t.Fatalf("eventDTO = %#v, want nil", eventDTO)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(reader.calls, []string{"inspect_access"}) {
				t.Fatalf("reader calls = %v, want only inspect_access", reader.calls)
			}
		})
	}
}

func TestEventAdminServiceGetEventDetailsHandlesDetailsFailureAndConcurrentDeletion(t *testing.T) {
	storageErr := errors.New("details storage failure")
	tests := []struct {
		name       string
		details    servicesevents.EventAdminDetailsView
		detailsErr error
		wantErr    error
	}{
		{
			name:       "storage error",
			detailsErr: storageErr,
			wantErr:    storageErr,
		},
		{
			name:    "event deleted after access check",
			details: servicesevents.EventAdminDetailsView{},
			wantErr: servicesevents.ErrEventNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &eventAdminReaderStub{
				access: servicesevents.EventAdminAccessView{
					EventFound:         true,
					ActorIsGroupMember: true,
					ActorRole:          groupmodels.RoleAdmin,
				},
				details:    tt.details,
				detailsErr: tt.detailsErr,
			}
			service := servicesevents.NewEventAdminService(&testLogger{}, reader, nil)

			eventDTO, err := service.GetEventDetailsForAdmin(context.Background(), 5, 77)

			if eventDTO != nil {
				t.Fatalf("eventDTO = %#v, want nil", eventDTO)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(reader.calls, []string{"inspect_access", "load_details"}) {
				t.Fatalf("reader calls = %v, want access then details", reader.calls)
			}
		})
	}
}

func TestEventAdminServiceGetEventDetailsRechecksAccessSnapshot(t *testing.T) {
	tests := []struct {
		name         string
		latestAccess servicesevents.EventAdminAccessView
		wantErr      error
	}{
		{
			name: "membership revoked",
			latestAccess: servicesevents.EventAdminAccessView{
				EventFound: true,
			},
			wantErr: servicesevents.ErrNotGroupMember,
		},
		{
			name: "role downgraded",
			latestAccess: servicesevents.EventAdminAccessView{
				EventFound:         true,
				ActorIsGroupMember: true,
				ActorRole:          groupmodels.RoleMember,
			},
			wantErr: servicesevents.ErrPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &eventAdminReaderStub{
				access: servicesevents.EventAdminAccessView{
					EventFound:         true,
					ActorIsGroupMember: true,
					ActorRole:          groupmodels.RoleAdmin,
				},
				details: servicesevents.EventAdminDetailsView{
					Found:  true,
					Access: tt.latestAccess,
					Event:  servicesevents.EventFullView{ID: 77},
				},
			}
			service := servicesevents.NewEventAdminService(&testLogger{}, reader, nil)

			eventDTO, err := service.GetEventDetailsForAdmin(context.Background(), 5, 77)

			if eventDTO != nil {
				t.Fatalf("eventDTO = %#v, want nil", eventDTO)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(reader.calls, []string{"inspect_access", "load_details"}) {
				t.Fatalf("reader calls = %v, want access then details", reader.calls)
			}
		})
	}
}

func TestEventAdminServiceKickUserFromEventUsesTransactionalAdminStore(t *testing.T) {
	adminStore := &eventAdminStoreStub{
		event: servicesevents.EventAdminEventSnapshot{
			ID:        77,
			GroupID:   42,
			CreatorID: 1,
			Title:     "Admin kick event",
		},
		role:         groupmodels.RoleAdmin,
		targetExists: true,
	}
	audit := &eventAuditStoreStub{}
	uow := &eventUnitOfWorkStub{admin: adminStore, audit: audit}
	service := servicesevents.NewEventAdminService(&testLogger{}, nil, uow)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "admin-context")

	kicked, err := service.KickUserFromEvent(ctx, 5, 77, 9)

	if err != nil {
		t.Fatalf("KickUserFromEvent returned error: %v", err)
	}
	if !kicked {
		t.Fatal("KickUserFromEvent returned false")
	}
	if uow.calls != 1 || uow.ctx != ctx {
		t.Fatalf("unit of work calls = %d, context preserved = %v", uow.calls, uow.ctx == ctx)
	}
	wantCalls := []string{"find_event", "find_group_role", "user_exists", "remove_participant", "decrement_participants"}
	if !reflect.DeepEqual(adminStore.calls, wantCalls) {
		t.Fatalf("admin store calls = %v, want %v", adminStore.calls, wantCalls)
	}
	if adminStore.findEventID != 77 || adminStore.findRoleActorID != 5 || adminStore.findRoleGroupID != 42 {
		t.Fatalf("admin store lookup args = event:%d actor:%d group:%d", adminStore.findEventID, adminStore.findRoleActorID, adminStore.findRoleGroupID)
	}
	if adminStore.userExistsID != 9 || adminStore.removedUserID != 9 || adminStore.removedEventID != 77 || adminStore.decrementedEventID != 77 {
		t.Fatalf(
			"admin store mutation args = exists:%d remove:(%d,%d) decrement:%d",
			adminStore.userExistsID,
			adminStore.removedUserID,
			adminStore.removedEventID,
			adminStore.decrementedEventID,
		)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.Action != groupmodels.ActionKickFromEvent || entry.GroupID != 42 || entry.ActorID != 5 || entry.TargetUserID == nil || *entry.TargetUserID != 9 {
		t.Fatalf("audit entry = %#v", entry)
	}
	if entry.EntityID == nil || *entry.EntityID != 77 || entry.EntityName != "Admin kick event" {
		t.Fatalf("audit entity = %#v", entry)
	}
	if entry.CreatedAt.IsZero() || entry.CreatedAt.After(time.Now().Add(time.Second)) {
		t.Fatalf("audit created_at = %v, want non-zero current timestamp", entry.CreatedAt)
	}
}

func TestEventAdminServiceKickUserFromEventRejectsSelfKick(t *testing.T) {
	adminStore := &eventAdminStoreStub{
		event: servicesevents.EventAdminEventSnapshot{
			ID:        77,
			GroupID:   42,
			CreatorID: 1,
			Title:     "Admin kick event",
		},
		role: groupmodels.RoleAdmin,
	}
	uow := &eventUnitOfWorkStub{admin: adminStore}
	service := servicesevents.NewEventAdminService(&testLogger{}, nil, uow)

	kicked, err := service.KickUserFromEvent(context.Background(), 5, 77, 5)

	if kicked {
		t.Fatal("KickUserFromEvent returned true")
	}
	if !errors.Is(err, servicesevents.ErrActorCantKickSelf) {
		t.Fatalf("err = %v, want ErrActorCantKickSelf", err)
	}
	if !reflect.DeepEqual(adminStore.calls, []string{"find_event", "find_group_role"}) {
		t.Fatalf("admin store calls = %v, want event lookup then role lookup only", adminStore.calls)
	}
}

func TestEventAdminServiceKickUserFromEventStopsOnDomainFailures(t *testing.T) {
	storageErr := errors.New("admin storage failure")
	tests := []struct {
		name         string
		configure    func(*eventAdminStoreStub)
		wantErr      error
		wantCalls    []string
		wantAuditLen int
	}{
		{
			name: "event lookup error",
			configure: func(store *eventAdminStoreStub) {
				store.findEventErr = storageErr
			},
			wantErr:   storageErr,
			wantCalls: []string{"find_event"},
		},
		{
			name: "actor is not group member",
			configure: func(store *eventAdminStoreStub) {
				store.findRoleErr = servicesevents.ErrNotGroupMember
			},
			wantErr:   servicesevents.ErrNotGroupMember,
			wantCalls: []string{"find_event", "find_group_role"},
		},
		{
			name: "actor cannot moderate",
			configure: func(store *eventAdminStoreStub) {
				store.role = groupmodels.RoleMember
			},
			wantErr:   servicesevents.ErrPermissionDenied,
			wantCalls: []string{"find_event", "find_group_role"},
		},
		{
			name: "target is creator",
			configure: func(store *eventAdminStoreStub) {
				store.event.CreatorID = 9
			},
			wantErr:   servicesevents.ErrCreatorCantLeave,
			wantCalls: []string{"find_event", "find_group_role"},
		},
		{
			name: "target user missing",
			configure: func(store *eventAdminStoreStub) {
				store.targetExists = false
			},
			wantErr:   servicesevents.ErrUserNotFound,
			wantCalls: []string{"find_event", "find_group_role", "user_exists"},
		},
		{
			name: "participant missing",
			configure: func(store *eventAdminStoreStub) {
				store.removeErr = servicesevents.ErrNotJoined
			},
			wantErr: servicesevents.ErrNotJoined,
			wantCalls: []string{
				"find_event",
				"find_group_role",
				"user_exists",
				"remove_participant",
			},
		},
		{
			name: "counter update error",
			configure: func(store *eventAdminStoreStub) {
				store.decrementErr = storageErr
			},
			wantErr: storageErr,
			wantCalls: []string{
				"find_event",
				"find_group_role",
				"user_exists",
				"remove_participant",
				"decrement_participants",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &eventAdminStoreStub{
				event: servicesevents.EventAdminEventSnapshot{
					ID:        77,
					GroupID:   42,
					CreatorID: 1,
					Title:     "Admin kick event",
				},
				role:         groupmodels.RoleAdmin,
				targetExists: true,
			}
			tt.configure(store)
			audit := &eventAuditStoreStub{}
			service := servicesevents.NewEventAdminService(
				&testLogger{},
				nil,
				&eventUnitOfWorkStub{admin: store, audit: audit},
			)

			kicked, err := service.KickUserFromEvent(context.Background(), 5, 77, 9)

			if kicked {
				t.Fatal("KickUserFromEvent returned true")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(store.calls, tt.wantCalls) {
				t.Fatalf("admin store calls = %v, want %v", store.calls, tt.wantCalls)
			}
			if len(audit.entries) != tt.wantAuditLen {
				t.Fatalf("audit entries = %d, want %d", len(audit.entries), tt.wantAuditLen)
			}
		})
	}
}
