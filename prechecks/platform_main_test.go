package prechecks

import (
	"os"
	"testing"

	"github.com/cloudflare/cloudflared/apsy"
)

// Every test in this package runs with a platform suffix unless it clears it: the probes read it,
// and upstream's parallel tests cannot set it themselves.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.SuffixEnv, "platform.example")
	// The region labels carry the suffix, so they are taken once it is set.
	testRegion1Global, testRegion2Global = regionTargets("")
	testRegion1US, testRegion2US = regionTargets("us")
	testRegion1Fed, testRegion2Fed = regionTargets("fed")
	testRegion1EU, testRegion2EU = regionTargets("eu")
	os.Exit(m.Run())
}
