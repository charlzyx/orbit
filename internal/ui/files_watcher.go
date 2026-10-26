package orbitui

import "github.com/egoist/mygo/ui/editor"

// reloadWatchedFiles schedules disk reads away from the UI thread. Native
// edits can precede the next editor build, or arrive while the reads run.
func (a *App) reloadWatchedFiles() {
	if a.files == nil || !a.syncFileBuffers() {
		return
	}
	s := a.files
	ed := s.editor
	type target struct {
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
		targets = append(targets, target{p, st})
	}
	if len(targets) == 0 {
		return
	}
	a.Async(func() func() {
		for _, target := range targets {
			_ = ed.Reload(target.path)
		}
		return func() {
			if a.files != s || s.editor != ed {
				return
			}
			// Use current domain state, but only for this job's original native
			// targets. A newer tab or reopened path does not belong to this result.
			for _, d := range ed.Documents() {
				for _, target := range targets {
					p := target.path
					if d.Content.Path != p || target.state == nil || s.states[p] != target.state || s.closing[p] != nil {
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
