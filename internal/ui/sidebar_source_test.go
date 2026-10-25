package orbitui

import (
	"reflect"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
)

// Fixture order deliberately disagrees with displayed order: an old active
// task, a pin, and a background run must lead newer ordinary/archived rows.
func sourceSidebarFixture(t *testing.T) *App {
	t.Helper()
	a := fixture(t)
	a.Model.Workspace = "/work/alpha"
	a.Model.SessionID = "active"
	a.Model.SessionPath = "/store/active.jsonl"
	a.Workspaces.Workspaces = []preferences.Workspace{{Path: "/work/beta", AddedAt: 200}, {Path: "/work/alpha", AddedAt: 100}, {Path: "/work/empty", AddedAt: 300}}
	a.Workspaces.Sort = "last_updated"
	a.Workspaces.ArchivedFilter = "hide"
	makeRow := func(id, workspace string, stamp int64) SessionRow {
		return SessionRow{ID: id, Path: "/store/" + id + ".jsonl", Name: id + " task", Workspace: workspace, Summary: "Fixture preview for " + id, Updated: time.Unix(stamp, 0)}
	}
	a.Model.Sessions = []SessionRow{
		makeRow("archived", "/work/alpha", 900),
		makeRow("beta-new", "/work/beta", 800),
		makeRow("new", "/work/alpha", 700),
		makeRow("next", "/work/alpha", 600),
		makeRow("third", "/work/alpha", 500),
		makeRow("fourth", "/work/alpha", 400),
		makeRow("running", "/work/alpha", 200),
		makeRow("pin", "/work/alpha", 100),
		makeRow("active", "/work/alpha", 1),
		makeRow("unlisted", "/work/outside", 1000),
	}
	a.Model.Sessions[0].Archived = true
	a.Model.Sessions[6].Running = true
	a.Model.Sessions[7].Pinned = true
	return a
}

func sidebarSessionIDs(a *App) []string {
	ids := []string{}
	for _, r := range a.sidebarRows() {
		if r.isSession() {
			ids = append(ids, r.session.ID)
		}
	}
	return ids
}
func sidebarGroupLabels(a *App) []string {
	labels := []string{}
	for _, r := range a.sidebarRows() {
		if r.group {
			labels = append(labels, r.label)
		}
	}
	return labels
}
func requireSidebarSessions(t *testing.T, a *App, want ...string) {
	t.Helper()
	if got := sidebarSessionIDs(a); !reflect.DeepEqual(got, want) {
		t.Fatalf("visible sidebar sessions = %v; want %v", got, want)
	}
}
func sidebarOverflow(t *testing.T, a *App) sidebarRow {
	t.Helper()
	for _, r := range a.sidebarRows() {
		if r.more > 0 || r.less {
			return r
		}
	}
	t.Fatal("expected sidebar overflow row")
	return sidebarRow{}
}

func TestSourceSidebarGroupRanksCapsAndDisclosure(t *testing.T) {
	a := sourceSidebarFixture(t)
	if got := sidebarGroupLabels(a); !reflect.DeepEqual(got, []string{"alpha", "beta", "empty"}) {
		t.Fatal("active workspace must lead newest background group", got)
	}
	requireSidebarSessions(t, a, "active", "pin", "running")
	rows := a.sidebarRows()
	if rows[0].collapsed || rows[0].count != 7 || !rows[5].collapsed {
		t.Fatal("source active-open/background-collapsed defaults or filtered counts", rows)
	}
	more := sidebarOverflow(t, a)
	if more.more != 3 || more.label != "Show 3 more" || more.canCollapse {
		t.Fatal("initial three-session overflow", more)
	}
	a.activateSidebarRow(more)
	requireSidebarSessions(t, a, "active", "pin", "running", "new", "next", "third")
	more = sidebarOverflow(t, a)
	if more.more != 1 || !more.canCollapse {
		t.Fatal("tail overflow must reveal only remainder and offer collapse", more)
	}
	a.activateSidebarRow(more)
	requireSidebarSessions(t, a, "active", "pin", "running", "new", "next", "third", "fourth")
	less := sidebarOverflow(t, a)
	if !less.less || less.label != "Show less" {
		t.Fatal("fully expanded history needs Show less", less)
	}
	a.activateSidebarRow(less)
	requireSidebarSessions(t, a, "active", "pin", "running")
	a.toggleSidebarGroup("alpha")
	requireSidebarSessions(t, a, "active", "pin", "running")
	for _, r := range a.sidebarRows() {
		if r.more != 0 || r.less {
			t.Fatal("collapsed group must not show overflow", r)
		}
	}
	a.toggleSidebarGroup("beta")
	requireSidebarSessions(t, a, "active", "pin", "running", "beta-new")
}

