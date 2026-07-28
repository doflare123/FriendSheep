package middlewares

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"friendship/config"
	"friendship/logger"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

const (
	rateLimitKeyPrefix   = "friendship:rate-limit:v1"
	maxRateLimitBodySize = 64 << 10
)

var (
	errRateLimitBodyTooLarge = errors.New("request body exceeds rate limit inspection limit")
	errRateLimitBodyRead     = errors.New("request body cannot be read for rate limiting")
)

type RateLimitDecision struct {
	Allowed    bool
	Count      int
	ResetAfter time.Duration
}

type RateLimitStore interface {
	Take(ctx context.Context, key string, limit int, window time.Duration) (RateLimitDecision, error)
}

type rateLimitFailureMode uint8

const (
	rateLimitFailOpen rateLimitFailureMode = iota
	rateLimitFailClosed
)

type rateLimitKeyKind uint8

const (
	rateLimitByIP rateLimitKeyKind = iota
	rateLimitByJSONField
	rateLimitByUser
)

type rateLimitPolicy struct {
	name      string
	limit     int
	window    time.Duration
	keyKind   rateLimitKeyKind
	jsonField string
	onFailure rateLimitFailureMode
}

type RateLimitMiddleware struct {
	logger     logger.Logger
	store      RateLimitStore
	hashSecret []byte
	catalog    rateLimitCatalog
}

func NewRateLimitMiddleware(log logger.Logger, store RateLimitStore, hashSecret string) *RateLimitMiddleware {
	return NewRateLimitMiddlewareWithConfig(log, store, hashSecret, config.DefaultRateLimitConfig())
}

func NewRateLimitMiddlewareWithConfig(log logger.Logger, store RateLimitStore, hashSecret string, cfg config.RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		logger:     log,
		store:      store,
		hashSecret: []byte(hashSecret),
		catalog:    newRateLimitCatalog(cfg),
	}
}

// APIBaseLimit защищает каждый активный маршрут API по IP-адресу клиента
// и применяет более строгие лимиты к публичным маршрутам и точкам до аутентификации.
func (m *RateLimitMiddleware) APIBaseLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isV2APIPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		policies := []rateLimitPolicy{
			m.policy("api-global-ip", rateLimitByIP, "", rateLimitFailOpen),
		}
		policies = append(policies, publicRateLimitPoliciesWithCatalog(c.Request.Method, normalizedFullPath(c), m.policyCatalog())...)

		for _, policy := range policies {
			if !m.enforce(c, policy, 0) {
				return
			}
		}

		c.Next()
	}
}

// ApplyAuthenticatedLimit после проверки токена JWT применяет квоту пользователя.
// Общие квоты пропускают запросы при временной недоступности хранилища лимитов,
// а чувствительные к злоупотреблениям изменения дополнительно блокируются.
func (m *RateLimitMiddleware) ApplyAuthenticatedLimit(c *gin.Context, userID uint) bool {
	if m == nil || userID == 0 {
		return true
	}

	genericPolicy := m.policy("authenticated-read-user", rateLimitByUser, "", rateLimitFailOpen)
	if isMutationMethod(c.Request.Method) {
		genericPolicy = m.policy("authenticated-write-user", rateLimitByUser, "", rateLimitFailOpen)
	}

	if !m.enforce(c, genericPolicy, userID) {
		return false
	}

	for _, policy := range authenticatedRateLimitPoliciesWithCatalog(c.Request.Method, normalizedFullPath(c), m.policyCatalog()) {
		if !m.enforce(c, policy, userID) {
			return false
		}
	}

	return true
}

