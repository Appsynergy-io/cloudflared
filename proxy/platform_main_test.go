package proxy

import (
	"os"
	"testing"

	"github.com/cloudflare/cloudflared/apsy"
)

// Every test in this package runs with upstream's Quick Tunnel domain: protected quick tunnels are
// named under it, and upstream's parallel tests cannot set it themselves.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.QuickTunnelDomainEnv, "trycloudflare.com")
	os.Exit(m.Run())
}
