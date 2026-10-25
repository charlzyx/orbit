package orbitui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func (a *App) messages(c *ui.Context) {
	t := a.Theme
	ui.Column(c.Key("transcript-wrap")).Grow(1).MinHeight(0).Children(func() {
		a.transcript.Key = func(i int) any {
			m := a.Model.Messages[i]
			if m.ID != "" {
				return m.ID
			}
			return fmt.Sprintf("message-%d", i)
		}
		ui.List(c.Key("transcript"), &a.transcript, len(a.Model.Messages), func(i int) {
			m := a.Model.Messages[i]
			ui.Column(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Padding(10, 24).Children(func() {
				row := ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).Gap(t.Space(8))
				if a.Search != "" && strings.Contains(strings.ToLower(m.Text), strings.ToLower(a.Search)) {
					row.Background(t.Accent.Alpha(.07)).Radius(8)
				}
				row.Children(func() {
					if len(m.Images) > 0 {
						a.messageImageRow(c, m)
					}
					switch {
					case m.Role == "user":
						ui.Row(c).AlignSelf(ui.Stretch).Children(func() {
							ui.Spacer(c)
							ui.Text(c, m.Text).MaxWidth(800).Padding(12, 16).Radius(t.Px(12)).Background(t.SurfaceSecondary).Selectable()
						})
					case m.ToolName != "" || m.Role == "toolResult":
						a.toolMessage(c, m)
					default:
						if m.Thinking != "" {
							label := "Thinking"
							if !m.Streaming {
								label = "Thought process"
							}
							if a.button(c, label, "icons/thinking-medium.svg", false).Clicked() {
								a.collapsedMessages[m.ID+":thinking"] = !a.collapsedMessages[m.ID+":thinking"]
							}
							if a.collapsedMessages[m.ID+":thinking"] {
								ui.Text(c, m.Thinking).TextColor(t.TextMuted).FontSize(t.Px(12)).Selectable().Padding(10).Border(1, t.Border).Radius(8)
							}
						}
						if m.Text != "" {
							if m.Streaming {
								markdown.Stream(c, m.Text, &a.md)
							} else {
								markdown.View(c, m.Text, &a.md)
							}
						}
						if m.Streaming {
							ui.Spinner(c).Size(12, 12)
						}
						if m.Error != "" {
							ui.Text(c, m.Error).TextColor(t.Error).Padding(8).Background(t.Error.Alpha(.08)).Radius(6)
						}
						if !m.Streaming && m.Text != "" {
							ui.Row(c).Gap(4).Children(func() {
								if a.iconButton(c, "Copy response", "icons/copy.svg").Clicked() {
									c.Services().WriteClipboard(m.Text)
								}
								ui.Text(c, " ").TextColor(t.TextMuted)
							})
						}
					}
				})
			})
		}).Grow(1).MinHeight(0).Justify(ui.End)
		if !a.transcript.AtEnd() {
			ui.Row(c).AlignSelf(ui.Stretch).Justify(ui.Center).Children(func() {
				if a.button(c, "Jump to latest", "icons/arrow-down.svg", false).Clicked() {
					a.transcript.ScrollToEnd()
				}
			})
		}
	})
}
func (a *App) toolMessage(c *ui.Context, m Message) {
	t := a.Theme
	name := m.ToolName
	if name == "" {
		name = "Tool result"
	}
	open := a.collapsedMessages[m.ID]
	ui.Column(c).AlignSelf(ui.Stretch).Border(1, t.Border).Radius(8).Clip().Children(func() {
		b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Padding(10, 12).Gap(8).Label(name)
		b.Children(func() {
			a.icon(c, "icons/terminal.svg", 16, t.TextMuted)
			ui.Text(c, name).FontWeight(500)
			ui.Spacer(c)
			if m.Streaming {
				ui.Spinner(c).Size(12, 12)
			}
			glyph := "icons/chevron-down.svg"
			if !open {
				glyph = "icons/chevron-right.svg"
			}
			a.icon(c, glyph, 14, t.TextMuted)
		})
		if b.Clicked() {
			a.collapsedMessages[m.ID] = !open
		}
		if open {
			body := m.ToolOutput
			if body == "" {
				body = m.Text
			}
			ui.Scroll(c).MaxHeight(320).Padding(12).Children(func() { ui.Text(c, body).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize).Selectable() })
			if m.Error != "" {
				ui.Text(c, m.Error).TextColor(t.Error).Padding(10)
			}
		}
	})
}
func (a *App) composerView(c *ui.Context) {
	t := a.Theme
	ui.Column(c.Key("composer-wrap")).AlignSelf(ui.Stretch).AlignItems(ui.Center).Padding(12, 24, 10).Children(func() {
		ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).Gap(6).Children(func() {
			a.attachmentRow(c)
			for i, q := range a.Model.Queued {
				ui.Row(c.Key(fmt.Sprintf("queued-%d", i))).Padding(8, 12).Radius(8).Border(1, t.Border).Background(t.SurfaceSecondary).Children(func() {
					ui.Text(c, q).MaxLines(2).Grow(1).TextColor(t.TextMuted)
					if a.iconButton(c, "Clear queued messages", "icons/x.svg").Clicked() {
						a.Dispatch(Action{Kind: "ClearQueue"})
					}
				})
			}
			ui.Column(c.Key("composer")).AlignSelf(ui.Stretch).Padding(14, 16, 10).Gap(16).Radius(18).Border(1, t.Border).Background(t.Surface).Children(func() {
				ui.TextAreaBase(c.Key("draft"), &a.Draft).Bind(&a.composer).Label("Message").Placeholder(a.text("composer.placeholder", "Describe the task…")).Lines(2, 10).AlignSelf(ui.Stretch).OnTextKey(func(k ui.TextKey) bool {
					if k.Key == ui.KeyEnter && k.Mods == 0 {
						if a.submit(k.Text(), "") {
							k.SetText("")
						}
						return true
					}
					if k.Key == ui.KeyEnter && k.Mods == ui.Alt {
						if a.submit(k.Text(), "alternate") {
							k.SetText("")
						}
						return true
					}
					if k.Key == ui.KeyEnter && k.Mods == (ui.Cmd|ui.Shift) {
						if a.submit(k.Text(), "steer") {
							k.SetText("")
						}
						return true
					}
					return false
				}).OnPaste(func(text string) bool {
					if text == "" {
						a.Dispatch(Action{Kind: "PasteImage"})
						return true
					}
					return false
				})
				ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(5).Children(func() {
					a.iconButton(c, "Add attachment", "icons/plus.svg").Menu(func(m *ui.Menu) {
						if m.Item("Attach files…").Chosen() {
							a.Dispatch(Action{Kind: "BrowseAttachments"})
						}
						if m.Item("Paste image").Chosen() {
							a.Dispatch(Action{Kind: "PasteImage"})
						}
					})
					accessLabels := map[string]string{"supervised": "Supervised", "auto-accept-edits": "Auto-accept edits", "full": "Full access", "full-access": "Full access"}
					label := accessLabels[a.access]
					if label == "" {
						label = a.access
					}
					a.button(c, label, "icons/lock.svg", false).Menu(func(m *ui.Menu) {
						for _, v := range []struct{ id, title string }{{"supervised", "Supervised"}, {"auto-accept-edits", "Auto-accept edits"}, {"full-access", "Full access"}} {
							if m.Item(v.title).Checked(a.access == v.id).Chosen() {
								a.access = v.id
								a.save(a.Options.Store.SaveAccess(v.id))
								a.Dispatch(Action{Kind: "SetAccess", Value: v.id})
							}
						}
					})
					a.button(c, titleCase(a.workflow), "icons/start-up.svg", false).Menu(func(m *ui.Menu) {
						for _, v := range []string{"plan", "build", "ask"} {
							if m.Item(titleCase(v)).Checked(a.workflow == v).Chosen() {
								a.workflow = v
								if a.Model.SessionID != "" {
									a.save(a.Options.Store.SaveWorkflow(a.Model.SessionID, v))
								}
								a.Dispatch(Action{Kind: "SetWorkflow", Value: v})
							}
						}
					})
					ui.Spacer(c)
					model := a.Model.ModelID
					if model == "" {
						model = "Choose model"
					}
					if a.button(c, model, "icons/spark.svg", false).Clicked() {
						a.modelOpen = true
					}
					level := a.Model.Thinking
					if level == "" {
						level = "Thinking"
					}
					a.button(c, titleCase(level), "icons/thinking-medium.svg", false).Menu(func(m *ui.Menu) {
						if len(a.Model.ThinkingLevels) == 0 {
							m.Item("Not available for this model").Disabled(true)
						}
						for _, v := range a.Model.ThinkingLevels {
							if m.Item(titleCase(v)).Checked(v == a.Model.Thinking).Chosen() {
								a.Dispatch(Action{Kind: "SetThinking", Value: v})
							}
						}
					})
					if a.Model.Running {
						if a.iconButton(c, "Stop agent", "icons/stop.svg").Clicked() {
							a.Dispatch(Action{Kind: "AbortRun"})
						}
					} else {
						b := a.iconButton(c, "Send message", "icons/send.svg").Disabled(strings.TrimSpace(a.Draft) == "" && len(a.attachments) == 0)
						if b.Clicked() {
							if a.submit(a.Draft, "") {
								a.Draft = ""
							}
						}
					}
				})
			})
			ui.Row(c).Height(24).AlignItems(ui.Center).Gap(12).Children(func() {
				if a.button(c, filepath.Base(a.Model.Workspace), "icons/folder.svg", false).Height(22).Clicked() {
					a.Dispatch(Action{Kind: "BrowseWorkspace"})
				}
				if a.Options.WorkIn != nil {
					a.Options.WorkIn(c)
				}
				if a.Model.Branch != "" {
					if a.button(c, a.Model.Branch, "icons/git-branch.svg", false).Height(22).Clicked() {
						a.Navigate("git")
					}
				}
				ui.Spacer(c)
				if a.Model.Usage.ContextAvailable {
					usage := a.Model.Usage
					detail := fmt.Sprintf("Context: %d / %d tokens", usage.ContextTokens, usage.ContextWindow)
					if usage.TokensAvailable {
						detail += fmt.Sprintf("\nSession: %d input · %d output · %d cache read · %d cache write", usage.Input, usage.Output, usage.CacheRead, usage.CacheWrite)
					}
					if usage.CostAvailable {
						detail += fmt.Sprintf("\nSession cost: $%.4f", usage.Cost)
					}
					ui.Text(c, contextPercentLabel(usage)).Tooltip(detail).TextColor(t.TextMuted).FontSize(t.Px(12))
				}
				ui.Text(c, "↵ Send").TextColor(t.TextMuted).FontSize(t.Px(12))
			})
		})
	})
}
func titleCase(s string) string {
	if s == "" {
		return "Build"
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
func (a *App) submit(text, mode string) bool {
	text = strings.TrimSpace(text)
	if text == "" && len(a.attachments) == 0 {
		return false
	}
	if mode == "" {
		mode = a.Prefs.ComposerSendMode
	}
	if mode == "alternate" {
		if a.Prefs.ComposerSendMode == "steer" {
			mode = "follow_up"
		} else {
			mode = "steer"
		}
	}
	a.submitAccepted = true
	a.Dispatch(Action{Kind: "Submit", Value: text, Data: map[string]any{"mode": mode}})
	a.transcript.ScrollToEnd()
	return a.submitAccepted
}
func (a *App) findBar(c *ui.Context) {
	t := a.Theme
	ui.Row(c.Key("find-transcript")).Padding(8, 16).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c, &a.Search).Placeholder("Find in transcript…").Label("Find in transcript").Grow(1)
		matches := a.searchMatches()
		ui.Text(c, fmt.Sprintf("%d matches", len(matches))).TextColor(t.TextMuted).FontSize(t.Px(12))
		if a.iconButton(c, "Previous match", "icons/arrow-up.svg").Clicked() {
			a.moveMatch(-1)
		}
		if a.iconButton(c, "Next match", "icons/arrow-down.svg").Clicked() {
			a.moveMatch(1)
		}
		if a.iconButton(c, "Close find", "icons/x.svg").Clicked() {
			a.searchOpen = false
			a.Search = ""
		}
	})
}
func (a *App) searchMatches() []int {
	var out []int
	if a.Search == "" {
		return out
	}
	for i, m := range a.Model.Messages {
		if strings.Contains(strings.ToLower(m.Text+"\n"+m.ToolOutput), strings.ToLower(a.Search)) {
			out = append(out, i)
		}
	}
	return out
}
func (a *App) moveMatch(delta int) {
	m := a.searchMatches()
	if len(m) == 0 {
		return
	}
	a.searchCursor = (a.searchCursor + delta + len(m)) % len(m)
	a.transcript.ScrollTo(m[a.searchCursor], ui.Center)
}
