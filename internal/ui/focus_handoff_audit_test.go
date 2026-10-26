package orbitui

import (
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Source-only audit tests. Rust299 session.rs2303-2316 and pickers.rs360-372.
// All actions are mounted tester input; no Controller, RPC or real sessions.
func TestAuditSidebarHideHandsFocusToComposer(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Draft = ""
	a.Options.OnAction = func(Action) {}
	sidebarFocused := false
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); sidebarFocused = a.sessionList.Handle.FocusWithin(c) }, 1470, 922)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("Message") {
		t.Fatal("actual composer text input not focused")
	}
	tt.Type("甲乙")
	end, ok := tt.TextCaret()
	if !ok {
		t.Fatal("no native caret")
	}
	tt.Key(0, ui.KeyLeft)
	middle, ok := tt.TextCaret()
	if !ok || middle.X >= end.X {
		t.Fatalf("native Left did not move caret: %v -> %v", end, middle)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	if !sidebarFocused {
		t.Fatal("fixture did not focus actual sidebar")
	}
	tt.Key(ui.Cmd, ui.KeyB)
	tt.Frame()
	if a.SidebarVisible {
		t.Fatal("hide shortcut did not execute")
	}
	t.Logf("after hide composerFocused=%v sidebarFocused=%v cursor=%d", tt.Focused("Message"), sidebarFocused, a.sessionCursor)
	tt.Type("X")
	if a.Draft != "甲X乙" {
		t.Fatalf("hidden sidebar swallowed typing / lost caret: draft=%q", a.Draft)
	}
	if !tt.Focused("Message") {
		t.Fatal("composer did not own resumed input")
	}
}

