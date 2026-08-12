package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenUse = "access"

type JWTConfig struct {
	Issuer           string
	Audience         string
	KeyID            string
	VerificationKeys map[string]string
	AccessTTL        time.Duration
	ClockSkew        time.Duration
}

type JWTUtils struct {
	secretKey           string
	keyID               string
	verificationKeys    map[string][]byte
	issuer              string
	audience            string
	accessTokenDuration time.Duration
	clockSkew           time.Duration
}

func NewJWTUtils(secretKey string) *JWTUtils {
	return NewJWTUtilsWithConfig(secretKey, JWTConfig{})
}

func NewJWTUtilsWithConfig(secretKey string, cfg JWTConfig) *JWTUtils {
	if cfg.Issuer == "" {
		cfg.Issuer = "friendship"
	}
	if cfg.Audience == "" {
		cfg.Audience = "friendship-api"
	}
	if cfg.KeyID == "" {
		cfg.KeyID = "primary"
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 20 * time.Minute
	}
	verificationKeys := make(map[string][]byte, len(cfg.VerificationKeys)+1)
	verificationKeys[cfg.KeyID] = []byte(secretKey)
	for keyID, key := range cfg.VerificationKeys {
		if strings.TrimSpace(keyID) == "" || key == "" || keyID == cfg.KeyID {
			continue
		}
		verificationKeys[keyID] = []byte(key)
	}
	return &JWTUtils{
		secretKey:           secretKey,
		keyID:               cfg.KeyID,
		verificationKeys:    verificationKeys,
		issuer:              cfg.Issuer,
		audience:            cfg.Audience,
		accessTokenDuration: cfg.AccessTTL,
		clockSkew:           cfg.ClockSkew,
	}
}

type Claims struct {
	UserID    uint   `json:"-"`
	SessionID string `json:"sid,omitempty"`
	TokenUse  string `json:"token_use,omitempty"`
	TokenID   string `json:"-"`
	jwt.RegisteredClaims
}

func (j *JWTUtils) AccessTTL() time.Duration {
	return j.accessTokenDuration
}

func (j *JWTUtils) GenerateAccessToken(userID uint, sessionID string) (string, string, time.Time, error) {
	if j.secretKey == "" {
		return "", "", time.Time{}, fmt.Errorf("пустой секретный ключ")
	}
	if userID == 0 {
		return "", "", time.Time{}, fmt.Errorf("некорректный идентификатор пользователя")
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", "", time.Time{}, fmt.Errorf("идентификатор сессии обязателен")
	}

	now := time.Now().UTC()
	tokenID, err := randomTokenString(24)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("ошибка генерации идентификатора токена: %w", err)
	}

	claims := Claims{
		SessionID: sessionID,
		TokenUse:  accessTokenUse,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{j.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = j.keyID
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("ошибка создания access токена: %w", err)
	}

	return signedToken, tokenID, claims.ExpiresAt.Time, nil
}

func (j *JWTUtils) ParseAccessToken(tokenString string) (*Claims, error) {
	if j.secretKey == "" {
		return nil, fmt.Errorf("секретный ключ не установлен")
	}
	if err := validateAccessTokenPayloadShape(tokenString); err != nil {
		return nil, err
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(j.issuer),
		jwt.WithAudience(j.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(j.clockSkew),
	)

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok || strings.TrimSpace(keyID) == "" {
			return nil, fmt.Errorf("access токен не содержит kid")
		}
		key, ok := j.verificationKeys[keyID]
		if !ok || len(key) == 0 {
			return nil, fmt.Errorf("неизвестный kid")
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("невалидный токен: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("невалидные claims")
	}

	if claims.ExpiresAt == nil {
		return nil, fmt.Errorf("access токен не содержит exp")
	}
	if claims.IssuedAt == nil {
		return nil, fmt.Errorf("access токен не содержит iat")
	}
	if claims.ExpiresAt.Time.Before(claims.IssuedAt.Time) {
		return nil, fmt.Errorf("access токен содержит exp раньше iat")
	}
	if claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) > j.accessTokenDuration+j.clockSkew {
		return nil, fmt.Errorf("access токен превышает допустимый срок действия")
	}
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("невалидный issuer")
	}
	if len(claims.Audience) == 0 || !hasAudience(claims.Audience, j.audience) {
		return nil, fmt.Errorf("невалидный audience")
	}
	if claims.TokenUse != accessTokenUse {
		return nil, fmt.Errorf("токен не является access токеном")
	}

	sessionID := strings.TrimSpace(claims.SessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("access токен не содержит sid")
	}

	tokenID := strings.TrimSpace(claims.ID)
	if tokenID == "" {
		return nil, fmt.Errorf("access токен не содержит jti")
	}

	userID, err := parsePositiveSubject(claims.Subject)
	if err != nil {
		return nil, err
	}

	claims.UserID = userID
	claims.SessionID = sessionID
	claims.TokenID = tokenID
	return claims, nil
}

func validateAccessTokenPayloadShape(tokenString string) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("невалидный формат access токена")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("невалидный payload access токена: %w", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("невалидный JSON payload access токена: %w", err)
	}
	allowed := map[string]struct{}{
		"sub": {}, "iss": {}, "aud": {}, "exp": {}, "iat": {}, "jti": {}, "sid": {}, "token_use": {},
	}
	for claim := range payload {
		if _, ok := allowed[claim]; !ok {
			return fmt.Errorf("access токен содержит недопустимый claim %q", claim)
		}
	}
	return nil
}

func (j *JWTUtils) ValidateToken(tokenString string) error {
	_, err := j.ParseAccessToken(tokenString)
	return err
}

func parsePositiveSubject(subject string) (uint, error) {
	trimmed := strings.TrimSpace(subject)
	if trimmed == "" {
		return 0, fmt.Errorf("access токен не содержит subject")
	}

	value, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("access токен содержит некорректный subject")
	}
	if uint64(uint(value)) != value {
		return 0, fmt.Errorf("access токен содержит слишком большой subject")
	}

	return uint(value), nil
}

func randomTokenString(bytesCount int) (string, error) {
	if bytesCount <= 0 {
		return "", fmt.Errorf("bytesCount must be positive")
	}

	buffer := make([]byte, bytesCount)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hasAudience(audiences jwt.ClaimStrings, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}
