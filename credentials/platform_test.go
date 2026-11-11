package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

func TestTheAPIAndManagementHostsAreThePlatforms(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	apiURL, err := BaseApiURL()
	require.NoError(t, err)
	assert.Equal(t, "https://api.platform.example/client/v4", apiURL)
	hostname, err := ManagementHostname()
	require.NoError(t, err)
	assert.Equal(t, "management.platform.example", hostname)
}

func TestTheFedRAMPHostsAreThePlatforms(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	apiURL, err := FedRampBaseApiURL()
	require.NoError(t, err)
	assert.Equal(t, "https://api.fed.platform.example/client/v4", apiURL)
	hostname, err := FedRampHostname()
	require.NoError(t, err)
	assert.Equal(t, "management.fed.platform.example", hostname)
}

func TestNoSuffixNamesNoHost(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	_, err := BaseApiURL()
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
	_, err = ManagementHostname()
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
	_, err = FedRampBaseApiURL()
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
	_, err = FedRampHostname()
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
}
