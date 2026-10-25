package orbitui

import (
	"errors"
	"fmt"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"slices"
)

type workspaceMenuState struct {
	epoch      uint64
	open, busy bool
	rendered   bool
	target     *preferences.Workspace
}

func (a *App) workspaceEntry(path string) *preferences.Workspace {
	for i := range a.Workspaces.Workspaces {
		if a.Workspaces.Workspaces[i].Path == path {
			return &a.Workspaces.Workspaces[i]
		}
	}
	return nil
}
func (a *App) workspaceMark(w preferences.Workspace) (string, ui.Color) {
	icon := "folder"
	if w.Path == a.Model.Workspace {
		icon = "folder-open"
	} else {
		any, archived, running := false, true, false
		for _, s := range a.Model.Sessions {
			if s.Workspace == w.Path {
				any = true
				archived = archived && s.Archived
				running = running || s.Running
			}
		}
		if running {
			icon = "folder-sync"
		} else if any && archived {
			icon = "folder-archive"
		}
	}
	if slices.Contains(preferences.WorkspaceIcons, w.Icon) {
		icon = w.Icon
	}
	return "icons/" + icon + ".svg", a.workspaceTint(w.Tint)
}
func (a *App) workspaceTint(tint string) ui.Color {
	t := a.Theme
	switch tint {
	case "accent":
		return t.Accent
	case "green":
		return t.Success
	case "red":
		return t.Error
	case "yellow":
		return t.Warning
	case "blue":
		return t.SynFunction
	case "purple":
		return t.SynType
	case "pink":
		return t.SynString
	}
	return t.TextMuted
}
func (a *App) workspaceHeader(c *ui.Context, r sidebarRow, sticky, ownsActions bool) {
	w := a.workspaceEntry(r.path)
	if w == nil {
		return
	}
	s := &a.workspaceActions
	// The public button identity owns native-menu delivery, including A→B→A.
	ctx := c.Key(fmt.Sprintf("workspace-menu:%d:%s:%p:%d", s.epoch, r.path, w, w.AddedAt))
	icon, ink := a.workspaceMark(*w)
	// Keep the menu button left of the inset native scrollbar and resize grip.
	row := ui.Row(ctx).AlignSelf(ui.Stretch).Margin(0, sidebarScrollbarGutter(c), 0, 8).AlignItems(ui.Center)
	if !sticky {
		row.Bind(a.sidebarHeaderHandle(r.path)).Label("Sidebar workspace header " + r.path)
	}
	row.Children(func() {
		var b ui.Element
		if ownsActions {
			b = ui.ButtonBase(c.Key("header")).Label(r.label)
		} else {
			b = ui.Row(c.Key("header-presentation"))
		}
		b = b.Grow(1).MinWidth(0).Padding(6, 8).Gap(6).AlignItems(ui.Center)
		b.Children(func() {
			chevron := "icons/chevron-down.svg"
			if r.collapsed {
				chevron = "icons/chevron-right.svg"
			}
			a.icon(c, chevron, 10, a.Theme.TextMuted)
			a.icon(c, icon, 16, ink)
			ui.Text(c, r.label).MaxLines(1).Grow(1)
		})
		if ownsActions && b.Clicked() {
			a.activateSidebarRow(r)
		}
		menu := func(m *ui.Menu) {
			if m.Item(a.Tr("sidebar.copy_path")).Chosen() {
				c.WriteClipboard(r.path)
			}
			if m.Item("Worktrees").Chosen() {
				a.Navigate("worktrees")
			}
			if m.Item(a.Tr("sidebar.icon_and_color")).Chosen() {
				s.target = w
				s.open = true
			}
			if m.Item(a.Tr("sidebar.remove_from_sidebar")).Disabled(s.busy).Chosen() {
				a.editWorkspace(w, true, "", "")
			}
			m.Separator()
			if m.Item("New task here").Chosen() {
				a.Dispatch(Action{Kind: "SetWorkspace", Value: r.path})
				a.Navigate("chat")
			}
			if m.Item("Open in file manager").Chosen() {
				a.Dispatch(Action{Kind: "RevealPath", Value: r.path})
			}
		}
		if ownsActions {
			b.ContextMenu(menu)
			a.iconButton(c.Key("actions"), "Workspace actions "+r.path, "icons/more.svg").Size(20, 20).Menu(menu)
		} else {
			ui.Box(c.Key("actions-presentation")).Size(20, 20).Center().Children(func() { a.icon(c, "icons/more.svg", a.Theme.Px(16), a.Theme.TextSecondary) })
		}
		if ownsActions && s.open && s.target == w {
			a.workspaceAppearance(c, b, w)
		}
	})
}
func (a *App) editWorkspace(target *preferences.Workspace, remove bool, icon, tint string) {
	s := &a.workspaceActions
	if s.busy || target == nil || a.workspaceEntry(target.Path) != target {
		return
	}
	expected := *target
	revision := a.Workspaces.Revision
	s.busy = true
	if remove {
		s.open = false
	}
	a.Async(func() func() {
		err := a.Options.Store.EditWorkspace(expected, revision, remove, icon, tint)
		latest := a.Options.Store.Workspaces()
		return func() {
			s.busy = false
			if err != nil {
				if errors.Is(err, preferences.ErrWorkspaceListChanged) && a.workspaceEntry(expected.Path) == target {
					a.Workspaces = latest
					s.open = false
					s.target = nil
					s.epoch++
				}
				a.save(err)
				return
			}
			// A refreshed/replaced project list owns its own UI; never mutate the old
			// pointer or resurrect a workspace removed/re-added while the worker ran.
			if a.workspaceEntry(expected.Path) != target {
				return
			}
			a.Workspaces = latest
			s.epoch++
			s.target = nil
			if remove {
				if a.workspaceEntry(expected.Path) == nil {
					label := sidebarWorkspaceLabel(expected.Path)
					delete(a.sidebarState.collapsedWorkspaces, label)
					delete(a.sidebarState.expandedWorkspaces, label)
					delete(a.sidebarState.extraSessions, label)
				}
				s.open = false
			} else if next := a.workspaceEntry(expected.Path); next != nil && next.AddedAt == expected.AddedAt && next.Icon == icon && next.Tint == tint {
				s.target = next
			} else {
				s.open = false
			}

		}
	})
}
func (a *App) workspaceAppearance(c *ui.Context, anchor ui.Element, w *preferences.Workspace) {
	a.workspaceActions.rendered = true
	s := &a.workspaceActions
	t := a.Theme
	ui.PopoverBase(c.Key("workspace-appearance"), anchor, &s.open, func(panel ui.Element) {
		panel.Column().Width(t.Px(206)).Padding(t.Px(8), t.Px(16)).Gap(t.Px(8)).Background(t.Elevated).Border(1, t.Border).Radius(8).Label("Workspace appearance").Children(func() {
			ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.Tr("sidebar.icon_and_color")).Grow(1)
				if a.workspaceResetButton(c, s.busy).Clicked() {
					a.editWorkspace(w, false, "", "")
				}
			})
			ui.Row(c).AlignSelf(ui.Stretch).Justify(ui.SpaceBetween).Children(func() {
				for _, tint := range preferences.WorkspaceTints {
					b := ui.ButtonBase(c.Key(tint)).Label("Workspace tint "+tint).Size(t.Px(16), t.Px(16)).Radius(8).Background(a.workspaceTint(tint)).Disabled(s.busy)
					if w.Tint == tint || w.Tint == "" && tint == "default" {
						b.Border(2, t.Text)
					}
					if b.Clicked() {
						a.editWorkspace(w, false, w.Icon, tint)
					}
				}
			})
			for start := 0; start < len(preferences.WorkspaceIcons); start += 6 {
				ui.Row(c).Gap(t.Px(6)).Children(func() {
					for _, icon := range preferences.WorkspaceIcons[start:min(start+6, len(preferences.WorkspaceIcons))] {
						b := ui.ButtonBase(c.Key(icon)).Label("Workspace icon "+icon).Size(t.Px(24), t.Px(24)).Center().Disabled(s.busy)
						if w.Icon == icon || w.Icon == "" && icon == "folder" {
							b.Background(t.Selection)
						}
						b.Children(func() { a.icon(c, "icons/"+icon+".svg", 16, a.workspaceTint(w.Tint)) })
						if b.Clicked() {
							a.editWorkspace(w, false, icon, w.Tint)
						}
					}
				})
			}
		})
	})
}

