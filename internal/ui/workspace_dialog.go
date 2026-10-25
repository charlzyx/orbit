package orbitui

import (
	"time"

	"github.com/charlzyx/orbit/internal/diagnosticlog"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo"
)

// These fields are owned by the UI thread. A native dialog may outlive a view,
// session, or workspace; pending remains true until its answer is consumed so
// repeated clicks cannot create stacked native sheets.
type workspaceBrowseState struct {
	pending bool
	epoch   uint64
	open    func(*mygo.Window, string) (string, error)
}

// Rust session.rs::browse_for_folder deliberately returns from its handler
// before awaiting the native sheet, avoiding nested event-loop reentry. MyGo's
// Dialog.Open is blocking but documented safe from any goroutine: off the UI
// thread it waits on a channel instead of recursively pumping native events.
// Never move SetWorkspace itself off the UI thread.
func (x *Controller) browseWorkspace() {
	select {
	case <-x.stop:
		return
	default:
	}
	if x.workspaceBrowse.pending || x.fileDialogs.pending {
		return
	}
	parent, current, epoch := x.UI.Window, x.workspace, x.workspaceBrowse.epoch
	if parent != nil && parent.IsDestroyed() {
		return
	}
	open := x.workspaceBrowse.open
	if open == nil {
		open = native.BrowseWorkspace
	}
	x.workspaceBrowse.pending = true
	started := time.Now()
	x.diagnostics.Record(diagnosticlog.EventWorkspaceBrowse, diagnosticlog.ResultStarted, 0)
	go func() {
		select {
		case <-x.stop:
			return
		default:
		}
		path, err := open(parent, current)
		result := diagnosticlog.ResultOK
		if err != nil {
			result = diagnosticlog.ResultError
		} else if path == "" {
			result = diagnosticlog.ResultCancelled
		}
		x.diagnostics.Record(diagnosticlog.EventWorkspaceBrowse, result, time.Since(started))
		x.enqueue(func() {
			x.workspaceBrowse.pending = false
			select {
			case <-x.stop:
				return
			default:
			}
			// New task/workspace/session intent or a closed/replaced parent wins over
			// a late sheet answer. An error from that obsolete sheet is also stale.
			if epoch != x.workspaceBrowse.epoch || current != x.workspace || parent != x.UI.Window || (parent != nil && parent.IsDestroyed()) {
				return
			}
			if err != nil {
				x.UI.save(err)
				return
			}
			if path != "" {
				x.SetWorkspace(path)
			}
		})
	}()
}
