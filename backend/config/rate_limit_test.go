package config

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestDefaultRateLimitConfigUsesSafePositiveQuotas(t *testing.T) {
	defaults := DefaultRateLimitConfig()
	catalog := defaults.PolicyCatalog()

	if got, want := len(catalog), 30; got != want {
		t.Fatalf("default policy count = %d, want %d", got, want)
	}

	for name, quota := range catalog {
		if quota.Limit <= 0 {
			t.Errorf("%s default limit = %d, want > 0", name, quota.Limit)
		}
		if quota.Window <= 0 {
			t.Errorf("%s default window = %s, want > 0", name, quota.Window)
		}
	}

	assertDefaultQuota(t, catalog, "api-global-ip", 600, time.Minute)
	assertDefaultQuota(t, catalog, "auth-login-email", 10, 15*time.Minute)
	assertDefaultQuota(t, catalog, "authenticated-write-user", 120, time.Minute)
	assertDefaultQuota(t, catalog, "destructive-admin-user", 20, time.Hour)

	cfg := validConfigForTest(defaults)
	if err := cfg.NormalizeAndValidate(); err != nil {
		t.Fatalf("NormalizeAndValidate(defaults): %v", err)
	}
	if cfg.HTTP.TrustedProxies != nil {
		t.Fatalf("default trusted proxies = %#v, want nil (trust no proxies)", cfg.HTTP.TrustedProxies)
	}
}

func TestNewConfigReadsRateLimitAndTrustedProxyOverridesFromEnvironment(t *testing.T) {
	resetViperForConfigTest(t)
	t.Chdir(t.TempDir())
	t.Setenv("RATE_LIMIT_AUTH_LOGIN_IP_LIMIT", "17")
	t.Setenv("RATE_LIMIT_AUTH_LOGIN_IP_WINDOW", "42s")
	t.Setenv("TRUSTED_PROXIES", "192.0.2.10, 10.0.0.99/24, 192.0.2.10")
	t.Setenv("SECRET_KEY_JWT", strings.Repeat("s", 32))
	t.Setenv("JWT_KEY_ID", "test-primary")

	cfg := NewConfig()

	if got, want := cfg.RateLimit.AuthLoginIPLimit, 17; got != want {
		t.Fatalf("AuthLoginIPLimit = %d, want %d", got, want)
	}
	if got, want := cfg.RateLimit.AuthLoginIPWindow, 42*time.Second; got != want {
		t.Fatalf("AuthLoginIPWindow = %s, want %s", got, want)
	}
	wantProxies := []string{"192.0.2.10", "10.0.0.0/24"}
	if !reflect.DeepEqual(cfg.HTTP.TrustedProxies, wantProxies) {
		t.Fatalf("TrustedProxies = %#v, want %#v", cfg.HTTP.TrustedProxies, wantProxies)
	}
}

func TestNewConfigReadsRateLimitOverridesSetThroughViper(t *testing.T) {
	resetViperForConfigTest(t)
	t.Chdir(t.TempDir())
	viper.Set("RATE_LIMIT_REFERENCES_READ_IP_LIMIT", 23)
	viper.Set("RATE_LIMIT_REFERENCES_READ_IP_WINDOW", "90s")
	viper.Set("SECRET_KEY_JWT", strings.Repeat("s", 32))
	viper.Set("JWT_KEY_ID", "test-primary")

	cfg := NewConfig()

	if got, want := cfg.RateLimit.ReferencesReadIPLimit, 23; got != want {
		t.Fatalf("ReferencesReadIPLimit = %d, want %d", got, want)
	}
	if got, want := cfg.RateLimit.ReferencesReadIPWindow, 90*time.Second; got != want {
		t.Fatalf("ReferencesReadIPWindow = %s, want %s", got, want)
	}
}

