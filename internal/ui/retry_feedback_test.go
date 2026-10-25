package orbitui

import (
	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRetryCancellationFooterUsesSourceSlotAndTTL(t *testing.T) {
	a := fixture(t)
	a.Options.OnAction = func(Action) {}
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Connected: true, Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Revision: 1, Attempt: 1}, ActiveSession: agent.SessionRow{ID: "one"}}, "/one", nil, nil))
	tt := ui.NewTester(a.View, 1470, 922)
	a.retryFeedbackNow = tt.Now
	tt.Frame()
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	status, ok := tt.Find("Retry cancellation status")
	if !ok || !tt.HasText(a.Tr("settings.retry_aborted")) {
		t.Fatal("missing source cancellation status", tt.Texts())
	}
	meter, _ := tt.Find("Context usage")
	if status.W > 320 || status.X >= meter.X || status.Y < meter.Y-24 || status.Y > meter.Y+24 {
		t.Fatal("status not in source footer slot", status, meter)
	}
	tt.Advance(6*time.Second - time.Millisecond)
	if !tt.HasText(a.Tr("settings.retry_aborted")) {
		t.Fatal("status expired early")
	}
	tt.Advance(time.Millisecond)
	if tt.HasText(a.Tr("settings.retry_aborted")) {
		t.Fatal("status outlived source six-second TTL")
	}
	if !a.Model.Retrying {
		t.Fatal("status timer changed runtime")
	}
}

