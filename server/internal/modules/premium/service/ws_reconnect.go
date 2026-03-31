package service

import (
	"context"
	"math/rand/v2"
	"time"
)

// Backoff 指数退避 + 抖动
type Backoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	attempt    int
}

func NewBackoff() *Backoff {
	return &Backoff{
		Initial:    time.Second,
		Max:        30 * time.Second,
		Multiplier: 2,
	}
}

func (b *Backoff) Next() time.Duration {
	if b.attempt == 0 {
		b.attempt++
		return b.Initial
	}

	d := b.Initial
	for i := 0; i < b.attempt; i++ {
		d = time.Duration(float64(d) * b.Multiplier)
		if d > b.Max {
			d = b.Max
			break
		}
	}
	b.attempt++

	// 加 ±25% 抖动
	jitter := float64(d) * (0.75 + rand.Float64()*0.5)
	return time.Duration(jitter)
}

func (b *Backoff) Reset() {
	b.attempt = 0
}

// SleepWithContext 可被 context 取消的 sleep
func SleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
