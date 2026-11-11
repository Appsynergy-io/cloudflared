package tunnel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"

	"github.com/cloudflare/cloudflared/apsy"
	cfdflags "github.com/cloudflare/cloudflared/cmd/cloudflared/flags"
)

func stringDefault(t *testing.T, flags []cli.Flag, name string) string {
	t.Helper()
	for _, f := range flags {
		if f.Names()[0] != name {
			continue
		}
		switch f := f.(type) {
		case *altsrc.StringFlag:
			return f.StringFlag.Value
		case *cli.StringFlag:
			return f.Value
		}
		t.Fatalf("flag %s is a %T", name, f)
	}
	t.Fatalf("no flag %s", name)
	return ""
}

func TestTheAPIAndManagementDefaultsAreThePlatforms(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	flags := tunnelFlags(false)
	assert.Equal(t, "https://api.platform.example/client/v4", stringDefault(t, flags, cfdflags.ApiURL))
	assert.Equal(t, "management.platform.example", stringDefault(t, flags, cfdflags.ManagementHostname))
}

func TestNoSuffixDefaultsNoAPIOrManagementHost(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	flags := tunnelFlags(false)
	assert.Equal(t, "", stringDefault(t, flags, cfdflags.ApiURL))
	assert.Equal(t, "", stringDefault(t, flags, cfdflags.ManagementHostname))
}
