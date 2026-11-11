package diagnostic

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudflare/cloudflared/apsy"
)

func TestTracerouteTargetsAreThePlatformsEdge(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	defer func(f func(string, string, string) (string, []*net.SRV, error)) { diagLookupSRV = f }(diagLookupSRV)
	var asked []string
	diagLookupSRV = func(service, proto, name string) (string, []*net.SRV, error) {
		asked = append(asked, "_"+service+"._"+proto+"."+name)
		return "", []*net.SRV{{Target: "edge-a.platform.example.", Port: 7844}, {Target: "edge-b.platform.example.", Port: 7844}}, nil
	}
	assert.Equal(t, []string{"edge-a.platform.example", "edge-b.platform.example"}, edgeHosts())
	assert.Equal(t, []string{"_v2-origintunneld._tcp.platform.example"}, asked)

	diagLookupSRV = func(string, string, string) (string, []*net.SRV, error) {
		return "", nil, errors.New("no such host")
	}
	assert.Empty(t, edgeHosts(), "an unresolved record traces nothing rather than a built-in host")
}

func TestNoSuffixTracesNothing(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	defer func(f func(string, string, string) (string, []*net.SRV, error)) { diagLookupSRV = f }(diagLookupSRV)
	diagLookupSRV = func(_, _, name string) (string, []*net.SRV, error) {
		t.Fatalf("looked up %q with no platform suffix", name)
		return "", nil, nil
	}
	assert.Empty(t, edgeHosts())
}
