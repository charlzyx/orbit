package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
)

// Frozen theme/tokens.rs: Large=16, Small=12 (the view's 15/11 comments
// are stale). Compare actual native glyph bounds, including UI-font scaling.
func TestUsageActualSourceFontTokens(t *testing.T) {
	for _, size := range []float32{14, 21} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			a := fixtureUsageApp(t)
			a.Theme.FontSize = size
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Fill().Children(func() {
					a.usageCard(c, "Source title", "Source metadata", func() {})
					ui.Text(c, "Source title").FontSize(16 * size / 14).FontWeight(500).AlignSelf(ui.Start).Label("Large reference")
					ui.Text(c, "Source metadata").FontSize(12 * size / 14).FixedLineHeight(usageLineHeight(12 * size / 14)).AlignSelf(ui.Start).Label("Small reference")
				})
			}, 1000, 400)
			tt.Frame()
			title, _ := tt.Find("Source title")
			meta, _ := tt.Find("Source metadata")
			large, _ := tt.Find("Large reference")
			small, _ := tt.Find("Small reference")
			if title.W != large.W || meta.W != small.W || math.Abs(float64(meta.H-small.H)) > .0001 || title.H != 20*size/14 {
				t.Fatalf("actual source token mismatch size%v title%#v large%#v meta%#v small%#v", size, title, large, meta, small)
			}
		})
	}
}
