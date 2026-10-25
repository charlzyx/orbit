package orbitui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/diagnosticlog"
)

func TestWorkspaceDiagnosticsContainStagesNotUserData(t *testing.T) {
	home, initial := t.TempDir(), t.TempDir()
	target := filepath.Join(t.TempDir(), "private-workspace-name-not-for-logs")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	log, err := diagnosticlog.Open(home, "0.2.12-go-test.20261025.4")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	x, err := NewController(RunOptions{Home: home, Workspace: initial, PiBinary: filepath.Join(home, "missing-synthetic-pi"), Offline: true, NoExtensions: true, Diagnostics: log})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	x.UI.Draft = "private conversation text must not reach logs"
	x.SetWorkspace(target)
	x.SetWorkspace("")
	x.sessionGate.Lock()
	x.SetWorkspace(initial)
	x.sessionGate.Unlock()
	data, err := os.ReadFile(filepath.Join(home, ".orbit-pi", "logs", "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{home, initial, target, "private-workspace-name", x.UI.Draft, "missing-synthetic-pi"} {
		if forbidden != "" && strings.Contains(string(data), forbidden) {
			t.Fatalf("private application data reached diagnostic log: %q", forbidden)
		}
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		if len(row) != 6 {
			t.Fatalf("unexpected diagnostic fields: %v", row)
		}
		seen[row["event"].(string)+":"+row["result"].(string)] = true
	}
	for _, key := range []string{"controller_init:started", "controller_init:ok", "workspace_select:started", "workspace_select:cancelled", "workspace_select:busy", "workspace_services:started", "workspace_services:ok", "workspace_preferences:ok", "workspace_session:started", "workspace_files:started", "workspace_files:ok"} {
		if !seen[key] {
			t.Errorf("missing diagnostic stage %s", key)
		}
	}
	if x.workspace != target || x.Store.LastWorkspace() != target {
		t.Fatal("logging changed workspace/cancel/busy behavior")
	}
}
