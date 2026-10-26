package agent

import (
	"sync"
	"testing"
)

func captureDisplayedWorkflow(t *testing.T, m *Manager) WorkflowTarget {
	t.Helper()
	s := m.Snapshot()
	v, err := m.CaptureWorkflowTargetFor(s.RuntimeID, s.ActiveSession.ID, s.ActiveSession.Path, s.ActiveSession.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestWorkflowDisplayCaptureAllowsUnhydratedPendingTask(t *testing.T) {
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	defer release()
	m := testManager(t, func(o *Options) { o.ResolveChildEnv = func(string) ([]string, error) { <-gate; return nil, nil } })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); s.ActiveSession.ID != "" || s.Connected {
		t.Fatal("fixture should still be unhydrated")
	}
	target := captureDisplayedWorkflow(t, m)
	if err := m.SetWorkflowModeFor(target, "plan"); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); s.WorkflowMode != "plan" || s.ActiveSession.ID != "" || len(workflowValues(t, m)) != 0 {
		t.Fatal("pending mode not kept only in memory")
	}
	release()
	waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && workflowValues(t, m)[s.ActiveSession.ID] == "plan"
	})
}
func TestWorkflowDisplayCaptureRejectsUnprovableIdentity(t *testing.T) {
	m := testManager(t, nil)
	if _, err := m.CaptureWorkflowTargetFor("", "", "", ""); err == nil {
		t.Fatal("nil entry captured")
	}
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	s := waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	for _, v := range [][4]string{{"other", s.ActiveSession.ID, s.ActiveSession.Path, s.ActiveSession.Workspace}, {s.RuntimeID, "other", s.ActiveSession.Path, s.ActiveSession.Workspace}, {s.RuntimeID, s.ActiveSession.ID, "/different", s.ActiveSession.Workspace}, {s.RuntimeID, s.ActiveSession.ID, s.ActiveSession.Path, "/different"}} {
		if _, err := m.CaptureWorkflowTargetFor(v[0], v[1], v[2], v[3]); err == nil {
			t.Fatal("stale display captured", v)
		}
	}
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if _, err := m.CaptureWorkflowTargetFor(s.RuntimeID, s.ActiveSession.ID, s.ActiveSession.Path, s.ActiveSession.Workspace); err == nil {
		t.Fatal("stopped entry captured")
	}
	m.Close()
	if _, err := m.CaptureWorkflowTargetFor(s.RuntimeID, s.ActiveSession.ID, s.ActiveSession.Path, s.ActiveSession.Workspace); err == nil {
		t.Fatal("closed entry captured")
	}
}
func TestWorkflowDisplayCaptureGenerationAndReplacement(t *testing.T) {
	m := testManager(t, nil)
	workspace := t.TempDir()
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	before := waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	old := captureDisplayedWorkflow(t, m)
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	after := waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.ID != before.ActiveSession.ID
	})
	if after.RuntimeID != before.RuntimeID {
		t.Fatal("fixture did not reuse entry")
	}
	if _, err := m.CaptureWorkflowTargetFor(before.RuntimeID, before.ActiveSession.ID, before.ActiveSession.Path, before.ActiveSession.Workspace); err == nil {
		t.Fatal("old UI captured same-child new session")
	}
	if current := captureDisplayedWorkflow(t, m); current == old {
		t.Fatal("same key lost generation distinction")
	}
	if err := m.SetWorkflowModeFor(old, "ask"); err == nil {
		t.Fatal("old generation changed new task")
	}
	if err := m.SetWorkflowModeFor(captureDisplayedWorkflow(t, m), "unknown"); err == nil {
		t.Fatal("backend unknown mode accepted")
	}
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.CaptureWorkflowTargetFor(after.RuntimeID, after.ActiveSession.ID, after.ActiveSession.Path, after.ActiveSession.Workspace); err == nil {
		t.Fatal("old UI captured replacement entry")
	}
}
