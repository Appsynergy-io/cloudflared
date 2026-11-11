package origins

import (
	"context"
	"net/netip"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

// Every test in this package runs with a platform suffix unless it clears it: the resolver
// refresh looks the suffix up, and upstream's tests expect a refresh to succeed.
func TestMain(m *testing.M) {
	_ = os.Setenv(apsy.SuffixEnv, "platform.example")
	os.Exit(m.Run())
}

type recordingPeekResolver struct {
	asked []string
}

func (r *recordingPeekResolver) addr() (string, string) { return "udp", "127.0.0.2:53" }

func (r *recordingPeekResolver) lookupNetIP(_ context.Context, host string) ([]netip.Addr, error) {
	r.asked = append(r.asked, host)
	return nil, nil
}

func TestTheResolverProbeLooksUpThePlatformNotCloudflare(t *testing.T) {
	log := zerolog.Nop()
	service := NewDNSResolverService(NewDNSDialer(), &log, &noopMetrics{})
	peek := &recordingPeekResolver{}
	service.resolver = peek
	require.NoError(t, service.update(t.Context()))
	assert.Equal(t, []string{"platform.example"}, peek.asked)
	validateAddrs(t, []netip.AddrPort{netip.MustParseAddrPort("127.0.0.2:53")}, service.addresses)
}

func TestNoSuffixKeepsTheDefaultResolverAndAsksNothing(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	log := zerolog.Nop()
	service := NewDNSResolverService(NewDNSDialer(), &log, &noopMetrics{})
	peek := &recordingPeekResolver{}
	service.resolver = peek
	require.ErrorIs(t, service.update(t.Context()), apsy.ErrNoSuffix)
	assert.Empty(t, peek.asked)
	assert.Equal(t, defaultResolverAddr, service.getAddress())
}