func TestRetryExhaustionLocalizationThroughRealRPC(t *testing.T) {
	if err := LoadAssets(); err != nil {
		t.Fatal(err)
	}
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	_ = x.Agent.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestRunStatusRPCChild$", "--"}, Home: x.Store.Home, Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_STATUS_CHILD": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	wait := func(pred func(agent.Snapshot) bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			x.tickWindowless()
			if pred(x.Agent.Snapshot()) {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("feedback RPC timeout", x.Agent.Snapshot())
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !x.Agent.AnyBusy() })
	x.UI.catalog, err = orbitassets.Locale("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(x.View, 1470, 922)
	tt.Frame()
	for i, kind := range []string{"missing", "null", "literal"} {
		if err := x.Agent.Send("status_test_event", map[string]any{"event": map[string]any{"type": "auto_retry_start", "attempt": i + 1}}); err != nil {
			t.Fatal(err)
		}
		wait(func(s agent.Snapshot) bool { return s.RetryDetail.Attempt == uint64(i+1) && s.Retrying })
		event := map[string]any{"type": "auto_retry_end", "success": false}
		if kind == "null" {
			event["finalError"] = nil
		}
		if kind == "literal" {
			event["finalError"] = "pi exhausted its automatic retries"
		}
		if err := x.Agent.Send("status_test_event", map[string]any{"event": event}); err != nil {
			t.Fatal(err)
		}
		wait(func(s agent.Snapshot) bool { return !s.Retrying && s.ConnectionError != "" })
		tt.Frame()
		message := x.UI.Tr("events.retries_exhausted")
		if kind == "literal" {
			message = "pi exhausted its automatic retries"
		}
		want := strings.ReplaceAll(x.UI.Tr("events.auto_retry_failed"), "%{error}", message)
		if !tt.HasText(want) {
			t.Fatal("retry failure not localized/preserved", kind, want, tt.Texts())
		}
		generic := "Automatic retry failed: pi exhausted its automatic retries"
		if err := x.Agent.Send("status_test_event", map[string]any{"event": map[string]any{"type": "response", "command": "unrelated", "success": false, "error": generic}}); err != nil {
			t.Fatal(err)
		}
		wait(func(s agent.Snapshot) bool { return !s.RetryFailure.Active })
		tt.Frame()
		if !tt.HasText(generic) || tt.HasText(want) {
			t.Fatal("same-text generic error was localized as retry", tt.Texts())
		}
	}
}

func TestRetryCancellationFooterLocalizationAndSessionOwnership(t *testing.T) {
	a := fixture(t)
	a.Options.OnAction = func(Action) {}
	var err error
	a.catalog, err = orbitassets.Locale("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	state := Snapshot{Connected: true, SessionID: "one", RuntimeID: "runtime-one", Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Revision: 1, Attempt: 1}}
	a.SetSnapshot(state)
	tt := ui.NewTester(a.View, 1470, 922)
	a.retryFeedbackNow = tt.Now
	tt.Frame()
	a.Dispatch(Action{Kind: "AbortRetry", SessionID: "stale"})
	tt.Frame()
	if tt.HasText(a.Tr("settings.retry_aborted")) {
		t.Fatal("stale action claimed cancellation")
	}
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(a.Tr("settings.retry_aborted")) || tt.HasText("Retry aborted") {
		t.Fatal("footer not localized", tt.Texts())
	}
	state.ExtensionStatusText = "newer extension status"
	tt.Advance(time.Millisecond)
	state.ExtensionStatusAt = tt.Now()
	a.SetSnapshot(state)
	tt.Frame()
	if tt.HasText(a.Tr("settings.retry_aborted")) || !tt.HasText("newer extension status") {
		t.Fatal("newer status did not supersede local footer", tt.Texts())
	}
	state.ExtensionStatusText = ""
	state.ExtensionStatusAt = time.Time{}
	state.RetryDetail.Revision = 2
	a.SetSnapshot(state)
	tt.Frame()
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	state.RuntimeID = "runtime-two"
	a.SetSnapshot(state)
	tt.Frame()
	if tt.HasText(a.Tr("settings.retry_aborted")) {
		t.Fatal("cancel status leaked into restarted runtime")
	}
}

// The tester clock begins at time.Now (upstream headless.go237) and models
// native frameTime. Suppressing App.View models a renderer with no app frames;
// this is a deterministic resume-boundary check, not an OS occlusion test.
func TestRetryCancellationFooterSuspendedFramesResume(t *testing.T) {
	for _, newStatus := range []bool{false, true} {
		t.Run(map[bool]string{false: "expired", true: "new-status"}[newStatus], func(t *testing.T) {
			a := fixture(t)
			a.Options.OnAction = func(Action) {}
			state := Snapshot{Connected: true, SessionID: "one", RuntimeID: "runtime-one", Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Revision: 1, Attempt: 1}}
			a.SetSnapshot(state)
			visible := true
			appFrames := 0
			tt := ui.NewTester(func(c *ui.Context) {
				if visible {
					appFrames++
					a.View(c)
				} else {
					ui.Box(c).Fill()
				}
			}, 1470, 922)
			a.retryFeedbackNow = tt.Now
			tt.Frame()
			if err := tt.Click("Cancel retry"); err != nil {
				t.Fatal(err)
			}
			renderedBeforeHidden := appFrames
			visible = false
			tt.Frame()
			tt.Advance(7 * time.Second)
			if appFrames != renderedBeforeHidden {
				t.Fatal("fixture secretly rendered hidden app")
			}
			// Both controlled instants represent production's local monotonic receive
			// timestamp and frame timestamp; do not compare protocol timestamps here.
			if newStatus {
				state.ExtensionStatusText = "status received while hidden"
				state.ExtensionStatusAt = tt.Now()
			}
			state.RetryDetail.Revision = 2
			a.SetSnapshot(state)
			visible = true
			tt.Frame()
			if tt.HasText(a.Tr("settings.retry_aborted")) || !a.retryStatusAt.IsZero() {
				t.Fatal("resume revived expired cancellation", tt.Texts())
			}
			if newStatus && !tt.HasText("status received while hidden") {
				t.Fatal("newer receive status lost on resume", tt.Texts())
			}
			// A fresh user action after resume must start a fresh six-second lifetime,
			// not reuse the last frame from before the suspended interval.
			if err := tt.Click("Cancel retry"); err != nil {
				t.Fatal(err)
			}
			if !a.retryStatusAt.Equal(tt.Now()) || !tt.HasText(a.Tr("settings.retry_aborted")) {
				t.Fatal("resume cancellation used stale frame clock", a.retryStatusAt, tt.Now(), tt.Texts())
			}
			tt.Advance(6*time.Second - time.Millisecond)
			if !tt.HasText(a.Tr("settings.retry_aborted")) {
				t.Fatal("resumed status expired early")
			}
			tt.Advance(time.Millisecond)
			if tt.HasText(a.Tr("settings.retry_aborted")) {
				t.Fatal("resumed status failed to expire")
			}
		})
	}
}

// A display-paced frame may retain its earlier frame-start timestamp while
// new events arrive. The already-observed extension status is older than this
// user action, even if its receive time is later than the frame's c.Now.
func TestRetryCancellationFooterUsesActionClockNotDelayedFrameClock(t *testing.T) {
	a := fixture(t)
	a.Options.OnAction = func(Action) {}
	state := Snapshot{Connected: true, SessionID: "one", RuntimeID: "runtime-one", Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Revision: 1, Attempt: 1}}
	a.SetSnapshot(state)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	state.ExtensionStatusText = "already observed status"
	state.ExtensionStatusAt = time.Now()
	if !state.ExtensionStatusAt.After(tt.Now()) {
		t.Fatal("fixture did not establish later receive timestamp")
	}
	a.SetSnapshot(state)
	tt.Frame()
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(a.Tr("settings.retry_aborted")) {
		t.Fatal("older observed extension beat newer user cancellation", tt.Texts())
	}
}
