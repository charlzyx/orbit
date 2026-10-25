package orbitui

import (
	"fmt"
	"strings"

	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/egoist/mygo/ui"
)

func (a *App) modal(c *ui.Context, open *bool, width float32, fn func()) {
	t := a.Theme
	w, h := c.Size()
	ui.DialogBase(c, open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, .45))
		panel.Width(min(width, max(0, w-32))).MaxWidth(max(0, w-32)).MaxHeight(min(680, max(0, h-32))).Padding(20).Gap(12).Background(t.Elevated).Border(1, t.Border).Radius(12).Children(fn)
	})
}

// Pi owns all keyboard input in a custom component, including Escape. Unlike
// DialogBase this native modal only dismisses via its scrim or explicit Cancel.
func (a *App) customModal(c *ui.Context, open *bool, width float32, fn func()) {
	t := a.Theme
	w, h := c.Size()
	ui.Overlay(c, func() {
		back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Center().Modal().Background(ui.RGBA(0, 0, 0, .45))
		if back.Clicked() {
			*open = false
		}
		back.Children(func() {
			panel := ui.Box(c).Role(ui.RoleDialog).Width(min(width, max(0, w-32))).MaxWidth(max(0, w-32)).MaxHeight(min(680, max(0, h-32))).Padding(20).Gap(12).Background(t.Elevated).Border(1, t.Border).Radius(12)
			// Mark the panel clickable to keep inside clicks from hitting the scrim.
			_ = panel.Clicked()
			panel.Children(fn)
		})
	})
}
func (a *App) commandPalette(c *ui.Context) {
	a.modal(c.Key("command-palette"), &a.paletteOpen, 620, func() {
		type result struct{ title, icon, command, value string }
		var results []result
		input := ui.TextInput(c.Key("palette-search"), &a.paletteQuery).Placeholder("Search sessions, commands and settings…").Label("Command palette search").AutoFocus().AlignSelf(ui.Stretch).OnTextKey(func(k ui.TextKey) bool {
			if len(results) == 0 {
				return false
			}
			switch k.Key {
			case ui.KeyDown:
				a.commandCursor = (a.commandCursor + 1) % len(results)
				a.paletteList.ScrollTo(a.commandCursor, ui.Center)
				return true
			case ui.KeyUp:
				a.commandCursor = (a.commandCursor + len(results) - 1) % len(results)
				a.paletteList.ScrollTo(a.commandCursor, ui.Center)
				return true
			case ui.KeyEnter:
				if a.commandCursor >= 0 && a.commandCursor < len(results) {
					r := results[a.commandCursor]
					a.paletteOpen = false
					a.command(r.command, r.value)
				}
				return true
			}
			return false
		})
		if input.Changed() {
			a.commandCursor = 0
		}
		q := strings.ToLower(a.paletteQuery)
		for _, s := range a.Model.Sessions {
			if q == "" || strings.Contains(strings.ToLower(s.Name+" "+s.Summary), q) {
				results = append(results, result{s.Name, "icons/chat.svg", "SwitchSession", s.Path})
			}
		}
		for _, cmd := range Commands {
			if cmd.ID == "OpenSessionSlot" || cmd.Category == "Review" {
				continue
			}
			title := a.Tr(cmd.TitleKey)
			if q == "" || strings.Contains(strings.ToLower(title+" "+cmd.Keywords), q) {
				results = append(results, result{title, cmd.Icon, cmd.ID, ""})
			}
		}
		a.commandCursor = max(0, min(a.commandCursor, len(results)-1))
		a.paletteList.Selected = &a.commandCursor
		a.paletteList.Label = func(i int) string { return results[i].title }
		list := ui.List(c, &a.paletteList, len(results), func(i int) {
			r := results[i]
			b := a.button(c, r.title, r.icon, i == a.commandCursor).AlignSelf(ui.Stretch).Height(36)
			if b.Clicked() {
				a.command(r.command, r.value)
				a.paletteOpen = false
			}
		}).Height(420)
		if list.Submitted() && a.commandCursor >= 0 && a.commandCursor < len(results) {
			r := results[a.commandCursor]
			a.command(r.command, r.value)
			a.paletteOpen = false
		}
	})
}
func (a *App) modelPicker(c *ui.Context) {
	a.modal(c.Key("model-picker"), &a.modelOpen, 620, func() {
		ui.Text(c, "Choose model").FontSize(18).FontWeight(600)
		var rows []ModelChoice
		input := ui.TextInput(c, &a.modelQuery).Placeholder("Search models and providers…").Label("Search models").AutoFocus().AlignSelf(ui.Stretch).OnTextKey(func(k ui.TextKey) bool {
			if len(rows) == 0 {
				return false
			}
			switch k.Key {
			case ui.KeyDown:
				a.modelCursor = (a.modelCursor + 1) % len(rows)
				a.modelList.ScrollTo(a.modelCursor, ui.Center)
				return true
			case ui.KeyUp:
				a.modelCursor = (a.modelCursor + len(rows) - 1) % len(rows)
				a.modelList.ScrollTo(a.modelCursor, ui.Center)
				return true
			case ui.KeyEnter:
				if a.modelCursor >= 0 && a.modelCursor < len(rows) {
					m := rows[a.modelCursor]
					a.Dispatch(Action{Kind: "SetModel", Value: m.ID, Data: map[string]any{"provider": m.Provider}})
					a.modelOpen = false
				}
				return true
			}
			return false
		})
		if input.Changed() {
			a.modelCursor = 0
		}
		for _, m := range a.Model.Models {
			if a.modelQuery == "" || strings.Contains(strings.ToLower(m.Name+" "+m.Provider+" "+m.ID), strings.ToLower(a.modelQuery)) {
				rows = append(rows, m)
			}
		}
		a.modelCursor = max(0, min(a.modelCursor, len(rows)-1))
		if len(rows) == 0 {
			ui.Text(c, "No models available. Configure a provider or connect pi.").TextColor(a.Theme.TextMuted)
		}
		a.modelList.Selected = &a.modelCursor
		a.modelList.Label = func(i int) string { return rows[i].Name }
		ui.List(c, &a.modelList, len(rows), func(i int) {
			m := rows[i]
			name := m.Name
			if name == "" {
				name = m.ID
			}
			b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Padding(10).Radius(6).Label(name)
			if i == a.modelCursor {
				b.Background(a.Theme.Selection)
			} else if b.Hovered() {
				b.Background(a.Theme.Hover)
			}
			b.Children(func() {
				ui.Column(c).Grow(1).Children(func() { ui.Text(c, name); ui.Text(c, m.Provider+" / "+m.ID).FontSize(12).TextColor(a.Theme.TextMuted) })
				if m.ID == a.Model.ModelID && m.Provider == a.Model.ModelProvider {
					a.icon(c, "icons/check.svg", 16, a.Theme.Accent)
				}
			})
			if b.Clicked() {
				a.Dispatch(Action{Kind: "SetModel", Value: m.ID, Data: map[string]any{"provider": m.Provider}})
				a.modelOpen = false
			}
		}).Height(400)
	})
}
func (a *App) extensionDialog(c *ui.Context) {
	d := a.Model.Dialog
	open := true
	width := float32(560)
	modal := a.modal
	if d.Method == "custom" {
		width = 960
		modal = a.customModal
	}
	modal(c.Key("extension-dialog:"+d.ID), &open, width, func() {
		ui.Text(c, d.Title).FontSize(18).FontWeight(600)
		if d.Message != "" {
			ui.Text(c, d.Message).Selectable()
		}
		switch d.Method {
		case "custom":
			if a.Options.Custom != nil {
				_, height := c.Size()
				orbiterm.CustomView(c, a.Options.Custom).Height(min(400, max(120, height-180))).AlignSelf(ui.Stretch)
			} else {
				ui.Text(c, "Native extension UI renderer is not connected.").TextColor(a.Theme.Error)
			}
		case "select":
			for _, choice := range d.Options {
				if a.button(c, choice, "", false).AlignSelf(ui.Stretch).Clicked() {
					a.Dispatch(Action{Kind: "RespondDialog", Value: choice, Data: map[string]any{"id": d.ID}})
				}
			}
		case "input":
			ui.TextInput(c, &a.dialogValue).Label(d.Title).Placeholder(d.Placeholder).AutoFocus().AlignSelf(ui.Stretch)
		case "editor":
			ui.TextArea(c, &a.dialogValue).Label(d.Title).Placeholder(d.Placeholder).Height(240).AutoFocus().AlignSelf(ui.Stretch)
		}
		ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
			if a.button(c, "Cancel", "", false).Clicked() {
				open = false
			}
			if d.Method == "confirm" {
				if a.button(c, "Confirm", "icons/check.svg", true).Clicked() {
					a.Dispatch(Action{Kind: "RespondDialog", Data: map[string]any{"id": d.ID, "confirmed": true}})
				}
			} else if d.Method == "input" || d.Method == "editor" {
				if a.button(c, "Submit", "", true).Clicked() {
					a.Dispatch(Action{Kind: "RespondDialog", Value: a.dialogValue, Data: map[string]any{"id": d.ID}})
				}
			}
		})
	})
	if !open {
		a.Dispatch(Action{Kind: "RespondDialog", Data: map[string]any{"id": d.ID, "cancelled": true}})
	}
}
func (a *App) prompt(c *ui.Context) {
	if !a.promptOpen {
		return
	}
	a.modal(c.Key("text-prompt"), &a.promptOpen, 460, func() {
		ui.Text(c, a.promptTitle).FontSize(18).FontWeight(600)
		in := ui.TextInput(c, &a.promptValue).Label(a.promptTitle).AutoFocus().AlignSelf(ui.Stretch)
		submit := in.Submitted()
		ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
			if a.button(c, "Cancel", "", false).Clicked() {
				a.promptOpen = false
			}
			if a.button(c, "Save", "", true).Clicked() {
				submit = true
			}
		})
		if submit && strings.TrimSpace(a.promptValue) != "" {
			f := a.promptCommit
			a.promptOpen = false
			if f != nil {
				f(a.promptValue)
			}
		}
	})
}
func (a *App) Prompt(title, value string, commit func(string)) {
	a.promptTitle = title
	a.promptValue = value
	a.promptCommit = commit
	a.promptOpen = true
}
func (a *App) Confirm(title, detail string, commit func()) {
	a.confirmTitle = title
	a.confirmMessage = detail
	a.confirmCommit = commit
	a.confirmOpen = true
}
func (a *App) confirm(c *ui.Context) {
	if !a.confirmOpen {
		return
	}
	a.modal(c.Key("confirmation"), &a.confirmOpen, 480, func() {
		ui.Text(c, a.confirmTitle).FontSize(18).FontWeight(600)
		ui.Text(c, a.confirmMessage)
		ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
			if a.button(c, "Cancel", "", false).Clicked() {
				a.confirmOpen = false
			}
			if a.button(c, "Continue", "", true).Clicked() {
				f := a.confirmCommit
				a.confirmOpen = false
				if f != nil {
					f()
				}
			}
		})
	})
}
func (a *App) commandSession(id string, s SessionRow) {
	switch id {
	case "RenameSession":
		a.Prompt("Rename task", s.Name, func(name string) {
			a.Dispatch(Action{Kind: id, Value: name, SessionID: s.ID, Data: map[string]any{"path": s.Path}})
		})
	case "DeleteSession":
		a.Confirm("Delete task?", fmt.Sprintf("Delete %q and its saved session file?", s.Name), func() { a.Dispatch(Action{Kind: id, Value: s.Path, SessionID: s.ID}) })
	default:
		a.Dispatch(Action{Kind: id, Value: s.Path, SessionID: s.ID})
	}
}
