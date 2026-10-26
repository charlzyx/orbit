package workspace

import (
	"errors"
	"path/filepath"
)

var ErrEditorClosed = errors.New("editor closed")
var ErrNamespaceBusy = errors.New("file namespace is being renamed; load was not applied")
var ErrNamespaceChanged = errors.New("file namespace changed during loading; load was not applied")
var ErrReadSuperseded = errors.New("file load superseded by a newer read")
var ErrRenameChanged = errors.New("rename targets changed; retry rename")
var ErrRenameDirty = errors.New("save unsaved edits before renaming this path")
var ErrReopenOccupied = errors.New("closed file path is occupied by another document; reopen was not applied")

type RenameTarget struct {
	ID   DocumentID
	Path string
}
type RenameRequest struct {
	Old, Name string
	Targets   []RenameTarget
}
type RenameResult struct {
	Path    string
	Targets []RenameTarget
}
type renameReservation struct{ old, destination string }

func (e *Editor) findID(id DocumentID) (int, *Document) {
	for i, d := range e.tabs {
		if d.ID == id {
			return i, d
		}
	}
	return -1, nil
}
func (e *Editor) namespaceBusy(p string) bool {
	return e.reservation != nil && (pathUnder(p, e.reservation.old) || pathUnder(p, e.reservation.destination))
}

// The existing I/O mutex serializes filesystem mutation with writes; the state
// mutex is released throughout I/O so snapshots and late edits remain responsive.
func (e *Editor) RenameDocuments(req RenameRequest) (RenameResult, error) {
	name, err := ValidateName(req.Name)
	if err != nil {
		return RenameResult{}, err
	}
	old := filepath.Clean(req.Old)
	destination := filepath.Join(filepath.Dir(old), name)
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return RenameResult{}, ErrEditorClosed
	}
	expected := map[DocumentID]string{}
	for _, target := range req.Targets {
		if target.ID == 0 || expected[target.ID] != "" || !pathUnder(target.Path, old) {
			e.mu.Unlock()
			return RenameResult{}, ErrRenameChanged
		}
		expected[target.ID] = target.Path
	}
	affected := make([]*Document, 0, len(expected))
	for _, d := range e.tabs {
		if pathUnder(d.Content.Path, old) {
			if expected[d.ID] != d.Content.Path {
				e.mu.Unlock()
				return RenameResult{}, ErrRenameChanged
			}
			if d.Dirty || d.Saving {
				e.mu.Unlock()
				return RenameResult{}, ErrRenameDirty
			}
			affected = append(affected, d)
		} else if pathUnder(d.Content.Path, destination) {
			e.mu.Unlock()
			return RenameResult{}, ErrDestinationExists
		}
	}
	if len(affected) != len(expected) {
		e.mu.Unlock()
		return RenameResult{}, ErrRenameChanged
	}
	e.namespaceEpoch++
	e.reservation = &renameReservation{old, destination}
	rename := e.renamePath
	if rename == nil {
		rename = RenamePath
	}
	e.mu.Unlock()
	next, err := rename(old, name)
	e.mu.Lock()
	result := RenameResult{Path: next}
	if err == nil {
		for _, d := range affected {
			rel, _ := filepath.Rel(old, d.Content.Path)
			d.Content.Path = filepath.Join(next, rel)
			d.Content.Display = filepath.Base(d.Content.Path)
			result.Targets = append(result.Targets, RenameTarget{d.ID, d.Content.Path})
		}
	}
	e.namespaceEpoch++
	e.reservation = nil
	e.notify()
	e.mu.Unlock()
	return result, err
}
