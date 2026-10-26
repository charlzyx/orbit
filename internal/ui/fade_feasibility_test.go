package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// Test-only source fade mechanism; no production call site or toolkit mutation.
func testSourceFade(c *ui.Context, color ui.Color) ui.Element {
	var fade ui.Element
	ui.Box(c.Key("fade-clip")).Absolute().Left(0).Right(0).Top(0).Bottom(0).Clip().PassThrough().Children(func() {
		fade = ui.Box(c.Key("source-fade")).Label("Source fade fixture").Absolute().Left(0).Right(0).Bottom(0).Height(48).Gradient(color.Alpha(0), color, 180).PassThrough()
	})
	return fade
}
func TestFadeOwnWrapperClipsFixed48WithoutRescaling(t *testing.T) {
	for _, height := range []float32{24, 80} {
		t.Run(fmt.Sprint(height), func(t *testing.T) {
			var host, rawFade ui.Rect
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Fill().Background(ui.RGB(255, 0, 0))
				box := ui.Box(c.Key("viewport")).Absolute().Left(20).Top(60).Width(200).Height(height).Background(ui.RGB(255, 0, 0))
				box.Children(func() { rawFade = testSourceFade(c, ui.RGB(255, 255, 255)).Bounds() })
				host = box.Bounds()
			}, 260, 180)
			tt.Frame()
			tt.Frame()
			f, ok := tt.Find("Source fade fixture")
			if !ok || rawFade.H != 48 || rawFade.Y+rawFade.H != host.Y+host.H || f.H != min(height, float32(48)) {
				t.Fatalf("fixed fade bounds raw%+v visible%+v host%+v", rawFade, f, host)
			}
			im := tt.Image()
			x := int(host.X + 20)
			above := im.RGBAAt(x, int(host.Y)-2)
			below := im.RGBAAt(x, int(host.Y+host.H)+2)
			if above.G != 0 || below.G != 0 {
				t.Fatalf("fade escaped own viewport: above=%v below=%v", above, below)
			}
			top := im.RGBAAt(x, int(host.Y)+1)
			bottom := im.RGBAAt(x, int(host.Y+host.H)-2)
			if height == 24 && (top.G < 120 || top.G > 150) {
				t.Fatalf("tiny crop was rescaled instead of lower half: %v", top)
			}
			if height == 80 && top.G != 0 {
				t.Fatalf("fade extended beyond bottom48: %v", top)
			}
			if bottom.G < 235 {
				t.Fatalf("bottom not opaque enough: %v", bottom)
			}
			// The gradient axis is vertical: two x positions at same y agree exactly.
			if im.RGBAAt(x, int(host.Y+host.H)-12) != im.RGBAAt(x+100, int(host.Y+host.H)-12) {
				t.Fatal("horizontal gradient instead of vertical")
			}
			if dir := os.Getenv("ORBIT_FADE_CAPTURE_DIR"); dir != "" {
				if e := os.MkdirAll(dir, 0755); e != nil {
					t.Fatal(e)
				}
				f, e := os.Create(filepath.Join(dir, fmt.Sprintf("clip-%g.png", height)))
				if e != nil {
					t.Fatal(e)
				}
				defer f.Close()
				if e = png.Encode(f, im); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}
