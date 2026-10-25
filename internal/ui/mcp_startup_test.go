package orbitui

import (
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestControllerRegistersRealMCPStartupProbe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("synthetic shell; separate cross build")
	}
	home, workspace := t.TempDir(), t.TempDir()
	marker := filepath.Join(t.TempDir(), "probe-args")
	bin := filepath.Join(t.TempDir(), "synthetic-pi")
	script := "#!/bin/sh\nif [ \"$1\" = mcp ]; then printf '%s\\n' \"$*\" >> '" + marker + "'; printf '%s\\n' '{\"servers\":[{\"name\":\"fixture\",\"state\":\"connected\"}]}'; fi\n"
	if e := os.WriteFile(bin, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	if e := workbench.New(preferences.New(home), workspace).SaveMCP("global", "", "fixture", map[string]any{"command": "not-executed"}); e != nil {
		t.Fatal(e)
	}
	x, e := NewController(RunOptions{Home: home, Workspace: workspace, PiBinary: bin, Offline: true, NoExtensions: true})
	if e != nil {
		t.Fatal(e)
	}
	defer x.Close()
	x.MCP.Tick()
	if _, e := os.Stat(marker); !os.IsNotExist(e) {
		t.Fatal("controller immediately ran deferred probe", e)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		x.MCP.Tick()
		if x.MCP.Runtime["fixture"].State == "connected" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	raw, e := os.ReadFile(marker)
	if e != nil || strings.Count(string(raw), "mcp list --json") != 1 || x.MCP.Runtime["fixture"].State != "connected" {
		t.Fatal("controller's startup hook did not consume real probe", string(raw), e, x.MCP.Error)
	}
}
