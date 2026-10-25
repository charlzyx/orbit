package gitworkbench

import (
	"fmt"

	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func (s *State) diffSnapshot(staged bool) *workspace.ReviewSnapshot {
	if staged {
		return &s.Data.StagedReview
	}
	return &s.Data.UnstagedReview
}
func (s *State) fileDiff(staged bool, path string) []changeItem {
	snapshot := s.diffSnapshot(staged)
	file := -1
	for i, f := range snapshot.Files {
		if f.Path == path {
			file = i
			break
		}
	}
	if file < 0 {
		return nil
	}
	var lines []workspace.ReviewLine
	for _, line := range snapshot.Lines {
		if line.File == file && line.Kind != "file_header" {
			lines = append(lines, line)
		}
	}
	var result []changeItem
	if s.DiffSplit {
		for i, row := range workspace.SplitReview(lines) {
			r := row
			result = append(result, changeItem{staged: staged, path: path, diffIndex: i + 1, split: &r})
		}
	} else {
		for i, line := range lines {
			l := line
			result = append(result, changeItem{staged: staged, path: path, diffIndex: i + 1, line: &l})
		}
	}
	return result
}
func (s *State) embeddedDiffLine(c *ui.Context, item changeItem) {
	if item.split != nil {
		row := item.split
		if row.Full != nil {
			s.paintDiffLine(c, row.Full, item.staged)
			return
		}
		ui.Row(c).AlignSelf(ui.Stretch).Children(func() {
			ui.Column(c).WidthPercent(50).MinWidth(0).Children(func() { s.paintDiffLine(c, row.Left, item.staged) })
			ui.Column(c).WidthPercent(50).MinWidth(0).Children(func() { s.paintDiffLine(c, row.Right, item.staged) })
		})
		return
	}
	s.paintDiffLine(c, item.line, item.staged)
}
func (s *State) paintDiffLine(c *ui.Context, line *workspace.ReviewLine, staged bool) {
	if line == nil {
		ui.Box(c).Height(20)
		return
	}
	if line.Kind == "gap" {
		s.embeddedGap(c, *line, staged)
		return
	}

	bg, color, prefix := c.Theme().Background, c.Theme().Text, " "
	switch line.Kind {
	case "addition":
		bg = c.Theme().Success.Alpha(.1)
		prefix = "+"
	case "deletion":
		bg = c.Theme().Danger.Alpha(.1)
		prefix = "−"
	case "meta", "hunk_header":
		bg = c.Theme().Surface
		color = c.Theme().TextMuted
	}
	ui.Row(c).MinHeight(20).AlignSelf(ui.Stretch).Background(bg).PaddingX(10).Gap(6).Children(func() {
		old, newLine := "", ""
		if line.OldLine > 0 {
			old = fmt.Sprint(line.OldLine)
		}
		if line.NewLine > 0 {
			newLine = fmt.Sprint(line.NewLine)
		}
		muted(c, old).Width(35).Font("monospace")
		muted(c, newLine).Width(35).Font("monospace")
		ui.Text(c, prefix).Width(10).Font("monospace").TextColor(color)
		text := ui.RichText(c, s.diffSpans(c, *line)...).Grow(1).MinWidth(0).Font("monospace").FontSize(12).TextColor(color).Selectable()
		if !s.DiffWrap {
			text.NoWrap()
		}
	})
}

// Only Color is overridden. Font metrics, wrapping and selection are still the
// exact settings of the original text element; spans concatenate losslessly.
func (s *State) diffSpans(c *ui.Context, line workspace.ReviewLine) []ui.Span {
	runs := highlight.Runs(line.Content, line.Tokens)
	spans := make([]ui.Span, 0, len(runs))
	for _, run := range runs {
		span := ui.Span{Text: run.Text}
		if run.Class != "" {
			if s.Options.TokenColor != nil {
				span.Color = s.Options.TokenColor(run.Class)
			} else {
				switch run.Class {
				case highlight.Comment:
					span.Color = c.Theme().TextMuted
				case highlight.Keyword, highlight.Command:
					span.Color = c.Theme().Accent
				case highlight.String, highlight.Added:
					span.Color = c.Theme().Success
				case highlight.Removed:
					span.Color = c.Theme().Danger
				default:
					span.Color = c.Theme().Text
				}
			}
		}
		spans = append(spans, span)
	}
	return spans
}
