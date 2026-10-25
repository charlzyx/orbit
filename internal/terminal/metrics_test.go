package terminal

import (
	"math"
	"sync"
	"testing"

	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
)

func TestTerminalFontUsesSourcePitchNotNaturalLineHeight(t *testing.T) {
	for _, family := range []string{"monospace", "serif", "sans-serif"} {
		for _, size := range []float32{11, 13, 17, 20} {
			f := terminalFont(ghostty.Font{Family: family, Size: size, LineHeight: 8})
			m := (ui.Font{Family: f.Family, Size: f.Size, Weight: f.Weight}).Metrics()
			got := (m.Ascent + m.Descent + m.LineGap) * f.LineHeight
			if math.Abs(float64(got-size*1.4)) > .0001 {
				t.Fatalf("%s %g: pitch %g, expected %g", family, size, got, size*1.4)
			}
		}
	}
}

func TestLiveCustomRowPitchAndScale(t *testing.T) {
	requireTerminal(t)
	var mu sync.Mutex
	lastRows := 0
	p := NewCustom(CustomOptions{Font: ghostty.Font{Size: 12}, RowHeight: 18, OnResize: func(_ string, _, rows int) error {
		mu.Lock()
		lastRows = rows
		mu.Unlock()
		return nil
	}})
	defer p.Close()
	p.Update("metric", []string{"first row"}, 96, CustomRows, nil)
	until(t, nil, "custom metrics renderer", func() bool { return p.ActiveTerminal() != nil })
	tt := ui.NewTester(func(c *ui.Context) { CustomView(c, p).Width(640).Height(420) }, 660, 440)
	for _, pitch := range []float32{18, 19.384615, 25.5} {
		p.SetRowHeight(pitch)
		for _, scale := range []float32{1, 1.25, 1.5, 2} {
			tt.SetScale(scale)
			tt.Frame()
			until(t, tt, "custom actual device scale", func() bool {
				p.mu.Lock()
				defer p.mu.Unlock()
				return p.scale == scale
			})
			until(t, tt, "all custom rows fit", func() bool { mu.Lock(); defer mu.Unlock(); return lastRows == CustomRows })
		}
	}
}

func TestLiveSourcePitchReachesPTYGrid(t *testing.T) {
	requireTerminal(t)
	p := NewPanel(Options{Workspace: t.TempDir(), Command: []string{"/bin/sh"}, Env: []string{"PS1=", "ENV="}})
	defer p.Close()
	if e := p.NewTab(); e != nil {
		t.Fatal(e)
	}
	tt := ui.NewTester(func(c *ui.Context) { View(c, p) }, 640, 420)
	term := ready(t, p, tt)
	for _, family := range []string{"monospace", "serif"} {
		for _, size := range []float32{13, 17, 20} {
			p.SetFont(ghostty.Font{Family: family, Size: size})
			for _, scale := range []float32{1, 1.25, 1.5, 2} {
				tt.SetScale(scale)
				tt.Frame()
				_, rows := term.Size()
				pixels := int(math.Round(float64(420*scale))) - 2*int(math.Round(float64(4*scale)))
				cell := int(gridRowHeight(size*1.4, scale) * scale)
				if want := pixels / cell; rows != want {
					t.Fatalf("%s %g @%g: %d rows, expected %d (cell %d pixels)", family, size, scale, rows, want, cell)
				}
			}
		}
	}
}
