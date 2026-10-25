package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func apply(t *testing.T, s *sessionState, raw string) {
	t.Helper()
	event, err := pirpc.ParseEvent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	s.apply(event)
}
func TestStreamingBlockIdentityAndAuthoritativeMessage(t *testing.T) {
	s := newState("/workspace", "")
	apply(t, &s, `{"type":"agent_start"}`)
	apply(t, &s, `{"type":"message_start","message":{"role":"assistant","content":[],"timestamp":1000}}`)
	id := s.snapshot.Messages[0].ID
	for _, event := range []string{
		`{"type":"message_update","assistantMessageEvent":{"type":"thinking_delta","contentIndex":0,"delta":"consider"}}`,
		`{"type":"message_update","assistantMessageEvent":{"type":"text_delta","contentIndex":1,"delta":"one"}}`,
		`{"type":"message_update","assistantMessageEvent":{"type":"text_delta","contentIndex":2,"delta":"two"}}`,
		`{"type":"message_update","assistantMessageEvent":{"type":"text_end","contentIndex":1,"content":"corrected"}}`,
	} {
		apply(t, &s, event)
	}
	message := s.snapshot.Messages[0]
	if message.Thinking != "consider" || message.Text != "correctedtwo" || len(message.Blocks) != 3 {
		t.Fatal(message)
	}
	apply(t, &s, `{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"authoritative"}],"timestamp":1000}}`)
	if len(s.snapshot.Messages) != 1 || s.snapshot.Messages[0].ID != id || s.snapshot.Messages[0].Text != "authoritative" || s.snapshot.Messages[0].Streaming {
		t.Fatal(s.snapshot.Messages)
	}
	apply(t, &s, `{"type":"agent_end","willRetry":true}`)
	if !s.snapshot.Running || !s.snapshot.Retrying {
		t.Fatal("agent_end treated as settled")
	}
	apply(t, &s, `{"type":"agent_settled","aborted":false}`)
	if s.snapshot.Running || s.snapshot.Retrying {
		t.Fatal("run did not settle")
	}
}
func TestToolActivityMergesFinalResult(t *testing.T) {
	s := newState("/workspace", "")
	apply(t, &s, `{"type":"tool_execution_start","toolCallId":"t1","toolName":"bash","args":{"command":"pwd"}}`)
	apply(t, &s, `{"type":"tool_execution_update","toolCallId":"t1","toolName":"bash","partialResult":{"content":[{"type":"text","text":"/work"}]}}`)
	apply(t, &s, `{"type":"tool_execution_end","toolCallId":"t1","toolName":"bash","result":{"content":[{"type":"text","text":"/workspace"}]},"isError":false}`)
	apply(t, &s, `{"type":"message_start","message":{"role":"toolResult","toolCallId":"t1","toolName":"bash","content":[{"type":"text","text":"/workspace"}]}}`)
	apply(t, &s, `{"type":"message_end","message":{"role":"toolResult","toolCallId":"t1","toolName":"bash","content":[{"type":"text","text":"/workspace"}]}}`)
	if len(s.snapshot.Messages) != 1 || s.snapshot.Messages[0].ToolOutput != "/workspace" || s.snapshot.Messages[0].ID != "tool:t1" {
		t.Fatal(s.snapshot.Messages)
	}
}
func TestDialogsCustomStatusAndEditor(t *testing.T) {
	s := newState("", "")
	apply(t, &s, `{"type":"extension_ui_request","id":"d","method":"select","title":"Allow?","options":["Allow once","Deny"]}`)
	apply(t, &s, `{"type":"extension_ui_request","id":"c","method":"custom","phase":"open","lines":["A"],"width":80,"height":24}`)
	apply(t, &s, `{"type":"extension_ui_request","id":"c","method":"custom","phase":"render","lines":["B"],"width":90,"height":20}`)
	if len(s.snapshot.Dialogs) != 2 || s.snapshot.Dialogs[1].Lines[0] != "B" {
		t.Fatal(s.snapshot.Dialogs)
	}
	apply(t, &s, `{"type":"extension_ui_request","id":"c","method":"custom","phase":"close"}`)
	apply(t, &s, `{"type":"extension_ui_request","id":"s","method":"setStatus","statusKey":"work","statusText":"working"}`)
	apply(t, &s, `{"type":"extension_ui_request","id":"e","method":"set_editor_text","text":"seed"}`)
	if len(s.snapshot.Dialogs) != 1 || s.snapshot.StatusEntries["work"] != "working" || s.snapshot.ComposerText != "seed" {
		t.Fatal(s.snapshot)
	}
}
func TestAuthQuotaAndEntryCursor(t *testing.T) {
	s := newState("", "")
	apply(t, &s, `{"type":"response","command":"auth.list","success":true,"data":{"providers":[{"id":"p","name":"Provider","credential":"none","authenticated":false,"methods":[{"id":"browser","label":"Sign in"}],"token":"must-not-enter-state"}]}}`)
	apply(t, &s, `{"type":"auth_login_started","sessionId":"one","provider":"p","method":"browser","expiresAt":999}`)
	apply(t, &s, `{"type":"auth_login_succeeded","sessionId":"different","provider":"p","credential":"oauth"}`)
	if s.snapshot.Auth.Providers[0].Authenticated {
		t.Fatal("stale auth event was accepted")
	}
	apply(t, &s, `{"type":"auth_login_succeeded","sessionId":"one","provider":"p","credential":"oauth"}`)
	raw, _ := json.Marshal(s.snapshot.Auth)
	if strings.Contains(string(raw), "must-not-enter-state") || !s.snapshot.Auth.Providers[0].Authenticated {
		t.Fatal(string(raw))
	}
	apply(t, &s, `{"type":"response","command":"quota.list","success":false,"error":"Unknown command: quota.list"}`)
	apply(t, &s, `{"type":"response","command":"get_entries","success":true,"data":{"entries":[{"id":"e1","type":"custom","customType":"orbit:quota","data":{"providers":[{"provider":"p","kind":"subscription","windows":[{"id":"w","label":"Week","usedPercent":25}]}]}},{"id":"e2","type":"custom","customType":"orbit:workflow","data":{"mode":"plan"}}],"leafId":"e2"}}`)
	if s.entriesCursor != "e2" || len(s.snapshot.Quotas) != 1 || s.snapshot.QuotaSupport != "unsupported" || s.snapshot.WorkflowMode != "plan" {
		t.Fatal(s.snapshot)
	}
}
func TestSessionChangeResetsTransientState(t *testing.T) {
	s := newState("/workspace", "")
	s.snapshot.ActiveSession.ID = "old"
	s.entriesCursor = "old-entry"
	s.snapshot.Messages = []Message{{ID: "old"}}
	apply(t, &s, `{"type":"response","command":"get_state","success":true,"data":{"sessionId":"new","sessionFile":"/session","model":{"id":"m","provider":"p","name":"Model"},"thinkingLevel":"high","isStreaming":false,"autoCompactionEnabled":true,"capabilities":{"extension_ui":["custom"]}}}`)
	if len(s.snapshot.Messages) != 0 || s.entriesCursor != "" || s.snapshot.ActiveSession.ID != "new" || s.snapshot.Model.ID != "m" {
		t.Fatal(s.snapshot)
	}
}
func TestSnapshotDeepCopyWithPartialToolJSON(t *testing.T) {
	s := Snapshot{Messages: []Message{{Blocks: []ContentBlock{{Arguments: json.RawMessage(`{"unfinished":`)}}}}, StatusEntries: map[string]string{"a": "b"}, Dialogs: []PendingDialog{{Options: []string{"one"}}}}
	clone := cloneValue(reflect.ValueOf(s)).Interface().(Snapshot)
	clone.Messages[0].Blocks[0].Arguments[0] = '!'
	clone.StatusEntries["a"] = "changed"
	clone.Dialogs[0].Options[0] = "changed"
	if s.Messages[0].Blocks[0].Arguments[0] != '{' || s.StatusEntries["a"] != "b" || s.Dialogs[0].Options[0] != "one" {
		t.Fatal("snapshot shares mutable storage")
	}
}
func TestSessionDiscoveryUsesMessageActivityNotQuotaWrites(t *testing.T) {
	root := t.TempDir()
	group := filepath.Join(root, "group")
	if err := os.Mkdir(group, 0700); err != nil {
		t.Fatal(err)
	}
	content := `{"type":"session","id":"s1","cwd":"/project","timestamp":"2026-01-01T00:00:00Z"}
{"type":"message","message":{"role":"user","content":"Initial request","timestamp":1767312000000}}
{"type":"session_info","name":"Named session","timestamp":"2026-03-01T00:00:00Z"}
{"type":"custom","customType":"orbit:quota","timestamp":"2026-04-01T00:00:00Z"}
{"partial":`
	path := filepath.Join(group, "one.jsonl")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	rows, err := ScanSessions(root)
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	if rows[0].Name != "Named session" || rows[0].Summary != "Initial request" || rows[0].Workspace != "/project" || !rows[0].Updated.Equal(time.UnixMilli(1767312000000)) {
		t.Fatal(rows[0])
	}
}
func TestExtensionMaterializationAndAtomicPersistence(t *testing.T) {
	root := t.TempDir()
	resources := fstest.MapFS{"example/index.js": {Data: []byte("export default function(){}")}, "example/policy.js": {Data: []byte("export const allow=false")}}
	if err := materialize(resources, "example", []string{"index.js", "policy.js"}, root); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, "index.js"))
	if err != nil {
		t.Fatal(err)
	}
	before := info.ModTime()
	if err = materialize(resources, "example", []string{"index.js", "policy.js"}, root); err != nil {
		t.Fatal(err)
	}
	info, _ = os.Stat(filepath.Join(root, "index.js"))
	if !info.ModTime().Equal(before) {
		t.Fatal("unchanged extension rewritten")
	}
	if err = atomicWrite(filepath.Join(root, "nested", "settings.json"), []byte(`{"mode":"plan"}`)); err != nil {
		t.Fatal(err)
	}
}

