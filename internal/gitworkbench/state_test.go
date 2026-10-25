package gitworkbench

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func testRepo(t *testing.T) workspace.Git {
	t.Helper()
	g := workspace.Git{Dir: t.TempDir()}
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "Native Test"}, {"config", "user.email", "native@example.invalid"}, {"config", "commit.gpgsign", "false"}} {
		if _, e := g.Run(context.Background(), args...); e != nil {
			t.Fatal(e)
		}
	}
	return g
}
func writeTest(t *testing.T, g workspace.Git, name, text string) {
	t.Helper()
	if e := os.WriteFile(filepath.Join(g.Dir, name), []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
}
func awaitIdle(t *testing.T, s *State) {
	t.Helper()
	end := time.Now().Add(5 * time.Second)
	for time.Now().Before(end) {
		s.Tick()
		if !s.Busy {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("state did not finish", s.Activity)
}
func TestRefreshAndChangesAreReal(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "a", "hello")
	var wakes atomic.Int32
	s := New(g.Dir, Options{OnChange: func() { wakes.Add(1) }})
	defer s.Close()
	s.Refresh()
	awaitIdle(t, s)
	if s.Error != nil || len(s.Data.Changes) != 1 || s.Data.Changes[0].Path != "a" {
		t.Fatal(s.Error, s.Data)
	}
	s.RunGit("Stage", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.Stage(ctx, "a") })
	awaitIdle(t, s)
	if s.Error != nil || !s.Data.Changes[0].Staged() {
		t.Fatal(s.Error, s.Data)
	}
	s.RunGit("Commit", func(ctx context.Context, g workspace.Git) (string, error) { return g.Commit(ctx, "native", false) })
	awaitIdle(t, s)
	if s.Error != nil || len(s.Data.Changes) != 0 {
		t.Fatal(s.Error, s.Data)
	}
	s.SetTab(1)
	awaitIdle(t, s)
	if len(s.Data.History) != 1 || s.Data.History[0].Subject != "native" {
		t.Fatal(s.Data)
	}
	if wakes.Load() == 0 {
		t.Fatal("host was not notified")
	}
}
func TestStaleResultsAndWorkspaceReset(t *testing.T) {
	s := New("first", Options{})
	generation := s.generation
	s.Data.Branch = "old"
	s.CommitMessage = "draft"
	s.SetWorkspace("second")
	called := false
	s.results <- result{generation: generation, apply: func(*State) { called = true }, status: "stale", err: errors.New("old error")}
	s.Tick()
	if called || s.Data.Branch != "" || s.Status != "" || s.Error != nil || s.CommitMessage != "" {
		t.Fatal("stale workspace result applied")
	}
}
func TestTabSwitchDoesNotCancelMutation(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "a", "a")
	s := New(g.Dir, Options{})
	defer s.Close()
	started, release := make(chan struct{}), make(chan struct{})
	var count atomic.Int32
	s.RunGit("Commit", func(ctx context.Context, g workspace.Git) (string, error) {
		count.Add(1)
		close(started)
		select {
		case <-release:
			return g.Commit(ctx, "once", true)
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})
	<-started
	s.SetTab(1)
	s.RunGit("Duplicate", func(context.Context, workspace.Git) (string, error) { count.Add(10); return "", nil })
	close(release)
	awaitIdle(t, s)
	if count.Load() != 1 || s.Error != nil || s.Tab != 1 || len(s.Data.History) != 1 {
		t.Fatalf("count %d error %v tab %d history %v", count.Load(), s.Error, s.Tab, s.Data.History)
	}
}
func TestCancellationDoesNotApplyLateResults(t *testing.T) {
	s := New(".", Options{})
	started, stopped := make(chan struct{}), make(chan struct{})
	s.start("wait", func(ctx context.Context, _ workspace.Git, _ workspace.GitHub) (func(*State), string, error) {
		close(started)
		<-ctx.Done()
		close(stopped)
		return func(s *State) { s.Status = "late" }, "late", ctx.Err()
	})
	<-started
	s.Cancel()
	<-stopped
	s.Tick()
	if s.Busy || s.Status == "late" {
		t.Fatal("late result applied")
	}
}
func TestActionFailureSurvivesRefresh(t *testing.T) {
	g := testRepo(t)
	s := New(g.Dir, Options{})
	defer s.Close()
	sentinel := errors.New("action failed")
	s.RunGit("Fail", func(context.Context, workspace.Git) (string, error) { return "", sentinel })
	awaitIdle(t, s)
	if !errors.Is(s.Error, sentinel) {
		t.Fatal(s.Error)
	}
	s.Refresh()
	awaitIdle(t, s)
	if !errors.Is(s.Error, sentinel) {
		t.Fatal("refresh erased action error", s.Error)
	}
}
func TestSuccessfulMutationWithFailedRefreshCannotReplay(t *testing.T) {
	g := testRepo(t)
	s := New(g.Dir, Options{})
	defer s.Close()
	s.RunGit("Action", func(context.Context, workspace.Git) (string, error) {
		return "changed", os.Rename(filepath.Join(g.Dir, ".git"), filepath.Join(g.Dir, "saved-git"))
	})
	awaitIdle(t, s)
	if s.Error == nil || s.Status != "changed" || s.lastAction != nil {
		t.Fatal("successful action can be replayed after refresh failure", s.Error, s.Status)
	}
}
func TestChangeGroupingAndLiteralFilters(t *testing.T) {
	s := New(".", Options{})
	s.Data.Changes = []workspace.StatusRow{{Path: "src/a.go", Index: 'M', Worktree: 'M'}, {Path: "b.txt", Index: '?', Worktree: '?'}}
	items := s.changeItems()
	files := 0
	for _, i := range items {
		if i.header == "" {
			files++
		}
	}
	if files != 3 {
		t.Fatal(items)
	}
	s.collapsed["Staged changes:src"] = true
	items = s.changeItems()
	files = 0
	for _, i := range items {
		if i.header == "" {
			files++
		}
	}
	if files != 2 {
		t.Fatal(items)
	}
	s.ChangesFilter = "B.TXT"
	items = s.changeItems()
	files = 0
	for _, i := range items {
		if i.header == "" {
			files++
			if i.path != "b.txt" {
				t.Fatal(i)
			}
		}
	}
	if files != 1 {
		t.Fatal(items)
	}
}
func TestSixNativeTabsRender(t *testing.T) {
	for tab := 0; tab < 6; tab++ {
		t.Run([]string{"Changes", "History", "Graph", "Stashes", "Issues", "Pull requests"}[tab], func(t *testing.T) {
			s := New(".", Options{})
			s.Tab = tab
			s.initialized = true
			s.refreshed = time.Now()
			s.Data.Branch = "main"
			s.Data.GitHub = workspace.GitHubStatus{Installed: true, Authenticated: true, Supported: true}
			s.Data.Changes = []workspace.StatusRow{{Path: "file.go", Index: ' ', Worktree: 'M'}}
			s.Data.History = []workspace.CommitEntry{{Hash: "abc", Short: "abc", Subject: "Real commit", Author: "dev"}}
			s.Data.Graph = workspace.LayoutGraph([]workspace.GraphCommit{{CommitEntry: s.Data.History[0]}})
			s.Data.Stashes = []workspace.StashEntry{{Index: 0, Message: "Saved work"}}
			s.Data.Issues = []workspace.GitHubIssue{{Number: 1, Title: "Real issue", State: "OPEN"}}
			s.Data.Pulls = []workspace.GitHubPull{{Number: 2, Title: "Real PR", State: "OPEN"}}
			tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1280, 900)
			for _, label := range []string{"Changes", "History", "Graph", "Stashes", "Issues", "Pull requests"} {
				if !tt.HasText(label) {
					t.Fatalf("tab missing %s: %v", label, tt.Texts())
				}
			}
			content := []string{"file.go", "Real commit", "Real commit", "Saved work", "Real issue", "Real PR"}[tab]
			if !tt.HasText(content) {
				t.Fatalf("content missing %s: %v", content, tt.Texts())
			}
		})
	}
}
func TestNativeConfirmationDoesNotExecuteBeforeConfirm(t *testing.T) {
	s := New(".", Options{})
	called := 0
	s.Confirm("Destructive test?", "Only executes after explicit confirmation", func() { called++ })
	tt := ui.NewTester(func(c *ui.Context) { s.confirmView(c) }, 700, 500)
	if called != 0 {
		t.Fatal("confirmation executed immediately")
	}
	if e := tt.Click("Cancel"); e != nil {
		t.Fatal(e)
	}
	if called != 0 || s.confirmOpen {
		t.Fatal("cancel executed action")
	}
	s.Confirm("Destructive test?", "Proceed with this action", func() { called++ })
	tt.Frame()
	if e := tt.Click("Confirm"); e != nil {
		t.Fatal(e)
	}
	if called != 1 || s.confirmOpen {
		t.Fatal(called, s.confirmOpen)
	}
}
func TestGitHubFormValidation(t *testing.T) {
	s := New(".", Options{})
	s.initialized = true
	s.refreshed = time.Now()
	s.openDialog("new-issue")
	tt := ui.NewTester(func(c *ui.Context) { s.dialogView(c) }, 800, 700)
	if !tt.HasText("New issue") {
		t.Fatal(tt.Texts())
	}
	if e := tt.Click("Create"); e != nil && !strings.Contains(e.Error(), "disabled") {
		t.Fatal(e)
	}
	if s.Busy || !s.dialogOpen {
		t.Fatal("empty title submitted")
	}
	got := splitValues(" bug, urgent,bug, , docs ")
	if strings.Join(got, ",") != "bug,urgent,docs" {
		t.Fatal(got)
	}
}
