package orbitui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
)

// Exercise the real Controller callback, asynchronous Manager restart and saved
// session bootstrap using the existing credential-free JSONL helper process.
func newConfigApplyFixture(t *testing.T) *Controller {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture; separate cross-build covers Windows")
	}
	home, workspace := t.TempDir(), t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	bin := filepath.Join(t.TempDir(), "fixture pi")
	script := "#!/bin/sh\nif [ \"$1\" = mcp ]; then printf '{\"servers\":[]}\\n'; exit 0; fi\nexport ORBIT_MCP_RELOAD_HELPER=1\nexec " + quote(exe) + " -test.run=^TestMCPReloadRPCProcess$ -- \"$@\"\n"
	if err = os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "saved.jsonl")
	if err = os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"saved-fixture","cwd":"`+workspace+`"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	x, err := NewController(RunOptions{Home: home, Workspace: workspace, Session: path, PiBinary: bin, Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID == "saved-fixture" })
	return x
}
func waitConfigApply(t *testing.T, x *Controller, done func(agent.Snapshot) bool) {
	t.Helper()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		x.sync()
		if done(x.Agent.Snapshot()) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("configuration apply did not settle", x.Agent.Snapshot().Status, x.MCP.Status)
}
func TestMCPHostAcknowledgesOnlyCompletedSessionPreservingReload(t *testing.T) {
	x := newConfigApplyFixture(t)
	before := x.Agent.Snapshot()
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.RuntimeID != before.RuntimeID && x.MCP.Status == "MCP changes applied to Pi"
	})
	if x.MCP.Status != "MCP changes applied to Pi" {
		t.Fatal("host never acknowledged actual reload", x.MCP.Status)
	}
	after := x.Agent.Snapshot()
	if after.ActiveSession.ID != before.ActiveSession.ID || after.ActiveSession.Path != before.ActiveSession.Path {
		t.Fatal("reload lost session identity")
	}
}

func TestMCPReloadRPCProcess(t *testing.T) {
	if os.Getenv("ORBIT_MCP_RELOAD_HELPER") != "1" {
		return
	}
	home := os.Getenv("HOME")
	path := ""
	running := false
	write := func(v any) { b, _ := json.Marshal(v); fmt.Fprintln(os.Stdout, string(b)) }
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		r := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		switch kind {
		case "switch_session":
			path, _ = c["sessionPath"].(string)
			r["data"] = map[string]any{"cancelled": false}
		case "get_state":
			if _, e := os.Stat(filepath.Join(home, "hold-reload")); e == nil {
				_ = os.WriteFile(filepath.Join(home, "reload-entered"), []byte("ready"), 0600)
				for {
					if _, e = os.Stat(filepath.Join(home, "hold-reload")); os.IsNotExist(e) {
						break
					}
					time.Sleep(time.Millisecond)
				}
			}
			if _, e := os.Stat(filepath.Join(home, "fail-reload")); e == nil {
				r["success"] = false
				r["error"] = "synthetic bootstrap failure"
				break
			}
			var row struct{ ID string }
			if raw, e := os.ReadFile(path); e == nil {
				_ = json.Unmarshal([]byte(strings.Split(string(raw), "\n")[0]), &row)
			}
			r["data"] = map[string]any{"sessionId": row.ID, "sessionFile": path, "isStreaming": running, "model": map[string]any{"id": "fixture", "provider": "fixture"}}
		case "get_messages":
			r["data"] = map[string]any{"messages": []any{}}
		case "get_available_models":
			r["data"] = map[string]any{"models": []any{}}
		case "get_available_thinking_levels":
			r["data"] = map[string]any{"levels": []string{"off"}}
		case "get_commands":
			r["data"] = map[string]any{"commands": []any{}}
		case "get_entries":
			r["data"] = map[string]any{"entries": []any{}}
		case "auth.list", "quota.list":
			r["success"] = false
			r["error"] = "Unknown command"
		case "prompt":
			running = true
			write(r)
			write(map[string]any{"type": "agent_start"})
			continue
		case "abort":
			running = false
			write(r)
			write(map[string]any{"type": "agent_settled", "aborted": true})
			continue
		}
		write(r)
	}
	os.Exit(0)
}
func configControl(t *testing.T, x *Controller, name string) string {
	t.Helper()
	p := filepath.Join(x.Store.Home, name)
	if e := os.WriteFile(p, []byte("synthetic"), 0600); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestMCPReloadWaitsForBootstrapAndNewEditSupersedesOld(t *testing.T) {
	x := newConfigApplyFixture(t)
	before := x.Agent.Snapshot().RuntimeID
	hold := configControl(t, x, "hold-reload")
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		_, e := os.Stat(filepath.Join(x.Store.Home, "reload-entered"))
		return e == nil
	})
	first := x.Agent.Snapshot().RuntimeID
	if first == before || x.MCP.Status == "MCP changes applied to Pi" || x.Agent.Snapshot().Connected {
		t.Fatal("async admission was acknowledged")
	}
	x.MCP.Options.OnModified() // A newer edit must have its own completed reload.
	x.tickConfigApply(time.Now().Add(time.Second))
	if x.MCP.Status == "MCP changes applied to Pi" {
		t.Fatal("newer edit cleared by pending reload")
	}
	if e := os.Remove(hold); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.RuntimeID != first && x.MCP.Status == "MCP changes applied to Pi"
	})
}
func TestMCPReloadFailureRemainsPendingAndExplicitRetrySucceeds(t *testing.T) {
	x := newConfigApplyFixture(t)
	fail := configControl(t, x, "fail-reload")
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.ConnectionError != "" && strings.Contains(x.MCP.Status, "failed")
	})
	failed := x.Agent.Snapshot().RuntimeID
	for i := 0; i < 5; i++ {
		x.tickConfigApply(time.Now().Add(time.Hour))
	}
	if x.Agent.Snapshot().RuntimeID != failed {
		t.Fatal("failed reload entered automatic restart loop")
	}
	if e := os.Remove(fail); e != nil {
		t.Fatal(e)
	}
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.RuntimeID != failed && x.MCP.Status == "MCP changes applied to Pi"
	})
}
func TestMCPReloadDefersBusyActiveButNotBusyParked(t *testing.T) {
	x := newConfigApplyFixture(t)
	before := x.Agent.Snapshot()
	if e := x.Agent.Submit("synthetic hold", nil, agent.SendPrompt); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Running })
	x.MCP.Options.OnApply()
	x.tickConfigApply(time.Now().Add(time.Hour))
	if x.Agent.Snapshot().RuntimeID != before.RuntimeID {
		t.Fatal("busy active runtime was replaced")
	}
	if e := x.Agent.Abort(); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.RuntimeID != before.RuntimeID && x.MCP.Status == "MCP changes applied to Pi"
	})
	// The old busy runtime must remain running when an idle different session applies.
	if e := x.Agent.Submit("synthetic parked hold", nil, agent.SendPrompt); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Running })
	parked := x.Agent.RuntimeSnapshot().Active
	path := filepath.Join(x.Store.Home, "second.jsonl")
	if e := os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"second-fixture","cwd":"`+x.workspace+`"}`+"\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := x.Agent.Open(x.workspace, path); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID == "second-fixture" })
	second := x.Agent.Snapshot().RuntimeID
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Connected && s.RuntimeID != second })
	found := false
	for _, p := range x.Agent.RuntimeSnapshot().Parked {
		if p.Key == parked.Key {
			found = p.Busy && p.Alive && p.PID == parked.PID
		}
	}
	if !found {
		t.Fatal("busy parked process was stopped or replaced")
	}
}
func TestMCPReloadSessionSwitchCannotAcknowledgeOldCompletion(t *testing.T) {
	x := newConfigApplyFixture(t)
	hold := configControl(t, x, "hold-reload")
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		_, e := os.Stat(filepath.Join(x.Store.Home, "reload-entered"))
		return e == nil
	})
	path := filepath.Join(x.Store.Home, "other.jsonl")
	if e := os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"other-fixture","cwd":"`+x.workspace+`"}`+"\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := x.Agent.Open(x.workspace, path); e != nil {
		t.Fatal(e)
	}
	if e := os.Remove(hold); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID == "other-fixture" })
	if x.MCP.Status == "MCP changes applied to Pi" || x.configApplyFlight != nil {
		t.Fatal("old completion acknowledged different session")
	}
}
func TestGenericRuntimeReloadDoesNotAcknowledgeMCPPending(t *testing.T) {
	x := newConfigApplyFixture(t)
	before := x.Agent.Snapshot().RuntimeID
	x.scheduleConfigApply(0)
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.Connected && s.RuntimeID != before })
	x.tickConfigApply(time.Now())
	if x.MCP.Status == "MCP changes applied to Pi" {
		t.Fatal("generic resources reload acknowledged MCP request")
	}
}

