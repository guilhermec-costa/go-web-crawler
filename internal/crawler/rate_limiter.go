package crawler

import (
	"context"
	"time"
)

type RateLimiter struct {
	tokenRate   int
	tokenBucket chan struct{}
	refillTime  time.Duration
}

func NewRateLimiter(ctx context.Context, tokenRate int, refillTime time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokenRate:   tokenRate,
		tokenBucket: make(chan struct{}, tokenRate),
		refillTime:  refillTime,
	}

	rl.refill()
	go rl.StartAutorefill(ctx)
	return rl
}

func (rl *RateLimiter) refill() {
	for range rl.tokenRate {
		select {
		// if current token can not be refilled, continues to the next one
		case rl.tokenBucket <- struct{}{}:
		default:
		}
	}
}

func (rl *RateLimiter) Wait() {
	<-rl.tokenBucket
}

func (rl *RateLimiter) StartAutorefill(ctx context.Context) {
	ticker := time.NewTicker(rl.refillTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			rl.refill()
		}
	}
}

func (rl *RateLimiter) TokenCount() int {
	return len(rl.tokenBucket)
}
