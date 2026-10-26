package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestThoughtScrollActualAppOwnerChangesRejectCapturedCallback(t *testing.T) {
	for _, change := range []string{"session", "runtime", "roundtrip", "closed"} {
		t.Run(change, func(t *testing.T) {
			a := fixture(t)
			s := Snapshot{RuntimeID: "old-runtime", SessionID: "old-session", Workspace: "/synthetic-a"}
			a.SetSnapshot(s)
			var fn func(float32) bool
			tt := ui.NewTester(func(c *ui.Context) {
				a.View(c)
				if fn == nil {
					fn = a.transcriptOptions(c).ThoughtScroll
				}
			}, 1200, 800)
			tt.Frame()
			if fn == nil || !fn(0) {
				t.Fatal("current owner must accept")
			}
			switch change {
			case "session":
				s.SessionID = "new-session"
				a.SetSnapshot(s)
			case "runtime":
				s.RuntimeID = "new-runtime"
				a.SetSnapshot(s)
			case "roundtrip":
				old := s
				s.SessionID = "other"
				a.SetSnapshot(s)
				a.SetSnapshot(old)
			case "closed":
				a.closeInlineImages()
			}
			y := a.transcriptScroll.Y
			unread := a.transcriptActivity.unread
			if fn(0) || fn(100000) || a.transcriptScroll.Y != y || a.transcriptActivity.unread != unread {
				t.Fatal("stale callback changed new owner")
			}
		})
	}
}