func TestClonePreservesCompleteEntriesAndProvenance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.jsonl")
	content := `{"type":"session","version":3,"id":"source","cwd":"/workspace","timestamp":"2026-01-01T00:00:00Z","future":"preserved"}
{"type":"message","id":"m1","parentId":null,"message":{"role":"user","content":"hello","timestamp":1767225600000}}
{"partial":`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	clone, err := CloneSessionFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(clone)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatal(string(raw))
	}
	var header map[string]any
	if err = json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatal(err)
	}
	if header["id"] == "source" || header["parentSession"] != path || header["future"] != "preserved" {
		t.Fatal(header)
	}
	original, _ := os.ReadFile(path)
	if string(original) != content {
		t.Fatal("clone changed source")
	}
}

func TestClearQueueResponseContainsRemovedNotCurrentQueue(t *testing.T) {
	s := newState("", "")
	s.snapshot.Queue = Queue{Steering: []string{"first"}, FollowUp: []string{"second"}}
	apply(t, &s, `{"type":"response","command":"clear_queue","success":true,"data":{"steering":["first"],"followUp":["second"]}}`)
	if len(s.snapshot.Queue.Steering) != 0 || len(s.snapshot.Queue.FollowUp) != 0 {
		t.Fatal("removed queue was restored as pending")
	}
}

