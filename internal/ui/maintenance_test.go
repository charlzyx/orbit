package orbitui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
)

func maintenanceFixture(t *testing.T) *Controller {
	t.Helper()
	a := fixture(t)
	m := agent.NewManager(agent.Options{Home: a.Options.Store.Home, SessionDir: filepath.Join(a.Options.Store.PiDir, "sessions"), Offline: true})
	t.Cleanup(func() { m.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &Controller{UI: a, Agent: m, Store: a.Options.Store, background: ctx, cancelBackground: cancel, stop: make(chan struct{}), archived: map[string]bool{}}
}

func TestMaintenanceGateKeepsDraftAndBlocksNewProcess(t *testing.T) {
	x := maintenanceFixture(t)
	release, err := x.beginRuntimeOperation()
	if err != nil {
		t.Fatal(err)
	}
	if !x.maintenance.Load() {
		t.Fatal("maintenance guard missing")
	}
	x.UI.Draft = "keep this"
	x.UI.submitAccepted = true
	x.Action(Action{Kind: "Submit", Value: "keep this"})
	x.Action(Action{Kind: "NewSession"})
	x.OpenSession("not-opened.jsonl")
	x.SetWorkspace("not-opened")
	if x.UI.submitAccepted || x.UI.Draft != "keep this" || x.Agent.AnyBusy() || x.workspace != "" {
		t.Fatal("maintenance allowed a transition or lost the draft")
	}
	if next, e := x.beginRuntimeOperation(); e == nil {
		next()
		t.Fatal("nested runtime operation was admitted")
	}
	release()
	if x.maintenance.Load() {
		t.Fatal("maintenance guard leaked")
	}
	x.sessionGate.RLock()
	if next, e := x.beginRuntimeOperation(); e == nil {
		next()
		t.Fatal("runtime update raced an isolated generation reader")
	}
	x.sessionGate.RUnlock()
}

func TestRetentionRunsOffThreadAndReloadsArchiveSet(t *testing.T) {
	x := maintenanceFixture(t)
	path := filepath.Join(x.Store.PiDir, "sessions", "fixture", "old.jsonl")
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	data := `{"type":"session","version":3,"id":"old","timestamp":"2020-01-01T00:00:00Z","cwd":"/fixture"}
{"type":"message","id":"m1","parentId":null,"timestamp":"2020-01-01T00:00:01Z","message":{"role":"user","content":[{"type":"text","text":"old task"}],"timestamp":1577836801000}}
`
	if e := os.WriteFile(path, []byte(data), 0600); e != nil {
		t.Fatal(e)
	}
	if e := x.Store.SaveRetention(preferences.Retention{Enabled: true, Mode: "archive", Days: 30}); e != nil {
		t.Fatal(e)
	}
	x.runRetention(true)
	if !x.UI.retentionBusy {
		t.Fatal("retention did not start")
	}
	drainRegressionUpdates(t, x.UI, func() bool { return !x.UI.retentionBusy })
	if !x.archived[path] {
		t.Fatal("archive state not reloaded")
	}
	if _, e := os.Stat(path); e != nil {
		t.Fatal("archive deleted session", e)
	}
	if !x.sessionGate.TryRLock() {
		t.Fatal("retention gate leaked")
	}
	x.sessionGate.RUnlock()
}
