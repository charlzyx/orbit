package terminal

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
)

func TestShellProfiles(t *testing.T) {
	for path, want := range map[string]Profile{"/bin/zsh": Zsh, "/usr/local/bin/bash": Bash, "/opt/homebrew/bin/fish": Fish, "/bin/sh": Other, `C:\Windows\System32\cmd.exe`: Other, "ZSH.EXE": Zsh} {
		if got := ShellProfile(path); got != want {
			t.Errorf("%s: %s != %s", path, got, want)
		}
	}
}
func TestPaneShortcutsPreserveShellKeys(t *testing.T) {
	for _, os := range []string{"linux", "windows", "darwin"} {
		for _, key := range []ui.Key{ui.KeyT, ui.KeyW, ui.KeyTab, ui.KeyC, ui.KeyD, ui.KeyJ} {
			if got := KeyAction(os, ui.Ctrl, key); got != NoAction {
				t.Errorf("%s Ctrl+%v swallowed", os, key)
			}
		}
		mods := ui.Ctrl | ui.Shift
		if os == "darwin" {
			mods = ui.Super
		}
		if KeyAction(os, mods, ui.KeyT) != NewTabAction || KeyAction(os, mods, ui.KeyW) != CloseTabAction {
			t.Fatalf("missing tabs for %s", os)
		}
	}
}
func TestPanelHeightAndClose(t *testing.T) {
	p := NewPanel(Options{})
	if p.Height() != 260 {
		t.Fatal(p.Height())
	}
	p.SetHeight(2)
	if p.Height() != 96 {
		t.Fatal(p.Height())
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.NewTab(); err == nil {
		t.Fatal("opened closed panel")
	}
}
func requireTerminal(t *testing.T) {
	t.Helper()
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" {
		t.Skip("set ORBIT_TERMINAL_TEST=1 for pinned libghostty-vt + disposable PTY/native-view integration")
	}
	if runtime.GOOS == "windows" {
		t.Skip("POSIX disposable shell integration; ConPTY requires Windows runner")
	}
	// Tests must not depend on a writable real home. Respect an explicit
	// cache (CI may prewarm it), otherwise use only this test's scratch tree.
	if runtime.GOOS == "linux" && os.Getenv("XDG_CACHE_HOME") == "" {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())
	}
	if err := ghostty.Load(); err != nil {
		t.Fatal(err)
	}
}
func until(t *testing.T, tt *ui.Tester, what string, f func() bool) {
	t.Helper()
	guard := time.NewTimer(time.Minute)
	defer guard.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for !f() {
		select {
		case <-guard.C:
			t.Fatalf("hang waiting for %s", what)
		case <-tick.C:
			if tt != nil {
				tt.Frame()
			}
		}
	}
}
func ready(t *testing.T, p *Panel, tt *ui.Tester) *ghostty.Terminal {
	t.Helper()
	until(t, tt, "terminal creation", func() bool {
		s := p.Snapshot()
		if len(s.Tabs) == 0 {
			return false
		}
		return !s.Tabs[s.Active].Loading
	})
	s := p.Snapshot()
	if err := s.Tabs[s.Active].Error; err != "" {
		t.Fatal(err)
	}
	term := p.ActiveTerminal()
	if term == nil {
		t.Fatal("no terminal")
	}
	return term
}
func TestLiveTabLifecycleAndNativeGrid(t *testing.T) {
	requireTerminal(t)
	a, b := t.TempDir(), t.TempDir()
	p := NewPanel(Options{Workspace: a, Command: []string{"/bin/sh"}, Env: []string{"PS1=orbit> ", "ENV="}})
	defer p.Close()
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(func(c *ui.Context) { View(c, p) }, 640, 240)
	first := ready(t, p, tt)
	until(t, tt, "shell prompt", func() bool { return strings.Contains(first.Text(), "orbit>") })
	first.Send([]byte("printf '__FIRST_CWD__%s\\n' \"$PWD\"\r"))
	until(t, tt, "workspace cwd", func() bool { return strings.Contains(first.Text(), "__FIRST_CWD__"+a) })
	initialCols, initialRows := first.Size()
	tt.SetSize(900, 380)
	tt.Frame()
	cols, rows := first.Size()
	if cols <= initialCols || rows <= initialRows {
		t.Fatalf("resize did not reach PTY grid: %dx%d -> %dx%d", initialCols, initialRows, cols, rows)
	}
	p.SetWorkspace(b)
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	second := ready(t, p, tt)
	until(t, tt, "second prompt", func() bool { return strings.Contains(second.Text(), "orbit>") })
	second.Send([]byte("printf '__SECOND_CWD__%s\\n' \"$PWD\"\r"))
	until(t, tt, "new workspace cwd", func() bool { return strings.Contains(second.Text(), "__SECOND_CWD__"+b) })
	p.Hide()
	first.Send([]byte("printf '__BACKGROUND_%s__\\n' alive\r"))
	until(t, nil, "hidden background tab output", func() bool { return strings.Contains(first.Text(), "__BACKGROUND_alive__") })
	if err := p.Toggle(); err != nil {
		t.Fatal(err)
	}
	p.Select(0)
	tt.Frame()
	if p.ActiveTerminal() != first {
		t.Fatal("showing panel replaced shell")
	}
	first.Send([]byte("printf '\\033]2;build title\\007'\r"))
	until(t, tt, "OSC title", func() bool { return p.Snapshot().Tabs[0].Label == "build title" })
	tt.Command("selectAll")
	tt.Command("copy")
	if !strings.Contains(tt.Clipboard(), "__FIRST_CWD__"+a) {
		t.Fatalf("native selection/copy lost output: %q", tt.Clipboard())
	}
	if err := p.Restart(); err != nil {
		t.Fatal(err)
	}
	restarted := ready(t, p, tt)
	if restarted == first {
		t.Fatal("restart kept original process")
	}
	until(t, tt, "restarted prompt", func() bool { return strings.Contains(restarted.Text(), "orbit>") })
	restarted.Send([]byte("printf '__RESTART_CWD__%s\\n' \"$PWD\"\r"))
	until(t, tt, "original cwd on restart", func() bool { return strings.Contains(restarted.Text(), "__RESTART_CWD__"+a) })
	select {
	case <-first.Done():
	case <-time.After(time.Minute):
		t.Fatal("restarted shell did not stop")
	}
	if dir := os.Getenv("ORBIT_TEST_IMAGES"); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, "terminal-native-grid.png"))
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	p.Select(1)
	second.Send([]byte("exit 7\r"))
	until(t, tt, "exit status", func() bool { return p.Snapshot().Tabs[1].Exited })
	if code := p.Snapshot().Tabs[1].ExitCode; code != 7 {
		t.Fatalf("exit code %d", code)
	}
	if err := p.CloseTab(0); err != nil {
		t.Fatal(err)
	}
	if s := p.Snapshot(); s.Active != 0 || len(s.Tabs) != 1 || s.Tabs[0].Directory != b {
		t.Fatal(s)
	}
	if err := p.CloseActive(); err != nil {
		t.Fatal(err)
	}
	if p.IsOpen() || p.TabCount() != 0 {
		t.Fatal("last tab did not close panel")
	}
}
func TestLiveBracketedPasteAndScopedChords(t *testing.T) {
	requireTerminal(t)
	p := NewPanel(Options{Workspace: t.TempDir(), Command: []string{"/bin/sh", "-c", "stty raw -echo; printf '\\033[?2004hREADY\\r\\n'; od -An -tx1"}})
	defer p.Close()
	if err := p.NewTab(); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(func(c *ui.Context) { View(c, p) }, 640, 240)
	term := ready(t, p, tt)
	until(t, tt, "raw-shell ready", func() bool { return strings.Contains(term.Text(), "READY") })
	// A single 16-byte block causes od to emit without EOF. Newlines are CR
	// within bracketed paste, never an unguarded shell submit.
	tt.SetClipboard("ABCD")
	mods := ui.Ctrl | ui.Shift
	if runtime.GOOS == "darwin" {
		mods = ui.Super
	}
	tt.Key(mods, ui.KeyV)
	until(t, tt, "bracketed bytes", func() bool { return strings.Contains(term.Text(), "1b 5b 32 30 30 7e 41 42 43 44 1b 5b 32 30 31 7e") })
	before := p.TabCount()
	tt.Key(ui.Ctrl, ui.KeyT)
	tt.Key(ui.Ctrl, ui.KeyW)
	if p.TabCount() != before {
		t.Fatal("shell readline keys changed tabs")
	}
	tt.Key(mods, ui.KeyT)
	until(t, tt, "new-tab chord", func() bool { return p.TabCount() == before+1 })
	if err := p.Close(); err != nil {
		t.Fatal(fmt.Errorf("close: %w", err))
	}
}

