package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/pirpc"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func writeMCPFixture(t *testing.T, root, content string, global bool) {
	t.Helper()
	path := filepath.Join(root, ".pi", "mcp.json")
	if global {
		path = filepath.Join(root, ".pi", "agent", "mcp.json")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
func openMCPFixture(t *testing.T, m *Manager, workspace, path string) *RuntimeProcess {
	t.Helper()
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	return waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active != nil && s.Active.State == "idle" }).Active
}
func TestMCPIdleParkedReopenReplacesStaleConfig(t *testing.T) {
	m := testManager(t, nil)
	a, b := t.TempDir(), t.TempDir()
	path := filepath.Join(a, "saved.jsonl")
	writeMCPFixture(t, a, `{"mcpServers":{"a":{"command":"old"}}}`, false)
	first := openMCPFixture(t, m, a, path)
	openMCPFixture(t, m, b, filepath.Join(b, "saved.jsonl"))
	writeMCPFixture(t, a, `{"mcpServers":{"a":{"command":"new"}}}`, false)
	next := openMCPFixture(t, m, a, path)
	if next.PID == first.PID || next.Key == first.Key {
		t.Fatal("stale idle parked process was reused", next.PID)
	}
}

func TestMCPParkedUsesTargetWorkspaceAndGlobalRawBytes(t *testing.T) {
	m := testManager(t, nil)
	a, b := t.TempDir(), t.TempDir()
	pa, pb := filepath.Join(a, "saved.jsonl"), filepath.Join(b, "saved.jsonl")
	writeMCPFixture(t, a, `{"mcpServers":{"a":{"command":"a"}}}`, false)
	writeMCPFixture(t, b, `{"mcpServers":{"b":{"command":"b"}}}`, false)
	first := openMCPFixture(t, m, a, pa)
	other := openMCPFixture(t, m, b, pb)
	// Changing only the outgoing project's configuration must not invalidate A.
	writeMCPFixture(t, b, `{"mcpServers":{"b":{"command":"changed"}}}`, false)
	same := openMCPFixture(t, m, a, pa)
	if same.PID != first.PID {
		t.Fatal("outgoing workspace fingerprint killed target", first.PID, same.PID)
	}
	changed := openMCPFixture(t, m, b, pb)
	if changed.PID == other.PID {
		t.Fatal("target project change ignored")
	}
	// Both workspaces inherit the global file; even whitespace is in Rust's raw domain.
	writeMCPFixture(t, m.options.Home, "{\"mcpServers\":{}}\n", true)
	global := openMCPFixture(t, m, a, pa)
	if global.PID == same.PID {
		t.Fatal("global change ignored")
	}
	openMCPFixture(t, m, b, pb)
	writeMCPFixture(t, m.options.Home, "{\"mcpServers\":{}}\n\n", true)
	whitespace := openMCPFixture(t, m, a, pa)
	if whitespace.PID == global.PID {
		t.Fatal("raw bytes incorrectly reduced to parsed config")
	}
}

func TestMCPBusyParkedSurvivesUntilIdleReopen(t *testing.T) {
	m := testManager(t, nil)
	a, b := t.TempDir(), t.TempDir()
	pa, pb := filepath.Join(a, "saved.jsonl"), filepath.Join(b, "saved.jsonl")
	first := openMCPFixture(t, m, a, pa)
	if err := m.Submit("synthetic running", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	openMCPFixture(t, m, b, pb)
	writeMCPFixture(t, a, `{"mcpServers":{"changed":{"command":"fixture"}}}`, false)
	if err := m.Open(a, pa); err != nil {
		t.Fatal(err)
	}
	busy := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.Busy }).Active
	if busy.PID != first.PID || !busy.Alive {
		t.Fatal("busy stale process replaced")
	}
	if err := m.Abort(); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	// Settling alone must not eagerly restart the active or parked process.
	if m.RuntimeSnapshot().Active.PID != first.PID {
		t.Fatal("eager restart after settle")
	}
	openMCPFixture(t, m, b, pb)
	next := openMCPFixture(t, m, a, pa)
	if next.PID == first.PID {
		t.Fatal("stale process survived idle reopen")
	}
}

