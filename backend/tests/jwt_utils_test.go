package tests

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"friendship/utils"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testJWTSecret   = "test-secret"
	testJWTIssuer   = "friendship-test"
	testJWTAudience = "friendship-test-api"
)

func TestJWTUtilsAccessTokenHasMinimalSecurityClaims(t *testing.T) {
	jwtUtils := newConfiguredJWTUtils()

	tokenString, tokenID, expiresAt, err := jwtUtils.GenerateAccessToken(99, "session-99")
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if tokenID == "" {
		t.Fatal("GenerateAccessToken returned empty token ID")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expiresAt = %v, want future time", expiresAt)
	}

	header, payload := decodeJWTParts(t, tokenString)
	if got := header["alg"]; got != jwt.SigningMethodHS256.Alg() {
		t.Fatalf("alg = %#v, want %q", got, jwt.SigningMethodHS256.Alg())
	}
	if got := header["kid"]; got != "primary" {
		t.Fatalf("kid = %#v, want primary", got)
	}

	wantPayloadKeys := map[string]bool{
		"sub":       true,
		"iss":       true,
		"aud":       true,
		"exp":       true,
		"iat":       true,
		"jti":       true,
		"sid":       true,
		"token_use": true,
	}
	if len(payload) != len(wantPayloadKeys) {
		t.Fatalf("payload keys = %v, want exactly %v", mapKeys(payload), mapKeysBool(wantPayloadKeys))
	}
	for key := range wantPayloadKeys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("payload is missing required claim %q: %v", key, payload)
		}
	}
	for _, forbidden := range []string{"id", "username", "name", "us", "image", "typ"} {
		if _, ok := payload[forbidden]; ok {
			t.Fatalf("payload unexpectedly contains profile/legacy claim %q: %v", forbidden, payload)
		}
	}
	if payload["sub"] != "99" {
		t.Fatalf("sub = %#v, want %q", payload["sub"], "99")
	}
	if payload["sid"] != "session-99" {
		t.Fatalf("sid = %#v, want %q", payload["sid"], "session-99")
	}
	if payload["jti"] != tokenID {
		t.Fatalf("jti = %#v, want %q", payload["jti"], tokenID)
	}
	if payload["token_use"] != "access" {
		t.Fatalf("token_use = %#v, want access", payload["token_use"])
	}

	claims, err := jwtUtils.ParseAccessToken(tokenString)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.UserID != 99 || claims.Subject != "99" {
		t.Fatalf("parsed identity = userID %d subject %q, want 99", claims.UserID, claims.Subject)
	}
	if claims.SessionID != "session-99" || claims.ID != tokenID || claims.TokenID != tokenID || claims.TokenUse != "access" {
		t.Fatalf("parsed security claims = sid %q jti %q derived-jti %q use %q", claims.SessionID, claims.ID, claims.TokenID, claims.TokenUse)
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil || claims.Issuer != testJWTIssuer || !containsAudience(claims.Audience, testJWTAudience) {
		t.Fatalf("parsed registered claims are incomplete: %#v", claims.RegisteredClaims)
	}
}

