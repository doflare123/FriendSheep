package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type HTTPConfig struct {
	TrustedProxiesRaw string   `mapstructure:"TRUSTED_PROXIES"`
	TrustedProxies    []string `mapstructure:"-"`
}

type RateLimitPolicyQuota struct {
	Limit  int
	Window time.Duration
}

type RateLimitConfig struct {
	APIGlobalIPLimit              int           `mapstructure:"RATE_LIMIT_API_GLOBAL_IP_LIMIT"`
	APIGlobalIPWindow             time.Duration `mapstructure:"RATE_LIMIT_API_GLOBAL_IP_WINDOW"`
	AuthenticatedReadUserLimit    int           `mapstructure:"RATE_LIMIT_AUTHENTICATED_READ_USER_LIMIT"`
	AuthenticatedReadUserWindow   time.Duration `mapstructure:"RATE_LIMIT_AUTHENTICATED_READ_USER_WINDOW"`
	AuthenticatedWriteUserLimit   int           `mapstructure:"RATE_LIMIT_AUTHENTICATED_WRITE_USER_LIMIT"`
	AuthenticatedWriteUserWindow  time.Duration `mapstructure:"RATE_LIMIT_AUTHENTICATED_WRITE_USER_WINDOW"`
	AuthLoginIPLimit              int           `mapstructure:"RATE_LIMIT_AUTH_LOGIN_IP_LIMIT"`
	AuthLoginIPWindow             time.Duration `mapstructure:"RATE_LIMIT_AUTH_LOGIN_IP_WINDOW"`
	AuthLoginEmailLimit           int           `mapstructure:"RATE_LIMIT_AUTH_LOGIN_EMAIL_LIMIT"`
	AuthLoginEmailWindow          time.Duration `mapstructure:"RATE_LIMIT_AUTH_LOGIN_EMAIL_WINDOW"`
	AuthRefreshIPLimit            int           `mapstructure:"RATE_LIMIT_AUTH_REFRESH_IP_LIMIT"`
	AuthRefreshIPWindow           time.Duration `mapstructure:"RATE_LIMIT_AUTH_REFRESH_IP_WINDOW"`
	RegisterSessionIPLimit        int           `mapstructure:"RATE_LIMIT_REGISTER_SESSION_IP_LIMIT"`
	RegisterSessionIPWindow       time.Duration `mapstructure:"RATE_LIMIT_REGISTER_SESSION_IP_WINDOW"`
	RegisterSessionEmailLimit     int           `mapstructure:"RATE_LIMIT_REGISTER_SESSION_EMAIL_LIMIT"`
	RegisterSessionEmailWindow    time.Duration `mapstructure:"RATE_LIMIT_REGISTER_SESSION_EMAIL_WINDOW"`
	RegisterVerifyIPLimit         int           `mapstructure:"RATE_LIMIT_REGISTER_VERIFY_IP_LIMIT"`
	RegisterVerifyIPWindow        time.Duration `mapstructure:"RATE_LIMIT_REGISTER_VERIFY_IP_WINDOW"`
	RegisterVerifySessionLimit    int           `mapstructure:"RATE_LIMIT_REGISTER_VERIFY_SESSION_LIMIT"`
	RegisterVerifySessionWindow   time.Duration `mapstructure:"RATE_LIMIT_REGISTER_VERIFY_SESSION_WINDOW"`
	RegisterCreateIPLimit         int           `mapstructure:"RATE_LIMIT_REGISTER_CREATE_IP_LIMIT"`
	RegisterCreateIPWindow        time.Duration `mapstructure:"RATE_LIMIT_REGISTER_CREATE_IP_WINDOW"`
	RegisterCreateSessionLimit    int           `mapstructure:"RATE_LIMIT_REGISTER_CREATE_SESSION_LIMIT"`
	RegisterCreateSessionWindow   time.Duration `mapstructure:"RATE_LIMIT_REGISTER_CREATE_SESSION_WINDOW"`
	PasswordChangeIPLimit         int           `mapstructure:"RATE_LIMIT_PASSWORD_CHANGE_IP_LIMIT"`
	PasswordChangeIPWindow        time.Duration `mapstructure:"RATE_LIMIT_PASSWORD_CHANGE_IP_WINDOW"`
	PasswordChangeSessionLimit    int           `mapstructure:"RATE_LIMIT_PASSWORD_CHANGE_SESSION_LIMIT"`
	PasswordChangeSessionWindow   time.Duration `mapstructure:"RATE_LIMIT_PASSWORD_CHANGE_SESSION_WINDOW"`
	EventsPopularIPLimit          int           `mapstructure:"RATE_LIMIT_EVENTS_POPULAR_IP_LIMIT"`
	EventsPopularIPWindow         time.Duration `mapstructure:"RATE_LIMIT_EVENTS_POPULAR_IP_WINDOW"`
	EventsSearchIPLimit           int           `mapstructure:"RATE_LIMIT_EVENTS_SEARCH_IP_LIMIT"`
	EventsSearchIPWindow          time.Duration `mapstructure:"RATE_LIMIT_EVENTS_SEARCH_IP_WINDOW"`
	ReferencesReadIPLimit         int           `mapstructure:"RATE_LIMIT_REFERENCES_READ_IP_LIMIT"`
	ReferencesReadIPWindow        time.Duration `mapstructure:"RATE_LIMIT_REFERENCES_READ_IP_WINDOW"`
	ImageUploadUserLimit          int           `mapstructure:"RATE_LIMIT_IMAGE_UPLOAD_USER_LIMIT"`
	ImageUploadUserWindow         time.Duration `mapstructure:"RATE_LIMIT_IMAGE_UPLOAD_USER_WINDOW"`
	GroupCreateUserLimit          int           `mapstructure:"RATE_LIMIT_GROUP_CREATE_USER_LIMIT"`
	GroupCreateUserWindow         time.Duration `mapstructure:"RATE_LIMIT_GROUP_CREATE_USER_WINDOW"`
	EventCreateUserLimit          int           `mapstructure:"RATE_LIMIT_EVENT_CREATE_USER_LIMIT"`
	EventCreateUserWindow         time.Duration `mapstructure:"RATE_LIMIT_EVENT_CREATE_USER_WINDOW"`
	MembershipChangeUserLimit     int           `mapstructure:"RATE_LIMIT_MEMBERSHIP_CHANGE_USER_LIMIT"`
	MembershipChangeUserWindow    time.Duration `mapstructure:"RATE_LIMIT_MEMBERSHIP_CHANGE_USER_WINDOW"`
	GroupInviteUserLimit          int           `mapstructure:"RATE_LIMIT_GROUP_INVITE_USER_LIMIT"`
	GroupInviteUserWindow         time.Duration `mapstructure:"RATE_LIMIT_GROUP_INVITE_USER_WINDOW"`
	GroupInviteResponseUserLimit  int           `mapstructure:"RATE_LIMIT_GROUP_INVITE_RESPONSE_USER_LIMIT"`
	GroupInviteResponseUserWindow time.Duration `mapstructure:"RATE_LIMIT_GROUP_INVITE_RESPONSE_USER_WINDOW"`
	JoinRequestBulkUserLimit      int           `mapstructure:"RATE_LIMIT_JOIN_REQUEST_BULK_USER_LIMIT"`
	JoinRequestBulkUserWindow     time.Duration `mapstructure:"RATE_LIMIT_JOIN_REQUEST_BULK_USER_WINDOW"`
	JoinRequestReviewUserLimit    int           `mapstructure:"RATE_LIMIT_JOIN_REQUEST_REVIEW_USER_LIMIT"`
	JoinRequestReviewUserWindow   time.Duration `mapstructure:"RATE_LIMIT_JOIN_REQUEST_REVIEW_USER_WINDOW"`
	GroupPermissionUserLimit      int           `mapstructure:"RATE_LIMIT_GROUP_PERMISSION_USER_LIMIT"`
	GroupPermissionUserWindow     time.Duration `mapstructure:"RATE_LIMIT_GROUP_PERMISSION_USER_WINDOW"`
	GroupUpdateUserLimit          int           `mapstructure:"RATE_LIMIT_GROUP_UPDATE_USER_LIMIT"`
	GroupUpdateUserWindow         time.Duration `mapstructure:"RATE_LIMIT_GROUP_UPDATE_USER_WINDOW"`
	EventUpdateUserLimit          int           `mapstructure:"RATE_LIMIT_EVENT_UPDATE_USER_LIMIT"`
	EventUpdateUserWindow         time.Duration `mapstructure:"RATE_LIMIT_EVENT_UPDATE_USER_WINDOW"`
	DestructiveAdminUserLimit     int           `mapstructure:"RATE_LIMIT_DESTRUCTIVE_ADMIN_USER_LIMIT"`
	DestructiveAdminUserWindow    time.Duration `mapstructure:"RATE_LIMIT_DESTRUCTIVE_ADMIN_USER_WINDOW"`
	MemberModerationUserLimit     int           `mapstructure:"RATE_LIMIT_MEMBER_MODERATION_USER_LIMIT"`
	MemberModerationUserWindow    time.Duration `mapstructure:"RATE_LIMIT_MEMBER_MODERATION_USER_WINDOW"`
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		APIGlobalIPLimit:              600,
		APIGlobalIPWindow:             time.Minute,
		AuthenticatedReadUserLimit:    300,
		AuthenticatedReadUserWindow:   time.Minute,
		AuthenticatedWriteUserLimit:   120,
		AuthenticatedWriteUserWindow:  time.Minute,
		AuthLoginIPLimit:              30,
		AuthLoginIPWindow:             5 * time.Minute,
		AuthLoginEmailLimit:           10,
		AuthLoginEmailWindow:          15 * time.Minute,
		AuthRefreshIPLimit:            60,
		AuthRefreshIPWindow:           5 * time.Minute,
		RegisterSessionIPLimit:        20,
		RegisterSessionIPWindow:       15 * time.Minute,
		RegisterSessionEmailLimit:     5,
		RegisterSessionEmailWindow:    15 * time.Minute,
		RegisterVerifyIPLimit:         30,
		RegisterVerifyIPWindow:        15 * time.Minute,
		RegisterVerifySessionLimit:    10,
		RegisterVerifySessionWindow:   15 * time.Minute,
		RegisterCreateIPLimit:         30,
		RegisterCreateIPWindow:        15 * time.Minute,
		RegisterCreateSessionLimit:    10,
		RegisterCreateSessionWindow:   15 * time.Minute,
		PasswordChangeIPLimit:         30,
		PasswordChangeIPWindow:        15 * time.Minute,
		PasswordChangeSessionLimit:    10,
		PasswordChangeSessionWindow:   15 * time.Minute,
		EventsPopularIPLimit:          60,
		EventsPopularIPWindow:         time.Minute,
		EventsSearchIPLimit:           60,
		EventsSearchIPWindow:          time.Minute,
		ReferencesReadIPLimit:         120,
		ReferencesReadIPWindow:        time.Minute,
		ImageUploadUserLimit:          30,
		ImageUploadUserWindow:         time.Hour,
		GroupCreateUserLimit:          10,
		GroupCreateUserWindow:         time.Hour,
		EventCreateUserLimit:          30,
		EventCreateUserWindow:         time.Hour,
		MembershipChangeUserLimit:     30,
		MembershipChangeUserWindow:    5 * time.Minute,
		GroupInviteUserLimit:          30,
		GroupInviteUserWindow:         10 * time.Minute,
		GroupInviteResponseUserLimit:  30,
		GroupInviteResponseUserWindow: 10 * time.Minute,
		JoinRequestBulkUserLimit:      10,
		JoinRequestBulkUserWindow:     10 * time.Minute,
		JoinRequestReviewUserLimit:    30,
		JoinRequestReviewUserWindow:   10 * time.Minute,
		GroupPermissionUserLimit:      20,
		GroupPermissionUserWindow:     10 * time.Minute,
		GroupUpdateUserLimit:          30,
		GroupUpdateUserWindow:         10 * time.Minute,
		EventUpdateUserLimit:          30,
		EventUpdateUserWindow:         10 * time.Minute,
		DestructiveAdminUserLimit:     20,
		DestructiveAdminUserWindow:    time.Hour,
		MemberModerationUserLimit:     30,
		MemberModerationUserWindow:    10 * time.Minute,
	}
}

