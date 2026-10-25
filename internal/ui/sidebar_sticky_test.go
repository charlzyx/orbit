package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
	"time"
)

func stickySidebarFixture(t *testing.T) (*App, *ui.Tester) {
	a := fixture(t)
	a.Model.Workspace = "/work/alpha"
	a.Model.SessionID = "alpha-0"
	a.Model.SessionPath = "/store/alpha-0.jsonl"
	a.Model.Sessions = nil
	a.Workspaces.Workspaces = []preferences.Workspace{{Path: "/work/alpha"}, {Path: "/work/beta"}, {Path: "/work/gamma"}}
	a.Workspaces.Sort = "manual"
	a.Workspaces.GroupBy = "workspace"
	a.sidebarState.expandedWorkspaces = map[string]bool{"alpha": true, "beta": true, "gamma": true}
	a.sidebarState.extraSessions = map[string]int{"alpha": 40, "beta": 40, "gamma": 40}
	for _, group := range []string{"alpha", "beta", "gamma"} {
		for i := 0; i < 30; i++ {
			id := fmt.Sprintf("%s-%d", group, i)
			a.Model.Sessions = append(a.Model.Sessions, SessionRow{ID: id, Path: "/store/" + id + ".jsonl", Workspace: "/work/" + group, Name: id, Summary: "preview", Updated: time.Unix(int64(100-i), 0)})
		}
	}
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	return a, tt
}
func TestSidebarStickyExpandedWorkspace(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	tt.Scroll(100, 400, 0, 500)
	first, _ := a.sessionList.Visible()
	if first < 2 {
		t.Fatal("fixture did not scroll", first)
	}
	if _, ok := tt.Find("Sidebar sticky workspace /work/alpha"); !ok {
		t.Fatal("expanded workspace header disappeared")
	}
}

func TestSidebarStickyNonactiveGroupKeepsNativeAnchorAndWheel(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	a.sessionList.ScrollTo(42, ui.Start)
	tt.Frame()
	first, _ := a.sessionList.Visible()
	if first != 42 {
		t.Fatal("sticky changed native start anchor", first)
	}
	header, ok := tt.Find("Sidebar sticky workspace /work/beta")
	viewport, _ := tt.Find("Sidebar session viewport")
	if !ok || math.Abs(float64(header.Y-viewport.Y)) > 1 {
		t.Fatal("nonactive expanded group did not pin", header, viewport)
	}
	tt.Scroll(header.X+header.W/2, header.Y+header.H/2, 0, 130)
	if got, _ := a.sessionList.Visible(); got <= first {
		t.Fatal("sticky swallowed wheel", first, got)
	}
	if err := tt.Click("Workspace actions /work/beta"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy path"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "/work/beta" {
		t.Fatal("pinned menu has wrong owner", tt.Clipboard())
	}
	if err := tt.Click("Workspace actions /work/beta"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Icon & color"); err != nil {
		t.Fatal(err)
	}
	if a.workspaceActions.target != a.workspaceEntry("/work/beta") {
		t.Fatal("pinned picker target")
	}
	if _, ok := tt.Find("Workspace appearance"); !ok {
		t.Fatal("pinned popup not rendered")
	}
	tt.Key(0, ui.KeyEscape)
}
func TestSidebarStickyPushOffAndCollapsedFlatExclusion(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	rows := a.sidebarRows()
	next := -1
	for i, r := range rows {
		if r.group && r.path == "/work/beta" {
			next = i
			break
		}
	}
	a.sessionList.ScrollTo(next-2, ui.Start)
	tt.Frame()
	viewport, _ := tt.Find("Sidebar session viewport")
	pushed := false
	for i := 0; i < 35; i++ {
		tt.Scroll(viewport.X+100, viewport.Y+100, 0, 4)
		incoming, yes := tt.Find("Sidebar workspace header /work/beta")
		outgoing, old := tt.Find("Sidebar sticky workspace /work/alpha")
		if yes && old && incoming.Y > viewport.Y && incoming.Y < viewport.Y+32 {
			if math.Abs(float64(outgoing.Y+outgoing.H-incoming.Y)) > 1 {
				t.Fatal("next group overlapped pinned group", incoming, outgoing)
			}
			pushed = true
			break
		}
	}
	if !pushed {
		t.Fatal("never saw next group push current header")
	}
	a.sidebarState.expandedWorkspaces["beta"] = false
	// Always-visible pins survive a collapsed group; they must not pin its header.
	for i := range a.Model.Sessions {
		if a.Model.Sessions[i].Workspace == "/work/beta" {
			a.Model.Sessions[i].Pinned = true
		}
	}
	rows = a.sidebarRows()
	for i, r := range rows {
		if r.group && r.path == "/work/beta" {
			a.sessionList.ScrollTo(i+4, ui.Start)
			break
		}
	}
	tt.Frame()
	if _, ok := tt.Find("Sidebar sticky workspace /work/beta"); ok {
		t.Fatal("collapsed group pinned")
	}
	a.Workspaces.GroupBy = "one_list"
	tt.Frame()
	for _, path := range []string{"alpha", "beta", "gamma"} {
		if _, ok := tt.Find("Sidebar sticky workspace /work/" + path); ok {
			t.Fatal("flat list pinned", path)
		}
	}
}
func TestSidebarStickyMenuAndHeldPressCannotRetarget(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	a.sessionList.ScrollTo(42, ui.Start)
	tt.Frame()
	if err := tt.Click("Workspace actions /work/beta"); err != nil {
		t.Fatal(err)
	}
	a.sessionList.ScrollTo(75, ui.Start)
	tt.Frame()
	_ = tt.ChooseMenuItem("Copy path")
	if tt.Clipboard() != "" {
		t.Fatal("old pinned menu retargeted", tt.Clipboard())
	}
	tt.CloseMenu()
	header, ok := tt.Find("Sidebar sticky workspace /work/gamma")
	if !ok {
		t.Fatal("gamma fixture missing")
	}
	x, y := header.X+75, header.Y+header.H/2
	tt.Press(x, y)
	a.sessionList.ScrollTo(42, ui.Start)
	tt.Frame()
	tt.Release(x, y)
	tt.Frame()
	if a.sidebarGroupCollapsed("beta") {
		t.Fatal("press on gamma collapsed beta")
	}
}
func TestSidebarStickyResizeFontAndDatasetInvalidation(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	a.sessionList.ScrollTo(42, ui.Start)
	tt.Frame()
	before, ok := tt.Find("Sidebar sticky workspace /work/beta")
	if !ok {
		t.Fatal("missing fixture")
	}
	a.Prefs.UIFontSize = 22
	a.SidebarWidth = 300
	tt.SetSize(1000, 760)
	tt.Frame()
	a.sessionList.ScrollTo(42, ui.Start)
	tt.Frame()
	after, ok := tt.Find("Sidebar sticky workspace /work/beta")
	viewport, _ := tt.Find("Sidebar session viewport")
	if !ok || after.H <= before.H || math.Abs(float64(after.W-viewport.W)) > 1 {
		t.Fatal("stale sticky metrics", before, after, viewport)
	}
	a.Workspaces.Workspaces = []preferences.Workspace{{Path: "/work/gamma"}}
	tt.Frame()
	if _, ok := tt.Find("Sidebar sticky workspace /work/beta"); ok {
		t.Fatal("removed workspace retained pinned controls")
	}
	if _, ok := a.sidebarState.sticky.headers["/work/beta"]; ok {
		t.Fatal("obsolete header handle retained")
	}
}
func TestSidebarStickyNativeScrollbarAndFocus(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	viewport, _ := tt.Find("Sidebar session viewport")
	first, _ := a.sessionList.Visible()
	x, y := viewport.X+viewport.W-sidebarResizeGripWidth-5, viewport.Y+40
	tt.Move(x, y)
	tt.Frame() // native overlay scrollbar must be painted before hit-testing
	tt.Press(x, y)
	tt.Move(x, y+120)
	tt.Release(x, y+120)
	tt.Frame()
	if got, _ := a.sessionList.Visible(); got <= first {
		t.Fatal("native vertical scrollbar stopped dragging", first, got)
	}
	a.focusSidebarSessions()
	tt.Frame()
	if a.sessionCursor < 0 || !a.sidebarRows()[a.sessionCursor].isSession() {
		t.Fatal("focus lost active session")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEscape)
	if a.sessionCursor != -1 {
		t.Fatal("sidebar Escape focus behavior changed")
	}
}

