package orbitui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/agent"
)

func TestPaletteRenameOfflineColdSavedRowUsesMetadataLease(t *testing.T) {
	x := offlineRenameController(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "cold-saved.jsonl")
	const id = "01900000-0000-7000-8000-000000000002"
	header, _ := json.Marshal(map[string]any{"type": "session", "version": 3, "id": id, "cwd": workspace, "timestamp": "2026-01-01T00:00:00Z"})
	content := append(header, '\n')
	content = append(content, []byte(`{"type":"message","id":"entry456","parentId":null,"timestamp":"2026-01-01T00:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Cold synthetic fixture"}],"timestamp":1767225600000}}`+"\n")...)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	x.UI.commandSession("RenameSession", SessionRow{ID: id, Path: path, Name: "Cold fixture"})
	if !x.UI.promptOpen {
		t.Fatal("cold row prompt missing")
	}
	x.UI.promptCommit("Cold saved renamed")
	s := waitOfflineRename(t, x, func(s agent.Snapshot) bool {
		row, err := agent.ReadSession(path)
		return err == nil && row.Name == "Cold saved renamed" && row.ID == id
	})
	if s.ActiveSession.ID != "" || s.Running || len(s.Messages) != 0 {
		t.Fatal("metadata rename activated cold task or inference", s.ActiveSession)
	}
}
