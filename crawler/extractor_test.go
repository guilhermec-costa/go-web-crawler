package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testCfg = RetryConfig{
	maxRetries:     1,
	baseDelay:      time.Duration(1),
	jitterMaxRand:  0,
	multiplier:     1,
	requestTimeout: 100 * time.Millisecond,
}


func TestExtractPageNodes_SuccessfulExtraction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html><body><a href="/link1">link</a><h1>Title</h1></body></html>`))
	}))
	defer server.Close()

	ctx := context.Background()
	rl := NewRateLimiter(ctx, 10, time.Second)
	pageUrl, _ := url.Parse(server.URL)

	nodes, err := extractPageNodes(ctx, pageUrl, rl, testCfg)

	assert.NoError(t, err)
	assert.Equal(t, 1, nodes[NodeExtractorTypeA].Count)
	assert.Equal(t, 1, nodes[NodeExtractorTypeH1].Count)
}

func TestExtractPageNodes_NonRetryableStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	ctx := context.Background()
	rl := NewRateLimiter(ctx, 10, time.Second)
	pageUrl, _ := url.Parse(server.URL)

	nodes, err := extractPageNodes(ctx, pageUrl, rl, testCfg)
	assert.Nil(t, nodes)
	assert.ErrorContains(t, err, "not retryable")
}