func TestLiveDirectCommandTabPreservesArgv(t *testing.T) {
	requireTerminal(t)
	dir := t.TempDir()
	p := NewPanel(Options{Workspace: dir, Command: []string{"/bin/sh"}})
	defer p.Close()
	literal := "argument with spaces; $(touch must-not-exist)"
	if err := p.NewCommandTab([]string{"/bin/sh", "-c", "printf '__ARGV__%s\\n' \"$1\"; exit 4", "fixture", literal}); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(func(c *ui.Context) { View(c, p) }, 800, 240)
	term := ready(t, p, tt)
	until(t, tt, "direct argv output", func() bool { return strings.Contains(term.Text(), "__ARGV__"+literal) })
	until(t, tt, "direct command exit", func() bool { return p.Snapshot().Tabs[0].Exited })
	if p.Snapshot().Tabs[0].ExitCode != 4 {
		t.Fatal(p.Snapshot())
	}
	if _, err := os.Stat(filepath.Join(dir, "must-not-exist")); !os.IsNotExist(err) {
		t.Fatal("argument was interpreted by a shell")
	}
	if err := p.Restart(); err != nil {
		t.Fatal(err)
	}
	again := ready(t, p, tt)
	until(t, tt, "restarted same direct argv", func() bool { return strings.Contains(again.Text(), "__ARGV__"+literal) })
}

