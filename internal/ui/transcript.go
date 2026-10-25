package orbitui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func (a *App) projectTranscript(s *Snapshot) {
	if s.RawMessages == nil {
		return
	} // Explicit native fixtures may supply UI rows.
	rows := a.transcriptUI.Observe(transcriptui.Project(s.RawMessages), s.Running, time.Now())
	s.Messages = make([]Message, 0, len(rows)+1)
	for i := range rows {
		r := &rows[i]
		s.Messages = append(s.Messages, Message{ID: r.ID, Role: r.Role, Text: r.Text(), Error: r.Error, Streaming: r.Streaming, Images: r.Images, Logical: r})
	}
	if !s.Running {
		if files := transcriptui.ChangedFilesOf(rows); len(files) > 0 {
			last := -1
			for i := len(rows) - 1; i >= 0; i-- {
				if rows[i].Role == "assistant" {
					last = i
					break
				}
			}
			if last >= 0 {
				s.Messages[last].SuppressFooter = true
				s.Messages = append(s.Messages, Message{ID: "changed-files-summary", Role: "changedFiles", ChangedFiles: files, Logical: &rows[last]})
			}
		}
	}
}
func messageSearchText(m Message) string {
	if m.Role == "changedFiles" {
		var paths []string
		for _, f := range m.ChangedFiles {
			paths = append(paths, f.Path)
		}
		return strings.Join(paths, "\n")
	}
	if m.Logical != nil {
		return m.Logical.SearchText()
	}
	return m.Text + "\n" + m.Thinking + "\n" + m.ToolOutput
}
func (a *App) transcriptOptions(c *ui.Context) transcriptui.Options {
	t := a.Theme
	latest := a.Model.LatestTurn
	colors := transcriptui.Colors{Text: t.Text, Muted: t.TextMuted, Accent: t.AccentText, Border: t.Border, Surface: t.Surface, Code: t.CodeBg, Error: t.Error, Warning: t.Warning, Success: t.Success, Selection: t.Selection, Syntax: map[highlight.Class]ui.Color{}}
	for _, class := range []highlight.Class{highlight.Keyword, highlight.Literal, highlight.String, highlight.Comment, highlight.Number, highlight.Type, highlight.Function, highlight.Meta, highlight.Command, highlight.Flag, highlight.Path, highlight.Operator, highlight.Added, highlight.Removed} {
		colors.Syntax[class] = t.TokenColor(class)
	}
	return transcriptui.Options{Colors: colors, Font: t.Font, CodeFont: t.CodeFont, FontSize: t.FontSize, CodeSize: t.CodeSize, Text: a.text, Workspace: a.Model.Workspace,
		Icon: func(c *ui.Context, path string, size float32, color ui.Color) { a.icon(c, path, size, color) },
		Markdown: func(c *ui.Context, text string, streaming bool) {
			if streaming {
				markdown.Stream(c, text, &a.md)
			} else {
				markdown.View(c, text, &a.md)
			}
		},
		Copy: func(text string) { c.Services().WriteClipboard(text) }, Review: func() { a.OpenTurnReview(latest) },
	}
}
func (a *App) logicalMessage(c *ui.Context, m Message) {
	o := a.transcriptOptions(c)
	if m.Role == "changedFiles" {
		if m.Logical != nil {
			transcriptui.Tail(c, &a.transcriptUI, m.ChangedFiles, *m.Logical, o)
		} else {
			transcriptui.Files(c, &a.transcriptUI, m.ChangedFiles, o)
		}
		return
	}
	if m.Logical == nil {
		return
	}
	o.SuppressFooter = m.SuppressFooter
	transcriptui.Turn(c, &a.transcriptUI, *m.Logical, o)
}
func (a *App) transcriptRowKey(i int) any {
	m := a.Model.Messages[i]
	if m.ID != "" {
		return m.ID
	}
	return fmt.Sprintf("message-%d", i)
}

func (a *App) emptySidebar(c *ui.Context) {
	t := a.Theme
	ui.Column(c.Key("empty-projects")).Grow(1).MinHeight(0).Center().Gap(t.Space(6)).Padding(0, t.Space(20), t.Space(40)).Children(func() {
		ui.Box(c).Size(36, 36).Radius(18).Background(t.Elevated).Center().Children(func() { a.icon(c, "icons/spark.svg", 16, t.AccentText) })
		ui.Text(c, a.text("sidebar.no_projects_yet", "No projects yet")).TextColor(t.TextSecondary).FontWeight(500)
		ui.Text(c, a.text("sidebar.pick_a_folder_to_start_your_first_task", "Pick a folder to start your first task")).TextColor(t.TextMuted).FontSize(t.Px(12)).TextAlign(ui.Center)
	})
}
