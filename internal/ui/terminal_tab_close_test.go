package orbitui

import (
	"fmt"
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTerminalTabCloseTargetsInactivePTYAndPreservesOwner(t *testing.T) {
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" || runtime.GOOS == "windows" {
		t.Skip("disposable POSIX PTY opt-in")
	}
	root := t.TempDir()
	p := orbiterm.NewPanel(orbiterm.Options{Workspace: root, Command: []string{"/bin/sh", "-c", "printf 'READY PID=%s\\n' $$; exec cat"}, Env: []string{"PATH=/usr/bin:/bin", "HOME=" + root, "ENV="}})
	defer p.Close()
	a := sourceSidebarFixture(t)
	a.Options.Terminal = p
	focused := false
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); focused = a.surfaceFocus.terminal.FocusWithin(c) }, 1470, 922)
	var terms []*ghostty.Terminal
	var ids []uint64
	for i := 0; i < 3; i++ {
		if e := p.NewTab(); e != nil {
			t.Fatal(e)
		}
		terminalHostWait(t, tt, "new shell ready", func() bool { term := p.ActiveTerminal(); return term != nil && strings.Contains(term.Text(), "READY") })
		terms = append(terms, p.ActiveTerminal())
		s := p.Snapshot()
		ids = append(ids, s.Tabs[s.Active].ID)
	}
	closeTab := func(id uint64) {
		t.Helper()
		label := fmt.Sprintf("%s %d", a.text("terminal.close_tab", "Close tab"), id)
		if e := tt.Click(label); e != nil {
			t.Fatal("missing per-tab close control", e)
		}
		tt.Frame()
	}
	// Toolbar x hides instead of terminating; reopening retains exact grid/process.
	beforeText := terms[2].Text()
	if !strings.Contains(beforeText, "READY PID=") {
		t.Fatal("fixture did not report shell PID")
	}
	for _, id := range ids {
		chip, ok := tt.Find(fmt.Sprintf("Terminal tab %d", id))
		if !ok {
			t.Fatal("missing chip", id)
		}
		close, ok := tt.Find(fmt.Sprintf("%s %d", a.text("terminal.close_tab", "Close tab"), id))
		if !ok {
			t.Fatal("missing close", id)
		}
		if close.X < chip.X || close.X+close.W > chip.X+chip.W || close.Y < chip.Y || close.Y+close.H > chip.Y+chip.H {
			t.Fatal("close control not inside its tab", chip, close)
		}
	}
	if err := tt.Click("Hide terminal"); err != nil {
		t.Fatal(err)
	}
	if p.IsOpen() || p.TabCount() != 3 {
		t.Fatal("hide closed tabs")
	}
	for _, term := range terms {
		select {
		case <-term.Done():
			t.Fatal("hide terminated PTY")
		default:
		}
	}
	if err := p.Toggle(); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if p.ActiveTerminal() != terms[2] || terms[2].Text() != beforeText {
		t.Fatal("hide/reopen replaced process or content")
	}
	closeTab(ids[0])
	s := p.Snapshot()
	if len(s.Tabs) != 2 || s.Tabs[s.Active].ID != ids[2] || p.ActiveTerminal() != terms[2] || !focused {
		t.Fatal("inactive close selected stale index or lost shell owner", s, focused)
	}
	select {
	case <-terms[0].Done():
	case <-time.After(3 * time.Second):
		t.Fatal("closed inactive PTY still running")
	}
	for _, term := range terms[1:] {
		select {
		case <-term.Done():
			t.Fatal("unrelated PTY was closed")
		default:
		}
	}
	// A held intent for the removed ID must not close its former index's shell.
	if err := closeTerminalTab(p, ids[0]); err != nil {
		t.Fatal(err)
	}
	if p.TabCount() != 2 || p.ActiveTerminal() != terms[2] {
		t.Fatal("stale tab ID closed a replacement")
	}
	closeTab(ids[2])
	s = p.Snapshot()
	if len(s.Tabs) != 1 || s.Tabs[s.Active].ID != ids[1] || p.ActiveTerminal() != terms[1] || !focused {
		t.Fatal("active close did not focus surviving neighbor", s, focused)
	}
	closeTab(ids[1])
	if p.IsOpen() || p.TabCount() != 0 {
		t.Fatal("last close left empty panel")
	}
	for _, term := range terms {
		select {
		case <-term.Done():
		case <-time.After(3 * time.Second):
			t.Fatal("closed PTY did not exit")
		}
	}
}

func TestTerminalHeldCloseDoesNotRetargetWorkspaceABA(t *testing.T) {
	aRoot, bRoot := t.TempDir(), t.TempDir()
	a, p, tt := terminalHostFixture(t, aRoot, []string{"/bin/sh", "-c", "printf 'READY PID=%s\\n' $$; exec cat"})
	old := terminalHostReady(t, tt, p)
	oldID := p.Snapshot().Tabs[0].ID
	label := fmt.Sprintf("%s %d", a.text("terminal.close_tab", "Close tab"), oldID)
	rect, ok := tt.Find(label)
	if !ok {
		t.Fatal("close target absent")
	}
	x, y := float32(rect.X+rect.W/2), float32(rect.Y+rect.H/2)
	tt.Press(x, y)
	if p.TabCount() != 1 {
		t.Fatal("close acted on press instead of release")
	}
	held := func() error { return closeTerminalTab(p, oldID) }
	if err := p.CloseTab(0); err != nil {
		t.Fatal(err)
	}
	p.SetWorkspace(bRoot)
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	b := terminalHostReady(t, tt, p)
	p.SetWorkspace(aRoot)
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	again := terminalHostReady(t, tt, p)
	before := p.Snapshot()
	if before.Tabs[0].ID == oldID || before.Tabs[1].ID == oldID {
		t.Fatal("workspace switch reused removed ID")
	}
	if err := held(); err != nil {
		t.Fatal(err)
	}
	tt.Release(x, y)
	tt.Frame()
	after := p.Snapshot()
	if len(after.Tabs) != 2 || after.Tabs[0].ID != before.Tabs[0].ID || after.Tabs[1].ID != before.Tabs[1].ID || p.ActiveTerminal() != again {
		t.Fatal("held old press targeted ABA replacement", after)
	}
	for _, term := range []*ghostty.Terminal{b, again} {
		select {
		case <-term.Done():
			t.Fatal("stale close terminated replacement")
		default:
		}
	}
	select {
	case <-old.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("old target did not close")
	}
}
