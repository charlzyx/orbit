package pirpc

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Opt in explicitly with ORBIT_PI_INTEGRATION_BIN. PI_BIN alone never enables
// tests, and this test never sends prompt, bash, steer, follow_up, or login.
func TestLiveCredentialFreePi(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN to official Pi 1.4.0 for offline verification")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	agent := filepath.Join(home, ".pi", "agent")
	workspace := filepath.Join(home, "workspace")
	if err = os.MkdirAll(agent, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	environment := []string{"HOME=" + home, "USERPROFILE=" + home, "PI_CODING_AGENT_DIR=" + agent, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1", "PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH")}
	version := exec.Command(binary, "--version")
	version.Env = environment
	version.Dir = workspace
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected pinned Pi 1.4.0: %q %v", raw, err)
	}
	client, err := Start(Config{Executable: binary, WorkspaceDir: workspace, Environment: environment, Offline: true, NoSession: true, Args: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools"}})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, command := range []string{GetState, GetAvailableModels, Abort, GetAvailableThinkingLevels, GetMessages, GetSessionStats, GetEntries, GetCommands} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		response, err := client.Call(ctx, NewCommand(command, nil))
		cancel()
		if err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		if !response.Success {
			t.Fatal(response)
		}
		if command == GetSessionStats {
			var stats struct {
				Cost      float64
				Tokens    struct{ Total int }
				ToolCalls int
			}
			if err = json.Unmarshal(response.Data, &stats); err != nil {
				t.Fatal(err)
			}
			if stats.Cost != 0 || stats.Tokens.Total != 0 || stats.ToolCalls != 0 {
				t.Fatal("unexpected model/tool usage")
			}
		}
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	for _, event := range client.DrainEvents() {
		if event.Type == "agent_start" || event.Type == "message_start" || event.Type == "tool_execution_start" {
			t.Fatal("unexpected run event", event.Type)
		}
	}
	if client.ExitCode() != 0 {
		t.Fatal("unclean shutdown", client.ExitCode())
	}
}
