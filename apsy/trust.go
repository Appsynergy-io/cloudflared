package apsy

import (
	"errors"
	"fmt"
	"strings"
)

// Cloudflare serves the names its users choose on registrable domains of their own, apart from
// the hosts it trusts: Access team domains are {team}.cloudflareaccess.com, the Access login and
// token transfer service is login.cloudflareaccess.org, and Quick Tunnels are
// <label>.trycloudflare.com. Spelling them under the platform suffix would put user-chosen names
// inside the trusted one, so each is a setting of its own, built in like the suffix with
// -ldflags "-X github.com/cloudflare/cloudflared/apsy.builtAccessDomain=..." (and
// builtAccessLoginDomain, builtQuickTunnelDomain) and overridden by the env vars below. None is
// derived from the suffix, and none may equal or lie under (or over) the suffix or another of
// them. Unset, the feature that needs one fails with an error naming it.
const (
	// AccessDomainEnv is upstream's cloudflareaccess.com: team domains and JWT auth domains.
	AccessDomainEnv = "TUNNEL_PLATFORM_ACCESS_DOMAIN"
	// AccessLoginDomainEnv is upstream's cloudflareaccess.org: the login and transfer service.
	AccessLoginDomainEnv = "TUNNEL_PLATFORM_ACCESS_LOGIN_DOMAIN"
	// QuickTunnelDomainEnv is upstream's trycloudflare.com: Quick Tunnels and their broker.
	QuickTunnelDomainEnv = "TUNNEL_PLATFORM_QUICK_TUNNEL_DOMAIN"
)

var (
	builtAccessDomain      = ""
	builtAccessLoginDomain = ""
	builtQuickTunnelDomain = ""
)

var (
	ErrNoAccessDomain      = errors.New("no Access domain: this build names none and " + AccessDomainEnv + " is unset")
	ErrNoAccessLoginDomain = errors.New("no Access login domain: this build names none and " + AccessLoginDomainEnv + " is unset")
	ErrNoQuickTunnelDomain = errors.New("no Quick Tunnel domain: this build names none and " + QuickTunnelDomainEnv + " is unset")
)

type domainSettingSpec struct {
	env      string
	built    func() string
	errUnset error
}

func domainSettings() []domainSettingSpec {
	return []domainSettingSpec{
		{SuffixEnv, func() string { return builtSuffix }, ErrNoSuffix},
		{AccessDomainEnv, func() string { return builtAccessDomain }, ErrNoAccessDomain},
		{AccessLoginDomainEnv, func() string { return builtAccessLoginDomain }, ErrNoAccessLoginDomain},
		{QuickTunnelDomainEnv, func() string { return builtQuickTunnelDomain }, ErrNoQuickTunnelDomain},
	}
}

// separateDomain is the domain setting env, refused when it equals, lies under or lies over the
// value of another domain setting.
func separateDomain(env string) (string, error) {
	var mine string
	for _, d := range domainSettings() {
		if d.env != env {
			continue
		}
		var err error
		if mine, err = domainSetting(d.env, d.built(), d.errUnset); err != nil {
			return "", err
		}
	}
	for _, d := range domainSettings() {
		if d.env == env {
			continue
		}
		other, err := domainSetting(d.env, d.built(), d.errUnset)
		if err != nil {
			continue
		}
		if mine == other || strings.HasSuffix(mine, "."+other) || strings.HasSuffix(other, "."+mine) {
			return "", fmt.Errorf("%s %q overlaps %s %q: each must be a domain of its own", env, mine, d.env, other)
		}
	}
	return mine, nil
}

// AccessDomain is the domain of Access team domains and auth domains, upstream's
// cloudflareaccess.com.
func AccessDomain() (string, error) { return separateDomain(AccessDomainEnv) }

// AccessLoginDomain is the domain of the Access login and token transfer service, upstream's
// cloudflareaccess.org.
func AccessLoginDomain() (string, error) { return separateDomain(AccessLoginDomainEnv) }

// QuickTunnelDomain is the domain Quick Tunnels and their broker are served under, upstream's
// trycloudflare.com.
func QuickTunnelDomain() (string, error) { return separateDomain(QuickTunnelDomainEnv) }

// AccessTeamDomain is the team domain of an Access team, upstream's {team}.cloudflareaccess.com
// or, for FedRAMP, {team}.fed.cloudflareaccess.com, under the Access domain. It is always a DNS
// name under the Access domain, never a platform host.
func AccessTeamDomain(team string, fedramp bool) (string, error) {
	domain, err := AccessDomain()
	if err != nil {
		return "", err
	}
	host := strings.ToLower(team) + "."
	if fedramp {
		host += "fed."
	}
	host += domain
	if err := validName(host); err != nil {
		return "", fmt.Errorf("Access team %q: %q: %w", team, host, err)
	}
	return host, nil
}

// AccessLoginHost is label.ACCESS_LOGIN_DOMAIN, upstream's label.cloudflareaccess.org.
func AccessLoginHost(label string) (string, error) {
	domain, err := AccessLoginDomain()
	if err != nil {
		return "", err
	}
	return label + "." + domain, nil
}

// QuickTunnelHost is label.QUICK_TUNNEL_DOMAIN, upstream's label.trycloudflare.com.
func QuickTunnelHost(label string) (string, error) {
	domain, err := QuickTunnelDomain()
	if err != nil {
		return "", err
	}
	return label + "." + domain, nil
}
