package orbitui

import (
	"bytes"
	"encoding/base64"
	"github.com/egoist/mygo/ui"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestInlineImageContentAuditSameSlotReplacesRealBitmap(t *testing.T) {
	msg := inlineContentAuditMessage(t, color.RGBA{R: 255, A: 255})
	a, tt := inlineAuditApp(t, &msg, "same-session")
	inlineAuditDrainReady(t, a, tt)
	old := a.messageImages["same-message:0"]
	oldLength := len(msg.Images[0].Data)
	msg = inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
	if len(msg.Images[0].Data) != oldLength {
		t.Fatal("fixture lengths differ")
	}
	s := a.Model
	s.Messages = []Message{msg}
	a.SetSnapshot(s)
	tt.Frame()
	inlineAuditDrainReady(t, a, tt)
	hit, ok := tt.Find("Open image attachment")
	if !ok {
		t.Fatal("updated image missing")
	}
	red, green, _, _ := tt.Image().At(int(hit.X+hit.W/2), int(hit.Y+hit.H/2)).RGBA()
	if red>>8 > 80 || green>>8 < 200 || a.messageImages["same-message:0"] == old {
		t.Fatal("same ID:index still renders the old red payload, not replacement green")
	}
}
func TestInlineImageContentAuditInvalidOrEmptyCanBecomeComplete(t *testing.T) {
	for _, initial := range []string{"empty", "invalid", "mime-only"} {
		t.Run(initial, func(t *testing.T) {
			valid := inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
			msg := valid
			msg.Images = append(msg.Images[:0:0], msg.Images...)
			switch initial {
			case "empty":
				msg.Images[0].Data = ""
			case "invalid":
				msg.Images[0].Data = "not-base64"
			case "mime-only":
				msg.Images[0].MIMEType = "image/svg+xml"
			}
			a, tt := inlineAuditApp(t, &msg, "same-session")
			inlineAuditCompletion(t, a)()
			tt.Frame()
			if a.imageErrors["same-message:0"] == "" {
				t.Fatal("fixture failed to produce actual decode error")
			}
			msg = valid
			s := a.Model
			s.Messages = []Message{msg}
			a.SetSnapshot(s)
			tt.Frame()
			if a.imageErrors["same-message:0"] != "" {
				t.Fatal("previous payload error remains sticky after content becomes valid")
			}
			inlineAuditDrainReady(t, a, tt)
		})
	}
}
func TestInlineImageContentAuditHeldDecodeCannotOverwriteReplacement(t *testing.T) {
	msg := inlineContentAuditMessage(t, color.RGBA{R: 255, A: 255})
	a, tt := inlineAuditApp(t, &msg, "same-session")
	oldCompletion := inlineAuditCompletion(t, a)
	msg = inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
	s := a.Model
	s.Messages = []Message{msg}
	a.SetSnapshot(s)
	tt.Frame()
	oldCompletion()
	if a.messageImages["same-message:0"] != nil {
		t.Fatal("old payload's real held decode published after same-slot replacement")
	}
	inlineAuditDrainReady(t, a, tt)
}

func TestInlineImageContentAuditHiddenABAAndRemoval(t *testing.T) {
	for _, kind := range []string{"payload-ABA", "remove-readd"} {
		t.Run(kind, func(t *testing.T) {
			original := inlineContentAuditMessage(t, color.RGBA{R: 255, A: 255})
			msg := original
			a, tt := inlineAuditApp(t, &msg, "same-session")
			held := inlineAuditCompletion(t, a)
			s := a.Model
			if kind == "payload-ABA" {
				s.Messages = []Message{inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})}
			} else {
				s.Messages = nil
			}
			a.SetSnapshot(s) // no frame while content is B/removed
			s.Messages = []Message{original}
			a.SetSnapshot(s)
			held()
			if len(a.messageImages) != 0 || len(a.imageErrors) != 0 {
				t.Fatal("hidden content ABA accepted old completion")
			}
			tt.Frame()
			inlineAuditDrainReady(t, a, tt)
		})
	}
}
func TestInlineImageContentAuditOldErrorDoesNotOverwriteFreshSuccess(t *testing.T) {
	valid := inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
	msg := valid
	msg.Images = append(msg.Images[:0:0], msg.Images...)
	msg.Images[0].Data = "invalid"
	a, tt := inlineAuditApp(t, &msg, "same-session")
	held := inlineAuditCompletion(t, a)
	msg = valid
	s := a.Model
	s.Messages = []Message{msg}
	a.SetSnapshot(s)
	tt.Frame()
	inlineAuditDrainReady(t, a, tt)
	fresh := a.messageImages["same-message:0"]
	held()
	if a.imageErrors["same-message:0"] != "" || a.messageImages["same-message:0"] != fresh {
		t.Fatal("old error overwrote valid content")
	}
}
func TestInlineImageContentAuditUnchangedStreamingSnapshotKeepsCache(t *testing.T) {
	msg := inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
	a, tt := inlineAuditApp(t, &msg, "same-session")
	inlineAuditDrainReady(t, a, tt)
	original := a.messageImages["same-message:0"]
	revision := a.inlineImages.nextContentRevision
	for i := range 10 {
		msg.Images[0].Data = strings.Clone(msg.Images[0].Data) // equal value, distinct backing bytes
		msg.Streaming = i%2 == 0
		msg.Text = "text-only update"
		s := a.Model
		s.Messages = []Message{msg}
		a.SetSnapshot(s)
		tt.Frame()
	}
	if a.inlineImages.nextContentRevision != revision || a.messageImages["same-message:0"] != original || a.imageLoading["same-message:0"] {
		t.Fatal("unchanged image was decoded again for streaming text snapshots")
	}
	if len(a.updates) != 0 {
		t.Fatal("unchanged payload queued unexpected worker completion")
	}
}
func TestInlineImageContentAuditHeldPointerAndPreviewInvalidation(t *testing.T) {
	msg := inlineContentAuditMessage(t, color.RGBA{R: 255, A: 255})
	a, tt := inlineAuditApp(t, &msg, "same-session")
	inlineAuditDrainReady(t, a, tt)
	hit, _ := tt.Find("Open image attachment")
	px, py := hit.X+hit.W/2, hit.Y+hit.H/2
	tt.Press(px, py)
	msg = inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
	s := a.Model
	s.Messages = []Message{msg}
	a.SetSnapshot(s)
	tt.Frame()
	inlineAuditDrainReady(t, a, tt)
	tt.Release(px, py)
	tt.Frame()
	if a.imageOpen {
		t.Fatal("old-content pointer opened replacement preview")
	}
	if err := tt.Click("Open image attachment"); err != nil {
		t.Fatal(err)
	}
	if !a.imageOpen {
		t.Fatal("fresh content preview did not open")
	}
	same := a.Model
	a.SetSnapshot(same)
	if !a.imageOpen {
		t.Fatal("unchanged payload closed preview")
	}
	msg = inlineContentAuditMessage(t, color.RGBA{R: 255, A: 255})
	s.Messages = []Message{msg}
	a.SetSnapshot(s)
	if a.imageOpen || a.imagePreview != nil {
		t.Fatal("changed inline payload retained old preview")
	}
}

