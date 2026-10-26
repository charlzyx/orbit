package orbitui

import (
	"errors"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui/editor"
)

// reloadWatchedFiles schedules disk reads away from the UI thread. Native
// edits can precede the next editor build, or arrive while the reads run.
func (a *App) reloadWatchedFiles() {
	if a.files == nil || !a.syncFileBuffers() {
		return
	}
	s := a.files
	ed := s.editor
	type target struct {
		id    workspace.DocumentID
		path  string
		state *editor.State
	}
	targets := make([]target, 0)
	for _, d := range ed.Documents() {
		p := d.Content.Path
		st := s.states[p]
		// Closing native targets do not belong to a new refresh.
		if s.closing[p] != nil {
			continue
		}
		targets = append(targets, target{d.ID, p, st})
	}
	if len(targets) == 0 {
		return
	}
	a.Async(func() func() {
		outcomes := map[workspace.DocumentID]error{}
		for _, target := range targets {
			outcomes[target.id] = ed.ReloadID(target.id)
		}
		return func() {
			if a.files != s || s.editor != ed {
				return
			}
			failures := make([]error, 0, len(targets))
			for _, target := range targets {
				failures = append(failures, outcomes[target.id])
			}
			a.reportFileReloadErrors(failures)
			// Use current domain state, but only for this job's original native
			// targets. A newer tab or reopened path does not belong to this result.
			for _, d := range ed.Documents() {
				for _, target := range targets {
					p := target.path
					if outcomes[target.id] != nil || d.ID != target.id || d.Content.Path != p || target.state == nil || s.states[p] != target.state || s.closing[p] != nil {
						continue
					}
					if target.state.Buffer.Version() != s.versions[p] {
						continue
					}
					a.syncDocument(d)
				}
			}
		}
	})
}

func (a *App) reportFileReloadError(err error) {
	if err != nil && !errors.Is(err, workspace.ErrDocumentNotOpen) && !errors.Is(err, workspace.ErrEditorClosed) && !errors.Is(err, workspace.ErrReadSuperseded) {
		a.save(err)
	}
}

// Only the two transient namespace classes are coalesced, only within this batch.
func (a *App) reportFileReloadErrors(failures []error) {
	seen := map[error]bool{}
	for _, err := range failures {
		duplicate := false
		for _, class := range []error{workspace.ErrNamespaceBusy, workspace.ErrNamespaceChanged} {
			if errors.Is(err, class) {
				duplicate = seen[class]
				seen[class] = true
				break
			}
		}
		if !duplicate {
			a.reportFileReloadError(err)
		}
	}
}
