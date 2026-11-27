package pushdevices

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const installationDomain = "friendSheep:push-installation:v1:"

var (
	installationIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{42}[AEIMQUYcgkosw048]$`)
	uuidV4Pattern         = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

func verifyInstallationOwnership(installationID, secret string) error {
	if !installationIDPattern.MatchString(installationID) {
		return domainError(ErrorCodeInvalidRequest, errors.New("некорректный installationId"))
	}
	secretBytes, err := base64.RawURLEncoding.Strict().DecodeString(secret)
	if err != nil || len(secretBytes) != 32 || base64.RawURLEncoding.EncodeToString(secretBytes) != secret {
		return domainError(ErrorCodePushDeviceNotFound, nil)
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(installationDomain))
	_, _ = hash.Write(secretBytes)
	expected := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(installationID)) != 1 {
		return domainError(ErrorCodePushDeviceNotFound, nil)
	}
	return nil
}

func validateUUIDv4(value string) bool {
	return uuidV4Pattern.MatchString(value)
}

func validateSessionID(value string) bool {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	return err == nil && len(decoded) == 18 && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func validateExpectedVersion(value *string) bool {
	return value == nil || validateUUIDv4(*value)
}

func validateProviderToken(value string) bool {
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

func newUUIDv4() (string, error) {
	value := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value)
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32], nil
}

func keyedFingerprint(key []byte, domain string, payload []byte) []byte {
	mac := hmac.New(sha256.New, key)
	writeFingerprintPart(mac, []byte(domain))
	writeFingerprintPart(mac, payload)
	return mac.Sum(nil)
}

func writeFingerprintPart(writer io.Writer, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write(value)
}

type registerFingerprint struct {
	Method               string  `json:"method"`
	InstallationID       string  `json:"installationId"`
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
}

type detachFingerprint struct {
	Method          string  `json:"method"`
	InstallationID  string  `json:"installationId"`
	SchemaVersion   int     `json:"schemaVersion"`
	UserID          uint64  `json:"userId"`
	SessionID       string  `json:"sessionId"`
	OperationID     string  `json:"operationId"`
	ExpectedVersion *string `json:"expectedVersion"`
}

func (r *Registry) fingerprintRegister(input RegisterInput) ([]byte, error) {
	payload, err := json.Marshal(registerFingerprint{
		Method: "register", InstallationID: input.InstallationID, SchemaVersion: input.SchemaVersion,
		UserID: input.UserID, SessionID: input.SessionID, OperationID: input.OperationID,
		ExpectedVersion: input.ExpectedVersion, Platform: input.Platform, Provider: input.Provider,
		ProviderProject: input.ProviderProject, Environment: input.Environment,
		ProviderToken: input.ProviderToken, NotificationsEnabled: input.NotificationsEnabled,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать отпечаток регистрации: %w", err)
	}
	return keyedFingerprint(r.operationFingerprintKey, "push-device-operation-v1", payload), nil
}

func (r *Registry) fingerprintDetach(input DetachInput) ([]byte, error) {
	payload, err := json.Marshal(detachFingerprint{
		Method: "detach", InstallationID: input.InstallationID, SchemaVersion: input.SchemaVersion,
		UserID: input.UserID, SessionID: input.SessionID, OperationID: input.OperationID,
		ExpectedVersion: input.ExpectedVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать отпечаток отвязки: %w", err)
	}
	return keyedFingerprint(r.operationFingerprintKey, "push-device-operation-v1", payload), nil
}

func (r *Registry) tokenFingerprint(token string) []byte {
	return keyedFingerprint(r.tokenFingerprintKey, "push-device-provider-token-v1", []byte(token))
}

func (r *Registry) encryptProviderToken(input RegisterInput) ([]byte, []byte, error) {
	block, err := aes.NewCipher(r.encryptionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось создать шифр адреса: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось создать режим шифрования адреса: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("не удалось создать nonce адреса: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(input.ProviderToken), providerTokenAAD(
		input.InstallationID, input.Platform, input.Provider, input.ProviderProject, input.Environment,
	))
	return ciphertext, nonce, nil
}

func (r *Registry) decryptProviderToken(record installationRecord) (string, error) {
	if record.TokenKeyID != r.encryptionKeyID {
		return "", fmt.Errorf("неизвестный идентификатор ключа шифрования %q", record.TokenKeyID)
	}
	block, err := aes.NewCipher(r.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("не удалось создать шифр адреса: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("не удалось создать режим шифрования адреса: %w", err)
	}
	if len(record.TokenNonce) != gcm.NonceSize() || len(record.TokenCiphertext) <= gcm.Overhead() {
		return "", errors.New("зашифрованный адрес провайдера повреждён")
	}
	plaintext, err := gcm.Open(nil, record.TokenNonce, record.TokenCiphertext, providerTokenAAD(
		record.InstallationID, record.Platform, record.Provider, record.ProviderProject, record.Environment,
	))
	if err != nil {
		return "", errors.New("не удалось расшифровать адрес провайдера")
	}
	return string(plaintext), nil
}

func providerTokenAAD(installationID, platform, provider, project, environment string) []byte {
	return []byte(strings.Join([]string{
		"push-device-provider-token-v1", installationID, platform, provider, project, environment,
	}, "\x00"))
}
