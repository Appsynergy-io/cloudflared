package features

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/apsy"
)

func TestFeaturesRecordIsThePlatforms(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "platform.example")
	var asked []string
	dr := newDNSResolver()
	dr.lookupTXT = func(_ context.Context, name string) ([]string, error) {
		asked = append(asked, name)
		return nil, errors.New("no such host")
	}
	_, err := dr.lookupRecord(context.Background())
	require.Error(t, err)
	assert.Equal(t, []string{"cfd-features.platform.example"}, asked)
}

func TestNoSuffixAsksForNoFeaturesRecord(t *testing.T) {
	t.Setenv(apsy.SuffixEnv, "")
	dr := newDNSResolver()
	dr.lookupTXT = func(_ context.Context, name string) ([]string, error) {
		t.Fatalf("looked up %q with no platform suffix", name)
		return nil, nil
	}
	_, err := dr.lookupRecord(context.Background())
	require.ErrorIs(t, err, apsy.ErrNoSuffix)
}
