package orbitui

import (
	"context"
	"errors"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSetupMigrationCardOnlyForNPMOnVisibleSetup(t *testing.T) {
	for _, kind := range []string{"npm-missing-node", "npm-ready", "managed", "missing-pi", "other"} {
		t.Run(kind, func(t *testing.T) {
			a := fixture(t)
			deps := setupDeps(true, false, true)
			installation := native.PiInstallation{Bin: "/fixture/pi", Kind: native.PiNPM}
			switch kind {
			case "npm-ready":
				deps = setupDeps(true, true, true)
			case "managed":
				installation.Kind = native.PiManaged
			case "missing-pi":
				installation = native.PiInstallation{}
				deps = setupDeps(false, true, true)
			case "other":
				installation.Kind = native.PiOther
			}
			a.setup = onboardingState{enabled: true, loaded: true, dependencies: deps, installation: installation}
			tt := ui.NewTester(a.View, 1200, 800)
			_, shown := tt.Find("Migrate npm Pi installation")
			if shown != (kind == "npm-missing-node") {
				t.Fatal("source npm-only migration entry mismatch", kind, shown)
			}
			if _, shown := tt.Find("Install pi…"); shown {
				t.Fatal("invented missing-Pi installer button")
			}
		})
	}
}

func setupMigrationFixture(t *testing.T, after native.PiInstallation, node bool, run func(context.Context, bool) (native.CommandOutput, error)) (*Controller, *ui.Tester) {
	t.Helper()
	x := setupInstallFixture(t)
	before := native.PiInstallation{Bin: "/fixture/npm/pi", Kind: native.PiNPM, Version: "1.3.0"}
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(true, false, true), installation: before, facts: append([]native.HostFact{{Label: "Platform", Value: "fixture OS"}}, native.PiInstallFacts(before)...)}
	b := fixtureInstallBackend(before, after, run)
	b.dependencies = func(context.Context) []native.Dependency { return setupDeps(true, node, true) }
	x.setupInstall.backend = &b
	return x, ui.NewTester(x.View, 1200, 1100)
}
func clickSetupMigration(t *testing.T, x *Controller, tt *ui.Tester) {
	t.Helper()
	if err := tt.Click("Migrate npm Pi installation"); err != nil {
		t.Fatal(err)
	}
	if !x.UI.confirmOpen {
		t.Fatal("migration did not request explicit confirmation")
	}
}
func TestSetupMigrationConfirmationCancelStaleAndCloseNeverRun(t *testing.T) {
	for _, kind := range []string{"cancel", "generation", "closed", "maintenance"} {
		t.Run(kind, func(t *testing.T) {
			var calls atomic.Int32
			x, tt := setupMigrationFixture(t, managedSetupFixture(), true, func(context.Context, bool) (native.CommandOutput, error) {
				calls.Add(1)
				return native.CommandOutput{}, nil
			})
			clickSetupMigration(t, x, tt)
			if calls.Load() != 0 {
				t.Fatal("opening confirmation invoked installer")
			}
			switch kind {
			case "cancel":
				if err := tt.Click("Cancel"); err != nil {
					t.Fatal(err)
				}
			case "generation":
				x.setupProbe.generation++
				if err := tt.Click("Continue"); err != nil {
					t.Fatal(err)
				}
			case "closed":
				f := x.UI.confirmCommit
				x.Close()
				notices := len(x.UI.toastStack.Items())
				f()
				if len(x.UI.toastStack.Items()) != notices {
					t.Fatal("closed confirmation mutated UI")
				}
			case "maintenance":
				release, err := x.beginRuntimeOperation()
				if err != nil {
					t.Fatal(err)
				}
				if err = tt.Click("Continue"); err != nil {
					release()
					t.Fatal(err)
				}
				release()
			}
			if calls.Load() != 0 || x.setupInstall.pending != nil {
				t.Fatal("canceled/stale/rejected confirmation started installer", kind)
			}
		})
	}
}
func TestSetupMigrationVerifiedReceiptSeparatesNodeReadinessAndPATH(t *testing.T) {
	for _, kind := range []string{"ready", "node-missing", "path-shadow", "verify-failed", "runner-failed"} {
		t.Run(kind, func(t *testing.T) {
			after := managedSetupFixture()
			if kind == "path-shadow" {
				after.OnPath = false
			}
			if kind == "verify-failed" {
				after.Kind = native.PiNPM
			}
			var calls atomic.Int32
			x, tt := setupMigrationFixture(t, after, kind != "node-missing", func(_ context.Context, migrate bool) (native.CommandOutput, error) {
				calls.Add(1)
				if !migrate {
					return native.CommandOutput{}, errors.New("confirmation did not authorize migration")
				}
				if kind == "runner-failed" {
					return native.CommandOutput{}, errors.New("fixture failure")
				}
				return native.CommandOutput{}, nil
			})
			before := x.Agent.CaptureWorkflowTarget()
			resumes := 0
			x.setupProbe.resume = func(_ agent.WorkflowTarget, _ string) bool { resumes++; return true }
			clickSetupMigration(t, x, tt)
			if err := tt.Click("Continue"); err != nil {
				t.Fatal(err)
			}
			drainControllerDialogUpdates(t, x, func() bool { return x.setupInstall.pending == nil })
			tt.Frame()
			if calls.Load() != 1 || x.maintenance.Load() || x.UI.setup.migrating {
				t.Fatal("migration did not settle once", calls.Load())
			}
			if resumes != 0 || x.Agent.CaptureWorkflowTarget() != before {
				t.Fatal("migration implicitly started/replaced a task")
			}
			verified := kind != "verify-failed" && kind != "runner-failed"
			if x.setupInstall.receipt.verified != verified {
				t.Fatal("wrong verification result", kind, x.setupInstall.receipt)
			}
			if x.UI.setupBlocked() != (kind == "node-missing") {
				t.Fatal("managed installation confused with required Node readiness", kind)
			}
			notices := ""
			for _, n := range x.UI.toastStack.Items() {
				notices += n.Body + "\n"
			}
			if strings.Contains(notices, "pi migrated to the official install") != verified {
				t.Fatal("false/missing installation success", kind, notices)
			}
			if strings.Contains(notices, x.UI.Tr("pi_update.installed_not_on_path")) != (kind == "path-shadow") {
				t.Fatal("missing/invented PATH warning", notices)
			}
			if !verified && !strings.Contains(notices, "pi migration failed") {
				t.Fatal("failure not surfaced", notices)
			}
		})
	}
}
func TestSetupMigrationPendingDisablesDuplicateAndRetainsLeaseUntilOwner(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	var calls atomic.Int32
	x, tt := setupMigrationFixture(t, managedSetupFixture(), false, func(ctx context.Context, _ bool) (native.CommandOutput, error) {
		calls.Add(1)
		close(started)
		select {
		case <-release:
			return native.CommandOutput{}, nil
		case <-ctx.Done():
			return native.CommandOutput{}, ctx.Err()
		}
	})
	t.Cleanup(unblock)
	clickSetupMigration(t, x, tt)
	if err := tt.Click("Continue"); err != nil {
		t.Fatal(err)
	}
	<-started
	tt.Frame()
	if !x.UI.setup.migrating || !x.maintenance.Load() {
		t.Fatal("pending migration has no busy state/lease")
	}
	if err := tt.Click("Migrate npm Pi installation"); err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "RequestSetupMigration"})
	if x.UI.confirmOpen || calls.Load() != 1 {
		t.Fatal("pending migration admitted duplicate confirmation/runner")
	}
	unblock()
	drainControllerDialogUpdates(t, x, func() bool { return x.setupInstall.pending == nil })
	tt.Frame()
	if x.UI.setup.migrating || x.maintenance.Load() || !x.UI.setupBlocked() {
		t.Fatal("completion did not clear progress while retaining missing Node setup")
	}
}
func TestSetupMigrationRechecksIntentAfterMaintenanceAdmission(t *testing.T) {
	x := setupInstallFixture(t)
	intent := x.captureSetupInstallIntent(true)
	calls := 0
	b := fixtureInstallBackend(native.PiInstallation{}, managedSetupFixture(), func(context.Context, bool) (native.CommandOutput, error) { calls++; return native.CommandOutput{}, nil })
	b.supported = func() bool { x.setupProbe.generation++; return true }
	if err := x.beginSetupInstall(intent, b); err == nil {
		t.Fatal("stale admission accepted after backend capability check")
	}
	if calls != 0 || x.setupInstall.pending != nil || x.maintenance.Load() {
		t.Fatal("stale admission started worker or retained lease")
	}
}

