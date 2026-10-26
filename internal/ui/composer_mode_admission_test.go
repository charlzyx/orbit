package orbitui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func modeController(t *testing.T) (*Controller, *ui.Tester) {
	t.Helper()
	x := newConfigApplyFixture(t)
	if err := x.Agent.SetAccessMode("supervised"); err != nil {
		t.Fatal(err)
	}
	x.sync()
	a := x.UI
	a.Prefs.Language = "en"
	a.reloadLocale()
	a.Options.OnAction = x.Action
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	x.sync()
	tt.Frame()
	return x, tt
}
func TestComposerModeRejectedChoiceKeepsVisibleMode(t *testing.T) {
	for _, kind := range []string{"access", "workflow"} {
		t.Run(kind, func(t *testing.T) {
			x, tt := modeController(t)
			a := x.UI
			release, err := x.beginRuntimeOperation()
			if err != nil {
				t.Fatal(err)
			}
			defer release()
			trigger, choice := "Supervised", "Auto-accept edits"
			if kind == "workflow" {
				trigger, choice = "Build", "Ask"
			}
			modeClick(t, tt, trigger)
			modeClick(t, tt, choice)
			if a.access != "supervised" || a.workflow != "build" {
				t.Fatal("rejected mode projected")
			}
			if _, ok := tt.Find(trigger); !ok {
				t.Fatal("old chip lost")
			}
		})
	}
}
func TestComposerModeAccessMaintenanceRejectLeavesExactFile(t *testing.T) {
	x, tt := modeController(t)
	p := filepath.Join(x.Store.Home, ".orbit-pi", "access.json")
	before, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	release, err := x.beginRuntimeOperation()
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	modeClick(t, tt, "Supervised")
	modeClick(t, tt, "Full access")
	after, err := os.ReadFile(p)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("refused choice modified policy", err)
	}
}
func TestComposerModeAccessWriteFailureKeepsVisibleMode(t *testing.T) {
	x, tt := modeController(t)
	p := filepath.Join(x.Store.Home, ".orbit-pi", "access.json")
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0700); err != nil {
		t.Fatal(err)
	}
	modeClick(t, tt, "Supervised")
	modeClick(t, tt, "Full access")
	if x.UI.access != "supervised" || x.Agent.Snapshot().AccessMode != "supervised" {
		t.Fatal("failed atomic write advertised success")
	}
	if len(x.UI.toastStack.Items()) == 0 {
		t.Fatal("failure not reported")
	}
}
func TestComposerModeAcceptedChoiceProjectsConfirmedSnapshot(t *testing.T) {
	x, tt := modeController(t)
	a := x.UI
	modeClick(t, tt, "Supervised")
	modeClick(t, tt, "Auto-accept edits")
	if a.access != "supervised" || x.Store.Access() != "auto-accept-edits" {
		t.Fatal("unconfirmed projection or no actual persistence", a.access, x.Store.Access())
	}
	x.sync()
	tt.Frame()
	if a.access != "auto-accept-edits" || !tt.HasText("Auto-accept edits") {
		t.Fatal("confirmed policy not projected")
	}
}
func TestComposerModeAccessGlobalSuccessAndInvalidBackend(t *testing.T) {
	x, tt := modeController(t)
	original := x.Agent.Snapshot()
	if err := x.Agent.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Connected && s.RuntimeID != original.RuntimeID })
	tt.Frame()
	modeClick(t, tt, "Supervised")
	modeClick(t, tt, "Full access")
	if x.Agent.Snapshot().AccessMode != "full-access" {
		t.Fatal("active not updated")
	}
	if err := x.Agent.Open(original.ActiveSession.Workspace, original.ActiveSession.Path); err != nil {
		t.Fatal(err)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.RuntimeID == original.RuntimeID })
	if x.Agent.Snapshot().AccessMode != "full-access" {
		t.Fatal("parked policy not updated")
	}
	before, err := os.ReadFile(filepath.Join(x.Store.Home, ".orbit-pi", "access.json"))
	if err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "SetAccess", Value: "bypass"})
	after, _ := os.ReadFile(filepath.Join(x.Store.Home, ".orbit-pi", "access.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("invalid mode changed disk")
	}
}
func TestComposerModeWorkflowActualControllerPersistsExactTask(t *testing.T) {
	x, tt := modeController(t)
	id := x.Agent.Snapshot().ActiveSession.ID
	modeClick(t, tt, "Build")
	modeClick(t, tt, "Ask")
	if x.UI.workflow != "build" {
		t.Fatal("optimistic workflow projection")
	}
	drainRegressionUpdates(t, x.UI, func() bool { return x.Store.Workflow(id) == "ask" })
	x.sync()
	tt.Frame()
	if x.UI.workflow != "ask" || !tt.HasText("Ask") {
		t.Fatal("confirmed workflow not projected")
	}
}
func TestComposerModeWorkflowControllerUsesOriginalTargetAfterDispatchDelay(t *testing.T) {
	x, tt := modeController(t)
	a := x.UI
	var delayed Action
	a.Options.OnAction = func(v Action) { delayed = v }
	modeClick(t, tt, "Build")
	modeClick(t, tt, "Ask")
	if delayed.workflowTarget == nil {
		t.Fatal("menu did not capture exact task")
	}
	if err := x.Agent.NewSession(x.workspace); err != nil {
		t.Fatal(err)
	}
	a.Options.OnAction = x.Action
	x.Action(delayed)
	drainRegressionUpdates(t, a, func() bool { return len(a.toastStack.Items()) > 0 })
	if x.Agent.Snapshot().WorkflowMode == "ask" || a.workflow == "ask" {
		t.Fatal("stale action changed or projected new task")
	}
}
func TestComposerModeManagerReplacementDuringHeldChoice(t *testing.T) {
	x, tt := modeController(t)
	a := x.UI
	modeClick(t, tt, "Build")
	r, ok := tt.Find("Ask")
	if !ok {
		t.Fatal("no Ask")
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	if err := x.Agent.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	tt.Frame()
	if a.composerMode != nil || x.Agent.Snapshot().WorkflowMode == "ask" {
		t.Fatal("old held choice changed replaced entry")
	}
}

func TestComposerModePendingIdentitySuccessProjectsSnapshot(t *testing.T) {
	x := maintenanceFixture(t)
	x.Agent.Close()
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	defer release()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "pi")
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	script := "#!/bin/sh\nexport ORBIT_MODE_PENDING_HELPER=1\nexec " + quote(exe) + " -test.run=^TestComposerModePendingRPCProcess$ -- \"$@\"\n"
	if err = os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{Home: x.Store.Home, PiBinary: bin, Offline: true, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, ResolveChildEnv: func(string) ([]string, error) { <-gate; return nil, nil }})
	defer func() { release(); x.Agent.Close() }()
	x.workspace = t.TempDir()
	if err = x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	a := x.UI
	a.Options.OnAction = x.Action
	a.Options.CaptureWorkflowTarget = x.Agent.CaptureWorkflowTargetFor
	a.SetSnapshot(adaptSnapshot(x.Agent.Snapshot(), x.workspace, nil, nil))
	a.Prefs.Language = "en"
	a.reloadLocale()
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if a.Model.SessionID != "" {
		t.Fatal("fixture not pending")
	}
	modeClick(t, tt, "Build")
	modeClick(t, tt, "Plan")
	drainRegressionUpdates(t, a, func() bool { return x.Agent.Snapshot().WorkflowMode == "plan" })
	if a.workflow != "build" {
		t.Fatal("unconfirmed pending projection")
	}
	a.SetSnapshot(adaptSnapshot(x.Agent.Snapshot(), x.workspace, nil, nil))
	tt.Frame()
	if a.workflow != "plan" || !tt.HasText("Plan") {
		t.Fatal("accepted pending mode not projected")
	}
	release()
	drainRegressionUpdates(t, a, func() bool {
		x.Agent.Tick()
		return x.Agent.Snapshot().ActiveSession.ID == "hydrated-mode" && x.Store.Workflow("hydrated-mode") == "plan"
	})
}
func TestComposerModePendingRPCProcess(t *testing.T) {
	if os.Getenv("ORBIT_MODE_PENDING_HELPER") != "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var c map[string]any
		if json.Unmarshal(scanner.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		r := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		if kind == "get_state" {
			r["data"] = map[string]any{"sessionId": "hydrated-mode", "isStreaming": false}
		}
		if kind == "get_messages" {
			r["data"] = map[string]any{"messages": []any{}}
		}
		if kind == "get_available_models" {
			r["data"] = map[string]any{"models": []any{}}
		}
		b, _ := json.Marshal(r)
		fmt.Println(string(b))
	}
	os.Exit(0)
}

func TestComposerModeWorkflowWriteFailurePreservesExistingSnapshotContract(t *testing.T) {
	x, tt := modeController(t)
	p := filepath.Join(x.Store.Home, ".orbit-pi", "workflow.json")
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0700); err != nil {
		t.Fatal(err)
	}
	modeClick(t, tt, "Build")
	modeClick(t, tt, "Ask")
	drainRegressionUpdates(t, x.UI, func() bool { return len(x.UI.toastStack.Items()) > 0 })
	if x.UI.workflow != "build" {
		t.Fatal("UI optimistically changed before snapshot")
	}
	// Existing Manager (and Rust set_workflow_mode) accepts the in-memory mode
	// before persistence. Preserve that separate contract, not a fictitious rollback.
	x.sync()
	tt.Frame()
	if x.UI.workflow != "ask" || x.Agent.Snapshot().WorkflowMode != "ask" || !tt.HasText("Ask") {
		t.Fatal("confirmed in-memory mode not projected")
	}
	st, err := os.Stat(p)
	if err != nil || !st.IsDir() {
		t.Fatal("failed persistence replaced the obstruction", err)
	}
}