func TestJWTUtilsRejectsInvalidAccessTokenSecurityContract(t *testing.T) {
	jwtUtils := newConfiguredJWTUtils()
	now := time.Now().UTC().Truncate(time.Second)
	validClaims := func() utils.Claims {
		return utils.Claims{
			SessionID: "session-42",
			TokenUse:  "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "42",
				Issuer:    testJWTIssuer,
				Audience:  jwt.ClaimStrings{testJWTAudience},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				ID:        "token-42",
			},
		}
	}

	tests := []struct {
		name  string
		build func() string
	}{
		{
			name: "missing issuer",
			build: func() string {
				claims := validClaims()
				claims.Issuer = ""
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "wrong issuer",
			build: func() string {
				claims := validClaims()
				claims.Issuer = "attacker"
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "missing audience",
			build: func() string {
				claims := validClaims()
				claims.Audience = nil
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "wrong audience",
			build: func() string {
				claims := validClaims()
				claims.Audience = jwt.ClaimStrings{"other-api"}
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "wrong algorithm",
			build: func() string {
				return signAccessClaims(t, jwt.SigningMethodHS384, validClaims(), testJWTSecret)
			},
		},
		{
			name: "missing expiration",
			build: func() string {
				claims := validClaims()
				claims.ExpiresAt = nil
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "missing issued at",
			build: func() string {
				claims := validClaims()
				claims.IssuedAt = nil
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "issued in the future",
			build: func() string {
				claims := validClaims()
				claims.IssuedAt = jwt.NewNumericDate(now.Add(10 * time.Minute))
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(70 * time.Minute))
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "expiration before issued at",
			build: func() string {
				claims := validClaims()
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute))
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "excessive lifetime",
			build: func() string {
				claims := validClaims()
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(2 * time.Hour))
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "missing subject",
			build: func() string {
				claims := validClaims()
				claims.Subject = ""
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "zero subject",
			build: func() string {
				claims := validClaims()
				claims.Subject = "0"
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "negative subject",
			build: func() string {
				claims := validClaims()
				claims.Subject = "-1"
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "non-numeric subject",
			build: func() string {
				claims := validClaims()
				claims.Subject = "not-a-user"
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "wrong token use",
			build: func() string {
				claims := validClaims()
				claims.TokenUse = "refresh"
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "legacy profile claim",
			build: func() string {
				return signRawAccessPayload(t, map[string]interface{}{
					"sub": "42", "iss": testJWTIssuer, "aud": []string{testJWTAudience},
					"exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "jti": "token-42",
					"sid": "session-42", "token_use": "access", "image": "avatar.png",
				}, testJWTSecret, "primary")
			},
		},
		{
			name: "missing token use",
			build: func() string {
				claims := validClaims()
				claims.TokenUse = ""
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "missing session ID",
			build: func() string {
				claims := validClaims()
				claims.SessionID = ""
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "missing token ID",
			build: func() string {
				claims := validClaims()
				claims.ID = ""
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "missing key ID",
			build: func() string {
				return signAccessClaimsWithKeyID(t, jwt.SigningMethodHS256, validClaims(), testJWTSecret, "")
			},
		},
		{
			name: "unknown key ID",
			build: func() string {
				return signAccessClaimsWithKeyID(t, jwt.SigningMethodHS256, validClaims(), testJWTSecret, "unknown")
			},
		},
		{
			name: "expired",
			build: func() string {
				claims := validClaims()
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute))
				return signAccessClaims(t, jwt.SigningMethodHS256, claims, testJWTSecret)
			},
		},
		{
			name: "tampered signature",
			build: func() string {
				return tamperJWTSignature(t, signAccessClaims(t, jwt.SigningMethodHS256, validClaims(), testJWTSecret))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := jwtUtils.ParseAccessToken(tt.build()); err == nil {
				t.Fatal("ParseAccessToken accepted invalid token")
			}
		})
	}
}

func TestJWTUtilsRejectsInvalidGenerationInputsAndEmptySecret(t *testing.T) {
	jwtUtils := newConfiguredJWTUtils()
	if _, _, _, err := jwtUtils.GenerateAccessToken(0, "session"); err == nil {
		t.Fatal("GenerateAccessToken accepted zero user ID")
	}
	if _, _, _, err := jwtUtils.GenerateAccessToken(1, "  "); err == nil {
		t.Fatal("GenerateAccessToken accepted empty session ID")
	}

	emptySecret := utils.NewJWTUtilsWithConfig("", utils.JWTConfig{
		Issuer:   testJWTIssuer,
		Audience: testJWTAudience,
	})
	if _, _, _, err := emptySecret.GenerateAccessToken(1, "session"); err == nil {
		t.Fatal("GenerateAccessToken accepted empty secret")
	}
	if err := emptySecret.ValidateToken("token"); err == nil {
		t.Fatal("ValidateToken accepted empty secret")
	}
}

func TestJWTUtilsSupportsVerificationKeyRotation(t *testing.T) {
	oldKey := strings.Repeat("o", 32)
	newKey := strings.Repeat("n", 32)
	oldIssuer := utils.NewJWTUtilsWithConfig(oldKey, utils.JWTConfig{
		Issuer: testJWTIssuer, Audience: testJWTAudience, KeyID: "old", AccessTTL: time.Hour,
	})
	oldToken, _, _, err := oldIssuer.GenerateAccessToken(12, "session-old")
	if err != nil {
		t.Fatalf("GenerateAccessToken(old): %v", err)
	}

	rotatedVerifier := utils.NewJWTUtilsWithConfig(newKey, utils.JWTConfig{
		Issuer: testJWTIssuer, Audience: testJWTAudience, KeyID: "new", AccessTTL: time.Hour,
		VerificationKeys: map[string]string{"old": oldKey},
	})
	if _, err := rotatedVerifier.ParseAccessToken(oldToken); err != nil {
		t.Fatalf("rotated verifier rejected previous key token: %v", err)
	}

	withoutOldKey := utils.NewJWTUtilsWithConfig(newKey, utils.JWTConfig{
		Issuer: testJWTIssuer, Audience: testJWTAudience, KeyID: "new", AccessTTL: time.Hour,
	})
	if _, err := withoutOldKey.ParseAccessToken(oldToken); err == nil {
		t.Fatal("verifier accepted token signed by unknown kid")
	}
}

func newConfiguredJWTUtils() *utils.JWTUtils {
	return utils.NewJWTUtilsWithConfig(testJWTSecret, utils.JWTConfig{
		Issuer:    testJWTIssuer,
		Audience:  testJWTAudience,
		KeyID:     "primary",
		AccessTTL: time.Hour,
	})
}

func signAccessClaims(t *testing.T, method jwt.SigningMethod, claims utils.Claims, secret string) string {
	return signAccessClaimsWithKeyID(t, method, claims, secret, "primary")
}

func signAccessClaimsWithKeyID(t *testing.T, method jwt.SigningMethod, claims utils.Claims, secret, keyID string) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	if keyID != "" {
		token.Header["kid"] = keyID
	}
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func signRawAccessPayload(t *testing.T, payload map[string]interface{}, secret, keyID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
	token.Header["kid"] = keyID
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign raw access payload: %v", err)
	}
	return signed
}

func decodeJWTParts(t *testing.T, tokenString string) (map[string]interface{}, map[string]interface{}) {
	t.Helper()
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT parts = %d, want 3", len(parts))
	}
	decode := func(label, encoded string) map[string]interface{} {
		t.Helper()
		raw, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatalf("decode %s: %v", label, err)
		}
		var value map[string]interface{}
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatalf("unmarshal %s: %v", label, err)
		}
		return value
	}
	return decode("header", parts[0]), decode("payload", parts[1])
}

func tamperJWTSignature(t *testing.T, tokenString string) string {
	t.Helper()
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 || parts[2] == "" {
		t.Fatalf("cannot tamper malformed token %q", tokenString)
	}
	replacement := byte('a')
	if parts[2][0] == replacement {
		replacement = 'b'
	}
	parts[2] = string(replacement) + parts[2][1:]
	return strings.Join(parts, ".")
}

func mapKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func mapKeysBool(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func containsAudience(audience jwt.ClaimStrings, want string) bool {
	for _, value := range audience {
		if value == want {
			return true
		}
	}
	return false
}
