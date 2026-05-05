package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	val "guilhermec-costa/go-web-crawler/crawler/validation"
)

func TestShouldReturnErrorOnEmptyUrl(t *testing.T) {
	flags, err := ParseCliCrawlerFlags([]string{})

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "-url is required")
	assert.Equal(t, "", flags.RootUrl)
}

func TestShouldReturnDefaultArgs(t *testing.T) {
	url := "http://example.com"
	flags, err := ParseCliCrawlerFlags([]string{
		"-url=" + url,
	})

	da := val.DefaultCrawlerParams()
	assert.NoError(t, err)
	assert.Equal(t, flags.RootUrl, url)
	assert.Equal(t, flags.Depth, da.Depth)
	assert.Equal(t, flags.Workers, da.Workers)
	assert.Equal(t, flags.Verbose, da.Verbose)
	assert.Equal(t, flags.TimeoutMs, da.TimeoutMs)
	assert.Equal(t, flags.TickUpdateMs, da.TickUpdateMs)
	assert.Equal(t, flags.OutputPath, da.OutputPath)

	assert.True(t, strings.HasPrefix(da.OutputPath, "extractions-"))
	assert.True(t, strings.HasSuffix(da.OutputPath, ".jsonl"))
}
