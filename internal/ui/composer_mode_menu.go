package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"strings"
)

// A mode picker is bound to its opening task/intent, not a later rendered chip.
// Access is still global policy; ownership only guards the user interaction.
type composerModeMenu struct {
	kind    string
	cursor  int
	hovered int
	owner   attachmentOwner
	target  *agent.WorkflowTarget
	epoch   uint64
}

func composerModeValues(kind string) []string {
	switch kind {
	case "access":
		return []string{"supervised", "auto-accept-edits", "full-access"}
	case "workflow":
		return []string{"plan", "build", "ask"}
	}
	return nil
}
func (a *App) composerModeLabel(kind, value string) string {
	if kind == "workflow" {
		return a.workflowDisplay(value)
	}
	fallback := map[string]string{"supervised": "Supervised", "auto-accept-edits": "Auto-accept edits", "full-access": "Full access"}[value]
	return a.accessDisplay(value, fallback)
}
func (a *App) closeComposerModeMenu(focus bool) {
	if a.composerMode == nil {
		return
	}
	a.composerMode = nil
	a.composerModeEpoch++
	if focus {
		a.composer.Focus()
	}
}
func (a *App) pruneComposerModeMenu() {
	if s := a.composerMode; s != nil && !a.ownsAttachment(s.owner) {
		a.closeComposerModeMenu(false)
	}
}
func (a *App) toggleComposerModeMenu(kind string) {
	values := composerModeValues(kind)
	if len(values) == 0 {
		return
	}
	a.pruneComposerModeMenu()
	if a.composerMode != nil && a.composerMode.kind == kind {
		a.closeComposerModeMenu(true)
		return
	}
	owner := a.attachmentOwner()
	if !a.ownsAttachment(owner) {
		return
	}
	var target *agent.WorkflowTarget
	if kind == "workflow" && a.Options.CaptureWorkflowTarget != nil {
		t, err := a.Options.CaptureWorkflowTarget(a.Model.RuntimeID, a.Model.SessionID, a.Model.SessionPath, a.Model.Workspace)
		if err != nil {
			a.Notify(err.Error(), true)
			return
		}
		target = &t
	}
	a.closeComposerModeMenu(false)
	a.composerModeEpoch++ // invalidates native Add menu held before opening this picker
	s := &composerModeMenu{kind: kind, owner: owner, target: target, epoch: a.composerModeEpoch, hovered: -1}
	current := a.access
	if kind == "workflow" {
		current = a.workflow
	}
	if current == "full" {
		current = "full-access"
	}
	for i, v := range values {
		if v == current {
			s.cursor = i
		}
	}
	a.modelOpen, a.thinkingOpen, a.workspaceOpen, a.contextOpen = false, false, false, false
	if a.Models != nil {
		a.Models.DismissPicker(false)
	}
	a.autocomplete.Dismiss()
	a.composerMode = s
}
func (a *App) chooseComposerMode(s *composerModeMenu, value string) {
	if s == nil || a.composerMode != s || !a.ownsAttachment(s.owner) {
		a.pruneComposerModeMenu()
		return
	}
	valid := false
	for _, v := range composerModeValues(s.kind) {
		if value == v {
			valid = true
		}
	}
	if !valid {
		return
	}
	kind := "SetAccess"
	if s.kind == "workflow" {
		kind = "SetWorkflow"
	}
	a.closeComposerModeMenu(true)
	// Projection is exclusively from SetSnapshot after backend admission/success.
	// In particular neither a rejected maintenance request nor a failed write may
	// temporarily advertise a new global access policy.
	a.Dispatch(Action{Kind: kind, Value: value, workflowTarget: s.target})
}
func (a *App) composerModeTrigger(c *ui.Context, anchor ui.Element, kind string) {
	// Full Review retains an invisible composer for native undo. Popovers are
	// overlays, so never mount one from that retained-but-hidden subtree.
	if a.page != "chat" || a.reviewFull() || a.setupBlocked() {
		return
	}
	if anchor.Clicked() {
		a.toggleComposerModeMenu(kind)
	}
	a.pruneComposerModeMenu()
	s := a.composerMode
	if s == nil || s.kind != kind {
		return
	}
	values := composerModeValues(kind)
	titleKey := "command_palette.choose_access"
	if kind == "workflow" {
		titleKey = "command_palette.choose_workflow_mode"
	}
	open := true
	popup := ui.PopoverBase(c.Key(struct {
		Kind  string
		Epoch uint64
	}{"composer-mode-" + kind, s.epoch}), anchor, &open, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 5, 0).Column().Width(300).Padding(6).Gap(2).Background(a.Theme.Elevated).Border(1, a.Theme.Border).Radius(8).Label(a.Tr(titleKey))
		choices := ui.Column(c.Key("mode-options")).Focusable().AutoFocus().AlignSelf(ui.Stretch)
		if choices.Shortcut(0, ui.KeyDown) {
			s.cursor = (s.cursor + 1) % len(values)
		}
		if choices.Shortcut(0, ui.KeyUp) {
			s.cursor = (s.cursor + len(values) - 1) % len(values)
		}
		if choices.Shortcut(0, ui.KeyEnter) {
			a.chooseComposerMode(s, values[s.cursor])
		}
		if choices.Shortcut(0, ui.KeyEscape) {
			a.closeComposerModeMenu(true)
		}
		hovered := -1
		choices.Children(func() {
			for i, v := range values {
				label := a.composerModeLabel(kind, v)
				current := a.access
				if kind == "workflow" {
					current = a.workflow
				}
				if current == "full" {
					current = "full-access"
				}
				row := c.Key(v)
				b := ui.ButtonBase(row).KeepFocus().Label(label).AlignSelf(ui.Stretch).Padding(8).Gap(8).Radius(6).AlignItems(ui.Center)
				if b.Hovered() {
					hovered = i
					if s.hovered != i {
						s.cursor = i
					}
				}
				if i == s.cursor || current == v {
					b.Background(a.Theme.Selection)
				}
				hintKey := kind + "." + strings.ReplaceAll(v, "-", "_") + "_hint"
				icon := "icons/lock.svg"
				if v == "auto-accept-edits" {
					icon = "icons/compose.svg"
				}
				if kind == "workflow" {
					icon = map[string]string{"plan": "icons/tools/todo.svg", "build": "icons/rocket-01.svg", "ask": "icons/tools/ask.svg"}[v]
				}
				b.Children(func() {
					a.icon(row.Key("icon"), icon, 16, a.Theme.TextSecondary)
					ui.Column(row.Key("copy")).Grow(1).MinWidth(0).Gap(2).Children(func() {
						ui.Text(row.Key("title"), label).FontSize(a.Theme.Px(14))
						ui.Text(row.Key("hint"), a.Tr(hintKey)).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextSecondary)
					})
					if current == v {
						a.icon(row.Key("current"), "icons/check.svg", 14, a.Theme.Text)
					}
				})
				if current == v {
					b.Description("Current mode")
				}
				if b.Clicked() {
					a.chooseComposerMode(s, v)
				}
			}
		})
		s.hovered = hovered
	})
	if popup.Dismissed() && a.composerMode == s {
		a.closeComposerModeMenu(true)
	}
}
