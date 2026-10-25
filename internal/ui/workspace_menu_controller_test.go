package orbitui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceMenuRPCChild(t *testing.T) {
	if os.Getenv("ORBIT_WORKSPACE_MENU_CHILD") != "1" {
		return
	}
	n := 1
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		r := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		switch kind {
		case "new_session":
			n++
			r["data"] = map[string]any{"cancelled": false}
		case "get_state":
			r["data"] = map[string]any{"sessionId": fmt.Sprintf("menu-%d-%d", os.Getpid(), n)}
		case "get_messages":
			r["data"] = map[string]any{"messages": []any{}}
		case "get_available_models":
			r["data"] = map[string]any{"models": []any{}}
		case "get_available_thinking_levels":
			r["data"] = map[string]any{"levels": []string{"off"}}
		case "get_commands":
			r["data"] = map[string]any{"commands": []any{}}
		case "get_entries":
			r["data"] = map[string]any{"entries": []any{}}
		case "auth.list", "quota.list":
			r["success"], r["error"] = false, "Unknown command: "+kind
		}
		_ = json.NewEncoder(os.Stdout).Encode(r)
	}
	os.Exit(0)
}
func workspaceMenuController(t *testing.T) (*Controller, string, string, func(func(agent.Snapshot) bool)) {
	t.Helper()
	x, wait := newHeartbeatRPCController(t)
	_ = x.Agent.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestWorkspaceMenuRPCChild$", "--"}, Home: x.Store.Home, SessionDir: filepath.Join(x.Store.PiDir, "sessions"), Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_WORKSPACE_MENU_CHILD": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	a := x.workspace
	b := filepath.Join(t.TempDir(), "inactive-b")
	if err := os.MkdirAll(b, 0700); err != nil {
		t.Fatal(err)
	}
	if err := x.Agent.Open(a, ""); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !x.Agent.AnyBusy() })
	if err := x.Store.AddWorkspace(a); err != nil {
		t.Fatal(err)
	}
	if err := x.Store.AddWorkspace(b); err != nil {
		t.Fatal(err)
	}
	x.UI.Workspaces = x.Store.Workspaces()
	x.UI.Navigate("chat")
	return x, a, b, wait
}
func TestWorkspaceMenuWorktreesMatchesRustCurrentRepository(t *testing.T) {
	x, a, b, _ := workspaceMenuController(t)
	tt := ui.NewTester(x.UI.View, 1222, 878)
	tt.Frame()
	if err := tt.Click("Workspace actions " + b); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Worktrees"); err != nil {
		t.Fatal(err)
	}
	// Rust sidebar.rs:1581 closes menu then open_worktrees(); worktrees.rs:85
	// obtains current_workspace, not the inactive header's cwd. Preserve that.
	if x.UI.page != "worktrees" || x.Worktrees.Workspace != a || x.workspace != a || x.UI.Model.Workspace != a {
		t.Fatal("worktree action changed source target", x.UI.page, x.Worktrees.Workspace, x.workspace)
	}
}
func TestWorkspaceMenuNewTaskUsesRealControllerAndFreshIdentity(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(fmt.Sprint(same), func(t *testing.T) {
			x, a, b, wait := workspaceMenuController(t)
			target := b
			if same {
				target = a
			}
			before := x.Agent.Snapshot()
			tt := ui.NewTester(x.UI.View, 1222, 878)
			tt.Frame()
			if err := tt.Click("Workspace actions " + target); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem("New task here"); err != nil {
				t.Fatal(err)
			}
			wait(func(s agent.Snapshot) bool {
				return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.ID != before.ActiveSession.ID && s.ActiveSession.Workspace == target && !x.Agent.AnyBusy()
			})
			if x.workspace != target || x.Worktrees.Workspace != target || x.UI.page != "chat" {
				t.Fatal("new task did not coordinate target services", x.workspace, x.Worktrees.Workspace, x.UI.page)
			}
		})
	}
}