func (c RateLimitConfig) PolicyCatalog() map[string]RateLimitPolicyQuota {
	return map[string]RateLimitPolicyQuota{
		"api-global-ip":              {Limit: c.APIGlobalIPLimit, Window: c.APIGlobalIPWindow},
		"authenticated-read-user":    {Limit: c.AuthenticatedReadUserLimit, Window: c.AuthenticatedReadUserWindow},
		"authenticated-write-user":   {Limit: c.AuthenticatedWriteUserLimit, Window: c.AuthenticatedWriteUserWindow},
		"auth-login-ip":              {Limit: c.AuthLoginIPLimit, Window: c.AuthLoginIPWindow},
		"auth-login-email":           {Limit: c.AuthLoginEmailLimit, Window: c.AuthLoginEmailWindow},
		"auth-refresh-ip":            {Limit: c.AuthRefreshIPLimit, Window: c.AuthRefreshIPWindow},
		"register-session-ip":        {Limit: c.RegisterSessionIPLimit, Window: c.RegisterSessionIPWindow},
		"register-session-email":     {Limit: c.RegisterSessionEmailLimit, Window: c.RegisterSessionEmailWindow},
		"register-verify-ip":         {Limit: c.RegisterVerifyIPLimit, Window: c.RegisterVerifyIPWindow},
		"register-verify-session":    {Limit: c.RegisterVerifySessionLimit, Window: c.RegisterVerifySessionWindow},
		"register-create-ip":         {Limit: c.RegisterCreateIPLimit, Window: c.RegisterCreateIPWindow},
		"register-create-session":    {Limit: c.RegisterCreateSessionLimit, Window: c.RegisterCreateSessionWindow},
		"password-change-ip":         {Limit: c.PasswordChangeIPLimit, Window: c.PasswordChangeIPWindow},
		"password-change-session":    {Limit: c.PasswordChangeSessionLimit, Window: c.PasswordChangeSessionWindow},
		"events-popular-ip":          {Limit: c.EventsPopularIPLimit, Window: c.EventsPopularIPWindow},
		"events-search-ip":           {Limit: c.EventsSearchIPLimit, Window: c.EventsSearchIPWindow},
		"references-read-ip":         {Limit: c.ReferencesReadIPLimit, Window: c.ReferencesReadIPWindow},
		"image-upload-user":          {Limit: c.ImageUploadUserLimit, Window: c.ImageUploadUserWindow},
		"group-create-user":          {Limit: c.GroupCreateUserLimit, Window: c.GroupCreateUserWindow},
		"event-create-user":          {Limit: c.EventCreateUserLimit, Window: c.EventCreateUserWindow},
		"membership-change-user":     {Limit: c.MembershipChangeUserLimit, Window: c.MembershipChangeUserWindow},
		"group-invite-user":          {Limit: c.GroupInviteUserLimit, Window: c.GroupInviteUserWindow},
		"group-invite-response-user": {Limit: c.GroupInviteResponseUserLimit, Window: c.GroupInviteResponseUserWindow},
		"join-request-bulk-user":     {Limit: c.JoinRequestBulkUserLimit, Window: c.JoinRequestBulkUserWindow},
		"join-request-review-user":   {Limit: c.JoinRequestReviewUserLimit, Window: c.JoinRequestReviewUserWindow},
		"group-permission-user":      {Limit: c.GroupPermissionUserLimit, Window: c.GroupPermissionUserWindow},
		"group-update-user":          {Limit: c.GroupUpdateUserLimit, Window: c.GroupUpdateUserWindow},
		"event-update-user":          {Limit: c.EventUpdateUserLimit, Window: c.EventUpdateUserWindow},
		"destructive-admin-user":     {Limit: c.DestructiveAdminUserLimit, Window: c.DestructiveAdminUserWindow},
		"member-moderation-user":     {Limit: c.MemberModerationUserLimit, Window: c.MemberModerationUserWindow},
	}
}

