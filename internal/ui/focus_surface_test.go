package orbitui

import (
	"encoding/json"
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWorkbenchTabMovesComposerToSidebarSurface(t *testing.T) {
	a := sourceSidebarFixture(t)
	var sidebarFocused bool
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); sidebarFocused = a.sessionList.Handle.FocusWithin(c) }, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !sidebarFocused {
		t.Fatal("Tab left composer for ordinary control rather than sidebar surface")
	}
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("reverse surface cycle did not return to composer")
	}
}

func TestWorkbenchSurfaceCycleSkipsHiddenAndWrapsReview(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.ReviewVisible = true
	a.reviewState.root = a.Model.Workspace
	a.reviewState.tree = true
	a.ReviewWidth = 600
	a.reviewState.data = workspace.ReviewSnapshot{Files: []workspace.ReviewFile{{Path: "fixture.go", DiffLine: 0}}, Lines: []workspace.ReviewLine{{File: 0, Kind: "file_header", Content: "fixture.go"}, {File: 0, Kind: "addition", NewLine: 1, Content: "hello"}}}
	var sidebar, review bool
	tt := ui.NewTester(func(c *ui.Context) {
		a.View(c)
		sidebar = a.sessionList.Handle.FocusWithin(c)
		review = a.reviewState.treeHandle.FocusWithin(c)
	}, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !sidebar {
		t.Fatal("missing sidebar stop")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !review {
		t.Fatal("missing review tree stop")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("cycle failed to wrap")
	}
	a.SidebarVisible = false
	tt.Frame()
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !review {
		t.Fatal("hidden sidebar remained in cycle")
	}
	a.ReviewVisible = false
	tt.Frame()
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("single-surface cycle escaped into ordinary controls")
	}
}
func TestWorkbenchEditorTabAndShiftTabKeepIndentOwnership(t *testing.T) {
	a, _, path := mountedFile(t, "hello")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if e := tt.Click("File editor"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	st := a.files.states[path]
	if st.Buffer.String() == "hello" || !tt.Focused("File editor") {
		t.Fatal("surface traversal stole editor indentation", st.Buffer.String())
	}
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("File editor") || st.Buffer.String() != "hello" {
		t.Fatal("surface traversal stole editor outdent", st.Buffer.String())
	}
}
func TestWorkbenchModalAndAutocompleteOwnTab(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Model.CommandsRaw = []json.RawMessage{json.RawMessage(`{"name":"review","source":"extension"}`)}
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Type("/rev")
	tt.Frame()
	if !a.autocomplete.IsOpen() {
		t.Fatal("autocomplete fixture not open")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("surface traversal stole autocomplete accept")
	}
	a.Prompt("Fixture prompt", "", func(string) {})
	tt.Frame()
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Cancel") {
		t.Fatal("Tab did not remain on the prompt’s actual Cancel owner")
	}
}

func TestWorkbenchTerminalOwnsTabAndBacktabInRealPTY(t *testing.T) {
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" || runtime.GOOS == "windows" {
		t.Skip("disposable POSIX PTY opt-in")
	}
	root := t.TempDir()
	output := filepath.Join(root, "keys")
	p := orbiterm.NewPanel(orbiterm.Options{Workspace: root, Command: []string{"/bin/sh", "-c", "stty raw -echo; printf READY; cat > '" + output + "'"}, Env: []string{"PATH=/usr/bin:/bin", "HOME=" + root, "ENV="}})
	defer p.Close()
	if e := p.NewTab(); e != nil {
		t.Fatal(e)
	}
	a := sourceSidebarFixture(t)
	a.Options.Terminal = p
	var focused bool
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); focused = a.surfaceFocus.terminal.FocusWithin(c) }, 1470, 922)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		tt.Frame()
		term := p.ActiveTerminal()
		if term != nil && strings.Contains(term.Text(), "READY") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if term := p.ActiveTerminal(); term == nil || !strings.Contains(term.Text(), "READY") {
		t.Fatal("PTY fixture did not become ready")
	}
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !focused {
		t.Fatal("surface cycle did not reach terminal")
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	deadline = time.Now().Add(3 * time.Second)
	var raw []byte
	for time.Now().Before(deadline) {
		raw, _ = os.ReadFile(output)
		if string(raw) == "\t\x1b[Z" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !focused || string(raw) != "\t\x1b[Z" {
		t.Fatalf("terminal keys stolen: focused=%v bytes=%q", focused, raw)
	}
}

func TestWorkbenchEmptySidebarRemainsReachable(t *testing.T) {
	a := fixture(t)
	a.Model.Sessions = nil
	a.Workspaces.Workspaces = nil
	var focused bool
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); focused = a.surfaceFocus.emptySidebar.Focused(c) }, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !focused {
		t.Fatal("empty sidebar lost its source tab stop")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("empty sidebar Escape lost composer return")
	}
}

func TestWorkbenchExplorerTreeCanReachActualEditor(t *testing.T) {
	a, _, _ := mountedFile(t, "fixture")
	a.ExplorerVisible = true
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if e := tt.Click("Explorer files"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("File editor") {
		t.Fatal("Explorer surface did not reach actual editable owner")
	}
}
func TestWorkbenchSurfaceReturnPreservesComposerSelection(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Draft = "preserve selected text"
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	tt.Type("replacement")
	if a.Draft != "replacement" {
		t.Fatal("surface return damaged native selection", a.Draft)
	}
}

func TestWorkbenchFullReviewSkipsRetainedInvisibleComposer(t *testing.T) {
	a, _ := mountedReview(t)
	a.toggleReviewFull()
	var sidebar, review bool
	tt := ui.NewTester(func(c *ui.Context) {
		a.View(c)
		sidebar = a.sessionList.Handle.FocusWithin(c) || a.surfaceFocus.emptySidebar.Focused(c)
		review = a.reviewState.treeHandle.FocusWithin(c)
	}, 1470, 922)
	if e := tt.Click("Review files"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !sidebar || tt.Focused("Message") {
		t.Fatal("full review traversed retained hidden content")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !review {
		t.Fatal("full-review cycle did not wrap to visible tree")
	}
}

func TestWorkbenchAutocompleteBacktabUsesSurfaceTraversal(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Model.CommandsRaw = []json.RawMessage{json.RawMessage(`{"name":"review","source":"extension"}`)}
	var sidebar bool
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); sidebar = a.sessionList.Handle.FocusWithin(c) }, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Type("/rev")
	tt.Frame()
	if !a.autocomplete.IsOpen() {
		t.Fatal("completion fixture not open")
	}
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !sidebar || a.Draft != "/rev" {
		t.Fatal("ShiftTab must traverse surface without accepting completion", sidebar, a.Draft)
	}
}