func TestSetupMigrationActualBinaryOrTaskChangeInvalidatesConfirmation(t *testing.T) {
	for _, kind := range []string{"binary", "task"} {
		t.Run(kind, func(t *testing.T) {
			var calls atomic.Int32
			x, tt := setupMigrationFixture(t, managedSetupFixture(), true, func(context.Context, bool) (native.CommandOutput, error) {
				calls.Add(1)
				return native.CommandOutput{}, nil
			})
			before := x.Agent.CaptureWorkflowTarget()
			binary := x.Agent.PiBinary()
			clickSetupMigration(t, x, tt)
			if kind == "binary" {
				path := filepath.Join(t.TempDir(), "selected-pi")
				if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
					t.Fatal(err)
				}
				// Only LookPath/admission is used: no child exists to restart, and the
				// dependency refresh is injected so this executable is never run.
				x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
					r := setupFixtureResult(true, false, true)
					r.installation = native.PiInstallation{Kind: native.PiNPM, Bin: path}
					return r, nil
				}
				if err := x.Agent.SetPiBinary(path); err != nil {
					t.Fatal(err)
				}
			} else {
				x.Agent.Close()
				x.Agent = agent.NewManager(agent.Options{Home: t.TempDir(), Environment: []string{}, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true, ResolveChildEnv: func(string) ([]string, error) {
					return nil, errors.New("fixture intentionally prevents all child execution")
				}})
				if err := x.Agent.Open(t.TempDir(), ""); err != nil {
					t.Fatal(err)
				}
				if x.Agent.PiBinary() != binary || x.Agent.CaptureWorkflowTarget() == before {
					t.Fatal("fixture did not isolate task-only ownership change")
				}
			}
			if err := tt.Click("Continue"); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 0 || x.setupInstall.pending != nil {
				t.Fatal("old confirmation migrated for a new binary/task")
			}
		})
	}
}
func TestSetupMigrationProbeCarriesInstallationKindToActualCard(t *testing.T) {
	x := setupInstallFixture(t)
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		r := setupFixtureResult(true, false, true)
		r.installation = native.PiInstallation{Bin: "/fixture/npm/pi", Kind: native.PiNPM, Version: "1.3.0"}
		return r, nil
	}
	x.refreshSetup()
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	tt := ui.NewTester(x.View, 1200, 1100)
	if x.UI.setup.installation.Kind != native.PiNPM {
		t.Fatal("probe result lost install ownership")
	}
	if _, ok := tt.Find("Migrate npm Pi installation"); !ok {
		t.Fatal("actual async probe did not enable source npm-only card")
	}
}

