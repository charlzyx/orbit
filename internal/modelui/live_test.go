package modelui

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/preferences"
)

// This integration starts only a disposable credential-free, offline Pi. The
// only commands are the model/thinking catalog getters; it never sends a prompt,
// tool, authentication request or provider/network catalog refresh.
func TestLiveCredentialFreeCatalogProjection(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN to pinned Pi 1.4.0 for offline metadata verification")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	home, workspace := t.TempDir(), t.TempDir()
	store := preferences.New(home)
	if err := os.MkdirAll(store.PiDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.PiDir, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	environment := []string{"HOME=" + home, "USERPROFILE=" + home, "PI_CODING_AGENT_DIR=" + store.PiDir, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1", "PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH")}
	version := exec.Command(binary, "--version")
	version.Env = environment
	version.Dir = workspace
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected pinned Pi 1.4.0: %q %v", raw, err)
	}
	client, err := pirpc.Start(pirpc.Config{Executable: binary, WorkspaceDir: workspace, Environment: environment, Offline: true, NoSession: true, Args: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools"}})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	call := func(command string) json.RawMessage {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		response, err := client.Call(ctx, pirpc.NewCommand(command, nil))
		if err != nil || !response.Success {
			t.Fatalf("%s: %v %+v", command, err, response)
		}
		return response.Data
	}
	var catalog struct{ Models []json.RawMessage }
	if err := json.Unmarshal(call(pirpc.GetAvailableModels), &catalog); err != nil {
		t.Fatal(err)
	}
	var levels struct{ Levels []string }
	if err := json.Unmarshal(call(pirpc.GetAvailableThinkingLevels), &levels); err != nil {
		t.Fatal(err)
	}
	var models []agent.ModelChoice
	for _, raw := range catalog.Models {
		var m agent.ModelChoice
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		m.Raw = raw
		models = append(models, m)
	}
	s := New(store, Options{})
	s.SetSnapshot(Snapshot{Connected: true, Models: models, Levels: levels.Levels})
	if len(models) != 0 {
		t.Fatal("credential-free fixture unexpectedly has authenticated models", len(models))
	}
	if len(s.Groups()) != 0 || s.CountLabel() != "No models reported by the runtime" {
		t.Fatal("projection fabricated rows")
	}
	tt := nativeTester(s, "models", 1000, 700)
	tt.Frame()
	if _, ok := tt.Find("No models reported"); !ok {
		t.Fatal("real empty response missing native empty state")
	}
	s.PickerOpen(ModelPicker)
	if len(s.PickerRows()) != 0 || s.ConfirmPicker() {
		t.Fatal("real empty response selectable")
	}
	if len(levels.Levels) <= 1 && s.PickerOpen(ThinkingPicker) {
		t.Fatal("unavailable thinking opened")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	for _, event := range client.DrainEvents() {
		if event.Type == "agent_start" || event.Type == "message_start" || event.Type == "tool_execution_start" {
			t.Fatal("metadata test caused inference", event.Type)
		}
	}
}
