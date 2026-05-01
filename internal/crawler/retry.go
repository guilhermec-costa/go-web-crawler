package crawler

import (
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

func IsRetryableReq(resp *http.Response) bool {
	switch resp.StatusCode {
	case
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusRequestTimeout,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

type RetryConfig struct {
	maxRetries    int
	baseDelay     time.Duration
	multiplier    float64
	jitterMaxRand int32
}

func WithRetry(retryConfig RetryConfig, fn func() (*http.Response, error)) (*http.Response, error) {
	for r := 0; r < retryConfig.maxRetries; r++ {
		resp, err := fn()

		if err != nil {
			return nil, err
		}

		if is2xx(resp) {
			return resp, nil
		}

		if !IsRetryableReq(resp) {
			return resp, fmt.Errorf("not retryable: %d", resp.StatusCode)
		}

		delay := retryConfig.baseDelay * time.Duration(math.Pow(retryConfig.multiplier, float64(r)))

		jitter := rand.Int32N(retryConfig.jitterMaxRand)
		jitterTime := time.Duration(jitter * int32(time.Millisecond))

		<-time.After(delay + jitterTime)
	}

	return nil, fmt.Errorf("failed after %d retries", retryConfig.maxRetries)
}
