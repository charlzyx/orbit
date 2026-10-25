package orbitui

import (
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// Source sidepane.rs749–769 gives each click a650ms floor. Test historical
// boundaries on an immutable snapshot, never on state a queued expiry cleared.
func TestReviewRefreshFeedbackDeadlineBoundaries(t *testing.T) {
	start := time.Unix(100, 0)
	deadline := start.Add(650 * time.Millisecond)
	if reviewRefreshFeedback != 650*time.Millisecond {
		t.Fatal("source feedback duration changed")
	}
	settled := reviewState{refreshSpinUntil: deadline}
	if !settled.refreshFeedback(start.Add(649*time.Millisecond)) || !settled.refreshFeedback(deadline.Add(-time.Nanosecond)) {
		t.Fatal("feedback ended before650ms")
	}
	if settled.refreshFeedback(deadline) || settled.refreshFeedback(deadline.Add(time.Nanosecond)) {
		t.Fatal("settled feedback continued at/after deadline")
	}
	loading := reviewState{loading: true, refreshSpinUntil: deadline}
	if !loading.refreshFeedback(deadline.Add(time.Second)) {
		t.Fatal("content loading lost feedback")
	}
	expired := reviewState{}
	if expired.refreshFeedback(deadline.Add(-time.Nanosecond)) {
		t.Fatal("cleared live state is not a historical snapshot")
	}
}

func TestReviewRefreshKeepsSource650msFeedback(t *testing.T) {
	a := fixture(t)
	// Keep content loading while exercising actual native button input. No Git
	// completion/expiry callbacks are drained into this controlled mounted state.
	a.reviewState.loading = true
	tt := ui.NewTester(func(c *ui.Context) { c.SetTheme(a.Theme.Native()); a.reviewRefreshButton(c) }, 100, 80)
	tt.Frame()
	click := func() time.Time {
		t.Helper()
		before := time.Now()
		if err := tt.Click("Refresh review"); err != nil {
			t.Fatal(err)
		}
		after := time.Now()
		deadline := a.reviewState.refreshSpinUntil
		if deadline.Before(before.Add(650*time.Millisecond)) || deadline.After(after.Add(650*time.Millisecond)) {
			t.Fatal("native click did not set source650ms deadline", before, deadline, after)
		}
		if _, ok := tt.Find("Refreshing review"); !ok {
			t.Fatal("native feedback glyph missing")
		}
		return deadline
	}
	first := click()
	second := click()
	if !second.After(first) {
		t.Fatal("second native click did not extend deadline while loading")
	}
	// A snapshot with content completed retains the exact captured second-click
	// floor; asynchronous cleanup of live state cannot invalidate this boundary.
	settled := reviewState{refreshSpinUntil: second}
	if !settled.refreshFeedback(second.Add(-time.Nanosecond)) {
		t.Fatal("extended feedback ended early")
	}
	if settled.refreshFeedback(second) {
		t.Fatal("extended feedback failed to end")
	}
}

// Preserve real filesystem refresh completion coverage separately from the
// wall-clock deadline contract; loading may legitimately outlast its650ms floor.
func TestReviewRefreshNativeClickCompletesRealReload(t *testing.T) {
	root := regressionRepo(t)
	a := fixture(t)
	a.Model.Workspace = root
	a.ReviewVisible = true
	a.reviewState.source = workspace.ReviewUncommitted
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	for i := 0; i < 2; i++ {
		before := a.reviewState.generation
		if err := tt.Click("Refresh review"); err != nil {
			t.Fatal(err)
		}
		drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
		if a.reviewState.generation <= before || a.reviewState.err != "" {
			t.Fatal("native refresh did not complete", a.reviewState.generation, a.reviewState.err)
		}
	}
}
