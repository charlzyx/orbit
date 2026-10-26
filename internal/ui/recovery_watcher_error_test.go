package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"os"
	"testing"
)

func TestRecoveryWatcherBatchNamespaceErrors(t *testing.T) {
	a, _, p := mountedFile(t, "original")
	s := a.files
	st := s.states[p]
	st.Buffer.Set("pending")
	ack := s.versions[p]
	failures := []error{workspace.ErrNamespaceBusy, fmt.Errorf("wrapped: %w", workspace.ErrNamespaceBusy), workspace.ErrNamespaceChanged, fmt.Errorf("wrapped: %w", workspace.ErrNamespaceChanged), &os.PathError{Op: "read", Path: "one", Err: os.ErrPermission}, &os.PathError{Op: "read", Path: "two", Err: os.ErrPermission}}
	before := len(a.toastStack.Items())
	a.reportFileReloadErrors(failures)
	items := a.toastStack.Items()[before:]
	if len(items) != 4 {
		t.Fatalf("want two namespace classes + two original OS errors, got %d", len(items))
	}
	for _, err := range []error{failures[0], failures[2], failures[4], failures[5]} {
		found := false
		for _, item := range items {
			if item.Body == err.Error() {
				found = true
			}
		}
		if !found {
			t.Fatal("missing original error", err)
		}
	}
	next := fmt.Errorf("next batch: %w", workspace.ErrNamespaceBusy)
	a.reportFileReloadErrors([]error{next})
	found := false
	for _, item := range a.toastStack.Items() {
		if item.Body == next.Error() {
			found = true
		}
	}
	if !found {
		t.Fatal("cross-batch suppression")
	}
	n := len(a.toastStack.Items())
	a.reportFileReloadErrors([]error{nil, workspace.ErrReadSuperseded, workspace.ErrDocumentNotOpen, workspace.ErrEditorClosed})
	if len(a.toastStack.Items()) != n {
		t.Fatal("cancellation notified")
	}
	if st.Buffer.String() != "pending" || s.versions[p] != ack {
		t.Fatal("error acknowledgement destroyed pending")
	}
	select {
	case <-a.updates:
		t.Fatal("error triggered retry")
	default:
	}
}
