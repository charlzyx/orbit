package gitworkbench

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func TestCommitGenerationChoiceAndPartialPush(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "a", "new")
	var stagedOnly bool
	s := New(g.Dir, Options{GenerateCommitResult: func(_ context.Context, staged bool) (workspace.GeneratedText, error) {
		stagedOnly = staged
		return workspace.GeneratedText{Text: "feat: add a"}, nil
	}})
	defer s.Close()
	s.Refresh()
	awaitIdle(t, s)
	s.RequestGenerate(false, false)
	if !s.dialogOpen || s.dialog != "stage-generate" {
		t.Fatal("missing stage choice")
	}
	s.dialogOpen = false
	s.generateCommitFlow(true, false, false)
	awaitIdle(t, s)
	if s.Error != nil || !stagedOnly || s.CommitMessage != "feat: add a" || !s.Data.Changes[0].Staged() {
		t.Fatal(s.Error, s.CommitMessage, stagedOnly, s.Data)
	}
	s.Commit(true)
	awaitIdle(t, s)
	if s.Error == nil || s.CommitMessage != "" || !strings.Contains(s.Status, "Committed locally") {
		t.Fatal("partial success not explicit", s.Error, s.Status, s.CommitMessage)
	}
	history, e := g.History(context.Background(), workspace.HistoryOptions{})
	if e != nil || len(history) != 1 {
		t.Fatal(history, e)
	}
	s.Commit(true)
	awaitIdle(t, s)
	history, e = g.History(context.Background(), workspace.HistoryOptions{})
	if e != nil || len(history) != 1 {
		t.Fatal("repeated click committed again", history, e)
	}
}
func TestBlankCommitGeneratesThenCommits(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "a", "new")
	if e := g.StageAll(context.Background()); e != nil {
		t.Fatal(e)
	}
	s := New(g.Dir, Options{GenerateCommitResult: func(context.Context, bool) (workspace.GeneratedText, error) {
		return workspace.GeneratedText{Text: "feat: generated"}, nil
	}})
	defer s.Close()
	s.Refresh()
	awaitIdle(t, s)
	s.Commit(false)
	awaitIdle(t, s)
	if s.Error != nil || s.CommitMessage != "" || len(s.Data.Changes) != 0 {
		t.Fatal(s.Error, s.CommitMessage, s.Data)
	}
	history, e := g.History(context.Background(), workspace.HistoryOptions{})
	if e != nil || len(history) != 1 || history[0].Subject != "feat: generated" {
		t.Fatal(history, e)
	}
}
func TestTemplatePrefillPreservesEdits(t *testing.T) {
	s := New(".", Options{})
	defer s.Close()
	s.openDialog("new-issue")
	s.body = "user text before async template arrival"
	s.Data.IssueTemplates = []workspace.RepositoryTemplate{{Name: "Bug", Body: "## Repro", TitlePrefix: "bug: ", Labels: []string{"bug"}}}
	s.applyLateTemplate()
	if s.body != "user text before async template arrival" {
		t.Fatal("late template overwrote draft")
	}
	s.body = ""
	s.applyLateTemplate()
	if s.body != "## Repro" || s.title != "bug: " || s.labels != "bug" {
		t.Fatal(s.body, s.title, s.labels)
	}
	s.body += "\nUser edits"
	s.applyTemplate(false, -1)
	if !strings.Contains(s.body, "User edits") {
		t.Fatal("clear template discarded edits")
	}
	s.applyTemplate(false, 0)
	s.applyTemplate(false, -1)
	if s.body != "" {
		t.Fatal("untouched template not cleared")
	}
	s.Data.Branches = []string{"feature", "master", "main"}
	s.openDialog("new-pull")
	if s.base != "main" {
		t.Fatal(s.base)
	}
}
func TestCommitBarStateBranches(t *testing.T) {
	s := New(".", Options{})
	defer s.Close()
	if s.commitBarMode() != "up-to-date" {
		t.Fatal(s.commitBarMode())
	}
	s.Data.HasCommits = true
	if s.commitBarMode() != "publish" {
		t.Fatal(s.commitBarMode())
	}
	s.Data.Tracking = workspace.AheadBehind{HasUpstream: true, Ahead: 1, Behind: 1}
	if s.commitBarMode() != "merge" {
		t.Fatal(s.commitBarMode())
	}
	s.Data.Tracking.Behind = 0
	if s.commitBarMode() != "push" {
		t.Fatal(s.commitBarMode())
	}
	s.Data.Tracking.Ahead = 0
	s.Data.Tracking.Behind = 1
	if s.commitBarMode() != "pull" {
		t.Fatal(s.commitBarMode())
	}
	s.Data.Changes = []workspace.StatusRow{{Path: "a", Index: 'M', Worktree: ' '}}
	if s.commitBarMode() != "commit" || !s.CanCommit() {
		t.Fatal(s.commitBarMode())
	}
}
func TestEmbeddedDiffRendersAndExpands(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "a", "old line\n")
	if _, e := g.Commit(context.Background(), "initial", true); e != nil {
		t.Fatal(e)
	}
	writeTest(t, g, "a", "new line\n")
	s := New(g.Dir, Options{})
	defer s.Close()
	s.Refresh()
	awaitIdle(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	s.diffExpanded["false:a"] = true
	tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1280, 900)
	if !tt.HasText("new line") || !tt.HasText("old line") {
		t.Fatal("embedded patch missing", tt.Texts())
	}
	s.DiffSplit = true
	tt.Frame()
	if !tt.HasText("new line") || !tt.HasText("old line") {
		t.Fatal("split patch missing", tt.Texts())
	}
}
func TestAvatarAsyncCacheAndFallback(t *testing.T) {
	var buffer bytes.Buffer
	im := image.NewRGBA(image.Rect(0, 0, 4, 4))
	im.Set(0, 0, color.RGBA{R: 255, A: 255})
	if e := png.Encode(&buffer, im); e != nil {
		t.Fatal(e)
	}
	var calls atomic.Int32
	s := New(".", Options{AvatarLoader: func(context.Context, string) ([]byte, error) { calls.Add(1); return buffer.Bytes(), nil }})
	defer s.Close()
	tt := ui.NewTester(func(c *ui.Context) { s.Tick(); s.avatar(c, "tester", "") }, 200, 100)
	end := time.Now().Add(time.Second)
	for time.Now().Before(end) && len(s.avatars) == 0 {
		time.Sleep(time.Millisecond)
		tt.Frame()
	}
	if len(s.avatars) != 1 || calls.Load() != 1 {
		t.Fatal("avatar did not load/cache", len(s.avatars), calls.Load())
	}
	tt.Frame()
	if calls.Load() != 1 {
		t.Fatal("avatar fetched repeatedly")
	}
}
func TestGitEscapeClosesTransientOnly(t *testing.T) {
	s := New(".", Options{})
	defer s.Close()
	s.IssueDetail = &workspace.GitHubIssue{Number: 1}
	s.openDialog("new-issue")
	if !s.CloseTransient() || s.dialogOpen || s.IssueDetail == nil {
		t.Fatal("escape priority")
	}
	if !s.CloseTransient() || s.IssueDetail != nil {
		t.Fatal("issue detail not closed")
	}
	if s.CloseTransient() {
		t.Fatal("empty transient consumed escape")
	}
}

