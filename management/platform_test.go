package management

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

func corsAllows(t *testing.T, origin string) bool {
	t.Helper()
	mgmt := New("management.platform.example", false, "", uuid.Nil, "", &noopLogger, nil)
	req := httptest.NewRequest("GET", "https://management.platform.example/ping?access_token="+validToken, nil)
	req.Header.Set("Origin", origin)
	recorder := httptest.NewRecorder()
	mgmt.ServeHTTP(recorder, req)
	return recorder.Result().Header.Get("Access-Control-Allow-Origin") == origin
}

func TestTheDashboardOriginsAreThePlatforms(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	for _, origin := range []string{"https://dash.platform.example", "https://one.dash.platform.example", "https://dash.fed.platform.example", "https://DASH.platform.example"} {
		assert.True(t, corsAllows(t, origin), origin)
	}
	for origin, why := range map[string]string{
		"https://dash.cloudflare.com":                "an upstream dashboard is a foreign origin here",
		"http://dash.platform.example":               "only https",
		"https://dash.platform.example:8443":         "only the default port",
		"https://dash.platform.example/":             "an origin has no path",
		"https://evilplatform.example":               "not under the suffix",
		"https://platform.example.evil.test":         "not under the suffix",
		"https://platform.example":                   "the suffix itself is no dashboard",
		"https://api.platform.example":               "a platform host that is no dashboard",
		"https://myzone.platform.example":            "a name under the suffix that is no dashboard",
		"https://x.dash.platform.example":            "a name under a dashboard",
		"https://attacker.one.dash.platform.example": "a name under a dashboard",
		"https://label.tryplatform.example":          "a Quick Tunnel host, wherever its domain",
		"https://dash.tryplatform.example":           "a Quick Tunnel host named like a dashboard",
		"https://label.try.platform.example":         "no Quick Tunnel host is under the suffix",
		"https://dash.platformaccess.example":        "an Access team domain named like a dashboard",
		"https://myteam.platformaccess.example":      "an Access team domain, wherever its domain",
		"https://login.platformaccess.test":          "the Access login host",
	} {
		assert.False(t, corsAllows(t, origin), "%s: %s", origin, why)
	}
}

func logStreamAllows(t *testing.T, origin string) bool {
	t.Helper()
	mgmt := New("management.platform.example", false, "", uuid.Nil, "", &noopLogger, nil)
	server := httptest.NewServer(mgmt)
	t.Cleanup(server.Close)
	req, err := http.NewRequest("GET", server.URL+"/logs?access_token="+validToken, nil)
	require.NoError(t, err)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Origin", origin)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusSwitchingProtocols
}

func TestTheLogStreamOriginsAreTheDashboards(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	assert.Equal(t, []string{"dash.platform.example", "one.dash.platform.example", "dash.fed.platform.example"}, dashHosts())
	assert.True(t, logStreamAllows(t, "https://one.dash.platform.example"))
	for _, origin := range []string{
		"https://api.platform.example",
		"https://myzone.platform.example",
		"https://x.dash.platform.example",
		"https://label.tryplatform.example",
		"https://label.try.platform.example",
		"https://dash.cloudflare.com",
	} {
		assert.False(t, logStreamAllows(t, origin), origin)
	}
}

func TestNoSuffixAllowsNoCrossOrigin(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	assert.False(t, corsAllows(t, "https://dash.platform.example"))
	assert.False(t, corsAllows(t, "https://dash.cloudflare.com"))
	assert.Empty(t, dashHosts())
	assert.False(t, logStreamAllows(t, "https://dash.platform.example"))
}
