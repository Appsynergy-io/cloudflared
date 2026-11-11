package ingress

import (
	"os"
	"testing"

	"github.com/cloudflare/cloudflared/apsy"
)

// Every test in this package runs with upstream's Access domain unless it sets another: an Access
// rule's team domain is read under it, and upstream's parallel tests cannot set it themselves.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.AccessDomainEnv, "cloudflareaccess.com")
	os.Exit(m.Run())
}
