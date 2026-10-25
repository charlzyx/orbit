//go:build orbit_diagnostic_fractional_text_linux && linux && (amd64 || arm64)

package orbitui

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// This is a diagnostic-only startup choice, never init or a runtime toggle.
// Call before LoadAssets and before constructing or measuring any native text.
func configureDiagnosticTextMetrics() error {
	if err := ui.EnableOrbitFractionalTextMetrics(); err != nil {
		return fmt.Errorf("Orbit DIAGNOSTIC fractional text metrics: %w", err)
	}
	return nil
}
