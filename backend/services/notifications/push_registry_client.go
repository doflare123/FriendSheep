package notifications

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	PushDeviceSchemaVersion = 1

	PushDeviceStateActive   = "active"
	PushDeviceStateDetached = "detached"

	PushErrorInvalidRequest           = "invalid_request"
	PushErrorUnsupportedSchemaVersion = "unsupported_schema_version"
	PushErrorOptInRequired            = "push_opt_in_required"
	PushErrorDeviceNotFound           = "push_device_not_found"
	PushErrorVersionConflict          = "push_version_conflict"
	PushErrorOperationConflict        = "push_operation_conflict"
	PushErrorInstallationConflict     = "push_installation_conflict"
	PushErrorTokenConflict            = "push_token_conflict"
	PushErrorTokenInvalidated         = "push_token_invalidated"
)

const (
	pushProviderFCM       = "fcm"
	pushInstallationSalt  = "friendSheep:push-installation:v1:"
	pushResponseBodyLimit = 16_384
)

var (
	ErrPushRegistryUnavailable = errors.New("реестр push-устройств недоступен")

	pushScopePattern          = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	pushInstallationIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{42}[AEIMQUYcgkosw048]$`)
	pushUUIDv4Pattern         = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

type RegisterPushDeviceInput struct {
	InstallationID       string
	InstallationSecret   string
	SchemaVersion        int
	UserID               uint64
	SessionID            string
	OperationID          string
	ExpectedVersion      *string
	Platform             string
	ProviderToken        string
	NotificationsEnabled bool
}

type DetachPushDeviceInput struct {
	InstallationID     string
	InstallationSecret string
	SchemaVersion      int
	UserID             uint64
	SessionID          string
	OperationID        string
	ExpectedVersion    *string
}

type PushDeviceMutationResult struct {
	SchemaVersion  int    `json:"schemaVersion" binding:"required" enums:"1"`
	InstallationID string `json:"installationId" binding:"required"`
	Version        string `json:"version" binding:"required"`
	State          string `json:"state" binding:"required" enums:"active,detached"`
}

type PushDeviceRegistry interface {
	Register(ctx context.Context, input RegisterPushDeviceInput) (PushDeviceMutationResult, error)
	Detach(ctx context.Context, input DetachPushDeviceInput) (PushDeviceMutationResult, error)
}

type PushRegistryError struct {
	Code           string
	CurrentVersion *string
}

func (e *PushRegistryError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code
}

type HTTPPushDeviceRegistryClient struct {
	baseURL         *url.URL
	internalToken   string
	providerProject string
	environment     string
	client          *http.Client
}

func NewHTTPPushDeviceRegistryClient(
	baseURL string,
	internalToken string,
	timeout time.Duration,
	providerProject string,
	environment string,
) (*HTTPPushDeviceRegistryClient, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("NOTIFY_SERVICE_BASE_URL must be an absolute HTTP(S) URL without credentials, query or fragment")
	}
	if strings.TrimSpace(internalToken) == "" {
		return nil, errors.New("NOTIFY_SERVICE_TOKEN must not be empty")
	}
	if timeout <= 0 || timeout > 30*time.Second {
		return nil, errors.New("NOTIFY_SERVICE_HTTP_TIMEOUT must be between 1ns and 30s")
	}
	if !pushScopePattern.MatchString(providerProject) {
		return nil, errors.New("NOTIFY_PUSH_PROVIDER_PROJECT has invalid format")
	}
	if !pushScopePattern.MatchString(environment) {
		return nil, errors.New("NOTIFY_PUSH_ENVIRONMENT has invalid format")
	}
	return &HTTPPushDeviceRegistryClient{
		baseURL:         parsed,
		internalToken:   internalToken,
		providerProject: providerProject,
		environment:     environment,
		client: &http.Client{
			Timeout:       timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

func (c *HTTPPushDeviceRegistryClient) Register(ctx context.Context, input RegisterPushDeviceInput) (PushDeviceMutationResult, error) {
	if err := ValidateRegisterPushDeviceInput(input); err != nil {
		return PushDeviceMutationResult{}, err
	}
	if c == nil || c.baseURL == nil || c.client == nil {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	body := struct {
		SchemaVersion        int     `json:"schemaVersion"`
		UserID               uint64  `json:"userId"`
		SessionID            string  `json:"sessionId"`
		OperationID          string  `json:"operationId"`
		ExpectedVersion      *string `json:"expectedVersion"`
		Platform             string  `json:"platform"`
		Provider             string  `json:"provider"`
		ProviderProject      string  `json:"providerProject"`
		Environment          string  `json:"environment"`
		ProviderToken        string  `json:"providerToken"`
		NotificationsEnabled bool    `json:"notificationsEnabled"`
	}{
		SchemaVersion: input.SchemaVersion, UserID: input.UserID, SessionID: input.SessionID,
		OperationID: input.OperationID, ExpectedVersion: input.ExpectedVersion, Platform: input.Platform,
		Provider: pushProviderFCM, ProviderProject: c.providerProject, Environment: c.environment,
		ProviderToken: input.ProviderToken, NotificationsEnabled: input.NotificationsEnabled,
	}
	return c.mutate(ctx, http.MethodPut, input.InstallationID, "", input.InstallationSecret, body, PushDeviceStateActive)
}

func (c *HTTPPushDeviceRegistryClient) Detach(ctx context.Context, input DetachPushDeviceInput) (PushDeviceMutationResult, error) {
	if err := ValidateDetachPushDeviceInput(input); err != nil {
		return PushDeviceMutationResult{}, err
	}
	if c == nil || c.baseURL == nil || c.client == nil {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	body := struct {
		SchemaVersion   int     `json:"schemaVersion"`
		UserID          uint64  `json:"userId"`
		SessionID       string  `json:"sessionId"`
		OperationID     string  `json:"operationId"`
		ExpectedVersion *string `json:"expectedVersion"`
	}{
		SchemaVersion: input.SchemaVersion, UserID: input.UserID, SessionID: input.SessionID,
		OperationID: input.OperationID, ExpectedVersion: input.ExpectedVersion,
	}
	return c.mutate(ctx, http.MethodPost, input.InstallationID, "/detach", input.InstallationSecret, body, PushDeviceStateDetached)
}

func (c *HTTPPushDeviceRegistryClient) mutate(
	ctx context.Context,
	method string,
	installationID string,
	suffix string,
	installationSecret string,
	body any,
	wantState string,
) (PushDeviceMutationResult, error) {
	if c == nil || c.baseURL == nil || c.client == nil || strings.TrimSpace(c.internalToken) == "" {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	payload, err := marshalPushJSON(body)
	if err != nil {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	relative := &url.URL{Path: "/internal/v1/push-devices/" + installationID + suffix}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL.ResolveReference(relative).String(), bytes.NewReader(payload))
	if err != nil {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Token", c.internalToken)
	request.Header.Set("X-Installation-Secret", installationSecret)

	response, err := c.client.Do(request)
	if err != nil {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	defer response.Body.Close()
	raw, err := readBoundedPushResponse(response.Body)
	if err != nil {
		return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
	}
	if response.StatusCode == http.StatusOK {
		var result PushDeviceMutationResult
		if err := decodeExactJSON(raw, &result); err != nil || !hasExactJSONFields(raw,
			[]string{"schemaVersion", "installationId", "version", "state"}) || result.SchemaVersion != PushDeviceSchemaVersion ||
			result.InstallationID != installationID || !pushUUIDv4Pattern.MatchString(result.Version) || result.State != wantState {
			return PushDeviceMutationResult{}, ErrPushRegistryUnavailable
		}
		return result, nil
	}
	return PushDeviceMutationResult{}, decodePushRegistryError(response.StatusCode, raw)
}

func marshalPushJSON(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'}), nil
}

func ValidateRegisterPushDeviceInput(input RegisterPushDeviceInput) error {
	if err := validatePushInstallation(input.InstallationID, input.InstallationSecret); err != nil {
		return err
	}
	if input.SchemaVersion != PushDeviceSchemaVersion {
		return &PushRegistryError{Code: PushErrorUnsupportedSchemaVersion}
	}
	if !validatePushIdentity(input.UserID, input.SessionID) || !pushUUIDv4Pattern.MatchString(input.OperationID) ||
		!validatePushExpectedVersion(input.ExpectedVersion) {
		return &PushRegistryError{Code: PushErrorInvalidRequest}
	}
	if !input.NotificationsEnabled {
		return &PushRegistryError{Code: PushErrorOptInRequired}
	}
	if (input.Platform != "android" && input.Platform != "ios") || !validatePushProviderToken(input.ProviderToken) {
		return &PushRegistryError{Code: PushErrorInvalidRequest}
	}
	return nil
}

func ValidateDetachPushDeviceInput(input DetachPushDeviceInput) error {
	if err := validatePushInstallation(input.InstallationID, input.InstallationSecret); err != nil {
		return err
	}
	if input.SchemaVersion != PushDeviceSchemaVersion {
		return &PushRegistryError{Code: PushErrorUnsupportedSchemaVersion}
	}
	if !validatePushIdentity(input.UserID, input.SessionID) || !pushUUIDv4Pattern.MatchString(input.OperationID) ||
		!validatePushExpectedVersion(input.ExpectedVersion) {
		return &PushRegistryError{Code: PushErrorInvalidRequest}
	}
	return nil
}

func validatePushInstallation(installationID, secret string) error {
	if !pushInstallationIDPattern.MatchString(installationID) {
		return &PushRegistryError{Code: PushErrorInvalidRequest}
	}
	secretBytes, err := base64.RawURLEncoding.Strict().DecodeString(secret)
	if err != nil || len(secretBytes) != 32 || base64.RawURLEncoding.EncodeToString(secretBytes) != secret {
		return &PushRegistryError{Code: PushErrorDeviceNotFound}
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(pushInstallationSalt))
	_, _ = hash.Write(secretBytes)
	expected := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(installationID)) != 1 {
		return &PushRegistryError{Code: PushErrorDeviceNotFound}
	}
	return nil
}

func validatePushIdentity(userID uint64, sessionID string) bool {
	if userID == 0 || userID > math.MaxInt64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(sessionID)
	return err == nil && len(decoded) == 18 && base64.RawURLEncoding.EncodeToString(decoded) == sessionID
}

func validatePushExpectedVersion(version *string) bool {
	return version == nil || pushUUIDv4Pattern.MatchString(*version)
}

func validatePushProviderToken(value string) bool {
	if len(value) < 1 || len(value) > 4096 {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return false
		}
	}
	return true
}

func readBoundedPushResponse(body io.Reader) ([]byte, error) {
	limited := io.LimitReader(body, pushResponseBodyLimit+1)
	raw, err := io.ReadAll(limited)
	if err != nil || len(raw) > pushResponseBodyLimit {
		return nil, ErrPushRegistryUnavailable
	}
	return raw, nil
}

func decodeExactJSON(raw []byte, destination any) error {
	if err := rejectDuplicateJSONFields(raw); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("лишние данные JSON")
	}
	return nil
}

func rejectDuplicateJSONFields(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return errors.New("ожидался JSON-объект")
	}
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return errors.New("некорректное поле JSON")
		}
		if _, exists := seen[key]; exists {
			return errors.New("повторное поле JSON")
		}
		seen[key] = struct{}{}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return errors.New("некорректный JSON-объект")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("лишние данные JSON")
	}
	return nil
}

func decodePushRegistryError(status int, raw []byte) error {
	var envelope struct {
		Error   string `json:"error"`
		Details *struct {
			CurrentVersion *string `json:"currentVersion"`
		} `json:"details,omitempty"`
	}
	if err := decodeExactJSON(raw, &envelope); err != nil {
		return ErrPushRegistryUnavailable
	}
	if envelope.Details == nil {
		if !hasExactJSONFields(raw, []string{"error"}) {
			return ErrPushRegistryUnavailable
		}
	} else if !hasExactJSONFields(raw, []string{"error", "details"}) {
		return ErrPushRegistryUnavailable
	}
	known := false
	switch status {
	case http.StatusBadRequest:
		known = envelope.Error == PushErrorInvalidRequest || envelope.Error == PushErrorUnsupportedSchemaVersion ||
			envelope.Error == PushErrorOptInRequired
	case http.StatusNotFound:
		known = envelope.Error == PushErrorDeviceNotFound
	case http.StatusConflict:
		known = envelope.Error == PushErrorVersionConflict || envelope.Error == PushErrorOperationConflict ||
			envelope.Error == PushErrorInstallationConflict || envelope.Error == PushErrorTokenConflict ||
			envelope.Error == PushErrorTokenInvalidated
	}
	if !known {
		return ErrPushRegistryUnavailable
	}
	if envelope.Error == PushErrorVersionConflict {
		if envelope.Details == nil || !jsonNestedObjectHasExactFields(raw, "details", []string{"currentVersion"}) ||
			envelope.Details.CurrentVersion != nil && !pushUUIDv4Pattern.MatchString(*envelope.Details.CurrentVersion) {
			return ErrPushRegistryUnavailable
		}
		return &PushRegistryError{Code: envelope.Error, CurrentVersion: envelope.Details.CurrentVersion}
	}
	if envelope.Details != nil {
		return ErrPushRegistryUnavailable
	}
	return &PushRegistryError{Code: envelope.Error}
}

func hasExactJSONFields(raw []byte, expected []string) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || len(object) != len(expected) {
		return false
	}
	for _, field := range expected {
		if _, exists := object[field]; !exists {
			return false
		}
	}
	return true
}

func jsonNestedObjectHasExactFields(raw []byte, objectField string, expected []string) bool {
	var outer map[string]json.RawMessage
	if json.Unmarshal(raw, &outer) != nil {
		return false
	}
	nested := outer[objectField]
	return rejectDuplicateJSONFields(nested) == nil && hasExactJSONFields(nested, expected)
}
