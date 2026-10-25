package orbitui

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
)

func terminalHostFixture(t *testing.T, dir string, command []string) (*App, *orbiterm.Panel, *ui.Tester) {
	t.Helper()
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" || runtime.GOOS == "windows" {
		t.Skip("requires disposable POSIX PTY and pinned Ghostty")
	}
	if err := ghostty.Load(); err != nil {
		t.Fatal(err)
	}
	a := fixture(t)
	p := orbiterm.NewPanel(orbiterm.Options{Workspace: dir, Command: command, Env: []string{"PS1=host> ", "ENV="}})
	t.Cleanup(func() { _ = p.Close() })
	a.Options.Terminal = p
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(a.terminal, 900, 300)
	return a, p, tt
}
func terminalHostWait(t *testing.T, tt *ui.Tester, what string, check func() bool) {
	t.Helper()
	until := time.Now().Add(15 * time.Second)
	for time.Now().Before(until) {
		tt.Frame()
		if check() {
			// Paint the completion just observed: a worker may have published it
			// after the preceding frame sampled the terminal snapshot.
			tt.Frame()
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("waiting for " + what)
}
func terminalHostReady(t *testing.T, tt *ui.Tester, p *orbiterm.Panel) *ghostty.Terminal {
	t.Helper()
	terminalHostWait(t, tt, "shell ready", func() bool { s := p.Snapshot(); return len(s.Tabs) > 0 && !s.Tabs[s.Active].Loading })
	s := p.Snapshot()
	if s.Tabs[s.Active].Error != "" {
		t.Fatal(s.Tabs[s.Active].Error)
	}
	term := p.ActiveTerminal()
	if term == nil {
		t.Fatal("missing terminal")
	}
	return term
}
func terminalHostCapture(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	out := os.Getenv("ORBIT_TERMINAL_HOST_CAPTURE")
	if out == "" {
		return
	}
	if err := os.MkdirAll(out, 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}
func TestTerminalHostExitedRestartPreservesTabAndOtherShell(t *testing.T) {
	firstDir, secondDir := t.TempDir(), t.TempDir()
	a, p, tt := terminalHostFixture(t, firstDir, []string{"/bin/sh"})
	first := terminalHostReady(t, tt, p)
	firstID := p.Snapshot().Tabs[0].ID
	p.SetWorkspace(secondDir)
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	second := terminalHostReady(t, tt, p)
	second.Send([]byte("printf '__SECOND_%s__\\n' untouched\r"))
	terminalHostWait(t, tt, "other tab output", func() bool { return strings.Contains(second.Text(), "__SECOND_untouched__") })
	p.Select(0)
	first.Send([]byte("exit 7\r"))
	terminalHostWait(t, tt, "shell exit", func() bool { return p.Snapshot().Tabs[0].Exited })
	tt.Frame()
	if !tt.HasText("Shell exited") {
		t.Fatal("host omitted persistent shell-exit explanation", tt.Texts())
	}
	terminalHostCapture(t, tt, "exited")
	if err := tt.Click("Restart shell"); err != nil {
		t.Fatal(err)
	}
	terminalHostWait(t, tt, "replacement shell", func() bool {
		return p.ActiveTerminal() != nil && p.ActiveTerminal() != first && !p.Snapshot().Tabs[0].Loading
	})
	next := terminalHostReady(t, tt, p)
	s := p.Snapshot()
	if len(s.Tabs) != 2 || s.Active != 0 || s.Tabs[0].ID != firstID || s.Tabs[0].Directory != firstDir {
		t.Fatal("restart changed tab identity or initial cwd", s)
	}
	if tt.HasText("Shell exited") {
		t.Fatal("exit overlay remained after restart")
	}
	next.Send([]byte("printf '__RESTART_%s__\\n' \"$PWD\"\r"))
	terminalHostWait(t, tt, "original cwd", func() bool { return strings.Contains(next.Text(), "__RESTART_"+firstDir+"__") })
	p.Select(1)
	tt.Frame()
	if p.ActiveTerminal() != second || !strings.Contains(second.Text(), "__SECOND_untouched__") {
		t.Fatal("restarting first tab replaced or cleared other shell")
	}
	select {
	case <-first.Done():
	case <-time.After(time.Second):
		t.Fatal("old shell remained alive")
	}
	if err := tt.Click("Restart terminal"); err != nil {
		t.Fatal(err)
	}
	terminalHostWait(t, tt, "header restart", func() bool {
		return p.ActiveTerminal() != nil && p.ActiveTerminal() != second && !p.Snapshot().Tabs[1].Loading
	})
	if p.Snapshot().Tabs[1].Directory != secondDir {
		t.Fatal("header restart changed cwd")
	}
	if err := tt.Click("Hide terminal"); err != nil {
		t.Fatal(err)
	}
	if p.IsOpen() || p.TabCount() != 2 {
		t.Fatal("hide killed terminal tabs")
	}
	if err := p.Toggle(); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !p.IsOpen() {
		t.Fatal("panel did not reopen")
	}
	terminalHostCapture(t, tt, "restarted")
	if err := tt.Click("Restart terminal"); err != nil {
		t.Fatal(err)
	}
	state := p.Snapshot()
	if err := tt.Click(fmt.Sprintf("%s %d", a.text("terminal.close_tab", "Close tab"), state.Tabs[state.Active].ID)); err != nil {
		t.Fatal(err)
	}
	// A starter finishing after close must never recreate the removed tab.
	time.Sleep(100 * time.Millisecond)
	tt.Frame()
	if p.TabCount() != 1 || p.ActiveTerminal() != next {
		t.Fatal("closed/restarting tab was resurrected or remaining shell replaced")
	}
}
func TestTerminalHostStartupFailureHasPersistentRetry(t *testing.T) {
	dir := t.TempDir()
	shell := filepath.Join(dir, "fixture-shell")
	_, p, tt := terminalHostFixture(t, dir, []string{shell})
	terminalHostWait(t, tt, "startup failure", func() bool { s := p.Snapshot(); return !s.Tabs[0].Loading && s.Tabs[0].Error != "" })
	if !tt.HasText("Terminal could not start") {
		t.Fatal("host omitted persistent startup error", tt.Texts())
	}
	terminalHostCapture(t, tt, "startup-error")
	if err := os.WriteFile(shell, []byte("#!/bin/sh\nexec /bin/sh\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Restart shell"); err != nil {
		t.Fatal(err)
	}
	next := terminalHostReady(t, tt, p)
	if p.Snapshot().Tabs[0].Exited || tt.HasText("Terminal could not start") {
		t.Fatal("error not cleared after successful retry")
	}
	next.Send([]byte("printf '__RECOVERED_%s__\\n' yes\r"))
	terminalHostWait(t, tt, "recovered shell output", func() bool { return strings.Contains(next.Text(), "__RECOVERED_yes__") })
}

// The worker may publish state after the frame was sampled but before the
// predicate is observed. A successful wait must paint that observed state,
// rather than return the previous frame to subsequent visible-text assertions.
func TestTerminalHostWaitPaintsObservedCompletion(t *testing.T) {
	ready := false
	tt := ui.NewTester(func(c *ui.Context) {
		text := "pending"
		if ready {
			text = "observed terminal completion"
		}
		ui.Text(c, text)
	}, 300, 100)
	terminalHostWait(t, tt, "observed completion", func() bool { ready = true; return true })
	if !tt.HasText("observed terminal completion") {
		t.Fatal("wait returned a frame older than its observed completion", tt.Texts())
	}
}
