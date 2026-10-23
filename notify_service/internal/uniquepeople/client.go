package uniquepeople

import (
	"bytes"
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

const maxResponseBodyBytes int64 = 1 << 20

type HTTPClient struct {
	baseURL string
	token   string
	client  *http.Client
	clock   Clock
}

func NewHTTPClient(baseURL, token string, timeout time.Duration, transport http.RoundTripper, clock Clock) *HTTPClient {
	if timeout <= 0 {
		timeout = 35 * time.Second
	}
	if clock == nil {
		clock = SystemClock{}
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
	return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), token: token, client: client, clock: clock}
}

func (c *HTTPClient) Claim(ctx context.Context, limit int) (ClaimResponse, error) {
	if limit < 1 || limit > maxClaimLimit {
		return ClaimResponse{}, malformed(errors.New("claim limit должен быть от 1 до 8"))
	}
	u, err := url.Parse(c.baseURL + "/internal/v1/statistics/unique-people/claim")
	if err != nil {
		return ClaimResponse{}, malformed(err)
	}
	query := u.Query()
	query.Set("limit", strconv.Itoa(limit))
	u.RawQuery = query.Encode()
	var response ClaimResponse
	if err = c.request(ctx, http.MethodPost, u.String(), nil, &response); err != nil {
		return ClaimResponse{}, err
	}
	if err = validateClaimResponse(response, limit, c.clock.Now()); err != nil {
		return ClaimResponse{}, err
	}
	return response, nil
}

func (c *HTTPClient) Execute(ctx context.Context, claim Claim) (ExecuteResult, error) {
	if claim.UserID == 0 || claim.Generation == 0 || strings.TrimSpace(claim.LeaseToken) == "" || claim.LeaseUntil.IsZero() {
		return ExecuteResult{}, malformed(errors.New("некорректный unique people claim"))
	}
	request := struct {
		SchemaVersion uint16    `json:"schemaVersion"`
		UserID        uint64    `json:"userId"`
		Generation    uint64    `json:"generation"`
		LeaseToken    string    `json:"leaseToken"`
		LeaseUntil    time.Time `json:"leaseUntil"`
	}{SchemaVersion, claim.UserID, claim.Generation, claim.LeaseToken, claim.LeaseUntil}
	var result ExecuteResult
	if err := c.request(ctx, http.MethodPost, c.baseURL+"/internal/v1/statistics/unique-people/execute", request, &result); err != nil {
		return ExecuteResult{}, err
	}
	if err := validateExecuteResult(result, claim); err != nil {
		return ExecuteResult{}, err
	}
	return result, nil
}

func (c *HTTPClient) request(ctx context.Context, method, address string, body any, target any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return malformed(err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, address, reader)
	if err != nil {
		return malformed(err)
	}
	request.Header.Set("X-Internal-Token", c.token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		code := "network_error"
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			code = "timeout"
		}
		return &ClientError{Code: code, Cause: err}
	}
	defer response.Body.Close()
	encoded, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes+1))
	if err != nil {
		return malformed(err)
	}
	if int64(len(encoded)) > maxResponseBodyBytes {
		return &ClientError{Code: "response_too_large"}
	}
	if response.StatusCode != http.StatusOK {
		return &ClientError{Code: fmt.Sprintf("unexpected_status_%d", response.StatusCode)}
	}
	if err = json.Unmarshal(encoded, target); err != nil {
		return malformed(err)
	}
	return nil
}
