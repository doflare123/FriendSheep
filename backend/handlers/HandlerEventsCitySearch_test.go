package handlers

import "testing"

func TestBindEventSearchQueryCityID(t *testing.T) {
	ctx := newEventSearchTestContext(t, "cityId=42")

	input, err := bindEventSearchQuery(ctx)

	if err != nil {
		t.Fatalf("bindEventSearchQuery returned error: %v", err)
	}
	if input.CityID == nil || *input.CityID != 42 || input.City != "" {
		t.Fatalf("city filter = %#v, want cityId=42 without city name", input)
	}
}

func TestBindEventSearchQueryRejectsInvalidCityFilter(t *testing.T) {
	tests := []string{
		"cityId=0",
		"cityId=abc",
		"cityId=42&city=Москва",
	}

	for _, rawQuery := range tests {
		t.Run(rawQuery, func(t *testing.T) {
			ctx := newEventSearchTestContext(t, rawQuery)

			if _, err := bindEventSearchQuery(ctx); err == nil {
				t.Fatalf("bindEventSearchQuery(%q) returned nil error", rawQuery)
			}
		})
	}
}
