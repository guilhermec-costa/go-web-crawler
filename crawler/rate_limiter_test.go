package crawler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testCtx = context.Background()

func TestShouldStartRefilled(t *testing.T) {
	const bufMax = 5
	rl := NewRateLimiter(testCtx, bufMax, 2*time.Second)
	assert.Equal(t, rl.TokenCount(), bufMax)
}

func TestShouldConsumeTokenOnWait(t *testing.T) {
	const bufMax = 5
	rl := NewRateLimiter(testCtx, bufMax, 2*time.Second)
	assert.Equal(t, 5, rl.TokenCount())

	rl.Wait()
	assert.Equal(t, 4, rl.TokenCount())
}

func TestShouldNotExceedTokenRate(t *testing.T) {
	const bufMax = 10
	rl := NewRateLimiter(testCtx, bufMax, 2*time.Second)
	rl.refill()

	assert.Equal(t, bufMax, rl.TokenCount())
}

func TestShouldRefillAfterConsumption(t *testing.T) {
	const bufMax = 20
	rl := NewRateLimiter(testCtx, bufMax, 2*time.Second)

	consumptions := 3
	for range consumptions {
		rl.Wait()
	}

	assert.Equal(t, bufMax-consumptions, rl.TokenCount())

	rl.refill()

	assert.Equal(t, bufMax, rl.TokenCount())
}

func TestShouldStopRefillOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(testCtx)
	const bufMax = 5
	rl := NewRateLimiter(ctx, bufMax, 50*time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
	rl.Wait()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, bufMax-1, rl.TokenCount())
}
