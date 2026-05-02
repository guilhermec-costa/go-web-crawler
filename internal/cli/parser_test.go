package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnErrorOnEmptyUrl(t *testing.T) {
	flags, err := ParseCrawlerFlags([]string{})

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "-url is required")
	assert.Equal(t, "", flags.RootUrl)
}

func TestShouldReturnDefaultArgs(t *testing.T) {
	url := "http://example.com"
	flags, err := ParseCrawlerFlags([]string{
		"-url=" + url,
	})

	da := defaultArgs()
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
