package orbitui

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func TestSessionExportDialogChild(t *testing.T) {
	if os.Getenv("ORBIT_EXPORT_DIALOG_CHILD") != "1" {
		return
	}
	scan := bufio.NewScanner(os.Stdin)
	number := 1
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		data := map[string]any{}
		switch kind {
		case "get_state":
			data["sessionId"] = fmt.Sprintf("dialog-%d", number)
		case "new_session":
			number++
			data["cancelled"] = false
		case "export_html":
			p, _ := c["outputPath"].(string)
			_ = os.WriteFile(p, []byte(fmt.Sprintf("dialog-%d", number)), 0600)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"type": "response", "command": kind, "id": c["id"], "success": true, "data": data})
	}
	os.Exit(0)
}
func sessionExportDialogFixture(t *testing.T) *Controller {
	t.Helper()
	x := workspaceDialogFixture(t)
	_ = x.Agent.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestSessionExportDialogChild$", "--"}, Home: x.Store.Home, Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_EXPORT_DIALOG_CHILD": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		if x.Agent.Snapshot().Connected {
			return x
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("export fixture not ready")
	return nil
}
func TestSessionExportDialogAsyncExactTaskAndCancellation(t *testing.T) {
	for _, mode := range []string{"confirm", "cancel", "error", "new-task", "window", "close"} {
		t.Run(mode, func(t *testing.T) {
			x := sessionExportDialogFixture(t)
			path := filepath.Join(t.TempDir(), "导出 session.html")
			if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			paths := []string{path}
			var err error
			if mode == "cancel" {
				paths = nil
			}
			if mode == "error" {
				err = errors.New("synthetic native save error")
			}
			release := holdFileDialog(t, x, Action{Kind: "ExportSession"}, paths, err)
			switch mode {
			case "new-task":
				if err := x.Agent.NewSession(x.workspace); err != nil {
					t.Fatal(err)
				}
			case "window":
				x.UI.Window = new(mygo.Window)
			case "close":
				x.Close()
			}
			release()
			if mode == "close" {
				return
			}
			drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
			x.UI.Window = nil
			deadline := time.Now().Add(2 * time.Second)
			for {
				x.Agent.Tick()
				b, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				if mode != "confirm" {
					if string(b) != "original" {
						t.Fatal("stale/cancelled dialog overwrote file")
					}
					break
				}
				if string(b) == "dialog-1" {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("confirmed target not exported")
				}
				time.Sleep(time.Millisecond)
			}
		})
	}
}

func TestSessionExportCompletionConsumedByActualControllerFrame(t *testing.T) {
	x := sessionExportDialogFixture(t)
	path := filepath.Join(t.TempDir(), "frame-export.html")
	release := holdFileDialog(t, x, Action{Kind: "ExportSession"}, []string{path}, nil)
	tt := ui.NewTester(x.View, 1470, 922)
	tt.Frame()
	release()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		tt.Frame()
		if data, err := os.ReadFile(path); err == nil && string(data) == "dialog-1" && !x.fileDialogs.pending {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("actual Controller frame did not consume and export native dialog answer")
}
