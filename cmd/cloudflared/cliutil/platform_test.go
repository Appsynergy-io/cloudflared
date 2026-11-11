package cliutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

func TestPlatformDefault(t *testing.T) {
	assert.Equal(t, "management.platform.example", PlatformDefault(func() (string, error) {
		return "management.platform.example", nil
	}))
	assert.Equal(t, "", PlatformDefault(func() (string, error) { return "", apsy.ErrNoSuffix }))
}

func TestAnEmptyPlatformValueNamesTheFlagAndTheSuffix(t *testing.T) {
	v, err := RequirePlatformValue("api-url", "https://api.platform.example/client/v4")
	require.NoError(t, err)
	assert.Equal(t, "https://api.platform.example/client/v4", v)

	_, err = RequirePlatformValue("api-url", "")
	require.True(t, errors.Is(err, apsy.ErrNoSuffix))
	assert.Contains(t, err.Error(), "--api-url")
	assert.Contains(t, err.Error(), apsy.SuffixEnv)
}
