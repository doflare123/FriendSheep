package events

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestCreateEventInputUnmarshalJSONAcceptsStartTimeWithoutTimezone(t *testing.T) {
	payload := []byte(`{
		"title":"Board Game Night",
		"description":"Long enough description for binding",
		"groupId":1,
		"eventTypeId":1,
		"locationId":1,
		"imageUrl":"https://example.com/event.png",
		"startTime":"2027-01-02T15:04:05",
		"duration":60,
		"maxUsers":10,
		"genres":[1],
		"ageLimit":1
	}`)

	var input CreateEventInput
	if err := json.Unmarshal(payload, &input); err != nil {
		t.Fatalf("unmarshal create input: %v", err)
	}

	want := time.Date(2027, time.January, 2, 15, 4, 5, 0, time.UTC)
	if !input.StartTime.Equal(want) {
		t.Fatalf("startTime = %s, want %s", input.StartTime, want)
	}
}

func TestUpdateEventInputUnmarshalJSONAcceptsStartTimeWithoutTimezone(t *testing.T) {
	payload := []byte(`{"startTime":"2027-01-02T15:04:05"}`)

	var input UpdateEventInput
	if err := json.Unmarshal(payload, &input); err != nil {
		t.Fatalf("unmarshal update input: %v", err)
	}

	if input.StartTime == nil {
		t.Fatal("startTime is nil")
	}

	want := time.Date(2027, time.January, 2, 15, 4, 5, 0, time.UTC)
	if !input.StartTime.Equal(want) {
		t.Fatalf("startTime = %s, want %s", input.StartTime, want)
	}
}

func TestCreateEventInputUnmarshalJSONRejectsInvalidStartTimeWithFriendlyError(t *testing.T) {
	payload := []byte(`{
		"title":"Board Game Night",
		"description":"Long enough description for binding",
		"groupId":1,
		"eventTypeId":1,
		"locationId":1,
		"imageUrl":"https://example.com/event.png",
		"startTime":"not-a-time",
		"duration":60,
		"maxUsers":10,
		"genres":[1],
		"ageLimit":1
	}`)

	var input CreateEventInput
	err := json.Unmarshal(payload, &input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidStartTimeFormat) {
		t.Fatalf("err = %v, want ErrInvalidStartTimeFormat", err)
	}
}
