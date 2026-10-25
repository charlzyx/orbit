package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// Rust sidepane.rs:1750–1854: a 46px directional gutter beside the label.
// The label always reveals both edges, while Shift on either target reveals
// all retained context. Abbreviated patches with no retained lines stay inert.
func (a *App) reviewGap(c *ui.Context, line workspace.ReviewLine, index int) {
	gap := line.Gap
	if gap == nil {
		return
	}
	t := a.Theme
	directions := []string{"both"}
	switch gap.Position {
	case "leading":
		directions = []string{"end"}
	case "trailing":
		directions = []string{"start"}
	default:
		if gap.Count > workspace.DefaultExpansionLines {
			directions = []string{"start", "end"}
		}
	}
	activate := func(button ui.Element, direction string) {
		if button.Clicked() {
			if button.ClickModifiers()&ui.Shift != 0 {
				direction = "all"
			}
			a.expandReviewGap(line, index, direction)
			c.AnimationFrame()
		}
	}
	ui.Row(c).Label(fmt.Sprintf("Review context gap %d", gap.ID)).Height(32).FillWidth().Background(t.Surface).Children(func() {
		ui.Column(c).Width(46).Shrink(0).FillHeight().BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
			if !gap.Expandable() {
				return
			}
			for i, direction := range directions {
				b := ui.ButtonBase(c.Key(direction)).Label("Expand context " + direction).Tooltip("Expand context (Shift: all)").Grow(1).FillWidth().MinHeight(0).Center()
				if len(directions) > 1 && i == 0 {
					b.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
				}
				if b.Hovered() {
					b.Background(t.Hover)
				}
				icon := "icons/chevron-down.svg"
				if direction == "end" {
					icon = "icons/chevron-up.svg"
				}
				b.Children(func() { a.icon(c, icon, t.Px(12), t.TextMuted) })
				activate(b, direction)
			}
		})
		label := fmt.Sprintf("Show %d unchanged lines", gap.Count)
		b := ui.ButtonBase(c.Key("gap-label")).Label(label).Grow(1).MinWidth(0).FillHeight().PaddingX(t.Space(12)).AlignItems(ui.Center).Disabled(!gap.Expandable()).Clip()
		if b.Hovered() && gap.Expandable() {
			b.Background(t.Hover)
		}
		b.Children(func() { ui.Text(c, label).FontSize(t.Px(12)).TextColor(t.TextMuted).MaxLines(1) })
		activate(b, "both")
	})
}

func (a *App) expandReviewGap(line workspace.ReviewLine, index int, direction string) {
	s := &a.reviewState
	// Split rows do not carry an original index. Also revalidate the index if
	// an earlier event already changed the snapshot in this frame.
	if index < 0 || index >= len(s.data.Lines) || s.data.Lines[index].Gap == nil || s.data.Lines[index].Gap.ID != line.Gap.ID || s.data.Lines[index].File != line.File {
		index = -1
		for i, x := range s.data.Lines {
			if x.Gap != nil && x.Gap.ID == line.Gap.ID && x.File == line.File {
				index = i
				break
			}
		}
	}
	s.data.ExpandGap(index, direction)
}