func TestMCPLateValidationCannotStealSelectionOrConsumeNewerEdit(t *testing.T) {
	m := testManager(t, nil)
	a, b := t.TempDir(), t.TempDir()
	pa, pb := filepath.Join(a, "saved.jsonl"), filepath.Join(b, "saved.jsonl")
	first := openMCPFixture(t, m, a, pa)
	other := openMCPFixture(t, m, b, pb)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	var reads atomic.Int32
	m.mcpFingerprint = func(home, workspace string) [32]byte {
		stamp := mcpConfigStamp(home, workspace)
		if workspace == a && reads.Add(1) == 1 {
			close(entered)
			<-release
		}
		return stamp
	}
	if err := m.Open(a, pa); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("fingerprint worker did not start")
	}
	if m.Snapshot().Connected {
		t.Fatal("unvalidated runtime advertised ready")
	}
	if err := m.Submit("must not start on unvalidated configuration", nil, SendPrompt); err == nil {
		t.Fatal("prompt admitted before idle validation")
	}
	// Open and Snapshot returned while worker is held; no filesystem work on caller.
	back := openMCPFixture(t, m, b, pb)
	if back.PID != other.PID {
		t.Fatal("unrelated outgoing runtime replaced")
	}
	writeMCPFixture(t, a, `{"mcpServers":{"newer":{"command":"fixture"}}}`, false)
	fresh := openMCPFixture(t, m, a, pa)
	if fresh.PID == first.PID {
		t.Fatal("newer selection reused stale config")
	}
	once.Do(func() { close(release) })
	m.workers.Wait()
	if got := m.RuntimeSnapshot().Active; got.Key != fresh.Key || got.Loading || !got.Alive {
		t.Fatal("late completion consumed newer selection", got)
	}
}

func TestMCPValidationPreservesTurnThatArrivesWhileChecking(t *testing.T) {
	m := testManager(t, nil)
	a, b := t.TempDir(), t.TempDir()
	pa, pb := filepath.Join(a, "saved.jsonl"), filepath.Join(b, "saved.jsonl")
	first := openMCPFixture(t, m, a, pa)
	m.mu.Lock()
	client := m.sessions[first.Key].client
	m.mu.Unlock()
	openMCPFixture(t, m, b, pb)
	writeMCPFixture(t, a, `{"mcpServers":{"changed":{"command":"fixture"}}}`, false)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	m.mcpFingerprint = func(home, workspace string) [32]byte {
		if workspace == a {
			close(entered)
			<-release
		}
		return mcpConfigStamp(home, workspace)
	}
	if err := m.Open(a, pa); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("check did not start")
	}
	// A queued background event can arrive after admission checked idle. Use the
	// real fixture transport to simulate that already-issued process-side turn.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := client.Call(ctx, pirpc.NewCommand(pirpc.Prompt, map[string]any{"message": "background fixture"})); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	once.Do(func() { close(release) })
	m.workers.Wait()
	got := m.RuntimeSnapshot().Active
	if got.PID != first.PID || !got.Busy || !got.Alive || got.Loading {
		t.Fatal("late busy event was killed", got)
	}
}

