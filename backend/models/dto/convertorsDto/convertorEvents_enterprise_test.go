package convertorsdto

import (
	"testing"
	"time"

	eventmodels "friendship/models/events"
	groupmodels "friendship/models/groups"
)

func TestConvertToSearchItemDtoPropagatesGroupEnterprise(t *testing.T) {
	startTime := time.Date(2039, 3, 4, 15, 16, 17, 0, time.UTC)
	for _, enterprise := range []bool{true, false} {
		t.Run(map[bool]string{true: "enterprise", false: "regular"}[enterprise], func(t *testing.T) {
			result := ConvertToSearchItemDto(&eventmodels.Event{
				CurrentUsers: 7,
				StartTime:    startTime,
				AgeLimit:     eventmodels.AgeLimit{Name: "18+"},
				Status:       eventmodels.Status{Name: "Recruitment"},
				Group: groupmodels.Group{
					ID:         42,
					Name:       "Group",
					Image:      "https://example.com/group.png",
					Enterprise: enterprise,
				},
			})

			if result == nil {
				t.Fatal("result is nil")
			}
			if result.Group.Enterprise != enterprise {
				t.Fatalf("group enterprise = %t, want %t", result.Group.Enterprise, enterprise)
			}
			if result.Group.Image != "https://example.com/group.png" ||
				!result.StartTime.Equal(startTime) || result.AgeLimit != "18+" ||
				result.Status != "Recruitment" {
				t.Fatalf("search item = %#v, want shared group/event metadata", result)
			}
			if result.CurrentUsers != 7 {
				t.Fatalf("current users = %d, want 7", result.CurrentUsers)
			}
		})
	}
}
