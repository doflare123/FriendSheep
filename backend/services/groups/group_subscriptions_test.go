package group

import (
	"context"
	"errors"
	"testing"
)

type groupSubscriptionsStoreStub struct {
	calls int
}

func (s *groupSubscriptionsStoreStub) ListSubscribedGroups(_ context.Context, _ groupSubscriptionsQuery) (groupSubscriptionsPage, error) {
	s.calls++
	return groupSubscriptionsPage{}, nil
}

func TestGORMGroupSubscriptionsStoreRejectsNilContextBeforeRepositoryAccess(t *testing.T) {
	store := gormGroupSubscriptionsStore{}

	result, err := store.ListSubscribedGroups(nil, groupSubscriptionsQuery{
		UserID: 1,
		Page:   1,
		Limit:  20,
	})

	if !errors.Is(err, errGroupOperationContextMissing) {
		t.Fatalf("error = %v, want errGroupOperationContextMissing", err)
	}
	if result.Items != nil || result.Total != 0 {
		t.Fatalf("result = %#v, want zero page", result)
	}
}

func TestGroupServiceGetSubscribedGroupsRejectsInvalidInputBeforeStore(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name   string
		userID uint
		page   int
		limit  int
	}{
		{name: "missing user", userID: 0, page: 1, limit: 20},
		{name: "zero page", userID: 1, page: 0, limit: 20},
		{name: "zero limit", userID: 1, page: 1, limit: 0},
		{name: "limit above maximum", userID: 1, page: 1, limit: MaxSubscribedGroupsLimit + 1},
		{name: "offset overflow", userID: 1, page: maxInt, limit: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &groupSubscriptionsStoreStub{}
			service := groupService{subscriptions: store}

			result, err := service.GetSubscribedGroups(t.Context(), tt.userID, tt.page, tt.limit)

			if result != nil {
				t.Fatalf("result = %#v, want nil", result)
			}
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
			if store.calls != 0 {
				t.Fatalf("store calls = %d, want 0", store.calls)
			}
		})
	}
}
