package crawler

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnUrlHostError(t *testing.T) {
	u, _ := url.Parse("url")
	err := ValidateURL(u)

	assert.EqualError(t, err, "url host should not be empty")
}

func TestShouldReturnSchemeError(t *testing.T) {
	u, _ := url.Parse("scheme://example.com")
	err := ValidateURL(u)
	assert.EqualError(t, err, "url should be http or https only")
}

func TestShouldReturnNoErrorForValidUrl(t *testing.T) {
	_, err := url.Parse("http://example.com")
	assert.NoError(t, err)
}

func TestShouldReturnInvalidStatusCode(t *testing.T) {
	_300 := getStatusInfo(300)
	_429 := getStatusInfo(429)
	_504 := getStatusInfo(504)
	_500 := getStatusInfo(500)

	assert.False(t, _300.Retryable)
	assert.Equal(t, _300.Message, "unexpected status code")

	assert.True(t, _429.Retryable)
	assert.True(t, _504.Retryable)
	assert.True(t, _500.Retryable)
}
