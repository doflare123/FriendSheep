package reminders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBodyBytes int64 = 64 << 10

type sourceClient interface {
	ListReminderIntents(context.Context, uint64, int) (IntentPage, error)
}

type recipientClient interface {
	ResolveRecipients(context.Context, uint64, int) (RecipientSnapshot, error)
}

type HTTPClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewHTTPClient(baseURL string, token string, timeout time.Duration, transport http.RoundTripper) *HTTPClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	httpClient := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if transport != nil {
		httpClient.Transport = transport
	}
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  httpClient,
	}
}

func (c *HTTPClient) ListReminderIntents(ctx context.Context, after uint64, limit int) (IntentPage, error) {
	if limit < 1 {
		return IntentPage{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: errors.New("limit должен быть положительным")}
	}
	requestURL, err := url.Parse(c.baseURL + "/internal/v1/notification-intents/event-reminders")
	if err != nil {
		return IntentPage{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	query := requestURL.Query()
	query.Set("after", strconv.FormatUint(after, 10))
	query.Set("limit", strconv.Itoa(limit))
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return IntentPage{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	request.Header.Set("X-Internal-Token", c.token)

	var page IntentPage
	if err := c.doJSON(request, &page); err != nil {
		return IntentPage{}, err
	}
	return page, nil
}

func (c *HTTPClient) ResolveRecipients(ctx context.Context, eventID uint64, reminderOffsetMinutes int) (RecipientSnapshot, error) {
	if eventID == 0 {
		return RecipientSnapshot{}, &ClientError{Code: ErrorCodeInvalidEventID, Terminal: true}
	}
	if reminderOffsetMinutes <= 0 {
		return RecipientSnapshot{}, &ClientError{Code: ErrorCodeInvalidReminder, Terminal: true}
	}
	requestURL, err := url.Parse(fmt.Sprintf("%s/internal/v1/event-reminders/%d/recipients", c.baseURL, eventID))
	if err != nil {
		return RecipientSnapshot{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	query := requestURL.Query()
	query.Set("reminderOffsetMinutes", strconv.Itoa(reminderOffsetMinutes))
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return RecipientSnapshot{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	request.Header.Set("X-Internal-Token", c.token)

	var snapshot RecipientSnapshot
	if err := c.doJSON(request, &snapshot); err != nil {
		return RecipientSnapshot{}, err
	}
	if err := validateRecipientSnapshot(eventID, reminderOffsetMinutes, snapshot); err != nil {
		return RecipientSnapshot{}, err
	}
	return snapshot, nil
}

func (c *HTTPClient) doJSON(request *http.Request, target any) error {
	response, err := c.client.Do(request)
	if err != nil {
		return classifyTransportError(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes))
	if err != nil {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: err}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return classifyHTTPError(response.StatusCode, body)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: err}
	}
	return nil
}

func validateSchedulePage(page IntentPage, after uint64) error {
	if page.Items == nil {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("items отсутствует в reminder response")}
	}
	if page.HasMore && len(page.Items) == 0 {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("hasMore=true при пустом reminder batch")}
	}
	if len(page.Items) == 0 {
		if page.NextCursor != after {
			return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("пустой reminder batch вернул неожиданный nextCursor")}
		}
		return nil
	}
	if page.Items[0].Sequence <= after {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("reminder batch начинается не после after cursor")}
	}
	if page.NextCursor != page.Items[len(page.Items)-1].Sequence {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("reminder nextCursor не совпадает с последней sequence")}
	}
	return nil
}

func validateRecipientSnapshot(eventID uint64, offset int, snapshot RecipientSnapshot) error {
	if snapshot.EventID != eventID || snapshot.EventID == 0 {
		return &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: fmt.Errorf("unexpected eventId=%d", snapshot.EventID)}
	}
	if snapshot.ReminderOffsetMinutes != offset {
		return &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: fmt.Errorf("unexpected reminderOffsetMinutes=%d", snapshot.ReminderOffsetMinutes)}
	}
	if snapshot.Title == "" || snapshot.StartTime.IsZero() || snapshot.Recipients == nil {
		return &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: errors.New("recipient snapshot incomplete")}
	}
	for _, recipient := range snapshot.Recipients {
		if recipient.UserID == 0 {
			return &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: errors.New("recipient userId=0")}
		}
		for _, channel := range recipient.Channels {
			if strings.TrimSpace(channel) == "" {
				return &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: errors.New("recipient содержит пустой channel code")}
			}
		}
	}
	return nil
}

func classifyTransportError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.Canceled):
		return err
	case errors.Is(err, context.DeadlineExceeded):
		return &ClientError{Code: ErrorCodeTimeout, Retryable: true, Cause: err}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		code := ErrorCodeNetwork
		if netErr.Timeout() {
			code = ErrorCodeTimeout
		}
		return &ClientError{Code: code, Retryable: true, Cause: err}
	}
	return &ClientError{Code: ErrorCodeNetwork, Retryable: true, Cause: err}
}

func classifyHTTPError(status int, body []byte) error {
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	code := payload.Error
	if code == "" {
		code = fmt.Sprintf("%s_%d", ErrorCodeUnexpectedStatus, status)
	}

	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ClientError{Code: code, Retryable: true}
	case http.StatusNotFound:
		return &ClientError{Code: code, Terminal: true}
	case http.StatusBadRequest:
		return &ClientError{Code: code, Terminal: true}
	case http.StatusTooManyRequests:
		return &ClientError{Code: code, Retryable: true}
	default:
		if status >= 500 {
			return &ClientError{Code: code, Retryable: true}
		}
		return &ClientError{Code: code, Retryable: true}
	}
}
