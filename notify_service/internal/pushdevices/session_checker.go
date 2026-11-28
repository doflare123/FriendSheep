package pushdevices

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxSessionCheckResponseBytes int64 = 16 << 10

const (
	SessionCheckErrorNetwork           = "network_error"
	SessionCheckErrorTimeout           = "timeout"
	SessionCheckErrorResponseTooLarge  = "response_too_large"
	SessionCheckErrorMalformedResponse = "malformed_response"
)

// SessionChecker проверяет принадлежащую пользователю серверную сессию.
type SessionChecker interface {
	CheckSession(ctx context.Context, userID uint64, sessionID string) (bool, error)
}

// SessionCheckError описывает временную или контрактную ошибку проверки. Она
// никогда не означает подтверждённый отзыв сессии.
type SessionCheckError struct {
	Code  string
	cause error
}

func (e *SessionCheckError) Error() string {
	if e == nil {
		return ""
	}
	return "проверка push-сессии завершилась ошибкой: " + e.Code
}

func (e *SessionCheckError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// SessionCheckErrorCode возвращает безопасный код ошибки без тела ответа и
// идентификатора сессии.
func SessionCheckErrorCode(err error) string {
	var checkErr *SessionCheckError
	if errors.As(err, &checkErr) {
		return checkErr.Code
	}
	return ""
}

type HTTPMonolithSessionChecker struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewHTTPMonolithSessionChecker(baseURL, token string, timeout time.Duration, transport http.RoundTripper) (*HTTPMonolithSessionChecker, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	token = strings.TrimSpace(token)
	if baseURL == "" || token == "" {
		return nil, errors.New("проверка push-сессий не настроена")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil ||
		(parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("адрес монолита для проверки push-сессий должен быть HTTP(S) origin")
	}
	if timeout <= 0 {
		return nil, errors.New("таймаут проверки push-сессий должен быть положительным")
	}
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if transport != nil {
		client.Transport = transport
	}
	return &HTTPMonolithSessionChecker{
		endpoint: baseURL + "/internal/v1/push-sessions/check",
		token:    token,
		client:   client,
	}, nil
}

func (c *HTTPMonolithSessionChecker) CheckSession(ctx context.Context, userID uint64, sessionID string) (bool, error) {
	if c == nil || c.client == nil || userID == 0 || userID > math.MaxInt64 || !validateSessionID(sessionID) {
		return false, domainError(ErrorCodeInvalidRequest, errors.New("некорректная проверка push-сессии"))
	}
	body, err := json.Marshal(struct {
		SchemaVersion int    `json:"schemaVersion"`
		UserID        uint64 `json:"userId"`
		SessionID     string `json:"sessionId"`
	}{SchemaVersion: SchemaVersion, UserID: userID, SessionID: sessionID})
	if err != nil {
		return false, &SessionCheckError{Code: SessionCheckErrorMalformedResponse, cause: err}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return false, &SessionCheckError{Code: SessionCheckErrorMalformedResponse, cause: err}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Token", c.token)

	response, err := c.client.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false, context.Canceled
		}
		code := SessionCheckErrorNetwork
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			code = SessionCheckErrorTimeout
		}
		return false, &SessionCheckError{Code: code, cause: err}
	}
	defer response.Body.Close()
	encoded, err := io.ReadAll(io.LimitReader(response.Body, maxSessionCheckResponseBytes+1))
	if err != nil {
		return false, &SessionCheckError{Code: SessionCheckErrorMalformedResponse, cause: err}
	}
	if int64(len(encoded)) > maxSessionCheckResponseBytes {
		return false, &SessionCheckError{Code: SessionCheckErrorResponseTooLarge}
	}
	if response.StatusCode != http.StatusOK {
		return false, &SessionCheckError{Code: fmt.Sprintf("unexpected_status_%d", response.StatusCode)}
	}

	active, err := decodeSessionCheckResponse(encoded)
	if err != nil {
		return false, &SessionCheckError{Code: SessionCheckErrorMalformedResponse, cause: err}
	}
	return active, nil
}

func decodeSessionCheckResponse(encoded []byte) (bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false, errors.New("ответ проверки сессии должен быть JSON-объектом")
	}
	seen := make(map[string]struct{}, 2)
	var schemaVersion *int
	var active *bool
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return false, err
		}
		name, ok := token.(string)
		if !ok {
			return false, errors.New("имя поля ответа проверки сессии некорректно")
		}
		if _, duplicate := seen[name]; duplicate {
			return false, errors.New("ответ проверки сессии содержит повтор поля")
		}
		seen[name] = struct{}{}
		switch name {
		case "schemaVersion":
			var value int
			if err := decoder.Decode(&value); err != nil {
				return false, err
			}
			schemaVersion = &value
		case "active":
			if err := decoder.Decode(&active); err != nil {
				return false, err
			}
		default:
			return false, errors.New("ответ проверки сессии содержит неизвестное поле")
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return false, errors.New("ответ проверки сессии не завершён")
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return false, err
	}
	if schemaVersion == nil || *schemaVersion != SchemaVersion || active == nil {
		return false, errors.New("ответ проверки сессии не соответствует версии контракта")
	}
	return *active, nil
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("ответ содержит несколько JSON-значений")
		}
		return err
	}
	return nil
}
