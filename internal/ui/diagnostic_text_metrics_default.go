//go:build !orbit_diagnostic_fractional_text_linux

package orbitui

// Production builds retain the pinned toolkit's unmodified text metrics policy.
func configureDiagnosticTextMetrics() error { return nil }
