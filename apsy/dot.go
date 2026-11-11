package apsy

import (
	"errors"
	"fmt"
	"net"
)

// Edge discovery falls back to DNS over TLS when the system resolver cannot answer the SRV
// lookup (cloudflared GitHub issue #75). Upstream dials Cloudflare's resolver, 1.1.1.1:853 as
// cloudflare-dns.com; here the resolver is an operator input like the suffix: built in with
// -ldflags "-X github.com/cloudflare/cloudflared/apsy.builtDoTAddr=IP:PORT -X
// github.com/cloudflare/cloudflared/apsy.builtDoTServerName=NAME", overridden by the env vars
// below. Unset, there is no fallback and the system resolver's answer stands.
const (
	DoTAddrEnv       = "TUNNEL_PLATFORM_DOT_ADDR"
	DoTServerNameEnv = "TUNNEL_PLATFORM_DOT_SERVER_NAME"
)

var (
	builtDoTAddr       = ""
	builtDoTServerName = ""
)

// ErrNoDoTResolver is the fallback's answer when no DNS-over-TLS resolver is named.
var ErrNoDoTResolver = errors.New("no DNS-over-TLS resolver: this build names none and " + DoTAddrEnv + " is unset")

// DoTResolver is the DNS-over-TLS resolver's address (IP:port) and the name its certificate is
// verified against.
func DoTResolver() (addr, serverName string, err error) {
	addr = setting(DoTAddrEnv, builtDoTAddr)
	if addr == "" {
		return "", "", ErrNoDoTResolver
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil || net.ParseIP(host) == nil || port == "" {
		return "", "", fmt.Errorf("%s %q: not an IP:port address", DoTAddrEnv, addr)
	}
	serverName, err = domainSetting(DoTServerNameEnv, builtDoTServerName,
		fmt.Errorf("%s is %q but %s, the name its certificate is verified against, is unset", DoTAddrEnv, addr, DoTServerNameEnv))
	if err != nil {
		return "", "", err
	}
	return addr, serverName, nil
}