func TestSetupMigrationNativeBackendRejectsMismatchedTargetBeforeConfirmation(t *testing.T) {
	x, tt := setupMigrationFixture(t, managedSetupFixture(), true, func(context.Context, bool) (native.CommandOutput, error) {
		t.Fatal("injected runner must not run")
		return native.CommandOutput{}, nil
	})
	x.setupInstall.backend = nil // Exercise native-target admission only, never its runner.
	if err := tt.Click("Migrate npm Pi installation"); err != nil {
		t.Fatal(err)
	}
	if x.UI.confirmOpen || x.setupInstall.pending != nil {
		t.Fatal("isolated home admitted ambient native installer")
	}
	notices := x.UI.toastStack.Items()
	if len(notices) == 0 || !strings.Contains(notices[len(notices)-1].Body, "Pi home differs") {
		t.Fatal("unsupported native target not explained", notices)
	}
}

func TestSetupMigrationCapture(t *testing.T) {
	path := os.Getenv("ORBIT_MIGRATION_CAPTURE")
	if path == "" {
		t.Skip("optional CPU-only npm setup capture")
	}
	a := fixture(t)
	installation := native.PiInstallation{Bin: "/fixture/npm/pi", Kind: native.PiNPM, Version: "1.3.0"}
	deps := setupDeps(true, false, true)
	deps[0].Version = "1.3.0"
	deps[1].InstallHint = native.DependencyInstallHint("darwin", "node")
	deps[2].Version = "2.47.0"
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: deps, installation: installation, facts: append([]native.HostFact{{Label: "Platform", Value: "macOS fixture · aarch64"}, {Label: "pi sessions", Value: "~/.pi/agent/sessions"}, {Label: "Orbit config", Value: "~/.orbit-pi"}}, native.PiInstallFacts(installation)...)}
	tt := ui.NewTester(a.View, 1200, 1100)
	tt.Frame()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