// Uncompressed PNG keeps differently colored fixtures exactly the same length,
// so identity cannot be reduced to MIME/size or a prefix.
func inlineContentAuditMessage(t *testing.T, shade color.RGBA) Message {
	t.Helper()
	msg := inlineAuditMessage(t, shade)
	bitmap := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for y := range 24 {
		for x := range 24 {
			bitmap.SetRGBA(x, y, shade)
		}
	}
	var raw bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(&raw, bitmap); err != nil {
		t.Fatal(err)
	}
	msg.Images[0].Data = base64.StdEncoding.EncodeToString(raw.Bytes())
	return msg
}

func TestInlineImageContentAuditDoesNotCloseMarkdownPreview(t *testing.T) {
	a := fixture(t)
	t.Cleanup(a.markdownImages.Close)
	msg := inlineContentAuditMessage(t, color.RGBA{R: 255, A: 255})
	md := Message{ID: "markdown", Role: "assistant", Text: "![independent preview](" + markdownDataFixture(t) + ")"}
	s := Snapshot{RuntimeID: "runtime", SessionID: "same", Messages: []Message{msg, md}}
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1200, 800)
	inlineAuditDrainReady(t, a, tt)
	waitMarkdownRaster(t, tt, "independent preview")
	if err := tt.Click("Open image attachment"); err != nil {
		t.Fatal(err)
	}
	if !a.imageOpen || a.inlineImages.preview.revision == 0 {
		t.Fatal("inline preview provenance missing")
	}
	if err := tt.Click("Close image"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Open Markdown image independent preview"); err != nil {
		t.Fatal(err)
	}
	markdownPreview := a.imagePreview
	if !a.imageOpen || markdownPreview == nil || a.inlineImages.preview.revision != 0 {
		t.Fatal("markdown retained inline preview provenance")
	}
	msg = inlineContentAuditMessage(t, color.RGBA{G: 255, A: 255})
	s.Messages = []Message{msg, md}
	a.SetSnapshot(s)
	if !a.imageOpen || a.imagePreview != markdownPreview {
		t.Fatal("inline replacement closed unrelated markdown preview")
	}
	s.Messages = []Message{md}
	a.SetSnapshot(s)
	if len(a.inlineImages.contents) != 0 || len(a.messageImages) != 0 || len(a.imageErrors) != 0 || len(a.imageLoading) != 0 {
		t.Fatal("removed slot references were retained")
	}
	if !a.imageOpen || a.imagePreview != markdownPreview {
		t.Fatal("inline removal closed unrelated markdown preview")
	}
}
