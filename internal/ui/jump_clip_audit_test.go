package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAuditTinyTranscriptJumpDoesNotPaintOrClickAboveViewport(t *testing.T) {
	for _, mode := range []string{"light", "dark"} {
		t.Run(mode, func(t *testing.T) {
			a := fixture(t)
			id := a.Appearance.DarkTheme
			if mode == "light" {
				id = a.Appearance.LightTheme
			}
			a.Theme = NewTheme(id, a.Prefs, a.Fonts)
			s := Snapshot{SessionID: "jump-clip"}
			for i := 0; i < 30; i++ {
				s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Tiny row %d", i)})
			}
			a.SetSnapshot(s)
			var viewport ui.Rect
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Fill().Background(ui.RGB(255, 0, 0))
				ui.Column(c.Key("tiny-source-transcript")).Absolute().Left(20).Top(60).Width(240).Height(24).Children(func() { a.messages(c) })
				viewport = a.transcript.Handle.Bounds(c)
			}, 300, 150)
			tt.Frame()
			tt.Scroll(30, 70, 0, -40)
			tt.Frame()
			tt.Frame()
			if a.transcript.AtEnd() {
				t.Fatal("not away from tail")
			}
			b, ok := tt.Find("Jump to latest")
			if !ok {
				t.Fatal("jump absent")
			}
			if b.H != 16 || b.Y != viewport.Y {
				t.Fatalf("tiny visible button not clipped: %+v viewport %+v", b, viewport)
			}
			if dir := os.Getenv("ORBIT_JUMP_CLIP_CAPTURE_DIR"); dir != "" {
				if e := os.MkdirAll(dir, 0755); e != nil {
					t.Fatal(e)
				}
				f, e := os.Create(filepath.Join(dir, "tiny-"+mode+".png"))
				if e != nil {
					t.Fatal(e)
				}
				if e = png.Encode(f, tt.Image()); e != nil {
					t.Fatal(e)
				}
				f.Close()
			}
			x, y := viewport.X+viewport.W/2, viewport.Y-2
			p := tt.Image().RGBAAt(int(x), int(y))
			if p.R != 255 || p.G != 0 || p.B != 0 {
				t.Errorf("jump paints outside source viewport: %v bounds%+v viewport%+v", p, b, viewport)
			}

			tip := a.Tr("message_scroller.jump_to_latest")
			countTip := func() int {
				n := 0
				for _, v := range tt.Texts() {
					if v == tip {
						n++
					}
				}
				return n
			}
			// Tester.Texts includes the button accessible Label too; a real tooltip adds a second exact text.
			baseTip := countTip()
			if baseTip != 1 {
				t.Fatalf("unexpected pre-hover tooltip count %d", baseTip)
			}
			tt.Move(x, y)
			tt.Advance(time.Second)
			if countTip() != baseTip {
				t.Fatal("clipped-away pointer showed Jump tooltip")
			}
			tt.Move(b.X+b.W/2, b.Y+b.H/2)
			tt.Advance(time.Second)
			if countTip() != baseTip+1 {
				t.Fatal("visible clipped button lost tooltip", tt.Texts())
			}
			tt.Move(x, y)
			tt.Advance(time.Second)
			if countTip() != baseTip {
				t.Fatal("tooltip remained over clipped-away pointer")
			}
			tt.ClickAt(x, y)
			tt.Frame()
			if a.transcript.AtEnd() {
				t.Error("outside-viewport click activated Jump")
			}
			if e := tt.Click("Jump to latest"); e != nil {
				t.Fatal(e)
			}
			tt.Frame()
			if !a.transcript.AtEnd() {
				t.Fatal("visible clipped Jump cannot activate")
			}
		})
	}
}
func TestJumpOwnClipWrapperPublicMechanism(t *testing.T) {
	for _, height := range []float32{24, 80} {
		t.Run(fmt.Sprint(height), func(t *testing.T) {
			clicks := 0
			var host, rawButton ui.Rect
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Fill().Background(ui.RGB(255, 0, 0))
				box := ui.Column(c.Key("viewport")).Absolute().Left(20).Top(60).Width(240).Height(height)
				box.Children(func() {
					ui.Box(c.Key("jump-only-clip")).Absolute().Left(0).Right(0).Top(0).Bottom(0).Clip().PassThrough().Children(func() {
						ui.Row(c.Key("jump-layer")).Absolute().Left(0).Right(0).Bottom(8).Justify(ui.Center).PassThrough().Children(func() {
							b := ui.ButtonBase(c.Key("jump")).Label("Clipped Jump fixture").Size(32, 32).Radius(16).Background(ui.RGB(255, 255, 255))
							rawButton = b.Bounds()
							if b.Clicked() {
								clicks++
							}
						})
					})
				})
				host = box.Bounds()
			}, 300, 180)
			tt.Frame()
			tt.Frame()
			visible, ok := tt.Find("Clipped Jump fixture")
			if !ok {
				t.Fatal("visible jump absent")
			}
			if rawButton.H != 32 || rawButton.Y+rawButton.H != host.Y+host.H-8 {
				t.Fatalf("raw source geometry %+v host%+v", rawButton, host)
			}
			if visible.H != min(float32(32), height-8) || visible.Y < host.Y {
				t.Fatalf("bad clipped visible bounds %+v", visible)
			}
			tt.ClickAt(host.X+host.W/2, host.Y-2)
			if clicks != 0 {
				t.Fatal("clip allowed outside pointer")
			}
			if p := tt.Image().RGBAAt(int(host.X+host.W/2), int(host.Y)-2); p.R != 255 || p.G != 0 || p.B != 0 {
				t.Fatal("clip allowed outside paint", p)
			}
			tt.ClickAt(host.X+4, visible.Y+visible.H/2)
			if clicks != 0 {
				t.Fatal("blank clip wrapper intercepted activation")
			}
			if e := tt.Click("Clipped Jump fixture"); e != nil {
				t.Fatal(e)
			}
			if clicks != 1 {
				t.Fatalf("visible button must activate once: %d", clicks)
			}
		})
	}
}
