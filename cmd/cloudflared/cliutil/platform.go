package cliutil

import (
	"fmt"

	"github.com/cloudflare/cloudflared/apsy"
)

// PlatformDefault is a flag's default host or URL when the flags are built, read from the
// platform's settings (package apsy); with the setting unset there is none, and the command that
// needs the value fails with the error naming the setting.
func PlatformDefault(value func() (string, error)) string {
	s, _ := value()
	return s
}

// RequirePlatformValue is a flag's value, or apsy.ErrNoSuffix naming the flag when it is empty
// because the build has no platform suffix to default it from.
func RequirePlatformValue(flag, value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("--%s has no default host: %w", flag, apsy.ErrNoSuffix)
	}
	return value, nil
}
