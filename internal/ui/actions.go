package orbitui

import (
	"runtime"
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
	if a.composerMode != nil && c.Shortcut(0, ui.KeyEscape) {
		a.closeComposerModeMenu(true)
		return
	}
	if a.files != nil && a.files.entry != nil && a.files.entry.input.FocusWithin(c) {
		if c.Shortcut(0, ui.KeyEscape) {
			a.cancelExplorerEntry()
		}
		return
	}
	if a.modalOpen() {
		return
	}
	if a.surfaceTraversal(c) {
		return
	}
	if a.sidebarShortcut(c) {
		return
	}

	if a.autocomplete.IsOpen() && a.composer.Focused(c) && c.Shortcut(0, ui.KeyEscape) {
		a.autocomplete.Dismiss()
		return
	}
	if a.searchOpen && a.searchInput.Focused(c) && c.Shortcut(0, ui.KeyEscape) {
		a.closeFind()
		return
	}
	if a.page == "git" && c.Shortcut(0, ui.KeyEscape) {
		if a.Options.Escape == nil || !a.Options.Escape() {
			a.Navigate("chat")
		}
		return
	}
	if a.ReviewVisible && a.reviewState.treeHandle.FocusWithin(c) {
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
			a.shortcutCommand(b.Command, value)
		}
	}
	if a.thinkingOpen && a.Models == nil {
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
	return a.composerMode != nil || a.workspaceActions.open || a.nativeSettings.selectMenu.id != "" || a.providerState.usageProvider != "" || (a.Models != nil && a.Models.HasModal()) || a.workspaceOpen || a.usageState.filterMenu != "" || a.providerState.editor != nil || (a.Options.HasModal != nil && a.Options.HasModal()) || a.promptOpen || a.confirmOpen || a.Model.Dialog != nil || a.paletteOpen || a.modelOpen || a.jsonOpen || a.secretOpen || a.imageOpen || a.usageState.detailError != nil || a.usageState.customOpen || a.usageState.detailRecord != nil || a.usageState.detailSession != nil
}

// Native menu accelerators arrive outside MyGo's modal shortcut scope. Apply
// the same guard before routing them to normal UI commands.
func (a *App) nativeCommand(id string) {
	if id != "Quit" && a.modalOpen() {
		return
	}
	a.shortcutCommand(id, "")
}

// The source OpenGit keyboard action toggles, whereas the command palette's
// destination opens idempotently (session.rs::on_open_git vs pickers.rs).
func (a *App) shortcutCommand(id, value string) {
	if id == "OpenGit" {
		id = "ToggleGit"
	}
	a.command(id, value)
}

func (a *App) command(id, value string) {
	if strings.HasPrefix(id, "Review") && id != "ReviewChanges" {
		a.reviewCommand(id)
		return
	}
	switch id {
	case "AbortRun":
		// Pinned session.rs::on_abort closes feature pages before touching Pi.
		if a.page == "settings" || a.page == "worktrees" {
			a.Navigate("chat")
			a.composer.Focus()
			return
		}
		a.Dispatch(Action{Kind: id, Value: value})
	case "NewSession":
		if !a.admitNewTask() {
			return
		}
		a.drafts[a.Model.SessionID] = draftState{a.Draft, append([]attachment(nil), a.attachments...)}
		a.Draft = ""
		a.attachments = nil
		a.pendingNew = true
		a.Navigate("chat")
		a.composer.Focus()
	case "SwitchSession":
		a.Navigate("chat")
		a.Dispatch(Action{Kind: id, Value: value})
	case "ToggleSidebar":
		a.SidebarVisible = !a.SidebarVisible
	case "FocusSessions":
		a.focusSidebarSessions()
	case "FocusComposer":
		a.Navigate("chat")
		a.composer.Focus()
	case "ToggleCommandPalette":
		a.paletteOpen = !a.paletteOpen
		a.paletteQuery = ""
		a.commandCursor = 0
	case "ToggleSearch":
		a.searchOpen = !a.searchOpen
		a.Search = ""
		a.searchCursor = 0
		a.searchLastQuery = ""
		a.searchLastCount = 0
		a.Navigate("chat")
		if a.searchOpen {
			a.searchInput.Focus()
		} else {
			a.composer.Focus()
		}
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
		a.closeReviewDock()
		a.Navigate("git")
	case "ToggleGit":
		if a.page == "git" {
			a.Navigate("chat")
		} else {
			a.command("OpenGit", "")
		}
	case "OpenFiles":
		a.Navigate("files")
	case "OpenWorktrees":
		a.Navigate("worktrees")
	case "ToggleProjectPanel":
		a.ExplorerVisible = !a.ExplorerVisible
		if a.ExplorerVisible {
			a.closeReviewDock()
			a.Dispatch(Action{Kind: "RefreshFiles"})
		}
	case "ReviewChanges":
		a.openReviewDock()
	case "ToggleSidePanel":
		if a.ReviewVisible {
			// SidePane::toggle preserves its full-width state; explicit close
			// and switching to another dock use closeReviewDock instead.
			a.ReviewVisible = false
			a.reviewState.sourceMenuOpen = false
		} else {
			a.openReviewDock()
		}
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
		a.closeComposerModeMenu(false)
		a.modelOpen = true
	case "ChooseThinking":
		a.closeComposerModeMenu(false)
		a.thinkingOpen = true
	case "ChooseAccess":
		a.toggleComposerModeMenu("access")
	case "ChooseWorkflowMode":
		a.toggleComposerModeMenu("workflow")
	case "PinSession":
		// Palette commands have no row payload. Pin the active saved path,
		// never the empty value used by the command registry result.
		if a.Model.SessionPath != "" {
			a.Dispatch(Action{Kind: id, Value: a.Model.SessionPath, SessionID: a.Model.SessionID})
		}
	case "RenameSession":
		a.promptPaletteRename()
	case "DeleteSession":
		a.confirmPaletteDelete()
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
		a.openSidebarSessionSlot(value)
	case "NextSession":
		a.cycleSidebarSession(1)
	case "PrevSession":
		a.cycleSidebarSession(-1)
	case "PrevTurn", "NextTurn":
		delta := 1
		if id == "PrevTurn" {
			delta = -1
		}
		a.jumpTurn(delta)
		a.dismissRailHint()
	case "OpenMcpSettings":
		a.settingsPage = "mcp"
		a.Navigate("settings")
		a.loadSettings("mcp")
	default:
		if strings.HasPrefix(id, "GitTab") && a.page != "git" {
			return
		}
		a.Dispatch(Action{Kind: id, Value: value})
	}
}
