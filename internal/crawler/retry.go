package crawler

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

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

		statusInfo := getStatusInfo(resp.StatusCode)
		slog.Info("request failed", "status_code", resp.StatusCode, "message", statusInfo.Message, "attempt", r+1)

		if !statusInfo.Retryable {
			return resp, fmt.Errorf("not retryable: %d", resp.StatusCode)
		}

		jitter := rand.Int32N(retryConfig.jitterMaxRand)
		jitterTime := time.Duration(jitter * int32(time.Millisecond))
		delay := retryConfig.baseDelay * time.Duration(math.Pow(retryConfig.multiplier, float64(r)))
		<-time.After(delay + jitterTime)
	}

	return nil, fmt.Errorf("failed after %d retries", retryConfig.maxRetries)
}
