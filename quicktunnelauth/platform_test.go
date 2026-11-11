package quicktunnelauth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

// Upstream's constants, which its tests still name. They are what the Quick Tunnel domain TestMain
// sets gives, so those tests run as upstream wrote them.
const (
	quickTunnelAuthBrokerIssuer  = "https://login.trycloudflare.com"
	quickTunnelAuthBrokerJWKSURL = "https://login.trycloudflare.com/.well-known/jwks.json"
)

// Every test in this package runs with upstream's Quick Tunnel domain unless it sets another:
// upstream's parallel tests cannot set it themselves.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.QuickTunnelDomainEnv, "trycloudflare.com")
	os.Exit(m.Run())
}

func TestTheBrokerIsUnderTheQuickTunnelDomain(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.QuickTunnelDomainEnv, "tryplatform.example")
	authorize, err := QuickTunnelAuthBrokerAuthorizeURL()
	require.NoError(t, err)
	assert.Equal(t, "https://login.tryplatform.example/authorize", authorize)
	issuer, err := brokerIssuer()
	require.NoError(t, err)
	assert.Equal(t, "https://login.tryplatform.example", issuer)
	jwks, err := brokerJWKSURL()
	require.NoError(t, err)
	assert.Equal(t, "https://login.tryplatform.example/.well-known/jwks.json", jwks)
}

func TestAQuickTunnelHostnameIsOneLabelUnderTheQuickTunnelDomain(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.QuickTunnelDomainEnv, "tryplatform.example")
	assert.True(t, isQuickTunnelHostname("test-tunnel.tryplatform.example"))
	for _, hostname := range []string{
		"test-tunnel.trycloudflare.com",
		"test-tunnel.platform.example",
		"test-tunnel.try.platform.example",
		"api.platform.example",
		"a.test-tunnel.tryplatform.example",
		"tryplatform.example",
		"test-tunnel.eviltryplatform.example",
	} {
		assert.False(t, isQuickTunnelHostname(hostname), hostname)
		_, err := NewQuickTunnelAuthStateManager(hostname)
		require.Error(t, err, hostname)
	}
}

func TestNoQuickTunnelDomainNamesNoQuickTunnel(t *testing.T) {
	// The platform suffix is no substitute: the domain is never derived from it.
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.QuickTunnelDomainEnv, "")
	assert.False(t, isQuickTunnelHostname("test-tunnel.try.platform.example"))
	assert.False(t, isQuickTunnelHostname("test-tunnel.platform.example"))
	_, err := NewQuickTunnelAuthStateManager("test-tunnel.try.platform.example")
	require.ErrorIs(t, err, apsy.ErrNoQuickTunnelDomain)
	assert.Contains(t, err.Error(), apsy.QuickTunnelDomainEnv)
	_, err = NewQuickTunnelAuthAssertionValidator()
	require.ErrorIs(t, err, apsy.ErrNoQuickTunnelDomain)
}
