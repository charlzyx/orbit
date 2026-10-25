package orbitui

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
)

func piUpdateReceiptFixture(t *testing.T) *Controller {
	t.Helper()
	x := workspaceDialogFixture(t)
	_ = x.Agent.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestRunStatusRPCChild$", "--"}, Home: x.Store.Home, Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_STATUS_CHILD": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	waitPiUpdateFixture(t, x, func() bool {
		return x.Agent.Snapshot().Connected && x.Agent.Snapshot().ActiveSession.ID != "" && !x.Agent.AnyBusy()
	})
	return x
}
func waitPiUpdateFixture(t *testing.T, x *Controller, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		if ready() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("fixture did not settle: %+v", x.Agent.RuntimeSnapshot())
}
func TestPiUpdateReceiptKeepsLiveIdleProcessAndDraft(t *testing.T) {
	x := piUpdateReceiptFixture(t)
	before := x.Agent.RuntimeSnapshot().Active
	target := x.Agent.CaptureWorkflowTarget()
	x.UI.Draft = "unsent local draft"
	x.UI.attachments = []attachment{{ID: "local-image", Name: "local image"}}
	owner := x.capturePiUpdateOwner()
	x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "updated", Before: "0.60.1", Installation: native.PiInstallation{Kind: native.PiManaged, Version: "0.60.2", OnPath: true}})
	after := x.Agent.RuntimeSnapshot().Active
	if after == nil || after.Key != before.Key || after.PID != before.PID || !after.Alive || x.Agent.CaptureWorkflowTarget() != target {
		t.Fatalf("automatic receipt replaced live idle task: before=%+v after=%+v", before, after)
	}
	if x.UI.Draft != "unsent local draft" || len(x.UI.attachments) != 1 || x.UI.attachments[0].ID != "local-image" {
		t.Fatal("receipt changed local draft")
	}
}

func TestPiUpdateReceiptKeepsBusyAndParkedProcesses(t *testing.T) {
	x := piUpdateReceiptFixture(t)
	owner := x.capturePiUpdateOwner()
	old := x.Agent.RuntimeSnapshot().Active
	if err := x.Agent.Send("status_test_event", map[string]any{"event": map[string]any{"type": "agent_start"}}); err != nil {
		t.Fatal(err)
	}
	waitPiUpdateFixture(t, x, func() bool { return x.Agent.AnyBusy() })
	if err := x.Agent.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitPiUpdateFixture(t, x, func() bool { return x.Agent.Snapshot().Connected && x.Agent.RuntimeSnapshot().Active.PID != old.PID })
	before := x.Agent.RuntimeSnapshot()
	x.UI.Draft = "new active task draft"
	x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "updated", Installation: native.PiInstallation{Version: "0.60.2", OnPath: true}})
	after := x.Agent.RuntimeSnapshot()
	if after.Active.PID != before.Active.PID || after.Active.Key != before.Active.Key || len(after.Parked) != len(before.Parked) || len(after.Parked) != 1 || after.Parked[0].PID != old.PID || !after.Parked[0].Alive || !after.Parked[0].Busy || x.UI.Draft != "new active task draft" {
		t.Fatalf("receipt retargeted active/parked live tasks: before=%+v after=%+v", before, after)
	}
}

func TestPiUpdateReceiptAdoptsVerifiedFactsWithoutResume(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true), facts: []native.HostFact{{LabelKey: "onboarding.pi_version", Value: "old"}, {Label: "Platform", Value: "fixture"}}}
	cancelled, resumed := false, false
	x.setupProbe.cancel = func() { cancelled = true }
	x.setupProbe.resume = func(agent.WorkflowTarget, string) bool { resumed = true; return true }
	owner := x.capturePiUpdateOwner()
	event := native.PiUpdateEvent{State: "updated", Dependencies: setupDeps(true, false, true), Installation: native.PiInstallation{Kind: native.PiManaged, Bin: "/fixture/new-pi", Version: "0.60.2", OnPath: false}}
	x.applyPiUpdateEvent(owner, event)
	if !cancelled || resumed || !x.UI.setupBlocked() || !x.UI.setup.loaded || x.UI.setup.installation != event.Installation || x.setupProbe.generation == owner.generation {
		t.Fatal("receipt did not preserve independent Node readiness or cancel stale probe")
	}
	event.Dependencies[0].Name = "mutated external slice"
	if x.UI.setup.dependencies[0].Name == event.Dependencies[0].Name {
		t.Fatal("retained external dependency slice")
	}
	facts := x.UI.setup.facts
	if len(facts) < 2 || facts[0].Label != "Platform" {
		t.Fatalf("non-Pi host facts lost: %+v", facts)
	}
	for _, f := range facts {
		if f.Value == "old" {
			t.Fatal("stale Pi fact retained")
		}
	}
	items := x.UI.toastStack.Items()
	if len(items) != 2 || items[0].Kind != native.ToastSuccess || items[1].Kind != native.ToastError || items[1].Body != x.UI.Tr("pi_update.installed_not_on_path") {
		t.Fatalf("source receipt/path warning: %+v", items)
	}
	if !strings.Contains(items[0].Body, "0.60.2") || !strings.Contains(items[0].Body, "/fixture/new-pi") {
		t.Fatal("success omitted actual version/path")
	}
}

