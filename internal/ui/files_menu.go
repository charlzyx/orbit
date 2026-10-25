package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"path/filepath"
)

// Rust explorer/panel.rs:1238–1406: directories toggle before create/rename;
// both file and directory menus share default-app, reveal and path actions.
func (a *App) explorerRowMenu(c *ui.Context, m *ui.Menu, row workspace.TreeRow, path string) {
	if row.Directory {
		label := a.text("explorer.expand", "Expand")
		if row.Expanded {
			label = a.text("explorer.collapse", "Collapse")
		}
		if m.Item(label).Chosen() {
			a.files.expanded[row.Path] = !a.files.expanded[row.Path]
		}
		m.Separator()
		if m.Item("New file…").Chosen() {
			a.beginExplorerEntry(path, "", false)
		}
		if m.Item("New folder…").Chosen() {
			a.beginExplorerEntry(path, "", true)
		}
	} else if m.Item("Open").Chosen() {
		a.OpenFile(path)
	}
	m.Separator()
	if m.Item("Rename…").Chosen() {
		a.beginExplorerEntry(filepath.Dir(path), path, row.Directory)
	}
	if m.Item("Delete…").Chosen() {
		a.deleteFile(path)
	}
	m.Separator()
	if m.Item(a.text("explorer.open_default", "Open in default app")).Chosen() {
		a.Dispatch(Action{Kind: "OpenDefaultPath", Value: path})
	}
	if m.Item("Reveal in file manager").Chosen() {
		a.Dispatch(Action{Kind: "RevealPath", Value: path})
	}
	m.Separator()
	if m.Item("Copy path").Chosen() {
		c.Services().WriteClipboard(path)
	}
	if m.Item("Copy relative path").Chosen() {
		if relative, err := filepath.Rel(a.Model.Workspace, path); err == nil {
			c.Services().WriteClipboard(filepath.ToSlash(relative))
		}
	}
}

// Schedule the launcher outside the view handler and return errors on the owner
// thread. The native adapter may marshal back to its required main thread; this
// is not a promise that the OS launcher itself never blocks. Capture the selected
// path before starting, so later workspace changes cannot replace the target.
func (a *App) openPathDefault(path string, open func(string) error) {
	a.Async(func() func() {
		err := open(path)
		return func() { a.save(err) }
	})
}
