package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInboxUnauthorized = errors.New("notify service authentication failed")
	ErrInboxNotFound     = errors.New("уведомление не найдено")
	ErrInboxInvalidInput = errors.New("некорректный запрос уведомлений")
	ErrInboxUnavailable  = errors.New("сервис уведомлений недоступен")
)

type EventReminderPayload struct {
	SchemaVersion uint16    `json:"schemaVersion"`
	EventID       uint      `json:"eventId"`
	Title         string    `json:"title"`
	StartTime     time.Time `json:"startTime"`
}

type InboxNotification struct {
	ID                    string               `json:"id"`
	Kind                  string               `json:"kind"`
	ResourceType          string               `json:"resourceType"`
	ResourceID            uint                 `json:"resourceId"`
	Payload               EventReminderPayload `json:"payload"`
	ReminderOffsetMinutes int                  `json:"reminderOffsetMinutes"`
	ScheduleRevision      int64                `json:"scheduleRevision"`
	CreatedAt             time.Time            `json:"createdAt"`
	ReadAt                *time.Time           `json:"readAt"`
}

type InboxPage struct {
	Items      []InboxNotification `json:"items"`
	NextCursor string              `json:"nextCursor"`
	HasMore    bool                `json:"hasMore"`
}

type UnreadCount struct {
	UnreadCount int64 `json:"unreadCount"`
}

type MarkReadResult struct {
	ID     string    `json:"id"`
	ReadAt time.Time `json:"readAt"`
}

type NotificationInbox interface {
	List(ctx context.Context, userID uint, cursor string, limit int, unreadOnly bool) (InboxPage, error)
	UnreadCount(ctx context.Context, userID uint) (UnreadCount, error)
	MarkRead(ctx context.Context, userID uint, notificationID string) (MarkReadResult, error)
}

type HTTPNotificationInboxClient struct {
	baseURL *url.URL
	token   string
	client  *http.Client
}

func NewHTTPNotificationInboxClient(baseURL string, token string, timeout time.Duration) (*HTTPNotificationInboxClient, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("NOTIFY_SERVICE_BASE_URL must be an absolute HTTP(S) URL")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("NOTIFY_SERVICE_TOKEN must not be empty")
	}
	if timeout <= 0 || timeout > 30*time.Second {
		return nil, errors.New("NOTIFY_SERVICE_HTTP_TIMEOUT must be between 1ns and 30s")
	}
	return &HTTPNotificationInboxClient{
		baseURL: parsed,
		token:   token,
		client: &http.Client{
			Timeout:       timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

func (c *HTTPNotificationInboxClient) List(ctx context.Context, userID uint, cursor string, limit int, unreadOnly bool) (InboxPage, error) {
	query := url.Values{}
	if cursor != "" {
		query.Set("after", cursor)
	}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("unread", strconv.FormatBool(unreadOnly))
	var result InboxPage
	err := c.do(ctx, http.MethodGet, c.userPath(userID, "/notifications")+"?"+query.Encode(), &result)
	if result.Items == nil {
		result.Items = []InboxNotification{}
	}
	return result, err
}

func (c *HTTPNotificationInboxClient) UnreadCount(ctx context.Context, userID uint) (UnreadCount, error) {
	var internal struct {
		Count int64 `json:"unreadCount"`
	}
	err := c.do(ctx, http.MethodGet, c.userPath(userID, "/notifications/unread-count"), &internal)
	return UnreadCount{UnreadCount: internal.Count}, err
}

func (c *HTTPNotificationInboxClient) MarkRead(ctx context.Context, userID uint, notificationID string) (MarkReadResult, error) {
	if strings.TrimSpace(notificationID) == "" || strings.Contains(notificationID, "/") {
		return MarkReadResult{}, ErrInboxInvalidInput
	}
	var result MarkReadResult
	err := c.do(ctx, http.MethodPatch, c.userPath(userID, "/notifications/"+url.PathEscape(notificationID)+"/read"), &result)
	return result, err
}

func (c *HTTPNotificationInboxClient) userPath(userID uint, suffix string) string {
	return "/internal/v1/users/" + strconv.FormatUint(uint64(userID), 10) + suffix
}

func (c *HTTPNotificationInboxClient) do(ctx context.Context, method string, path string, destination any) error {
	if c == nil || c.baseURL == nil || c.client == nil || strings.TrimSpace(c.token) == "" {
		return ErrInboxUnavailable
	}
	relative, err := url.Parse(path)
	if err != nil {
		return ErrInboxInvalidInput
	}
	endpoint := c.baseURL.ResolveReference(relative)
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), nil)
	if err != nil {
		return ErrInboxInvalidInput
	}
	req.Header.Set("X-Internal-Token", c.token)
	req.Header.Set("Accept", "application/json")
	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInboxUnavailable, err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if err := json.NewDecoder(limited).Decode(destination); err != nil {
			return fmt.Errorf("%w: malformed response", ErrInboxUnavailable)
		}
		return nil
	}
	_, _ = io.Copy(io.Discard, limited)
	switch response.StatusCode {
	case http.StatusBadRequest:
		return ErrInboxInvalidInput
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrInboxUnauthorized
	case http.StatusNotFound:
		return ErrInboxNotFound
	default:
		return ErrInboxUnavailable
	}
}
