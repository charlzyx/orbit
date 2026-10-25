package orbitui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
)

func piUpdateTargetFixture(t *testing.T) *Controller {
	t.Helper()
	x := workspaceDialogFixture(t)
	_ = x.Agent.Close()
	x.Agent = agent.NewManager(agent.Options{Home: x.Store.Home, UserHome: x.Store.UserHome, PiAgentDir: x.Store.PiDir, DisableExtensions: true, DisablePatches: true})
	t.Setenv("HOME", x.Store.UserHome)
	t.Setenv("USERPROFILE", x.Store.UserHome)
	t.Setenv("PI_CODING_AGENT_DIR", x.Store.PiDir)
	return x
}
func TestPiUpdateTargetMismatchDoesNotStartAutomaticCheck(t *testing.T) {
	for _, which := range []string{"home", "agent-directory"} {
		t.Run(which, func(t *testing.T) {
			x := piUpdateTargetFixture(t)
			if which == "home" {
				t.Setenv("HOME", t.TempDir())
				t.Setenv("USERPROFILE", os.Getenv("HOME"))
			} else {
				t.Setenv("PI_CODING_AGENT_DIR", t.TempDir())
			}
			calls := 0
			x.startPiLaunchUpdate(func(context.Context, native.PiLaunchUpdateOptions) <-chan struct{} { calls++; return nil })
			if calls != 0 || x.maintenance.Load() {
				t.Fatal("automatic check started for a different native target", calls)
			}
		})
	}
}
func TestPiUpdateTargetRecheckedBeforeWriterAdmission(t *testing.T) {
	x := piUpdateTargetFixture(t)
	var options native.PiLaunchUpdateOptions
	calls := 0
	x.startPiLaunchUpdate(func(_ context.Context, o native.PiLaunchUpdateOptions) <-chan struct{} {
		calls++
		options = o
		return nil
	})
	if calls != 1 || options.BeginUpdate == nil || options.OnEvent == nil || options.Busy == nil {
		t.Fatal("matching default target was not wired")
	}
	release, ok := options.BeginUpdate()
	if !ok || release == nil || !x.maintenance.Load() {
		t.Fatal("matching target lost maintenance contract")
	}
	release()
	t.Setenv("PI_CODING_AGENT_DIR", t.TempDir())
	release, ok = options.BeginUpdate()
	if release != nil {
		release()
	}
	if ok || x.maintenance.Load() {
		t.Fatal("changed native target acquired permission to update")
	}
	if !x.sessionGate.TryLock() {
		t.Fatal("rejected target leaked maintenance lease")
	}
	x.sessionGate.Unlock()
}

func TestPiUpdateTargetRetainsClosedCustomAndBusyRejection(t *testing.T) {
	for _, reason := range []string{"closed", "custom-binary", "busy-reader"} {
		t.Run(reason, func(t *testing.T) {
			x := piUpdateTargetFixture(t)
			var options native.PiLaunchUpdateOptions
			x.startPiLaunchUpdate(func(_ context.Context, o native.PiLaunchUpdateOptions) <-chan struct{} { options = o; return nil })
			switch reason {
			case "closed":
				x.Close()
			case "custom-binary":
				exe, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				if err = x.Agent.SetPiBinary(exe); err != nil {
					t.Fatal(err)
				}
			case "busy-reader":
				x.sessionGate.RLock()
				defer x.sessionGate.RUnlock()
			}
			release, ok := options.BeginUpdate()
			if release != nil {
				release()
			}
			if ok || x.maintenance.Load() {
				t.Fatal("unsafe writer admission", reason)
			}
		})
	}
}
func TestPiUpdateTargetDefaultDirectoryAndOwnerDelivery(t *testing.T) {
	x := piUpdateTargetFixture(t)
	t.Setenv("PI_CODING_AGENT_DIR", "")
	// Store's default Pi directory and native default must remain supported.
	if x.Store.PiDir != filepath.Join(x.Store.UserHome, ".pi", "agent") {
		t.Fatal("fixture does not use the default Pi directory")
	}
	calls := 0
	x.startPiLaunchUpdate(func(_ context.Context, o native.PiLaunchUpdateOptions) <-chan struct{} {
		calls++
		o.OnEvent(native.PiUpdateEvent{State: "available", Latest: "0.60.2"})
		if len(x.UI.toastStack.Items()) != 0 {
			t.Fatal("worker callback applied UI off-owner")
		}
		return nil
	})
	if calls != 1 {
		t.Fatal("default native target was disabled")
	}
	x.drainHostUpdates()
	if len(x.UI.toastStack.Items()) != 1 {
		t.Fatal("matching target lost queued event delivery")
	}
	x.Close()
	x.startPiLaunchUpdate(func(context.Context, native.PiLaunchUpdateOptions) <-chan struct{} {
		t.Fatal("closed host started checker")
		return nil
	})
}
