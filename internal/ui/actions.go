package orbitui

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

func keyChord(chord string) (ui.Modifiers, ui.Key, bool) {
	var mods ui.Modifiers
	parts := strings.Split(chord, "-")
	for _, p := range parts[:len(parts)-1] {
		switch p {
		case "secondary":
			mods |= ui.Cmd
		case "ctrl":
			mods |= ui.Ctrl
		case "shift":
			mods |= ui.Shift
		case "alt":
			mods |= ui.Alt
		case "cmd":
			mods |= ui.Super
		default:
			return 0, 0, false
		}
	}
	s := parts[len(parts)-1]
	if len(s) == 1 && s[0] >= 'a' && s[0] <= 'z' {
		return mods, ui.KeyA + ui.Key(s[0]-'a'), true
	}
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return mods, ui.Key0 + ui.Key(s[0]-'0'), true
	}
	k, ok := map[string]ui.Key{"enter": ui.KeyEnter, "escape": ui.KeyEscape, "space": ui.KeySpace, "tab": ui.KeyTab, "up": ui.KeyUp, "down": ui.KeyDown, "left": ui.KeyLeft, "right": ui.KeyRight, ".": ui.KeyPeriod, ",": ui.KeyComma, "/": ui.KeySlash, "[": ui.KeyBracketLeft, "]": ui.KeyBracketRight, "{": ui.KeyBracketLeft, "}": ui.KeyBracketRight}[s]
	if s == "{" || s == "}" {
		mods |= ui.Shift
	}
	return mods, k, ok
}
func shortcutLabel(chord string) string {
	if runtime.GOOS == "darwin" {
		r := strings.NewReplacer("secondary-", "⌘", "ctrl-", "⌃", "shift-", "⇧", "alt-", "⌥", "enter", "↵", "escape", "Esc", "up", "↑", "down", "↓", "tab", "⇥", "{", "⇧[", "}", "⇧]")
		return r.Replace(strings.ToLower(chord))
	}
	r := strings.NewReplacer("secondary-", "Ctrl+", "ctrl-", "Ctrl+", "shift-", "Shift+", "alt-", "Alt+")
	return r.Replace(chord)
}
func (a *App) shortcuts(c *ui.Context) {
	if a.modalOpen() {
		return
	}

	if a.page == "git" && c.Shortcut(0, ui.KeyEscape) {
		if a.Options.Escape == nil || !a.Options.Escape() {
			a.Navigate("chat")
		}
		return
	}
	if a.ReviewVisible && a.reviewState.handle.FocusWithin(c) {
		for _, b := range Bindings {
			if b.Context != "ReviewDiffTree" {
				continue
			}
			mods, key, ok := keyChord(b.Chord)
			if ok && c.Shortcut(mods, key) {
				a.reviewCommand(b.Command)
				return
			}
		}
	}
	for _, b := range Bindings {
		if b.Context != "" {
			continue
		}
		mods, key, ok := keyChord(b.Chord)
		if !ok {
			continue
		}
		if c.Shortcut(mods, key) {
			value := ""
			if b.Command == "OpenSessionSlot" {
				value = b.Chord[strings.LastIndexByte(b.Chord, '-')+1:]
			}
			a.command(b.Command, value)
		}
	}
	if a.thinkingOpen {
		a.thinkingOpen = false
		a.Prompt("Thinking level", a.Model.Thinking, func(v string) {
			for _, valid := range a.Model.ThinkingLevels {
				if v == valid {
					a.Dispatch(Action{Kind: "SetThinking", Value: v})
					return
				}
			}
			a.Notify("This thinking level is not available for the current model", true)
		})
	}
}
func (a *App) modalOpen() bool {
	return a.promptOpen || a.confirmOpen || a.Model.Dialog != nil || a.paletteOpen || a.modelOpen || a.jsonOpen || a.secretOpen || a.imageOpen
}

