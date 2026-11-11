package quicktunnelauth

import "github.com/cloudflare/cloudflared/apsy"

// Upstream names quick tunnels under trycloudflare.com and its broker login.trycloudflare.com,
// a registrable domain apart from Cloudflare's own; here that is the Quick Tunnel domain, apart
// from the platform suffix, labels kept under it.

func quickTunnelAuthBrokerURL(path string) (string, error) {
	host, err := apsy.QuickTunnelHost("login")
	if err != nil {
		return "", err
	}
	return "https://" + host + path, nil
}

// QuickTunnelAuthBrokerAuthorizeURL is upstream's login.trycloudflare.com/authorize under the Quick
// Tunnel domain.
func QuickTunnelAuthBrokerAuthorizeURL() (string, error) {
	return quickTunnelAuthBrokerURL("/authorize")
}

func brokerIssuer() (string, error) {
	return quickTunnelAuthBrokerURL("")
}

func brokerJWKSURL() (string, error) {
	return quickTunnelAuthBrokerURL("/.well-known/jwks.json")
}
