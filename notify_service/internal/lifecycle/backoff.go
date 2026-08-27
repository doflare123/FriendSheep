package lifecycle

import "time"

type ExponentialBackoff struct {
	Min    time.Duration
	Max    time.Duration
	Jitter Jitter
}

func (b ExponentialBackoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	base := b.Min
	if base <= 0 {
		base = time.Second
	}
	max := b.Max
	if max < base {
		max = base
	}
	for i := 1; i < attempt && base < max; i++ {
		if base > max/2 {
			base = max
			break
		}
		base *= 2
	}
	if base > max {
		base = max
	}
	if b.Jitter != nil {
		return b.Jitter.Apply(base)
	}
	return base
}
