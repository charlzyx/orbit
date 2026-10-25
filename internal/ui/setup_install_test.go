package orbitui

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
)

func setupInstallFixture(t *testing.T) *Controller {
	t.Helper()
	t.Setenv("PI_BIN", "")
	x := workspaceDialogFixture(t)
	x.Agent.Close()
	// A never-opened fixture manager has no executable to launch. All install
	// work below is injected; no test calls the native installer adapter.
	x.Agent = agent.NewManager(agent.Options{Home: t.TempDir()})
	return x
}
func fixtureInstallBackend(before, after native.PiInstallation, run func(context.Context, bool) (native.CommandOutput, error)) setupInstallBackend {
	calls := 0
	return setupInstallBackend{supported: func() bool { return true }, detect: func(context.Context) native.PiInstallation {
		calls++
		if calls == 1 {
			return before
		}
		return after
	}, run: run, dependencies: func(context.Context) []native.Dependency { return setupDeps(true, true, true) }}
}
func managedSetupFixture() native.PiInstallation {
	return native.PiInstallation{Bin: "/fixture/managed/pi", Root: "/fixture/managed", Kind: native.PiManaged, Version: "1.4.0", OnPath: true}
}
func waitSetupInstallMailbox(t *testing.T, x *Controller) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		x.hostMailboxMu.Lock()
		n := len(x.hostMailbox)
		x.hostMailboxMu.Unlock()
		if n > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("installer completion did not reach owner")
}
func TestSetupInstallExitZeroRequiresVerifiedManagedReceipt(t *testing.T) {
	for _, kind := range []string{"missing", "npm", "version-error", "shadowed", "managed"} {
		t.Run(kind, func(t *testing.T) {
			x := setupInstallFixture(t)
			after := managedSetupFixture()
			switch kind {
			case "missing":
				after = native.PiInstallation{}
			case "npm":
				after.Kind = native.PiNPM
			case "version-error":
				after.Version = ""
				after.Error = "cannot execute"
			case "shadowed":
				after.OnPath = false
			}
			var calls atomic.Int32
			b := fixtureInstallBackend(native.PiInstallation{}, after, func(context.Context, bool) (native.CommandOutput, error) {
				calls.Add(1)
				return native.CommandOutput{Stdout: []byte("installer exited zero")}, nil
			})
			if err := x.beginSetupInstall(x.captureSetupInstallIntent(false), b); err != nil {
				t.Fatal(err)
			}
			op := x.setupInstall.pending
			waitSetupInstallMailbox(t, x)
			if !x.maintenance.Load() {
				t.Fatal("maintenance released before verification owner apply")
			}
			if x.sessionGate.TryRLock() {
				x.sessionGate.RUnlock()
				t.Fatal("session mutation admitted before receipt")
			}
			if x.setupInstall.receipt.verified {
				t.Fatal("worker mutated owner state")
			}
			x.drainHostUpdates()
			<-op.done
			r := x.setupInstall.receipt
			want := kind == "shadowed" || kind == "managed"
			if r.verified != want || (r.err == nil) != want || calls.Load() != 1 {
				t.Fatal("false success or missing verified receipt", r, calls.Load())
			}
			if (r.warning != "") != (kind == "shadowed") {
				t.Fatal("PATH shadow warning missing or invented", r.warning)
			}
			if x.maintenance.Load() || x.setupInstall.pending != nil {
				t.Fatal("receipt stranded maintenance")
			}
		})
	}
}
func TestSetupInstallAdmissionAndKindGuardsNeverInvokeInstaller(t *testing.T) {
	for _, kind := range []string{"stale", "custom", "unsupported", "busy", "managed", "other", "npm-unconfirmed"} {
		t.Run(kind, func(t *testing.T) {
			x := setupInstallFixture(t)
			intent := x.captureSetupInstallIntent(false)
			before := native.PiInstallation{}
			switch kind {
			case "stale":
				x.setupProbe.generation++
			case "custom":
				t.Setenv("PI_BIN", "/fixture/custom")
			case "managed":
				before = managedSetupFixture()
			case "other":
				before = native.PiInstallation{Bin: "/fixture/third-party", Kind: native.PiOther}
			case "npm-unconfirmed":
				before = native.PiInstallation{Bin: "/fixture/npm", Kind: native.PiNPM}
			}
			var calls atomic.Int32
			b := fixtureInstallBackend(before, managedSetupFixture(), func(context.Context, bool) (native.CommandOutput, error) {
				calls.Add(1)
				return native.CommandOutput{}, nil
			})
			if kind == "unsupported" {
				b.supported = func() bool { return false }
			}
			if kind == "busy" {
				x.sessionGate.Lock()
				defer x.sessionGate.Unlock()
			}
			err := x.beginSetupInstall(intent, b)
			if kind == "managed" || kind == "other" || kind == "npm-unconfirmed" {
				if err != nil {
					t.Fatal(err)
				}
				waitSetupInstallMailbox(t, x)
				x.drainHostUpdates()
				if x.setupInstall.receipt.err == nil {
					t.Fatal("kind guard accepted install")
				}
			} else if err == nil {
				t.Fatal("invalid admission accepted")
			}
			if calls.Load() != 0 {
				t.Fatal("guard invoked installer")
			}
		})
	}
}
func TestSetupInstallCloseWaitsForActualRunnerReturn(t *testing.T) {
	x := setupInstallFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	b := fixtureInstallBackend(native.PiInstallation{}, managedSetupFixture(), func(ctx context.Context, migrate bool) (native.CommandOutput, error) {
		close(started)
		<-release
		return native.CommandOutput{}, ctx.Err()
	})
	if err := x.beginSetupInstall(x.captureSetupInstallIntent(false), b); err != nil {
		t.Fatal(err)
	}
	op := x.setupInstall.pending
	<-started
	if err := x.beginSetupInstall(x.captureSetupInstallIntent(false), b); err == nil {
		t.Fatal("second install admitted")
	}
	x.Close()
	select {
	case <-op.done:
		t.Fatal("Close falsely settled running installer")
	default:
	}
	if !x.maintenance.Load() {
		t.Fatal("cancellation released active runner")
	}
	unblock()
	select {
	case <-op.done:
	case <-time.After(3 * time.Second):
		t.Fatal("completed runner leaked lease after Close")
	}
	if x.maintenance.Load() || x.setupInstall.receipt.verified {
		t.Fatal("closed host kept lease or applied success")
	}
}
func TestSetupInstallFailureAndStaleOwnerCannotReportSuccess(t *testing.T) {
	for _, kind := range []string{"run-error", "exit-nonzero", "generation", "close-after-queue", "npm-confirmed"} {
		t.Run(kind, func(t *testing.T) {
			x := setupInstallFixture(t)
			before := native.PiInstallation{}
			migrate := kind == "npm-confirmed"
			if migrate {
				before = native.PiInstallation{Bin: "/fixture/npm", Kind: native.PiNPM}
			}
			b := fixtureInstallBackend(before, managedSetupFixture(), func(_ context.Context, got bool) (native.CommandOutput, error) {
				if got != migrate {
					return native.CommandOutput{}, errors.New("wrong confirmation")
				}
				if kind == "exit-nonzero" {
					return native.CommandOutput{ExitCode: 7}, nil
				}
				if kind == "run-error" {
					return native.CommandOutput{Stderr: []byte("fixture failure")}, errors.New("fixture installer failed")
				}
				return native.CommandOutput{}, nil
			})
			if err := x.beginSetupInstall(x.captureSetupInstallIntent(migrate), b); err != nil {
				t.Fatal(err)
			}
			op := x.setupInstall.pending
			waitSetupInstallMailbox(t, x)
			if kind == "generation" {
				x.setupProbe.generation++
			}
			if kind == "close-after-queue" {
				x.Close()
			} else {
				x.drainHostUpdates()
			}
			select {
			case <-op.done:
			case <-time.After(3 * time.Second):
				t.Fatal("settlement leaked")
			}
			if x.setupInstall.receipt.verified != (kind == "npm-confirmed") {
				t.Fatal("invalid success", x.setupInstall.receipt)
			}
			if (kind == "run-error" || kind == "exit-nonzero") && x.setupInstall.receipt.err == nil {
				t.Fatal("installer error hidden")
			}
			if x.maintenance.Load() {
				t.Fatal("maintenance leaked")
			}
		})
	}
}
