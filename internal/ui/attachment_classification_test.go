package orbitui

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Rust app.rs Attachment::from_path rejects these suffixes before reading;
// composer_ops.rs on_file_drop inserts the path at the active caret instead.
func TestAttachmentFileSVGAndTIFFInsertAtMountedCaret(t *testing.T) {
	for _, ext := range []string{"svg", "SVG", "tif", "TIFF"} {
		t.Run(ext, func(t *testing.T) {
			a := fixture(t)
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			tt.Type("left右")
			tt.Key(ui.Shift, ui.KeyLeft) // active caret precedes selected suffix; do not replace it
			path := filepath.Join(t.TempDir(), "文 件."+ext)
			if err := os.WriteFile(path, []byte("non-raster fixture"), 0600); err != nil {
				t.Fatal(err)
			}
			a.AddPaths([]string{path})
			tt.Frame()
			if want := "left" + path + " 右"; a.Draft != want {
				t.Fatalf("file should insert path at active caret: got %q want %q", a.Draft, want)
			}
			if len(a.attachments) != 0 {
				t.Fatal("unsupported file became image attachment")
			}
			tt.Key(ui.Cmd, ui.KeyZ)
			if a.Draft != "left右" {
				t.Fatalf("path insertion undo: %q", a.Draft)
			}
		})
	}
}

func drainClassificationAttachments(t *testing.T, a *App, tt *ui.Tester, count int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for len(a.attachments) < count && time.Now().Before(deadline) {
		select {
		case apply := <-a.updates:
			apply()
		case <-time.After(time.Millisecond):
		}
		tt.Frame()
	}
	if len(a.attachments) != count {
		t.Fatalf("attachments=%d want %d", len(a.attachments), count)
	}
}

func TestAttachmentFilesRetainSixSourceImageSuffixes(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1200, 800)
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	paths := []string{"a.png", "b.JPG", "c.jpeg", "d.webp", "e.gif", "f.bmp"}
	a.addPaths(paths, func(string) ([]byte, string, error) { return raw.Bytes(), "image/png", nil })
	drainClassificationAttachments(t, a, tt, len(paths))
	if a.Draft != "" || len(a.pendingAttachmentPaths) != 0 {
		t.Fatal("supported image became a path")
	}
	names := map[string]bool{}
	for _, im := range a.attachments {
		names[im.Name] = true
	}
	for _, p := range paths {
		if !names[p] {
			t.Fatalf("missing %s", p)
		}
	}
}

func TestAttachmentClipboardSVGAndTIFFRemainPassThrough(t *testing.T) {
	// Clipboard enters AddImage directly, not the file-suffix classifier.
	for _, mime := range []string{"image/svg+xml", "image/tiff"} {
		t.Run(mime, func(t *testing.T) {
			a := fixture(t)
			tt := ui.NewTester(a.View, 1200, 800)
			raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"><rect width="1" height="1"/></svg>`)
			if mime == "image/tiff" {
				raw = []byte{'I', 'I', 42, 0, 0, 0, 0, 0}
			}
			a.AddImage(raw, mime, "clipboard")
			drainClassificationAttachments(t, a, tt, 1)
			im := a.attachments[0].Image
			if im.MIMEType != mime || im.Data != base64.StdEncoding.EncodeToString(raw) {
				t.Fatal("clipboard passthrough changed")
			}
			if a.Draft != "" {
				t.Fatal("clipboard inserted path")
			}
		})
	}
}
