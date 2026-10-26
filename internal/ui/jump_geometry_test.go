package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func jumpGeometryFixture(t *testing.T, mode string) (*App, *ui.Tester, *ui.Rect, Snapshot) {
	t.Helper()
	a := fixture(t)
	a.Appearance.Mode = mode
	s := Snapshot{RuntimeID: "jump-r", SessionID: "jump-s"}
	for i := 0; i < 120; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Geometry row %03d", i)})
	}
	a.SetSnapshot(s)
	bounds := new(ui.Rect)
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); *bounds = a.transcript.Handle.Bounds(c) }, 1200, 800)
	tt.Frame()
	for i := 0; i < 6; i++ {
		tt.Scroll(700, 400, 0, -30)
	}
	tt.Frame()
	return a, tt, bounds, s
}
func TestJumpGeometryLightDarkAndUnread(t *testing.T) {
	for _, mode := range []string{"light", "dark"} {
		t.Run(mode, func(t *testing.T) {
			a, tt, viewport, s := jumpGeometryFixture(t, mode)
			h := viewport.H
			capture := func(state string) {
				if dir := os.Getenv("ORBIT_JUMP_CAPTURE_DIR"); dir != "" {
					if e := os.MkdirAll(dir, 0755); e != nil {
						t.Fatal(e)
					}
					f, e := os.Create(filepath.Join(dir, mode+"-"+state+".png"))
					if e != nil {
						t.Fatal(e)
					}
					defer f.Close()
					if e = png.Encode(f, tt.Image()); e != nil {
						t.Fatal(e)
					}
				}
			}
			verify := func(unread bool) {
				b, ok := tt.Find("Jump to latest")
				if !ok {
					t.Fatal("accessible jump label missing")
				}
				if b.H != 32 || (!unread && b.W != 32) || (unread && b.W <= 32) {
					t.Fatalf("bad geometry %+v unread=%t", b, unread)
				}
				if math.Abs(float64(b.X+b.W/2-(viewport.X+viewport.W/2))) > .01 || math.Abs(float64(viewport.Y+viewport.H-(b.Y+b.H)-8)) > .01 {
					t.Fatalf("not centered/bottom8 b=%+v viewport=%+v", b, *viewport)
				}
				if viewport.H != h {
					t.Fatal("unread resized List")
				}
			}
			verify(false)
			capture("idle")
			s.Messages = append([]Message(nil), s.Messages...)
			s.Messages = append(s.Messages, Message{ID: "new", Role: "assistant", Text: "New activity visible"})
			a.SetSnapshot(s)
			tt.Frame()
			tt.Frame()
			verify(true)
			capture("unread")
			if e := tt.Click("Jump to latest"); e != nil {
				t.Fatal(e)
			}
			tt.Frame()
			if !a.transcript.AtEnd() || a.transcriptActivity.unread {
				t.Fatal("jump failed")
			}
			if viewport.H != h {
				t.Fatal("jump removal resized viewport")
			}
		})
	}
}
func TestJumpGeometryEmptyLayerPassesWheelSelectionAndMenu(t *testing.T) {
	a, tt, viewport, s := jumpGeometryFixture(t, "light")
	b, ok := tt.Find("Jump to latest")
	if !ok {
		t.Fatal("jump absent")
	}
	before := a.transcriptScroll.Y
	tt.Scroll(viewport.X+10, b.Y+b.H/2, 0, -30)
	tt.Frame()
	if a.transcriptScroll.Y >= before {
		t.Fatal("blank jump layer swallowed wheel")
	}
	tt.RightClickAt(viewport.X+10, b.Y+b.H/2)
	if len(tt.Menu()) != 0 {
		t.Fatalf("blank layer opened stale message menu %v", tt.Menu())
	}
	// Locate actual rendered text crossing the overlay strip, away from its
	// centered button; only public measured bounds drive this fixture.
	var r ui.Rect
	var text string
	var y float32
	for tries := 0; tries < 20 && text == ""; tries++ {
		b, _ = tt.Find("Jump to latest")
		for _, m := range s.Messages {
			q, found := tt.Find(m.Text)
			if !found {
				continue
			}
			lo := max(q.Y+14, b.Y+2)
			hi := min(q.Y+q.H-14, b.Y+b.H-2)
			if lo < hi && q.X > b.X+b.W+8 {
				r, text, y = q, m.Text, (lo+hi)/2
				break
			}
		}
		if text == "" {
			tt.Scroll(viewport.X+10, b.Y+16, 0, -5)
			tt.Frame()
		}
	}
	if text == "" {
		t.Fatal("no real text/overlay intersection fixture")
	}
	tt.RightClickAt(r.X+20, y)
	if e := tt.ChooseMenuItem("Copy Message"); e != nil {
		t.Fatal(e)
	}
	if tt.Clipboard() != text {
		t.Fatalf("overlay retargeted context menu %q want %q", tt.Clipboard(), text)
	}
	tt.Press(r.X+17, y)
	tt.Move(r.X+r.W-17, y)
	tt.Release(r.X+r.W-17, y)
	tt.Key(ui.Cmd, ui.KeyC)
	if got := tt.Clipboard(); got == "" || !strings.Contains(got, text) {
		t.Fatalf("overlay selection omitted its first measured row: %q missing %q", got, text)
	}
	selected := strings.Split(tt.Clipboard(), "\n")
	start := -1
	for i, m := range s.Messages {
		if m.Text == text {
			start = i
			break
		}
	}
	if start < 0 || start+len(selected) > len(s.Messages) {
		t.Fatal("selection outside fixture messages")
	}
	for i, line := range selected {
		if line != s.Messages[start+i].Text {
			t.Fatalf("non-contiguous selected text at %d: %q", i, line)
		}
	}

}
func TestJumpGeometryFocusedButtonKeyboardActivation(t *testing.T) {
	a, tt, viewport, _ := jumpGeometryFixture(t, "dark")
	b, ok := tt.Find("Jump to latest")
	if !ok {
		t.Fatal("button absent")
	}
	tt.Press(b.X+b.W/2, b.Y+b.H/2)
	tt.Move(viewport.X+10, b.Y+b.H/2)
	tt.Release(viewport.X+10, b.Y+b.H/2)
	if !tt.Focused("Jump to latest") || a.transcript.AtEnd() {
		t.Fatal("cancelled pointer press did not retain focus without activation")
	}
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	if !a.transcript.AtEnd() {
		t.Fatal("focused button Space did not activate jump")
	}
}