func (m *RateLimitMiddleware) enforce(c *gin.Context, policy rateLimitPolicy, userID uint) bool {
	if m == nil || m.store == nil || policy.limit <= 0 || policy.window <= 0 {
		return m.handleStoreFailure(c, policy, nil)
	}

	identity, err := m.policyIdentity(c, policy, userID)
	if err != nil {
		return abortRateLimitIdentityError(c, err)
	}
	if identity == "" {
		identity = clientAddress(c)
	}

	decision, err := m.store.Take(
		c.Request.Context(),
		m.rateLimitKey(policy.name, identity),
		policy.limit,
		policy.window,
	)
	if err != nil {
		return m.handleStoreFailure(c, policy, err)
	}

	setRateLimitHeaders(c, policy, decision)
	if decision.Allowed {
		return true
	}

	retryAfter := resetSeconds(decision.ResetAfter, policy.window)
	c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
	utils.AbortJSONError(
		c,
		http.StatusTooManyRequests,
		"rate_limit_exceeded",
		utils.WithMessage("Слишком много запросов. Повторите попытку позже"),
	)
	return false
}

func (m *RateLimitMiddleware) handleStoreFailure(c *gin.Context, policy rateLimitPolicy, err error) bool {
	if m != nil && m.logger != nil && err != nil {
		m.logger.Warn("Rate limiter store unavailable", "policy", policy.name, "error", err)
	}
	if policy.onFailure == rateLimitFailOpen {
		return true
	}

	utils.AbortJSONError(
		c,
		http.StatusServiceUnavailable,
		"rate_limit_unavailable",
		utils.WithMessage("Защита от частых запросов временно недоступна"),
	)
	return false
}

func (m *RateLimitMiddleware) policyIdentity(
	c *gin.Context,
	policy rateLimitPolicy,
	userID uint,
) (string, error) {
	switch policy.keyKind {
	case rateLimitByUser:
		if userID == 0 {
			return "", nil
		}
		return "user:" + strconv.FormatUint(uint64(userID), 10), nil
	case rateLimitByJSONField:
		value, err := requestJSONField(c, policy.jsonField)
		if err != nil {
			return "", err
		}
		if value == "" {
			return "", nil
		}
		return policy.jsonField + ":" + strings.ToLower(strings.TrimSpace(value)), nil
	default:
		return "ip:" + clientAddress(c), nil
	}
}

func (m *RateLimitMiddleware) policyCatalog() rateLimitCatalog {
	if m == nil || len(m.catalog) == 0 {
		return defaultRateLimitCatalog()
	}
	return m.catalog
}

func (m *RateLimitMiddleware) policy(name string, keyKind rateLimitKeyKind, jsonField string, onFailure rateLimitFailureMode) rateLimitPolicy {
	return m.policyCatalog().policy(name, keyKind, jsonField, onFailure)
}

func (m *RateLimitMiddleware) rateLimitKey(policyName, identity string) string {
	var digest []byte
	if len(m.hashSecret) > 0 {
		mac := hmac.New(sha256.New, m.hashSecret)
		_, _ = mac.Write([]byte(identity))
		digest = mac.Sum(nil)
	} else {
		sum := sha256.Sum256([]byte(identity))
		digest = sum[:]
	}
	return rateLimitKeyPrefix + ":" + policyName + ":" + hex.EncodeToString(digest[:16])
}

func publicRateLimitPolicies(method, path string) []rateLimitPolicy {
	return publicRateLimitPoliciesWithCatalog(method, path, defaultRateLimitCatalog())
}

