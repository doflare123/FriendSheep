package middlewares

import (
	"time"

	"friendship/config"
)

type rateLimitQuota struct {
	limit  int
	window time.Duration
}

type rateLimitCatalog map[string]rateLimitQuota

func defaultRateLimitCatalog() rateLimitCatalog {
	return newRateLimitCatalog(config.DefaultRateLimitConfig())
}

func newRateLimitCatalog(cfg config.RateLimitConfig) rateLimitCatalog {
	quotas := cfg.PolicyCatalog()
	catalog := make(rateLimitCatalog, len(quotas))
	for name, quota := range quotas {
		catalog[name] = rateLimitQuota{
			limit:  quota.Limit,
			window: quota.Window,
		}
	}
	return catalog
}

func (c rateLimitCatalog) policy(name string, keyKind rateLimitKeyKind, jsonField string, onFailure rateLimitFailureMode) rateLimitPolicy {
	quota, ok := c[name]
	if !ok {
		quota = defaultRateLimitCatalog()[name]
	}

	return rateLimitPolicy{
		name:      name,
		limit:     quota.limit,
		window:    quota.window,
		keyKind:   keyKind,
		jsonField: jsonField,
		onFailure: onFailure,
	}
}
