package agent

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func TestLiveManagerCredentialFree(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN for the pinned offline Pi integration")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	agentDir := filepath.Join(home, ".pi", "agent")
	if err = os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(agentDir, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	environment := []string{"PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME=" + home, "PI_CODING_AGENT_DIR=" + agentDir, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1"}
	version := exec.Command(binary, "--version")
	version.Env = environment
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected pinned Pi 1.4.0: %q %v", raw, err)
	}
	manager := NewManager(Options{PiBinary: binary, Home: home, Environment: environment, DisableExtensions: true, DisablePatches: true, Offline: true, NoSession: true, CommandTimeout: 15 * time.Second, PiArgs: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools", "--no-approve"}})
	defer manager.Close()
	if err = manager.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, manager, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	if snapshot.Running || len(snapshot.Messages) != 0 || len(snapshot.Models) != 0 {
		t.Fatal("unexpected authenticated/model state")
	}
	runtime := manager.RuntimeSnapshot()
	if runtime.Active == nil || runtime.Active.PID <= 0 || runtime.Active.StartedAt.IsZero() || !runtime.Active.Alive || runtime.Active.AutoRetry.Known {
		t.Fatalf("unexpected real process metadata: %+v", runtime.Active)
	}
	manager.mu.Lock()
	client := manager.sessions[manager.active].client
	manager.mu.Unlock()
	state, err := manager.callClient(client, pirpc.NewCommand(pirpc.GetState, nil))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = state.DecodeData(&fields); err != nil {
		t.Fatal(err)
	}
	if _, present := fields["autoRetryEnabled"]; present {
		t.Fatal("pinned Pi unexpectedly exposes retry getter; revisit documented contract")
	}
	for _, enabled := range []bool{false, true} {
		if err = manager.SetAutoRetry(enabled); err != nil {
			t.Fatal(err)
		}
		value := waitRuntime(t, manager, func(s RuntimeSnapshot) bool {
			return s.Active.AutoRetry.Known && !s.Active.AutoRetry.Pending && s.Active.AutoRetry.Enabled == enabled
		})
		if value.Active.AutoRetry.Support != "supported" {
			t.Fatal(value.Active.AutoRetry)
		}
	}
	if err = manager.Abort(); err != nil {
		t.Fatal(err)
	}
	initialID := snapshot.ActiveSession.ID
	if err = manager.NewSession(snapshot.ActiveSession.Workspace); err != nil {
		t.Fatal(err)
	}
	waitFor(t, manager, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.ID != initialID
	})
	if value := manager.RuntimeSnapshot(); value.Active.PID != runtime.Active.PID || !value.Active.StartedAt.Equal(runtime.Active.StartedAt) || !value.Active.AutoRetry.Known || !value.Active.AutoRetry.Enabled {
		t.Fatal("same-process new_session lost process metadata", value.Active)
	}
	if value := manager.Snapshot(); value.Running || len(value.Messages) != 0 {
		t.Fatal("metadata test unexpectedly started inference")
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLiveRenameSavedSessionNoInference(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN for offline metadata verification")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	workspace := t.TempDir()
	directory := filepath.Join(home, ".pi", "agent", "sessions", "test")
	if err = os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(home, ".pi", "agent", "settings.json")
	if err = os.WriteFile(settings, []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "saved.jsonl")
	header, _ := json.Marshal(map[string]any{"type": "session", "version": 3, "id": "01900000-0000-7000-8000-000000000001", "cwd": workspace, "timestamp": "2026-01-01T00:00:00Z"})
	content := append(header, '\n')
	content = append(content, []byte(`{"type":"message","id":"entry123","parentId":null,"timestamp":"2026-01-01T00:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Saved fixture only"}],"timestamp":1767225600000}}`+"\n")...)
	if err = os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(Options{PiBinary: binary, Home: home, Environment: []string{"PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH")}, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true, CommandTimeout: 15 * time.Second})
	defer manager.Close()
	if err = manager.RenameSavedSession(path, "Renamed offline"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		row, err := ReadSession(path)
		if err == nil && row.Name == "Renamed offline" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("rename not persisted", manager.Snapshot().ConnectionError)
}
