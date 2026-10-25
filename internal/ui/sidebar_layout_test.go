package orbitui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestNativeSidebarSourcePointerResizeAndPersist(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	drag := func(dx float32) float32 {
		t.Helper()
		r, ok := tt.Find("Resize sidebar")
		if !ok || r.W != 6 || r.X != a.SidebarWidth-6 {
			t.Fatal("source inside-edge resize handle", r, ok, a.SidebarWidth)
		}
		x, y := r.X+r.W/2, r.Y+200
		tt.Press(x, y)
		tt.Move(x+dx, y)
		tt.Release(x+dx, y)
		tt.Frame()
		if a.Options.Store.Layout().SidebarWidth != float64(a.SidebarWidth) {
			t.Fatal("drag did not persist", a.SidebarWidth, a.Options.Store.Layout())
		}
		divider, ok := tt.Find("Divider")
		if !ok || divider.X != a.SidebarWidth-1 {
			t.Fatal("divider width counted twice", divider)
		}
		return x + dx
	}
	if pointer := drag(100); a.SidebarWidth != pointer {
		t.Fatal("source drag uses absolute pointer x", a.SidebarWidth, pointer)
	}
	drag(-2000)
	if a.SidebarWidth != 200 {
		t.Fatal("source minimum", a.SidebarWidth)
	}
	drag(3000)
	if a.SidebarWidth != 1070 {
		t.Fatal("source400px workspace reserve", a.SidebarWidth)
	}
	a.SidebarWidth = 400 // Begin the narrow-window gesture with its handle on screen.
	tt.SetSize(960, 640)
	drag(1000)
	if a.SidebarWidth != 560 {
		t.Fatal("source narrow-window ceiling", a.SidebarWidth)
	}
}