func TestSourceSidebarFlatFilterAndStableLeadRanks(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Workspaces.GroupBy = "one_list"
	requireSidebarSessions(t, a, "active", "pin", "running", "beta-new", "new", "next", "third", "fourth")
	for _, r := range a.sidebarRows() {
		if !r.isSession() {
			t.Fatal("flat mode invented a header/overflow row", r)
		}
	}
	a.Workspaces.ArchivedFilter = "only"
	requireSidebarSessions(t, a, "archived")
	a.Workspaces.ArchivedFilter = "show"
	requireSidebarSessions(t, a, "active", "pin", "running", "archived", "beta-new", "new", "next", "third", "fourth")
	// A watcher reload moving a busy run to newest cannot move the open row.
	a.Model.Sessions[6].Updated = time.Unix(10000, 0)
	requireSidebarSessions(t, a, "active", "pin", "running", "archived", "beta-new", "new", "next", "third", "fourth")
}

func TestSourceSidebarWorkspaceSortAndDuplicateBasenames(t *testing.T) {
	a := sourceSidebarFixture(t)
	for _, tc := range []struct {
		mode string
		want []string
	}{
		{"manual", []string{"beta", "alpha", "empty"}},
		{"alphabetical", []string{"alpha", "beta", "empty"}},
		{"alphabetical_desc", []string{"empty", "beta", "alpha"}},
		{"last_updated", []string{"alpha", "beta", "empty"}},
		{"date_added", []string{"empty", "beta", "alpha"}},
		{"session_count", []string{"alpha", "beta", "empty"}},
	} {
		a.Workspaces.Sort = tc.mode
		if got := sidebarGroupLabels(a); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s groups = %v; want %v", tc.mode, got, tc.want)
		}
	}
	a.Workspaces.Sort = "manual"
	a.Workspaces.Workspaces = append(a.Workspaces.Workspaces, preferences.Workspace{Path: "/other/alpha", AddedAt: 999})
	a.Model.Sessions = append(a.Model.Sessions, SessionRow{ID: "same-label", Path: "/store/same-label.jsonl", Name: "Same label task", Workspace: "/other/alpha", Pinned: true})
	if got := sidebarGroupLabels(a); !reflect.DeepEqual(got, []string{"beta", "alpha", "empty"}) {
		t.Fatal("source duplicate-basename group identity", got)
	}
	for _, r := range a.sidebarRows() {
		if r.group && r.label == "alpha" && (r.path != "/work/alpha" || r.count != 8) {
			t.Fatal("first saved path must own merged label and count", r)
		}
	}
}