func publicRateLimitPoliciesWithCatalog(method, path string, catalog rateLimitCatalog) []rateLimitPolicy {
	key := method + " " + path
	switch key {
	case http.MethodPost + " /api/v2/auth/login":
		return []rateLimitPolicy{
			ipPolicy(catalog, "auth-login-ip", rateLimitFailClosed),
			jsonPolicy(catalog, "auth-login-email", "email"),
		}
	case http.MethodPost + " /api/v2/auth/refresh":
		return []rateLimitPolicy{ipPolicy(catalog, "auth-refresh-ip", rateLimitFailClosed)}
	case http.MethodPost + " /api/v2/register/session/register":
		return []rateLimitPolicy{
			ipPolicy(catalog, "register-session-ip", rateLimitFailClosed),
			jsonPolicy(catalog, "register-session-email", "email"),
		}
	case http.MethodPatch + " /api/v2/register/session/verify":
		return []rateLimitPolicy{
			ipPolicy(catalog, "register-verify-ip", rateLimitFailClosed),
			jsonPolicy(catalog, "register-verify-session", "session_id"),
		}
	case http.MethodPost + " /api/v2/register/":
		return []rateLimitPolicy{
			ipPolicy(catalog, "register-create-ip", rateLimitFailClosed),
			jsonPolicy(catalog, "register-create-session", "session_id"),
		}
	case http.MethodPost + " /api/v2/register/password/change":
		return []rateLimitPolicy{
			ipPolicy(catalog, "password-change-ip", rateLimitFailClosed),
			jsonPolicy(catalog, "password-change-session", "session_id"),
		}
	case http.MethodGet + " /api/v2/events/popular":
		return []rateLimitPolicy{ipPolicy(catalog, "events-popular-ip", rateLimitFailOpen)}
	case http.MethodGet + " /api/v2/events/search":
		return []rateLimitPolicy{ipPolicy(catalog, "events-search-ip", rateLimitFailOpen)}
	case http.MethodGet + " /api/v2/references",
		http.MethodGet + " /api/v2/references/genres":
		return []rateLimitPolicy{ipPolicy(catalog, "references-read-ip", rateLimitFailOpen)}
	default:
		return nil
	}
}

func authenticatedRateLimitPolicies(method, path string) []rateLimitPolicy {
	return authenticatedRateLimitPoliciesWithCatalog(method, path, defaultRateLimitCatalog())
}

func authenticatedRateLimitPoliciesWithCatalog(method, path string, catalog rateLimitCatalog) []rateLimitPolicy {
	key := method + " " + path
	switch key {
	case http.MethodPost + " /api/v2/sub/UploadImg":
		return []rateLimitPolicy{userPolicy(catalog, "image-upload-user")}
	case http.MethodPost + " /api/v2/groups":
		return []rateLimitPolicy{userPolicy(catalog, "group-create-user")}
	case http.MethodPost + " /api/v2/admin/events":
		return []rateLimitPolicy{userPolicy(catalog, "event-create-user")}
	case http.MethodPost + " /api/v2/groups/:groupId/join",
		http.MethodPost + " /api/v2/groups/:groupId/leave",
		http.MethodPost + " /api/v2/events/:eventId/join",
		http.MethodPost + " /api/v2/events/:eventId/leave":
		return []rateLimitPolicy{userPolicy(catalog, "membership-change-user")}
	case http.MethodPost + " /api/v2/groups/invites":
		return []rateLimitPolicy{userPolicy(catalog, "group-invite-user")}
	case http.MethodPost + " /api/v2/groups/invites/:inviteId/accept",
		http.MethodPost + " /api/v2/groups/invites/:inviteId/reject":
		return []rateLimitPolicy{userPolicy(catalog, "group-invite-response-user")}
	case http.MethodPost + " /api/v2/groups/:groupId/requests/approve-all",
		http.MethodPost + " /api/v2/groups/:groupId/requests/reject-all":
		return []rateLimitPolicy{userPolicy(catalog, "join-request-bulk-user")}
	case http.MethodPost + " /api/v2/groups/requests/:requestId/approve",
		http.MethodPost + " /api/v2/groups/requests/:requestId/reject":
		return []rateLimitPolicy{userPolicy(catalog, "join-request-review-user")}
	case http.MethodPost + " /api/v2/groups/permissions/add",
		http.MethodPost + " /api/v2/groups/permissions/remove":
		return []rateLimitPolicy{userPolicy(catalog, "group-permission-user")}
	case http.MethodPut + " /api/v2/groups":
		return []rateLimitPolicy{userPolicy(catalog, "group-update-user")}
	case http.MethodPut + " /api/v2/admin/events/:eventId":
		return []rateLimitPolicy{userPolicy(catalog, "event-update-user")}
	case http.MethodDelete + " /api/v2/groups/:groupId",
		http.MethodDelete + " /api/v2/admin/events/:eventId":
		return []rateLimitPolicy{userPolicy(catalog, "destructive-admin-user")}
	case http.MethodDelete + " /api/v2/groups/:groupId/members/:userId",
		http.MethodDelete + " /api/v2/groups/:groupId/blacklist/:userId",
		http.MethodDelete + " /api/v2/admin/events/:eventId/kick/:userId":
		return []rateLimitPolicy{userPolicy(catalog, "member-moderation-user")}
	default:
		return nil
	}
}

