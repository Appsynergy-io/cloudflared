package apsy

import (
	"errors"
	"testing"
)

func withBuilt(t *testing.T, s string) {
	t.Helper()
	old := builtSuffix
	builtSuffix = s
	t.Cleanup(func() { builtSuffix = old })
}

func TestNoSuffixRefusesRatherThanFallingBack(t *testing.T) {
	withBuilt(t, "")
	t.Setenv(SuffixEnv, "")
	if _, err := Suffix(); !errors.Is(err, ErrNoSuffix) {
		t.Fatalf("Suffix() err = %v, want ErrNoSuffix", err)
	}
	if h, err := Host("cfd-features"); err == nil {
		t.Fatalf("Host() = %q with no suffix", h)
	}
}

func TestBuiltSuffixAndItsOverride(t *testing.T) {
	withBuilt(t, "platform.example")
	t.Setenv(SuffixEnv, "")
	if h, _ := Host("cfd-features"); h != "cfd-features.platform.example" {
		t.Fatalf("Host() = %q", h)
	}
	t.Setenv(SuffixEnv, "Other.Example.")
	if s, _ := Suffix(); s != "other.example" {
		t.Fatalf("Suffix() = %q, want the override lower-cased without its root dot", s)
	}
}

func TestASuffixThatIsNotADomainIsRefused(t *testing.T) {
	withBuilt(t, "")
	for _, bad := range []string{"localhost", "a..b", "-a.b", "a.b/c", "a.b:443", "a b.c", "https://a.b"} {
		t.Setenv(SuffixEnv, bad)
		if s, err := Suffix(); err == nil {
			t.Errorf("Suffix() = %q for %q", s, bad)
		}
	}
}
