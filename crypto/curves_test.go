package crypto

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/features"
)

// TestCurvePreferences verifies that GetCurvePreferences returns the
// documented curve list for each supported PostQuantumMode. The expected
// values correspond to the contract described in the package documentation
// and must be identical under FIPS and non-FIPS builds (see TUN-10413).
func TestCurvePreferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedCurves []tls.CurveID
		pqMode         features.PostQuantumMode
	}{
		{
			name:           "Prefer PQ",
			pqMode:         features.PostQuantumPrefer,
			expectedCurves: []tls.CurveID{tls.X25519MLKEM768, P256Kyber768Draft00, tls.CurveP256},
		},
		{
			name:           "Strict PQ",
			pqMode:         features.PostQuantumStrict,
			expectedCurves: []tls.CurveID{tls.X25519MLKEM768, P256Kyber768Draft00},
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()
			curves, err := getCurvePreferences(tcase.pqMode)
			require.NoError(t, err)
			require.Equal(t, tcase.expectedCurves, curves)
		})
	}
}

// TestCurvePreferenceUnknownMode asserts that passing a PostQuantumMode
// value outside of the documented constants produces an error instead of
// silently returning a nil or default curve list. This protects callers
// from accidentally negotiating with an unintended curve set.
func TestCurvePreferenceUnknownMode(t *testing.T) {
	t.Parallel()

	_, err := getCurvePreferences(features.PostQuantumMode(255))
	require.Error(t, err)
}

// TestReturnedSliceIsIndependent ensures GetCurvePreferences returns a
// fresh slice on every call, so that callers cannot corrupt the
// package-level defaults by mutating the result.
func TestReturnedSliceIsIndependent(t *testing.T) {
	t.Parallel()

	first, err := getCurvePreferences(features.PostQuantumPrefer)
	require.NoError(t, err)
	// Mutate the returned slice.
	first[0] = tls.CurveP521

	second, err := getCurvePreferences(features.PostQuantumPrefer)
	require.NoError(t, err)
	require.Equal(t, tls.X25519MLKEM768, second[0], "package defaults must not be affected by caller mutation")
}

// runClientServerHandshake drives a TLS 1.3 handshake with the given curve
// preferences set on the client and captures the SupportedCurves list
// advertised by the client in its ClientHello. The helper is used by
// TestSupportedCurvesNegotiation to exercise the curves end-to-end against
// the standard library's TLS stack.
func runClientServerHandshake(t *testing.T, curves []tls.CurveID) []tls.CurveID {
	var advertisedCurves []tls.CurveID
	ts := httptest.NewUnstartedServer(nil)
	ts.TLS = &tls.Config{ // nolint: gosec
		GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			advertisedCurves = slices.Clone(chi.SupportedCurves)
			return nil, nil
		},
	}
	ts.StartTLS()
	defer ts.Close()
	clientTLSConfig := ts.Client().Transport.(*http.Transport).TLSClientConfig
	clientTLSConfig.CurvePreferences = curves
	resp, err := ts.Client().Head(ts.URL)
	if err != nil {
		t.Error(err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	return advertisedCurves
}

// TestSupportedCurvesNegotiation verifies the exact advertised order for each mode.
// Standard Go omits Cloudflare's private legacy curve; cfgo implements it. Both are
// supported upstream build paths, and neither may omit ML-KEM or reorder preferences.
func TestSupportedCurvesNegotiation(t *testing.T) {
	t.Parallel()
	for _, mode := range []features.PostQuantumMode{features.PostQuantumPrefer, features.PostQuantumStrict} {
		curves, err := getCurvePreferences(mode)
		require.NoError(t, err)
		want := []tls.CurveID{tls.X25519MLKEM768}
		if legacyKyberInToolchain {
			want = append(want, P256Kyber768Draft00)
		}
		if mode == features.PostQuantumPrefer {
			want = append(want, tls.CurveP256)
		}
		require.Equal(t, want, runClientServerHandshake(t, curves))
	}
}
