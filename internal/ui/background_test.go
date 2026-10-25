package orbitui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/background"
	"github.com/egoist/mygo/ui"
)

func TestNativeBackdropBoundsAndFeatureVisibility(t *testing.T) {
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	if a.backdrop.field == nil {
		t.Fatal("production mark not loaded")
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	node, ok := tt.Find("New task backdrop")
	if !ok {
		t.Fatal("empty task backdrop missing")
	}
	r := node
	if r.X != a.SidebarWidth+1 || r.Y != HeaderHeight || r.W != 1470-a.SidebarWidth-1 || r.H != 922-HeaderHeight {
		t.Fatalf("backdrop escaped main below-header bounds: %+v", r)
	}
	a.Navigate("settings")
	tt.Frame()
	if _, ok := tt.Find("New task backdrop"); ok {
		t.Fatal("backdrop painted on settings")
	}
	a.Navigate("chat")
	a.Model.Messages = []Message{{ID: "real", Role: "user", Text: "Task"}}
	tt.Frame()
	if _, ok := tt.Find("New task backdrop"); ok {
		t.Fatal("backdrop painted over transcript")
	}
}

func TestBackgroundChooseCoalesceFadeReuseFailureAndClear(t *testing.T) {
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	path := filepath.Join(t.TempDir(), "chosen.png")
	im := image.NewNRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			im.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 7), G: uint8(y * 13), A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, im)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	a.chooseBackground("")
	if a.backdrop.loading {
		t.Fatal("cancelled native picker started processing")
	}
	a.chooseBackground(path)
	for _, v := range []float32{.5, 1, 1.5} {
		a.setBackgroundTuning(background.Tuning{Blur: v, Cell: 6, Fade: .14})
	}
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading && a.backdrop.pending == nil })
	if a.backdrop.err != "" || a.backdrop.bitmap == nil || a.backdrop.config.Source == nil || a.backdrop.config.Blur != 1.5 || a.backdrop.config.Cell != 6 {
		t.Fatalf("photo result: %+v", a.backdrop.config)
	}
	old := a.backdrop.bitmap
	next := a.backdrop.config.Tuning
	next.Fade = .32
	a.setBackgroundTuning(next)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	if a.backdrop.bitmap != old {
		t.Fatal("fade-only change reallocated image bitmap")
	}
	a.chooseBackground(filepath.Join(t.TempDir(), "missing.png"))
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	if a.backdrop.err == "" || a.backdrop.bitmap != old {
		t.Fatal("failed replacement lost last usable picture or error")
	}
	a.clearBackground()
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	if a.backdrop.bitmap != nil || a.backdrop.config.Source != nil {
		t.Fatal("clear retained photo")
	}
	if _, err = os.Stat(filepath.Join(a.Options.Store.RootDir, "new-task-background.png")); !os.IsNotExist(err) {
		t.Fatal("persisted background not removed", err)
	}
}

func TestBackdropPaintsBehindOpaqueComposer(t *testing.T) {
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	r, ok := tt.Find("Message")
	if !ok {
		t.Fatal("composer missing")
	}
	pixels := tt.Image()
	want := a.Theme.Surface
	for y := int(r.Y + 5); y < int(r.Y+20); y++ {
		for x := int(r.X + r.W - 100); x < int(r.X+r.W-20); x++ {
			red, green, blue, alpha := pixels.At(x, y).RGBA()
			if uint8(red>>8) != want.R || uint8(green>>8) != want.G || uint8(blue>>8) != want.B || alpha != 65535 {
				t.Fatalf("backdrop painted over opaque composer at%d,%d: %d/%d/%d", x, y, red>>8, green>>8, blue>>8)
			}
		}
	}
}
