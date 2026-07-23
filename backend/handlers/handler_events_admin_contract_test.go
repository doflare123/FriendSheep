package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEventsHandlerAdminSourceUsesSeparateAdminDependency(t *testing.T) {
	content, err := os.ReadFile("HandlerEvents.go")
	if err != nil {
		t.Fatalf("read handler source: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "Admin") {
		t.Fatal("EventsHandlerDependencies does not expose separate Admin dependency")
	}
	if strings.Contains(text, "h.srv.GetEventDetailsForAdmin(") {
		t.Fatal("GetEventDetailsForAdmin still uses legacy srv dependency")
	}
	if strings.Contains(text, "h.srv.KickUserFromEvent(") {
		t.Fatal("KickUserFromEvent still uses legacy srv dependency")
	}
	if !strings.Contains(text, ".GetEventDetailsForAdmin(c.Request.Context(),") {
		t.Fatal("GetEventDetailsForAdmin does not pass request context to admin service")
	}
	if !strings.Contains(text, ".KickUserFromEvent(c.Request.Context(),") {
		t.Fatal("KickUserFromEvent does not pass request context to admin service")
	}
}

func TestEventsHandlerAdminSourceMapsGroupMembershipErrorDirectly(t *testing.T) {
	content, err := os.ReadFile("HandlerEvents.go")
	if err != nil {
		t.Fatalf("read handler source: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "case errors.Is(err, events.ErrNotGroupMember):") {
		t.Fatal("admin handler does not map ErrNotGroupMember explicitly")
	}
	if strings.Contains(text, "case errors.Is(err, events.ErrNotInGroup):") {
		t.Fatal("admin handler still maps legacy ErrNotInGroup instead of ErrNotGroupMember")
	}
}
