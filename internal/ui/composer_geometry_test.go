package orbitui

import (
	"github.com/egoist/mygo/ui"
	"math"
	"strings"
	"testing"
)

func TestSourceComposerGeometryAndConditionalChips(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{Workspace: "/workspace/example", ModelID: "fixture"})
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	input, ok := tt.Find("Message")
	if !ok {
		t.Fatal("composer missing")
	}
	if math.Abs(float64(input.H-a.Theme.FontSize*sourceDefaultLineHeight)) > 1 {
		t.Fatal("source editor should initially use one inherited line", input)
	}
	if _, ok = tt.Find("Thinking"); ok {
		t.Fatal("unknown or one-level thinking chip must not be shown")
	}
	if access, ok := tt.Find("Full access"); !ok || access.W > 30 {
		t.Fatal("default access chip should paint glyph only")
	}
	if tt.HasText("↵ Send") {
		t.Fatal("idle send hint should not appear in source footer")
	}
	for _, entry := range []struct {
		label string
		h     float32
	}{{"Add attachment", 22}, {"Full access", 22}, {"Build", 22}, {"fixture", 22}, {"Send message", 28}} {
		r, ok := tt.Find(entry.label)
		if !ok || r.H != entry.h {
			t.Fatal("source control frame", entry, r)
		}
	}
	box, _ := tt.Find("Composer box")
	if box.W != 960 || box.H > 100 {
		t.Fatal("source composer width/height", box)
	}
	a.Draft = strings.Repeat("line\n", 12)
	tt.Frame()
	input, _ = tt.Find("Message")
	if math.Abs(float64(input.H-8*a.Theme.FontSize*sourceDefaultLineHeight)) > 1 {
		t.Fatal("source editor should cap eight lines", input)
	}
	a.Draft = ""
	a.SetSnapshot(Snapshot{Workspace: "/workspace/example", ModelID: "fixture", Thinking: "high", ThinkingLevels: []string{"off", "high"}})
	tt.Frame()
	if _, ok = tt.Find("High"); !ok {
		t.Fatal("available thinking ladder missing")
	}
}
func TestSourceComposerCompactKeepsSendReachable(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{ModelID: strings.Repeat("long-model", 12), Thinking: "high", ThinkingLevels: []string{"off", "high"}})
	tt := ui.NewTester(a.View, 800, 560)
	for i := 0; i < 3; i++ {
		tt.Frame()
	}
	if _, ok := tt.Find("Full access"); ok {
		t.Fatal("compact composer kept access chip")
	}
	if _, ok := tt.Find("Build"); ok {
		t.Fatal("compact composer kept workflow chip")
	}
	r, ok := tt.Find("Send message")
	if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 800 || r.Y+r.H > 560 {
		t.Fatal("send clipped", r)
	}
}

func TestSourceComposerContactShadowPaintsOutsideOpaqueBox(t *testing.T) {
	for _, dark := range []bool{true, false} {
		t.Run(map[bool]string{true: "dark", false: "light"}[dark], func(t *testing.T) {
			a := fixture(t)
			drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
			a.backdrop.field = nil // Isolate the shadow from the animated field.
			if !dark {
				a.Appearance.Mode = "light"
				a.Theme = NewTheme("orbit-azure-light", a.Prefs, a.Fonts)
			}
			alpha := float32(.05)
			if !dark {
				alpha = .06
			}
			if a.Theme.overlay() != a.Theme.Text.Alpha(alpha) {
				t.Fatal("quiet send overlay differs from source theme wash")
			}
			tt := ui.NewTester(a.View, 1470, 922)
			tt.SetDark(dark)
			tt.Frame()
			box, ok := tt.Find("Composer box")
			if !ok {
				t.Fatal("composer missing")
			}
			pixels, x := tt.Image(), int(box.X+box.W/2)
			shadow := false
			for y := int(math.Ceil(float64(box.Y + box.H))); y < int(box.Y+box.H)+8; y++ {
				r, g, b, alpha := pixels.At(x, y).RGBA()
				bg := a.Theme.Background
				if r>>8 < uint32(bg.R) && g>>8 < uint32(bg.G) && b>>8 < uint32(bg.B) && alpha == 65535 {
					shadow = true
				}
			}
			if !shadow {
				t.Fatal("source contact shadow missing below composer", box)
			}
			r, g, b, _ := pixels.At(x, int(box.Y+10)).RGBA()
			want := a.Theme.SurfaceSecondary
			if r>>8 != uint32(want.R) || g>>8 != uint32(want.G) || b>>8 != uint32(want.B) {
				t.Fatal("shadow changed the opaque composer fill")
			}
		})
	}
}
