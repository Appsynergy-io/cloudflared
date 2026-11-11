//go:build !cfgo

package crypto

// Cloudflare Go sets the cfgo tag; its private legacy Kyber group is absent in standard Go.
const legacyKyberInToolchain = false
