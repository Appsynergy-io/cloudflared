package tail

import (
	"encoding/base64"
	"flag"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cloudflare/cloudflared/apsy"
	"github.com/cloudflare/cloudflared/cfapi"
	cfdflags "github.com/cloudflare/cloudflared/cmd/cloudflared/flags"
)

// A management token of the shape the edge issues, FedRAMP's issuer included; nothing here
// verifies its signature, as upstream does not.
func managementToken(issuer string) string {
	enc := base64.RawURLEncoding.EncodeToString
	header := enc([]byte(`{"alg":"ES256","typ":"JWT"}`))
	claims := enc([]byte(`{"tun":{"id":"7b098149-51fe-4ee5-a687-3e374466efc7","account_tag":"cd391e9c0626a8f76cb1f670f6591b05"},"actor":{"id":"operator","support":false},"res":["logs"],"iss":"` + issuer + `"}`))
	return header + "." + claims + "." + enc(make([]byte, 64))
}

func tailContext(t *testing.T, token, hostname string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("tail", flag.ContinueOnError)
	set.String("token", token, "")
	set.String("connector-id", "", "")
	set.String(cfdflags.ManagementHostname, hostname, "")
	return cli.NewContext(cli.NewApp(), set, nil)
}

func TestAFedTokenTailsThePlatformsFedManagementHost(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	log := zerolog.Nop()
	for issuer, want := range map[string]string{
		"tunnelstore":     "management.platform.example",
		"fed-tunnelstore": "management.fed.platform.example",
	} {
		u, err := buildURL(tailContext(t, managementToken(issuer), "management.platform.example"), &log, cfapi.Logs)
		require.NoError(t, err, issuer)
		assert.Equal(t, want, u.Host, issuer)
	}
}

func TestAFedTokenWithoutASuffixNamesNoHost(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	log := zerolog.Nop()
	_, err := buildURL(tailContext(t, managementToken("fed-tunnelstore"), "management.platform.example"), &log, cfapi.Logs)
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
}

func managementHostnameDefault(t *testing.T) string {
	t.Helper()
	for _, f := range Command().Flags {
		if f.Names()[0] == cfdflags.ManagementHostname {
			return f.(*cli.StringFlag).Value
		}
	}
	t.Fatalf("no flag %s", cfdflags.ManagementHostname)
	return ""
}

func TestTheManagementDefaultIsThePlatforms(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	assert.Equal(t, "management.platform.example", managementHostnameDefault(t))
}

func TestNoSuffixTailsNoManagementHost(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	assert.Equal(t, "", managementHostnameDefault(t))
	log := zerolog.Nop()
	_, err := buildURL(tailContext(t, managementToken("tunnelstore"), ""), &log, cfapi.Logs)
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
}
