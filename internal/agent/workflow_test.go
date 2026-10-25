package agent

import (
	"encoding/json"
	"github.com/charlzyx/orbit/internal/preferences"
	"os"
	"path/filepath"
	"testing"
)

func workflowValues(t *testing.T, m *Manager) map[string]string {
	t.Helper()
	v := map[string]string{}
	b, e := os.ReadFile(filepath.Join(m.options.Home, ".orbit-pi", "workflow.json"))
	if os.IsNotExist(e) {
		return v
	}
	if e != nil || json.Unmarshal(b, &v) != nil {
		t.Fatal(string(b), e)
	}
	return v
}

func TestWorkflowSelectionBeforeFirstIdentityPersistsOnlyNewTask(t *testing.T) {
	gate := make(chan struct{})
	m := testManager(t, func(o *Options) { o.ResolveChildEnv = func(string) ([]string, error) { <-gate; return nil, nil } })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	if err := m.SetWorkflowMode("plan"); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); s.WorkflowMode != "plan" || s.ActiveSession.ID != "" {
		t.Fatal(s.WorkflowMode, s.ActiveSession.ID)
	}
	if got := workflowValues(t, m); len(got) != 0 {
		t.Fatal("pending choice written without identity", got)
	}
	close(gate)
	waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && workflowValues(t, m)[s.ActiveSession.ID] == "plan"
	})
	if got := m.Snapshot().WorkflowMode; got != "plan" {
		t.Fatal(got)
	}
}

func TestWorkflowNewSessionClearsOldIdentityBeforeAwaitingRPC(t *testing.T) {
	m := testManager(t, func(o *Options) { o.EnvOverrides["ORBIT_NEW_DELAY_MS"] = "90" })
	workspace := t.TempDir()
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == "s-1" })
	if err := m.SetWorkflowMode("ask"); err != nil {
		t.Fatal(err)
	}
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); s.ActiveSession.ID != "" || s.WorkflowMode != "build" {
		t.Fatal("new task retained old identity/mode", s.ActiveSession.ID, s.WorkflowMode)
	}
	if err := m.SetWorkflowMode("plan"); err != nil {
		t.Fatal(err)
	}
	if v := workflowValues(t, m); v["s-1"] != "ask" || v[""] != "" {
		t.Fatal("wrote parked/empty identity", v)
	}
	waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID == "s-2" && workflowValues(t, m)["s-2"] == "plan"
	})
	if v := workflowValues(t, m); v["s-1"] != "ask" || v["s-2"] != "plan" {
		t.Fatal(v)
	}
}

func TestWorkflowFailedSpawnRetainsPendingAndNewTaskResetsIt(t *testing.T) {
	m := testManager(t, nil)
	bad := filepath.Join(t.TempDir(), "absent")
	if err := m.Open(bad, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ConnectionError != "" })
	if err := m.SetWorkflowMode("ask"); err != nil {
		t.Fatal(err)
	}
	m.Tick()
	if s := m.Snapshot(); s.WorkflowMode != "ask" {
		t.Fatal(s.WorkflowMode)
	}
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); s.WorkflowMode != "build" {
		t.Fatal("new task inherited failed pending choice", s.WorkflowMode)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if v := workflowValues(t, m); len(v) != 0 {
		t.Fatal(v)
	}
}

func TestWorkflowCancelledNewSessionDoesNotMutatePreviousChoice(t *testing.T) {
	m := testManager(t, func(o *Options) {
		o.EnvOverrides["ORBIT_NEW_DELAY_MS"] = "90"
		o.EnvOverrides["ORBIT_NEW_CANCEL"] = "1"
	})
	workspace := t.TempDir()
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	if err := m.SetWorkflowMode("ask"); err != nil {
		t.Fatal(err)
	}
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	if err := m.SetWorkflowMode("plan"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ActiveSession.ID == "s-1" && s.WorkflowMode == "ask" })
	if v := workflowValues(t, m); len(v) != 1 || v["s-1"] != "ask" {
		t.Fatal(v)
	}
}

func TestWorkflowStaleHydrationCannotOverwriteNewChoice(t *testing.T) {
	m := testManager(t, nil)
	entry := &liveSession{state: newState(t.TempDir(), "")}
	entry.state.snapshot.ActiveSession.ID = "fixture"
	m.mu.Lock()
	m.sessions["fixture"] = entry
	m.active = "fixture"
	load := m.workflowIdentityLocked(entry)
	m.mu.Unlock()
	if err := m.SetWorkflowMode("ask"); err != nil {
		t.Fatal(err)
	}
	load()
	if got := m.Snapshot().WorkflowMode; got != "ask" {
		t.Fatal("late stored mode replaced explicit choice", got)
	}
	path := filepath.Join(m.options.Home, ".orbit-pi", "workflow.json")
	bad := []byte(`{"unknown":`)
	if err := os.WriteFile(path, bad, 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.SetWorkflowMode("plan"); err == nil {
		t.Fatal("corrupt persistence falsely accepted")
	}
	b, _ := os.ReadFile(path)
	if string(b) != string(bad) {
		t.Fatal("corrupt store destroyed")
	}
	if err := m.SetWorkflowMode("invalid"); err == nil {
		t.Fatal("invalid mode")
	}
}

func TestWorkflowSavedSessionLoadsItsOwnModeNotPendingOtherTask(t *testing.T) {
	m := testManager(t, nil)
	if err := preferences.New(m.options.Home).SaveWorkflow("switched", "ask"); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(filepath.Join(t.TempDir(), "absent"), ""); err != nil {
		t.Fatal(err)
	}
	if err := m.SetWorkflowMode("plan"); err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	if err := os.WriteFile(path, []byte(`{"type":"session","id":"switched","cwd":"`+filepath.ToSlash(workspace)+`"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ActiveSession.ID == "switched" && s.WorkflowMode == "ask" })
}

func TestWorkflowCapturedTargetCannotWriteLaterDraft(t *testing.T) {
	m := testManager(t, func(o *Options) { o.EnvOverrides["ORBIT_NEW_DELAY_MS"] = "90" })
	workspace := t.TempDir()
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	target := m.CaptureWorkflowTarget()
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	if err := m.SetWorkflowModeFor(target, "ask"); err == nil {
		t.Fatal("stale draft generation accepted")
	}
	if got := m.Snapshot().WorkflowMode; got != "build" {
		t.Fatal(got)
	}
}

func TestWorkflowRetryPreservesChoiceAfterFailedSpawn(t *testing.T) {
	m := testManager(t, nil)
	workspace := filepath.Join(t.TempDir(), "initially-missing")
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ConnectionError != "" })
	if err := m.SetWorkflowMode("ask"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && workflowValues(t, m)[s.ActiveSession.ID] == "ask"
	})
	if got := m.Snapshot().WorkflowMode; got != "ask" {
		t.Fatal(got)
	}
}
