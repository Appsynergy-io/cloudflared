package apsy

import (
	"errors"
	"strings"
	"testing"
)

func withBuiltDoT(t *testing.T, addr, name string) {
	t.Helper()
	oldAddr, oldName := builtDoTAddr, builtDoTServerName
	builtDoTAddr, builtDoTServerName = addr, name
	t.Cleanup(func() { builtDoTAddr, builtDoTServerName = oldAddr, oldName })
}

func TestNoDoTResolverByDefault(t *testing.T) {
	withBuiltDoT(t, "", "")
	t.Setenv(DoTAddrEnv, "")
	t.Setenv(DoTServerNameEnv, "dns.platform.example")
	if _, _, err := DoTResolver(); !errors.Is(err, ErrNoDoTResolver) {
		t.Fatalf("DoTResolver() err = %v, want ErrNoDoTResolver", err)
	}
	if !strings.Contains(ErrNoDoTResolver.Error(), DoTAddrEnv) {
		t.Fatalf("ErrNoDoTResolver does not name %s: %v", DoTAddrEnv, ErrNoDoTResolver)
	}
}

func TestBuiltDoTResolverAndItsOverride(t *testing.T) {
	withBuiltDoT(t, "192.0.2.53:853", "dns.platform.example")
	t.Setenv(DoTAddrEnv, "")
	t.Setenv(DoTServerNameEnv, "")
	addr, name, err := DoTResolver()
	if err != nil || addr != "192.0.2.53:853" || name != "dns.platform.example" {
		t.Fatalf("DoTResolver() = %q, %q, %v", addr, name, err)
	}
	t.Setenv(DoTAddrEnv, "[2001:db8::53]:853")
	t.Setenv(DoTServerNameEnv, "DNS.Other.Example.")
	addr, name, err = DoTResolver()
	if err != nil || addr != "[2001:db8::53]:853" || name != "dns.other.example" {
		t.Fatalf("DoTResolver() = %q, %q, %v", addr, name, err)
	}
}

func TestADoTAddressWithoutAServerNameNamesTheSetting(t *testing.T) {
	withBuiltDoT(t, "", "")
	t.Setenv(DoTAddrEnv, "192.0.2.53:853")
	t.Setenv(DoTServerNameEnv, "")
	_, _, err := DoTResolver()
	if err == nil || !strings.Contains(err.Error(), DoTServerNameEnv) {
		t.Fatalf("DoTResolver() err = %v, want one naming %s", err, DoTServerNameEnv)
	}
}

func TestADoTAddressThatIsNotAnIPAndPortIsRefused(t *testing.T) {
	withBuiltDoT(t, "", "")
	t.Setenv(DoTServerNameEnv, "dns.platform.example")
	for _, bad := range []string{"192.0.2.53", "dns.platform.example:853", "192.0.2.53:", "tls://192.0.2.53:853"} {
		t.Setenv(DoTAddrEnv, bad)
		if addr, _, err := DoTResolver(); err == nil {
			t.Errorf("DoTResolver() = %q for %q", addr, bad)
		}
	}
}
