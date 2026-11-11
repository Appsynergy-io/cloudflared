package tunnel

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cloudflare/cloudflared/apsy"
)

func withPlatformDomains(t *testing.T) {
	t.Helper()
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "platformaccess.example")
	t.Setenv(apsy.AccessLoginDomainEnv, "platformaccess.test")
	t.Setenv(apsy.QuickTunnelDomainEnv, "tryplatform.example")
}

func TestTheQuickTunnelServiceIsUnderTheQuickTunnelDomain(t *testing.T) {
	withPlatformDomains(t)
	assert.Equal(t, "https://api.tryplatform.example", stringDefault(t, tunnelFlags(false), "quick-service"))
}

func TestNoQuickTunnelDomainDefaultsNoQuickTunnelService(t *testing.T) {
	// The platform suffix is no substitute: the Quick Tunnel domain is never derived from it.
	withPlatformDomains(t)
	t.Setenv(apsy.QuickTunnelDomainEnv, "")
	assert.Equal(t, "", stringDefault(t, tunnelFlags(false), "quick-service"))
	_, err := quickServiceURL()
	require.ErrorIs(t, err, apsy.ErrNoQuickTunnelDomain)
	assert.Contains(t, err.Error(), apsy.QuickTunnelDomainEnv)
}

func TestAQuickTunnelWithoutAQuickTunnelDomainNamesTheSetting(t *testing.T) {
	withPlatformDomains(t)
	t.Setenv(apsy.QuickTunnelDomainEnv, "")
	set := flag.NewFlagSet("tunnel", flag.ContinueOnError)
	for _, f := range tunnelFlags(false) {
		require.NoError(t, f.Apply(set))
	}
	require.NoError(t, set.Parse([]string{"--url", "http://localhost:8080"}))
	err := TunnelCommand(cli.NewContext(cli.NewApp(), set, nil))
	require.ErrorIs(t, err, apsy.ErrNoQuickTunnelDomain)
}

func TestTheLoginURLsAreThePlatforms(t *testing.T) {
	withPlatformDomains(t)
	for _, tc := range []struct {
		url  func() (string, error)
		want string
	}{
		{baseLoginURL, "https://dash.platform.example/argotunnel"},
		{callbackURL, "https://login.platformaccess.test/"},
		{fedBaseLoginURL, "https://dash.fed.platform.example/argotunnel"},
		{fedCallbackStoreURL, "https://login.fed.platformaccess.test/"},
	} {
		got, err := tc.url()
		require.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}
}

func loginError(t *testing.T, fed bool) error {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	set := flag.NewFlagSet("login", flag.ContinueOnError)
	for _, f := range buildLoginSubcommand(false).Flags {
		require.NoError(t, f.Apply(set))
	}
	if fed {
		require.NoError(t, set.Set(fedRAMPParamName, "true"))
	}
	return login(cli.NewContext(cli.NewApp(), set, nil))
}

func TestTunnelLoginWithoutASuffixNamesNoHost(t *testing.T) {
	withPlatformDomains(t)
	t.Setenv(apsy.SuffixEnv, "")
	for _, fed := range []bool{false, true} {
		require.ErrorIs(t, loginError(t, fed), apsy.ErrNoSuffix, "fedramp %v", fed)
	}
}

func TestTunnelLoginWithoutAnAccessLoginDomainNamesTheSetting(t *testing.T) {
	withPlatformDomains(t)
	t.Setenv(apsy.AccessLoginDomainEnv, "")
	for _, fed := range []bool{false, true} {
		err := loginError(t, fed)
		require.ErrorIs(t, err, apsy.ErrNoAccessLoginDomain, "fedramp %v", fed)
		assert.Contains(t, err.Error(), apsy.AccessLoginDomainEnv)
	}
}