func applyConfigDefaults() {
	defaults := DefaultRateLimitConfig()
	viper.SetDefault("SECRET_KEY_JWT", "")
	viper.SetDefault("JWT_ISSUER", "friendSheep")
	viper.SetDefault("JWT_AUDIENCE", "friendSheep-api")
	viper.SetDefault("JWT_KEY_ID", "primary-v1")
	viper.SetDefault("JWT_PREVIOUS_SECRET_KEY", "")
	viper.SetDefault("JWT_PREVIOUS_KEY_ID", "")
	viper.SetDefault("JWT_ACCESS_TTL", "20m")
	viper.SetDefault("AUTH_REFRESH_TTL", "720h")
	viper.SetDefault("JWT_CLOCK_SKEW", "30s")

	setPolicyDefault := func(limitKey string, limit int, windowKey string, window time.Duration) {
		viper.SetDefault(limitKey, limit)
		viper.SetDefault(windowKey, window.String())
	}

	viper.SetDefault("TRUSTED_PROXIES", "")
	setPolicyDefault("RATE_LIMIT_API_GLOBAL_IP_LIMIT", defaults.APIGlobalIPLimit, "RATE_LIMIT_API_GLOBAL_IP_WINDOW", defaults.APIGlobalIPWindow)
	setPolicyDefault("RATE_LIMIT_AUTHENTICATED_READ_USER_LIMIT", defaults.AuthenticatedReadUserLimit, "RATE_LIMIT_AUTHENTICATED_READ_USER_WINDOW", defaults.AuthenticatedReadUserWindow)
	setPolicyDefault("RATE_LIMIT_AUTHENTICATED_WRITE_USER_LIMIT", defaults.AuthenticatedWriteUserLimit, "RATE_LIMIT_AUTHENTICATED_WRITE_USER_WINDOW", defaults.AuthenticatedWriteUserWindow)
	setPolicyDefault("RATE_LIMIT_AUTH_LOGIN_IP_LIMIT", defaults.AuthLoginIPLimit, "RATE_LIMIT_AUTH_LOGIN_IP_WINDOW", defaults.AuthLoginIPWindow)
	setPolicyDefault("RATE_LIMIT_AUTH_LOGIN_EMAIL_LIMIT", defaults.AuthLoginEmailLimit, "RATE_LIMIT_AUTH_LOGIN_EMAIL_WINDOW", defaults.AuthLoginEmailWindow)
	setPolicyDefault("RATE_LIMIT_AUTH_REFRESH_IP_LIMIT", defaults.AuthRefreshIPLimit, "RATE_LIMIT_AUTH_REFRESH_IP_WINDOW", defaults.AuthRefreshIPWindow)
	setPolicyDefault("RATE_LIMIT_REGISTER_SESSION_IP_LIMIT", defaults.RegisterSessionIPLimit, "RATE_LIMIT_REGISTER_SESSION_IP_WINDOW", defaults.RegisterSessionIPWindow)
	setPolicyDefault("RATE_LIMIT_REGISTER_SESSION_EMAIL_LIMIT", defaults.RegisterSessionEmailLimit, "RATE_LIMIT_REGISTER_SESSION_EMAIL_WINDOW", defaults.RegisterSessionEmailWindow)
	setPolicyDefault("RATE_LIMIT_REGISTER_VERIFY_IP_LIMIT", defaults.RegisterVerifyIPLimit, "RATE_LIMIT_REGISTER_VERIFY_IP_WINDOW", defaults.RegisterVerifyIPWindow)
	setPolicyDefault("RATE_LIMIT_REGISTER_VERIFY_SESSION_LIMIT", defaults.RegisterVerifySessionLimit, "RATE_LIMIT_REGISTER_VERIFY_SESSION_WINDOW", defaults.RegisterVerifySessionWindow)
	setPolicyDefault("RATE_LIMIT_REGISTER_CREATE_IP_LIMIT", defaults.RegisterCreateIPLimit, "RATE_LIMIT_REGISTER_CREATE_IP_WINDOW", defaults.RegisterCreateIPWindow)
	setPolicyDefault("RATE_LIMIT_REGISTER_CREATE_SESSION_LIMIT", defaults.RegisterCreateSessionLimit, "RATE_LIMIT_REGISTER_CREATE_SESSION_WINDOW", defaults.RegisterCreateSessionWindow)
	setPolicyDefault("RATE_LIMIT_PASSWORD_CHANGE_IP_LIMIT", defaults.PasswordChangeIPLimit, "RATE_LIMIT_PASSWORD_CHANGE_IP_WINDOW", defaults.PasswordChangeIPWindow)
	setPolicyDefault("RATE_LIMIT_PASSWORD_CHANGE_SESSION_LIMIT", defaults.PasswordChangeSessionLimit, "RATE_LIMIT_PASSWORD_CHANGE_SESSION_WINDOW", defaults.PasswordChangeSessionWindow)
	setPolicyDefault("RATE_LIMIT_EVENTS_POPULAR_IP_LIMIT", defaults.EventsPopularIPLimit, "RATE_LIMIT_EVENTS_POPULAR_IP_WINDOW", defaults.EventsPopularIPWindow)
	setPolicyDefault("RATE_LIMIT_EVENTS_SEARCH_IP_LIMIT", defaults.EventsSearchIPLimit, "RATE_LIMIT_EVENTS_SEARCH_IP_WINDOW", defaults.EventsSearchIPWindow)
	setPolicyDefault("RATE_LIMIT_REFERENCES_READ_IP_LIMIT", defaults.ReferencesReadIPLimit, "RATE_LIMIT_REFERENCES_READ_IP_WINDOW", defaults.ReferencesReadIPWindow)
	setPolicyDefault("RATE_LIMIT_IMAGE_UPLOAD_USER_LIMIT", defaults.ImageUploadUserLimit, "RATE_LIMIT_IMAGE_UPLOAD_USER_WINDOW", defaults.ImageUploadUserWindow)
	setPolicyDefault("RATE_LIMIT_GROUP_CREATE_USER_LIMIT", defaults.GroupCreateUserLimit, "RATE_LIMIT_GROUP_CREATE_USER_WINDOW", defaults.GroupCreateUserWindow)
	setPolicyDefault("RATE_LIMIT_EVENT_CREATE_USER_LIMIT", defaults.EventCreateUserLimit, "RATE_LIMIT_EVENT_CREATE_USER_WINDOW", defaults.EventCreateUserWindow)
	setPolicyDefault("RATE_LIMIT_MEMBERSHIP_CHANGE_USER_LIMIT", defaults.MembershipChangeUserLimit, "RATE_LIMIT_MEMBERSHIP_CHANGE_USER_WINDOW", defaults.MembershipChangeUserWindow)
	setPolicyDefault("RATE_LIMIT_GROUP_INVITE_USER_LIMIT", defaults.GroupInviteUserLimit, "RATE_LIMIT_GROUP_INVITE_USER_WINDOW", defaults.GroupInviteUserWindow)
	setPolicyDefault("RATE_LIMIT_GROUP_INVITE_RESPONSE_USER_LIMIT", defaults.GroupInviteResponseUserLimit, "RATE_LIMIT_GROUP_INVITE_RESPONSE_USER_WINDOW", defaults.GroupInviteResponseUserWindow)
	setPolicyDefault("RATE_LIMIT_JOIN_REQUEST_BULK_USER_LIMIT", defaults.JoinRequestBulkUserLimit, "RATE_LIMIT_JOIN_REQUEST_BULK_USER_WINDOW", defaults.JoinRequestBulkUserWindow)
	setPolicyDefault("RATE_LIMIT_JOIN_REQUEST_REVIEW_USER_LIMIT", defaults.JoinRequestReviewUserLimit, "RATE_LIMIT_JOIN_REQUEST_REVIEW_USER_WINDOW", defaults.JoinRequestReviewUserWindow)
	setPolicyDefault("RATE_LIMIT_GROUP_PERMISSION_USER_LIMIT", defaults.GroupPermissionUserLimit, "RATE_LIMIT_GROUP_PERMISSION_USER_WINDOW", defaults.GroupPermissionUserWindow)
	setPolicyDefault("RATE_LIMIT_GROUP_UPDATE_USER_LIMIT", defaults.GroupUpdateUserLimit, "RATE_LIMIT_GROUP_UPDATE_USER_WINDOW", defaults.GroupUpdateUserWindow)
	setPolicyDefault("RATE_LIMIT_EVENT_UPDATE_USER_LIMIT", defaults.EventUpdateUserLimit, "RATE_LIMIT_EVENT_UPDATE_USER_WINDOW", defaults.EventUpdateUserWindow)
	setPolicyDefault("RATE_LIMIT_DESTRUCTIVE_ADMIN_USER_LIMIT", defaults.DestructiveAdminUserLimit, "RATE_LIMIT_DESTRUCTIVE_ADMIN_USER_WINDOW", defaults.DestructiveAdminUserWindow)
	setPolicyDefault("RATE_LIMIT_MEMBER_MODERATION_USER_LIMIT", defaults.MemberModerationUserLimit, "RATE_LIMIT_MEMBER_MODERATION_USER_WINDOW", defaults.MemberModerationUserWindow)
}