func TestSetupMigrationHeldRunnerRejectsRefreshAndKeepsFinalReceipt(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	x, tt := setupMigrationFixture(t, managedSetupFixture(), false, func(ctx context.Context, _ bool) (native.CommandOutput, error) {
		close(started)
		select {
		case <-release:
			return native.CommandOutput{}, nil
		case <-ctx.Done():
			return native.CommandOutput{}, ctx.Err()
		}
	})
	t.Cleanup(unblock)
	var probes atomic.Int32
	x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
		probes.Add(1)
		r := setupFixtureResult(true, true, true)
		r.installation = managedSetupFixture()
		return r, nil
	}
	clickSetupMigration(t, x, tt)
	if err := tt.Click("Continue"); err != nil {
		t.Fatal(err)
	}
	<-started
	generation := x.setupProbe.generation
	if err := tt.Click("Refresh setup"); err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "RefreshSetup"})
	tt.Frame()
	if x.UI.setup.refreshing || x.setupProbe.generation != generation {
		t.Fatal("pending migration allowed refresh to invalidate its outcome")
	}
	if probes.Load() != 0 || !x.UI.setupBlocked() {
		t.Fatal("pending migration exposed half-installed readiness")
	}
	unblock()
	drainControllerDialogUpdates(t, x, func() bool { return x.setupInstall.pending == nil })
	tt.Frame()
	if !x.setupInstall.receipt.verified || !x.UI.setupBlocked() {
		t.Fatal("final migration outcome or missing Node readiness was lost")
	}
	if err := tt.Click("Refresh setup"); err != nil {
		t.Fatal(err)
	}
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	tt.Frame()
	if probes.Load() != 1 || x.UI.setupBlocked() {
		t.Fatal("settled migration did not restore explicit dependency refresh")
	}
}
func TestSetupMigrationRejectsWhileDependencyProbeIsActive(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	var calls atomic.Int32
	x, tt := setupMigrationFixture(t, managedSetupFixture(), true, func(context.Context, bool) (native.CommandOutput, error) {
		calls.Add(1)
		return native.CommandOutput{}, nil
	})
	t.Cleanup(unblock)
	x.setupProbe.probe = func(ctx context.Context, _ string) (setupProbeResult, error) {
		close(started)
		select {
		case <-release:
			r := setupFixtureResult(true, false, true)
			r.installation = native.PiInstallation{Bin: "/fixture/npm/pi", Kind: native.PiNPM}
			return r, nil
		case <-ctx.Done():
			return setupProbeResult{}, ctx.Err()
		}
	}
	x.refreshSetup()
	<-started
	tt.Frame()
	if err := tt.Click("Migrate npm Pi installation"); err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "RequestSetupMigration"})
	if x.UI.confirmOpen {
		t.Fatal("active dependency probe admitted migration confirmation")
	}
	if err := x.beginSetupInstall(x.captureSetupInstallIntent(true), *x.setupInstall.backend); err == nil {
		t.Fatal("active probe admitted migration worker")
	}
	if calls.Load() != 0 || x.setupInstall.pending != nil {
		t.Fatal("active probe raced installer")
	}
	unblock()
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
	tt.Frame()
	clickSetupMigration(t, x, tt)
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
}