func TestLiveHiddenPTYKeepsRunningWithoutWindowFrames(t *testing.T) {
	requireTerminal(t)
	p := NewPanel(Options{Workspace: t.TempDir(), Command: []string{"/bin/sh"}, Env: []string{"PS1=orbit> ", "ENV="}, ReduceMotion: true})
	defer p.Close()
	if e := p.NewTab(); e != nil {
		t.Fatal(e)
	}
	frames := 0
	tt := ui.NewTester(func(c *ui.Context) { frames++; View(c, p) }, 640, 240)
	term := ready(t, p, tt)
	until(t, tt, "visible shell ready", func() bool { return strings.Contains(term.Text(), "orbit>") })
	p.Hide()
	tt.Frame()
	before := frames
	term.Send([]byte("printf '__HIDDEN_%s__\\n' alive\r"))
	until(t, nil, "hidden shell completed command", func() bool { return strings.Contains(term.Text(), "__HIDDEN_alive__") })
	tt.Advance(time.Second)
	if frames != before {
		t.Fatalf("hidden PTY requested %d window frames", frames-before)
	}
	if e := p.Toggle(); e != nil {
		t.Fatal(e)
	}
	tt.Advance(time.Second)
	if frames == before || p.ActiveTerminal() != term {
		t.Fatal("showing did not redraw the existing PTY")
	}
	// A visible preference can belong to an unrendered settings/parked page.
	p.SetVisible(false)
	tt.Frame()
	before = frames
	term.Send([]byte("printf '__PARKED_%s__\\n' alive\r"))
	until(t, nil, "parked shell completed command", func() bool { return strings.Contains(term.Text(), "__PARKED_alive__") })
	tt.Advance(time.Second)
	if frames != before || !p.IsOpen() {
		t.Fatal("page visibility changed saved openness or invalidated window")
	}
	p.SetVisible(true)
	tt.Advance(time.Second)
	if frames == before {
		t.Fatal("restored page did not redraw")
	}
}
