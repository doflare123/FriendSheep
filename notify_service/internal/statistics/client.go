package statistics

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

// Ответ ограничен 1 MiB; чтение лишнего байта выявляет усечённую страницу.
const maxResponseBodyBytes int64 = 1 << 20

type HTTPClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewHTTPClient(baseURL, token string, timeout time.Duration, transport http.RoundTripper) *HTTPClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if transport != nil {
		client.Transport = transport
	}
	return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), token: token, client: client}
}

func (c *HTTPClient) ListTriggers(ctx context.Context, after int64, limit int) (Page, error) {
	if after < 0 || limit < 1 || limit > 500 {
		return Page{}, malformed(errors.New("некорректные cursor или limit"))
	}
	u, err := url.Parse(c.baseURL + "/internal/v1/statistics/triggers")
	if err != nil {
		return Page{}, malformed(err)
	}
	query := u.Query()
	query.Set("afterSequence", strconv.FormatInt(after, 10))
	query.Set("limit", strconv.Itoa(limit))
	u.RawQuery = query.Encode()
	var page Page
	if err = c.request(ctx, http.MethodGet, u.String(), &page); err != nil {
		return Page{}, err
	}
	if err = validatePage(page, after, limit); err != nil {
		return Page{}, err
	}
	return page, nil
}

func (c *HTTPClient) ProcessEvent(ctx context.Context, eventID uint64) (ProcessResult, error) {
	var result ProcessResult
	if eventID == 0 {
		return result, malformed(errors.New("eventId должен быть положительным"))
	}
	if err := c.request(ctx, http.MethodPost, fmt.Sprintf("%s/internal/v1/events/%d/statistics/process", c.baseURL, eventID), &result); err != nil {
		return result, err
	}
	if result.EventID != eventID {
		return ProcessResult{}, malformed(errors.New("eventId ответа не совпадает с запросом"))
	}
	switch result.Outcome {
	case "processed", "already_processed":
		if result.ProcessedAt == nil || result.ProcessedAt.IsZero() {
			return ProcessResult{}, malformed(errors.New("processedAt отсутствует"))
		}
	case "missing", "not_completed":
		if result.ProcessedAt != nil {
			return ProcessResult{}, malformed(errors.New("необработанное событие содержит processedAt"))
		}
	default:
		return ProcessResult{}, malformed(errors.New("неизвестный statistics outcome"))
	}
	return result, nil
}

func (c *HTTPClient) request(ctx context.Context, method, address string, target any) error {
	request, err := http.NewRequestWithContext(ctx, method, address, nil)
	if err != nil {
		return malformed(err)
	}
	request.Header.Set("X-Internal-Token", c.token)
	response, err := c.client.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		errorCode := "network_error"
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			errorCode = "timeout"
		}
		return &ClientError{Code: errorCode, Cause: err}
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes+1))
	if err != nil {
		return malformed(err)
	}
	if int64(len(body)) > maxResponseBodyBytes {
		return &ClientError{Code: "response_too_large"}
	}
	if response.StatusCode != http.StatusOK {
		// Ошибки маршрутизации/конфигурации не доказывают отсутствие события; missing приходит в DTO.
		return &ClientError{Code: fmt.Sprintf("unexpected_status_%d", response.StatusCode)}
	}
	if err = json.Unmarshal(body, target); err != nil {
		return malformed(err)
	}
	return nil
}
