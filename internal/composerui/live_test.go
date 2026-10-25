package composerui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
)

// This starts the real pinned Pi RPC process with an isolated HOME and local
// fixture resources. It sends no prompts, reads no credentials, and is offline.
func TestLivePiCommandCatalogCredentialFree(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN to the pinned real Pi 1.4.0 executable")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	home, workspace := t.TempDir(), t.TempDir()
	agentDir := filepath.Join(home, ".pi", "agent")
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	for path, text := range map[string]string{".pi/prompts/contract.md": "---\ndescription: Review the fixture contract\n---\nRead the local contract.\n", ".pi/skills/scan/SKILL.md": "---\nname: scan\ndescription: Inspect the fixture workspace\n---\nRead the fixture workspace.\n"} {
		full := filepath.Join(workspace, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	env := []string{"PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME=" + home, "PI_CODING_AGENT_DIR=" + agentDir, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1"}
	version := exec.Command(binary, "--version")
	version.Env = env
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected Pi 1.4.0, got %q %v", raw, err)
	}
	manager := agent.NewManager(agent.Options{PiBinary: binary, Home: home, Environment: env, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true, NoSession: true, CommandTimeout: 15 * time.Second, PiArgs: []string{"--no-extensions", "--no-context-files", "--no-tools"}})
	defer manager.Close()
	if err := manager.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(20 * time.Second)
	var snapshot agent.Snapshot
	for time.Now().Before(deadline) {
		manager.Tick()
		snapshot = manager.Snapshot()
		if len(snapshot.Commands) > 0 || snapshot.ConnectionError != "" {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	commands := ParseCommands(snapshot.Commands, workspace, home)
	found := map[string]Command{}
	for _, command := range commands {
		found[command.Name] = command
	}
	if command, ok := found["contract"]; !ok || command.Scope != Project || command.Description != "Review the fixture contract" {
		t.Fatalf("actual RPC project prompt missing: %+v; diagnostics=%v", commands, append(snapshot.Diagnostics, snapshot.ConnectionError, snapshot.Status))
	}
	if command, ok := found["skill:scan"]; !ok || command.Scope != Skill || command.Description != "Inspect the fixture workspace" {
		t.Fatalf("actual RPC skill missing: %+v", commands)
	}
	c := New(nil)
	defer c.Close()
	c.SetContext(workspace, home, snapshot.ActiveSession.ID)
	c.SetCommands(snapshot.Commands)
	c.Sync("/sca", 4, true, false)
	if s := c.Snapshot(); !s.Open || len(s.Entries) != 1 || s.Entries[0].Name != "skill:scan" {
		t.Fatal(s)
	}
	if snapshot.Running || len(snapshot.Messages) != 0 {
		t.Fatal("catalog lookup must not run inference")
	}
}
