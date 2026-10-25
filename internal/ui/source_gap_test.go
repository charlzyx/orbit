package orbitui

import (
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// This formerly failing acceptance probe now checks the predicate used by the
// native spinner, independently of fast content reload completion.
func TestReviewRefreshKeepsSource650msFeedback(t *testing.T) {
	root := regressionRepo(t)
	a := fixture(t)
	a.Model.Workspace = root
	a.ReviewVisible = true
	a.reviewState.source = workspace.ReviewUncommitted
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	started := time.Now()
	if e := tt.Click("Refresh review"); e != nil {
		t.Fatal(e)
	}
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if !a.reviewState.refreshFeedback(started.Add(649 * time.Millisecond)) {
		t.Fatal("feedback ended before650ms")
	}
	if a.reviewState.refreshFeedback(a.reviewState.refreshSpinUntil) {
		t.Fatal("feedback did not end at deadline after content completed")
	}
	tt.Frame()
	if _, ok := tt.Find("Refreshing review"); !ok {
		t.Fatal("native feedback spinner missing")
	}
	// A second native click extends the deadline even while a reload is active.
	first := a.reviewState.refreshSpinUntil
	time.Sleep(5 * time.Millisecond)
	if err := tt.Click("Refresh review"); err != nil {
		t.Fatal(err)
	}
	if !a.reviewState.refreshSpinUntil.After(first) {
		t.Fatal("second click did not extend feedback")
	}
	deadline := a.reviewState.refreshSpinUntil
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if !a.reviewState.refreshFeedback(deadline.Add(-time.Nanosecond)) {
		t.Fatal("extended feedback ended early")
	}
	if a.reviewState.refreshFeedback(deadline) {
		t.Fatal("extended feedback failed to end")
	}
}
