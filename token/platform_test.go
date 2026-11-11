package token

import (
	"net/url"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

// Upstream's constants, which its tests still name. They are what the Access login domain TestMain
// sets gives, so those tests run as upstream wrote them.
const (
	baseStoreURL = "https://login.cloudflareaccess.org/"
	fedStoreURL  = "https://login.fed.cloudflareaccess.org/"
)

// Every test in this package runs with upstream's Access domains unless it sets others: upstream's
// parallel tests cannot set them themselves.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.AccessDomainEnv, "cloudflareaccess.com")
	_ = os.Setenv(apsy.AccessLoginDomainEnv, "cloudflareaccess.org")
	os.Exit(m.Run())
}

func withPlatformDomains(t *testing.T) {
	t.Helper()
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "platformaccess.example")
	t.Setenv(apsy.AccessLoginDomainEnv, "platformaccess.test")
	t.Setenv(apsy.QuickTunnelDomainEnv, "tryplatform.example")
}

func TestAnAuthDomainIsUnderTheAccessDomainAlone(t *testing.T) {
	withPlatformDomains(t)
	for _, domain := range []string{"team.platformaccess.example", "team.fed.platformaccess.example", "API.PlatformAccess.Example"} {
		_, err := parseAuthDomain(domain)
		require.NoError(t, err, domain)
	}
	for domain, why := range map[string]string{
		"team.platform.example":           "a host under the platform suffix",
		"api.platform.example":            "a platform host",
		"team.fed.platform.example":       "a host under the platform suffix",
		"platform.example":                "the platform suffix",
		"label.tryplatform.example":       "a Quick Tunnel host",
		"login.platformaccess.test":       "the Access login host",
		"platformaccess.example":          "the Access domain itself is no team domain",
		"team.evilplatformaccess.example": "not under the Access domain",
		"team.cloudflareaccess.com":       "Cloudflare's Access domain",
	} {
		_, err := parseAuthDomain(domain)
		require.Error(t, err, "%s: %s", domain, why)
		assert.False(t, isFedAuthDomain(domain), domain)
	}
	assert.True(t, isFedAuthDomain("team.fed.platformaccess.example"))
	assert.False(t, isFedAuthDomain("team.fed.platform.example"))
}

func TestNoAccessDomainAcceptsNoAuthDomain(t *testing.T) {
	// The platform suffix is no substitute: the Access domain is never derived from it.
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "")
	for _, domain := range []string{"team.platform.example", "team.cloudflareaccess.com"} {
		_, err := parseAuthDomain(domain)
		require.ErrorIs(t, err, apsy.ErrNoAccessDomain)
		assert.Contains(t, err.Error(), apsy.AccessDomainEnv)
	}
}

func TestTheTransferServiceIsUnderTheAccessLoginDomain(t *testing.T) {
	withPlatformDomains(t)
	for fedramp, want := range map[bool]string{
		false: "https://login.platformaccess.test/",
		true:  "https://login.fed.platformaccess.test/",
	} {
		got, err := transferStoreURL(fedramp)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

func TestNoAccessLoginDomainOpensNoBrowser(t *testing.T) {
	withPlatformDomains(t)
	t.Setenv(apsy.AccessLoginDomainEnv, "")
	opened := false
	oldOpen := openBrowser
	openBrowser = func(string) error { opened = true; return nil }
	t.Cleanup(func() { openBrowser = oldOpen })
	log := zerolog.Nop()
	u, err := url.Parse("https://app.example.com")
	require.NoError(t, err)
	for _, fedramp := range []bool{false, true} {
		_, err = RunTransfer(u, "aud", "token", "token", "", true, false, true, fedramp, &log, "")
		require.ErrorIs(t, err, apsy.ErrNoAccessLoginDomain)
		assert.Contains(t, err.Error(), apsy.AccessLoginDomainEnv)
	}
	assert.False(t, opened)
}
