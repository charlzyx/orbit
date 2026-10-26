package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestMessageMenuSameOwnerStreamUsesOpeningSnapshot(t *testing.T) {
	a := fixture(t)
	s := menuAuditSnapshot()
	s.Messages = []Message{{ID: "stream", Role: "assistant", Text: "OPENING_MESSAGE", Streaming: true}}
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1470, 922)
	if err := tt.RightClick("OPENING_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	s.Messages = []Message{{ID: "stream", Role: "assistant", Text: "OPENING_MESSAGE APPENDED_MESSAGE", Streaming: true}}
	a.SetSnapshot(s)
	tt.Frame()
	tt.Move(1460, 10)
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "OPENING_MESSAGE" {
		t.Fatal("selection-time retarget", tt.Clipboard())
	}
	if err := tt.RightClick("OPENING_MESSAGE APPENDED_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "OPENING_MESSAGE APPENDED_MESSAGE" {
		t.Fatal("reopen kept stale snapshot", tt.Clipboard())
	}
}
