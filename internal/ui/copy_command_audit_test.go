package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestAuditHostMenuCommandCopy(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "s", Messages: []Message{{ID: "u", Role: "user", Text: "First 用户 message."}, {ID: "a", Role: "assistant", Text: "Second **styled** message.\n\n> [!NOTE]\n> Last alert line."}}})
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	first, ok := tt.Find("First 用户 message.")
	if !ok {
		t.Fatal("user missing")
	}
	last, ok := tt.Find("Last alert line.")
	if !ok {
		t.Fatal("alert missing")
	}
	tt.Press(first.X, first.Y+first.H/2)
	tt.Move(last.X+last.W-1, last.Y+last.H/2)
	tt.Release(last.X+last.W-1, last.Y+last.H/2)
	tt.Command("copy")
	text := tt.Clipboard()
	for _, want := range []string{"First 用户 message.", "Second styled message.", "Last alert line."} {
		if !strings.Contains(text, want) {
			t.Fatalf("cross-message alert selection missing %q: %q", want, text)
		}
	}
}

func TestTranscriptMenuCopyComposerAndReset(t *testing.T) {
	a := fixture(t)
	snap := Snapshot{SessionID: "copy-command", Messages: []Message{{ID: "a", Role: "assistant", Text: "SELECTED transcript text."}}}
	a.SetSnapshot(snap)
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	selectTranscript := func() {
		t.Helper()
		r, ok := tt.Find("SELECTED transcript text.")
		if !ok {
			t.Fatal("message missing")
		}
		tt.Press(r.X, r.Y+r.H/2)
		tt.Move(r.X+r.W-1, r.Y+r.H/2)
		tt.Release(r.X+r.W-1, r.Y+r.H/2)
	}
	selectTranscript()
	tt.Command("copy")
	if tt.Clipboard() != "SELECTED transcript text." {
		t.Fatal("selection copy", tt.Clipboard())
	}
	// Clicking the composer clears retained transcript selection; native edit
	// commands then belong exclusively to the text input.
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("draft 用户")
	tt.Command("selectAll")
	tt.Command("copy")
	if tt.Clipboard() != "draft 用户" {
		t.Fatal("composer copy", tt.Clipboard())
	}
	tt.Command("cut")
	if a.Draft != "" || tt.Clipboard() != "draft 用户" {
		t.Fatal("cut", a.Draft, tt.Clipboard())
	}
	tt.Command("paste")
	if a.Draft != "draft 用户" {
		t.Fatal("paste", a.Draft)
	}
	tt.Key(0, ui.KeyRight)
	tt.SetClipboard("sentinel")
	tt.Command("copy")
	if tt.Clipboard() != "sentinel" {
		t.Fatal("empty selection copied", tt.Clipboard())
	}
	selectTranscript()
	a.SetSnapshot(Snapshot{SessionID: "other", Messages: []Message{{ID: "b", Role: "assistant", Text: "replacement"}}})
	tt.Frame()
	tt.SetClipboard("owner sentinel")
	tt.Command("copy")
	if tt.Clipboard() != "owner sentinel" {
		t.Fatal("old owner copy", tt.Clipboard())
	}
}
