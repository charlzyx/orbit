package orbitui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charlzyx/orbit/internal/composerui"
	"github.com/charlzyx/orbit/internal/modelui"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func (a *App) messages(c *ui.Context) {
	t := a.Theme
	hits := a.searchMatches()
	active := -1
	if a.searchOpen && len(hits) > 0 {
		active = hits[min(a.searchCursor, len(hits)-1)]
	}
	scope := ui.Column(c.Key("transcript-wrap")).Selectable().Grow(1).MinHeight(0)
	transcriptui.RetainSelection(scope, a.Model.SessionID)
	scope.Children(func() {
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
				if row.Hovered() {
					a.transcriptHover = m.Text
				}
				if a.searchOpen && strings.TrimSpace(a.Search) != "" && strings.Contains(strings.ToLower(messageSearchText(m)), strings.ToLower(strings.TrimSpace(a.Search))) {
					alpha := float32(.07)
					if i == active {
						alpha = .15
					}
					row.Background(t.Accent.Alpha(alpha)).Radius(8)
				}
				row.Children(func() {
					if len(m.Images) > 0 {
						a.messageImageRow(c, m)
					}
					switch {
					case (m.Logical != nil && m.Role != "user") || m.Role == "changedFiles":
						a.logicalMessage(c, m)
					case m.Role == "user":
						ui.Row(c).AlignSelf(ui.Stretch).Children(func() {
							ui.Spacer(c)
							ui.Text(c, m.Text).MaxWidth(800).Padding(12, 16).Radius(t.Px(12)).Background(t.SurfaceSecondary).Selectable()
						})
						if m.Logical != nil {
							transcriptui.Footer(c, &a.transcriptUI, *m.Logical, a.transcriptOptions(c))
						}
					case m.Role == "toolResult" || (m.ToolName != "" && m.Role != "assistant"):
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
	transcriptui.SelectionMenu(c, scope, a.transcriptHover, a.transcriptOptions(c))
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
	a.autocomplete.SetContext(a.Model.Workspace, a.Options.Store.Home, a.Model.SessionID)
	a.autocomplete.SetCommands(a.Model.CommandsRaw)
	borderAlpha, badgeAlpha := float32(.14), float32(.09)
	if !t.Dark {
		borderAlpha, badgeAlpha = .16, .10
	}
	style := composerui.Style{FontSize: t.FontSize, Density: t.Density, Font: t.Font, Dark: t.Dark, Surface: t.Elevated, Border: t.Text.Alpha(borderAlpha), Selection: t.Selection, Hover: t.Hover, Text: t.Text, Secondary: t.TextSecondary, Muted: t.TextMuted, Badge: t.Text.Alpha(badgeAlpha), CommandColor: t.AccentText, FileColor: composerui.MentionFileColor(t.AccentText), TextFor: a.text, Icon: func(c *ui.Context, path string, size float32, color ui.Color) { a.icon(c, path, size, color) }}
	ui.Column(c.Key("composer-wrap")).AlignSelf(ui.Stretch).AlignItems(ui.Center).Padding(0, 16, 16).LineHeight(sourceDefaultLineHeight).Children(func() {
		ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).Gap(0).Children(func() {
			a.runStatusStrip(c)
			a.inlineExtensionPanels(c)
			a.queueBar(c)
			a.extensionWidgets(c, "aboveEditor")
			anchor := ui.Column(c.Key("composer")).AlignSelf(ui.Stretch).Padding(t.Space(12), t.Space(16), t.Space(12)).Gap(t.Space(16)).Radius(t.Px(16)).Border(1, t.Border).Background(t.SurfaceSecondary).Shadow(0, 2, 10, -3, t.composerContact()).Label("Composer box")
			var input ui.Element
			anchor.Children(func() {
				a.attachmentRow(c)
				input = ui.TextAreaBase(c.Key(struct {
					Field, Session, Workspace string
					ComposerEpoch             uint64
				}{"draft", a.Model.SessionID, a.Model.Workspace, a.extensionUI.composerEpoch}), &a.Draft).Bind(&a.composer).Label("Message").Placeholder(a.text("composer.placeholder", "Describe the task…")).Lines(1, 8).AlignSelf(ui.Stretch).OnTextKey(func(k ui.TextKey) bool {
					if a.autocomplete.HandleKey(k) {
						return true
					}
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

				if a.reviewComposerAutoFocus(c) {
					input.AutoFocus()
				}
				if a.autocomplete.SyncInput(input, &a.Draft) {
					c.AnimationFrame()
				}
				a.retainReviewComposer(c, input)
				a.trackExtensionComposer(c, input)
				composerui.PaintMentions(input, a.Draft, style)
				ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(8).Children(func() {
					a.composerChip(c, "Add attachment", "", "icons/plus.svg", false, 0).Menu(func(m *ui.Menu) {
						if m.Item("Attach files…").Chosen() {
							a.Dispatch(Action{Kind: "BrowseAttachments"})
						}
						if m.Item("Paste image").Chosen() {
							a.Dispatch(Action{Kind: "PasteImage"})
						}
					})
					compact := anchor.Bounds().W > 0 && anchor.Bounds().W < 600
					if !compact {
						accessLabels := map[string]string{"supervised": "Supervised", "auto-accept-edits": "Auto-accept edits", "full": "Full access", "full-access": "Full access"}
						label := accessLabels[a.access]
						if label == "" {
							label = a.access
						}
						display := label
						terse := a.access == "full-access" || a.access == "full"
						if terse {
							display = ""
						}
						a.composerChip(c, label, display, "icons/lock.svg", !terse, 0).Menu(func(m *ui.Menu) {
							for _, v := range []struct{ id, title string }{{"supervised", "Supervised"}, {"auto-accept-edits", "Auto-accept edits"}, {"full-access", "Full access"}} {
								if m.Item(v.title).Checked(a.access == v.id).Chosen() {
									a.access = v.id
									a.save(a.Options.Store.SaveAccess(v.id))
									a.Dispatch(Action{Kind: "SetAccess", Value: v.id})
								}
							}
						})
						a.composerChip(c, titleCase(a.workflow), titleCase(a.workflow), "icons/start-up.svg", true, 0).Menu(func(m *ui.Menu) {
							for _, v := range []string{"plan", "build", "ask"} {
								if m.Item(titleCase(v)).Checked(a.workflow == v).Chosen() {
									a.workflow = v
									a.Dispatch(Action{Kind: "SetWorkflow", Value: v})
								}
							}
						})
					}
					ui.Spacer(c)
					model := a.Model.ModelID
					if model == "" {
						model = "unknown"
					}
					displayModel := a.Model.CurrentModel.Name
					if displayModel == "" {
						displayModel = model
					}
					modelWidth := float32(220)
					if compact {
						modelWidth = 120
					}
					modelIcon := "icons/cloud.svg"
					provider := a.Model.ModelProvider
					if provider == "ollama-cloud" {
						provider = "ollama"
					}
					if icons["icons/providers/"+provider+".svg"] != nil {
						modelIcon = "icons/providers/" + provider + ".svg"
					}
					modelAnchor := a.composerChip(c, model, displayModel, modelIcon, true, modelWidth).Disabled(a.Model.Running)
					if modelAnchor.Clicked() {
						a.openModelPicker(modelui.ModelPicker, c.Now())
					}
					a.modelPickerPopup(c, modelAnchor, modelui.ModelPicker)
					if len(a.Model.ThinkingLevels) > 1 {
						level := a.Model.Thinking
						if level == "" {
							level = "Thinking"
						}
						thinkingAnchor := a.composerChip(c, titleCase(level), titleCase(level), "icons/thinking-medium.svg", true, 0).Disabled(a.Model.Running)
						if thinkingAnchor.Clicked() {
							a.openModelPicker(modelui.ThinkingPicker, c.Now())
						}
						a.modelPickerPopup(c, thinkingAnchor, modelui.ThinkingPicker)
					}

					if a.Model.Running {
						if strings.TrimSpace(a.Draft) != "" || len(a.attachments) > 0 {
							if a.composerRoundButton(c, "Steer agent", "icons/arrow-up-right.svg", t.SurfaceSecondary, t.AccentText).Clicked() && a.submit(a.Draft, "steer") {
								a.Draft = ""
							}
						}
						if a.composerRoundButton(c, "Stop agent", "icons/stop.svg", t.Error, t.ActionFg).Clicked() {
							a.Dispatch(Action{Kind: "AbortRun"})
						}
					} else {
						bg, fg := t.Action, t.ActionFg
						if strings.TrimSpace(a.Draft) == "" && len(a.attachments) == 0 {
							bg, fg = t.overlay(), t.TextMuted
						}
						b := a.composerRoundButton(c, "Send message", "icons/send.svg", bg, fg)
						if b.Clicked() {
							if a.autocomplete.CommitInput(input, &a.Draft) {
								c.AnimationFrame()
							} else if a.submit(a.Draft, "") {
								a.Draft = ""
							}
						}
					}
				})
			})
			a.extensionWidgets(c, "belowEditor")
			a.autocomplete.Popup(c, anchor, input, &a.Draft, style)
			ui.Row(c).Height(t.Px(22)).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(16).Children(func() {
				if a.composerChip(c, filepath.Base(a.Model.Workspace), filepath.Base(a.Model.Workspace), "icons/folder.svg", false, 0).Clicked() {
					a.Dispatch(Action{Kind: "BrowseWorkspace"})
				}
				if a.Options.WorkIn != nil {
					a.Options.WorkIn(c)
				}
				if a.Model.Branch != "" {
					if a.composerChip(c, a.Model.Branch, a.Model.Branch, "icons/git-branch.svg", false, 0).Clicked() {
						a.Navigate("git")
					}
				}
				ui.Spacer(c)
				a.extensionStatus(c)
				a.contextControl(c)
				if a.Model.Running {
					primary, alternate := a.text("composer.queue", "Queue"), a.text("composer.steer", "Steer")
					if a.Prefs.ComposerSendMode == "steer" {
						primary, alternate = alternate, primary
					}
					ui.Text(c, shortcutLabel("enter")+" "+primary+"   "+shortcutLabel("alt-enter")+" "+alternate).TextColor(t.TextMuted).FontSize(t.Px(12))
				}
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
		ui.TextInput(c, &a.Search).Bind(&a.searchInput).Placeholder("Find in transcript…").Label("Find in transcript").Grow(1).OnTextKey(func(k ui.TextKey) bool {
			if k.Key == ui.KeyEscape && k.Mods == 0 {
				a.closeFind()
				return true
			}
			if k.Key == ui.KeyEnter && (k.Mods == 0 || k.Mods == ui.Shift) {
				delta := 1
				if k.Mods == ui.Shift {
					delta = -1
				}
				a.moveMatch(delta)
				return true
			}
			return false
		})
		a.syncFind()
		matches := a.searchMatches()
		ui.Text(c, fmt.Sprintf("%d matches", len(matches))).TextColor(t.TextMuted).FontSize(t.Px(12))
		if a.iconButton(c, "Previous match", "icons/arrow-up.svg").Clicked() {
			a.moveMatch(-1)
		}
		if a.iconButton(c, "Next match", "icons/arrow-down.svg").Clicked() {
			a.moveMatch(1)
		}
		if a.iconButton(c, "Close find", "icons/x.svg").Clicked() {
			a.closeFind()
		}
	})
}
func (a *App) searchMatches() []int {
	var out []int
	query := strings.TrimSpace(a.Search)
	if query == "" {
		return out
	}
	for i, m := range a.Model.Messages {
		if strings.Contains(strings.ToLower(messageSearchText(m)), strings.ToLower(query)) {
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

func (a *App) closeFind() { a.searchOpen = false; a.Search = ""; a.composer.Focus() }
func (a *App) syncFind() {
	query := strings.TrimSpace(a.Search)
	count := len(a.Model.Messages)
	if query == a.searchLastQuery && count == a.searchLastCount {
		return
	}
	a.searchLastQuery = query
	a.searchLastCount = count
	matches := a.searchMatches()
	a.searchCursor = max(0, min(a.searchCursor, len(matches)-1))
	if len(matches) > 0 {
		a.transcript.ScrollTo(matches[a.searchCursor], ui.Center)
	}
}

// Source ButtonSize::Default: 22px frame, 12px label, 14px icon, 4px gap/pad.
func (a *App) composerChip(c *ui.Context, label, text, icon string, caret bool, maxText float32) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Label(label).Tooltip(label).Height(t.Px(22)).Padding(0, t.Space(4)).Radius(t.Px(8)).AlignItems(ui.Center).Gap(t.Space(4)).Shrink(0)
	if b.Hovered() {
		b.Background(t.Hover)
	}
	b.Children(func() {
		if icon != "" {
			a.icon(c, icon, t.Px(14), t.TextMuted)
		}
		if text != "" {
			e := ui.Text(c, text).FontSize(t.Px(12)).TextColor(t.TextSecondary).MaxLines(1)
			if maxText > 0 {
				e.MaxWidth(maxText).Ellipsis("…")
			}
		}
		if caret {
			a.icon(c, "icons/chevron-down.svg", t.Px(12), t.TextMuted)
		}
	})
	return b
}
func (a *App) composerRoundButton(c *ui.Context, label, icon string, bg, fg ui.Color) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Label(label).Tooltip(label).Size(t.Px(28), t.Px(28)).Radius(t.Px(14)).Background(bg).Center().Shrink(0)
	b.Children(func() { a.icon(c, icon, t.Px(14), fg) })
	return b
}
