package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testRetryConfig = RetryConfig{
	maxRetries:    3,
	baseDelay:     time.Duration(1),
	jitterMaxRand: 1,
	multiplier:    2,
}

func TestStopsOnClosureError(t *testing.T) {
	attemps := 0
	fn := func() (*http.Response, error) {
		attemps++
		return nil, errors.New("error")
	}

	resp, err := WithRetry(testRetryConfig, fn)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, attemps, 1)
}

func TestReturnsOnSuccess(t *testing.T) {
	attempts := 0
	fn := func() (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: 200,
		}, nil
	}

	resp, err := WithRetry(testRetryConfig, fn)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, attempts, 1)
}

func TestStopsOnNonRetryableStatus(t *testing.T) {
	fn := func() (*http.Response, error) {
		return &http.Response{
			StatusCode: 350,
		}, nil
	}

	resp, err := WithRetry(testRetryConfig, fn)

	assert.NotNil(t, resp)
	assert.ErrorContains(t, err, "not retryable")
}

func TestRetriesUntilSuccess(t *testing.T) {

	cfg := RetryConfig{
		maxRetries:    3,
		baseDelay:     time.Duration(1),
		jitterMaxRand: 1,
		multiplier:    2,
	}

	attempts := 0
	response := &http.Response{}

	fn := func() (*http.Response, error) {
		attempts++
		switch attempts {
		case 1, 2:
			{
				response.StatusCode = 429 // 429 is retryable
			}

		case 3:
			response.StatusCode = 200
		}

		return response, nil
	}

	resp, err := WithRetry(cfg, fn)

	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, attempts, 3)
}

func TestFailsAfterMaxRetries(t *testing.T) {
	cfg := RetryConfig{
		maxRetries:    10,
		baseDelay:     time.Duration(1),
		jitterMaxRand: 1,
		multiplier:    2,
	}

	attempts := 0

	fn := func() (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: 429,
		}, nil
	}

	resp, err := WithRetry(cfg, fn)

	assert.ErrorContains(t, err, fmt.Sprintf("failed after %d retries", cfg.maxRetries))
	assert.Equal(t, attempts, cfg.maxRetries)
	assert.Nil(t, resp)
}
