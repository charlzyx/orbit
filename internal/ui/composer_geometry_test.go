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