func TestSuccessfulCompactionInvalidatesOnlyContextBaseline(t *testing.T) {
	for _, test := range []struct {
		name, event string
		invalidated bool
	}{
		{"success", `{"type":"compaction_end","result":{"summary":"summary"},"aborted":false}`, true},
		{"aborted", `{"type":"compaction_end","aborted":true}`, false},
		{"failed", `{"type":"compaction_end","errorMessage":"failed","aborted":false}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := newState("workspace", "session")
			s.snapshot.Compacting = true
			s.snapshot.Stats = json.RawMessage(`{"tokens":{"total":100},"cost":0.25,"contextUsage":{"tokens":80,"contextWindow":100,"percent":80}}`)
			apply(t, &s, test.event)
			var stats map[string]json.RawMessage
			if err := json.Unmarshal(s.snapshot.Stats, &stats); err != nil {
				t.Fatal(err)
			}
			if s.snapshot.Compacting || string(stats["cost"]) != "0.25" || string(stats["tokens"]) != `{"total":100}` {
				t.Fatal(string(s.snapshot.Stats))
			}
			if (len(stats["contextUsage"]) == 0) != test.invalidated {
				t.Fatal("wrong context baseline", string(s.snapshot.Stats))
			}
		})
	}
}

func TestAutoCompactionKnownRequiresObservedStateField(t *testing.T) {
	s := newState("workspace", "session")
	if s.snapshot.AutoCompactionKnown {
		t.Fatal("default is not observed state")
	}
	apply(t, &s, `{"type":"response","command":"get_state","success":true,"data":{"sessionId":"s","autoCompactionEnabled":false}}`)
	if !s.snapshot.AutoCompactionKnown || s.snapshot.AutoCompactionEnabled {
		t.Fatal(s.snapshot)
	}
	apply(t, &s, `{"type":"response","command":"get_state","success":true,"data":{"sessionId":"s"}}`)
	if s.snapshot.AutoCompactionKnown {
		t.Fatal("missing field treated as false")
	}
}
