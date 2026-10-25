package orbitui

import (
	"strings"
	"sync"
	"testing"
)

// Drain the real background initialization callback before retaining a specific
// attachment completion; an arbitrary mailbox item is not an image barrier.
func attachmentIngressFixture(t *testing.T) *App {
	a := fixture(t)
	for a.backdrop.loading {
		takeAttachmentApply(t, a)()
	}
	return a
}

func TestAttachmentIngressExactSafetyLimits(t *testing.T) {
	paths := make([]string, attachmentMaxPaths)
	for i := range paths {
		paths[i] = "a"
	}
	if !attachmentPathsFit(paths) || attachmentPathsFit(append(paths, "b")) {
		t.Fatal("path count boundary")
	}
	if !attachmentPathsFit([]string{strings.Repeat("x", attachmentMaxPathBytes)}) || attachmentPathsFit([]string{strings.Repeat("x", attachmentMaxPathBytes), "x"}) {
		t.Fatal("UTF-8 byte boundary")
	}
	for _, tc := range []struct {
		size, pending int
		want          bool
	}{
		{attachmentMaxClipboard, 0, true}, {attachmentMaxClipboard + 1, 0, false},
		{attachmentMaxClipboard, attachmentMaxClipboard, true}, {attachmentMaxClipboard, attachmentMaxClipboard + 1, false},
		{1, attachmentMaxRawPending, false}, {0, attachmentMaxRawPending, true}, {1, int(^uint(0) >> 1), false},
	} {
		if got := attachmentRawFits(tc.size, tc.pending); got != tc.want {
			t.Fatalf("raw boundary %+v: %v", tc, got)
		}
	}
}
func TestAttachmentIngressFIFOIncludesClipboardAndFilePicker(t *testing.T) {
	a := attachmentIngressFixture(t)
	raw := attachmentOrderPNG(t)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	a.addPaths([]string{"drop.png"}, func(string) ([]byte, string, error) { close(started); <-release; return raw, "image/png", nil })
	<-started
	a.AddImage(raw, "image/png", "clipboard")
	a.addPathsOrigin([]string{"file.png"}, attachmentFilePicker, func(string) ([]byte, string, error) { return raw, "image/png", nil })
	if len(a.attachments) != 0 {
		t.Fatal("later origin overtook held drop")
	}
	unblock()
	drainAttachmentOrder(t, a)
	if len(a.attachments) != 3 {
		t.Fatal("missing attachments", len(a.attachments))
	}
	for i, want := range []string{"drop.png", "clipboard", "file.png"} {
		if a.attachments[i].Name != want {
			t.Fatal("mixed origin reordered", a.attachments)
		}
	}
	if a.attachmentIngress.requests != 0 || a.attachmentIngress.slots != 0 || a.attachmentIngress.rawBytes != 0 {
		t.Fatal("settled resources leaked", a.attachmentIngress)
	}
}
func TestAttachmentIngressRequestLimitRejectsWholeRequest(t *testing.T) {
	a := attachmentIngressFixture(t)
	raw := attachmentOrderPNG(t)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	a.Draft = "keep"
	a.addPaths([]string{"first.png"}, func(string) ([]byte, string, error) { close(started); <-release; return raw, "image/png", nil })
	<-started
	for i := 1; i < attachmentMaxIngress; i++ {
		a.AddImage(raw, "image/png", "queued")
	}
	before := a.attachmentIngress.rawBytes
	a.addPaths([]string{"must-not-insert.txt", "ninth.png"}, readAttachmentPath)
	if a.attachmentIngress.requests != 8 || a.pendingAttachmentImages.count != 8 || a.attachmentIngress.rawBytes != before || a.Draft != "keep" {
		t.Fatal("non-atomic rejection", a.attachmentIngress, a.Draft)
	}
	unblock()
	drainAttachmentOrder(t, a)
	if len(a.attachments) != 8 {
		t.Fatal("accepted requests lost", len(a.attachments))
	}
}
func TestAttachmentIngressRetiredOwnersKeepActualResources(t *testing.T) {
	a := attachmentIngressFixture(t)
	raw := attachmentOrderPNG(t)
	a.AddImage(raw, "image/png", "old-one")
	first := takeAttachmentApply(t, a) // decoded but not consumed still owns raw and slot
	a.Model.RuntimeID = "two"
	a.pruneAttachmentImages()
	if a.attachmentIngress.rawBytes != len(raw) || a.attachmentIngress.slots != 1 {
		t.Fatal("cancel freed retained data")
	}
	a.AddImage(raw, "image/png", "old-two")
	second := takeAttachmentApply(t, a)
	a.Model.RuntimeID = "three"
	a.pruneAttachmentImages()
	if a.attachmentIngress.rawBytes != 2*len(raw) || a.attachmentIngress.slots != 2 {
		t.Fatal("retired owners unaccounted")
	}
	a.AddImage(raw, "image/png", "rejected")
	if a.hasPendingAttachmentImages() || a.attachmentIngress.requests != 2 {
		t.Fatal("slot rejection created token")
	}
	first()
	first()
	if a.attachmentIngress.rawBytes != len(raw) || a.attachmentIngress.slots != 1 {
		t.Fatal("idempotent old cleanup accounting")
	}
	a.AddImage(raw, "image/png", "current")
	current := takeAttachmentApply(t, a)
	second()
	second()
	if !a.hasPendingAttachmentImages() || a.attachmentIngress.rawBytes != len(raw) || len(a.attachments) != 0 {
		t.Fatal("old result unlocked or wrote new owner")
	}
	current()
	current()
	if a.hasPendingAttachmentImages() || len(a.attachments) != 1 || a.attachments[0].Name != "current" || a.attachmentIngress.rawBytes != 0 || a.attachmentIngress.slots != 0 {
		t.Fatal("new owner did not recover")
	}
}
func TestAttachmentIngressRejectedPathsDoNotBlockTextSubmit(t *testing.T) {
	x, log, _ := attachmentSubmitFixture(t)
	a := x.UI
	a.Draft = "only explicit text"
	paths := make([]string, attachmentMaxPaths+1)
	for i := range paths {
		paths[i] = "unaccepted.txt"
	}
	a.AddPaths(paths)
	if a.hasPendingAttachmentImages() || len(a.pendingAttachmentPaths) != 0 || a.Draft != "only explicit text" || a.attachmentIngress.requests != 0 {
		t.Fatal("rejected paths altered draft or admission")
	}
	if !a.submit(a.Draft, "") {
		t.Fatal("rejection blocked text submit")
	}
	sent := waitAttachmentPrompt(t, log)
	if sent["message"] != "only explicit text" || sent["images"] != nil {
		t.Fatal("rejected input sent", sent)
	}
	a.AddPaths([]string{"normal.txt"})
	if len(a.pendingAttachmentPaths) != 1 || a.pendingAttachmentPaths[0].text != "normal.txt " {
		t.Fatal("small retry did not recover", a.Draft)
	}
}

func TestAttachmentIngressCancelReleasesQueuedButNotRunningRaw(t *testing.T) {
	a := attachmentIngressFixture(t)
	raw := attachmentOrderPNG(t)
	a.AddImage(raw, "image/png", "running")
	apply := takeAttachmentApply(t, a)
	a.AddImage(raw, "image/png", "queued-one")
	a.AddImage(raw, "image/png", "queued-two")
	if a.attachmentIngress.rawBytes != 3*len(raw) || a.attachmentIngress.requests != 3 {
		t.Fatal("queued budget missing")
	}
	a.cancelAttachmentImages()
	if a.hasPendingAttachmentImages() || a.attachmentIngress.rawBytes != len(raw) || a.attachmentIngress.requests != 1 || a.attachmentIngress.slots != 1 {
		t.Fatal("cancel released running or retained queued inputs", a.attachmentIngress)
	}
	apply()
	apply()
	if a.attachmentIngress.rawBytes != 0 || a.attachmentIngress.requests != 0 || a.attachmentIngress.slots != 0 || len(a.attachments) != 0 {
		t.Fatal("retired completion did not clean exactly once", a.attachmentIngress)
	}
}
