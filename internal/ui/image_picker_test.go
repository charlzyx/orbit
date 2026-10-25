package orbitui

import (
	"bytes"
	"errors"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestImagePickerAdmissionLimitAndMenu(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.attachments = make([]attachment, 8)
	for i := range x.UI.attachments {
		x.UI.attachments[i].ID = strconv.Itoa(i)
		x.UI.attachments[i].Name = "fixture " + strconv.Itoa(i)
	}
	called := false
	x.fileDialogs.image = func(*mygo.Window, string) ([]string, error) { called = true; return nil, nil }
	tt := ui.NewTester(x.UI.View, 1200, 800)
	if err := tt.Click("Add attachment"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(x.UI.text("composer.attach_image", "Attach image…")); err != nil {
		t.Fatal(err)
	}
	if called || x.fileDialogs.pending || len(x.UI.attachments) != 8 {
		t.Fatal("full attachment list opened picker")
	}
	if len(x.UI.toastStack.Items()) == 0 {
		t.Fatal("limit has no feedback")
	}
}
func TestImagePickerAsyncSingleImageAndSingleton(t *testing.T) {
	x := workspaceDialogFixture(t)
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "一 张.png")
	if err := os.WriteFile(path, raw.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	release := holdFileDialog(t, x, Action{Kind: "BrowseImage"}, []string{path, "/ignored.txt"}, nil)
	x.Action(Action{Kind: "BrowseImage"})
	x.Action(Action{Kind: "BrowseAttachments"})
	x.Action(Action{Kind: "BrowseWorkspace"})
	if x.workspaceBrowse.pending || len(x.UI.attachments) > 0 {
		t.Fatal("stacked native panel or premature attachment")
	}
	release()
	drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending && len(x.UI.attachments) == 1 })
	if x.UI.attachments[0].Name != "一 张.png" || x.UI.Draft != "" || len(x.UI.pendingAttachmentPaths) != 0 {
		t.Fatal("image picker became prompt text/multi-selection")
	}
}
func TestImagePickerCancelErrorUnsupportedAndReopen(t *testing.T) {
	for _, kind := range []string{"cancel", "error", "unsupported", "svg", "tiff"} {
		t.Run(kind, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			var err error
			var paths []string
			if kind == "error" {
				err = errors.New("synthetic image picker error")
			}
			if kind == "unsupported" {
				paths = []string{"/not-an-image.txt"}
			}
			if kind == "svg" || kind == "tiff" {
				paths = []string{"/not-file-attachable." + kind, "/ignored.png"}
			}
			release := holdFileDialog(t, x, Action{Kind: "BrowseImage"}, paths, err)
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
			if len(x.UI.attachments) != 0 || x.UI.Draft != "" || len(x.UI.pendingAttachmentPaths) != 0 {
				t.Fatal("unsuccessful image selection altered draft")
			}
			if kind != "cancel" && len(x.UI.toastStack.Items()) == 0 {
				t.Fatal("error missing")
			}
			if kind == "unsupported" || kind == "svg" || kind == "tiff" {
				found := false
				for _, toast := range x.UI.toastStack.Items() {
					found = found || toast.Body == "Unsupported image format"
				}
				if !found {
					t.Fatal("unexpected first image format was not explicitly rejected")
				}
			}
			release = holdFileDialog(t, x, Action{Kind: "BrowseImage"}, nil, nil)
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
		})
	}
}
func TestImagePickerStaleTaskWindowClose(t *testing.T) {
	for _, kind := range []string{"task", "window", "close"} {
		t.Run(kind, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			release := holdFileDialog(t, x, Action{Kind: "BrowseImage"}, []string{"/obsolete.png"}, errors.New("obsolete error"))
			switch kind {
			case "task":
				x.Action(Action{Kind: "NewSession"})
			case "window":
				x.UI.Window = new(mygo.Window)
			case "close": // Close after the completion has entered the mailbox.
			}
			release()
			apply := takeControllerDialogUpdate(t, x)
			if kind == "close" {
				x.Close()
			}
			apply()
			if len(x.UI.attachments) != 0 || len(x.UI.pendingAttachmentPaths) != 0 || x.fileDialogs.pending {
				t.Fatal("stale image selection applied")
			}
			for _, toast := range x.UI.toastStack.Items() {
				if toast.Body == "obsolete error" {
					t.Fatal("stale picker error leaked")
				}
			}
			x.UI.Window = nil
		})
	}
}

func TestImagePickerLimitRecheckedAfterSelection(t *testing.T) {
	x := workspaceDialogFixture(t)
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "late.png")
	if err := os.WriteFile(path, raw.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	release := holdFileDialog(t, x, Action{Kind: "BrowseImage"}, []string{path}, nil)
	x.UI.attachments = make([]attachment, 8)
	release()
	drainControllerDialogUpdates(t, x, func() bool {
		for _, toast := range x.UI.toastStack.Items() {
			if toast.Body == "Maximum 8 attachments per message" {
				return true
			}
		}
		return false
	})
	if len(x.UI.attachments) != 8 {
		t.Fatal("late picker exceeded attachment cap")
	}
}

func TestImagePickerPathGuardMatchesRustFromPath(t *testing.T) {
	for _, ext := range []string{"png", "jpg", "jpeg", "webp", "gif", "bmp", "PNG", "JpEg"} {
		if !imagePickerPathSupported("/file." + ext) {
			t.Fatal("rejected source format", ext)
		}
	}
	for _, path := range []string{"/file.svg", "/file.tiff", "/file.tif", "/file.txt", "/folder"} {
		if imagePickerPathSupported(path) {
			t.Fatal("accepted non-source file format", path)
		}
	}
}
