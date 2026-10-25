package orbitui

import (
	"bytes"
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func assertFileStatus(t *testing.T, a *App, tt *ui.Tester, path, suffix string) {
	t.Helper()
	tt.Frame()
	p := a.files.states[path].CursorPos()
	label := fmt.Sprintf("Ln %d, Col %d · %s", p.Line+1, p.Col+1, suffix)
	if _, ok := tt.Find(label); !ok {
		t.Fatalf("missing real editor status %q", label)
	}
}
func TestMountedFileDiskLinesPreserveTrailingNewlineAndUnsavedUndo(t *testing.T) {
	original := "a\nb\nc\nd\ne\n"
	a, tt, path := mountedFile(t, original)
	if _, ok := tt.Find("5 lines"); !ok {
		t.Fatal("disk metadata missing source-equivalent 5 lines")
	}
	if a.files.states[path].Buffer.LineCount() != 6 {
		t.Fatal("editable trailing line was removed")
	}
	assertFileStatus(t, a, tt, path, "Saved")
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("new\nsecond\nthird\n")
	tt.Frame()
	d, _ := a.files.editor.Active()
	if !d.Dirty || d.Saving {
		t.Fatal("expected genuine debounce-pending edit", d.Dirty, d.Saving)
	}
	if _, ok := tt.Find("5 lines"); !ok {
		t.Fatal("unsaved buffer replaced disk metadata")
	}
	assertFileStatus(t, a, tt, path, "Unsaved")
	bytes, e := os.ReadFile(path)
	if e != nil || string(bytes) != original {
		t.Fatal("metadata display changed file bytes", string(bytes), e)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	tt.Frame()
	if a.files.states[path].Buffer.String() != original {
		t.Fatal("undo lost original newline")
	}
	assertFileStatus(t, a, tt, path, "Saved")
	tt.Key(ui.Cmd|ui.Shift, ui.KeyZ)
	tt.Frame()
	if err := a.files.editor.Save(path); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("3 lines"); !ok {
		t.Fatal("successful save did not update disk metadata")
	}
	assertFileStatus(t, a, tt, path, "Saved")
	bytes, e = os.ReadFile(path)
	if e != nil || string(bytes) != "new\nsecond\nthird\n" {
		t.Fatal("save altered terminal newline", string(bytes), e)
	}
}

func TestMountedFileSaveFailureNeverClaimsSaving(t *testing.T) {
	a, tt, path := mountedFile(t, "saved\n")
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("unsaved\n")
	tt.Frame()
	backup := path + ".original"
	if err := os.Rename(path, backup); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(path); os.Rename(backup, path) })
	if err := a.files.editor.Save(path); err == nil {
		t.Fatal("fixture save unexpectedly succeeded")
	}
	d, _ := a.files.editor.Active()
	if !d.Dirty || d.Saving || d.LastError == "" {
		t.Fatal("fixture not failed dirty state", d)
	}
	assertFileStatus(t, a, tt, path, "Save failed: "+d.LastError)
}

func TestMountedFileConflictShowsDiskChangeNotSaving(t *testing.T) {
	a, tt, path := mountedFile(t, "saved\n")
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("my draft\n")
	tt.Frame()
	if err := os.WriteFile(path, []byte("external\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.files.editor.Reload(path); err != nil {
		t.Fatal(err)
	}
	d, _ := a.files.editor.Active()
	if !d.Conflict || !d.Dirty || d.Text != "my draft\n" {
		t.Fatal("fixture did not retain conflicting draft", d)
	}
	assertFileStatus(t, a, tt, path, "Changed on disk")
}

func TestFileStatusMatchesDocumentFlags(t *testing.T) {
	a := fixture(t)
	base := workspace.Document{Content: workspace.FileContent{Mode: workspace.FileText}}
	for _, tc := range []struct {
		name, want string
		change     func(*workspace.Document)
	}{
		{"clean", "Saved", func(*workspace.Document) {}},
		{"pending", "Unsaved", func(d *workspace.Document) { d.Dirty = true }},
		{"inflight", "Saving…", func(d *workspace.Document) { d.Dirty = true; d.Saving = true }},
		{"failed", "Save failed: raw disk failure", func(d *workspace.Document) { d.Dirty = true; d.LastError = "raw disk failure" }},
		{"retry", "Saving…", func(d *workspace.Document) { d.Dirty = true; d.Saving = true; d.LastError = "old attempt" }},
		{"conflict", "Changed on disk", func(d *workspace.Document) { d.Dirty = true; d.Conflict = true; d.LastError = "old attempt" }},
		{"readonly", "Read-only", func(d *workspace.Document) { d.Content.Lossy = true; d.Conflict = true; d.Dirty = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := base
			tc.change(&d)
			if got := a.fileSaveStatus(d); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestMountedImageHeaderRetainsRustZeroLineMetadata(t *testing.T) {
	a, tt, path := mountedFile(t, "temporary text fixture")
	im := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			im.SetNRGBA(x, y, color.NRGBA{R: 231, G: 17, B: 219, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, im); err != nil {
		t.Fatal(err)
	}
	imagePath := filepath.Join(filepath.Dir(path), "raster.png")
	if err := os.WriteFile(imagePath, encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := a.files.editor.Open(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	if d.Content.Mode != workspace.FileImage || len(d.Content.Lines) != 0 {
		t.Fatal("fixture is not image content")
	}
	a.syncDocument(d)
	tt.Frame()
	if _, ok := tt.Find("0 lines"); !ok {
		t.Fatal("Rust image toolbar's zero-line metadata not rendered")
	}
	rendered := tt.Image()
	pixels := 0
	for y := rendered.Bounds().Min.Y; y < rendered.Bounds().Max.Y; y++ {
		for x := rendered.Bounds().Min.X; x < rendered.Bounds().Max.X; x++ {
			r, g, b, _ := rendered.At(x, y).RGBA()
			if r>>8 == 231 && g>>8 == 17 && b>>8 == 219 {
				pixels++
			}
		}
	}
	if pixels < 100 {
		t.Fatal("image body was not actually rendered", pixels)
	}
	got, err := os.ReadFile(imagePath)
	if err != nil || !bytes.Equal(got, encoded.Bytes()) {
		t.Fatal("render changed PNG bytes", err)
	}
}
