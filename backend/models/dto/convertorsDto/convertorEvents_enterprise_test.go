package convertorsdto

import (
	"testing"

	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
)

func TestConvertToSearchItemDtoPropagatesGroupEnterprise(t *testing.T) {
	for _, enterprise := range []bool{true, false} {
		t.Run(map[bool]string{true: "enterprise", false: "regular"}[enterprise], func(t *testing.T) {
			result := ConvertToSearchItemDto(&eventmodels.Event{
				CurrentUsers: 7,
				Group: groupmodels.Group{
					ID:         42,
					Name:       "Group",
					Enterprise: enterprise,
				},
			})

			if result == nil {
				t.Fatal("result is nil")
			}
			if result.Group.Enterprise != enterprise {
				t.Fatalf("group enterprise = %t, want %t", result.Group.Enterprise, enterprise)
			}
			if result.CurrentUsers != 7 {
				t.Fatalf("current users = %d, want 7", result.CurrentUsers)
			}
		})
	}
}
