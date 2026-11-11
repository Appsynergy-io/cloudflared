package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
	"github.com/cloudflare/cloudflared/ingress/middleware"
)

const stockAccessRule = `
ingress:
 - hostname: app.example.com
   service: https://localhost:8080
   originRequest:
     access:
       required: true
       teamName: myteam
       audTag: [aud]
 - service: http_status:404
`

func TestAStockAccessTeamNameParses(t *testing.T) {
	ing, err := ParseIngress(MustReadIngress(stockAccessRule))
	require.NoError(t, err)
	require.Len(t, ing.Rules[0].Handlers, 1)
	require.IsType(t, &middleware.JWTValidator{}, ing.Rules[0].Handlers[0])
}

func TestAnAccessRuleWithoutAnAccessDomainIsNotParsed(t *testing.T) {
	// The platform suffix is no substitute: the Access domain is never derived from it.
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "")
	_, err := ParseIngress(MustReadIngress(stockAccessRule))
	require.ErrorIs(t, err, apsy.ErrNoAccessDomain)
	assert.Contains(t, err.Error(), `access.teamName "myteam"`)
	assert.Contains(t, err.Error(), apsy.AccessDomainEnv)
}
