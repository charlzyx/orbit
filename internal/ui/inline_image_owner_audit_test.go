package orbitui

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func inlineAuditMessage(t *testing.T, shade color.RGBA) Message {
	t.Helper()
	bitmap := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for y := range 24 {
		for x := range 24 {
			bitmap.SetRGBA(x, y, shade)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, bitmap); err != nil {
		t.Fatal(err)
	}
	return Message{ID: "same-message", Role: "assistant", Images: []agent.Image{{Type: "image", MIMEType: "image/png", Data: base64.StdEncoding.EncodeToString(b.Bytes())}}}
}
func inlineAuditCompletion(t *testing.T, a *App) func() {
	t.Helper()
	select {
	case apply := <-a.updates:
		return apply
	case <-time.After(3 * time.Second):
		t.Fatal("real inline decode did not complete")
		return nil
	}
}
func inlineAuditApp(t *testing.T, msg *Message, sessionID string) (*App, *ui.Tester) {
	t.Helper()
	a := fixture(t)
	// Finish the unrelated constructor background worker before mounting images.
	for a.backdrop.loading {
		inlineAuditCompletion(t, a)()
	}
	path := "/session-A"
	if sessionID == "" {
		path = ""
	}
	a.SetSnapshot(Snapshot{RuntimeID: "runtime-A", SessionID: sessionID, SessionPath: path, Workspace: "/workspace-A", Messages: []Message{*msg}})
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() { a.messageImageRow(c, *msg); a.imageDialog(c) })
	}, 400, 300)
	return a, tt
}
func TestInlineImageAuditLateRealDecodeCannotCrossOwner(t *testing.T) {
	for _, kind := range []string{"runtime", "workspace", "path", "empty-draft", "roundtrip"} {
		t.Run(kind, func(t *testing.T) {
			msg := inlineAuditMessage(t, color.RGBA{R: 255, A: 255})
			id := "same-session"
			if kind == "empty-draft" {
				id = ""
			}
			a, tt := inlineAuditApp(t, &msg, id)
			_ = tt
			apply := inlineAuditCompletion(t, a) // hold the actual decoder's callback, not a stub result
			s := a.Model
			switch kind {
			case "runtime":
				s.RuntimeID = "runtime-B"
			case "workspace":
				s.Workspace = "/workspace-B"
			case "path":
				s.SessionPath = "/session-B"
			case "empty-draft":
				// Keep both before/after IDs empty while retaining distinct runtime owners.
				s.RuntimeID = "runtime-B"
			case "roundtrip":
				s.RuntimeID = "runtime-B"
			}
			original := a.Model
			a.SetSnapshot(s)
			if kind == "roundtrip" {
				a.SetSnapshot(original)
			}
			apply()
			if len(a.messageImages) != 0 || len(a.imageErrors) != 0 {
				t.Fatal("held old-owner real decode published into current inline cache")
			}
		})
	}
}
func TestInlineImageAuditCachedBitmapCannotReopenAcrossRuntime(t *testing.T) {
	msg := inlineAuditMessage(t, color.RGBA{R: 255, A: 255})
	a, tt := inlineAuditApp(t, &msg, "same-session")
	inlineAuditCompletion(t, a)()
	tt.Frame()
	if err := tt.Click("Open image attachment"); err != nil {
		t.Fatal(err)
	}
	old := a.imagePreview
	if old == nil {
		t.Fatal("real decoded old-owner preview not opened")
	}
	s := a.Model
	s.RuntimeID = "runtime-B"
	msg = inlineAuditMessage(t, color.RGBA{G: 255, A: 255})
	s.Messages = []Message{msg}
	a.SetSnapshot(s)
	if a.imageOpen {
		t.Fatal("existing markdown-owner reset should already close the old modal")
	}
	tt.Frame()
	// A correct cache can show loading here. The current code renders the old
	// bitmap immediately and allows its actual mounted button to reopen it.
	if _, ok := tt.Find("Open image attachment"); ok {
		if err := tt.Click("Open image attachment"); err != nil {
			t.Fatal(err)
		}
		if a.imagePreview == old {
			t.Fatal("new runtime reopened previous runtime's actual cached bitmap")
		}
	}
	a.imageOpen = false
	inlineAuditDrainReady(t, a, tt)
	hit, ok := tt.Find("Open image attachment")
	if !ok {
		t.Fatal("fresh owner's decoded image button missing")
	}
	red, green, _, _ := tt.Image().At(int(hit.X+hit.W/2), int(hit.Y+hit.H/2)).RGBA()
	if green>>8 < 200 || red>>8 > 80 {
		t.Fatal("new owner's green bitmap was not actually rendered")
	}
	if err := tt.Click("Open image attachment"); err != nil {
		t.Fatal(err)
	}
	if !a.imageOpen || a.imagePreview == old {
		t.Fatal("fresh new-owner preview unavailable")
	}
}

