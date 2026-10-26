package orbitui

import (
	"context"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCommitModelFullWindowLayoutAndSafeClick(t *testing.T) {
	natural := map[string]float32{}
	for _, size := range []struct{ w, sidebar int }{{1470, 248}, {1200, 248}, {1024, 248}, {1470, 400}, {1200, 400}} {
		t.Run(fmt.Sprintf("window%d-sidebar%d", size.w, size.sidebar), func(t *testing.T) {
			home, dir := t.TempDir(), t.TempDir()
			if out, e := exec.Command("git", "init", "-q", dir).CombinedOutput(); e != nil {
				t.Fatal(e, string(out))
			}
			if e := os.WriteFile(filepath.Join(dir, "fixture.txt"), []byte("synthetic only\n"), 0600); e != nil {
				t.Fatal(e)
			}
			x, e := NewController(RunOptions{Home: home, Workspace: dir, PiBinary: filepath.Join(home, "nonexistent-fixture-pi"), Offline: true, NoExtensions: true})
			if e != nil {
				t.Fatal(e)
			}
			defer x.Close()
			if e = x.Store.SaveCommitModel(&preferences.Model{Provider: strings.Repeat("long-provider-", 12), ID: strings.Repeat("long-model-", 16)}); e != nil {
				t.Fatal(e)
			}
			var calls atomic.Int32
			x.Git.Options.GenerateSelectedCommit = func(context.Context, bool, agent.ModelChoice) (workspace.GeneratedText, error) {
				calls.Add(1)
				return workspace.GeneratedText{}, context.Canceled
			}
			deadline := time.Now().Add(5 * time.Second)
			for x.Git.Busy && time.Now().Before(deadline) {
				x.Git.Tick()
				time.Sleep(time.Millisecond)
			}
			if x.Git.Busy || x.Git.Error != nil {
				t.Fatal(x.Git.Busy, x.Git.Error)
			}
			x.Git.IncludeUnstaged = true
			x.UI.SidebarWidth = float32(size.sidebar)
			x.UI.Navigate("git")
			tt := ui.NewTester(x.UI.View, size.w, 922)
			tt.Frame()
			deadline = time.Now().Add(5 * time.Second)
			for x.Git.Busy && time.Now().Before(deadline) {
				x.Git.Tick()
				tt.Frame()
				time.Sleep(time.Millisecond)
			}
			if x.Git.Busy {
				t.Fatal("mounted initial refresh not settled")
			}
			tt.Frame()
			path := os.Getenv("ORBIT_LAYOUT_PROOF_DIR")
			if path != "" {
				f, e := os.Create(filepath.Join(path, fmt.Sprintf("commit-window%d-sidebar%d.png", size.w, size.sidebar)))
				if e != nil {
					t.Fatal(e)
				}
				if e = png.Encode(f, tt.Image()); e != nil {
					t.Fatal(e)
				}
				if e = f.Close(); e != nil {
					t.Fatal(e)
				}
			}
			labels := []string{"Include unstaged changes", "Commit model", "Generate message", "Commit", "Commit & push"}

			for index, label := range labels {
				r, ok := tt.Find(label)
				t.Logf("%s bounds=%+v window=%d sidebar=%d", label, r, size.w, size.sidebar)
				if size.w == 1470 && size.sidebar == 248 {
					natural[label] = r.W
				}
				if r.W < natural[label]-.5 {
					t.Errorf("control label visibly clipped: %s width=%f natural=%f", label, r.W, natural[label])
				}
				if !ok || r.X < float32(size.sidebar) || r.X+r.W > float32(size.w)+.5 || r.W <= 0 || r.Y < 0 || r.Y+r.H > 922 {
					t.Errorf("control clipped/outside actual window: %s %+v", label, r)
				}
				for _, previous := range labels[:index] {
					previousRect, _ := tt.Find(previous)
					if r.X < previousRect.X+previousRect.W-.5 && r.X+r.W > previousRect.X+.5 && r.Y < previousRect.Y+previousRect.H-.5 && r.Y+r.H > previousRect.Y+.5 {
						t.Errorf("controls overlap: %s %+v and %s %+v", label, r, previous, previousRect)
					}
				}
			}
			// Generators are deliberately canceled before any commit/index/push code.
			// These are mounted MyGo hit tests of the existing buttons, not git mutations.
			for _, label := range []string{"Generate message", "Commit", "Commit & push"} {
				before := calls.Load()
				if e := tt.Click(label); e != nil {
					t.Error(e)
					continue
				}
				deadline = time.Now().Add(3 * time.Second)
				for x.Git.Busy && time.Now().Before(deadline) {
					x.Git.Tick()
					tt.Frame()
					time.Sleep(time.Millisecond)
				}
				if calls.Load() != before+1 {
					t.Errorf("visible button not hit: %s calls %d->%d", label, before, calls.Load())
				}
				x.Git.Error = nil
				tt.Frame()
			}

			// Hold a real generation; the model picker must not open while busy.
			release := make(chan struct{})
			x.Git.Options.GenerateSelectedCommit = func(ctx context.Context, _ bool, _ agent.ModelChoice) (workspace.GeneratedText, error) {
				select {
				case <-release:
				case <-ctx.Done():
				}
				return workspace.GeneratedText{}, context.Canceled
			}
			if e := tt.Click("Generate message"); e != nil {
				t.Fatal(e)
			}
			if !x.Git.Busy {
				close(release)
				t.Fatal("real generation did not start")
			}
			if e := tt.Click("Commit model"); e != nil {
				close(release)
				t.Fatal(e)
			}
			if tt.HasText("Search commit models") {
				close(release)
				t.Fatal("busy picker opened")
			}
			close(release)
			deadline = time.Now().Add(5 * time.Second)
			for x.Git.Busy && time.Now().Before(deadline) {
				x.Git.Tick()
				tt.Frame()
				time.Sleep(time.Millisecond)
			}
			if x.Git.Busy {
				t.Fatal("held generator did not settle")
			}
			x.Git.Error = nil
			tt.Frame()
			if e := tt.Click("Commit model"); e != nil {
				t.Fatal(e)
			}
			tt.Frame()
			panel, ok := tt.Find("Commit model picker")
			anchor, _ := tt.Find("Commit model")
			if !ok || panel.X < 0 || panel.X+panel.W > float32(size.w)+.5 || panel.Y < 0 || panel.Y+panel.H > anchor.Y+.5 {
				t.Fatal("popover not anchored above chip within window", panel, anchor)
			}
			tt.Key(0, ui.KeyEscape)
			tt.Frame()
			tt.Type("escape draft")
			if x.Git.CommitMessage != "escape draft" || tt.HasText("Search commit models") {
				t.Fatal("Escape did not return composer focus", x.Git.CommitMessage)
			}
			if e := tt.Click("Commit model"); e != nil {
				t.Fatal(e)
			}
			if !tt.HasText("Search commit models") {
				t.Fatal("picker did not reopen")
			}
			// Outside header, not an action and not a guessed native OS target.
			if e := tt.Click("Git"); e != nil {
				t.Fatal(e)
			}
			tt.Frame()
			tt.Type(" outside")
			if x.Git.CommitMessage != "escape draft outside" || tt.HasText("Search commit models") {
				t.Fatal("outside dismissal did not return composer focus", x.Git.CommitMessage)
			}
			g := workspace.Git{Dir: dir}
			if head, e := g.HasCommits(context.Background()); e != nil || head {
				t.Fatal("test created a commit", head, e)
			}
			rows, e := g.Status(context.Background())
			if e != nil || len(rows) != 1 || rows[0].Path != "fixture.txt" || rows[0].Index != '?' {
				t.Fatal("index/status changed", rows, e)
			}
			raw, e := os.ReadFile(filepath.Join(dir, "fixture.txt"))
			if e != nil || string(raw) != "synthetic only\n" {
				t.Fatal("fixture bytes changed", e)
			}
		})
	}
}
