package orbitui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
)

// Source: app.rs::SIDEBAR_GROUP_SESSIONS_VISIBLE and app/sidebar.rs. Group
// disclosure is transient; only sort, grouping and archive filters persist.
const sidebarGroupSessionsVisible = 3

type sidebarState struct {
	collapsedWorkspaces map[string]bool
	expandedWorkspaces  map[string]bool
	extraSessions       map[string]int
}

type sidebarRow struct {
	key, label, path string
	group            bool
	collapsed        bool
	count            int
	more             int
	less             bool
	canCollapse      bool
	session          SessionRow
}

func (r sidebarRow) isSession() bool { return !r.group && r.more == 0 && !r.less }

func sidebarWorkspaceLabel(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Base(filepath.Clean(path))
}

func (a *App) sidebarSessionActive(s SessionRow) bool {
	if a.Model.SessionPath != "" {
		return s.Path == a.Model.SessionPath
	}
	return a.Model.SessionID != "" && s.ID == a.Model.SessionID
}

func (a *App) sidebarSessionRank(s SessionRow) int {
	switch {
	case a.sidebarSessionActive(s):
		return 0
	case s.Pinned:
		return 1
	case s.Running:
		return 2
	default:
		return 3
	}
}

// The same filtered snapshot feeds rendering and navigation. Keeping complete
// SessionRows here avoids the reference's filtered/unfiltered index mismatch.
func (a *App) sidebarSessions() []SessionRow {
	listed := make(map[string]bool, len(a.Workspaces.Workspaces))
	for _, w := range a.Workspaces.Workspaces {
		listed[filepath.Clean(w.Path)] = true
	}
	sessions := make([]SessionRow, 0, len(a.Model.Sessions)+1)
	foundActive := false
	for _, s := range a.Model.Sessions {
		if !listed[filepath.Clean(s.Workspace)] {
			continue
		}
		if a.sidebarSessionActive(s) {
			foundActive = true
			if name := strings.TrimSpace(a.Model.SessionName); name != "" {
				s.Name = name
			}
		}
		sessions = append(sessions, s)
	}
	// An active, started session can precede its first session-store scan. A
	// blank draft must not create a sidebar task. Grouped mode still requires
	// its workspace header, as in sessions_with_placeholder/build_sidebar_rows.
	if !foundActive && a.Model.SessionPath != "" && len(a.Model.Messages) > 0 {
		preview := ""
		for _, m := range a.Model.Messages {
			if m.Role == "user" {
				preview = m.Text
				break
			}
		}
		name := strings.TrimSpace(a.Model.SessionName)
		if name == "" {
			name = sidebarCapChars(preview, 80)
		}
		sessions = append(sessions, SessionRow{ID: a.Model.SessionID, Path: a.Model.SessionPath, Name: name, Workspace: a.Model.Workspace, Summary: sidebarCapChars(preview, 110), Updated: time.Now(), Running: a.Model.Running})
	}
	visible := sessions[:0]
	for _, s := range sessions {
		if (a.Workspaces.ArchivedFilter == "hide" && s.Archived) || (a.Workspaces.ArchivedFilter == "only" && !s.Archived) {
			continue
		}
		if s.Name == "" {
			s.Name = a.text("menu.new_task", "New task")
		}
		visible = append(visible, s)
	}
	// Store scans are newest-first. Sort defensively for live snapshots while
	// retaining store order for equal activity timestamps.
	sort.SliceStable(visible, func(i, j int) bool { return visible[i].Updated.After(visible[j].Updated) })
	return visible
}

