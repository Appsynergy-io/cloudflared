package middleware

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
	"github.com/cloudflare/cloudflared/credentials"
)

// Upstream's variable, which its test still names. It is what the Access domain TestMain sets
// gives, so that test runs as upstream wrote it.
var cloudflareAccessCertsURL = "https://%s.cloudflareaccess.com"

// Every test in this package runs with upstream's Access domain unless it sets another.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.AccessDomainEnv, "cloudflareaccess.com")
	os.Exit(m.Run())
}

func TestTheTeamDomainIsUnderTheAccessDomain(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "platformaccess.example")
	for environment, want := range map[string]string{
		"":                      "https://myteam.platformaccess.example",
		credentials.FedEndpoint: "https://myteam.fed.platformaccess.example",
	} {
		got, err := teamDomainURL("myteam", environment)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		validator, err := NewJWTValidator("myteam", environment, []string{"aud"})
		require.NoError(t, err)
		require.NotNil(t, validator)
	}
}

func TestATeamNameCannotAliasAPlatformHost(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "platformaccess.example")
	t.Setenv(apsy.AccessLoginDomainEnv, "platformaccess.test")
	for _, team := range []string{"api", "login", "dash", "management", "login.fed"} {
		for _, environment := range []string{"", credentials.FedEndpoint} {
			got, err := teamDomainURL(team, environment)
			require.NoError(t, err)
			assert.True(t, strings.HasSuffix(got, ".platformaccess.example"), got)
			for _, platformHost := range []string{
				"https://" + team + ".platform.example",
				"https://" + team + ".fed.platform.example",
				"https://" + team + ".platformaccess.test",
			} {
				assert.NotEqual(t, platformHost, got)
			}
		}
	}
	// A name that would leave the Access domain names no team domain.
	for _, team := range []string{"api.platform.example/", "api.platform.example#", "api.platform.example?", "user@api.platform.example", ""} {
		_, err := NewJWTValidator(team, "", []string{"aud"})
		require.Error(t, err, team)
	}
}

func TestNoAccessDomainNamesNoTeamDomain(t *testing.T) {
	// The platform suffix is no substitute: the Access domain is never derived from it.
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "")
	_, err := NewJWTValidator("myteam", "", []string{"aud"})
	require.ErrorIs(t, err, apsy.ErrNoAccessDomain)
	assert.Contains(t, err.Error(), apsy.AccessDomainEnv)
}
