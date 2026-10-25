package orbitui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func TestAttachmentSubmitRPCChild(t *testing.T) {
	if os.Getenv("ORBIT_ATTACHMENT_SUBMIT_CHILD") != "1" {
		return
	}
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		data := map[string]any{}
		if kind == "get_state" {
			data["sessionId"] = fmt.Sprintf("attachment-%d", os.Getpid())
		}
		if kind == "prompt" || kind == "steer" || kind == "follow_up" {
			_ = os.WriteFile(os.Getenv("ORBIT_ATTACHMENT_SUBMIT_LOG"), append([]byte(nil), scan.Bytes()...), 0600)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"type": "response", "command": kind, "id": c["id"], "success": true, "data": data})
	}
	os.Exit(0)
}
func attachmentSubmitFixture(t *testing.T) (*Controller, string, []byte) {
	t.Helper()
	x := workspaceDialogFixture(t)
	_ = x.Agent.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "prompt.json")
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestAttachmentSubmitRPCChild$", "--"}, Home: x.Store.Home, Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_ATTACHMENT_SUBMIT_CHILD": "1", "ORBIT_ATTACHMENT_SUBMIT_LOG": log}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		x.tickWindowless()
		if x.Agent.Snapshot().Connected && x.Agent.Snapshot().ActiveSession.ID != "" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !x.Agent.Snapshot().Connected {
		t.Fatal("synthetic child did not connect")
	}
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return x, log, raw.Bytes()
}
func waitAttachmentPrompt(t *testing.T, log string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(log)
		if err == nil {
			var out map[string]any
			if json.Unmarshal(raw, &out) == nil {
				return out
			}
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("synthetic child did not receive expected prompt")
	return nil
}
func TestAttachmentPendingReadMountedSubmitKeepsDraft(t *testing.T) {
	x, log, raw := attachmentSubmitFixture(t)
	a := x.UI
	a.Draft = "send with the pending image"
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	a.addPaths([]string{"synthetic.png"}, func(string) ([]byte, string, error) { close(started); <-release; return raw, "image/png", nil })
	<-started
	tt.Key(0, ui.KeyEnter)
	if a.Draft != "send with the pending image" {
		sent := waitAttachmentPrompt(t, log)
		t.Fatalf("pending image was omitted and draft cleared: draft=%q actualRPC=%+v", a.Draft, sent)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("pending image allowed actual prompt")
	}
}
func TestAttachmentPendingDecodeOwnerCallbackPreventsEarlyActualSubmit(t *testing.T) {
	x, log, raw := attachmentSubmitFixture(t)
	a := x.UI
	a.Draft = "paste then immediately send"
	a.AddImage(raw, "image/png", "clipboard")
	var apply func()
	select {
	case apply = <-a.updates:
	case <-time.After(2 * time.Second):
		t.Fatal("image decode did not queue owner completion")
	}
	accepted := a.submit(a.Draft, "")
	if accepted {
		sent := waitAttachmentPrompt(t, log)
		apply()
		t.Fatalf("queued image callback lost race to real submit: RPC=%+v lateAttachments=%d", sent, len(a.attachments))
	}
	apply()
	if len(a.attachments) != 1 {
		t.Fatal("completed image was not retained after rejected submit")
	}
}

func takeAttachmentApply(t *testing.T, a *App) func() {
	t.Helper()
	select {
	case apply := <-a.updates:
		return apply
	case <-time.After(2 * time.Second):
		t.Fatal("attachment owner callback missing")
		return nil
	}
}
func assertAttachmentRPCImage(t *testing.T, log, text string) {
	t.Helper()
	sent := waitAttachmentPrompt(t, log)
	images, _ := sent["images"].([]any)
	if sent["message"] != text || len(images) != 1 {
		t.Fatalf("final prompt lost text/image: %+v", sent)
	}
}
func TestAttachmentPendingReadDecodeHandoffAndDuplicateCallback(t *testing.T) {
	x, log, raw := attachmentSubmitFixture(t)
	a := x.UI
	a.Draft = "complete image then send"
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	a.addPaths([]string{"synthetic.png"}, func(string) ([]byte, string, error) { close(started); <-release; return raw, "image/png", nil })
	<-started
	unblock()
	readApply := takeAttachmentApply(t, a)
	readApply()
	readApply()
	if !a.hasPendingAttachmentImages() || a.pendingAttachmentImages.count != 1 {
		t.Fatal("read/decode handoff opened submit gap or duplicate read reserved twice")
	}
	a.submitAccepted = true
	x.Action(Action{Kind: "Submit", Value: a.Draft, Data: map[string]any{"mode": "steer"}})
	if a.submitAccepted {
		t.Fatal("direct Controller Submit bypassed pending-image guard")
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("handoff sent prompt")
	}
	decodeApply := takeAttachmentApply(t, a)
	decodeApply()
	decodeApply()
	if a.hasPendingAttachmentImages() || len(a.attachments) != 1 {
		t.Fatal("duplicate decode altered pending count or appended image twice")
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("completion automatically sent a prompt")
	}
	if !a.submit(a.Draft, "") {
		t.Fatal("completed image still blocked explicit submit")
	}
	assertAttachmentRPCImage(t, log, "complete image then send")
	if len(a.attachments) != 0 {
		t.Fatal("accepted host submit retained already-sent attachment")
	}
}
func TestAttachmentPendingFailureAllowsRetryWithoutChangingExistingError(t *testing.T) {
	x, log, _ := attachmentSubmitFixture(t)
	a := x.UI
	a.Draft = "text survives image read error"
	a.Notify("existing diagnostic", true)
	a.addPaths([]string{"synthetic.png"}, func(string) ([]byte, string, error) { return nil, "", errors.New("fixture read failed") })
	if a.submit(a.Draft, "") {
		t.Fatal("pending failed-read completion allowed early send")
	}
	if items := a.toastStack.Items(); len(items) != 1 || items[0].Body != "existing diagnostic" {
		t.Fatal("pending guard overwrote original error", items)
	}
	apply := takeAttachmentApply(t, a)
	apply()
	apply()
	if a.hasPendingAttachmentImages() || a.Draft != "text survives image read error" {
		t.Fatal("failed read stranded pending or lost draft")
	}
	items := a.toastStack.Items()
	if len(items) != 2 || items[1].Body != "fixture read failed" {
		t.Fatal("read error changed or repeated callback duplicated error", items)
	}
	if !a.submit(a.Draft, "") {
		t.Fatal("read failure prevented a later explicit submit")
	}
	sent := waitAttachmentPrompt(t, log)
	if sent["message"] != a.Draft || sent["images"] != nil {
		t.Fatalf("failed-read retry altered prompt: %+v", sent)
	}
}
func TestAttachmentPendingBlockedEnterPreservesSelectionAndExistingImage(t *testing.T) {
	x, log, raw := attachmentSubmitFixture(t)
	a := x.UI
	a.attachments = []attachment{{ID: "existing", Name: "existing image", Image: agent.Image{Type: "image", Data: "fixture", MIMEType: "image/png"}}}
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("甲😀乙尾")
	tt.Key(0, ui.KeyLeft)
	tt.Key(0, ui.KeyLeft)
	tt.Key(ui.Shift, ui.KeyRight)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	a.addPaths([]string{"synthetic.png"}, func(string) ([]byte, string, error) { close(started); <-release; return raw, "image/png", nil })
	<-started
	tt.Key(0, ui.KeyEnter)
	if a.Draft != "甲😀乙尾" || len(a.attachments) != 1 || a.attachments[0].ID != "existing" {
		t.Fatal("blocked Enter changed text or existing image")
	}
	tt.Type("中")
	if a.Draft != "甲😀中尾" {
		t.Fatalf("blocked submit lost native selection/caret: %q", a.Draft)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("blocked Enter sent prompt")
	}
}
func TestAttachmentPendingOldOwnerCannotBlockOrUnlockNewDraft(t *testing.T) {
	x, log, raw := attachmentSubmitFixture(t)
	a := x.UI
	oldStarted, oldRelease := make(chan struct{}), make(chan struct{})
	var oldOnce sync.Once
	releaseOld := func() { oldOnce.Do(func() { close(oldRelease) }) }
	defer releaseOld()
	a.addPaths([]string{"old.png"}, func(string) ([]byte, string, error) { close(oldStarted); <-oldRelease; return raw, "image/png", nil })
	<-oldStarted
	oldBatch := a.pendingAttachmentImages
	if err := x.Agent.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		x.tickWindowless()
		if x.Agent.Snapshot().Connected && x.Agent.Snapshot().ActiveSession.ID != "" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	a.Navigate("settings")
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if a.hasPendingAttachmentImages() || !oldBatch.cancelled {
		t.Fatal("hidden new task retained old owner's pending gate")
	}
	a.Draft = "new task text"
	if !a.submit(a.Draft, "") {
		t.Fatal("old owner blocked new draft")
	}
	_ = waitAttachmentPrompt(t, log)
	if err := os.Remove(log); err != nil {
		t.Fatal(err)
	}
	newStarted, newRelease := make(chan struct{}), make(chan struct{})
	var newOnce sync.Once
	releaseNew := func() { newOnce.Do(func() { close(newRelease) }) }
	defer releaseNew()
	a.addPaths([]string{"new.png"}, func(string) ([]byte, string, error) { close(newStarted); <-newRelease; return raw, "image/png", nil })
	<-newStarted
	newBatch := a.pendingAttachmentImages
	releaseOld()
	oldApply := takeAttachmentApply(t, a)
	oldApply()
	oldApply()
	if a.pendingAttachmentImages != newBatch || newBatch.count != 1 || !a.hasPendingAttachmentImages() || len(a.attachments) != 0 {
		t.Fatal("old completion unlocked/adopted new owner")
	}
	if a.submit("new image prompt", "") {
		t.Fatal("new pending image gate was lost")
	}
	releaseNew()
	takeAttachmentApply(t, a)()
	takeAttachmentApply(t, a)()
	if len(a.attachments) != 1 || a.attachments[0].Name != "new.png" || a.hasPendingAttachmentImages() {
		t.Fatal("new owner's image failed to settle")
	}
	if !a.submit("new image prompt", "") {
		t.Fatal("settled new image still blocked")
	}
	assertAttachmentRPCImage(t, log, "new image prompt")
}
func TestAttachmentPendingCancellationAndCloseDiscardLateCompletion(t *testing.T) {
	for _, kind := range []string{"cancel", "close"} {
		t.Run(kind, func(t *testing.T) {
			x, _, raw := attachmentSubmitFixture(t)
			a := x.UI
			a.AddImage(raw, "image/png", "late clipboard")
			apply := takeAttachmentApply(t, a)
			if kind == "close" {
				x.Close()
			} else {
				a.cancelAttachmentImages()
			}
			if a.hasPendingAttachmentImages() {
				t.Fatal("cancellation retained submit gate")
			}
			apply()
			apply()
			if len(a.attachments) != 0 || a.hasPendingAttachmentImages() {
				t.Fatal("canceled completion resurrected image or pending work")
			}
		})
	}
}
