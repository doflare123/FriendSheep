package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"notify_service/internal/pushdevices"
)

func loadPushRegistry(getenv Getenv) (*pushdevices.CryptoConfig, error) {
	scopes := getenv("NOTIFY_PUSH_ALLOWED_SCOPES")
	keyID := getenv("NOTIFY_PUSH_ENCRYPTION_KEY_ID")
	encryptionKey := getenv("NOTIFY_PUSH_ENCRYPTION_KEY")
	tokenKey := getenv("NOTIFY_PUSH_TOKEN_FINGERPRINT_KEY")
	operationKey := getenv("NOTIFY_PUSH_OPERATION_FINGERPRINT_KEY")
	if scopes == "" && keyID == "" && encryptionKey == "" && tokenKey == "" && operationKey == "" {
		return nil, nil
	}
	settings := &pushdevices.CryptoConfig{EncryptionKeyID: keyID}
	for _, pair := range strings.Split(scopes, ",") {
		project, environment, found := strings.Cut(pair, ":")
		if !found || project == "" || environment == "" {
			return nil, errors.New("требуется NOTIFY_PUSH_ALLOWED_SCOPES в формате project:environment[,project:environment]")
		}
		settings.AllowedScopes = append(settings.AllowedScopes, pushdevices.ProviderScope{Project: project, Environment: environment})
	}
	for _, key := range []struct {
		name string
		raw  string
		out  *[]byte
	}{
		{"NOTIFY_PUSH_ENCRYPTION_KEY", encryptionKey, &settings.EncryptionKey},
		{"NOTIFY_PUSH_TOKEN_FINGERPRINT_KEY", tokenKey, &settings.TokenFingerprintKey},
		{"NOTIFY_PUSH_OPERATION_FINGERPRINT_KEY", operationKey, &settings.OperationFingerprintKey},
	} {
		value, err := base64.RawURLEncoding.Strict().DecodeString(key.raw)
		if err != nil || len(value) != 32 || base64.RawURLEncoding.EncodeToString(value) != key.raw {
			return nil, fmt.Errorf("%s должен содержать 32 байта в каноническом base64url без заполнения", key.name)
		}
		*key.out = value
	}
	return settings, nil
}