func (c *Config) NormalizeAndValidate() error {
	var errs []error

	if len(strings.TrimSpace(c.JWTSecretKey)) < 32 {
		errs = append(errs, fmt.Errorf("SECRET_KEY_JWT must contain at least 32 characters"))
	}
	if strings.TrimSpace(c.JWTKeyID) == "" {
		errs = append(errs, fmt.Errorf("JWT_KEY_ID must not be empty"))
	}
	previousSecret := strings.TrimSpace(c.JWTPreviousSecretKey)
	previousKeyID := strings.TrimSpace(c.JWTPreviousKeyID)
	if (previousSecret == "") != (previousKeyID == "") {
		errs = append(errs, fmt.Errorf("JWT_PREVIOUS_SECRET_KEY and JWT_PREVIOUS_KEY_ID must be configured together"))
	}
	if previousSecret != "" && len(previousSecret) < 32 {
		errs = append(errs, fmt.Errorf("JWT_PREVIOUS_SECRET_KEY must contain at least 32 characters"))
	}
	if previousKeyID != "" && previousKeyID == strings.TrimSpace(c.JWTKeyID) {
		errs = append(errs, fmt.Errorf("JWT_PREVIOUS_KEY_ID must differ from JWT_KEY_ID"))
	}
	if strings.TrimSpace(c.Auth.Issuer) == "" {
		errs = append(errs, fmt.Errorf("JWT_ISSUER must not be empty"))
	}
	if strings.TrimSpace(c.Auth.Audience) == "" {
		errs = append(errs, fmt.Errorf("JWT_AUDIENCE must not be empty"))
	}
	if c.Auth.AccessTokenTTL < time.Minute || c.Auth.AccessTokenTTL > time.Hour {
		errs = append(errs, fmt.Errorf("JWT_ACCESS_TTL must be between 1m and 1h"))
	}
	if c.Auth.RefreshTokenTTL <= c.Auth.AccessTokenTTL || c.Auth.RefreshTokenTTL > 90*24*time.Hour {
		errs = append(errs, fmt.Errorf("AUTH_REFRESH_TTL must be greater than JWT_ACCESS_TTL and no more than 2160h"))
	}
	if c.Auth.ClockSkew < 0 || c.Auth.ClockSkew > 5*time.Minute {
		errs = append(errs, fmt.Errorf("JWT_CLOCK_SKEW must be between 0 and 5m"))
	}

	trustedProxies, err := parseTrustedProxies(c.HTTP.TrustedProxiesRaw)
	if err != nil {
		errs = append(errs, fmt.Errorf("TRUSTED_PROXIES: %w", err))
	} else {
		c.HTTP.TrustedProxies = trustedProxies
	}

	for policyName, quota := range c.RateLimit.PolicyCatalog() {
		if quota.Limit <= 0 {
			errs = append(errs, fmt.Errorf("%s limit must be > 0", policyName))
		}
		if quota.Window <= 0 {
			errs = append(errs, fmt.Errorf("%s window must be > 0", policyName))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func parseTrustedProxies(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	trusted := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}

		normalized := value
		if strings.Contains(value, "/") {
			_, network, err := net.ParseCIDR(value)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR %q: %w", value, err)
			}
			normalized = network.String()
		} else {
			ip := net.ParseIP(value)
			if ip == nil {
				return nil, fmt.Errorf("invalid IP %q", value)
			}
			normalized = ip.String()
		}

		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		trusted = append(trusted, normalized)
	}

	if len(trusted) == 0 {
		return nil, nil
	}
	return trusted, nil
}
