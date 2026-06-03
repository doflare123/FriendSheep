package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindEventSearchQueryRejectsInvalidCombinations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manyIDs := make([]string, 101)
	for i := range manyIDs {
		manyIDs[i] = strconv.Itoa(i + 1)
	}

	tests := []struct {
		name  string
		query string
	}{
		{"category include exclude conflict", "categoryIds=2798&excludeCategoryIds=2798"},
		{"genre include exclude conflict", "genreIds=7&excludeGenreIds=7"},
		{"event type include exclude conflict", "eventTypeIds=3&excludeEventTypeIds=3"},
		{"location type include exclude conflict", "locationType=online&excludeLocationType=онлайн"},
		{"removed location id", "locationId=1"},
		{"removed location id with location type", "locationId=1&locationType=offline"},
		{"duplicate ids", "genreIds=1&genreIds=1"},
		{"empty csv part", "genreIds=1,,2"},
		{"too many ids", "genreIds=" + strings.Join(manyIDs, ",")},
		{"date range inverted", "dateFrom=2027-01-10&dateTo=2027-01-09"},
		{"page zero", "page=0"},
		{"page negative", "page=-1"},
		{"page too large", "page=10001"},
		{"limit zero", "limit=0"},
		{"limit too large", "limit=101"},
		{"unknown feed", "feed=recommended"},
		{"unknown location type", "locationType=hybrid"},
		{"invalid bool", "hasFreeSlots=maybe"},
		{"invalid date", "dateFrom=2027-99-99"},
		{"bad uint", "categoryIds=abc"},
		{"zero uint", "categoryIds=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newEventSearchTestContext(t, tt.query)

			_, err := bindEventSearchQuery(ctx)

			if err == nil {
				t.Fatalf("bindEventSearchQuery(%q) returned nil error, want validation error", tt.query)
			}
		})
	}
}

func TestBindEventSearchQuerySupportsDocumentedAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := newEventSearchTestContext(t, "query=alias&categoryId=1&categories=2&genreId=3&genres=4&locationTypes=offline&type=офлайн&excludeLocationTypes=online&excludeType=онлайн&from=2027-01-02T10:00:00Z&to=2027-01-02&pageSize=10&category=subscription-news")

	input, err := bindEventSearchQuery(ctx)

	if err != nil {
		t.Fatalf("bindEventSearchQuery returned error: %v", err)
	}
	if input.Query != "alias" {
		t.Fatalf("Query = %q, want alias", input.Query)
	}
	if len(input.CategoryIDs) != 2 || input.CategoryIDs[0] != 1 || input.CategoryIDs[1] != 2 {
		t.Fatalf("CategoryIDs = %#v, want [1 2]", input.CategoryIDs)
	}
	if len(input.GenreIDs) != 2 || input.GenreIDs[0] != 3 || input.GenreIDs[1] != 4 {
		t.Fatalf("GenreIDs = %#v, want [3 4]", input.GenreIDs)
	}
	if len(input.LocationTypes) != 2 || len(input.ExcludeLocationTypes) != 2 {
		t.Fatalf("location aliases were not collected: %#v / %#v", input.LocationTypes, input.ExcludeLocationTypes)
	}
	if input.DateFrom == nil || input.DateTo == nil || input.DateFrom.After(*input.DateTo) {
		t.Fatalf("same-day date range = %#v..%#v, want inclusive dateTo", input.DateFrom, input.DateTo)
	}
	if input.Limit != 10 {
		t.Fatalf("Limit = %d, want 10", input.Limit)
	}
	if !input.OnlySubscriptionNews {
		t.Fatal("OnlySubscriptionNews = false, want true")
	}
}

func TestBindEventSearchQueryAcceptsValidEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := newEventSearchTestContext(t, "q=board&categoryIds=1,2&genreIds=3&genreIds=4&excludeGenreIds=5&eventTypeIds=6&excludeEventTypeIds=7&locationType=offline&excludeLocationType=online&city=Moscow&dateFrom=2027-01-01&dateTo=2027-01-02&hasFreeSlots=true&page=2&limit=50&feed=subscription_news")

	input, err := bindEventSearchQuery(ctx)

	if err != nil {
		t.Fatalf("bindEventSearchQuery returned error: %v", err)
	}
	if input.Page != 2 || input.Limit != 50 {
		t.Fatalf("page/limit = %d/%d, want 2/50", input.Page, input.Limit)
	}
	if !input.OnlySubscriptionNews {
		t.Fatal("OnlySubscriptionNews = false, want true")
	}
	if input.DateFrom == nil || input.DateTo == nil || input.DateFrom.After(*input.DateTo) {
		t.Fatalf("date range = %#v..%#v, want valid range", input.DateFrom, input.DateTo)
	}
	if input.HasFreeSlots == nil || !*input.HasFreeSlots {
		t.Fatalf("HasFreeSlots = %#v, want true", input.HasFreeSlots)
	}
	if len(input.CategoryIDs) != 2 || len(input.GenreIDs) != 2 || len(input.ExcludeGenreIDs) != 1 {
		t.Fatalf("parsed ids mismatch: %#v", input)
	}
}

func newEventSearchTestContext(t *testing.T, rawQuery string) *gin.Context {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v2/events/search?"+rawQuery, nil)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	return ctx
}
