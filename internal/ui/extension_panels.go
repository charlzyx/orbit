package orbitui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
)

type extensionUIState struct {
	composerEpoch                       uint64
	composerRevisions                   map[string]uint64
	composerPending, composerWasFocused bool
	dialogID                            string
	dialogCursor                        int
	dialogConfirmed                     bool
	dialogPointerX, dialogPointerY      float32
	dialogPointerValid                  bool
	dialogSent                          bool
	dialogReveal                        bool
	approvalID                          string
	approvalCursor                      int
	approvalSent                        bool
	approvalFocusPending                bool
	askID                               string
	askCursor                           int
	askHighlight                        []int
	askAnswers                          []agent.AskAnswer
	askSent                             bool
	askFocusPending                     bool
	notices                             map[string]uint64
}

func (a *App) syncExtensionUI(s Snapshot) {
	st := &a.extensionUI
	if st.composerRevisions == nil {
		st.composerRevisions = make(map[string]uint64)
	}
	composerKey := s.RuntimeID + ":" + s.SessionID
	if s.ComposerRevision > 0 && s.ComposerRevision != st.composerRevisions[composerKey] {
		st.composerRevisions[composerKey] = s.ComposerRevision
		st.composerEpoch++
		st.composerPending = true
		a.reviewState.layout.composer.valid = false
		if a.reviewFull() {
			st.composerWasFocused = false
		}
		a.Draft = s.ComposerText
	}
	if s.Dialog == nil {
		if st.dialogID != "" {
			a.composer.Focus()
		}
		st.dialogID = ""
		st.dialogSent = false
	} else if st.dialogID != s.Dialog.ID {
		st.dialogID = s.Dialog.ID
		st.dialogCursor = 0
		st.dialogConfirmed = true
		st.dialogPointerValid = false
		st.dialogSent = false
		st.dialogReveal = false
	}
	if s.Approval == nil {
		st.approvalID = ""
		st.approvalSent = false
	} else if st.approvalID != s.Approval.ID {
		st.approvalFocusPending = true
		st.approvalID = s.Approval.ID
		st.approvalCursor = 0
		st.approvalSent = false
	}
	if s.Ask == nil {
		st.askID = ""
		st.askSent = false
	} else if st.askID != s.Ask.ID {
		st.askFocusPending = true
		st.askID = s.Ask.ID
		st.askCursor = 0
		st.askSent = false
		st.askHighlight = make([]int, len(s.Ask.Questions))
		st.askAnswers = make([]agent.AskAnswer, len(s.Ask.Questions))
		for i, q := range s.Ask.Questions {
			st.askAnswers[i] = agent.AskAnswer{Kind: "option"}
			if q.MultiSelect {
				st.askAnswers[i] = agent.AskAnswer{Kind: "multi", Checked: make([]bool, len(q.Options))}
			} else if len(q.Options) == 0 {
				st.askAnswers[i].Kind = "custom"
			}
		}
	}
	if st.notices == nil {
		st.notices = make(map[string]uint64)
	}
	for _, notice := range s.ExtensionNotices {
		key := s.RuntimeID + ":" + s.SessionID + ":" + notice.ID
		if notice.Sequence > st.notices[key] {
			a.Notify(notice.Message, false)
			st.notices[key] = notice.Sequence
		}
	}
}
func (a *App) respondExtensionDialog(d *DialogRequest, value string, confirmed *bool, cancelled bool) {
	if d == nil || a.extensionUI.dialogSent {
		return
	}
	a.extensionUI.dialogSent = true
	data := map[string]any{"id": d.ID}
	if confirmed != nil {
		data["confirmed"] = *confirmed
	}
	if cancelled {
		data["cancelled"] = true
	}
	a.Dispatch(Action{Kind: "RespondDialog", Value: value, Data: data})
	a.composer.Focus()
}
func (a *App) respondApproval(choice int, cancelled bool) {
	p := a.Model.Approval
	if p == nil || a.extensionUI.approvalSent {
		return
	}
	a.extensionUI.approvalSent = true
	data := map[string]any{"id": p.ID}
	value := ""
	if cancelled || choice < 0 || choice >= len(p.Options) {
		data["cancelled"] = true
	} else {
		value = p.Options[choice]
	}
	a.Dispatch(Action{Kind: "RespondDialog", Value: value, Data: data})
	a.composer.Focus()
}
func (a *App) inlineExtensionPanels(c *ui.Context) {
	if a.Model.Ask != nil {
		a.askPanel(c.Key("inline-questionnaire:" + a.Model.Ask.ID))
	}
	if a.Model.Approval != nil {
		a.approvalPanel(c.Key("inline-approval:" + a.Model.Approval.ID))
	}
}
func (a *App) approvalPanel(c *ui.Context) {
	p, t, st := a.Model.Approval, a.Theme, &a.extensionUI
	heading := a.text("approval.permission_needed", "Permission needed")
	if server, tool, ok := transcriptui.MCPParts(p.Tool); ok {
		heading = strings.NewReplacer("%{server}", server, "%{tool}", tool).Replace(a.text("mcp.allow_tool", "Allow MCP · %{server} · %{tool}?"))
	} else if strings.TrimSpace(p.Tool) != "" {
		heading = strings.ReplaceAll(a.text("approval.allow_tool", "Allow %{tool}?"), "%{tool}", p.Tool)
	}
	card := ui.Row(c).AlignSelf(ui.Stretch).Height(48).Margin(0, 0, 8).Padding(0, t.Space(12)).Gap(10).AlignItems(ui.Center).Radius(t.Px(10)).Background(t.Elevated).Border(1, t.Accent.Alpha(.3)).Clip().Focusable().Label("Tool approval")
	if a.reviewFull() {
		st.approvalFocusPending = true
	} else if st.approvalFocusPending && !st.approvalSent {
		card.Focus()
		st.approvalFocusPending = false
	}
	card.Disabled(st.approvalSent)
	card.Children(func() {
		ui.Box(c).Size(28, 28).Radius(8).Center().Background(t.Accent.Alpha(.16)).Children(func() { a.icon(c, "icons/lock.svg", 16, t.AccentText) })
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, heading).FontSize(t.Px(12)).FontWeight(500).MaxLines(1)
			if strings.TrimSpace(p.Detail) != "" {
				ui.Text(c, p.Detail).Font(t.CodeFont).FontSize(t.Px(12)).TextColor(t.TextMuted).MaxLines(1)
			}
		})
		ui.Row(c).Gap(t.Space(6)).Children(func() {
			for i, label := range p.Options {
				deny := strings.EqualFold(label, "deny")
				primary := !deny && !strings.Contains(strings.ToLower(label), "always")
				b := a.button(c, label, "", false).KeepFocus().Border(1, ui.Color{})
				if primary {
					b.Background(t.Accent.Alpha(.16)).TextColor(t.AccentText)
				} else {
					b.Background(t.overlay()).TextColor(t.TextSecondary)
				}
				if i == st.approvalCursor {
					b.Border(1, t.BorderStrong())
				}
				if b.Clicked() {
					a.respondApproval(i, false)
				}
			}
		})
	})
	if card.Shortcut(0, ui.KeyDown) && len(p.Options) > 0 {
		st.approvalCursor = (st.approvalCursor + 1) % len(p.Options)
	}
	if card.Shortcut(0, ui.KeyUp) && len(p.Options) > 0 {
		st.approvalCursor = (st.approvalCursor + len(p.Options) - 1) % len(p.Options)
	}
	if card.Shortcut(0, ui.KeyEnter) {
		a.respondApproval(st.approvalCursor, false)
	}
	if card.Shortcut(0, ui.KeyEscape) {
		a.respondApproval(0, true)
	}
}
func (a *App) askChoose(index int) {
	p, st := a.Model.Ask, &a.extensionUI
	if p == nil || p.Submitted || st.askSent || st.askCursor >= len(p.Questions) {
		return
	}
	q := p.Questions[st.askCursor]
	answer := &st.askAnswers[st.askCursor]
	if index < 0 || index > len(q.Options) || q.MultiSelect && index == len(q.Options) {
		return
	}
	st.askHighlight[st.askCursor] = index
	if q.MultiSelect {
		answer.Checked[index] = !answer.Checked[index]
	} else if index == len(q.Options) {
		answer.Kind = "custom"
	} else {
		answer.Kind = "option"
		answer.Option = index
	}
}
func (a *App) askMove(delta int) {
	p, st := a.Model.Ask, &a.extensionUI
	if p == nil || p.Submitted || st.askSent || st.askCursor >= len(p.Questions) {
		return
	}
	q := p.Questions[st.askCursor]
	n := len(q.Options)
	if !q.MultiSelect {
		n++
	}
	if n == 0 {
		return
	}
	next := (st.askHighlight[st.askCursor] + n + delta) % n
	if q.MultiSelect {
		st.askHighlight[st.askCursor] = next
	} else {
		a.askChoose(next)
	}
}
func (a *App) askAdvance() {
	p, st := a.Model.Ask, &a.extensionUI
	if p == nil || p.Submitted || st.askSent {
		return
	}
	if st.askCursor+1 < len(p.Questions) {
		st.askCursor++
		st.askFocusPending = true
		return
	}
	answers := make([]agent.AskAnswer, len(st.askAnswers))
	copy(answers, st.askAnswers)
	for i := range answers {
		answers[i].Checked = append([]bool(nil), st.askAnswers[i].Checked...)
		if answers[i].Kind == "custom" {
			answers[i].CustomLabel = a.text("view.type_something", "Type something.")
		}
	}
	st.askSent = true
	a.Dispatch(Action{Kind: "SubmitAsk", Data: map[string]any{"id": p.ID, "answers": answers}})
	a.composer.Focus()
}
func (a *App) askCancel() {
	p, st := a.Model.Ask, &a.extensionUI
	if p == nil || p.Submitted || st.askSent {
		return
	}
	st.askSent = true
	a.Dispatch(Action{Kind: "RespondDialog", Data: map[string]any{"id": p.ID, "cancelled": true}})
	a.composer.Focus()
}
func (a *App) askPanel(c *ui.Context) {
	p, t, st := a.Model.Ask, a.Theme, &a.extensionUI
	if st.askCursor >= len(p.Questions) || st.askCursor >= len(st.askAnswers) {
		return
	}
	q := p.Questions[st.askCursor]
	current := st.askHighlight[st.askCursor]
	answer := &st.askAnswers[st.askCursor]
	submitted := p.Submitted || st.askSent
	card := ui.Column(c.Key(st.askCursor)).AlignSelf(ui.Stretch).Margin(0, 0, t.Space(8)).Gap(t.Space(8)).Padding(t.Space(12)).Radius(t.Px(10)).Background(t.Elevated).Border(1, t.Accent.Alpha(.3)).Focusable().Label("Questionnaire")
	if a.reviewFull() {
		st.askFocusPending = true
	} else if st.askFocusPending && !submitted {
		card.Focus()
		st.askFocusPending = false
	}
	card.Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			a.icon(c, "icons/task.svg", 16, t.AccentText)
			if q.Header != "" {
				ui.Text(c, q.Header).FontSize(t.Px(12)).TextColor(t.TextSecondary)
			}
			ui.Spacer(c)
			if len(p.Questions) > 1 {
				ui.Text(c, strings.NewReplacer("%{current}", fmt.Sprint(st.askCursor+1), "%{total}", fmt.Sprint(len(p.Questions))).Replace(a.text("ask.question_of", "Question %{current} of %{total}"))).FontSize(t.Px(12)).TextColor(t.TextMuted)
			}
			if a.iconButton(c, "Decline questionnaire", "icons/x.svg").KeepFocus().Disabled(submitted).Clicked() {
				a.askCancel()
			}
		})
		ui.Text(c, q.Question).FontWeight(500)
		ui.Column(c).Gap(2).AlignSelf(ui.Stretch).Children(func() {
			for i, option := range q.Options {
				selected := current == i
				checked := q.MultiSelect && answer.Checked[i]
				row := ui.ButtonBase(c.Key(i)).AlignSelf(ui.Stretch).Padding(8, 10).Radius(6).KeepFocus().Label(option.Label).Disabled(submitted)
				if selected || checked {
					row.Background(t.Selection)
				}
				row.Children(func() {
					ui.Text(c, fmt.Sprintf("%d.", i+1)).TextColor(t.TextMuted)
					ui.Column(c).Grow(1).MinWidth(0).Children(func() {
						ui.Text(c, option.Label)
						if selected && option.Description != "" && !submitted {
							ui.Text(c, option.Description).FontSize(t.Px(12)).TextColor(t.TextMuted)
						}
					})
					if checked {
						a.icon(c, "icons/check.svg", 14, t.AccentText)
					}
				})
				if row.Clicked() {
					a.askChoose(i)
				}
			}
			if !q.MultiSelect && a.button(c, a.text("view.type_something", "Type something."), "icons/compose.svg", current == len(q.Options)).KeepFocus().AlignSelf(ui.Stretch).Disabled(submitted).Clicked() {
				a.askChoose(len(q.Options))
			}
		})
		if !q.MultiSelect && current < len(q.Options) && q.Options[current].Preview != "" {
			preview := []rune(q.Options[current].Preview)
			text := q.Options[current].Preview
			if len(preview) > 600 {
				text = string(preview[:600]) + "…"
			}
			ui.Column(c.Key("option-preview")).AlignSelf(ui.Stretch).Border(1, t.Border).Background(t.Background).Radius(8).Padding(8, 10).Gap(6).Children(func() {
				ui.Text(c, strings.ReplaceAll(a.text("ask.preview", "Preview · %{label}"), "%{label}", q.Options[current].Label)).FontSize(t.Px(11)).TextColor(t.TextMuted)
				ui.Text(c, text).Font(t.CodeFont).FontSize(11 * t.CodeSize / 13).FixedLineHeight(16 * t.CodeSize / 13).TextColor(t.TextSecondary)
			})
		}
		if q.MultiSelect || current == len(q.Options) {
			input := ui.TextArea(c.Key("custom-answer"), &answer.Custom).Lines(1, 3).Label("Custom answer").Placeholder(a.text("ask.type_your_answer", "Type your answer…")).AlignSelf(ui.Stretch).Disabled(submitted).OnTextKey(func(k ui.TextKey) bool {
				if composing, ok := any(k).(interface{ Composing() bool }); ok && composing.Composing() {
					return true
				}
				if k.Key == ui.KeyEnter && k.Mods == 0 {
					answer.Custom = k.Text()
					a.askAdvance()
					return true
				}
				if k.Key == ui.KeyEscape {
					a.askCancel()
					return true
				}
				return false
			})
			if !a.reviewFull() && !q.MultiSelect && !submitted {
				input.AutoFocus()
			}
		}
		ui.Row(c).AlignItems(ui.Center).AlignSelf(ui.Stretch).Height(46).Gap(8).Children(func() {
			hint := a.text("ask.hint_single", "↑↓ Navigate · ⏎ Next · esc Cancel")
			if q.MultiSelect {
				hint = a.text("ask.hint_multi", "↑↓ Navigate · ⏎ Toggle · esc Cancel")
			}
			if submitted {
				hint = a.text("ask.sending", "Sending your answers…")
			}
			ui.Text(c, hint).Grow(1).MinWidth(0).FontSize(t.Px(12)).TextColor(t.TextMuted)
			if a.button(c, a.text("view.back", "Back"), "", false).Label("Previous question").KeepFocus().Disabled(submitted || st.askCursor == 0).Clicked() {
				st.askCursor--
				st.askFocusPending = true
			}
			label := a.text("ask.next", "Next")
			if st.askCursor+1 == len(p.Questions) {
				label = a.text("ask.submit", "Submit")
			}
			if a.button(c, label, "", true).KeepFocus().Disabled(submitted).Clicked() {
				a.askAdvance()
			}
		})
	})
	if submitted {
		return
	}
	if card.Shortcut(0, ui.KeyDown) {
		a.askMove(1)
	}
	if card.Shortcut(0, ui.KeyUp) {
		a.askMove(-1)
	}
	if card.Shortcut(0, ui.KeyEnter) || card.Shortcut(0, ui.KeySpace) {
		if q.MultiSelect {
			a.askChoose(st.askHighlight[st.askCursor])
		} else {
			a.askAdvance()
		}
	}
	if card.Shortcut(0, ui.KeyEscape) {
		a.askCancel()
	}
}
func (a *App) extensionWidgets(c *ui.Context, placement string) {
	var widgets []agent.ExtensionWidget
	for _, widget := range a.Model.ExtensionWidgets {
		if widget.Placement == placement && len(widget.Lines) != 0 {
			widgets = append(widgets, widget)
		}
	}
	if len(widgets) == 0 {
		return
	}
	// The source groups each placement with Base06 gaps, but adds no outer
	// composer gap. Typography follows the editor preference, not UI size.
	ui.Column(c.Key("extension-widgets:" + placement)).AlignSelf(ui.Stretch).Gap(a.Theme.Space(6)).Children(func() {
		for _, widget := range widgets {
			ui.Column(c.Key("extension-widget:"+widget.Key)).AlignSelf(ui.Stretch).Padding(a.Theme.Space(8), a.Theme.Space(12)).Gap(a.Theme.Space(2)).Background(a.Theme.CodeBg).Border(1, a.Theme.Border).Radius(a.Theme.Px(10)).Label("Extension widget " + widget.Key).Children(func() {
				for i, line := range widget.Lines {
					ui.RichText(c.Key(i), widgetSpans(line, a.Theme)...).Font(a.Theme.CodeFont).FontSize(12 * a.Theme.CodeSize / 13).FixedLineHeight(18 * a.Theme.CodeSize / 13).Selectable()
				}
			})
		}
	})
}

// Whole-text extension replacement owns a new editor identity. Retain focus
// only when the old editor had it; changing text is not a request to steal it.
func (a *App) trackExtensionComposer(c *ui.Context, input ui.Element) {
	if a.reviewFull() {
		return
	}
	st := &a.extensionUI
	if st.composerPending {
		st.composerPending = false
		if st.composerWasFocused {
			input.Focus()
		}
	}
	st.composerWasFocused = input.Focused()
}

func (a *App) extensionStatus(c *ui.Context) {
	if extensionStatusVisible(a.Model.ExtensionStatusAt, time.Now()) && a.Model.ExtensionStatusText != "" {
		ui.Text(c, statusPlainText(a.Model.ExtensionStatusText)).MaxWidth(a.Theme.Px(320)).MaxLines(1).TextColor(a.Theme.TextMuted).FontSize(a.Theme.Px(12))
	}
}
func extensionStatusVisible(at, now time.Time) bool {
	return !at.IsZero() && now.Sub(at) < 6*time.Second
}