func TestPiUpdateReceiptRejectsStaleOwnerAndClosedHost(t *testing.T) {
	for _, reason := range []string{"generation", "binary", "manager", "closed", "migration"} {
		t.Run(reason, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			_ = x.Agent.Close()
			x.Agent = agent.NewManager(agent.Options{Home: x.Store.Home, PiBinary: "old-fixture", DisableExtensions: true, DisablePatches: true})
			owner := x.capturePiUpdateOwner()
			switch reason {
			case "generation":
				x.invalidateSetupProbe()
			case "binary":
				exe, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				// No entry exists: selecting this fixture executable cannot spawn it.
				if err := x.Agent.SetPiBinary(exe); err != nil {
					t.Fatal(err)
				}
			case "manager":
				_ = x.Agent.Close()
				x.Agent = agent.NewManager(agent.Options{Home: x.Store.Home, PiBinary: owner.binary, DisableExtensions: true, DisablePatches: true})
			case "closed":
				x.Close()
			case "migration":
				x.setupInstall.pending = &setupInstallOperation{}
				defer func() { x.setupInstall.pending = nil }()
			}
			before := x.UI.setup
			x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "updated", Dependencies: setupDeps(true, true, true), Installation: native.PiInstallation{Version: "0.60.2", OnPath: true}})
			if !reflect.DeepEqual(before, x.UI.setup) || len(x.UI.toastStack.Items()) != 0 {
				t.Fatal("stale owner receipt affected newer host state")
			}
		})
	}
}

func TestPiUpdateReceiptQueuedBeforeNewProbeCannotOverwriteIt(t *testing.T) {
	x := workspaceDialogFixture(t)
	owner := x.capturePiUpdateOwner()
	x.enqueue(func() {
		x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "updated", Dependencies: setupDeps(true, true, true), Installation: native.PiInstallation{Version: "old-receipt", OnPath: true}})
	})
	x.invalidateSetupProbe()
	x.UI.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true), error: "new probe failed"}
	x.drainHostUpdates()
	if x.UI.setup.error != "new probe failed" || !x.UI.setupBlocked() || len(x.UI.toastStack.Items()) != 0 {
		t.Fatal("queued old receipt overwrote new probe state")
	}
}

func TestPiUpdateReceiptFailureAndUnchangedFacts(t *testing.T) {
	for _, state := range []string{"failed", "unchanged"} {
		t.Run(state, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			owner := x.capturePiUpdateOwner()
			x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: state, Detail: "redacted failure", Dependencies: setupDeps(false, true, true), Installation: native.PiInstallation{Kind: native.PiManaged, Bin: "/fixture/pi", Version: "0.60.1"}})
			if x.UI.setup.installation.Version != "0.60.1" || !x.UI.setup.loaded {
				t.Fatal("terminal verification facts were dropped")
			}
			items := x.UI.toastStack.Items()
			if state == "unchanged" && len(items) != 0 {
				t.Fatal("unchanged update announced success")
			}
			if state == "failed" && (len(items) != 1 || items[0].Kind != native.ToastError || items[0].Body != "Pi update failed: redacted failure") {
				t.Fatalf("existing failure handling changed: %+v", items)
			}
		})
	}
}

func TestPiUpdateReceiptNewProcessUsesFixtureLauncherWithoutReplacingOld(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX harmless launcher fixture")
	}
	x := workspaceDialogFixture(t)
	_ = x.Agent.Close()
	dir := t.TempDir()
	launcher, log := filepath.Join(dir, "pi-fixture"), filepath.Join(dir, "launches")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	quote := func(v string) string { return "'" + strings.ReplaceAll(v, "'", "'\\''") + "'" }
	writeLauncher := func(version string) {
		t.Helper()
		script := "#!/bin/sh\nprintf '%s\\n' " + quote(version) + " >> " + quote(log) + "\nexport ORBIT_STATUS_CHILD=1\nexec " + quote(exe) + " -test.run=^TestRunStatusRPCChild$ -- \"$@\"\n"
		if err := os.WriteFile(launcher, []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	writeLauncher("old")
	x.Agent = agent.NewManager(agent.Options{PiBinary: launcher, Home: x.Store.Home, Environment: []string{}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	waitPiUpdateFixture(t, x, func() bool { return x.Agent.Snapshot().Connected && !x.Agent.AnyBusy() })
	old := x.Agent.RuntimeSnapshot().Active
	owner := x.capturePiUpdateOwner()
	// This is a synthetic test launcher rewrite, not an installer invocation.
	writeLauncher("new")
	x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "updated", Dependencies: setupDeps(true, true, true), Installation: native.PiInstallation{Kind: native.PiManaged, Bin: launcher, Version: "0.60.2", OnPath: true}})
	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "old\n" || x.Agent.RuntimeSnapshot().Active.PID != old.PID {
		t.Fatalf("receipt itself spawned a process: %q", raw)
	}
	if err := x.Agent.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitPiUpdateFixture(t, x, func() bool {
		return x.Agent.Snapshot().Connected && x.Agent.RuntimeSnapshot().Active.PID != old.PID && !x.Agent.AnyBusy()
	})
	raw, err = os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	snap := x.Agent.RuntimeSnapshot()
	if string(raw) != "old\nnew\n" || len(snap.Parked) != 1 || snap.Parked[0].PID != old.PID || !snap.Parked[0].Alive {
		t.Fatalf("new process did not use current fixture launcher while old stayed live: %q %+v", raw, snap)
	}
	for _, action := range []string{"RestartRuntime", "RuntimeChanged"} {
		before := x.Agent.RuntimeSnapshot().Active.PID
		x.Action(Action{Kind: action})
		waitPiUpdateFixture(t, x, func() bool {
			return x.Agent.Snapshot().Connected && x.Agent.RuntimeSnapshot().Active.PID != before && !x.Agent.AnyBusy()
		})
	}
}
