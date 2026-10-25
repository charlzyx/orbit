package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// diff_view.rs:20,253–323,386–464. The number gutter is fixed at round(12.5*3+14),
// independent of editor font scaling. There is no separate +/- text column.
const reviewCodeGutter float32 = 52

func (a *App) reviewCodeFontSize() float32 { return 12.5 * a.Theme.CodeSize / 13 }

func (a *App) reviewCodeRow(c *ui.Context, l workspace.ReviewLine, number int, split bool) ui.Element {
	t := a.Theme
	fontSize, lineHeight := a.reviewCodeFontSize(), 16*a.Theme.CodeSize/13
	bodyBG, gutterBG, numberColor := t.Background, t.Background, t.TextMuted
	var edge ui.Color
	changed := l.Kind == "addition" || l.Kind == "deletion"
	if changed {
		edge = t.Success
		if l.Kind == "deletion" {
			edge = t.Error
		}
		bodyWash, gutterWash := float32(.12), float32(.09)
		if t.Dark {
			bodyWash, gutterWash = .20, .15
		}
		bodyBG, gutterBG, numberColor = edge.Alpha(bodyWash), edge.Alpha(gutterWash), edge
	}
	row := a.reviewTextPolicy().apply(ui.Row(c)).FillWidth().MinWidth(0).MinHeight(19).AlignItems(ui.Stretch)
	if changed && !split {
		row.BorderWidth(0, 0, 0, 2).BorderColor(edge)
	}
	row.Children(func() {
		gutter := ui.Row(c).Width(reviewCodeGutter).Shrink(0).MinHeight(19).Padding(0, 9, 0, 0).Justify(ui.End).Background(gutterBG)
		if !split {
			gutter.BorderWidth(0, 1, 0, 0).BorderColor(t.Border)
		}
		gutter.Children(func() {
			shown := ""
			if number > 0 {
				shown = fmt.Sprint(number)
			}
			ui.Text(c, shown).Font(a.Fonts.CodeFamily).FontSize(fontSize).FixedLineHeight(lineHeight).TextColor(numberColor).NoWrap()
		})
		right := float32(10)
		if split {
			right = t.Space(12)
		}
		ui.Column(c).Grow(1).MinWidth(0).MinHeight(19).Padding(0, right, 0, t.Space(12)).Background(bodyBG).Clip().Children(func() {
			text := ui.RichText(c, a.reviewSpans(l)...).Font(a.Fonts.CodeFamily).FontSize(fontSize).FixedLineHeight(lineHeight).TextColor(t.TextSecondary).FillWidth().MinWidth(0)
			if !a.reviewState.wrap {
				text.NoWrap()
			}
		})
	})
	return row
}
