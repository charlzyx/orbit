package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
)

func TestModelBirthProcess(t *testing.T) {
	if os.Getenv("ORBIT_MODEL_BIRTH_HELPER") != "1" {
		return
	}
	model, thinking, id := "old", "off", "initial"
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var c map[string]any
		if json.Unmarshal(scanner.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		if f, err := os.OpenFile(os.Getenv("MODEL_COMMAND_LOG"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			f.Write(append(append([]byte(nil), scanner.Bytes()...), '\n'))
			f.Close()
		}
		r := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		switch kind {
		case "get_state":
			r["data"] = map[string]any{"sessionId": id, "model": map[string]any{"provider": "p", "id": model}, "thinkingLevel": thinking}
		case "get_available_models":
			r["data"] = map[string]any{"models": []any{map[string]any{"provider": "p", "id": "target", "name": "Target", "reasoning": true}}}
		case "get_available_thinking_levels":
			r["data"] = map[string]any{"levels": []string{"off", "high"}}
		case "new_session":
			id = "new"
			model = "old"
			thinking = "off"
			r["data"] = map[string]any{"cancelled": false}
		case "set_model":
			model, _ = c["modelId"].(string)
			thinking = "off"
			r["data"] = map[string]any{"provider": "p", "id": model}
		case "set_thinking_level":
			thinking, _ = c["level"].(string)
		case "get_messages":
			r["data"] = map[string]any{"messages": []any{}}
		case "get_commands":
			r["data"] = map[string]any{"commands": []any{}}
		case "get_entries":
			r["data"] = map[string]any{"entries": []any{}}
		case "auth.list", "quota.list":
			r["success"] = false
			r["error"] = "unsupported"
		}
		b, _ := json.Marshal(r)
		fmt.Println(string(b))
	}
	os.Exit(0)
}

func TestBirthDefaultRealChildOrdersThinkingAfterModelConfirmation(t *testing.T) {
	home, work := t.TempDir(), t.TempDir()
	log := filepath.Join(t.TempDir(), "commands.jsonl")
	if err := preferences.New(home).SaveSessionDefault(preferences.SessionDefault{Provider: "p", ModelID: "target", Thinking: "high"}); err != nil {
		t.Fatal(err)
	}
	m := testManager(t, func(o *Options) {
		o.Home = home
		o.PiPrefixArgs = []string{"-test.run=^TestModelBirthProcess$", "--"}
		o.EnvOverrides["ORBIT_MODEL_BIRTH_HELPER"] = "1"
		o.EnvOverrides["MODEL_COMMAND_LOG"] = log
	})
	if err := m.Open(work, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.NewSession(work); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		return s.ActiveSession.ID == "new" && s.Model.ID == "target" && s.ThinkingLevel == "high"
	})
	// Several repeated state refreshes must never reapply a completed birth.
	for i := 0; i < 5; i++ {
		m.Tick()
		time.Sleep(5 * time.Millisecond)
	}
	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	var sequence []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var c map[string]any
		json.Unmarshal([]byte(line), &c)
		sequence = append(sequence, c["type"].(string))
	}
	modelIndex, thinkingIndex, sets, thoughts := -1, -1, 0, 0
	for i, kind := range sequence {
		if kind == "set_model" {
			modelIndex = i
			sets++
		}
		if kind == "set_thinking_level" {
			thinkingIndex = i
			thoughts++
		}
	}
	if sets != 1 || thoughts != 1 || thinkingIndex <= modelIndex {
		t.Fatal(sequence)
	}
	sawState := false
	for _, kind := range sequence[modelIndex+1 : thinkingIndex] {
		if kind == "get_state" {
			sawState = true
		}
	}
	if !sawState {
		t.Fatal("thinking sent before model confirmation", sequence)
	}
}
