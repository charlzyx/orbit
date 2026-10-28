package transcriptui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/egoist/mygo/ui"
)

func defaults(c *ui.Context, o Options) Options {
	if o.FontSize == 0 {
		o.FontSize = 14
	}
	if o.CodeSize == 0 {
		o.CodeSize = 12.5
	}
	if o.CodeFont == "" {
		o.CodeFont = "monospace"
	}
	if o.Colors.Text.A == 0 {
		t := c.Theme()
		o.Colors = Colors{Text: t.Text, Muted: t.TextMuted, Accent: t.Accent, Border: t.Border, Surface: t.Surface, Code: t.Background, Error: t.Danger, Warning: t.Warning, Success: t.Success, Selection: t.Selection}
	}
	return o
}
func (o Options) tr(key, fallback string, vars ...string) string {
	s := fallback
	if o.Text != nil {
		s = o.Text(key, fallback)
	}
	for i := 0; i+1 < len(vars); i += 2 {
		s = strings.ReplaceAll(s, "%{"+vars[i]+"}", vars[i+1])
	}
	return s
}
func (o Options) icon(c *ui.Context, path string, size float32, color ui.Color) {
	if o.Icon != nil {
		o.Icon(c, path, size, color)
	}
}
func (o Options) now(c *ui.Context) time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return c.Now()
}
func (o Options) prose(c *ui.Context, text string, live bool) {
	if o.Markdown != nil {
		o.Markdown(c, text, live)
	} else {
		ui.Text(c, text).Selectable().TextColor(o.Colors.Text).Font(o.Font).FontSize(o.FontSize)
	}
}
func copyButton(c *ui.Context, s *State, key, text, label string, o Options) {
	now := o.now(c)
	copied := !s.copied[key].IsZero() && now.Sub(s.copied[key]) < 2*time.Second
	if copied {
		c.After(2*time.Second - now.Sub(s.copied[key]))
	}
	b := ui.ButtonBase(c.Key("copy:"+key)).Label(label).Tooltip(label).Padding(3, 6)
	b.Children(func() {
		icon := "icons/copy.svg"
		color := o.Colors.Muted
		if copied {
			icon = "icons/check.svg"
			color = o.Colors.Success
		}
		o.icon(c, icon, 14, color)
		if o.Icon == nil {
			v := o.tr("common.copy", "Copy")
			if copied {
				v = o.tr("common.copied", "Copied")
			}
			ui.Text(c, v).FontSize(12).TextColor(color)
		}
	})
	if b.Clicked() {
		if o.Copy != nil {
			o.Copy(text)
		} else {
			c.WriteClipboard(text)
		}
		if s.copied == nil {
			s.Reset(s.session)
		}
		s.copied[key] = now
	}
}
func toggle(c *ui.Context, s *State, key, label string, defaultOpen bool, o Options) bool {
	open := s.Open(key, defaultOpen)
	b := ui.ButtonBase(c.Key(key)).AlignSelf(ui.Stretch).Padding(5, 8).Label(label)
	b.Children(func() {
		chevron := "icons/chevron-right.svg"
		if open {
			chevron = "icons/chevron-down.svg"
		}
		o.icon(c, chevron, 12, o.Colors.Muted)
		ui.Text(c, label).FontSize(o.FontSize - 1).TextColor(o.Colors.Muted)
	})
	if b.Clicked() {
		s.Toggle(key, defaultOpen)
	}
	return open
}

