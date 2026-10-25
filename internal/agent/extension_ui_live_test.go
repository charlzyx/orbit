package agent

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This exercises the real pinned Pi extension host's ctx.ui requests and
// responses over JSONL, without provider credentials or any model inference.
// Structured tool execution/replay is covered by deterministic event transcripts;
// this command fixture deliberately does not fabricate tool_execution events.
func TestLiveExtensionUIWireCredentialFree(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN for pinned Pi 1.4.0 extension RPC")
	}
	home, work := t.TempDir(), t.TempDir()
	agentDir := filepath.Join(home, ".pi", "agent")
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	env := []string{"PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME=" + home, "PI_CODING_AGENT_DIR=" + agentDir, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1"}
	version := exec.Command(binary, "--version")
	version.Env = env
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected Pi 1.4.0: %q %v", raw, err)
	}
	extension := filepath.Join(home, "ui-contract.ts")
	source := `export default function(pi) {
 pi.registerCommand("orbit-ui-contract", {description:"Credential-free UI protocol fixture",handler:async(_args,ctx)=>{
  ctx.ui.setStatus("fixture","Quiet status");
  ctx.ui.notify("Initial toast","info");
  ctx.ui.setEditorText("seed");ctx.ui.setEditorText("seed");ctx.ui.setEditorText("");
  ctx.ui.setWidget("fixture",["Above"],{placement:"aboveEditor"});
  const guard=await ctx.ui.select("[orbit-guard] bash\techo fixture",["Allow once","Always allow this tool","Deny"]);
  ctx.ui.setWidget("fixture",["Below"],{placement:"belowEditor"});
  const input=await ctx.ui.input("Live input","Live placeholder");
  const confirmed=await ctx.ui.confirm("Live confirm","Confirm detail");
  const editor=await ctx.ui.editor("Live editor","Seed text");
  const [first,second]=await Promise.all([ctx.ui.select("First concurrent",["A","B"]),ctx.ui.select("Second concurrent",["C","D"])]);
  ctx.ui.setWidget("fixture",undefined);
  ctx.ui.setStatus("fixture",undefined);
  ctx.ui.notify("RESULT:"+JSON.stringify({guard,input,confirmed,editor,first,second:second??null}),"info");
 }});
}`
	if err = os.WriteFile(extension, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	m := NewManager(Options{PiBinary: binary, Home: home, Environment: env, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true, NoSession: true, CommandTimeout: 15 * time.Second, PiArgs: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools", "--no-approve", "--extension", extension}})
	defer m.Close()
	if err = m.Open(work, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err = m.Submit("/orbit-ui-contract", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	s := waitFor(t, m, func(s Snapshot) bool { return s.Approval != nil })
	if s.ComposerRevision != 3 || s.ComposerText != "" || len(s.ExtensionWidgets) != 1 || s.ExtensionWidgets[0].Placement != "aboveEditor" || len(s.ExtensionNotices) != 1 || s.StatusEntries["fixture"] != "Quiet status" {
		t.Fatal(s)
	}
	if err = m.RespondDialog(s.Approval.ID, "Allow once", nil, false); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 1 && s.Dialogs[0].Title == "Live input" })
	if s.Dialogs[0].Placeholder != "Live placeholder" || s.ExtensionWidgets[0].Placement != "belowEditor" {
		t.Fatal(s)
	}
	if err = m.RespondDialog(s.Dialogs[0].ID, "", nil, false); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 1 && s.Dialogs[0].Title == "Live confirm" })
	no := false
	if err = m.RespondDialog(s.Dialogs[0].ID, "", &no, false); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 1 && s.Dialogs[0].Title == "Live editor" })
	if s.Dialogs[0].Value != "Seed text" {
		t.Fatal(s.Dialogs)
	}
	if err = m.RespondDialog(s.Dialogs[0].ID, "Edited\ntext", nil, false); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 1 && s.Dialogs[0].Title == "First concurrent" })
	if err = m.RespondDialog(s.Dialogs[0].ID, "B", nil, false); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool {
		for _, n := range s.ExtensionNotices {
			if strings.HasPrefix(n.Message, "RESULT:") {
				return true
			}
		}
		return false
	})
	var result struct {
		Guard, Input, Editor, First string
		Confirmed                   bool
		Second                      *string
	}
	for _, notice := range s.ExtensionNotices {
		if strings.HasPrefix(notice.Message, "RESULT:") {
			if err = json.Unmarshal([]byte(strings.TrimPrefix(notice.Message, "RESULT:")), &result); err != nil {
				t.Fatal(err)
			}
		}
	}
	if result.Guard != "Allow once" || result.Input != "" || result.Confirmed || result.Editor != "Edited\ntext" || result.First != "B" || result.Second != nil {
		t.Fatal(result)
	}
	if len(s.ExtensionWidgets) != 0 || len(s.StatusEntries) != 0 || len(s.Dialogs) != 0 || len(s.Messages) != 0 {
		t.Fatal("fixture left UI or started model inference", s)
	}
}
