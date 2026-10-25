package gitworkbench

import (
	"fmt"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"sync"
)

var diffGapIconsOnce sync.Once
var diffGapIcons map[string]*ui.SVG

func diffGapIcon(c *ui.Context, name string) {
	diffGapIconsOnce.Do(func() {
		diffGapIcons = map[string]*ui.SVG{}
		for _, name := range []string{"chevron-up", "chevron-down"} {
			if data, err := orbitassets.Asset("icons/" + name + ".svg"); err == nil {
				if svg, err := ui.ParseSVG(data); err == nil {
					diffGapIcons[name] = svg
				}
			}
		}
	})
	if svg := diffGapIcons[name]; svg != nil {
		ui.Icon(c, svg).Size(12, 12).TextColor(c.Theme().TextMuted)
	}
}

// Rust sidepane.rs:1750–1854: edge-dependent gutter, both-edge label,
// Shift reveals all retained context. Abbreviated patches remain inert.
func (s *State) embeddedGap(c *ui.Context, line workspace.ReviewLine, staged bool) {
	gap := line.Gap
	if gap == nil {
		return
	}
	t := c.Theme()
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
	activate := func(b ui.Element, direction string) {
		if !b.Clicked() {
			return
		}
		if b.ClickModifiers()&ui.Shift != 0 {
			direction = "all"
		}
		snapshot := s.diffSnapshot(staged)
		for i, row := range snapshot.Lines {
			if row.File == line.File && row.Gap != nil && row.Gap.ID == gap.ID {
				snapshot.ExpandGap(i, direction)
				c.AnimationFrame()
				break
			}
		}
	}
	snapshot := s.diffSnapshot(staged)
	key := fmt.Sprintf("diff-gap:%d:%p:%t:%d", s.generation, firstReviewLine(snapshot.Lines), staged, gap.ID)
	ui.Row(c.Key(key)).Label(fmt.Sprintf("Git context gap %t:%d", staged, gap.ID)).Height(32).FillWidth().Background(t.Surface).Children(func() {
		ui.Column(c).Label("Git context gutter").Width(46).Shrink(0).FillHeight().BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
			if !gap.Expandable() {
				return
			}
			for i, direction := range directions {
				b := ui.ButtonBase(c.Key(direction)).Label("Expand context " + direction).Tooltip("Expand context (Shift: all)").Grow(1).FillWidth().MinHeight(0).Center()
				if len(directions) > 1 && i == 0 {
					b.Height(16).MinHeight(16).MaxHeight(16).Grow(0).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
				}
				if b.Hovered() {
					b.Background(t.SurfaceHover)
				}
				icon := "chevron-down"
				if direction == "end" {
					icon = "chevron-up"
				}
				b.Children(func() { diffGapIcon(c, icon) })
				activate(b, direction)
			}
		})
		label := fmt.Sprintf("Show %d unchanged lines", gap.Count)
		b := ui.ButtonBase(c.Key("gap-label")).Label(label).Grow(1).MinWidth(0).FillHeight().PaddingX(12).AlignItems(ui.Center).Disabled(!gap.Expandable()).Clip()
		if b.Hovered() && gap.Expandable() {
			b.Background(t.SurfaceHover)
		}
		b.Children(func() { ui.Text(c, label).FontSize(12).TextColor(t.TextMuted).MaxLines(1) })
		activate(b, "both")
	})
}
