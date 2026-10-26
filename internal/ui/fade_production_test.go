package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestTranscriptFadeProductionOrderBoundsAndCapture(t *testing.T) {
	for _, mode := range []string{"light", "dark"} {
		t.Run(mode, func(t *testing.T) {
			a, tt, viewport, _ := jumpGeometryFixture(t, mode)
			f, ok := tt.Find("Transcript bottom fade")
			if !ok || f.H != 48 || f.Y+f.H != viewport.Y+viewport.H || f.W != viewport.W {
				t.Fatalf("fade geometry %+v viewport%+v", f, *viewport)
			}
			b, ok := tt.Find("Jump to latest")
			if !ok {
				t.Fatal("jump missing")
			}
			// At an interior fill point away from arrow/border, Jump is painted after
			// fade and keeps its original opaque surface fill.
			p := tt.Image().RGBAAt(int(b.X+b.W/2+8), int(b.Y+b.H/2))
			want := a.Theme.SurfaceSecondary
			if p.R != want.R || p.G != want.G || p.B != want.B {
				t.Fatalf("fade painted over jump fill: pixel%v want%v", p, want)
			}
			if dir := os.Getenv("ORBIT_FADE_FINAL_CAPTURE_DIR"); dir != "" {
				if e := os.MkdirAll(dir, 0755); e != nil {
					t.Fatal(e)
				}
				out, e := os.Create(filepath.Join(dir, "production-fade-"+mode+".png"))
				if e != nil {
					t.Fatal(e)
				}
				if e = png.Encode(out, tt.Image()); e != nil {
					t.Fatal(e)
				}
				out.Close()
			}
			if e := tt.Click("Jump to latest"); e != nil {
				t.Fatal(e)
			}
			tt.Frame()
			if !a.transcript.AtEnd() {
				t.Fatal("fade prevented jump")
			}
			if _, ok := tt.Find("Transcript bottom fade"); ok {
				t.Fatal("fade still shown at tail")
			}
		})
	}
}
func TestTranscriptFadeProductionTinyViewportClip(t *testing.T) {
	a := fixture(t)
	s := Snapshot{SessionID: "tiny-fade"}
	for i := 0; i < 30; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Tiny %d", i)})
	}
	a.SetSnapshot(s)
	var viewport ui.Rect
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Box(c).Fill().Background(ui.RGB(255, 0, 0))
		ui.Column(c.Key("tiny-transcript")).Absolute().Left(20).Top(60).Width(240).Height(24).Children(func() { a.messages(c) })
		viewport = a.transcript.Handle.Bounds(c)
	}, 300, 150)
	tt.Frame()
	tt.Scroll(30, 70, 0, -40)
	tt.Frame()
	tt.Frame()
	if a.transcript.AtEnd() {
		t.Fatal("tiny actual transcript not scrolled")
	}
	f, ok := tt.Find("Transcript bottom fade")
	if !ok || f.H != 24 || f.Y != viewport.Y || f.Y+f.H != viewport.Y+viewport.H {
		t.Fatalf("tiny actual fade not clipped: f%+v viewport%+v", f, viewport)
	}
	im := tt.Image()
	for _, y := range []int{int(viewport.Y) - 2, int(viewport.Y+viewport.H) + 2} {
		p := im.RGBAAt(int(viewport.X)+4, y)
		if p.R != 255 || p.G != 0 || p.B != 0 {
			t.Fatalf("fade escaped at y%d: %v", y, p)
		}
	}
	if dir := os.Getenv("ORBIT_FADE_FINAL_CAPTURE_DIR"); dir != "" {
		out, e := os.Create(filepath.Join(dir, "production-fade-tiny.png"))
		if e != nil {
			t.Fatal(e)
		}
		defer out.Close()
		if e = png.Encode(out, im); e != nil {
			t.Fatal(e)
		}
	}
}
