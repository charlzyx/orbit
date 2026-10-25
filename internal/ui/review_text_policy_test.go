package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"math"
	"runtime"
	"testing"
)

func TestReviewTextCapabilityFallback(t *testing.T) {
	for _, tc := range []struct {
		platform           string
		available, enabled bool
		status             string
		calls              int
	}{
		{"darwin", true, false, "unchanged-platform", 0}, {"windows", false, false, "unchanged-platform", 0},
		{"linux", false, false, "unsupported-native-api", 1}, {"linux", true, true, "enabled", 1},
	} {
		calls := 0
		p := resolveReviewTextPolicy(tc.platform, func() bool { calls++; return tc.available })
		if p.enabled != tc.enabled || p.status != tc.status || calls != tc.calls {
			t.Fatalf("%+v: policy %+v calls %d", tc, p, calls)
		}
	}
}
func TestReviewTextLocalPolicyMeasureAndCache(t *testing.T) {
	if runtime.GOOS != "linux" || !ui.OrbitFractionalGlyphPositionsAvailable() {
		t.Skip("native fractional API unavailable; fallback covered separately")
	}
	a := fixture(t)
	a.reviewState.wrap = true
	line := workspace.ReviewLine{Kind: "addition", NewLine: 7, Content: "Untracked fixture note."}
	a.reviewState.data.Lines = []workspace.ReviewLine{line}
	var width, ordinary, fractional, header float32
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		span := ui.Span{Text: line.Content, Font: a.Fonts.CodeFamily, Size: a.reviewCodeFontSize()}
		ordinary, _ = c.MeasureText(0, span)
		fractional, _ = a.reviewTextPolicy().measure(c, "addition", span)
		header, _ = a.reviewTextPolicy().measure(c, "hunk_header", span)
		width = a.reviewUnwrappedWidth(c)
		ui.Column(c).Width(254).Children(func() {
			a.reviewCodeRow(c, line, 7, false)
			ui.Text(c, line.Content).Label("ordinary neighbor").Font(a.Fonts.CodeFamily).FontSize(a.reviewCodeFontSize()).Width(178)
		})
	}, 254, 250)
	tt.Frame()
	if header != ordinary {
		t.Fatalf("non-code measure changed %g %g", header, ordinary)
	}
	if !a.reviewState.widthCache.fractional {
		t.Fatal("policy absent from width cache")
	}
	want := fractional + reviewCodeGutter + a.Theme.Space(12) + 10 + 2
	if math.Abs(float64(width-want)) > .01 {
		t.Fatalf("measure/render chrome mismatch %g %g", width, want)
	}
	r, ok := tt.Find(line.Content)
	if !ok || r.X != 66 || r.W != 178 {
		t.Fatalf("source geometry changed %+v", r)
	}
	originalWidth := width
	fallback := resolveReviewTextPolicy("linux", func() bool { return false })
	a.reviewState.textPolicy = &fallback
	tt.Frame()
	if a.reviewState.widthCache.fractional || header != ordinary || fractional != ordinary {
		t.Fatal("unsupported fallback retained opted-in measure/cache")
	}
	if ordinary > 0 && originalWidth > width+.01 {
		t.Fatal("unexpected rounded width smaller than fractional")
	}
	neighbor, ok := tt.Find("ordinary neighbor")
	if !ok || neighbor.W != 178 {
		t.Fatal("neighbor geometry changed")
	}
	t.Logf("native status enabled→%s; default %g, initial local width %g", fallback.status, ordinary, originalWidth)
}
func TestReviewTextHiddenCodeAndMetaMeasurements(t *testing.T) {
	a := fixture(t)
	var want float32
	hidden := workspace.ReviewLine{Kind: "context", Content: "hidden long context long context long context"}
	// Hidden context must use the same policy as expanded visible code.
	a.reviewState.data.Lines = []workspace.ReviewLine{{Kind: "meta", Content: "meta"}, {Kind: "gap", Gap: &workspace.ReviewGap{Hidden: []workspace.ReviewLine{hidden}}}}
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		w, _ := a.reviewTextPolicy().measure(c, hidden.Kind, ui.Span{Text: hidden.Content, Font: a.Fonts.CodeFamily, Size: a.reviewCodeFontSize()})
		want = w + reviewCodeGutter + a.Theme.Space(12) + 10 + 2
		if got := a.reviewUnwrappedWidth(c); math.Abs(float64(got-want)) > .01 {
			t.Errorf("code measure %g want %g", got, want)
		}
	}, 500, 100)
	tt.Frame()
}