func sidebarCapChars(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func (a *App) sidebarGroupCollapsed(label string) bool {
	if label == sidebarWorkspaceLabel(a.Model.Workspace) {
		return a.sidebarState.collapsedWorkspaces[label]
	}
	return !a.sidebarState.expandedWorkspaces[label]
}

func (a *App) toggleSidebarGroup(label string) {
	if label == sidebarWorkspaceLabel(a.Model.Workspace) {
		if a.sidebarState.collapsedWorkspaces == nil {
			a.sidebarState.collapsedWorkspaces = map[string]bool{}
		}
		a.sidebarState.collapsedWorkspaces[label] = !a.sidebarState.collapsedWorkspaces[label]
	} else {
		if a.sidebarState.expandedWorkspaces == nil {
			a.sidebarState.expandedWorkspaces = map[string]bool{}
		}
		a.sidebarState.expandedWorkspaces[label] = !a.sidebarState.expandedWorkspaces[label]
	}
	a.clampSidebarCursor()
}

func (a *App) sidebarRows() []sidebarRow {
	sessions := a.sidebarSessions()
	sessionRow := func(s SessionRow) sidebarRow {
		return sidebarRow{key: "session:" + s.Path, label: s.Name, session: s}
	}
	leadFirst := func(rows []SessionRow) {
		sort.SliceStable(rows, func(i, j int) bool { return a.sidebarSessionRank(rows[i]) < a.sidebarSessionRank(rows[j]) })
	}
	if a.Workspaces.GroupBy == "one_list" {
		leadFirst(sessions)
		rows := make([]sidebarRow, 0, len(sessions))
		for _, s := range sessions {
			rows = append(rows, sessionRow(s))
		}
		return rows
	}
	type group struct {
		label, path string
		added       int64
		sessions    []SessionRow
		updated     time.Time
		active      bool
	}
	var groups []group
	byLabel := map[string]int{}
	for _, w := range a.Workspaces.Workspaces {
		label := sidebarWorkspaceLabel(w.Path)
		if _, ok := byLabel[label]; ok {
			continue
		}
		byLabel[label] = len(groups)
		groups = append(groups, group{label: label, path: w.Path, added: w.AddedAt})
	}
	for _, s := range sessions {
		if ix, ok := byLabel[sidebarWorkspaceLabel(s.Workspace)]; ok {
			g := &groups[ix]
			g.sessions = append(g.sessions, s)
			if s.Updated.After(g.updated) {
				g.updated = s.Updated
			}
			g.active = g.active || a.sidebarSessionActive(s)
		}
	}
	sort.SliceStable(groups, func(i, j int) bool {
		l, r := groups[i], groups[j]
		byName := strings.ToLower(l.label) < strings.ToLower(r.label)
		switch a.Workspaces.Sort {
		case "alphabetical":
			return byName
		case "alphabetical_desc":
			return strings.ToLower(l.label) > strings.ToLower(r.label)
		case "date_added":
			if l.added != r.added {
				return l.added > r.added
			}
		case "session_count":
			if len(l.sessions) != len(r.sessions) {
				return len(l.sessions) > len(r.sessions)
			}
		case "last_updated", "":
			if l.active != r.active {
				return l.active
			}
			if !l.updated.Equal(r.updated) {
				return l.updated.After(r.updated)
			}
		default: // manual preserves the explicitly saved project order.
			return false
		}
		return byName
	})
	var rows []sidebarRow
	for _, g := range groups {
		leadFirst(g.sessions)
		collapsed := a.sidebarGroupCollapsed(g.label)
		rows = append(rows, sidebarRow{key: "workspace:" + g.label, label: g.label, path: g.path, group: true, collapsed: collapsed, count: len(g.sessions)})
		if collapsed {
			for _, s := range g.sessions {
				if a.sidebarSessionRank(s) < 3 {
					rows = append(rows, sessionRow(s))
				}
			}
			continue
		}
		extra := a.sidebarState.extraSessions[g.label]
		limit := min(len(g.sessions), sidebarGroupSessionsVisible+extra)
		for _, s := range g.sessions[:limit] {
			rows = append(rows, sessionRow(s))
		}
		if hidden := len(g.sessions) - limit; hidden > 0 {
			count := min(hidden, sidebarGroupSessionsVisible)
			label := orbitassets.Translate(a.catalog, a.english, "sidebar.show_count_more", map[string]string{"count": strconv.Itoa(count)})
			if label == "sidebar.show_count_more" {
				label = fmt.Sprintf("Show %d more", count)
			}
			rows = append(rows, sidebarRow{key: "more:" + g.label, path: g.label, label: label, more: count, canCollapse: extra > 0})
		} else if extra > 0 {
			rows = append(rows, sidebarRow{key: "less:" + g.label, path: g.label, label: a.text("sidebar.show_less", "Show less"), less: true})
		}
	}
	return rows
}

func (a *App) activateSidebarRow(r sidebarRow) {
	switch {
	case r.group:
		a.toggleSidebarGroup(r.label)
	case r.more > 0:
		if a.sidebarState.extraSessions == nil {
			a.sidebarState.extraSessions = map[string]int{}
		}
		a.sidebarState.extraSessions[r.path] += r.more
		a.clampSidebarCursor()
	case r.less:
		a.collapseSidebarSessions(r.path)
	default:
		a.command("SwitchSession", r.session.Path)
		a.sessionCursor = -1
		a.composer.Focus()
	}
}

func (a *App) collapseSidebarSessions(label string) {
	delete(a.sidebarState.extraSessions, label)
	a.clampSidebarCursor()
}

func (a *App) clampSidebarCursor() {
	if n := len(a.sidebarRows()); n == 0 {
		a.sessionCursor = -1
	} else if a.sessionCursor >= n {
		a.sessionCursor = n - 1
	}
}

func (a *App) navigableSidebarSessions() []SessionRow {
	var sessions []SessionRow
	for _, r := range a.sidebarRows() {
		if r.isSession() {
			sessions = append(sessions, r.session)
		}
	}
	return sessions
}

func (a *App) openSidebarSessionSlot(value string) {
	n, err := strconv.Atoi(value)
	sessions := a.navigableSidebarSessions()
	if err == nil && n > 0 && n <= len(sessions) {
		a.command("SwitchSession", sessions[n-1].Path)
	}
}

func (a *App) cycleSidebarSession(direction int) {
	sessions := a.navigableSidebarSessions()
	if len(sessions) < 2 {
		return
	}
	next := 0
	if direction < 0 {
		next = len(sessions) - 1
	}
	for i, s := range sessions {
		if a.sidebarSessionActive(s) {
			next = (i + direction + len(sessions)) % len(sessions)
			break
		}
	}
	a.command("SwitchSession", sessions[next].Path)
}

func (a *App) focusSidebarSessions() {
	if a.page == "settings" {
		return
	}
	a.SidebarVisible = true
	rows := a.sidebarRows()
	a.sessionCursor = -1
	if len(rows) > 0 {
		a.sessionCursor = 0
	}
	for i, r := range rows {
		if r.isSession() && a.sidebarSessionActive(r.session) {
			a.sessionCursor = i
			break
		}
	}
	if a.sessionCursor >= 0 {
		a.sessionList.ScrollIntoView(a.sessionCursor)
	}
	a.sessionList.Handle.Focus()
}

func (a *App) sidebarShortcut(c *ui.Context) bool {
	if a.SidebarVisible && a.sessionList.Handle.FocusWithin(c) && c.Shortcut(0, ui.KeyEscape) {
		a.sessionCursor = -1
		a.composer.Focus()
		return true
	}
	return false
}

func (a *App) sidebarOverflowRow(c *ui.Context, r sidebarRow) {
	t := a.Theme
	ui.Row(c).AlignSelf(ui.Stretch).MarginX(8).Height(t.Px(26)).Padding(0, t.Space(4), 0, 14).AlignItems(ui.Center).Children(func() {
		icon := "icons/chevron-down.svg"
		if r.less {
			icon = "icons/chevron-up.svg"
		}
		b := a.button(c, r.label, icon, false).Grow(1).MinWidth(0).Height(t.Px(26))
		if b.Clicked() {
			a.activateSidebarRow(r)
		}
		if r.canCollapse && a.iconButton(c, a.text("sidebar.show_less", "Show less"), "icons/chevron-up.svg").Size(t.Px(24), t.Px(24)).Clicked() {
			a.collapseSidebarSessions(r.path)
		}
	})
}
