package orbitui

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
)

func offlineRenameController(t *testing.T) *Controller {
	t.Helper()
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("requires pinned offline Pi1.4.0")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	x := workspaceDialogFixture(t)
	_ = x.Agent.Close()
	home := t.TempDir()
	agentDir := filepath.Join(home, ".pi", "agent")
	if err = os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(agentDir, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	env := []string{"PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME=" + home, "PI_CODING_AGENT_DIR=" + agentDir, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1"}
	version := exec.Command(binary, "--version")
	version.Env = env
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected Pi1.4.0 %q %v", raw, err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: binary, Home: home, UserHome: home, PiAgentDir: agentDir, Environment: env, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true, CommandTimeout: 15 * time.Second, PiArgs: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools", "--no-approve"}})
	return x
}
func waitOfflineRename(t *testing.T, x *Controller, pred func(agent.Snapshot) bool) agent.Snapshot {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		s := x.Agent.Snapshot()
		if pred(s) {
			return s
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("offline rename state timeout", x.Agent.Snapshot())
	return agent.Snapshot{}
}
func TestPaletteRenameOfflineSavedFilePositiveController(t *testing.T) {
	x := offlineRenameController(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	header, _ := json.Marshal(map[string]any{"type": "session", "version": 3, "id": "01900000-0000-7000-8000-000000000001", "cwd": workspace, "timestamp": "2026-01-01T00:00:00Z"})
	content := append(header, '\n')
	content = append(content, []byte(`{"type":"message","id":"entry123","parentId":null,"timestamp":"2026-01-01T00:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Synthetic offline fixture only"}],"timestamp":1767225600000}}`+"\n")...)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := x.Agent.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	s := waitOfflineRename(t, x, func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.Path == path })
	x.UI.SetSnapshot(adaptSnapshot(s, workspace, x.pinned, x.archived))
	x.UI.command("RenameSession", "")
	x.UI.promptCommit("Saved fixture renamed")
	s = waitOfflineRename(t, x, func(s agent.Snapshot) bool {
		row, err := agent.ReadSession(path)
		return err == nil && row.Name == "Saved fixture renamed"
	})
	if s.Running || len(s.Messages) != 1 || s.ActiveSession.ID != "01900000-0000-7000-8000-000000000001" {
		t.Fatal("rename changed identity or started inference", s.ActiveSession, s.Running)
	}
}
func TestPaletteRenameOfflineFreshUnpersistedPiContract(t *testing.T) {
	x := offlineRenameController(t)
	workspace := t.TempDir()
	if err := x.Agent.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	before := waitOfflineRename(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.Path != ""
	})
	if _, err := os.Stat(before.ActiveSession.Path); !os.IsNotExist(err) {
		t.Fatal("new session already persisted", err)
	}
	if err := x.Agent.RenameSession("Fresh in-memory Pi contract"); err != nil {
		t.Fatal(err)
	}
	s := waitOfflineRename(t, x, func(s agent.Snapshot) bool { return s.ActiveSession.Name == "Fresh in-memory Pi contract" })
	if _, err := os.Stat(s.ActiveSession.Path); !os.IsNotExist(err) {
		t.Fatal("name-only session unexpectedly created file", err)
	}
	if s.ActiveSession.ID != before.ActiveSession.ID || s.Running || len(s.Messages) != 0 {
		t.Fatal("identity/inference changed")
	}
}
func TestPaletteRenameOfflineFreshUnpersistedControllerParity(t *testing.T) {
	x := offlineRenameController(t)
	workspace := t.TempDir()
	if err := x.Agent.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	s := waitOfflineRename(t, x, func(s agent.Snapshot) bool {
		return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.Path != ""
	})
	if _, err := os.Stat(s.ActiveSession.Path); !os.IsNotExist(err) {
		t.Fatal("new session already persisted", err)
	}
	x.UI.SetSnapshot(adaptSnapshot(s, workspace, x.pinned, x.archived))
	x.UI.command("RenameSession", "")
	x.UI.promptCommit("Fresh Controller rename")
	waitOfflineRename(t, x, func(s agent.Snapshot) bool { return s.ActiveSession.Name == "Fresh Controller rename" })
}

func TestPaletteRenameOfflineStaleUIRejectedBeforeAndAfterArm(t *testing.T) {
	for _, arm := range []bool{false, true} {
		t.Run(map[bool]string{false: "before-arm", true: "after-arm"}[arm], func(t *testing.T) {
			x := offlineRenameController(t)
			workspace := t.TempDir()
			if err := x.Agent.Open(workspace, ""); err != nil {
				t.Fatal(err)
			}
			before := waitOfflineRename(t, x, func(s agent.Snapshot) bool {
				return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.Path != ""
			})
			a := x.UI
			a.SetSnapshot(adaptSnapshot(before, workspace, x.pinned, x.archived))
			var commit func(string)
			if arm {
				a.command("RenameSession", "")
				if !a.promptOpen || a.paletteRename.target == nil {
					t.Fatal("live target not captured")
				}
				commit = a.promptCommit
			}
			if err := x.Agent.NewSession(workspace); err != nil {
				t.Fatal(err)
			}
			after := waitOfflineRename(t, x, func(s agent.Snapshot) bool {
				return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.ID != before.ActiveSession.ID
			})
			if after.RuntimeID != before.RuntimeID {
				t.Fatal("expected same-child NewSession fixture")
			}
			if a.Model.SessionID != before.ActiveSession.ID {
				t.Fatal("UI snapshot should deliberately remain old")
			}
			if arm {
				commit("Must not rename replacement")
			} else {
				a.command("RenameSession", "")
				if a.promptOpen || a.paletteRename != nil {
					t.Fatal("stale displayed identity captured replacement")
				}
			}
			current := x.Agent.Snapshot()
			if current.ActiveSession.ID != after.ActiveSession.ID || current.ActiveSession.Name != after.ActiveSession.Name || current.Running {
				t.Fatal("stale rename changed replacement", current.ActiveSession)
			}
			if _, err := os.Stat(before.ActiveSession.Path); !os.IsNotExist(err) {
				t.Fatal("old fresh path unexpectedly persisted", err)
			}
		})
	}
}