// Native menu accelerators arrive outside MyGo's modal shortcut scope. Apply
// the same guard before routing them to normal UI commands.
func (a *App) nativeCommand(id string) {
	if id != "Quit" && a.modalOpen() {
		return
	}
	a.command(id, "")
}
func (a *App) command(id, value string) {
	if strings.HasPrefix(id, "Review") && id != "ReviewChanges" {
		a.reviewCommand(id)
		return
	}
	switch id {
	case "NewSession":
		a.drafts[a.Model.SessionID] = draftState{a.Draft, append([]attachment(nil), a.attachments...)}
		a.Draft = ""
		a.attachments = nil
		a.pendingNew = true
		a.Navigate("chat")
		a.Dispatch(Action{Kind: id, Value: a.Model.Workspace})
		a.composer.Focus()
	case "SwitchSession":
		a.Navigate("chat")
		a.Dispatch(Action{Kind: id, Value: value})
	case "ToggleSidebar":
		a.SidebarVisible = !a.SidebarVisible
	case "FocusSessions":
		a.SidebarVisible = true
		a.sessionList.Handle.Focus()
	case "FocusComposer":
		a.Navigate("chat")
		a.composer.Focus()
	case "ToggleCommandPalette":
		a.paletteOpen = !a.paletteOpen
		a.paletteQuery = ""
		a.commandCursor = 0
	case "ToggleSearch":
		a.searchOpen = !a.searchOpen
		a.Navigate("chat")
	case "OpenSettings":
		a.Navigate("settings")
	case "OpenAbout":
		a.settingsPage = "about"
		a.Navigate("settings")
	case "OpenShortcutHelp":
		a.settingsPage = "shortcuts"
		a.Navigate("settings")
	case "ToggleUsage":
		if a.page == "usage" {
			a.Navigate("chat")
		} else {
			a.Navigate("usage")
		}
	case "OpenGit":
		a.Navigate("git")
	case "OpenFiles":
		a.Navigate("files")
	case "OpenWorktrees":
		a.Navigate("worktrees")
	case "ToggleProjectPanel":
		a.ExplorerVisible = !a.ExplorerVisible
		a.ReviewVisible = false
		a.Dispatch(Action{Kind: "RefreshFiles"})
	case "ReviewChanges", "ToggleSidePanel":
		a.ReviewVisible = !a.ReviewVisible
		a.ExplorerVisible = false
		a.Dispatch(Action{Kind: "RefreshReview"})
	case "ReviewClose":
		a.ReviewVisible = false
		a.composer.Focus()
	case "ToggleTerminal":
		if a.Options.Terminal != nil {
			a.save(a.Options.Terminal.Toggle())
		} else {
			a.Dispatch(Action{Kind: id})
		}
	case "NewTerminalTab":
		if a.Options.Terminal != nil {
			a.save(a.Options.Terminal.NewTab())
		}
	case "CloseTerminalTab":
		if a.Options.Terminal != nil {
			a.save(a.Options.Terminal.CloseActive())
		}
	case "NextTerminalTab":
		if a.Options.Terminal != nil {
			a.Options.Terminal.Cycle(1)
		}
	case "PrevTerminalTab":
		if a.Options.Terminal != nil {
			a.Options.Terminal.Cycle(-1)
		}
	case "ChooseModel":
		a.modelOpen = true
	case "ChooseThinking":
		a.thinkingOpen = true
	case "ChooseAccess":
		a.Prompt("Access mode (supervised, auto-accept-edits, full-access)", a.access, func(v string) {
			if e := a.Options.Store.SaveAccess(v); e != nil {
				a.Notify(e.Error(), true)
				return
			}
			a.access = v
			a.Dispatch(Action{Kind: "SetAccess", Value: v})
		})
	case "ChooseWorkflowMode":
		a.Prompt("Workflow mode (plan, build, ask)", a.workflow, func(v string) {
			if e := a.Options.Store.SaveWorkflow(a.Model.SessionID, v); e != nil {
				a.Notify(e.Error(), true)
				return
			}
			a.workflow = v
			a.Dispatch(Action{Kind: "SetWorkflow", Value: v})
		})
	case "RenameSession":
		a.Prompt("Rename task", a.Model.SessionName, func(v string) { a.Dispatch(Action{Kind: id, Value: v}) })
	case "DeleteSession":
		a.Confirm("Delete task?", "This removes the saved session file. The active agent will be stopped.", func() { a.Dispatch(Action{Kind: id, Value: a.Model.SessionPath}) })
	case "CopySessionId":
		a.Dispatch(Action{Kind: "CopyText", Value: a.Model.SessionID})
	case "CopyLastResponse":
		for i := len(a.Model.Messages) - 1; i >= 0; i-- {
			m := a.Model.Messages[i]
			if m.Role == "assistant" && m.Text != "" {
				a.Dispatch(Action{Kind: "CopyText", Value: m.Text})
				break
			}
		}
	case "OpenSessionSlot":
		n, e := strconv.Atoi(value)
		if e == nil && n > 0 && n <= len(a.Model.Sessions) {
			a.Dispatch(Action{Kind: "SwitchSession", Value: a.Model.Sessions[n-1].Path})
		}
	case "NextSession", "PrevSession":
		if len(a.Model.Sessions) == 0 {
			return
		}
		i := 0
		for ix, s := range a.Model.Sessions {
			if s.ID == a.Model.SessionID {
				i = ix
				break
			}
		}
		delta := 1
		if id == "PrevSession" {
			delta = -1
		}
		i = (i + delta + len(a.Model.Sessions)) % len(a.Model.Sessions)
		a.Dispatch(Action{Kind: "SwitchSession", Value: a.Model.Sessions[i].Path})
	case "PrevTurn", "NextTurn":
		delta := 1
		if id == "PrevTurn" {
			delta = -1
		}
		first, _ := a.transcript.Visible()
		for i := first + delta; i >= 0 && i < len(a.Model.Messages); i += delta {
			if a.Model.Messages[i].Role == "user" {
				a.transcript.ScrollTo(i, ui.Start)
				break
			}
		}
	case "OpenMcpSettings":
		a.settingsPage = "mcp"
		a.Navigate("settings")
		a.loadSettings("mcp")
	default:
		if strings.HasPrefix(id, "GitTab") {
			a.Navigate("git")
		}
		a.Dispatch(Action{Kind: id, Value: value})
	}
}