func ipPolicy(catalog rateLimitCatalog, name string, failureMode rateLimitFailureMode) rateLimitPolicy {
	return catalog.policy(name, rateLimitByIP, "", failureMode)
}

func jsonPolicy(catalog rateLimitCatalog, name, field string) rateLimitPolicy {
	return catalog.policy(name, rateLimitByJSONField, field, rateLimitFailClosed)
}

func userPolicy(catalog rateLimitCatalog, name string) rateLimitPolicy {
	return catalog.policy(name, rateLimitByUser, "", rateLimitFailClosed)
}

func requestJSONField(c *gin.Context, field string) (string, error) {
	if c == nil || c.Request == nil || c.Request.Body == nil || field == "" {
		return "", nil
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxRateLimitBodySize+1))
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("%w: %v", errRateLimitBodyRead, err)
	}
	if len(body) > maxRateLimitBodySize {
		return "", errRateLimitBodyTooLarge
	}

	var values map[string]json.RawMessage
	if err := json.Unmarshal(body, &values); err != nil {
		return "", nil
	}

	raw, ok := values[field]
	if !ok {
		return "", nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", nil
	}
	return strings.TrimSpace(value), nil
}

func abortRateLimitIdentityError(c *gin.Context, err error) bool {
	if errors.Is(err, errRateLimitBodyTooLarge) {
		utils.AbortJSONError(
			c,
			http.StatusRequestEntityTooLarge,
			"request_too_large",
			utils.WithMessage("Тело запроса слишком большое"),
		)
		return false
	}

	utils.AbortJSONError(
		c,
		http.StatusBadRequest,
		"invalid_request",
		utils.WithMessage("Не удалось прочитать тело запроса"),
	)
	return false
}

func setRateLimitHeaders(c *gin.Context, policy rateLimitPolicy, decision RateLimitDecision) {
	remaining := policy.limit - decision.Count
	if remaining < 0 {
		remaining = 0
	}
	reset := resetSeconds(decision.ResetAfter, policy.window)
	resetAt := time.Now().Add(time.Duration(reset) * time.Second).Unix()

	c.Header("RateLimit-Limit", strconv.Itoa(policy.limit))
	c.Header("RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("RateLimit-Reset", strconv.FormatInt(reset, 10))
	c.Header("X-RateLimit-Limit", strconv.Itoa(policy.limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))
}

func resetSeconds(resetAfter, fallback time.Duration) int64 {
	if resetAfter <= 0 {
		resetAfter = fallback
	}
	seconds := int64((resetAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}

func normalizedFullPath(c *gin.Context) string {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func isV2APIPath(path string) bool {
	return path == "/api/v2" || strings.HasPrefix(path, "/api/v2/")
}

func isMutationMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func remoteAddress(request *http.Request) string {
	if request == nil {
		return "unknown"
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if ip := net.ParseIP(request.RemoteAddr); ip != nil {
		return ip.String()
	}
	if request.RemoteAddr == "" {
		return "unknown"
	}
	return request.RemoteAddr
}

func clientAddress(c *gin.Context) string {
	if c != nil {
		if ip := strings.TrimSpace(c.ClientIP()); ip != "" {
			return ip
		}
		if c.Request != nil {
			return remoteAddress(c.Request)
		}
	}
	return "unknown"
}
