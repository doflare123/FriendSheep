package group

import (
	"context"
	"errors"
	"friendship/models/dto"
	"reflect"
	"testing"
)

type managedGroupsReadStoreStub struct {
	result     *dto.ManagedGroupsDto
	err        error
	lastUserID uint
	lastCtx    context.Context
	calls      int
}

type managedGroupsLoggerStub struct{}

func (*managedGroupsLoggerStub) Info(string, ...interface{})  {}
func (*managedGroupsLoggerStub) Error(string, ...interface{}) {}
func (*managedGroupsLoggerStub) Debug(string, ...interface{}) {}
func (*managedGroupsLoggerStub) Warn(string, ...interface{})  {}
func (*managedGroupsLoggerStub) Fatal(string, ...interface{}) {}
func (*managedGroupsLoggerStub) Panic(string, ...interface{}) {}

func (s *managedGroupsReadStoreStub) GetManagedGroups(ctx context.Context, userID uint) (*dto.ManagedGroupsDto, error) {
	s.calls++
	s.lastCtx = ctx
	s.lastUserID = userID
	return s.result, s.err
}

func (*managedGroupsReadStoreStub) GetGroupDetails(context.Context, uint, uint) (*dto.GroupFullDto, error) {
	panic("unexpected GetGroupDetails call")
}

func (*managedGroupsReadStoreStub) ListJoinRequests(context.Context, uint, string, int) ([]JoinRequestInfo, error) {
	panic("unexpected ListJoinRequests call")
}

func (*managedGroupsReadStoreStub) ListGroupBlacklist(context.Context, uint, int) ([]BlacklistUser, error) {
	panic("unexpected ListGroupBlacklist call")
}

func (*managedGroupsReadStoreStub) ListGroupActions(context.Context, uint, GroupActionFilter) ([]GroupAction, error) {
	panic("unexpected ListGroupActions call")
}

func TestGroupServiceGetManagedGroupsPreservesRoleSectionsAndCurrentUser(t *testing.T) {
	want := &dto.ManagedGroupsDto{
		Admin: []dto.ManagedGroupItemDto{{
			ID:               11,
			Name:             "Admin group",
			Categories:       []string{"Board games"},
			SmallDescription: "Administered by current user",
			MemberCount:      14,
			Image:            "https://example.com/admin.png",
		}},
		Moderator: []dto.ManagedGroupItemDto{{
			ID:               22,
			Name:             "Moderator group",
			Categories:       []string{"Sport", "Travel"},
			SmallDescription: "Moderated by current user",
			MemberCount:      27,
			Image:            "https://example.com/moderator.png",
		}},
	}
	store := &managedGroupsReadStoreStub{result: want}
	service := groupService{logger: &managedGroupsLoggerStub{}, reads: store}

	type contextKey string
	const key contextKey = "managed-groups-service-context"
	ctx := context.WithValue(context.Background(), key, "request-value")
	got, err := service.GetManagedGroups(ctx, 73)

	if err != nil {
		t.Fatalf("GetManagedGroups returned error: %v", err)
	}
	if store.lastUserID != 73 {
		t.Fatalf("store userID = %d, want current user 73", store.lastUserID)
	}
	if store.lastCtx != ctx {
		t.Fatal("read store did not receive the exact service context")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want role-separated %#v", got, want)
	}
}

func TestGroupServiceGetManagedGroupsPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("managed groups store unavailable")
	store := &managedGroupsReadStoreStub{err: storeErr}
	service := groupService{logger: &managedGroupsLoggerStub{}, reads: store}

	got, err := service.GetManagedGroups(context.Background(), 81)

	if got != nil {
		t.Fatalf("result = %#v, want nil", got)
	}
	if !errors.Is(err, storeErr) {
		t.Fatalf("error = %v, want errors.Is(..., %v)", err, storeErr)
	}
	if store.lastUserID != 81 {
		t.Fatalf("store userID = %d, want current user 81", store.lastUserID)
	}
}

func TestGroupServiceGetManagedGroupsRejectsZeroUserIDBeforeStore(t *testing.T) {
	store := &managedGroupsReadStoreStub{}
	service := groupService{logger: &managedGroupsLoggerStub{}, reads: store}

	got, err := service.GetManagedGroups(context.Background(), 0)

	if got != nil {
		t.Fatalf("result = %#v, want nil", got)
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	if store.calls != 0 {
		t.Fatalf("store calls = %d, want 0", store.calls)
	}
}
