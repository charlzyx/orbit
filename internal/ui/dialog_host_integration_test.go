package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
	"time"
)

// Controller native completions use its host mailbox; App I/O still uses its
// separate queue. Exercise the production owner drain rather than the retired
// transport detail. Neither callback kind runs on the worker here.
func drainControllerDialogUpdates(t *testing.T, x *Controller, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		x.drainHostUpdates()
		if done() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("controller dialog completion did not settle")
}
func takeControllerDialogUpdate(t *testing.T, x *Controller) func() {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		x.hostMailboxMu.Lock()
		if len(x.hostMailbox) > 0 {
			apply := x.hostMailbox[0]
			x.hostMailbox[0] = nil
			x.hostMailbox = x.hostMailbox[1:]
			x.hostMailboxMu.Unlock()
			return apply
		}
		x.hostMailboxMu.Unlock()
		time.Sleep(time.Millisecond)
	}
	t.Fatal("native completion did not reach host mailbox")
	return nil
}
func TestWorkspaceDialogCompletionConsumedByActualControllerFrame(t *testing.T) {
	x := workspaceDialogFixture(t)
	target := t.TempDir()
	release := holdWorkspaceDialog(t, x, target, nil)
	tt := ui.NewTester(x.View, 1470, 922)
	tt.Frame()
	release()
	deadline := time.Now().Add(5 * time.Second)
	for x.workspaceBrowse.pending && time.Now().Before(deadline) {
		tt.Frame()
		time.Sleep(time.Millisecond)
	}
	if x.workspaceBrowse.pending || x.workspace != target {
		t.Fatal("real Controller.View did not consume native answer", x.workspace)
	}
}