// Rust sidebar.rs:963–1070: flat rows use their own status, not the current
// workspace's status. A deliberately chosen project icon still wins.
func (a *App) sessionWorkspaceMark(s SessionRow) (string, ui.Color) {
	icon := "folder"
	if a.sidebarSessionActive(s) {
		icon = "folder-open"
	} else if s.Running {
		icon = "folder-sync"
	} else if s.Archived {
		icon = "folder-archive"
	}
	tint := ""
	if w := a.workspaceEntry(s.Workspace); w != nil {
		tint = w.Tint
		if slices.Contains(preferences.WorkspaceIcons, w.Icon) {
			icon = w.Icon
		}
	}
	return "icons/" + icon + ".svg", a.workspaceTint(tint)
}
func (a *App) workspaceSessionSummary(c *ui.Context, s SessionRow) {
	t := a.Theme
	if a.Workspaces.GroupBy != "one_list" {
		ui.Text(c, s.Summary).MaxLines(1).FontSize(t.Px(12)).TextColor(t.TextMuted)
		return
	}
	ui.Row(c).Gap(t.Px(6)).AlignSelf(ui.Stretch).AlignItems(ui.Center).Children(func() {
		icon, ink := a.sessionWorkspaceMark(s)
		a.icon(c, icon, t.Px(12), ink)
		ui.Text(c, sidebarWorkspaceLabel(s.Workspace)).MaxWidth(t.Px(120)).MaxLines(1).FontSize(t.Px(12)).TextColor(t.TextSecondary)
		ui.Text(c, s.Summary).Grow(1).MinWidth(0).MaxLines(1).FontSize(t.Px(12)).TextColor(t.TextMuted)
	})
}

func (a *App) invalidateWorkspaceMenu(next string) {
	if next != a.Model.Workspace {
		a.workspaceActions.epoch++
		a.workspaceActions.open = false
	}
}

// Stable state identity is independent of the visible/accessibility language.
func (a *App) workspaceResetButton(c *ui.Context, busy bool) ui.Element {
	return a.button(c.Key("workspace-appearance-reset"), a.Tr("sidebar.reset"), "", false).Disabled(busy)
}