// Tool renders one typed activity. State must be scoped by its turn's keyed
// context, and ToolCall.ID supplies the stable tool identity within that turn.
func Tool(c *ui.Context, s *State, t ToolCall, o Options) ui.Element {
	o = defaults(c, o)
	if qs := t.Questions(); len(qs) > 0 {
		return questionCard(c, t, qs, o)
	}
	key := "tool:" + t.ID
	if t.ID == "" {
		key += "unknown:" + t.Name
	}
	c = c.Key(key)
	open := s.Open(key, false)
	hasDetail := present(t.Arguments) || present(t.Output)
	diff := t.Diff()
	return ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Border(1, o.Colors.Border).Radius(8).Background(o.Colors.Surface).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(4).Children(func() {
			label := ActionLabel(t.Name)
			preview := t.Preview()
			if server, tool, ok := MCPParts(t.Name); ok {
				label = o.tr("mcp.card_action", "MCP · %{server}", "server", server)
				preview = tool
			}
			b := ui.ButtonBase(c.Key("header")).Grow(1).MinWidth(0).Height(32).Padding(0, 8).Gap(8).Label(label + " " + preview).Disabled(!hasDetail)
			b.Children(func() {
				tone := o.Colors.Muted
				if t.Running {
					tone = o.Colors.Accent
				}
				if t.Failed {
					tone = o.Colors.Error
				}
				o.icon(c, Icon(t.Name), 16, tone)
				ui.Text(c, label).FontWeight(600).FontSize(o.FontSize).TextColor(o.Colors.Text)
				if _, command := t.Command(); command {
					ui.Text(c, "$").TextColor(o.Colors.Muted)
					code(c, preview, highlight.Shell, o).Grow(1).MinWidth(0).MaxHeight(22)
				} else {
					ui.Text(c, preview).Grow(1).MinWidth(0).MaxHeight(22).TextColor(o.Colors.Muted).FontSize(o.FontSize - 1)
				}
				if t.Facts.Truncated {
					label := o.tr("transcript.truncated", "truncated")
					if t.Facts.OutputLines != nil && t.Facts.TotalLines != nil && *t.Facts.TotalLines > *t.Facts.OutputLines {
						label += fmt.Sprintf(" · %d/%d", *t.Facts.OutputLines, *t.Facts.TotalLines)
					}
					ui.Text(c, label).TextColor(o.Colors.Warning).FontSize(12)
				}
				if t.Added > 0 || t.Removed > 0 {
					ui.Text(c, fmt.Sprintf("+%d", t.Added)).TextColor(o.Colors.Success).FontSize(12)
					ui.Text(c, fmt.Sprintf("-%d", t.Removed)).TextColor(o.Colors.Error).FontSize(12)
				}
				if t.Running {
					ui.Spinner(c).Size(12, 12)
				} else if t.Failed {
					o.icon(c, "icons/stop.svg", 12, o.Colors.Error)
				} else if t.Complete {
					o.icon(c, "icons/check.svg", 12, o.Colors.Muted)
				}
				if hasDetail {
					glyph := "icons/chevron-right.svg"
					if open {
						glyph = "icons/chevron-down.svg"
					}
					o.icon(c, glyph, 12, o.Colors.Muted)
				}
			})
			if b.Clicked() {
				s.Toggle(key, false)
			}
			if len(diff) > 0 {
				copyButton(c, s, key+":diff", DiffText(diff), o.tr("transcript.copy_diff", "Copy diff"), o)
			}
			if in(t.Name, "web_fetch webfetch fetch open_url browse") {
				if url := str(obj(t.Arguments)["url"]); url != "" {
					ui.Link(c, o.tr("common.open", "Open"), url).FontSize(12).Padding(4).Tooltip(url)
				}
			}
		})
		if t.Failed {
			error := t.ErrorText()
			if error == "" {
				error = o.tr("transcript.run_failed", "The run failed without an error message.")
			}
			ui.Text(c, error).Selectable().AlignSelf(ui.Stretch).Padding(4, 8).TextColor(o.Colors.Error).Background(o.Colors.Error.Alpha(.07)).FontSize(o.FontSize - 1)
		}
		if open && hasDetail {
			if len(diff) > 0 {
				diffView(c, s, key, diff, o)
			}
			for _, section := range t.Sections() {
				sectionView(c, s, key, section, o)
			}
		}
	})
}
func sectionView(c *ui.Context, s *State, toolKey string, section Section, o Options) {
	key := toolKey + ":" + section.ID
	c = c.Key(key)
	open := s.Open(key, false)
	window := SectionLines(section.Text, open)
	ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Border(1, o.Colors.Border).Background(o.Colors.Code).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).Unselectable().Padding(4, 8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, o.tr("transcript."+strings.ToLower(section.Label), section.Label)).FontSize(o.FontSize - 2).TextColor(o.Colors.Muted)
			ui.Spacer(c)
			copyButton(c, s, key, section.Text, o.tr("transcript.copy_section", "Copy section"), o)
		})
		text := strings.Join(window.Visible, "\n")
		ui.Row(c).AlignSelf(ui.Stretch).MinWidth(0).Padding(4, 12).Gap(6).Children(func() {
			if section.Prompt {
				ui.Text(c, "$").Unselectable().TextColor(o.Colors.Muted).Font(o.CodeFont).FontSize(o.CodeSize)
			}
			code(c, text, section.Language, o).Grow(1).MinWidth(0).Selectable()
		})
		if window.Clipped {
			ui.Text(c, o.tr("transcript.showing_first_output", "Showing first %{count} — copy for the full log", "count", fmt.Sprint(OutputPaintLines))).Unselectable().FontSize(12).TextColor(o.Colors.Muted).Padding(4, 8)
		}
		if window.Foldable {
			label := o.tr("transcript.show_all_lines", "Show all %{count} lines", "count", fmt.Sprint(window.Total))
			if open {
				label = o.tr("transcript.show_less", "Show less")
			}
			toggle(c, s, key, label, false, o)
		}
	})
}
func code(c *ui.Context, text string, lang highlight.Language, o Options) ui.Element {
	var spans []ui.Span
	tokens := highlight.TokenizeCached(lang, text).Lines
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			spans = append(spans, ui.Span{Text: "\n"})
		}
		var ts []highlight.Token
		if i < len(tokens) {
			ts = tokens[i]
		}
		at := 0
		for _, token := range ts {
			if token.Start > at {
				spans = append(spans, ui.Span{Text: line[at:token.Start]})
			}
			color := o.Colors.Syntax[token.Class]
			if color.A == 0 {
				color = o.Colors.Text
			}
			spans = append(spans, ui.Span{Text: line[token.Start:token.End], Color: color})
			at = token.End
		}
		if at < len(line) {
			spans = append(spans, ui.Span{Text: line[at:]})
		}
	}
	return ui.RichText(c, spans...).Font(o.CodeFont).FontSize(o.CodeSize).TextColor(o.Colors.Text)
}
func diffView(c *ui.Context, s *State, key string, rows []DiffRow, o Options) {
	key += ":diff-lines"
	c = c.Key(key)
	open := s.Open(key, false)
	n := len(rows)
	foldable := n > DiffCollapseLines
	limit := n
	if foldable && !open {
		limit = DiffPreviewLines
	} else {
		limit = min(n, DiffPaintLines)
	}
	ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Background(o.Colors.Code).Children(func() {
		for i, r := range rows[:limit] {
			rc := c.Key(fmt.Sprintf("diff-row-%d", i))
			if r.Kind == Break {
				ui.Box(rc).AlignSelf(ui.Stretch).Height(9).BorderWidth(1, 0, 0, 0).BorderColor(o.Colors.Border).Unselectable()
				continue
			}
			marker, tone, body, gutter := "", o.Colors.Muted, ui.Transparent, ui.Transparent
			switch r.Kind {
			case Addition:
				marker, tone = "+", o.Colors.Success
			case Deletion:
				marker, tone = "-", o.Colors.Error
			}
			row := ui.Row(rc).AlignSelf(ui.Stretch).MinWidth(0).MinHeight(19)
			if r.Kind == Addition || r.Kind == Deletion {
				body = tone.Alpha(.18)
				gutter = tone.Alpha(.24)
				row.BorderWidth(0, 0, 0, 2).BorderColor(tone)
			}
			row.Background(body).Children(func() {
				ui.Text(rc, marker).Unselectable().Width(24).MinHeight(19).TextAlign(ui.Center).Background(gutter).Font(o.CodeFont).FontSize(o.CodeSize).TextColor(tone)
				var spans []ui.Span
				at := 0
				for _, tk := range r.Tokens {
					if tk.Start > at {
						spans = append(spans, ui.Span{Text: r.Text[at:tk.Start]})
					}
					color := o.Colors.Syntax[tk.Class]
					if color.A == 0 {
						color = o.Colors.Text
					}
					spans = append(spans, ui.Span{Text: r.Text[tk.Start:tk.End], Color: color})
					at = tk.End
				}
				if at < len(r.Text) {
					spans = append(spans, ui.Span{Text: r.Text[at:]})
				}
				ui.RichText(rc, spans...).Grow(1).MinWidth(0).MinHeight(19).Selectable().Font(o.CodeFont).FontSize(o.CodeSize).TextColor(o.Colors.Text).Padding(0, 12, 0, 0)
			})
		}

		if open && n > limit {
			ui.Text(c, o.tr("transcript.showing_first_diff", "Showing first %{count} — copy for the whole diff", "count", fmt.Sprint(limit))).Unselectable().FontSize(12).TextColor(o.Colors.Muted).Padding(4, 8)
		}
		if foldable {
			label := o.tr("transcript.show_all_lines", "Show all %{count} lines", "count", fmt.Sprint(n))
			if open {
				label = o.tr("transcript.show_less", "Show less")
			}
			toggle(c, s, key, label, false, o)
		}
	})
}

