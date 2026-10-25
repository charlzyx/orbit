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

func TestWorkbenchExitedTerminalTargetsActionableRestart(t *testing.T) {
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" || runtime.GOOS == "windows" {
		t.Skip("disposable POSIX PTY opt-in")
	}
	root := t.TempDir()
	marker := filepath.Join(root, "starts")
	p := orbiterm.NewPanel(orbiterm.Options{Workspace: root, Command: []string{"/bin/sh", "-c", "printf x >> '" + marker + "'; exit 0"}, Env: []string{"PATH=/usr/bin:/bin", "HOME=" + root, "ENV="}})
	defer p.Close()
	if e := p.NewTab(); e != nil {
		t.Fatal(e)
	}
	a := sourceSidebarFixture(t)
	a.Options.Terminal = p
	tt := ui.NewTester(a.View, 1470, 922)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		tt.Frame()
		s := p.Snapshot()
		if len(s.Tabs) > 0 && s.Tabs[0].Exited {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	s := p.Snapshot()
	if len(s.Tabs) != 1 || !s.Tabs[0].Exited || s.Tabs[0].ExitCode != 0 {
		t.Fatal("real exit0 fixture did not settle", s)
	}
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Restart shell") {
		t.Fatal("exited grid is an inert Tab stop; restart has no keyboard focus")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("exited terminal restart trapped surface traversal")
	}
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Restart shell") {
		t.Fatal("reverse traversal lost restart stop")
	}
	tt.Key(0, ui.KeyEnter)
	deadline = time.Now().Add(3 * time.Second)
	var raw []byte
	for time.Now().Before(deadline) {
		tt.Frame()
		raw, _ = os.ReadFile(marker)
		if string(raw) == "xx" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if string(raw) != "xx" {
		t.Fatal("Enter did not invoke real restart", string(raw))
	}
}
func TestWorkbenchMarkdownPreviewRootIsActualFocusStop(t *testing.T) {
	a, _, path := mountedFile(t, "source")
	md := filepath.Join(filepath.Dir(path), "preview.md")
	if e := os.WriteFile(md, []byte("# Preview\n\nfixture body\n"), 0600); e != nil {
		t.Fatal(e)
	}
	d, e := a.files.editor.Open(md)
	if e != nil {
		t.Fatal(e)
	}
	a.syncDocument(d)
	if e := a.files.editor.SetPreview(md, true); e != nil {
		t.Fatal(e)
	}
	a.ExplorerVisible = true
	var preview, tree bool
	tt := ui.NewTester(func(c *ui.Context) {
		a.View(c)
		preview = a.files.handle.Focused(c)
		tree = a.files.treeHandle.FocusWithin(c)
	}, 1470, 922)
	if e := tt.Click("Explorer files"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !preview {
		t.Fatal("preview root is not a real focus stop")
	}
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !tree {
		t.Fatal("preview reverse traversal did not return to Explorer")
	}
	if d, _ := a.files.editor.Active(); !d.Preview {
		t.Fatal("focus traversal toggled preview")
	}
}

func TestWorkbenchTerminalExitTransfersOnlyItsOwnFocus(t *testing.T) {
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" || runtime.GOOS == "windows" {
		t.Skip("disposable POSIX PTY opt-in")
	}
	for _, foreground := range []bool{false, true} {
		t.Run(map[bool]string{false: "background", true: "foreground"}[foreground], func(t *testing.T) {
			root := t.TempDir()
			release := filepath.Join(root, "release")
			p := orbiterm.NewPanel(orbiterm.Options{Workspace: root, Command: []string{"/bin/sh", "-c", "printf READY; while [ ! -f '" + release + "' ]; do sleep 0.01; done; exit 0"}, Env: []string{"PATH=/usr/bin:/bin", "HOME=" + root, "ENV="}})
			defer p.Close()
			if e := p.NewTab(); e != nil {
				t.Fatal(e)
			}
			a := sourceSidebarFixture(t)
			a.Options.Terminal = p
			tt := ui.NewTester(a.View, 1470, 922)
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
				t.Fatal("live fixture not ready")
			}
			if e := tt.Click("Message"); e != nil {
				t.Fatal(e)
			}
			if foreground {
				tt.Key(0, ui.KeyTab)
				tt.Frame()
				tt.Key(0, ui.KeyTab)
				tt.Frame()
			}
			if e := os.WriteFile(release, nil, 0600); e != nil {
				t.Fatal(e)
			}
			deadline = time.Now().Add(3 * time.Second)
			for time.Now().Before(deadline) {
				tt.Frame()
				s := p.Snapshot()
				if len(s.Tabs) > 0 && s.Tabs[0].Exited {
					tt.Frame()
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if !p.Snapshot().Tabs[0].Exited {
				t.Fatal("fixture did not exit")
			}
			if foreground && !tt.Focused("Restart shell") {
				t.Fatal("focused exit retained inert grid instead of restart")
			}
			if !foreground && !tt.Focused("Message") {
				t.Fatal("background exit stole composer focus")
			}
		})
	}
}