func TestCommitDetailExpandsInlineAndGithubDetailsReplaceLists(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "a", "initial")
	if _, e := g.Commit(context.Background(), "Inline commit detail", true); e != nil {
		t.Fatal(e)
	}
	s := New(g.Dir, Options{})
	defer s.Close()
	s.Tab = 1
	s.Refresh()
	awaitIdle(t, s)
	s.SelectCommit(s.Data.History[0].Hash)
	awaitIdle(t, s)
	tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1280, 900)
	if !tt.HasText("A a") || !tt.HasText(s.Data.History[0].Hash) {
		t.Fatal("inline detail missing", tt.Texts())
	}
	s.Tab = 4
	s.Data.GitHub = workspace.GitHubStatus{Installed: true, Authenticated: true, Supported: true}
	s.Data.Issues = []workspace.GitHubIssue{{Number: 2, Title: "List-only issue"}}
	s.IssueDetail = &workspace.GitHubIssue{Number: 1, Title: "Selected issue", Body: "Report body"}
	tt.Frame()
	if !tt.HasText("All issues") || !tt.HasText("Selected issue") || tt.HasText("List-only issue") {
		t.Fatal("detail did not replace list", tt.Texts())
	}
	s.Tab = 5
	s.Data.Pulls = []workspace.GitHubPull{{Number: 2, Title: "List-only pull"}}
	s.PullDetail = &workspace.GitHubPull{Number: 1, Title: "Selected pull", Body: "Pull body"}
	tt.Frame()
	if !tt.HasText("All pull requests") || !tt.HasText("Selected pull") || tt.HasText("List-only pull") {
		t.Fatal("pull detail did not replace list", tt.Texts())
	}
}

func TestWorkspaceSwitchClearsPendingActionsAndFilters(t *testing.T) {
	s := New("old", Options{})
	defer s.Close()
	s.HistoryOptions.Path = "old-file"
	s.GraphLimit = 400
	s.GraphAllRefs = true
	s.IssueFilter.Search = "old issue"
	s.PullFilter.Search = "old pull"
	s.openDialog("new-issue")
	s.body = "private old draft"
	calls := 0
	s.Confirm("Old destructive action", "Old repository", func() { calls++ })
	s.SetWorkspace("new")
	if s.dialogOpen || s.confirmOpen || s.confirmAction != nil || s.body != "" || s.HistoryOptions.Path != "" || s.IssueFilter.Search != "" || s.PullFilter.Search != "" || s.GraphLimit != 80 || s.GraphAllRefs {
		t.Fatal("workspace-specific state leaked", s)
	}
	if calls != 0 {
		t.Fatal("workspace change executed confirmation")
	}
}