func TestAuditPaletteEscapeFromSidebarHandsFocusToComposer(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Draft = ""
	a.Options.OnAction = func(Action) {}
	sidebarFocused := false
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); sidebarFocused = a.sessionList.Handle.FocusWithin(c) }, 1470, 922)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("Message") {
		t.Fatal("actual composer text input not focused")
	}
	tt.Type("甲乙")
	end, ok := tt.TextCaret()
	if !ok {
		t.Fatal("no native caret")
	}
	tt.Key(0, ui.KeyLeft)
	middle, ok := tt.TextCaret()
	if !ok || middle.X >= end.X {
		t.Fatalf("native Left did not move caret: %v -> %v", end, middle)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	if !sidebarFocused {
		t.Fatal("fixture did not focus actual sidebar")
	}
	tt.Key(ui.Cmd, ui.KeyP)
	tt.Frame()
	if !a.paletteOpen || !tt.Focused("Command palette search") {
		t.Fatal("palette/search not open and focused")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.paletteOpen {
		t.Fatal("Escape did not dismiss palette")
	}
	t.Logf("after Escape composerFocused=%v sidebarFocused=%v restoreComposer=%v", tt.Focused("Message"), sidebarFocused, a.restoreComposerAfterModal)
	tt.Type("X")
	if a.Draft != "甲X乙" {
		t.Fatalf("palette dismissal did not return typing/caret: draft=%q", a.Draft)
	}
	if !tt.Focused("Message") {
		t.Fatal("composer did not own resumed input")
	}
}

func TestFocusHandoffSidebarPointerAndEmpty(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(map[bool]string{false: "list", true: "empty"}[empty], func(t *testing.T) {
			a := sourceSidebarFixture(t)
			a.Options.OnAction = func(Action) {}
			if empty {
				a.Model.Sessions = nil
				a.Workspaces.Workspaces = nil
			}
			focused := false
			tt := ui.NewTester(func(c *ui.Context) {
				a.View(c)
				focused = a.sessionList.Handle.FocusWithin(c) || a.surfaceFocus.emptySidebar.Focused(c)
			}, 1470, 922)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			tt.Type("甲乙")
			tt.Key(0, ui.KeyLeft)
			tt.Key(0, ui.KeyTab)
			tt.Frame()
			if !focused {
				t.Fatal("actual sidebar not focused")
			}
			if err := tt.Click("Toggle sidebar"); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			if a.SidebarVisible || !tt.Focused("Message") {
				t.Fatal("pointer hide did not hand focus back")
			}
			tt.Type("X")
			if a.Draft != "甲X乙" {
				t.Fatal("caret changed", a.Draft)
			}
			tt.Key(ui.Cmd, ui.KeyZ)
			if a.Draft != "甲乙" {
				t.Fatal("handoff damaged native undo", a.Draft)
			}
		})
	}
}
func TestFocusHandoffSidebarDoesNotStealOtherInput(t *testing.T) {
	a := sourceSidebarFixture(t)
	other := ""
	a.Options.Overlay = func(c *ui.Context) {
		ui.TextInput(c, &other).Label("Other editable surface").Absolute().Left(900).Top(40).Size(180, 26)
	}
	tt := ui.NewTester(a.View, 1470, 922)
	if err := tt.Click("Other editable surface"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		tt.Key(ui.Cmd, ui.KeyB)
		tt.Frame()
		if !tt.Focused("Other editable surface") {
			t.Fatal("sidebar toggle stole another editable surface")
		}
		tt.Type("X")
	}
	if other != "XX" || a.Draft != "" {
		t.Fatalf("input routed wrongly other=%q composer=%q", other, a.Draft)
	}
}
func TestFocusHandoffPaletteCommandKeepsDestination(t *testing.T) {
	a := sourceSidebarFixture(t)
	sidebar := false
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); sidebar = a.sessionList.Handle.FocusWithin(c) }, 1470, 922)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("draft")
	tt.Key(ui.Cmd, ui.KeyP)
	tt.Type("focus sessions")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if a.paletteOpen || !sidebar || tt.Focused("Message") {
		t.Fatalf("dismiss restoration stole command destination: palette=%v sidebar=%v composer=%v query=%q title=%q", a.paletteOpen, sidebar, tt.Focused("Message"), a.paletteQuery, a.Tr("command_palette.focus_sessions"))
	}
	tt.Key(ui.Cmd, ui.KeyP)
	tt.Type("settings preferences configure")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.page != "settings" || a.paletteOpen {
		t.Fatal("settings command did not open destination", a.page)
	}
	if tt.Focused("Message") {
		t.Fatal("palette restored hidden composer")
	}
}
func TestFocusHandoffPaletteOffChatDoesNotNavigate(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Navigate("settings")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Key(ui.Cmd, ui.KeyP)
	tt.Frame()
	if !a.paletteOpen {
		t.Fatal("off-chat palette did not open")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.page != "settings" || a.paletteOpen {
		t.Fatal("dismiss changed page or stayed open")
	}
}

func TestFocusHandoffSidebarPreservesNativeComposition(t *testing.T) {
	a := sourceSidebarFixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("甲乙")
	tt.Key(0, ui.KeyLeft)
	tt.Compose("候", 1)
	if err := tt.Click("Toggle sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("sidebar button stole composing owner")
	}
	tt.Type("候")
	if a.Draft != "甲候乙" {
		t.Fatal("composition was reset/duplicated", a.Draft)
	}
}
func TestFocusHandoffSidebarPreservesRealTerminal(t *testing.T) {
	if runtime.GOOS == "windows" || os.Getenv("ORBIT_TERMINAL_TEST") != "1" {
		t.Skip("disposable POSIX PTY opt-in")
	}
	root := t.TempDir()
	output := filepath.Join(root, "typed")
	p := orbiterm.NewPanel(orbiterm.Options{Workspace: root, Command: []string{"/bin/sh", "-c", "stty raw -echo; printf READY; cat > '" + output + "'"}, Env: []string{"PATH=/usr/bin:/bin", "HOME=" + root, "ENV="}})
	defer p.Close()
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	a := sourceSidebarFixture(t)
	a.Options.Terminal = p
	focused := false
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); focused = a.surfaceFocus.terminal.FocusWithin(c) }, 1470, 922)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		tt.Frame()
		if term := p.ActiveTerminal(); term != nil && strings.Contains(term.Text(), "READY") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if term := p.ActiveTerminal(); term == nil || !strings.Contains(term.Text(), "READY") {
		t.Fatal("controlled PTY not ready")
	}
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !focused {
		t.Fatal("fixture real terminal not focused")
	}
	if err := tt.Click("Toggle sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !focused {
		t.Fatal("sidebar pointer hide stole real terminal")
	}
	tt.Type("owned-terminal")
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		b, _ := os.ReadFile(output)
		if string(b) == "owned-terminal" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	b, _ := os.ReadFile(output)
	t.Fatalf("terminal did not receive marker %q", b)
}
