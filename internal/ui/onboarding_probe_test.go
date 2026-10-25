package orbitui

import (
	"context"
	"errors"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func setupFixtureResult(pi, node, git bool) setupProbeResult {
	return setupProbeResult{dependencies: setupDeps(pi, node, git), facts: []native.HostFact{{Label: "Platform", Value: "fixture OS"}}}
}
func TestOnboardingProbeSingleflightRefreshAndResumeOwnership(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	x.setupProbe.probe = func(ctx context.Context, binary string) (setupProbeResult, error) {
		calls.Add(1)
		close(started)
		select {
		case <-release:
			return setupFixtureResult(true, true, false), nil
		case <-ctx.Done():
			return setupProbeResult{}, ctx.Err()
		}
	}
	resumed := 0
	target := x.Agent.CaptureWorkflowTarget()
	x.setupProbe.resume = func(got agent.WorkflowTarget, binary string) bool {
		if got != target || binary != x.Agent.PiBinary() {
			t.Fatal("wrong resume target")
		}
		resumed++
		return true
	}
	x.refreshSetup()
	<-started
	x.refreshSetup()
	if !x.UI.setup.refreshing || calls.Load() != 1 {
		t.Fatal("probe not singleton")
	}
	close(release)
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	if x.UI.setupBlocked() || resumed != 1 {
		t.Fatal("ready transition failed", resumed)
	}
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		return setupFixtureResult(true, true, false), nil
	}
	x.refreshSetup()
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	if resumed != 1 {
		t.Fatal("repeated ready refresh duplicated startup")
	}
}
func TestOnboardingProbeStaleTaskDoesNotResumeNewIntent(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	started, release := make(chan struct{}), make(chan struct{})
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		close(started)
		<-release
		return setupFixtureResult(true, true, true), nil
	}
	resumed := false
	x.setupProbe.resume = func(agent.WorkflowTarget, string) bool { resumed = true; return true }
	x.refreshSetup()
	<-started
	x.Action(Action{Kind: "NewSession"})
	close(release)
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	if resumed || x.UI.setupBlocked() {
		t.Fatal("global readiness must not retarget old task", resumed)
	}
}
func TestOnboardingProbeErrorThenRetryAndCloseCancellation(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		return setupProbeResult{}, errors.New("fixture probe failed")
	}
	x.refreshSetup()
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	if !x.UI.setupBlocked() || x.UI.setup.error != "fixture probe failed" {
		t.Fatal("probe failure claimed ready")
	}
	started, cancelled := make(chan struct{}), make(chan struct{})
	x.setupProbe.probe = func(ctx context.Context, _ string) (setupProbeResult, error) {
		close(started)
		<-ctx.Done()
		close(cancelled)
		return setupProbeResult{}, ctx.Err()
	}
	x.refreshSetup()
	<-started
	x.Close()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("close leaked read-only probe")
	}
}
func TestOnboardingProbeGenerationRejectsOldCompletion(t *testing.T) {
	x := workspaceDialogFixture(t)
	oldStarted, oldRelease := make(chan struct{}), make(chan struct{})
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		close(oldStarted)
		<-oldRelease
		return setupFixtureResult(true, true, true), nil
	}
	x.refreshSetup()
	<-oldStarted
	x.invalidateSetupProbe()
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		return setupFixtureResult(false, true, true), nil
	}
	x.refreshSetup()
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	close(oldRelease)
	apply := takeControllerDialogUpdate(t, x)
	apply()
	if !x.UI.setupBlocked() || native.MissingRequiredCount(x.UI.setup.dependencies) != 1 {
		t.Fatal("old probe overwrote new missing snapshot")
	}
}

func TestOnboardingReadyDoesNotStartDuringRuntimeMaintenance(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		return setupFixtureResult(true, true, true), nil
	}
	resumed := false
	x.setupProbe.resume = func(agent.WorkflowTarget, string) bool { resumed = true; return true }
	x.sessionGate.Lock()
	x.refreshSetup()
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	x.sessionGate.Unlock()
	if resumed {
		t.Fatal("setup raced runtime maintenance")
	}
}

func TestOnboardingMountedRefreshReturnsChatWithoutSubmitting(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	x.UI.Draft = "preserved draft"
	started, release := make(chan struct{}), make(chan struct{})
	x.setupProbe.probe = func(ctx context.Context, _ string) (setupProbeResult, error) {
		close(started)
		select {
		case <-release:
			return setupFixtureResult(true, true, false), nil
		case <-ctx.Done():
			return setupProbeResult{}, ctx.Err()
		}
	}
	x.setupProbe.resume = func(agent.WorkflowTarget, string) bool { return false }
	tt := ui.NewTester(x.UI.View, 1200, 800)
	if err := tt.Click("Refresh setup"); err != nil {
		t.Fatal(err)
	}
	<-started
	if !x.UI.setup.refreshing {
		t.Fatal("refresh UI not loading")
	}
	close(release)
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	tt.Frame()
	if _, ok := tt.Find("Message"); !ok || x.UI.Draft != "preserved draft" {
		t.Fatal("ready failed to restore original draft")
	}
}

func TestOnboardingActualBinarySwitchInvalidatesProbe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX synthetic executable")
	}
	x := workspaceDialogFixture(t)
	oldStarted, oldRelease := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(oldRelease) }) }
	t.Cleanup(unblock)
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		close(oldStarted)
		<-oldRelease
		return setupFixtureResult(true, true, true), nil
	}
	x.refreshSetup()
	<-oldStarted
	binary := filepath.Join(t.TempDir(), "new-pi")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	x.setupProbe.probe = func(_ context.Context, got string) (setupProbeResult, error) {
		if got != binary {
			return setupProbeResult{}, errors.New("wrong selected executable")
		}
		return setupFixtureResult(false, true, true), nil
	}
	if err := x.Agent.SetPiBinary(binary); err != nil {
		t.Fatal(err)
	}
	x.syncAt(time.Now())
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	unblock()
	apply := takeControllerDialogUpdate(t, x)
	apply()
	if x.setupProbe.binary != binary || !x.UI.setupBlocked() || x.UI.setup.error != "" {
		t.Fatal("binary switch accepted old probe")
	}
}
func TestOnboardingQueuedCompletionIgnoredAfterClose(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		return setupFixtureResult(true, true, true), nil
	}
	x.refreshSetup()
	apply := takeControllerDialogUpdate(t, x)
	x.Close()
	apply()
	if x.UI.setup.loaded {
		t.Fatal("closed host adopted setup completion")
	}
}

func TestOnboardingWindowlessOwnerAdoptsProbeWithoutStartingAnotherTask(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	x.window.Store(nil)
	target := x.Agent.CaptureWorkflowTarget()
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		return setupFixtureResult(true, true, true), nil
	}
	calls := 0
	x.setupProbe.resume = func(got agent.WorkflowTarget, binary string) bool {
		if got != target {
			t.Fatal("windowless result retargeted")
		}
		calls++
		return false
	}
	x.refreshSetup()
	deadline := time.Now().Add(3 * time.Second)
	for x.UI.setup.refreshing && time.Now().Before(deadline) {
		x.tickWindowless()
		time.Sleep(time.Millisecond)
	}
	if x.UI.setup.refreshing || x.UI.setupBlocked() || calls != 1 {
		t.Fatal("retained host stranded setup", calls)
	}
	x.tickWindowless()
	if calls != 1 || x.Agent.CaptureWorkflowTarget() != target {
		t.Fatal("host repeated resume or replaced task")
	}
}
