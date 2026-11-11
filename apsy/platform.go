// Package apsy names the platform this build of cloudflared connects to. Every default that
// upstream spells under argotunnel.com or cloudflare.com -- the hosts Cloudflare lists for
// cloudflared to reach -- is spelled under the platform
// suffix instead, the domain the operator's platform is served under the way cloudflare.com is
// Cloudflare's. The suffix is an operator input, never a constant:
// it is built in with -ldflags "-X github.com/cloudflare/cloudflared/apsy.builtSuffix=SUFFIX"
// (the release build passes the operator's suffix) and TUNNEL_PLATFORM_SUFFIX
// overrides it at run time. A build with neither discovers nothing: the edge must then be given
// by --edge or TUNNEL_EDGE, and no Cloudflare host is ever the fallback.
package apsy

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// SuffixEnv overrides the built-in suffix.
const SuffixEnv = "TUNNEL_PLATFORM_SUFFIX"

// builtSuffix is set at link time; empty unless the build named one.
var builtSuffix = ""

// ErrNoSuffix is what every platform host answers when no platform suffix is known; it names the
// setting that supplies one.
var ErrNoSuffix = errors.New("no platform suffix: this build names none and " + SuffixEnv + " is unset")

// Suffix is the platform suffix, from TUNNEL_PLATFORM_SUFFIX when set, else the one built in.
func Suffix() (string, error) {
	return domainSetting(SuffixEnv, builtSuffix, ErrNoSuffix)
}

// setting is the value of env when set, else the one built in; unset is "".
func setting(env, built string) string {
	if s := os.Getenv(env); s != "" {
		return s
	}
	return built
}

// domainSetting is a domain name read from env or built in, lower-cased without its root dot;
// unset answers errUnset, which names the setting.
func domainSetting(env, built string, errUnset error) (string, error) {
	s := setting(env, built)
	if s == "" {
		return "", errUnset
	}
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	if err := validName(s); err != nil {
		return "", fmt.Errorf("%s %q: %w", env, s, err)
	}
	return s, nil
}

// Host is label.SUFFIX, the platform's own spelling of a host upstream names under
// argotunnel.com or cloudflare.com.
func Host(label string) (string, error) {
	s, err := Suffix()
	if err != nil {
		return "", err
	}
	return label + "." + s, nil
}

func validName(s string) error {
	if len(s) > 253 {
		return errors.New("longer than 253 characters")
	}
	labels := strings.Split(s, ".")
	if len(labels) < 2 {
		return errors.New("not a domain name")
	}
	for _, l := range labels {
		if l == "" || len(l) > 63 || strings.HasPrefix(l, "-") || strings.HasSuffix(l, "-") {
			return fmt.Errorf("label %q is not a DNS label", l)
		}
		for _, c := range l {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return fmt.Errorf("label %q is not a DNS label", l)
			}
		}
	}
	return nil
}
