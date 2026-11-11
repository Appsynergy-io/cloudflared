package apsy

import (
	"errors"
	"strings"
	"testing"
)

func withTrustDomains(t *testing.T, suffix, access, accessLogin, quick string) {
	t.Helper()
	withBuilt(t, "")
	old := [3]string{builtAccessDomain, builtAccessLoginDomain, builtQuickTunnelDomain}
	builtAccessDomain, builtAccessLoginDomain, builtQuickTunnelDomain = "", "", ""
	t.Cleanup(func() { builtAccessDomain, builtAccessLoginDomain, builtQuickTunnelDomain = old[0], old[1], old[2] })
	t.Setenv(SuffixEnv, suffix)
	t.Setenv(AccessDomainEnv, access)
	t.Setenv(AccessLoginDomainEnv, accessLogin)
	t.Setenv(QuickTunnelDomainEnv, quick)
}

func TestTrustDomainsAreNeverDerivedFromTheSuffix(t *testing.T) {
	withTrustDomains(t, "platform.example", "", "", "")
	for name, get := range map[string]func() (string, error){
		AccessDomainEnv:      AccessDomain,
		AccessLoginDomainEnv: AccessLoginDomain,
		QuickTunnelDomainEnv: QuickTunnelDomain,
	} {
		d, err := get()
		if err == nil {
			t.Fatalf("%s unset gave %q", name, d)
		}
		if !strings.Contains(err.Error(), name) {
			t.Errorf("%s unset: error %q does not name it", name, err)
		}
	}
	if _, err := AccessTeamDomain("myteam", false); !errors.Is(err, ErrNoAccessDomain) {
		t.Errorf("AccessTeamDomain err = %v, want ErrNoAccessDomain", err)
	}
	if _, err := AccessLoginHost("login"); !errors.Is(err, ErrNoAccessLoginDomain) {
		t.Errorf("AccessLoginHost err = %v, want ErrNoAccessLoginDomain", err)
	}
	if _, err := QuickTunnelHost("api"); !errors.Is(err, ErrNoQuickTunnelDomain) {
		t.Errorf("QuickTunnelHost err = %v, want ErrNoQuickTunnelDomain", err)
	}
}

func TestBuiltTrustDomainsAndTheirOverrides(t *testing.T) {
	withTrustDomains(t, "", "", "", "")
	builtAccessDomain, builtAccessLoginDomain, builtQuickTunnelDomain = "platformaccess.example", "platformaccess.test", "tryplatform.example"
	for _, tc := range []struct {
		get  func() (string, error)
		want string
	}{
		{AccessDomain, "platformaccess.example"},
		{AccessLoginDomain, "platformaccess.test"},
		{QuickTunnelDomain, "tryplatform.example"},
	} {
		if got, err := tc.get(); err != nil || got != tc.want {
			t.Errorf("got %q, %v; want %q", got, err, tc.want)
		}
	}
	t.Setenv(AccessDomainEnv, "Other-Access.Example.")
	if got, _ := AccessDomain(); got != "other-access.example" {
		t.Errorf("AccessDomain() = %q, want the override lower-cased without its root dot", got)
	}
}

func TestTrustDomainsMustBeApartFromTheSuffixAndEachOther(t *testing.T) {
	for _, tc := range []struct{ suffix, access, accessLogin, quick, refused string }{
		{"platform.example", "platform.example", "", "", AccessDomainEnv},
		{"platform.example", "access.platform.example", "", "", AccessDomainEnv},
		{"edge.platform.example", "platform.example", "", "", AccessDomainEnv},
		{"platform.example", "", "login.platform.example", "", AccessLoginDomainEnv},
		{"platform.example", "", "", "try.platform.example", QuickTunnelDomainEnv},
		{"platform.example", "shared.example", "", "shared.example", QuickTunnelDomainEnv},
		{"platform.example", "shared.example", "", "shared.example", AccessDomainEnv},
		{"platform.example", "access.example", "x.access.example", "", AccessLoginDomainEnv},
	} {
		withTrustDomains(t, tc.suffix, tc.access, tc.accessLogin, tc.quick)
		get := map[string]func() (string, error){
			AccessDomainEnv:      AccessDomain,
			AccessLoginDomainEnv: AccessLoginDomain,
			QuickTunnelDomainEnv: QuickTunnelDomain,
		}[tc.refused]
		if d, err := get(); err == nil || !strings.Contains(err.Error(), tc.refused) {
			t.Errorf("%+v: %s = %q, %v; want an error naming it", tc, tc.refused, d, err)
		}
	}
}

func TestATeamNameCannotAliasAPlatformHost(t *testing.T) {
	withTrustDomains(t, "platform.example", "platformaccess.example", "platformaccess.test", "tryplatform.example")
	for _, team := range []string{"api", "login", "dash", "management", "API", "login.fed", "x.api"} {
		for _, fedramp := range []bool{false, true} {
			host, err := AccessTeamDomain(team, fedramp)
			if err != nil {
				t.Fatalf("AccessTeamDomain(%q, %v): %v", team, fedramp, err)
			}
			if !strings.HasSuffix(host, ".platformaccess.example") {
				t.Errorf("team %q (fedramp %v) has team domain %q outside the Access domain", team, fedramp, host)
			}
			for _, domain := range []string{"platform.example", "platformaccess.test", "tryplatform.example"} {
				if host == domain || strings.HasSuffix(host, "."+domain) {
					t.Errorf("team %q (fedramp %v) has team domain %q under %q", team, fedramp, host, domain)
				}
			}
			for _, label := range []string{"api", "login", "dash", "login.fed"} {
				platformHost, _ := Host(label)
				loginHost, _ := AccessLoginHost(label)
				if host == platformHost || host == loginHost {
					t.Errorf("team %q (fedramp %v) aliases the platform host %q", team, fedramp, host)
				}
			}
		}
	}
	if host, _ := AccessTeamDomain("myteam", true); host != "myteam.fed.platformaccess.example" {
		t.Errorf("FedRAMP team domain = %q", host)
	}
}

func TestATeamNameThatLeavesTheAccessDomainIsRefused(t *testing.T) {
	withTrustDomains(t, "platform.example", "platformaccess.example", "", "")
	for _, team := range []string{"", "evil.example/", "evil.example#", "evil.example?", "user@evil.example", "evil.example:443", "a..b", "-a", "a b"} {
		if host, err := AccessTeamDomain(team, false); err == nil {
			t.Errorf("AccessTeamDomain(%q) = %q", team, host)
		}
	}
}