func TestConfigApplyWithoutLiveSavedIdentityDoesNotManufactureSession(t *testing.T) {
	x := maintenanceFixture(t)
	x.scheduleConfigApply(0)
	x.tickConfigApply(time.Now().Add(time.Hour))
	if x.Agent.Snapshot().RuntimeID != "" || x.configApplyFlight != nil {
		t.Fatal("reload manufactured a new session")
	}
}

func TestMCPReloadNewEditThenBootstrapFailureAllowsExplicitRetry(t *testing.T) {
	x := newConfigApplyFixture(t)
	hold := configControl(t, x, "hold-reload")
	fail := configControl(t, x, "fail-reload")
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		_, e := os.Stat(filepath.Join(x.Store.Home, "reload-entered"))
		return e == nil
	})
	failed := x.Agent.Snapshot().RuntimeID
	x.MCP.Options.OnModified()
	x.tickConfigApply(time.Now().Add(time.Second)) // supersede A while it is still loading
	if e := os.Remove(hold); e != nil {
		t.Fatal(e)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.ConnectionError != "" })
	for i := 0; i < 5; i++ {
		x.tickConfigApply(time.Now().Add(time.Hour))
	}
	if x.Agent.Snapshot().RuntimeID != failed {
		t.Fatal("superseded failure automatically restarted")
	}
	if x.MCP.Status == "MCP changes applied to Pi" {
		t.Fatal("failed A acknowledged newer B")
	}
	if e := os.Remove(fail); e != nil {
		t.Fatal(e)
	}
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.RuntimeID != failed && x.MCP.Status == "MCP changes applied to Pi"
	})
}

func TestMCPReloadIgnoresUnrelatedSavedSessionDiagnostic(t *testing.T) {
	x := newConfigApplyFixture(t)
	if err := x.Agent.CloneSavedSession(filepath.Join(t.TempDir(), "missing.jsonl")); err != nil {
		t.Fatal(err)
	}
	waitConfigApply(t, x, func(s agent.Snapshot) bool { return s.ConnectionError != "" })
	if p := x.Agent.RuntimeSnapshot().Active; p == nil || p.Error != "" {
		t.Fatal("fixture did not isolate manager diagnostic")
	}
	before := x.Agent.Snapshot().RuntimeID
	x.MCP.Options.OnApply()
	waitConfigApply(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.RuntimeID != before && x.MCP.Status == "MCP changes applied to Pi"
	})
}
