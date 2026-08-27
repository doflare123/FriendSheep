package lifecycle

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

type scheduleClient interface {
	ListScheduleEvents(context.Context, uint64, int) (SchedulePage, error)
}

type advanceClient interface {
	AdvanceLifecycle(context.Context, uint64) (AdvanceResult, error)
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

func (c *HTTPClient) ListScheduleEvents(ctx context.Context, after uint64, limit int) (SchedulePage, error) {
	if limit < 1 {
		return SchedulePage{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: errors.New("limit должен быть положительным")}
	}
	requestURL, err := url.Parse(c.baseURL + "/internal/v1/event-lifecycle/schedule-events")
	if err != nil {
		return SchedulePage{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	query := requestURL.Query()
	query.Set("after", strconv.FormatUint(after, 10))
	query.Set("limit", strconv.Itoa(limit))
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return SchedulePage{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	request.Header.Set("X-Internal-Token", c.token)

	var page SchedulePage
	if err := c.doJSON(request, &page); err != nil {
		return SchedulePage{}, err
	}
	return page, nil
}

func (c *HTTPClient) AdvanceLifecycle(ctx context.Context, eventID uint64) (AdvanceResult, error) {
	if eventID == 0 {
		return AdvanceResult{}, &ClientError{Code: ErrorCodeInvalidEventID, Terminal: true}
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/internal/v1/events/%d/lifecycle/advance", c.baseURL, eventID),
		nil,
	)
	if err != nil {
		return AdvanceResult{}, &ClientError{Code: ErrorCodeMalformedResponse, Terminal: true, Cause: err}
	}
	request.Header.Set("X-Internal-Token", c.token)

	var result AdvanceResult
	if err := c.doJSON(request, &result); err != nil {
		return AdvanceResult{}, err
	}
	if err := validateAdvanceResult(eventID, result); err != nil {
		return AdvanceResult{}, err
	}
	return result, nil
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
	case http.StatusTooManyRequests:
		return &ClientError{Code: code, Retryable: true}
	case http.StatusBadRequest, http.StatusConflict:
		return &ClientError{Code: code, Terminal: true}
	default:
		if status >= 500 {
			return &ClientError{Code: code, Retryable: true}
		}
		return &ClientError{Code: code, Retryable: true}
	}
}

func validateAdvanceResult(requestEventID uint64, result AdvanceResult) error {
	if result.EventID != requestEventID || result.EventID == 0 {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: fmt.Errorf("unexpected eventId=%d", result.EventID)}
	}
	if result.StartTime.IsZero() || result.EndTime.IsZero() || !result.EndTime.After(result.StartTime) {
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: errors.New("invalid lifecycle time window")}
	}
	switch result.Outcome {
	case OutcomeStarted, OutcomeCompleted, OutcomeCompletedCatchUp:
		if !result.Applied {
			return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: fmt.Errorf("outcome %q must be applied", result.Outcome)}
		}
	case OutcomeNotDue, OutcomeAlreadyActive, OutcomeAlreadyCompleted:
		if result.Applied {
			return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: fmt.Errorf("outcome %q must not be applied", result.Outcome)}
		}
	default:
		return &ClientError{Code: ErrorCodeMalformedResponse, Retryable: true, Cause: fmt.Errorf("неизвестный outcome %q", result.Outcome)}
	}
	return nil
}
