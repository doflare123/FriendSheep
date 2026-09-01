package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"friendship/handlers"
	"friendship/models/dto"
	"friendship/routes"
	"friendship/services/notifications"

	"github.com/gin-gonic/gin"
)

type notificationInboxStub struct {
	page          notifications.InboxPage
	count         notifications.UnreadCount
	mark          notifications.MarkReadResult
	err           error
	listUserID    uint
	listCursor    string
	listLimit     int
	listUnread    bool
	countUserID   uint
	markUserID    uint
	markID        string
	listCalls     int
	countCalls    int
	markReadCalls int
}

func (s *notificationInboxStub) List(_ context.Context, userID uint, cursor string, limit int, unreadOnly bool) (notifications.InboxPage, error) {
	s.listCalls++
	s.listUserID, s.listCursor, s.listLimit, s.listUnread = userID, cursor, limit, unreadOnly
	return s.page, s.err
}

func (s *notificationInboxStub) UnreadCount(_ context.Context, userID uint) (notifications.UnreadCount, error) {
	s.countCalls++
	s.countUserID = userID
	return s.count, s.err
}

func (s *notificationInboxStub) MarkRead(_ context.Context, userID uint, notificationID string) (notifications.MarkReadResult, error) {
	s.markReadCalls++
	s.markUserID, s.markID = userID, notificationID
	return s.mark, s.err
}

func TestPublicNotificationsHandlerDerivesUserIDAndIgnoresSpoofedUserID(t *testing.T) {
	readAt := time.Date(2036, 8, 30, 12, 0, 0, 0, time.UTC)
	stub := &notificationInboxStub{
		page:  notifications.InboxPage{Items: []notifications.InboxNotification{}, NextCursor: "cursor-2"},
		count: notifications.UnreadCount{UnreadCount: 3},
		mark:  notifications.MarkReadResult{ID: "notification-1", ReadAt: readAt},
	}
	router := newNotificationsTestRouter(stub, 7)

	t.Run("list", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v2/users/me/notifications?userId=999&cursor=cursor-1&limit=11&unread=true", nil)
		request.Header.Set("X-User-ID", "999")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
		}
		if stub.listCalls != 1 || stub.listUserID != 7 || stub.listCursor != "cursor-1" || stub.listLimit != 11 || !stub.listUnread {
			t.Fatalf("List() inputs = calls:%d user:%d cursor:%q limit:%d unread:%t", stub.listCalls, stub.listUserID, stub.listCursor, stub.listLimit, stub.listUnread)
		}
	})

	t.Run("unread count", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v2/users/me/notifications/unread-count?userId=999", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK || stub.countUserID != 7 {
			t.Fatalf("response = status:%d body:%s; Count user = %d, want 7", response.Code, response.Body.String(), stub.countUserID)
		}
	})

	t.Run("mark read", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPatch, "/api/v2/users/me/notifications/notification-1/read?userId=999", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK || stub.markUserID != 7 || stub.markID != "notification-1" {
			t.Fatalf("response = status:%d body:%s; MarkRead user/id = %d/%q, want 7/notification-1", response.Code, response.Body.String(), stub.markUserID, stub.markID)
		}
	})
}

func TestPublicNotificationsHandlerRejectsMissingAuthenticatedIdentity(t *testing.T) {
	stub := &notificationInboxStub{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := handlers.NewNotificationsHandler(stub)
	router.GET("/notifications", handler.List)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/notifications?userId=999", nil))
	if response.Code != http.StatusUnauthorized || stub.listCalls != 0 {
		t.Fatalf("response = status:%d calls:%d body:%s, want 401/0", response.Code, stub.listCalls, response.Body.String())
	}
}

func TestRegisterNotificationRoutesProtectsEveryRouteWithAuthentication(t *testing.T) {
	stub := &notificationInboxStub{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authCalls := 0
	auth := func(c *gin.Context) {
		authCalls++
		if c.GetHeader("Authorization") != "Bearer valid" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userID", uint(7))
		c.Next()
	}
	routes.RegisterNotificationRoutes(router, handlers.NewNotificationsHandler(stub), auth)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v2/users/me/notifications"},
		{method: http.MethodGet, path: "/api/v2/users/me/notifications/unread-count"},
		{method: http.MethodPatch, path: "/api/v2/users/me/notifications/notification-1/read"},
	}
	for _, test := range tests {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want 401", test.method, test.path, response.Code)
		}
	}
	if authCalls != len(tests) {
		t.Fatalf("auth middleware calls = %d, want %d", authCalls, len(tests))
	}
	if stub.listCalls != 0 || stub.countCalls != 0 || stub.markReadCalls != 0 {
		t.Fatalf("notification handler was called without auth: list=%d count=%d mark=%d", stub.listCalls, stub.countCalls, stub.markReadCalls)
	}
}

func TestPublicNotificationsHandlerUsesCommonSanitizedErrors(t *testing.T) {
	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid request", err: notifications.ErrInboxInvalidInput, wantStatus: http.StatusBadRequest, wantCode: "invalid_notification_request"},
		{name: "not found", err: notifications.ErrInboxNotFound, wantStatus: http.StatusNotFound, wantCode: "notification_not_found"},
		{name: "upstream secret", err: errors.New("notify internal token leaked"), wantStatus: http.StatusInternalServerError, wantCode: "notification_service_unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &notificationInboxStub{err: test.err}
			response := httptest.NewRecorder()
			newNotificationsTestRouter(stub, 7).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v2/users/me/notifications", nil))
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			var body dto.ErrorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error != test.wantCode || body.Message == "" {
				t.Fatalf("error DTO = %#v decode err:%v, want code %q", body, err, test.wantCode)
			}
			if strings.Contains(response.Body.String(), "token leaked") {
				t.Fatalf("response leaked upstream error: %s", response.Body.String())
			}
		})
	}
}

func TestPublicNotificationsHandlerValidatesPaginationBeforeProxy(t *testing.T) {
	for _, path := range []string{
		"/api/v2/users/me/notifications?limit=0",
		"/api/v2/users/me/notifications?limit=101",
		"/api/v2/users/me/notifications?unread=maybe",
	} {
		stub := &notificationInboxStub{}
		response := httptest.NewRecorder()
		newNotificationsTestRouter(stub, 7).ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusBadRequest || stub.listCalls != 0 {
			t.Fatalf("path %q response = status:%d calls:%d body:%s, want 400/0", path, response.Code, stub.listCalls, response.Body.String())
		}
	}
}

func newNotificationsTestRouter(inbox notifications.NotificationInbox, authenticatedUserID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	auth := func(c *gin.Context) {
		c.Set("userID", authenticatedUserID)
		c.Next()
	}
	routes.RegisterNotificationRoutes(router, handlers.NewNotificationsHandler(inbox), auth)
	return router
}

var _ notifications.NotificationInbox = (*notificationInboxStub)(nil)
