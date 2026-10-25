package orbitui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func usageShellFixture(t *testing.T) (*App, *ui.Tester) {
	t.Helper()
	a := fixture(t)
	a.Options.Pages = map[string]func(*ui.Context){"usage": func(c *ui.Context) {
		ui.Column(c).Fill().Label("Usage test body").Children(func() {
			ui.Column(c).FillWidth().MaxWidth(1180).AlignSelf(ui.Center).Padding(20).Children(func() {
				ui.Box(c).FillWidth().Height(20).Label("Usage test content")
			})
		})
	}}
	a.Navigate("usage")
	tt := ui.NewTester(a.View, 1470, 922)
	return a, tt
}

func TestNativeUsageSourceShellGeometry(t *testing.T) {
	a, tt := usageShellFixture(t)
	pane, ok := tt.Find("Usage pane")
	if !ok || pane != (ui.Rect{X: 260, Y: 44, W: 1210, H: 878}) {
		t.Fatal("source leading12 card below shared44 header", pane, ok)
	}
	body, ok := tt.Find("Usage test body")
	if !ok || body != (ui.Rect{X: 261, Y: 45, W: 1208, H: 876}) {
		t.Fatal("source hairline belongs to card, not page padding", body, ok)
	}
	content, ok := tt.Find("Usage test content")
	if !ok || content.X != 295 || content.W != 1140 {
		t.Fatal("centered1180 content plus20 inner padding", content, ok)
	}
	back, ok := tt.Find("Back to task")
	if !ok || back.X != 268 || back.Y != 8 || back.H != 28 {
		t.Fatal("source medium Back at shared top-bar inset", back, ok)
	}
	if !tt.HasText("Back") {
		t.Fatal("source Back label missing")
	}
	drag, ok := tt.Find("Usage window drag region")
	if !ok || drag.X <= back.X+back.W || drag.H != 44 {
		t.Fatal("only space after interactive title should drag", drag, ok)
	}
	a.SidebarVisible = false
	tt.Frame()
	pane, _ = tt.Find("Usage pane")
	if pane.X != 12 || pane.W != 1458 {
		t.Fatal("collapsed sidebar should not retain old page inset", pane)
	}
	tt.SetSize(960, 640)
	pane, _ = tt.Find("Usage pane")
	if pane != (ui.Rect{X: 12, Y: 44, W: 948, H: 596}) {
		t.Fatal("minimum viewport card", pane)
	}
}

func TestNativeUsageBackClosesOnPointerDownOnce(t *testing.T) {
	a, tt := usageShellFixture(t)
	var pages []string
	a.Options.OnPage = func(page string) { pages = append(pages, page) }
	back, _ := tt.Find("Back to task")
	x, y := back.X+back.W/2, back.Y+back.H/2
	tt.Press(x, y)
	if a.page != "chat" || len(pages) != 1 || pages[0] != "chat" {
		t.Fatal("source on_mouse_down close did not occur", a.page, pages)
	}
	tt.Release(x, y)
	if len(pages) != 1 {
		t.Fatal("release duplicated close action", pages)
	}
	if _, ok := tt.Find("Usage pane"); ok {
		t.Fatal("Usage card remains after close")
	}
}

func TestNativeUsageShellKeepsDIPGeometryAcrossScale(t *testing.T) {
	a, tt := usageShellFixture(t)
	base, _ := tt.Find("Usage pane")
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		tt.SetScale(scale)
		pane, _ := tt.Find("Usage pane")
		if pane != base {
			t.Fatal("device scale changed logical page bounds", scale, pane, base)
		}
	}
	tt.SetScale(1)
	a.Prefs.UIFontSize, a.Prefs.SpacingDensity = 18, 120
	tt.Frame()
	pane, _ := tt.Find("Usage pane")
	if pane.X != a.SidebarWidth+a.Theme.Space(12) || pane.Y != 44 || pane.X+pane.W != 1470 {
		t.Fatal("source font/density inset or fixed header was ignored", pane, a.Theme.Space(12))
	}
	back, _ := tt.Find("Back to task")
	if back.H != a.Theme.Px(28) || back.Y < 0 || back.Y+back.H > 44 {
		t.Fatal("scaled Back control clipped", back)
	}
}

func TestNativeUsageShellCapture(t *testing.T) {
	out := os.Getenv("ORBIT_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_CAPTURE_DIR for explicit CPU fixture")
	}
	a := fixtureUsageApp(t)
	a.Navigate("usage")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "usage-shell-fixture.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}