func TestMCPParkedFailureProcess(t *testing.T) {
	if os.Getenv("ORBIT_AGENT_HELPER") != "1" {
		return
	}
	if _, err := os.Stat(os.Getenv("ORBIT_MCP_FAIL_MARKER")); err == nil {
		mode := os.Getenv("ORBIT_MCP_FAILURE_MODE")
		if mode == "exit" {
			os.Exit(19)
		}
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			var command map[string]any
			if json.Unmarshal(scanner.Bytes(), &command) != nil {
				continue
			}
			kind, _ := command["type"].(string)
			response := map[string]any{"type": "response", "id": command["id"], "command": kind, "success": true, "data": map[string]any{}}
			if kind == "get_state" {
				response["success"] = false
				response["error"] = "synthetic bootstrap RPC failure"
			}
			if kind == "switch_session" && mode == "cancel" {
				response["data"] = map[string]any{"cancelled": true}
			}
			raw, _ := json.Marshal(response)
			fmt.Fprintln(os.Stdout, string(raw))
		}
		os.Exit(0)
	}
	TestAgentProcess(t)
}
func TestMCPStaleReplacementFailureCanBeReopenedWithoutHurtingParked(t *testing.T) {
	for _, mode := range []string{"exit", "rpc", "cancel"} {
		t.Run(mode, func(t *testing.T) { testMCPStaleReplacementFailure(t, mode) })
	}
}
func testMCPStaleReplacementFailure(t *testing.T, mode string) {
	marker := filepath.Join(t.TempDir(), "fail-next-bootstrap")
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestMCPParkedFailureProcess$", "--"}
		o.EnvOverrides["ORBIT_MCP_FAIL_MARKER"] = marker
		o.EnvOverrides["ORBIT_MCP_FAILURE_MODE"] = mode
	})
	a, b := t.TempDir(), t.TempDir()
	pa, pb := filepath.Join(a, "saved.jsonl"), filepath.Join(b, "saved.jsonl")
	first := openMCPFixture(t, m, a, pa)
	other := openMCPFixture(t, m, b, pb)
	writeMCPFixture(t, a, `{"mcpServers":{"changed":{"command":"fixture"}}}`, false)
	if err := os.WriteFile(marker, []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(a, pa); err != nil {
		t.Fatal(err)
	}
	failed := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active != nil && !s.Active.Loading && s.Active.Error != "" }).Active
	if failed.Key == first.Key {
		t.Fatal("failure did not belong to replacement")
	}
	if mode != "exit" && !failed.Alive {
		t.Fatal("fixture did not preserve failed alive client")
	}
	m.workers.Wait()
	if m.RuntimeSnapshot().Active.Key != failed.Key {
		t.Fatal("failure entered automatic replacement loop")
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	recovered := openMCPFixture(t, m, a, pa)
	if recovered.Key == failed.Key || recovered.PID == first.PID {
		t.Fatal("failed replacement could not recover")
	}
	back := openMCPFixture(t, m, b, pb)
	if back.PID != other.PID {
		t.Fatal("unrelated parked process affected by failure")
	}
}

func TestMCPStampIncludesPresenceRawBytesAndTargetPaths(t *testing.T) {
	home, a, b := t.TempDir(), t.TempDir(), t.TempDir()
	absent := mcpConfigStamp(home, a)
	if absent == mcpConfigStamp(home, b) {
		t.Fatal("workspace paths absent from fingerprint")
	}
	writeMCPFixture(t, a, "", false)
	empty := mcpConfigStamp(home, a)
	if empty == absent {
		t.Fatal("empty file confused with absent")
	}
	writeMCPFixture(t, a, "{malformed", false)
	malformed := mcpConfigStamp(home, a)
	if malformed == empty {
		t.Fatal("invalid config bytes excluded")
	}
	if err := os.Remove(filepath.Join(a, ".pi", "mcp.json")); err != nil {
		t.Fatal(err)
	}
	if mcpConfigStamp(home, a) != absent {
		t.Fatal("deletion did not restore absent stamp")
	}
}

type heldMCPMessage struct {
	entered chan struct{}
	release chan struct{}
}

func (m heldMCPMessage) MarshalJSON() ([]byte, error) {
	close(m.entered)
	<-m.release
	return []byte(`"synthetic delayed message"`), nil
}
func TestMCPReplacementRejectsCommandAdmittedBeforeMarshal(t *testing.T) {
	m := testManager(t, nil)
	a, b := t.TempDir(), t.TempDir()
	pa, pb := filepath.Join(a, "saved.jsonl"), filepath.Join(b, "saved.jsonl")
	first := openMCPFixture(t, m, a, pa)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	result := make(chan error, 1)
	go func() { result <- m.Send(pirpc.Prompt, map[string]any{"message": heldMCPMessage{entered, release}}) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("marshal did not pause")
	}
	openMCPFixture(t, m, b, pb)
	writeMCPFixture(t, a, `{"mcpServers":{"changed":{"command":"fixture"}}}`, false)
	fresh := openMCPFixture(t, m, a, pa)
	if fresh.PID == first.PID {
		t.Fatal("fixture did not replace stale target")
	}
	once.Do(func() { close(release) })
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("command was accepted after its runtime was detached")
		}
	case <-time.After(time.Second):
		t.Fatal("command did not return")
	}
	if m.Snapshot().Running {
		t.Fatal("stale draft reached replacement runtime")
	}
}
