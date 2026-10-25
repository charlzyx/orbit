package gitworkbench

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func TestDiffColorSpansPreserveEveryByteAndFontMetrics(t *testing.T) {
	line := workspace.ReviewLine{Content: "let π = 42; // 世界😀"}
	line.Tokens, _ = highlight.TokenizeLine(highlight.Rust, line.Content, highlight.Carry{})
	color := ui.RGB(0x12, 0x34, 0x56)
	state := &State{Options: Options{TokenColor: func(highlight.Class) ui.Color { return color }}}
	first := state.diffSpans(nil, line)
	var rebuilt strings.Builder
	for _, span := range first {
		rebuilt.WriteString(span.Text)
		paintOnly := span
		paintOnly.Text = ""
		paintOnly.Color = ui.Color{}
		if !reflect.DeepEqual(paintOnly, ui.Span{}) {
			t.Fatal("syntax changed text metrics", span)
		}
	}
	if rebuilt.String() != line.Content {
		t.Fatal("syntax changed content", rebuilt.String())
	}
	color = ui.RGB(0x65, 0x43, 0x21)
	second := state.diffSpans(nil, line)
	if first[0].Text != second[0].Text || first[0].Color == second[0].Color {
		t.Fatal("palette must repaint without retokenizing", first, second)
	}
}
