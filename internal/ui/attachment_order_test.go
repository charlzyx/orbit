package orbitui

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func attachmentOrderPNG(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func attachmentOrderExisting(a *App, count int) {
	for i := 0; i < count; i++ {
		a.attachments = append(a.attachments, attachment{ID: fmt.Sprintf("existing-%d", i), Name: fmt.Sprintf("existing-%d", i)})
	}
}
func drainAttachmentOrder(t *testing.T, a *App) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for a.hasPendingAttachmentImages() && time.Now().Before(deadline) {
		select {
		case apply := <-a.updates:
			apply()
		case <-time.After(time.Millisecond):
		}
	}
	if a.hasPendingAttachmentImages() {
		t.Fatal("image work did not settle")
	}
}
func TestAttachmentDropOrderFollowsInputNotReadCompletion(t *testing.T) {
	a := fixture(t)
	raw := attachmentOrderPNG(t)
	slowStarted, fastStarted, releaseSlow := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(releaseSlow) }) }
	defer unblock()
	a.addPaths([]string{"slow.png", "fast.png"}, func(path string) ([]byte, string, error) {
		if path == "slow.png" {
			close(slowStarted)
			<-releaseSlow
		} else {
			close(fastStarted)
		}
		return raw, "image/png", nil
	})
	<-slowStarted
	// A serialized implementation cannot start fast before releasing slow. If
	// fast did start, force its actual owner completion before releasing slow.
	select {
	case <-fastStarted:
		deadline := time.Now().Add(2 * time.Second)
		for len(a.attachments) == 0 && time.Now().Before(deadline) {
			select {
			case apply := <-a.updates:
				apply()
			case <-time.After(time.Millisecond):
			}
		}
		if len(a.attachments) != 1 || a.attachments[0].Name != "fast.png" {
			t.Fatal("controlled fast completion did not settle")
		}
	case <-time.After(200 * time.Millisecond):
	}

	unblock()
	drainAttachmentOrder(t, a)
	var names []string
	for _, im := range a.attachments {
		names = append(names, im.Name)
	}
	if strings.Join(names, ",") != "slow.png,fast.png" {
		t.Fatalf("drop order followed completion instead of input: %v", names)
	}
}
func TestAttachmentDropAtFullCapStopsOrdinaryPath(t *testing.T) {
	a := fixture(t)
	attachmentOrderExisting(a, 8)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("prefix")
	tt.DropFiles(600, 400, "tail.txt")
	tt.Frame()
	if a.Draft != "prefix" {
		t.Fatalf("drop at full source cap inserted a later file: %q", a.Draft)
	}
}
func TestAttachmentDropSevenImagesStopsAfterFirstInputImage(t *testing.T) {
	a := fixture(t)
	attachmentOrderExisting(a, 7)
	raw := attachmentOrderPNG(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("prefix")
	reads := make(chan string, 4)
	a.addPaths([]string{"first.png", "second.png", "tail.txt"}, func(path string) ([]byte, string, error) { reads <- path; return raw, "image/png", nil })
	drainAttachmentOrder(t, a)
	tt.Frame()
	var gotReads []string
	for len(reads) > 0 {
		gotReads = append(gotReads, <-reads)
	}
	if len(a.attachments) != 8 || a.attachments[7].Name != "first.png" || a.Draft != "prefix" || len(gotReads) != 1 || gotReads[0] != "first.png" {
		t.Fatalf("near-cap drop did not stop in source order: attachments=%v draft=%q reads=%v", len(a.attachments), a.Draft, gotReads)
	}
}
func TestAttachmentDropFailedFirstImageFallsBackThenContinues(t *testing.T) {
	a := fixture(t)
	attachmentOrderExisting(a, 7)
	raw := attachmentOrderPNG(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("prefix")
	a.addPaths([]string{"missing.png", "good.png", "after.txt"}, func(path string) ([]byte, string, error) {
		if path == "missing.png" {
			return nil, "", errors.New("synthetic unreadable image")
		}
		return raw, "image/png", nil
	})
	drainAttachmentOrder(t, a)
	tt.Frame()
	if len(a.attachments) != 8 || a.attachments[7].Name != "good.png" || a.Draft != "prefixmissing.png " {
		t.Fatalf("failed first image lost source path fallback/continuation: count=%d draft=%q", len(a.attachments), a.Draft)
	}
}

func TestAttachmentOriginNativePickerReadFailurePolicies(t *testing.T) {
	for _, origin := range []string{"file", "image"} {
		t.Run(origin, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			a := x.UI
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			tt.Type("prefix")
			kind := "BrowseAttachments"
			if origin == "image" {
				kind = "BrowseImage"
			}
			missing := filepath.Join(t.TempDir(), "missing.png")
			release := holdFileDialog(t, x, Action{Kind: kind}, []string{missing}, nil)
			release()
			takeControllerDialogUpdate(t, x)()
			drainAttachmentOrder(t, a)
			tt.Frame()
			if origin == "file" {
				if a.Draft != "prefix"+missing+" " || len(a.attachments) != 0 {
					t.Fatalf("file picker read failure lost source caret fallback: %q", a.Draft)
				}
			} else {
				if a.Draft != "prefix" || len(a.pendingAttachmentPaths) != 0 || len(a.toastStack.Items()) == 0 {
					t.Fatal("image-only picker must report failure without inserting prompt path")
				}
			}
		})
	}
}
func TestAttachmentOriginFilePickerAtCapStillReferencesOrdinaryFile(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	attachmentOrderExisting(a, 8)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("prefix")
	release := holdFileDialog(t, x, Action{Kind: "BrowseAttachments"}, []string{"tail.txt"}, nil)
	release()
	takeControllerDialogUpdate(t, x)()
	tt.Frame()
	if a.Draft != "prefixtail.txt " {
		t.Fatalf("drop-only cap was incorrectly imposed on file picker: %q", a.Draft)
	}
}
