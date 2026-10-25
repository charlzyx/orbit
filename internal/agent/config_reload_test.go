package agent

import (
	"errors"
	"testing"
)

func TestConfigReloadAtomicGuardsKeepBusyAndSupersededRuntime(t *testing.T) {
	m := NewManager(Options{Home: t.TempDir(), Offline: true})
	defer m.Close()
	s := &liveSession{state: newState(t.TempDir(), "/synthetic/session.jsonl")}
	s.state.snapshot.ActiveSession.ID = "session"
	m.active = "runtime"
	m.sessions[m.active] = s
	for _, v := range []struct{ runtime, path, id string }{{"older", "/synthetic/session.jsonl", "session"}, {"runtime", "/different.jsonl", "session"}, {"runtime", "/synthetic/session.jsonl", "different"}} {
		if e := m.RestartIdle(v.runtime, v.path, v.id); !errors.Is(e, ErrReloadSuperseded) {
			t.Fatal("changed target not rejected", e)
		}
		if m.sessions["runtime"] != s || m.active != "runtime" {
			t.Fatal("superseded reload detached active state")
		}
	}
	s.state.snapshot.Running = true
	if e := m.RestartIdle("runtime", "/synthetic/session.jsonl", "session"); !errors.Is(e, ErrReloadDeferred) {
		t.Fatal("busy target not deferred", e)
	}
	if m.sessions["runtime"] != s || m.active != "runtime" {
		t.Fatal("busy reload detached active state")
	}
}
