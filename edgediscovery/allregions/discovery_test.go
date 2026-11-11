package allregions

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

func (ea *EdgeAddr) String() string {
	return fmt.Sprintf("%s-%s", ea.TCP, ea.UDP)
}

func TestEdgeDiscovery(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	mockAddrs := newMockAddrs(19, 2, 5)
	netLookupSRV = mockNetLookupSRV(mockAddrs)
	netLookupIP = mockNetLookupIP(mockAddrs)

	expectedAddrSet := map[string]bool{}
	for _, addrs := range mockAddrs.addrMap {
		for _, addr := range addrs {
			expectedAddrSet[addr.String()] = true
		}
	}

	l := zerolog.Nop()
	addrLists, err := EdgeDiscovery(&l, "")
	require.NoError(t, err)
	actualAddrSet := map[string]bool{}
	for _, addrs := range addrLists {
		for _, addr := range addrs {
			actualAddrSet[addr.String()] = true
		}
	}

	assert.Equal(t, expectedAddrSet, actualAddrSet)
}

func TestEdgeDiscoveryAsksThePlatformSuffix(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.DoTAddrEnv, "")
	defer func(s func(string, string, string) (string, []*net.SRV, error)) { netLookupSRV = s }(netLookupSRV)
	var asked []string
	netLookupSRV = func(service, proto, name string) (string, []*net.SRV, error) {
		asked = append(asked, "_"+service+"._"+proto+"."+name)
		return "", nil, errors.New("no such host")
	}
	l := zerolog.Nop()
	_, err := EdgeDiscovery(&l, RegionalServiceName(""))
	require.Error(t, err)
	_, err = EdgeDiscovery(&l, RegionalServiceName("us"))
	require.Error(t, err)
	// One lookup each, at the platform's own name, and with no DNS-over-TLS resolver named no
	// second resolver asked behind it.
	assert.Equal(t, []string{
		"_v2-origintunneld._tcp.platform.example",
		"_us-v2-origintunneld._tcp.platform.example",
	}, asked)
	assert.NotContains(t, err.Error(), "argotunnel")
}

func TestEdgeDiscoveryWithoutASuffixAsksNothing(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	defer func(s func(string, string, string) (string, []*net.SRV, error)) { netLookupSRV = s }(netLookupSRV)
	netLookupSRV = func(_, _, name string) (string, []*net.SRV, error) {
		t.Fatalf("looked up an SRV record under %q with no platform suffix", name)
		return "", nil, nil
	}
	l := zerolog.Nop()
	_, err := EdgeDiscovery(&l, RegionalServiceName(""))
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
}

func TestTheDoTFallbackWithoutAResolverDialsNothing(t *testing.T) {
	t.Setenv(apsy.DoTAddrEnv, "")
	_, addrs, err := lookupSRVWithDOT(srvService, srvProto, "platform.example")
	require.ErrorIs(t, err, apsy.ErrNoDoTResolver)
	assert.Nil(t, addrs)
}

func TestTheDoTFallbackDialsTheNamedResolver(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	serverNames := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverNames <- "accept: " + err.Error()
			return
		}
		defer conn.Close()
		server := tls.Server(conn, &tls.Config{GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			serverNames <- hello.ServerName
			return nil, errors.New("no certificate")
		}})
		_ = server.Handshake()
	}()
	t.Setenv(apsy.DoTAddrEnv, listener.Addr().String())
	t.Setenv(apsy.DoTServerNameEnv, "dns.platform.example")
	_, _, err = lookupSRVWithDOT(srvService, srvProto, "platform.example")
	require.Error(t, err)
	assert.Equal(t, "dns.platform.example", <-serverNames)
}

func TestEdgeDiscoveryAcceptsTheDoTFallbackAnswer(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	defer func(s func(string, string, string) (string, []*net.SRV, error)) { netLookupSRV = s }(netLookupSRV)
	defer func(s func(string, string, string) (string, []*net.SRV, error)) { fallbackLookupSRV = s }(fallbackLookupSRV)
	defer func(l func(string) ([]net.IP, error)) { netLookupIP = l }(netLookupIP)
	mockAddrs := newMockAddrs(19, 2, 5)
	netLookupSRV = func(string, string, string) (string, []*net.SRV, error) {
		return "", nil, errors.New("compressed SRV records")
	}
	var fallbackAsked []string
	fallbackLookupSRV = func(service, proto, name string) (string, []*net.SRV, error) {
		fallbackAsked = append(fallbackAsked, "_"+service+"._"+proto+"."+name)
		return mockNetLookupSRV(mockAddrs)(service, proto, name)
	}
	netLookupIP = mockNetLookupIP(mockAddrs)
	l := zerolog.Nop()
	addrLists, err := EdgeDiscovery(&l, RegionalServiceName(""))
	require.NoError(t, err)
	assert.Equal(t, []string{"_v2-origintunneld._tcp.platform.example"}, fallbackAsked)
	assert.Len(t, addrLists, len(mockAddrs.addrMap))
}
