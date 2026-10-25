//go:build orbit_diagnostic_fractional_text_linux && (!linux || (!amd64 && !arm64))

package orbitui

import "errors"

func configureDiagnosticTextMetrics() error {
	return errors.New("Orbit DIAGNOSTIC fractional text metrics require Linux amd64 or arm64; no fallback")
}