func TestSidebarStickyResizeGripStillOwnsBoundary(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	grip, ok := tt.Find("Resize sidebar")
	if !ok {
		t.Fatal("resize grip missing")
	}
	before := a.SidebarWidth
	x, y := grip.X+grip.W/2, grip.Y+grip.H/2
	tt.Press(x, y)
	tt.Move(x+40, y)
	tt.Release(x+40, y)
	tt.Frame()
	if a.SidebarWidth <= before+20 {
		t.Fatal("scrollbar fix stole resize grip", before, a.SidebarWidth)
	}
}
func TestSidebarStickyPartialHeaderHasSinglePopupOwner(t *testing.T) {
	_, tt := stickySidebarFixture(t)
	tt.Scroll(100, 400, 0, 8)
	if _, ok := tt.Find("Sidebar sticky workspace /work/alpha"); !ok {
		t.Fatal("partial header fixture missing")
	}
	if err := tt.Click("Workspace actions /work/alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Icon & color"); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, label := range tt.Texts() {
		if label == "Workspace appearance" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("pinned and normal rows rendered duplicate popups", count)
	}
}

func TestSidebarStickyMenuRightEdgeDoesNotHitScrollbar(t *testing.T) {
	a, tt := stickySidebarFixture(t)
	a.sessionList.ScrollTo(42, ui.Start)
	tt.Frame()
	b, ok := tt.Find("Workspace actions /work/beta")
	if !ok {
		t.Fatal("pinned actions missing")
	}
	// Hover paints the native bar before testing the rightmost action pixel.
	tt.Move(b.X+b.W-1, b.Y+b.H/2)
	tt.Frame()
	tt.ClickAt(b.X+b.W-1, b.Y+b.H/2)
	if err := tt.ChooseMenuItem("Copy path"); err != nil {
		t.Fatal("action edge overlapped native scrollbar", err)
	}
	if tt.Clipboard() != "/work/beta" {
		t.Fatal(tt.Clipboard())
	}
}
