package pushdevices

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
)

const (
	SchemaVersion = 1

	StateActive         = "active"
	StateInvalidAddress = "invalid_address"
	StateDetached       = "detached"
	ProviderFCM         = "fcm"

	ErrorCodeInvalidRequest           = "invalid_request"
	ErrorCodeUnsupportedSchemaVersion = "unsupported_schema_version"
	ErrorCodeInvalidProviderScope     = "invalid_provider_scope"
	ErrorCodePushOptInRequired        = "push_opt_in_required"
	ErrorCodePushDeviceNotFound       = "push_device_not_found"
	ErrorCodePushVersionConflict      = "push_version_conflict"
	ErrorCodePushOperationConflict    = "push_operation_conflict"
	ErrorCodePushInstallationConflict = "push_installation_conflict"
	ErrorCodePushTokenConflict        = "push_token_conflict"
	ErrorCodePushTokenInvalidated     = "push_token_invalidated"
)

var (
	scopePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
)

type ProviderScope struct {
	Project     string
	Environment string
}

type CryptoConfig struct {
	EncryptionKey           []byte
	EncryptionKeyID         string
	TokenFingerprintKey     []byte
	OperationFingerprintKey []byte
	AllowedScopes           []ProviderScope
}

func (c CryptoConfig) Validate() error {
	if len(c.EncryptionKey) != 32 {
		return errors.New("ключ шифрования адресов должен содержать 32 байта")
	}
	if !keyIDPattern.MatchString(c.EncryptionKeyID) {
		return errors.New("идентификатор ключа шифрования некорректен")
	}
	if len(c.TokenFingerprintKey) != 32 {
		return errors.New("ключ отпечатков адресов должен содержать 32 байта")
	}
	if len(c.OperationFingerprintKey) != 32 {
		return errors.New("ключ отпечатков операций должен содержать 32 байта")
	}
	if bytes.Equal(c.EncryptionKey, c.TokenFingerprintKey) ||
		bytes.Equal(c.EncryptionKey, c.OperationFingerprintKey) ||
		bytes.Equal(c.TokenFingerprintKey, c.OperationFingerprintKey) {
		return errors.New("ключи реестра устройств должны различаться")
	}
	if len(c.AllowedScopes) == 0 {
		return errors.New("список разрешённых scope провайдера пуст")
	}
	seen := make(map[ProviderScope]struct{}, len(c.AllowedScopes))
	for _, scope := range c.AllowedScopes {
		if !scopePattern.MatchString(scope.Project) || !scopePattern.MatchString(scope.Environment) {
			return errors.New("разрешённый scope провайдера некорректен")
		}
		if _, exists := seen[scope]; exists {
			return errors.New("разрешённый scope провайдера повторяется")
		}
		seen[scope] = struct{}{}
	}
	return nil
}

type RegisterInput struct {
	InstallationID       string
	InstallationSecret   string
	SchemaVersion        int
	UserID               uint64
	SessionID            string
	OperationID          string
	ExpectedVersion      *string
	Platform             string
	Provider             string
	ProviderProject      string
	Environment          string
	ProviderToken        string
	NotificationsEnabled bool
}

type DetachInput struct {
	InstallationID     string
	InstallationSecret string
	SchemaVersion      int
	UserID             uint64
	SessionID          string
	OperationID        string
	ExpectedVersion    *string
}

type MutationResult struct {
	SchemaVersion  int
	InstallationID string
	Version        string
	State          string
}

type Binding struct {
	InstallationID string
	BindingID      string
	UserID         uint64
	SessionID      string
	Version        string
}

type ResolvedAddress struct {
	Binding
	Platform        string
	Provider        string
	ProviderProject string
	Environment     string
	ProviderToken   string
}

type DomainError struct {
	Code           string
	CurrentVersion *string
	cause          error
}

func (e *DomainError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause != nil {
		return e.Code + ": " + e.cause.Error()
	}
	return e.Code
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func ErrorCode(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return ""
}

func CurrentVersion(err error) (*string, bool) {
	var domainErr *DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != ErrorCodePushVersionConflict {
		return nil, false
	}
	return domainErr.CurrentVersion, true
}

type Registry struct {
	db                      *sql.DB
	encryptionKey           []byte
	encryptionKeyID         string
	tokenFingerprintKey     []byte
	operationFingerprintKey []byte
	allowedScopes           map[ProviderScope]struct{}
}

func NewRegistry(db *sql.DB, cfg CryptoConfig) (*Registry, error) {
	if db == nil {
		return nil, errors.New("база данных реестра устройств не настроена")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	allowedScopes := make(map[ProviderScope]struct{}, len(cfg.AllowedScopes))
	for _, scope := range cfg.AllowedScopes {
		allowedScopes[scope] = struct{}{}
	}
	return &Registry{
		db:                      db,
		encryptionKey:           append([]byte(nil), cfg.EncryptionKey...),
		encryptionKeyID:         cfg.EncryptionKeyID,
		tokenFingerprintKey:     append([]byte(nil), cfg.TokenFingerprintKey...),
		operationFingerprintKey: append([]byte(nil), cfg.OperationFingerprintKey...),
		allowedScopes:           allowedScopes,
	}, nil
}

func domainError(code string, cause error) error {
	return &DomainError{Code: code, cause: cause}
}

func versionConflict(current *string) error {
	return &DomainError{Code: ErrorCodePushVersionConflict, CurrentVersion: current}
}

func wrapStoreError(action string, err error) error {
	return fmt.Errorf("%s: %w", action, err)
}
