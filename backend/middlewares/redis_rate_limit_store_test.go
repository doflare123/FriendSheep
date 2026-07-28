package middlewares

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisEvalCall struct {
	script string
	keys   []string
	args   []interface{}
}

type fakeRedisRateLimitEvaluator struct {
	mu     sync.Mutex
	calls  []redisEvalCall
	result interface{}
	err    error
}

func (f *fakeRedisRateLimitEvaluator) Eval(
	_ context.Context,
	script string,
	keys []string,
	args ...interface{},
) *redis.Cmd {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, redisEvalCall{
		script: script,
		keys:   append([]string(nil), keys...),
		args:   append([]interface{}(nil), args...),
	})
	return redis.NewCmdResult(f.result, f.err)
}

func (f *fakeRedisRateLimitEvaluator) snapshotCalls() []redisEvalCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]redisEvalCall(nil), f.calls...)
}

func TestRedisRateLimitStoreTakeUsesOneAtomicEvalAndDecodesDecision(t *testing.T) {
	evaluator := &fakeRedisRateLimitEvaluator{
		result: []interface{}{int64(1), int64(3), int64(2750)},
	}
	store := &redisRateLimitStore{client: evaluator}

	decision, err := store.Take(context.Background(), "rate-limit-key", 7, 3*time.Second)
	if err != nil {
		t.Fatalf("Take returned error: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("Allowed = false, want true")
	}
	if decision.Count != 3 {
		t.Fatalf("Count = %d, want 3", decision.Count)
	}
	if decision.ResetAfter != 2750*time.Millisecond {
		t.Fatalf("ResetAfter = %v, want 2.75s", decision.ResetAfter)
	}

	calls := evaluator.snapshotCalls()
	if len(calls) != 1 {
		t.Fatalf("Eval calls = %d, want one atomic call: %#v", len(calls), calls)
	}
	call := calls[0]
	if len(call.keys) != 1 || call.keys[0] != "rate-limit-key" {
		t.Fatalf("Eval keys = %#v, want [rate-limit-key]", call.keys)
	}
	if len(call.args) != 2 || call.args[0] != 7 || call.args[1] != int64(3000) {
		t.Fatalf("Eval args = %#v, want [7 3000]", call.args)
	}
	if !strings.Contains(call.script, `redis.call("INCR", KEYS[1])`) {
		t.Fatal("Lua script does not increment the counter atomically")
	}
	if !strings.Contains(call.script, `redis.call("PEXPIRE", KEYS[1], window)`) {
		t.Fatal("Lua script does not set the counter expiration in the same Eval")
	}
}

func TestRedisRateLimitStoreTakeDecodesDeniedAndByteValues(t *testing.T) {
	evaluator := &fakeRedisRateLimitEvaluator{
		result: []interface{}{[]byte("0"), "8", int64(900)},
	}
	store := &redisRateLimitStore{client: evaluator}

	decision, err := store.Take(context.Background(), "rate-limit-key", 7, time.Second)
	if err != nil {
		t.Fatalf("Take returned error: %v", err)
	}
	if decision.Allowed {
		t.Fatal("Allowed = true, want false")
	}
	if decision.Count != 8 {
		t.Fatalf("Count = %d, want 8", decision.Count)
	}
	if decision.ResetAfter != 900*time.Millisecond {
		t.Fatalf("ResetAfter = %v, want 900ms", decision.ResetAfter)
	}
}

func TestRedisRateLimitStoreTakeValidatesInputAndPropagatesFailures(t *testing.T) {
	storeWithoutClient := &redisRateLimitStore{}
	if _, err := storeWithoutClient.Take(context.Background(), "key", 1, time.Second); err == nil {
		t.Fatal("Take without Redis client returned nil error")
	}

	evaluator := &fakeRedisRateLimitEvaluator{}
	store := &redisRateLimitStore{client: evaluator}
	for _, tt := range []struct {
		name   string
		key    string
		limit  int
		window time.Duration
	}{
		{name: "empty key", limit: 1, window: time.Second},
		{name: "zero limit", key: "key", window: time.Second},
		{name: "zero window", key: "key", limit: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := store.Take(context.Background(), tt.key, tt.limit, tt.window); err == nil {
				t.Fatal("Take returned nil error")
			}
		})
	}
	if calls := evaluator.snapshotCalls(); len(calls) != 0 {
		t.Fatalf("invalid input reached Redis Eval: %#v", calls)
	}

	evalErr := errors.New("redis unavailable")
	evaluator.err = evalErr
	if _, err := store.Take(context.Background(), "key", 1, time.Second); !errors.Is(err, evalErr) {
		t.Fatalf("Take error = %v, want wrapped %v", err, evalErr)
	}

	evaluator.err = nil
	evaluator.result = "unexpected"
	if _, err := store.Take(context.Background(), "key", 1, time.Second); err == nil {
		t.Fatal("Take with malformed result returned nil error")
	}
}

func TestRedisRateLimitStoreTakeRestoresWindowWhenTTLIsMissing(t *testing.T) {
	evaluator := &fakeRedisRateLimitEvaluator{
		result: []interface{}{int64(1), int64(1), int64(-1)},
	}
	store := &redisRateLimitStore{client: evaluator}

	decision, err := store.Take(context.Background(), "rate-limit-key", 5, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("Take returned error: %v", err)
	}
	if decision.ResetAfter != 1500*time.Millisecond {
		t.Fatalf("ResetAfter = %v, want window fallback 1.5s", decision.ResetAfter)
	}
}