func TestNormalizeAndValidateConfigRejectsNonPositiveRateLimitValues(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*RateLimitConfig)
		wantErr string
	}{
		{
			name: "zero limit",
			mutate: func(cfg *RateLimitConfig) {
				cfg.APIGlobalIPLimit = 0
			},
			wantErr: "api-global-ip limit must be > 0",
		},
		{
			name: "negative limit",
			mutate: func(cfg *RateLimitConfig) {
				cfg.AuthLoginIPLimit = -1
			},
			wantErr: "auth-login-ip limit must be > 0",
		},
		{
			name: "zero window",
			mutate: func(cfg *RateLimitConfig) {
				cfg.AuthLoginIPWindow = 0
			},
			wantErr: "auth-login-ip window must be > 0",
		},
		{
			name: "negative window",
			mutate: func(cfg *RateLimitConfig) {
				cfg.MemberModerationUserWindow = -time.Second
			},
			wantErr: "member-moderation-user window must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimits := DefaultRateLimitConfig()
			tt.mutate(&rateLimits)
			cfg := validConfigForTest(rateLimits)

			err := cfg.NormalizeAndValidate()
			if err == nil {
				t.Fatalf("NormalizeAndValidate() error = nil, want contains %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NormalizeAndValidate() error = %q, want contains %q", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeAndValidateConfigRejectsInvalidTrustedProxy(t *testing.T) {
	tests := []string{
		"not-an-ip",
		"192.0.2.1/99",
		"10.0.0.1,proxy.internal",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			cfg := Config{
				HTTP:         HTTPConfig{TrustedProxiesRaw: raw},
				RateLimit:    DefaultRateLimitConfig(),
				JWTSecretKey: strings.Repeat("s", 32),
				Auth: AuthConfig{
					Issuer:          "friendSheep",
					Audience:        "friendSheep-api",
					AccessTokenTTL:  20 * time.Minute,
					RefreshTokenTTL: 30 * 24 * time.Hour,
					ClockSkew:       30 * time.Second,
				},
			}

			err := cfg.NormalizeAndValidate()
			if err == nil {
				t.Fatalf("NormalizeAndValidate() error = nil for TRUSTED_PROXIES=%q", raw)
			}
			if !strings.Contains(err.Error(), "TRUSTED_PROXIES") {
				t.Fatalf("NormalizeAndValidate() error = %q, want TRUSTED_PROXIES context", err)
			}
		})
	}
}

func TestNormalizeAndValidateConfigRejectsUnsafeAuthSettings(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "short signing secret",
			mutate: func(cfg *Config) {
				cfg.JWTSecretKey = "too-short"
			},
			wantErr: "SECRET_KEY_JWT",
		},
		{
			name: "missing issuer",
			mutate: func(cfg *Config) {
				cfg.Auth.Issuer = ""
			},
			wantErr: "JWT_ISSUER",
		},
		{
			name: "missing key id",
			mutate: func(cfg *Config) {
				cfg.JWTKeyID = ""
			},
			wantErr: "JWT_KEY_ID",
		},
		{
			name: "previous key pair incomplete",
			mutate: func(cfg *Config) {
				cfg.JWTPreviousKeyID = "previous"
			},
			wantErr: "JWT_PREVIOUS_SECRET_KEY",
		},
		{
			name: "missing audience",
			mutate: func(cfg *Config) {
				cfg.Auth.Audience = ""
			},
			wantErr: "JWT_AUDIENCE",
		},
		{
			name: "access ttl too long",
			mutate: func(cfg *Config) {
				cfg.Auth.AccessTokenTTL = 2 * time.Hour
			},
			wantErr: "JWT_ACCESS_TTL",
		},
		{
			name: "refresh ttl not longer than access",
			mutate: func(cfg *Config) {
				cfg.Auth.RefreshTokenTTL = cfg.Auth.AccessTokenTTL
			},
			wantErr: "AUTH_REFRESH_TTL",
		},
		{
			name: "clock skew too large",
			mutate: func(cfg *Config) {
				cfg.Auth.ClockSkew = 10 * time.Minute
			},
			wantErr: "JWT_CLOCK_SKEW",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfigForTest(DefaultRateLimitConfig())
			tt.mutate(&cfg)

			err := cfg.NormalizeAndValidate()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NormalizeAndValidate() error = %v, want contains %q", err, tt.wantErr)
			}
		})
	}
}

func TestDerivedRateLimitHashSecretIsDomainSeparated(t *testing.T) {
	cfg := validConfigForTest(DefaultRateLimitConfig())
	derived := cfg.DerivedRateLimitHashSecret()
	if derived == "" || derived == cfg.JWTSecretKey {
		t.Fatalf("derived rate-limit secret must be non-empty and different from JWT signing secret")
	}
	if got := cfg.DerivedRateLimitHashSecret(); got != derived {
		t.Fatalf("derived rate-limit secret is not deterministic")
	}
}

func validConfigForTest(rateLimits RateLimitConfig) Config {
	return Config{
		JWTSecretKey: strings.Repeat("s", 32),
		JWTKeyID:     "test-primary",
		Auth: AuthConfig{
			Issuer:          "friendSheep",
			Audience:        "friendSheep-api",
			AccessTokenTTL:  20 * time.Minute,
			RefreshTokenTTL: 30 * 24 * time.Hour,
			ClockSkew:       30 * time.Second,
		},
		RateLimit: rateLimits,
	}
}

func assertDefaultQuota(
	t *testing.T,
	catalog map[string]RateLimitPolicyQuota,
	name string,
	wantLimit int,
	wantWindow time.Duration,
) {
	t.Helper()

	quota, ok := catalog[name]
	if !ok {
		t.Fatalf("default policy %q is missing", name)
	}
	if quota.Limit != wantLimit || quota.Window != wantWindow {
		t.Fatalf(
			"default policy %q = {Limit:%d Window:%s}, want {Limit:%d Window:%s}",
			name,
			quota.Limit,
			quota.Window,
			wantLimit,
			wantWindow,
		)
	}
}

func resetViperForConfigTest(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}
