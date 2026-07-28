package middlewares

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"friendship/repository"

	"github.com/redis/go-redis/v9"
)

const takeRateLimitScript = `
local current = redis.call("INCR", KEYS[1])
local window = tonumber(ARGV[2])
local ttl

if current == 1 then
	redis.call("PEXPIRE", KEYS[1], window)
	ttl = window
else
	ttl = redis.call("PTTL", KEYS[1])
	if ttl < 0 then
		redis.call("PEXPIRE", KEYS[1], window)
		ttl = window
	end
end

local allowed = 0
if current <= tonumber(ARGV[1]) then
	allowed = 1
end

return {allowed, current, ttl}
`

type redisRateLimitEvaluator interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

type redisRateLimitStore struct {
	client redisRateLimitEvaluator
}

func NewRedisRateLimitStore(repo repository.RedisRepository) RateLimitStore {
	if repo == nil {
		return &redisRateLimitStore{}
	}
	return &redisRateLimitStore{client: repo.Client()}
}

func (s *redisRateLimitStore) Take(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (RateLimitDecision, error) {
	if s == nil || s.client == nil {
		return RateLimitDecision{}, errors.New("redis rate limit client is unavailable")
	}
	if key == "" || limit <= 0 || window <= 0 {
		return RateLimitDecision{}, errors.New("invalid rate limit request")
	}

	windowMilliseconds := window.Milliseconds()
	if windowMilliseconds < 1 {
		windowMilliseconds = 1
	}

	result, err := s.client.Eval(
		ctx,
		takeRateLimitScript,
		[]string{key},
		limit,
		windowMilliseconds,
	).Result()
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("take redis rate limit: %w", err)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return RateLimitDecision{}, fmt.Errorf("unexpected redis rate limit result %T", result)
	}

	allowedValue, err := redisInteger(values[0])
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("decode redis rate limit allowed flag: %w", err)
	}
	count, err := redisInteger(values[1])
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("decode redis rate limit count: %w", err)
	}
	ttlMilliseconds, err := redisInteger(values[2])
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("decode redis rate limit ttl: %w", err)
	}
	if ttlMilliseconds < 1 {
		ttlMilliseconds = windowMilliseconds
	}

	return RateLimitDecision{
		Allowed:    allowedValue == 1,
		Count:      int(count),
		ResetAfter: time.Duration(ttlMilliseconds) * time.Millisecond,
	}, nil
}

func redisInteger(value interface{}) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int:
		return int64(typed), nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	case []byte:
		return strconv.ParseInt(string(typed), 10, 64)
	default:
		return 0, fmt.Errorf("unsupported integer type %T", value)
	}
}
