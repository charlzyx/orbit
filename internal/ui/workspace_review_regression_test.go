package orbitui

import (
	"context"
	"github.com/charlzyx/orbit/internal/workspace"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func drainRegressionUpdates(t *testing.T, a *App, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case f := <-a.updates:
			if f != nil {
				f()
			}
		default:
			time.Sleep(time.Millisecond)
		}
		if done() {
			return
		}
	}
	t.Fatal("background UI operation did not settle")
}
func regressionRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	g := workspace.Git{Dir: root}
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "UI regression"}, {"config", "user.email", "ui@example.invalid"}, {"config", "commit.gpgsign", "false"}, {"config", "core.hooksPath", filepath.Join(root, "no-hooks")}} {
		if _, e := g.Run(context.Background(), args...); e != nil {
			t.Fatal(e)
		}
	}
	os.WriteFile(filepath.Join(root, "seed"), []byte("seed\n"), 0600)
	g.Run(context.Background(), "add", ".")
	if _, e := g.Run(context.Background(), "commit", "-m", "seed"); e != nil {
		t.Fatal(e)
	}
	return root
}
func TestFilesRapidHiddenToggleAndWorkspaceRejectStaleScan(t *testing.T) {
	a := fixture(t)
	root := t.TempDir()
	other := t.TempDir()
	os.WriteFile(filepath.Join(root, ".hidden"), []byte("hidden"), 0600)
	os.WriteFile(filepath.Join(root, "first"), []byte("first"), 0600)
	os.WriteFile(filepath.Join(other, "second"), []byte("second"), 0600)
	a.Model.Workspace = root
	a.RefreshFiles()
	a.files.hidden = true
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == root })
	found := false
	for _, p := range a.files.tree.Files {
		found = found || p == ".hidden"
	}
	if !found {
		t.Fatal("in-flight hidden toggle was dropped", a.files.tree)
	}
	a.files.hidden = false
	a.RefreshFiles()
	a.Model.Workspace = other
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == other })
	if len(a.files.tree.Files) != 1 || a.files.tree.Files[0] != "second" {
		t.Fatal("old workspace scan applied", a.files.tree)
	}
}
func TestReviewRapidSourceAndWorkspaceRejectStaleResult(t *testing.T) {
	a := fixture(t)
	first := regressionRepo(t)
	second := regressionRepo(t)
	os.WriteFile(filepath.Join(first, "first-change"), []byte("first"), 0600)
	os.WriteFile(filepath.Join(second, "second-change"), []byte("second"), 0600)
	(workspace.Git{Dir: second}).Run(context.Background(), "add", "second-change")
	a.Model.Workspace = first
	a.Model.SessionID = "first"
	a.reviewState.source = workspace.ReviewUncommitted
	a.RefreshReview()
	a.Model.Workspace = second
	a.Model.SessionID = "second"
	a.reviewState.source = workspace.ReviewStaged
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading && a.reviewState.root == second })
	if a.reviewState.err != "" {
		t.Fatal(a.reviewState.err)
	}
	s := a.reviewState.data
	if s.Source != workspace.ReviewStaged || len(s.Files) != 1 || s.Files[0].Path != "second-change" {
		t.Fatal("stale review applied", s.Source, s.Files)
	}
}
func TestLastTurnUsesPersistedCounterAfterCompaction(t *testing.T) {
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	ctx := context.Background()
	if e := g.CaptureTurnStart(ctx, "compacted", 3); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, "agent-change"), []byte("third turn\n"), 0600)
	if e := g.CaptureTurn(ctx, "compacted", 3); e != nil {
		t.Fatal(e)
	}
	a := fixture(t)
	a.Model = Snapshot{Workspace: root, SessionID: "compacted", Turn: 3, LatestTurn: 3, Messages: []Message{{Role: "assistant", Text: "Earlier turns were compacted."}}}
	a.reviewState.source = workspace.ReviewLastTurn
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading && a.reviewState.root == root })
	if a.reviewState.err != "" || len(a.reviewState.data.Files) != 1 || a.reviewState.data.Files[0].Path != "agent-change" {
		t.Fatal(a.reviewState.err, a.reviewState.data.Files)
	}
}

// Source app/session.rs keeps latest_turn separate from the in-flight
// turn_count, and app.rs::sync_review_context passes only the completed one.
func TestReviewLastTurnKeepsCompletedDiffDuringNextPrompt(t *testing.T) {
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	ctx := context.Background()
	if e := g.CaptureTurnStart(ctx, "active", 1); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "completed.txt"), []byte("first completed turn\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := g.CaptureTurn(ctx, "active", 1); e != nil {
		t.Fatal(e)
	}
	a := fixture(t)
	a.Model = Snapshot{Workspace: root, SessionID: "active", Turn: 1, LatestTurn: 1}
	a.reviewState.source = workspace.ReviewLastTurn
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if a.reviewState.err != "" || len(a.reviewState.data.Files) != 1 {
		t.Fatal("completed turn did not load", a.reviewState.err, a.reviewState.data.Files)
	}
	if e := g.CaptureTurnStart(ctx, "active", 2); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "in-flight.txt"), []byte("unfinished second turn\n"), 0600); e != nil {
		t.Fatal(e)
	}
	a.Model.Turn = 2
	a.Model.Running = true
	for i := 0; i < 3; i++ {
		a.RefreshReview()
		drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
		if a.reviewState.err != "" || len(a.reviewState.data.Files) != 1 || a.reviewState.data.Files[0].Path != "completed.txt" {
			t.Fatalf("Last Turn targeted unfinished turn instead of retained completed turn: error=%q files=%#v", a.reviewState.err, a.reviewState.data.Files)
		}
	}
}

func TestReviewHistoricalTurnStaysPinnedAndEmptySessionFallsBack(t *testing.T) {
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	ctx := context.Background()
	for turn, name := range []string{"first.txt", "second.txt"} {
		if e := g.CaptureTurnStart(ctx, "history", turn+1); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(root, name), []byte(name+"\n"), 0600); e != nil {
			t.Fatal(e)
		}
		if e := g.CaptureTurn(ctx, "history", turn+1); e != nil {
			t.Fatal(e)
		}
	}
	a := fixture(t)
	a.Model = Snapshot{Workspace: root, SessionID: "history", Turn: 2, LatestTurn: 2}
	a.OpenTurnReview(1)
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	for i := 0; i < 3; i++ {
		if a.currentReviewLabel() != "Turn 1" || a.reviewState.err != "" || len(a.reviewState.data.Files) != 1 || a.reviewState.data.Files[0].Path != "first.txt" {
			t.Fatal("historical turn changed to latest", a.currentReviewLabel(), a.reviewState.err, a.reviewState.data.Files)
		}
		a.RefreshReview()
		drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	}
	a.SetSnapshot(Snapshot{Workspace: root, SessionID: "new-session", Turn: 1, Running: true})
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if a.reviewState.source != workspace.ReviewUncommitted || a.reviewState.turn != 0 || a.reviewState.err != "" {
		t.Fatal("session without completed checkpoint stayed on old source", a.reviewState.source, a.reviewState.turn, a.reviewState.err)
	}
}
