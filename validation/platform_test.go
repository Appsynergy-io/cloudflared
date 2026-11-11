package validation

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

// signedValidator is an Access validator that trusts issuer and the key it returns, so Validate
// reaches its issuer check.
func signedValidator(t *testing.T, issuer string) (*Access, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: key}, nil)
	require.NoError(t, err)
	now := time.Now()
	payload, err := json.Marshal(map[string]any{
		"iss": issuer,
		"aud": "app-aud",
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	})
	require.NoError(t, err)
	jws, err := signer.Sign(payload)
	require.NoError(t, err)
	token, err := jws.CompactSerialize()
	require.NoError(t, err)
	keySet := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{key.Public()}}
	verifier := oidc.NewVerifier(issuer, keySet, &oidc.Config{ClientID: "app-aud", SupportedSigningAlgs: []string{string(jose.ES256)}})
	return &Access{verifier}, token
}

func TestAnIssuerUnderTheAccessDomainIsAccepted(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "platformaccess.example")
	access, token := signedValidator(t, "https://myteam.platformaccess.example")
	require.NoError(t, access.Validate(context.Background(), token))
}

func TestAnIssuerOutsideTheAccessDomainIsRefused(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "platformaccess.example")
	for _, issuer := range []string{
		"https://myteam.platform.example",
		"https://api.platform.example",
		"https://myteam.cloudflareaccess.com",
	} {
		access, token := signedValidator(t, issuer)
		require.Error(t, access.Validate(context.Background(), token), issuer)
	}
}

func TestNoAccessDomainAcceptsNoIssuer(t *testing.T) {
	// The platform suffix is no substitute: the Access domain is never derived from it.
	t.Setenv(apsy.SuffixEnv, "platform.example")
	t.Setenv(apsy.AccessDomainEnv, "")
	access, token := signedValidator(t, "https://myteam.platform.example")
	err := access.Validate(context.Background(), token)
	require.ErrorIs(t, err, apsy.ErrNoAccessDomain)
	require.ErrorContains(t, err, apsy.AccessDomainEnv)
}
