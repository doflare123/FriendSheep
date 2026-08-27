package lifecycle

import (
	"math/rand"
	"time"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type DelaySource interface {
	After(time.Duration) <-chan time.Time
}

type RealDelaySource struct{}

func (RealDelaySource) After(delay time.Duration) <-chan time.Time {
	return time.After(delay)
}

type Jitter interface {
	Apply(time.Duration) time.Duration
}

type RandomJitter struct {
	rng *rand.Rand
}

func NewRandomJitter(seed int64) RandomJitter {
	return RandomJitter{rng: rand.New(rand.NewSource(seed))}
}

func (j RandomJitter) Apply(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	if j.rng == nil {
		return base
	}
	window := int64(base / 5)
	if window <= 0 {
		return base
	}
	return base + time.Duration(j.rng.Int63n(window))
}
