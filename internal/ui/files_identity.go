package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"path/filepath"
	"strings"
)

// Path maps are current indexes only. Native ownership survives a path change
// through the editor-local ID, never through an old-path alias.
type fileNative struct {
	state   *editor.State
	version uint64
	image   ui.ImageSource
}

func (a *App) captureFileNative(d workspace.Document) {
	if !a.liveFileDocument(d) {
		return
	}
	s := a.files
	if s.native == nil {
		s.native = map[workspace.DocumentID]*fileNative{}
		s.pathIDs = map[string]workspace.DocumentID{}
	}
	s.native[d.ID] = &fileNative{s.states[d.Content.Path], s.versions[d.Content.Path], s.images[d.Content.Path]}
	s.pathIDs[d.Content.Path] = d.ID
}
func (a *App) bindFileNative(d workspace.Document) bool {
	if !a.liveFileDocument(d) {
		return false
	}
	s := a.files
	if s.native == nil {
		s.native = map[workspace.DocumentID]*fileNative{}
		s.pathIDs = map[string]workspace.DocumentID{}
	}
	a.pruneFileNative()
	p := d.Content.Path
	if s.pathIDs[p] == d.ID && d.ID != 0 {
		return true
	}
	// Remove only this ID's previous index. Another document may now own it.
	for old, id := range s.pathIDs {
		if id == d.ID {
			delete(s.pathIDs, old)
			delete(s.states, old)
			delete(s.versions, old)
			delete(s.images, old)
		}
	}
	n := s.native[d.ID]
	delete(s.states, p)
	delete(s.versions, p)
	delete(s.images, p)
	s.pathIDs[p] = d.ID
	if n != nil {
		s.states[p] = n.state
		s.versions[p] = n.version
		s.images[p] = n.image
	}
	return true
}

func (a *App) liveFileDocument(d workspace.Document) bool {
	if a.files == nil || d.ID == 0 {
		return false
	}
	for _, live := range a.files.editor.Documents() {
		if live.ID == d.ID {
			return live.Content.Path == d.Content.Path
		}
	}
	return false
}

type fileRenameTarget struct {
	id    workspace.DocumentID
	old   string
	state *editor.State
}
type fileRenameReceipt struct {
	owner      *filesState
	editor     *workspace.Editor
	generation uint64
	old        string
	targets    []fileRenameTarget
}

func (a *App) captureFileRename(old string) fileRenameReceipt {
	s := a.files
	r := fileRenameReceipt{owner: s, editor: s.editor, generation: s.openGeneration, old: old}
	if old == "" {
		return r
	}
	for _, d := range s.editor.Documents() {
		rel, err := filepath.Rel(old, d.Content.Path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		a.bindFileNative(d)
		r.targets = append(r.targets, fileRenameTarget{d.ID, d.Content.Path, s.states[d.Content.Path]})
	}
	return r
}
func (a *App) applyFileRename(r fileRenameReceipt, renamed string) {
	s := a.files
	if s != r.owner || s.editor != r.editor {
		return
	}
	navigationCurrent := s.openGeneration == r.generation
	invalidateOpen := false
	for _, target := range r.targets {
		rel, _ := filepath.Rel(r.old, target.old)
		next := filepath.Join(renamed, rel)
		for _, d := range s.editor.Documents() {
			if d.ID != target.id || d.Content.Path != next {
				continue
			}
			n := s.native[d.ID]
			if (n == nil && target.state != nil) || (n != nil && n.state != target.state) {
				continue
			}
			if !a.bindFileNative(d) {
				continue
			}
			if navigationCurrent {
				if s.desired == target.old {
					s.desired = next
					invalidateOpen = true
				}
				if s.focusPending == target.old {
					s.focusPending = next
				}
				if active, ok := s.editor.Active(); ok && active.ID == target.id {
					s.focusPending = next
				}
			}
		}
	}
	// Only supersede an intent belonging to this accepted rename, never newer navigation.
	if invalidateOpen {
		s.openGeneration++
	}
}

func (a *App) pruneFileNative() {
	s := a.files
	live := map[workspace.DocumentID]bool{}
	for _, d := range s.editor.Documents() {
		live[d.ID] = true
	}
	for id := range s.native {
		if !live[id] {
			delete(s.native, id)
		}
	}
	for p, id := range s.pathIDs {
		if !live[id] {
			delete(s.pathIDs, p)
			delete(s.states, p)
			delete(s.versions, p)
			delete(s.images, p)
		}
	}
}
