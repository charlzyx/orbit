package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestAuditJumpAffordanceDoesNotResizeTranscriptViewport(t *testing.T) {
	a := fixture(t)
	s := Snapshot{RuntimeID: "geometry", SessionID: "geometry"}
	for i := 0; i < 120; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Geometry row %03d", i)})
	}
	a.SetSnapshot(s)
	var bounds ui.Rect
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); bounds = a.transcript.Handle.Bounds(c) }, 1200, 800)
	tt.Frame()
	tt.Frame()
	before := bounds
	if !a.transcript.AtEnd() || before.H <= 0 {
		t.Fatalf("not initialized tail: %v", before)
	}
	for i := 0; i < 6; i++ {
		tt.Scroll(700, 400, 0, -30)
	}
	tt.Frame()
	tt.Frame()
	after := bounds
	if a.transcript.AtEnd() {
		t.Fatal("wheel did not leave tail")
	}
	button, ok := tt.Find("Jump to latest")
	if !ok {
		t.Fatal("jump affordance absent")
	}
	first, last := a.transcript.Visible()
	row, found := tt.Find(s.Messages[first+1].Text)
	t.Logf("tail viewport=%+v; off-tail viewport=%+v; button=%+v; visible=%d..%d; anchorFound=%t anchorY=%g; scrollY=%g maxY=%g", before, after, button, first, last, found, row.Y, a.transcriptScroll.Y, a.transcriptScroll.MaxY)
	// Rust message_scroller.rs viewport is h_full and the jump layer is absolute
	// bottom8; showing the affordance must not consume viewport layout height.
	if after.H != before.H {
		t.Errorf("jump row consumed viewport height: %g -> %g (delta %g)", before.H, after.H, before.H-after.H)
	}
	if button.H != 32 {
		t.Errorf("non-unread source Large button height: got %g want 32", button.H)
	}
}
