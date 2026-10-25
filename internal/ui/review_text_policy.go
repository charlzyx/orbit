package orbitui

import (
	"github.com/egoist/mygo/ui"
	"runtime"
)

// This immutable policy is resolved on the UI owner before any Review measure
// or layout. Unsupported platforms keep their original native behavior.
type reviewTextPolicy struct {
	enabled bool
	status  string
}

func resolveReviewTextPolicy(platform string, available func() bool) reviewTextPolicy {
	if platform != "linux" {
		return reviewTextPolicy{status: "unchanged-platform"}
	}
	if !available() {
		return reviewTextPolicy{status: "unsupported-native-api"}
	}
	return reviewTextPolicy{enabled: true, status: "enabled"}
}
func (a *App) reviewTextPolicy() reviewTextPolicy {
	if a.reviewState.textPolicy == nil {
		policy := resolveReviewTextPolicy(runtime.GOOS, ui.OrbitFractionalGlyphPositionsAvailable)
		a.reviewState.textPolicy = &policy
	}
	return *a.reviewState.textPolicy
}
func (p reviewTextPolicy) apply(e ui.Element) ui.Element {
	if p.enabled {
		return e.OrbitFractionalGlyphPositions(true)
	}
	return e
}
func reviewCodeKind(kind string) bool {
	return kind == "context" || kind == "addition" || kind == "deletion"
}
func (p reviewTextPolicy) measure(c *ui.Context, kind string, span ui.Span) (float32, float32) {
	if p.enabled && reviewCodeKind(kind) {
		return c.MeasureTextOrbitFractional(0, span)
	}
	return c.MeasureText(0, span)
}