func inlineAuditDrainReady(t *testing.T, a *App, tt *ui.Tester) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for a.messageImages["same-message:0"] == nil && time.Now().Before(deadline) {
		select {
		case apply := <-a.updates:
			apply()
		case <-time.After(time.Millisecond):
		}
		tt.Frame()
	}
	if a.messageImages["same-message:0"] == nil {
		t.Fatal("fresh inline image did not decode")
	}
}
func TestInlineImageAuditHeldPointerAndFreshOwnerClick(t *testing.T) {
	msg := inlineAuditMessage(t, color.RGBA{R: 255, A: 255})
	a, tt := inlineAuditApp(t, &msg, "same-session")
	inlineAuditDrainReady(t, a, tt)
	hit, _ := tt.Find("Open image attachment")
	px, py := hit.X+hit.W/2, hit.Y+hit.H/2
	tt.Press(px, py)
	original := a.Model
	s := original
	s.RuntimeID = "runtime-B"
	a.SetSnapshot(s)
	a.SetSnapshot(original)
	tt.Frame()
	inlineAuditDrainReady(t, a, tt)
	tt.Release(px, py)
	tt.Frame()
	if a.imageOpen {
		t.Fatal("old-owner press survived epoch and opened preview")
	}
	if err := tt.Click("Open image attachment"); err != nil {
		t.Fatal(err)
	}
	if !a.imageOpen || a.imagePreview == nil {
		t.Fatal("fresh same-owner click did not open image")
	}
	a.SetSnapshot(original)
	if !a.imageOpen {
		t.Fatal("unchanged-owner snapshot closed valid preview")
	}
}
func TestInlineImageAuditLateDecodeErrorCannotCrossOwner(t *testing.T) {
	msg := Message{ID: "same-message", Images: []agent.Image{{MIMEType: "image/png", Data: "not-base64"}}}
	a, tt := inlineAuditApp(t, &msg, "")
	_ = tt
	apply := inlineAuditCompletion(t, a)
	s := a.Model
	s.RuntimeID = "runtime-B"
	a.SetSnapshot(s)
	apply()
	if len(a.imageErrors) != 0 || len(a.messageImages) != 0 {
		t.Fatal("old-owner error polluted new cache")
	}
	tt.Frame()
	inlineAuditCompletion(t, a)()
	if a.imageErrors["same-message:0"] == "" {
		t.Fatal("fresh current-owner error was suppressed")
	}
}
func TestInlineImageAuditServiceRoundtripAndActualTaskGeneration(t *testing.T) {
	for _, kind := range []string{"service-roundtrip", "new-task-generation", "close"} {
		t.Run(kind, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			a := x.UI
			for a.backdrop.loading {
				inlineAuditCompletion(t, a)()
			}
			// Snapshot deliberately stays empty-ID: task identity is supplied by the
			// real manager's opaque target, not a test-written generation field.
			a.SetSnapshot(Snapshot{RuntimeID: "same-runtime", Workspace: x.workspace})
			msg := inlineAuditMessage(t, color.RGBA{R: 255, A: 255})
			tt := ui.NewTester(func(c *ui.Context) { a.messageImageRow(c, msg) }, 400, 300)
			_ = tt
			apply := inlineAuditCompletion(t, a)
			switch kind {
			case "service-roundtrip":
				initial := x.workspace
				x.changeWorkspaceServices(t.TempDir())
				x.changeWorkspaceServices(initial)
			case "new-task-generation":
				old := x.Agent.CaptureWorkflowTarget()
				if err := x.Agent.Open(x.workspace, ""); err != nil {
					t.Fatal(err)
				}
				if old == x.Agent.CaptureWorkflowTarget() {
					t.Fatal("fixture did not create another actual task")
				}
			case "close":
				x.Close()
			}
			apply()
			if len(a.messageImages) != 0 || len(a.imageErrors) != 0 {
				t.Fatal("actual lifecycle allowed late inline completion")
			}
		})
	}
}
