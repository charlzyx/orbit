package orbitui

import (
	"bytes"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"image"
	"image/png"
	"testing"
	"time"
)

func TestAttachmentPathsNativeCaretSelectionUndo(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			a := fixture(t)
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			tt.Type("甲😀乙尾")
			tt.Key(0, ui.KeyLeft)
			if reverse {
				tt.Key(ui.Shift, ui.KeyLeft)
			} else {
				tt.Key(0, ui.KeyLeft)
				tt.Key(ui.Shift, ui.KeyRight)
			}
			a.AddPaths([]string{"/tmp/文 件.txt", "/tmp/two.md"})
			tt.Frame()
			want := "甲😀乙/tmp/文 件.txt /tmp/two.md 尾"
			if reverse {
				want = "甲😀/tmp/文 件.txt /tmp/two.md 乙尾"
			}
			if a.Draft != want {
				t.Fatalf("insert: %q want %q", a.Draft, want)
			}
			tt.Key(ui.Cmd, ui.KeyZ)
			first := "甲😀乙/tmp/文 件.txt 尾"
			if reverse {
				first = "甲😀/tmp/文 件.txt 乙尾"
			}
			if a.Draft != first {
				t.Fatalf("first undo: %q", a.Draft)
			}
			tt.Key(ui.Cmd, ui.KeyZ)
			if a.Draft != "甲😀乙尾" {
				t.Fatalf("second undo: %q", a.Draft)
			}
			tt.Key(ui.Cmd|ui.Shift, ui.KeyZ)
			if a.Draft != first {
				t.Fatalf("redo: %q", a.Draft)
			}
		})
	}
}
func TestAttachmentPathsQueuedEnterAndIME(t *testing.T) {
	a := fixture(t)
	var sent []string
	a.Options.OnAction = func(x Action) {
		if x.Kind == "Submit" {
			sent = append(sent, x.Value)
		}
	}
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("headtail")
	for range 4 {
		tt.Key(0, ui.KeyLeft)
	}
	a.AddPaths([]string{"/tmp/a.txt"})
	tt.Key(0, ui.KeyEnter)
	if len(sent) != 1 || sent[0] != "head/tmp/a.txt tail" {
		t.Fatalf("queued submit: %#v draft=%q", sent, a.Draft)
	}
	tt.Type("before")
	tt.Compose("你", 1)
	a.AddPaths([]string{"/tmp/中.txt"})
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if len(sent) != 1 || len(a.pendingAttachmentPaths) != 1 || a.Draft != "before" {
		t.Fatalf("IME altered: %#v %q pending=%d", sent, a.Draft, len(a.pendingAttachmentPaths))
	}
	tt.Type("你")
	tt.Frame()
	if a.Draft != "before你/tmp/中.txt " || len(a.pendingAttachmentPaths) != 0 {
		t.Fatalf("after composition: %q", a.Draft)
	}
}
func TestAttachmentPathsDropStaleOwnerBeforeMount(t *testing.T) {
	a := fixture(t)
	a.AddPaths([]string{"/tmp/old.txt"})
	a.Model.Workspace = "/different"
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if a.Draft != "" || len(a.pendingAttachmentPaths) != 0 {
		t.Fatal("stale path reached new draft")
	}
}

func TestAttachmentPathsManagedHydrationAndIntent(t *testing.T) {
	for _, change := range []string{"hydrate", "generation", "close", "window"} {
		t.Run(change, func(t *testing.T) {
			a := fixture(t)
			generation := uint64(1)
			live := true
			a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, generation, live }
			a.AddPaths([]string{"/tmp/kept.txt"})
			switch change {
			case "hydrate":
				a.Model.RuntimeID = "r"
				a.Model.SessionID = "s"
				a.Model.SessionPath = "p"
			case "generation":
				generation++
			case "close":
				live = false
			case "window":
				a.Window = new(mygo.Window)
			}
			tt := ui.NewTester(a.View, 1200, 800)
			tt.Frame()
			want := ""
			if change == "hydrate" {
				want = "/tmp/kept.txt "
			}
			if a.Draft != want || len(a.pendingAttachmentPaths) != 0 {
				t.Fatalf("%q pending %d", a.Draft, len(a.pendingAttachmentPaths))
			}
			a.Window = nil
		})
	}
}
func TestAttachmentPathsPendingCannotSubmitDirectly(t *testing.T) {
	a := fixture(t)
	a.Draft = "must wait"
	a.AddPaths([]string{"/tmp/queued.txt"})
	called := false
	a.Options.OnAction = func(Action) { called = true }
	if a.submit(a.Draft, "") || called || a.Draft != "must wait" {
		t.Fatal("pending edit submitted old draft")
	}
}

func TestAttachmentPathsMixedImageAndText(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("left右")
	tt.Key(0, ui.KeyLeft)
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	read := make(chan string, 1)
	a.addPaths([]string{"/tmp/one.txt", "/tmp/picture.png", "/tmp/二.md"}, func(p string) ([]byte, string, error) { read <- p; return b.Bytes(), "image/png", nil })
	tt.Frame()
	if a.Draft != "left/tmp/one.txt /tmp/二.md 右" {
		t.Fatal(a.Draft)
	}
	select {
	case p := <-read:
		if p != "/tmp/picture.png" {
			t.Fatal(p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("image not read")
	}
	deadline := time.Now().Add(3 * time.Second)
	for len(a.attachments) == 0 && time.Now().Before(deadline) {
		select {
		case apply := <-a.updates:
			apply()
		case <-time.After(time.Millisecond):
		}
		tt.Frame()
	}
	if len(a.attachments) != 1 || a.attachments[0].Name != "picture.png" {
		t.Fatalf("mixed image lost: %+v", a.attachments)
	}
	tt.Type("!")
	if a.Draft != "left/tmp/one.txt /tmp/二.md !右" {
		t.Fatalf("focus/caret not retained: %q", a.Draft)
	}
}

func TestAttachmentPathsPrunedWithHiddenComposer(t *testing.T) {
	a := fixture(t)
	generation := uint64(1)
	a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, generation, true }
	a.AddPaths([]string{"/tmp/obsolete.txt"})
	generation++
	a.SetSnapshot(a.Model)
	if len(a.pendingAttachmentPaths) != 0 {
		t.Fatal("snapshot kept obsolete intent")
	}
	a.AddPaths([]string{"/tmp/also-obsolete.txt"})
	generation++
	a.Navigate("settings")
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if len(a.pendingAttachmentPaths) != 0 || a.Draft != "" {
		t.Fatal("hidden composer kept obsolete intent")
	}
}