// Summary renders a context boundary, never an assistant answer footer.
func Summary(c *ui.Context, s *State, row Row, o Options) ui.Element {
	o = defaults(c, o)
	if row.Summary == nil {
		return ui.Column(c)
	}
	key := "summary:" + row.ID
	c = c.Key(key)
	open := s.Open(key, false)
	summary := row.Summary
	return ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Border(1, o.Colors.Border).Radius(8).Background(o.Colors.Surface).Children(func() {
		title := o.tr("transcript.compaction", "Compaction")
		if summary.Branch {
			title = o.tr("transcript.branch_summary", "Branch summary")
		}
		if summary.TokensBefore != nil {
			title += " · " + o.tr("transcript.compacted_from", "Compacted from %{tokens} tokens", "tokens", FormatTokens(*summary.TokensBefore))
		}
		toggle(c, s, key, title, false, o)
		if open {
			ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Padding(8, 12).Children(func() { o.prose(c, summary.Text, false) })
		}
	})
}

// Files is the source's end-of-task changed-files card. Call it once after
// the settled transcript with ChangedFilesOf(rows), not after every turn.
func Files(c *ui.Context, s *State, files []ChangedFile, o Options) ui.Element {
	o = defaults(c, o)
	key := "changed-files"
	c = c.Key(key)
	open := s.Open(key, false)
	visible := min(3, len(files))
	if open {
		visible = min(12, len(files))
	}
	var added, removed uint64
	for _, f := range files {
		added += f.Added
		removed += f.Removed
	}
	return ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Border(1, o.Colors.Border).Radius(o.FontSize * 12 / 14).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).Unselectable().MinHeight(o.FontSize*56/14).Padding(o.FontSize*10/14, o.FontSize).Gap(o.FontSize*10/14).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(o.Colors.Border).Children(func() {
			ui.Box(c).Size(o.FontSize*2, o.FontSize*2).Radius(o.FontSize*8/14).Background(o.Colors.Surface).Border(1, o.Colors.Border).Center().Children(func() { o.icon(c, "icons/file-diff.svg", o.FontSize, o.Colors.Muted) })
			ui.Column(c).Grow(1).MinWidth(0).Gap(o.FontSize * 2 / 14).Children(func() {
				ui.Text(c, changedFilesTitle(len(files), o)).FontSize(o.FontSize).FontWeight(500).FixedLineHeight(o.FontSize * 17 / 14).TextColor(o.Colors.Text)
				ui.Row(c).Gap(o.FontSize * 6 / 14).Children(func() {
					ui.Text(c, fmt.Sprintf("+%d", added)).FontSize(o.FontSize * 12 / 14).FixedLineHeight(o.FontSize).TextColor(o.Colors.Success)
					ui.Text(c, fmt.Sprintf("-%d", removed)).FontSize(o.FontSize * 12 / 14).FixedLineHeight(o.FontSize).TextColor(o.Colors.Error)
				})
			})
			if o.Review != nil {
				label := o.tr("transcript_view.review", "Review")
				review := ui.ButtonBase(c).Label(label).Height(o.FontSize*2).Padding(0, o.FontSize*8/14).Gap(o.FontSize*4/14).Radius(o.FontSize*8/14).Background(o.Colors.Surface).Border(1, o.Colors.Border).AlignItems(ui.Center)
				review.Children(func() {
					o.icon(c, "icons/file-diff.svg", o.FontSize, o.Colors.Muted)
					ui.Text(c, label).FontSize(o.FontSize).TextColor(o.Colors.Muted)
				})
				if review.Clicked() {
					o.Review()
				}
			}
		})
		for i, f := range files[:visible] {
			ui.Row(c.Key(i)).AlignSelf(ui.Stretch).Unselectable().Height(32).Padding(0, 14).Gap(8).AlignItems(ui.Center).Children(func() {
				label := f.Path
				if o.Workspace != "" {
					if rel, e := filepath.Rel(o.Workspace, f.Path); e == nil && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) {
						label = rel
					}
				}
				o.icon(c, "icons/tools/edit.svg", 14, o.Colors.Muted)
				ui.Text(c, label).Grow(1).MinWidth(0).MaxHeight(20).FontSize(o.FontSize - 1).TextColor(o.Colors.Muted)
				ui.Text(c, fmt.Sprintf("+%d", f.Added)).FontSize(12).TextColor(o.Colors.Success)
				ui.Text(c, fmt.Sprintf("-%d", f.Removed)).FontSize(12).TextColor(o.Colors.Error)
			})
		}
		if len(files) > 3 {
			label := o.tr("transcript_view.show_more_files", "Show %{count} more files", "count", fmt.Sprint(len(files)-3))
			if len(files) == 4 {
				label = o.tr("transcript_view.show_one_more_file", "Show 1 more file")
			}
			if open {
				label = o.tr("transcript_view.show_fewer_files", "Show fewer files")
			}
			if open && len(files) > 12 {
				label += " · " + o.tr("transcript_view.showing_first_of", "Showing first %{limit} of %{total}", "limit", "12", "total", fmt.Sprint(len(files)))
			}
			toggle(c, s, key, label, false, o)
		}
	})
}

func changedFilesTitle(n int, o Options) string {
	key, fallback := "transcript_view.changed_file_other", "Changed %{count} files"
	if n == 1 {
		key, fallback = "transcript_view.changed_file_one", "Changed %{count} file"
	}
	return o.tr(key, fallback, "count", fmt.Sprint(n))
}
