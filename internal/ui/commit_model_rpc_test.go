package orbitui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workspace"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommitModelRPCPeer(t *testing.T) {
	if os.Getenv("ORBIT_COMMIT_MODEL_PEER") != "1" {
		return
	}
	logPath := os.Getenv("ORBIT_COMMIT_MODEL_LOG")
	if !strings.Contains(strings.Join(os.Args, " "), "--no-tools") {
		logPath += ".active"
	}
	log, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		os.Exit(3)
	}
	defer log.Close()
	scan := bufio.NewScanner(os.Stdin)
	scan.Buffer(make([]byte, 4096), 4<<20)
	enc := json.NewEncoder(os.Stdout)
	for scan.Scan() {
		var q map[string]any
		if json.Unmarshal(scan.Bytes(), &q) != nil {
			os.Exit(4)
		}
		fmt.Fprintln(log, string(scan.Bytes()))
		r := map[string]any{"type": "response", "id": q["id"], "command": q["type"], "success": true}
		if q["type"] == "get_state" && os.Getenv("ORBIT_COMMIT_ACTIVE_MODEL") == "1" {
			r["data"] = map[string]any{"sessionId": "synthetic-session", "model": map[string]any{"provider": func() string {
				if os.Getenv("ORBIT_COMMIT_PARTIAL_MODEL") == "1" {
					return ""
				}
				return "active-provider"
			}(), "id": "active-model", "contextWindow": 32000}}
		}
		if q["type"] == "set_model" && os.Getenv("ORBIT_COMMIT_PARTIAL_MODEL") == "1" {
			r["success"] = false
			r["error"] = "synthetic unknown provider"
		}
		if q["type"] == "get_last_assistant_text" {
			r["data"] = map[string]any{"text": "feat: describe fixture"}
		}
		_ = enc.Encode(r)
		if q["type"] == "prompt" {
			_ = enc.Encode(map[string]any{"type": "agent_end"})
		}
	}
	os.Exit(0)
}
func commitHostFixture(t *testing.T, active bool) (*Controller, string, string) {
	t.Helper()
	home, dir := t.TempDir(), t.TempDir()
	cmd := exec.Command("git", "init", "-q", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	if err := os.WriteFile(filepath.Join(dir, "fixture.txt"), []byte("synthetic change\n"), 0600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(home, "rpc.jsonl")
	bin := filepath.Join(home, "fixture-pi")
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	script := "#!/bin/sh\nexport ORBIT_COMMIT_MODEL_PEER=1\nexport ORBIT_COMMIT_MODEL_LOG=" + quote(log) + "\nexec " + quote(os.Args[0]) + " -test.run=^TestCommitModelRPCPeer$ -- \"$@\"\n"
	if active {
		script = strings.Replace(script, "export ORBIT_COMMIT_MODEL_PEER=1", "export ORBIT_COMMIT_MODEL_PEER=1\nexport ORBIT_COMMIT_ACTIVE_MODEL=1", 1)
	}
	if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	x, err := NewController(RunOptions{Home: home, Workspace: dir, PiBinary: bin, Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		x.Git.Tick()
		if !x.Git.Busy && (!active || x.Agent.Snapshot().Model.ID == "active-model") {
			return x, dir, log
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("host fixture did not settle", x.Agent.Snapshot().Model)
	return nil, "", ""
}
func TestCommitDedicatedModelRealJSONL(t *testing.T) {
	x, dir, log := commitHostFixture(t, false)
	var err error
	if err = x.Store.Write("commit-model.json", map[string]any{"model": map[string]string{"provider": "dedicated-provider", "id": "dedicated-model"}}); err != nil {
		t.Fatal(err)
	}
	x.bindGenerators(dir)
	startupDeadline := time.Now().Add(5 * time.Second)
	for x.Git.Busy && time.Now().Before(startupDeadline) {
		x.Git.Tick()
		time.Sleep(time.Millisecond)
	}
	if x.Git.Busy {
		t.Fatal("initial refresh not settled")
	}
	x.Git.IncludeUnstaged = true
	x.Git.RequestGenerate(false, false)
	deadline := time.Now().Add(5 * time.Second)
	for x.Git.Busy && time.Now().Before(deadline) {
		x.Git.Tick()
		time.Sleep(time.Millisecond)
	}
	if x.Git.Busy || x.Git.Error != nil {
		t.Fatal(x.Git.Busy, x.Git.Error)
	}
	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var q map[string]any
		if err = json.Unmarshal([]byte(line), &q); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, q["type"].(string))
		if q["type"] == "set_model" && (q["provider"] != "dedicated-provider" || q["modelId"] != "dedicated-model") {
			t.Fatal(q)
		}
	}
	t.Logf("disposable commit child RPC commands: %v", kinds)
	if len(kinds) < 2 || kinds[0] != "set_model" || kinds[1] != "prompt" {
		t.Fatalf("commit child did not use explicit private model before prompt: %v", kinds)
	}
	if x.Agent.Snapshot().Model.ID != "" || x.Git.CommitMessage != "feat: describe fixture" {
		t.Fatal("chat changed or draft not applied", x.Agent.Snapshot().Model, x.Git.CommitMessage)
	}
	out, err := exec.CommandContext(context.Background(), "git", "-C", dir, "diff", "--cached", "--name-only").CombinedOutput()
	if err != nil || len(out) != 0 {
		t.Fatal("index mutated", err, string(out))
	}
}

func TestCommitFollowAndDraftUseOnlyActiveModelJSONL(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(fmt.Sprint(active), func(t *testing.T) {
			x, _, log := commitHostFixture(t, active)
			x.Git.IncludeUnstaged = true
			x.Git.RequestGenerate(false, false)
			wait := func() {
				deadline := time.Now().Add(5 * time.Second)
				for x.Git.Busy && time.Now().Before(deadline) {
					x.Git.Tick()
					time.Sleep(time.Millisecond)
				}
				if x.Git.Busy || x.Git.Error != nil {
					t.Fatal(x.Git.Busy, x.Git.Error)
				}
			}
			wait()
			check := func() {
				t.Helper()
				raw, e := os.ReadFile(log)
				if e != nil {
					t.Fatal(e)
				}
				var kinds []string
				for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
					var q map[string]any
					if e = json.Unmarshal([]byte(line), &q); e != nil {
						t.Fatal(e)
					}
					kinds = append(kinds, q["type"].(string))
					if q["type"] == "set_model" && (q["provider"] != "active-provider" || q["modelId"] != "active-model") {
						t.Fatal("private override leaked into draft/follow", q)
					}
				}
				t.Logf("active-model=%t disposable RPC commands: %v", active, kinds)
				if active {
					if len(kinds) < 2 || kinds[0] != "set_model" || kinds[1] != "prompt" {
						t.Fatal(kinds)
					}
				} else {
					for _, kind := range kinds {
						if kind == "set_model" {
							t.Fatal("no model forced", kinds)
						}
					}
				}
			}
			check()
			if e := x.Store.SaveCommitModel(&preferences.Model{Provider: "never-draft", ID: "private"}); e != nil {
				t.Fatal(e)
			}
			for _, kind := range []string{"issue", "pull"} {
				if e := os.Remove(log); e != nil {
					t.Fatal(e)
				}
				_, e := x.Git.Options.GenerateDraft(context.Background(), workspace.DraftRequest{Kind: kind, Notes: "synthetic request"})
				if e != nil {
					t.Fatal(e)
				}
				check()
			}
			raw, e := os.ReadFile(log + ".active")
			if e != nil {
				t.Fatal(e)
			}
			if strings.Contains(string(raw), `"type":"set_model"`) {
				t.Fatal("active chat received set_model", string(raw))
			}
		})
	}
}

func TestCommitPartialActiveIdentityIsAttemptedNotDefaultedJSONL(t *testing.T) {
	t.Setenv("ORBIT_COMMIT_PARTIAL_MODEL", "1")
	x, _, log := commitHostFixture(t, true)
	if m := x.Agent.Snapshot().Model; m.ID != "active-model" || m.Provider != "" {
		t.Fatal(m)
	}
	x.Git.IncludeUnstaged = true
	x.Git.RequestGenerate(false, false)
	deadline := time.Now().Add(5 * time.Second)
	for x.Git.Busy && time.Now().Before(deadline) {
		x.Git.Tick()
		time.Sleep(time.Millisecond)
	}
	if x.Git.Busy {
		t.Fatal("generation did not settle")
	}
	raw, e := os.ReadFile(log)
	if e != nil {
		t.Fatal(e)
	}
	var commands []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var q map[string]any
		if e = json.Unmarshal([]byte(line), &q); e != nil {
			t.Fatal(e)
		}
		commands = append(commands, q)
	}
	if len(commands) != 1 || commands[0]["type"] != "set_model" || commands[0]["modelId"] != "active-model" || commands[0]["provider"] != "" {
		t.Fatalf("partial identity silently defaulted or prompted after failed set_model: %s", raw)
	}
	if !strings.Contains(x.Git.Status, "Editable fallback:") || x.Git.CommitMessage == "" {
		t.Fatal("explicit fallback warning/draft missing", x.Git.Error, x.Git.CommitMessage)
	}
	active, e := os.ReadFile(log + ".active")
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(active), `"type":"set_model"`) {
		t.Fatal("active model mutated")
	}
}
