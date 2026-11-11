package updater

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/facebookgo/grace/gracenet"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/cliutil"
)

type noNetwork struct{ t *testing.T }

func (n noNetwork) RoundTrip(req *http.Request) (*http.Response, error) {
	n.t.Errorf("the update check went to the network: %s", req.URL)
	return nil, http.ErrUseLastResponse
}

// withoutNetwork fails the test on any request the default HTTP transport is asked to send.
func withoutNetwork(t *testing.T) {
	t.Helper()
	oldTransport, oldInfo := http.DefaultTransport, buildInfo
	http.DefaultTransport = noNetwork{t}
	buildInfo = &cliutil.BuildInfo{CloudflaredVersion: "2026.10.7"}
	t.Cleanup(func() { http.DefaultTransport, buildInfo = oldTransport, oldInfo })
}

func TestCloudflaresUpdateServiceAnswersNoUpdateWithoutTheNetwork(t *testing.T) {
	withoutNetwork(t)
	for _, opts := range []updateOptions{
		{},
		{isStaging: true},
		{isBeta: true, isForced: true, intendedVersion: "2026.11.0"},
	} {
		result, err := CheckForUpdate(opts)
		require.NoError(t, err, "%+v", opts)
		assert.Equal(t, "", result.Version(), "%+v", opts)
		assert.Equal(t, "", result.UserMessage(), "%+v", opts)
		outcome := applyUpdate(opts, result)
		assert.True(t, outcome.noUpdate(), "%+v", opts)
		assert.NoError(t, outcome.Error, "%+v", opts)
	}
}

func TestAnyOtherUpdateServiceIsStillAsked(t *testing.T) {
	asked := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked++
		assert.Equal(t, "2026.10.7", r.URL.Query().Get(ClientVersionName))
		_, _ = w.Write([]byte(`{"shouldUpdate":true,"version":"2026.11.0","url":"https://releases.platform.example/cloudflared","checksum":"abc"}`))
	}))
	defer server.Close()
	result, err := NewWorkersService("2026.10.7", server.URL, "cloudflared", Options{}).Check()
	require.NoError(t, err)
	assert.Equal(t, 1, asked)
	assert.Equal(t, "2026.11.0", result.Version())
}

func TestTheUpdateCommandFindsCloudflaredUpToDate(t *testing.T) {
	withoutNetwork(t)
	flagSet := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	flagSet.Bool("beta", false, "")
	flagSet.Bool("staging", false, "")
	flagSet.Bool("force", false, "")
	flagSet.String("version", "", "")
	require.NoError(t, Update(cli.NewContext(cli.NewApp(), flagSet, nil)))
}

func TestTheStartupCheckWarnsOfNothing(t *testing.T) {
	withoutNetwork(t)
	flagSet := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	flagSet.Bool("beta", false, "")
	flagSet.Bool("staging", false, "")
	checker := StartWarningCheck(cli.NewContext(cli.NewApp(), flagSet, nil))
	// Blocks until the check has answered, where getWarning would not wait for it.
	warning, answered := <-checker.warningChan
	assert.True(t, answered)
	assert.Equal(t, "", warning)
}

// --no-autoupdate and --autoupdate-freq keep upstream's meaning: where upstream supports
// autoupdate at all, they decide whether the check runs and how often.
func TestTheAutoupdateFlagsKeepTheirMeaning(t *testing.T) {
	log := zerolog.Nop()
	supported := supportAutoUpdate(&log)
	for _, tc := range []struct {
		disabled bool
		freq     time.Duration
	}{
		{false, time.Hour},
		{true, time.Hour},
		{false, 0},
	} {
		want := supported && !tc.disabled && tc.freq != 0
		autoupdater := NewAutoUpdater(tc.disabled, tc.freq, &gracenet.Net{}, &log)
		assert.Equal(t, want, autoupdater.configurable.enabled, "%+v", tc)
		if want {
			assert.Equal(t, tc.freq, autoupdater.configurable.freq, "%+v", tc)
		}
	}
}
