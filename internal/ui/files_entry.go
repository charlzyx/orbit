package orbitui

import (
	"path/filepath"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// Explorer EntryEditor contract: panel.rs:592–687,957–1014,1079–1110.
// Generic viewer prompts remain separate; only Explorer owns this inline editor.
type explorerEntry struct {
	root, parent, target, relative, value, err string
	epoch, serial                              uint64
	directory, pending                         bool
	input                                      ui.Handle
}

func (a *App) beginExplorerEntry(parent, target string, directory bool) {
	a.initFiles()
	s := a.files
	if s.mutating {
		return
	}
	if target != "" && s.editor.HasDirtyUnder(target) {
		a.Notify(a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path."), true)
		return
	}
	root := a.Model.Workspace
	rel, err := filepath.Rel(root, parent)
	if root == "" || err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return
	}
	s.entrySerial++
	e := &explorerEntry{root: root, parent: parent, target: target, directory: directory, epoch: s.menuEpoch, serial: s.entrySerial}
	if target != "" {
		e.value = filepath.Base(target)
		relative, err := filepath.Rel(root, target)
		if err != nil {
			return
		}
		e.relative = filepath.ToSlash(relative)
		for i, r := range workspace.TreeRows(s.tree, s.expanded, s.query, s.badges) {
			if r.Path == e.relative {
				s.list.ScrollIntoView(i)
				break
			}
		}
	}
	s.entry = e

}
func (a *App) cancelExplorerEntry() {
	a.files.entry = nil
	a.files.treeHandle.Focus()
}
func (a *App) explorerEntryCurrent(e *explorerEntry) bool {
	return e != nil && a.files.entry == e && e.root == a.Model.Workspace && e.epoch == a.files.menuEpoch
}
func (a *App) commitExplorerEntry(e *explorerEntry) {
	s := a.files
	if !a.explorerEntryCurrent(e) || e.pending || s.mutating {
		return
	}
	if e.target != "" && !a.syncFileBuffers() {
		return
	}
	name, err := workspace.ValidateName(e.value)
	if err != nil {
		e.err = err.Error()
		return
	}
	if e.target != "" && s.editor.HasDirtyUnder(e.target) {
		e.err = a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path.")
		return
	}
	receipt := a.captureFileRename(e.target)
	e.pending, s.mutating = true, true
	a.Async(func() func() {
		var path string
		var err error
		if e.target != "" {
			path, err = workspace.RenamePath(e.target, name)
		} else if e.directory {
			path, err = workspace.CreateDirectory(e.parent, name)
		} else {
			path, err = workspace.CreateFile(e.parent, name)
		}
		if err == nil && e.target != "" {
			s.editor.RenameUnder(e.target, path)
		}
		return func() {
			e.pending, s.mutating = false, false
			current := a.explorerEntryCurrent(e)
			if err == nil && e.target != "" {
				focus := s.focusPending
				a.applyFileRename(receipt, path)
				if !current {
					s.focusPending = focus
				}
			}
			if !current {
				return
			}
			if err != nil {
				e.err = err.Error()
				return
			}
			s.entry = nil
			a.RefreshFiles()
			if e.target == "" && !e.directory {
				a.OpenFile(path)
			} else {
				s.treeHandle.Focus()
			}
		}
	})
}
func (a *App) explorerEntryInput(c *ui.Context, e *explorerEntry) {
	before := e.value
	in := ui.TextInputBase(c.Key(e.serial), &e.value).Bind(&e.input).Label("Explorer entry name").Grow(1).MinWidth(0).Height(24).FontSize(a.Theme.Px(12)).AutoFocus().Disabled(e.pending)
	if in.Changed() || before != e.value {
		e.err = ""
	}
	if in.Shortcut(0, ui.KeyEscape) {
		a.cancelExplorerEntry()
		return
	}
	if in.Submitted() {
		a.commitExplorerEntry(e)
	}
}
func (a *App) explorerEntryTop(c *ui.Context) {
	s, t := a.files, a.Theme
	e := s.entry
	if e != nil && e.target == "" {
		ui.Row(c).Label("Explorer new entry row").Height(28).Shrink(0).PaddingX(8).Gap(6).Children(func() {
			icon := "icons/file-plus.svg"
			if e.directory {
				icon = "icons/folder-plus.svg"
			}
			a.icon(c, icon, t.Px(14), t.TextMuted)
			if rel, err := filepath.Rel(e.root, e.parent); err == nil && rel != "." {
				ui.Text(c, filepath.ToSlash(rel)+"/").MaxLines(1).Shrink(1).TextColor(t.TextMuted)
			}
			a.explorerEntryInput(c, e)
		})
	}
	message := ""
	if e != nil && e.err != "" {
		message = e.err
	}
	if message != "" {
		ui.Text(c, message).Label("Explorer entry error").Padding(6).TextColor(t.Error)
	}
}
