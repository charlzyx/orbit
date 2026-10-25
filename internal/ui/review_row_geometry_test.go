package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
)

func TestMountedReviewCodeRowSourceGeometry(t *testing.T) {
	for _, kind := range []string{"context", "addition", "deletion"} {
		t.Run(kind, func(t *testing.T) {
			a := fixture(t)
			a.reviewState.wrap = true
			line := workspace.ReviewLine{Kind: kind, OldLine: 4, NewLine: 4, Content: "Untracked fixture note."}
			var sourceWidth, oldWidth float32
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(a.Theme.Native())
				sourceWidth, _ = a.reviewTextPolicy().measure(c, kind, ui.Span{Text: line.Content, Font: a.Fonts.CodeFamily, Size: 12.5 * a.Theme.CodeSize / 13})
				oldWidth, _ = c.MeasureText(0, ui.Span{Text: line.Content, Font: a.Fonts.CodeFamily, Size: a.Theme.CodeSize})
				ui.Column(c).Width(254).Height(200).Children(func() {
					a.reviewLineContent(c, line, 0)
					ui.Text(c, "M").Label("Source diff font probe").Font(a.Fonts.CodeFamily).FontSize(12.5 * a.Theme.CodeSize / 13).FixedLineHeight(16 * a.Theme.CodeSize / 13)
				})
			}, 254, 200)
			tt.Frame()
			text, ok := tt.Find(line.Content)
			if !ok {
				t.Fatal("real selectable diff text not mounted")
			}
			t.Logf("kind=%s text=%+v sourceFont=%g sourceTextWidth=%g currentFont=%g currentTextWidth=%g", kind, text, 12.5*a.Theme.CodeSize/13, sourceWidth, a.Theme.CodeSize, oldWidth)
			x := float32(64)
			if kind != "context" {
				x += 2
			}
			if math.Abs(float64(text.X-x)) > .5 || math.Abs(float64(text.W-(254-x-10))) > .5 {
				t.Fatalf("source single gutter geometry differs: got %+v want x=%g rightPadding=10", text, x)
			}
			probe, ok := tt.Find("Source diff font probe")
			if !ok {
				t.Fatal("native font probe missing")
			}
			wantHeight := probe.H
			if sourceWidth > text.W+.5 {
				wantHeight *= 2
			}
			if math.Abs(float64(text.H-wantHeight)) > .5 {
				t.Fatalf("explicit source line height differs: %+v want %g", text, wantHeight)
			}
			if _, ok := tt.Find("+"); ok {
				t.Fatal("extra sign column not present in Rust source")
			}
			if _, ok := tt.Find("−"); ok {
				t.Fatal("extra sign column not present in Rust source")
			}
		})
	}
}

func TestMountedReviewSplitNumbersUseTheirOwnSourceSide(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	s.data.Lines = []workspace.ReviewLine{{File: 0, Kind: "file_header", Content: "one.go"}, {File: 0, Kind: "context", OldLine: 123, NewLine: 456, Content: "shared context"}}
	s.data.Files = s.data.Files[:1]
	s.generation++
	if err := tt.Click("Side-by-side view"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	left, lok := tt.Find("123")
	right, rok := tt.Find("456")
	if !lok || !rok || left.X >= right.X {
		t.Fatal("split context did not show old/new side numbers", left, right, lok, rok)
	}
	viewport, _ := tt.Find("Review horizontal viewport")
	if left.X >= viewport.X+viewport.W/2 || right.X < viewport.X+viewport.W/2 {
		t.Fatal("number gutters not in their own half", left, right, viewport)
	}
	if err := tt.Click("Unified view"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("456"); !ok {
		t.Fatal("unified did not prefer new line number")
	}
	if _, ok := tt.Find("123"); ok {
		t.Fatal("unified retained old-side gutter")
	}
}