func TestSourceSidebarMountedShortcutsUseDisplayedFilteredOrder(t *testing.T) {
	a := sourceSidebarFixture(t)
	var dispatched []Action
	a.Options.OnAction = func(v Action) { dispatched = append(dispatched, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	var lastY float32
	for _, id := range []string{"active", "pin", "running"} {
		r, ok := tt.Find(id + " task")
		if !ok || r.Y <= lastY {
			t.Fatalf("mounted rank order: %s at %+v found=%v", id, r, ok)
		}
		lastY = r.Y
	}
	for _, id := range []string{"archived", "new", "beta-new", "unlisted"} {
		if _, ok := tt.Find(id + " task"); ok {
			t.Fatal("hidden row mounted", id)
		}
	}
	checkKey := func(mods ui.Modifiers, key ui.Key, want string) {
		t.Helper()
		dispatched = nil
		tt.Key(mods, key)
		if len(dispatched) != 1 || dispatched[0].Kind != "SwitchSession" || dispatched[0].Value != "/store/"+want+".jsonl" {
			t.Fatalf("key %v/%v actions = %+v; want %s", mods, key, dispatched, want)
		}
	}
	checkKey(ui.Cmd, ui.Key1, "active")
	checkKey(ui.Cmd, ui.Key2, "pin")
	checkKey(ui.Ctrl, ui.KeyTab, "pin")
	checkKey(ui.Ctrl|ui.Shift, ui.KeyTab, "running")
	dispatched = nil
	tt.Key(ui.Cmd, ui.Key9)
	if len(dispatched) != 0 {
		t.Fatal("past-end slot dispatched", dispatched)
	}
	if err := tt.Click("alpha"); err != nil {
		t.Fatal(err)
	}
	requireSidebarSessions(t, a, "active", "pin", "running")
	checkKey(ui.Cmd, ui.Key3, "running")
	// Archive first in the backing snapshot must not corrupt navigation's
	// session index after filtering, grouping, or changing to flat mode.
	a.Workspaces.GroupBy = "one_list"
	a.Workspaces.ArchivedFilter = "only"
	tt.Frame()
	checkKey(ui.Cmd, ui.Key1, "archived")
	dispatched = nil
	tt.Key(ui.Ctrl, ui.KeyTab)
	if len(dispatched) != 0 {
		t.Fatal("single visible session must not cycle", dispatched)
	}
	a.Workspaces.ArchivedFilter = "hide"
	tt.Frame()
	checkKey(ui.Cmd, ui.Key4, "beta-new")
}

func TestSourceSidebarMountedKeyboardFocusEnterEscape(t *testing.T) {
	a := sourceSidebarFixture(t)
	var dispatched []Action
	a.Options.OnAction = func(v Action) { dispatched = append(dispatched, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	if a.sessionCursor != 1 {
		t.Fatal("focus must select active row, not header", a.sessionCursor)
	}
	tt.Key(0, ui.KeyDown)
	if a.sessionCursor != 2 {
		t.Fatal("sidebar Down must select pin", a.sessionCursor)
	}
	tt.Key(0, ui.KeyEnter)
	if len(dispatched) != 1 || dispatched[0].Value != "/store/pin.jsonl" || !tt.Focused("Message") {
		t.Fatal("Enter did not dispatch selected visible task and return composer focus", dispatched, a.sessionCursor)
	}
	dispatched = nil
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if !a.sidebarGroupCollapsed("alpha") || len(dispatched) != 0 {
		t.Fatal("Enter on header should toggle only group", dispatched)
	}
	tt.Key(0, ui.KeyEscape)
	if a.sessionCursor != -1 || !tt.Focused("Message") || len(dispatched) != 0 {
		t.Fatal("focused sidebar Escape must not abort", dispatched, a.sessionCursor)
	}
	tt.Key(0, ui.KeyEscape)
	if len(dispatched) != 1 || dispatched[0].Kind != "AbortRun" {
		t.Fatal("genuine chat Escape should still abort", dispatched)
	}
}

func TestSourceSidebarMountedOverflowClickAndKeyboard(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Workspaces.Workspaces = a.Workspaces.Workspaces[1:2] // alpha only
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Show 3 more"); err != nil {
		t.Fatal(err)
	}
	requireSidebarSessions(t, a, "active", "pin", "running", "new", "next", "third")
	if _, ok := tt.Find("Show less"); !ok {
		t.Fatal("partial overflow lacks collapse control")
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	requireSidebarSessions(t, a, "active", "pin", "running", "new", "next", "third", "fourth")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	requireSidebarSessions(t, a, "active", "pin", "running")
	if a.sessionCursor < 0 || a.sessionCursor >= len(a.sidebarRows()) {
		t.Fatal("overflow collapse left stale cursor", a.sessionCursor)
	}
}

func TestSourceSidebarPreferencesPersistButDisclosureResets(t *testing.T) {
	a := sourceSidebarFixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	choose := func(label string) {
		t.Helper()
		if err := tt.Click("Project options"); err != nil {
			t.Fatal(err)
		}
		if err := tt.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
	}
	choose("Sort: date added")
	choose("Archived: only")
	choose("Group by workspace")
	stored := a.Options.Store.Workspaces()
	if stored.Sort != "date_added" || stored.ArchivedFilter != "only" || stored.GroupBy != "one_list" {
		t.Fatal("mounted menu choices not persisted", stored)
	}
	b := New(Options{Store: preferences.New(a.Options.Store.Home)})
	b.Model = a.Model
	requireSidebarSessions(t, b, "archived")
	if !reflect.DeepEqual(b.Workspaces, stored) {
		t.Fatal("restart lost sidebar choices")
	}
	a.Workspaces.GroupBy = "workspace"
	a.Workspaces.ArchivedFilter = "hide"
	a.toggleSidebarGroup("alpha")
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	b = New(Options{Store: preferences.New(a.Options.Store.Home)})
	b.Model = a.Model
	if b.sidebarGroupCollapsed("alpha") || !b.sidebarGroupCollapsed("beta") {
		t.Fatal("disclosure should reset to source defaults on restart")
	}
}

func TestSourceSidebarFilterEmptyAndCycleWithoutActive(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Workspaces.GroupBy = "one_list"
	a.Workspaces.ArchivedFilter = "only"
	a.Model.Sessions[0].Archived = false
	var dispatched []Action
	a.Options.OnAction = func(v Action) { dispatched = append(dispatched, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if len(a.sidebarRows()) != 0 {
		t.Fatal("archive filter-empty must have no session rows")
	}
	for _, cmd := range []string{"OpenSessionSlot", "NextSession", "PrevSession"} {
		a.command(cmd, "1")
	}
	if len(dispatched) != 0 {
		t.Fatal("filter-empty navigation dispatched", dispatched)
	}
	a.Workspaces.ArchivedFilter = "hide"
	a.Model.SessionID, a.Model.SessionPath = "", ""
	for _, tc := range []struct{ cmd, want string }{{"NextSession", "pin"}, {"PrevSession", "active"}} {
		dispatched = nil
		a.command(tc.cmd, "")
		if len(dispatched) != 1 || dispatched[0].Value != "/store/"+tc.want+".jsonl" {
			t.Fatal("missing-current cycling must start at first/last displayed session", tc, dispatched)
		}
	}
}

func TestSourceSidebarActivePlaceholderAndLiveTitle(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Model.Sessions = nil
	if len(a.navigableSidebarSessions()) != 0 {
		t.Fatal("unsent draft became sidebar task")
	}
	a.Model.Messages = []Message{{Role: "user", Text: "  First task request  "}}
	s := a.navigableSidebarSessions()
	if len(s) != 1 || s[0].Name != "First task request" || s[0].Path != a.Model.SessionPath {
		t.Fatal("started placeholder", s)
	}
	a.Model.SessionName = "Live title"
	a.Model.Sessions = []SessionRow{{ID: "active", Path: a.Model.SessionPath, Name: "Stale disk title", Workspace: a.Model.Workspace}}
	s = a.navigableSidebarSessions()
	if len(s) != 1 || s[0].Name != "Live title" {
		t.Fatal("live title must replace disk title without duplicate placeholder", s)
	}
	if got := sidebarCapChars("  界界界  ", 2); got != "界界…" {
		t.Fatal("placeholder cap must trim and be Unicode-safe", got)
	}
}

func TestSourceSidebarCollapsedBackgroundKeepsPinsAndRuns(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Model.Sessions[6].Workspace = "/work/beta"
	a.Model.Sessions[7].Workspace = "/work/beta"
	requireSidebarSessions(t, a, "active", "new", "next", "pin", "running")
	if !a.sidebarGroupCollapsed("beta") {
		t.Fatal("background group should start collapsed")
	}
	a.Model.Sessions[6].Running = false
	requireSidebarSessions(t, a, "active", "new", "next", "pin")
	a.Model.Sessions[7].Pinned = false
	requireSidebarSessions(t, a, "active", "new", "next")
}