func TestSetupMigrationQueuedReceiptBoundaryRestoresRefreshAfterFailureOrDeadline(t *testing.T) {
	for _, kind := range []string{"verified", "failure", "deadline", "canceled-context"} {
		t.Run(kind, func(t *testing.T) {
			started, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			x, tt := setupMigrationFixture(t, managedSetupFixture(), false, func(ctx context.Context, _ bool) (native.CommandOutput, error) {
				close(started)
				select {
				case <-release:
				case <-ctx.Done():
					return native.CommandOutput{}, ctx.Err()
				}
				switch kind {
				case "failure":
					return native.CommandOutput{}, errors.New("fixture installer failed")
				case "deadline":
					return native.CommandOutput{}, context.DeadlineExceeded
				}
				return native.CommandOutput{}, nil
			})
			t.Cleanup(unblock)
			var probes atomic.Int32
			x.setupProbe.probe = func(context.Context, string) (setupProbeResult, error) {
				probes.Add(1)
				r := setupFixtureResult(true, true, true)
				r.installation = managedSetupFixture()
				return r, nil
			}
			clickSetupMigration(t, x, tt)
			if err := tt.Click("Continue"); err != nil {
				t.Fatal(err)
			}
			<-started
			generation := x.setupProbe.generation
			if kind == "canceled-context" {
				x.setupInstall.pending.cancel()
			} else {
				unblock()
			}
			waitSetupInstallMailbox(t, x)
			// The runner has returned but the final receipt is still queued. Refresh
			// cannot invalidate it or read partial state before owner settlement.
			x.Action(Action{Kind: "RefreshSetup"})
			if x.setupProbe.generation != generation || x.UI.setup.refreshing || probes.Load() != 0 || !x.maintenance.Load() {
				t.Fatal("queued receipt was bypassed by dependency refresh")
			}
			x.drainHostUpdates()
			tt.Frame()
			if x.setupInstall.pending != nil || x.maintenance.Load() || x.UI.setup.migrating {
				t.Fatal("actual receipt did not release progress/lease")
			}
			if x.setupInstall.receipt.verified != (kind == "verified") {
				t.Fatal("queued action outcome lost", kind, x.setupInstall.receipt)
			}
			if kind != "verified" && x.setupInstall.receipt.err == nil {
				t.Fatal("error/deadline outcome silently discarded")
			}
			if err := tt.Click("Refresh setup"); err != nil {
				t.Fatal(err)
			}
			drainControllerDialogUpdates(t, x, func() bool { return !x.UI.setup.refreshing })
			tt.Frame()
			if probes.Load() != 1 || x.UI.setupBlocked() {
				t.Fatal("error/deadline settlement did not restore explicit refresh")
			}
		})
	}
}

func TestSetupMigrationCloseDuringRunDoesNotApplyReceiptToClosedUI(t *testing.T) {
	started := make(chan struct{})
	x, tt := setupMigrationFixture(t, managedSetupFixture(), true, func(ctx context.Context, _ bool) (native.CommandOutput, error) {
		close(started)
		<-ctx.Done()
		return native.CommandOutput{}, ctx.Err()
	})
	clickSetupMigration(t, x, tt)
	if err := tt.Click("Continue"); err != nil {
		t.Fatal(err)
	}
	<-started
	before := len(x.UI.toastStack.Items())
	op := x.setupInstall.pending
	x.Close()
	<-op.done
	if x.setupInstall.receipt.verified || len(x.UI.toastStack.Items()) != before {
		t.Fatal("closed host adopted old installation receipt")
	}
}
