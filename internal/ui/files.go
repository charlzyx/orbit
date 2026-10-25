package orbitui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charlzyx/orbit/internal/composerui"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/markdown"
	_ "golang.org/x/image/tiff"
)

type filesState struct {
	tree                  workspace.TreeSnapshot
	expanded              map[string]bool
	query, loaded, err    string
	hidden, loading, wrap bool
	list                  ui.ListState
	editor                *workspace.Editor
	states                map[string]*editor.State
	versions              map[string]uint64
	images                map[string]ui.ImageSource
	badges                map[string]string
	selected              int
	active, treeWorkspace string
	resizeDirty           bool
	generation            uint64
	menuEpoch             uint64
	entrySerial           uint64
	entry                 *explorerEntry
	treeHandle            ui.Handle
	desired               string
	openGeneration        uint64
	mutating              bool
	focusPending          string
	handle                ui.Handle
	closing               map[string]*fileClose
}

type fileClose struct {
	reopen     bool
	root       string
	generation uint64
}

func (a *App) initFiles() {
	if a.files != nil {
		return
	}
	a.files = &filesState{selected: -1, expanded: map[string]bool{}, editor: workspace.NewEditor(), versions: map[string]uint64{}, states: map[string]*editor.State{}, images: map[string]ui.ImageSource{}}
}
func (a *App) RefreshFiles() {
	a.initFiles()
	s := a.files
	if s.treeWorkspace != a.Model.Workspace {
		s.entry = nil
		s.menuEpoch++ // A -> B -> A must not revive an old native menu choice.
		s.treeWorkspace = a.Model.Workspace
		s.tree = workspace.TreeSnapshot{Root: a.Model.Workspace}
		s.expanded = map[string]bool{}
		s.selected, s.active, s.err = -1, "", ""
		s.list.ScrollTo(0, ui.Start)
	}
	s.generation++
	generation := s.generation
	if s.loading {
		return
	}
	root := a.Model.Workspace
	if root == "" {
		return
	}
	s.loading = true
	hidden := s.hidden
	a.Async(func() func() {
		tree, e := workspace.Snapshot(context.Background(), root, hidden)
		badges := map[string]string{}
		rows, _ := (workspace.Git{Dir: root}).Status(context.Background())
		for _, r := range rows {
			badges[r.Path] = string([]byte{r.Index, r.Worktree})
		}
		return func() {
			s.loading = false
			if generation != s.generation || root != a.Model.Workspace || hidden != s.hidden {
				a.RefreshFiles()
				return
			}
			s.loaded = root
			s.tree = tree
			s.badges = badges
			s.err = ""
			if e != nil {
				s.err = e.Error()
			}
		}
	})
}
func (a *App) OpenFile(path string) {
	a.initFiles()
	s := a.files
	if absolute, e := filepath.Abs(path); e == nil {
		path = absolute
	}
	root := a.Model.Workspace
	s.desired = path
	s.openGeneration++
	generation := s.openGeneration
	a.Navigate("files")
	if operation := s.closing[path]; operation != nil {
		operation.reopen, operation.root = true, root
		operation.generation = generation
		return
	}
	// Source open(existing) activates the existing buffer synchronously. A
	// background copy can arrive after another keystroke and rewind that draft.
	for _, d := range s.editor.Documents() {
		if d.Content.Path == path {
			a.save(s.editor.Activate(path))
			if s.states[path] == nil {
				a.syncDocument(d)
			}
			s.focusPending = path
			return
		}
	}
	a.Async(func() func() {
		_, e := s.editor.OpenInactive(path)
		return func() {
			if generation != s.openGeneration || root != a.Model.Workspace {
				return
			}
			if e != nil {
				a.Notify(e.Error(), true)
				return
			}
			if e := s.editor.Activate(path); e != nil {
				a.save(e)
				return
			}
			// A watcher may have reloaded the registered document while the
			// UI completion was queued. Use its current state, not the worker's
			// stale value copy from before that reload.
			if d, ok := s.editor.Active(); ok && d.Content.Path == path {
				a.syncDocument(d)
			}
			s.focusPending = path
		}
	})
}
func (a *App) syncDocument(d workspace.Document) {
	s := a.files
	p := d.Content.Path
	viewText := d.Text
	if d.Content.Truncated {
		viewText = strings.Join(d.Content.Lines, "\n")
	}
	st := s.states[p]
	if st == nil {
		b := editor.NewBuffer(viewText)
		b.SetLanguage(editor.LanguageFor(p))
		st = &editor.State{Buffer: b}
		s.states[p] = st
	} else if st.Buffer.String() != viewText {
		st.Buffer.Set(viewText)
	}
	s.versions[p] = st.Buffer.Version()
	if d.Content.Mode == workspace.FileImage {
		var im ui.ImageSource
		var e error
		if strings.EqualFold(filepath.Ext(d.Content.Path), ".svg") {
			im, e = ui.ParseSVG(d.Content.ImageBytes)
		} else {
			im, e = ui.DecodeBitmap(d.Content.ImageBytes)
		}
		if e == nil {
			s.images[p] = im
		} else {
			a.Notify(e.Error(), true)
		}
	}
}

// Pinned explorer/panel.rs: width includes the left gutter, not only the card.
const explorerDefaultWidth float32 = 248
const explorerMinWidth float32 = 172
const explorerMaxWidth float32 = 460

func (a *App) explorerPaneWidth() float32 {
	w := a.ExplorerWidth
	if w <= 0 {
		w = explorerDefaultWidth
	}
	return max(explorerMinWidth, min(w, explorerMaxWidth))
}
func (a *App) explorer(c *ui.Context) {
	a.initFiles()
	t := a.Theme
	ui.Column(c.Key("explorer")).Label("Explorer dock").Width(a.explorerPaneWidth()).Shrink(0).FillHeight().MinHeight(0).Padding(0, 0, 0, t.Space(12)).Children(func() {
		ui.Column(c.Key("explorer-card")).Label("Explorer card").Fill().MinHeight(0).Background(t.Surface).Border(1, t.Border).Radius(t.Px(12)).Clip().Children(func() { a.fileTree(c) })
		h := ui.Box(c.Key("explorer-resize")).Label("Resize explorer").Absolute().Left(9).Top(0).Bottom(0).Width(6).Cursor(ui.CursorResizeEW)
		if h.Hovered() {
			h.Background(t.Accent.Alpha(.4))
		}
		if dx, _, dragging := h.Dragged(); dragging && dx != 0 {
			// app/view.rs uses distance from the actual pointer to the dock's
			// right edge, not a delta from the initial width or a main reserve.
			viewport, _ := c.Size()
			x, _, _ := h.PointerPosition()
			width := max(explorerMinWidth, min(viewport-h.Bounds().X-x, explorerMaxWidth))
			if width != a.explorerPaneWidth() {
				a.ExplorerWidth = width
				a.Layout.ExplorerWidth = float64(width)
				a.files.resizeDirty = true
				c.AnimationFrame()
			}
		}
		if _, _, ended := h.DragEnded(); ended && a.files.resizeDirty {
			a.files.resizeDirty = false
			a.save(a.Options.Store.SaveLayout(a.Layout))
		}
	})
}
func (a *App) explorerIcon(c *ui.Context, label, icon string, loading bool) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Label(label).Tooltip(label).Size(t.Px(22), t.Px(22)).Shrink(0).Radius(t.Px(6)).Center()
	if b.Hovered() {
		b.Background(t.Hover)
	}
	b.Children(func() {
		if loading {
			ui.Spinner(c).Size(t.Px(14), t.Px(14))
		} else {
			a.icon(c, icon, t.Px(14), t.TextMuted)
		}
	})
	return b
}
func (a *App) fileTree(c *ui.Context) {
	a.initFiles()
	s, t := a.files, a.Theme
	if s.treeWorkspace != a.Model.Workspace || (s.loaded != a.Model.Workspace && !s.loading) {
		a.RefreshFiles()
	}
	ui.Row(c).Label("Explorer toolbar").Height(40).Shrink(0).Padding(0, 6, 0, t.Space(12)).Gap(2).AlignItems(ui.Center).Children(func() {
		title := filepath.Base(a.Model.Workspace)
		if a.Model.Workspace == "" {
			title = a.text("explorer.files", "Files")
		}
		ui.Text(c, title).Grow(1).MinWidth(0).MaxLines(1).FontSize(t.Px(12)).FontWeight(500)
		if a.explorerIcon(c, "New file", "icons/file-plus.svg", false).Clicked() {
			a.beginExplorerEntry(a.explorerCreateParent(), "", false)
		}
		if a.explorerIcon(c, "New folder", "icons/folder-plus.svg", false).Clicked() {
			a.beginExplorerEntry(a.explorerCreateParent(), "", true)
		}
		if a.explorerIcon(c, "Refresh files", "icons/refresh.svg", s.loading).Clicked() {
			a.RefreshFiles()
		}
		if a.explorerIcon(c, "Collapse all", "icons/chevron-up.svg", false).Clicked() {
			s.expanded = map[string]bool{}
		}
		hiddenLabel, hiddenIcon := "Show hidden files", "icons/eye-off.svg"
		if s.hidden {
			hiddenLabel, hiddenIcon = "Hide hidden files", "icons/eye.svg"
		}
		if a.explorerIcon(c, hiddenLabel, hiddenIcon, false).Clicked() {
			s.hidden = !s.hidden
			a.RefreshFiles()
		}
		if a.explorerIcon(c, "Hide explorer", "icons/x.svg", false).Clicked() {
			a.ExplorerVisible = false
		}
	})
	ui.Row(c).Label("Explorer filter row").Height(t.Px(36)).Shrink(0).PaddingX(t.Px(10)).Gap(t.Space(6)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		a.icon(c, "icons/search.svg", t.Px(14), t.TextMuted)
		input := ui.TextInputBase(c, &s.query).Placeholder("Filter files…").Label("Filter files").Grow(1).MinWidth(0)
		if input.Shortcut(0, ui.KeyEscape) {
			s.query = ""
		}
		if s.query != "" && a.explorerIcon(c, "Clear file filter", "icons/x.svg", false).Clicked() {
			s.query = ""
		}
	})
	a.explorerEntryTop(c)
	if s.err != "" {
		ui.Text(c, s.err).Padding(10).TextColor(t.Error)
	}
	active := ""
	if d, ok := s.editor.Active(); ok {
		if relative, err := filepath.Rel(a.Model.Workspace, d.Content.Path); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			active = filepath.ToSlash(relative)
		}
	}
	changedActive := active != s.active
	if changedActive {
		s.active = active
		for parent := filepath.ToSlash(filepath.Dir(active)); active != "" && parent != "." && parent != ""; parent = filepath.ToSlash(filepath.Dir(parent)) {
			s.expanded[parent] = true
		}
	}
	rows := workspace.TreeRows(s.tree, s.expanded, s.query, s.badges)
	if e := s.entry; e != nil && e.target != "" {
		found := false
		for _, r := range rows {
			if r.Path == e.relative {
				found = true
				break
			}
		}
		if !found && !s.loading {
			a.cancelExplorerEntry()
		}
	}
	if s.selected >= len(rows) {
		s.selected = -1
	}
	if changedActive {
		for i, r := range rows {
			if r.Path == active {
				s.list.ScrollIntoView(i)
				break
			}
		}
	}
	tree := ui.Column(c.Key("explorer-tree")).Bind(&s.treeHandle).Label("Explorer files").Focusable().Grow(1).MinHeight(0)
	// Source cursor navigation does not select the active file. In particular,
	// an unopened tree starts with no cursor, and Left never jumps to a parent.
	if s.entry == nil && tree.Shortcut(0, ui.KeyEscape) {
		s.query = ""
	}
	if len(rows) > 0 && s.entry == nil {
		for _, k := range []struct {
			key ui.Key
			row int
		}{
			{ui.KeyDown, min(s.selected+1, len(rows)-1)}, {ui.KeyUp, max(s.selected-1, 0)},
			{ui.KeyHome, 0}, {ui.KeyEnd, len(rows) - 1},
		} {
			if tree.Shortcut(0, k.key) {
				s.selected = k.row
				s.list.ScrollIntoView(k.row)
			}
		}
		if tree.Shortcut(ui.Cmd, ui.KeyRight) {
			for _, d := range s.tree.Dirs {
				s.expanded[d] = true
			}
		}
		if tree.Shortcut(ui.Cmd, ui.KeyLeft) {
			s.expanded = map[string]bool{}
		}
		if s.selected >= 0 && s.selected < len(rows) {
			r := rows[s.selected]
			if tree.Shortcut(0, ui.KeyRight) && r.Directory {
				s.expanded[r.Path] = true
			}
			if tree.Shortcut(0, ui.KeyLeft) && r.Directory {
				s.expanded[r.Path] = false
			}
			if tree.Shortcut(0, ui.KeyEnter) || tree.Shortcut(0, ui.KeySpace) {
				a.activateExplorerRow(r)
			}
		}
	}
	tree.Children(func() {
		s.list.Key = func(i int) any { return rows[i].Path }
		s.list.Label = func(i int) string { return rows[i].Name }
		// Native Selected paints a blue keyboard cursor and owns different
		// activation semantics. Keep virtualization, but paint/route the source
		// cursor and active file separately on our sole tree.
		s.list.Selected = nil
		ui.List(c.Key(fmt.Sprintf("file-tree:%d:%s", s.menuEpoch, s.tree.Root)), &s.list, len(rows), func(i int) {
			r := rows[i]
			path := filepath.Join(s.tree.Root, filepath.FromSlash(r.Path))
			ui.Row(c).FillWidth().PaddingX(t.Space(4)).Children(func() {
				if e := s.entry; e != nil && e.relative == r.Path && e.target != "" {
					ui.Row(c).Label("Explorer rename row " + r.Path).Height(28).Grow(1).MinWidth(0).PaddingX(8).Gap(6).AlignItems(ui.Center).Background(t.Surface).Children(func() {
						if r.Directory {
							a.icon(c, "icons/folder.svg", t.Px(14), t.TextMuted)
						} else {
							a.explorerFileBadge(c, r.Path)
						}
						a.explorerEntryInput(c, e)
					})
					return
				}
				indent := float32(23 + r.Depth*14)
				if r.Directory {
					indent = float32(7 + r.Depth*14)
				}
				b := ui.ButtonBase(c.Key(r.Path)).Label("Explorer row "+r.Path).Height(28).Grow(1).MinWidth(0).Padding(0, t.Space(8), 0, indent).Radius(t.Px(6)).Gap(t.Space(6)).AlignItems(ui.Center).Justify(ui.Start)
				isActive := r.Path == s.active
				if isActive {
					b.Background(t.Selection)
				} else if i == s.selected {
					b.Background(t.overlay())
				} else if b.Hovered() {
					b.Background(t.Hover)
				}
				b.Children(func() {
					if r.Directory {
						arrow := "icons/chevron-right.svg"
						if r.Expanded {
							arrow = "icons/chevron-down.svg"
						}
						a.icon(c, arrow, t.Px(10), t.TextMuted)
						a.icon(c, "icons/folder.svg", t.Px(14), t.TextMuted)
					} else {
						a.explorerFileBadge(c, r.Path)
					}
					color := t.TextSecondary
					if isActive {
						color = t.Text
					}
					name := ui.Text(c, r.Name).FontSize(t.Px(12)).TextColor(color).Grow(1).MinWidth(0).MaxLines(1)
					if r.Directory {
						name.FontWeight(500)
					}
					if r.GitStatus != "" {
						badge, color := a.explorerStatusBadge(r.GitStatus)
						ui.Text(c, badge).FontSize(t.Px(12)).FontWeight(600).TextColor(color).Shrink(0)
					}
				})
				if b.Clicked() {
					if s.entry != nil {
						a.cancelExplorerEntry()
						return
					}
					s.selected = i
					tree.Focus()
					a.activateExplorerRow(r)
					c.AnimationFrame()
				}
				if b.RightClicked() {
					if s.entry != nil {
						a.cancelExplorerEntry()
					}
					s.selected = i
					tree.Focus()
				}
				b.ContextMenu(func(m *ui.Menu) { a.explorerRowMenu(c, m, r, path) })
			})
		}).Grow(1).MinHeight(0).PaddingY(t.Space(4))
	})
	ui.Row(c).Label("Explorer footer").Height(26).Shrink(0).PaddingX(t.Space(12)).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
		label, color := fmt.Sprintf("%d files", len(s.tree.Files)), t.TextMuted
		if s.tree.Truncated {
			label = fmt.Sprintf("%d+ files (truncated)", len(s.tree.Files))
			color = t.Warning
		}
		ui.Text(c, label).FontSize(t.Px(12)).TextColor(color)
	})
}
func (a *App) activateExplorerRow(r workspace.TreeRow) {
	if r.Directory {
		a.files.expanded[r.Path] = !a.files.expanded[r.Path]
	} else {
		a.OpenFile(filepath.Join(a.files.tree.Root, filepath.FromSlash(r.Path)))
	}
}
func (a *App) explorerCreateParent() string {
	s := a.files
	rows := workspace.TreeRows(s.tree, s.expanded, s.query, s.badges)
	if s.selected >= 0 && s.selected < len(rows) {
		r := rows[s.selected]
		if r.Directory {
			return filepath.Join(a.Model.Workspace, filepath.FromSlash(r.Path))
		}
		return filepath.Dir(filepath.Join(a.Model.Workspace, filepath.FromSlash(r.Path)))
	}
	return a.Model.Workspace
}
func (a *App) explorerFileBadge(c *ui.Context, path string) {
	t, badge := a.Theme, composerui.FileBadge(path)
	hex := badge.Light
	if t.Dark {
		hex = badge.Dark
	}
	color := ui.RGB(uint8(hex>>16), uint8(hex>>8), uint8(hex))
	ui.Box(c).Label("Explorer file badge "+path).Size(17, 17).Shrink(0).Radius(t.Px(4)).Background(color.Alpha(.16)).Center().Children(func() {
		ui.Text(c, badge.Label).Label("Explorer file badge text " + path).FontSize(t.Px(7.5)).FixedLineHeight(t.Px(8)).FontWeight(600).TextColor(color)
	})
}
func (a *App) explorerStatusBadge(status string) (string, ui.Color) {
	t := a.Theme
	switch {
	case status == "??":
		return "U", t.Success
	case strings.Contains(status, "U"):
		return "!", t.Error
	case strings.Contains(status, "D"):
		return "D", t.Error
	case strings.ContainsAny(status, "RC"):
		return "R", t.AccentText
	case strings.HasPrefix(status, "A"):
		return "A", t.Success
	default:
		return "M", t.Warning
	}
}
func (a *App) createFile(parent string, directory bool) {
	title := "New file"
	if directory {
		title = "New folder"
	}
	a.Prompt(title, "", func(name string) {
		a.Async(func() func() {
			var p string
			var e error
			if directory {
				p, e = workspace.CreateDirectory(parent, name)
			} else {
				p, e = workspace.CreateFile(parent, name)
			}
			return func() {
				a.save(e)
				if e == nil {
					a.RefreshFiles()
					if !directory {
						a.OpenFile(p)
					}
				}
			}
		})
	})
}
func (a *App) renameFile(path string) {
	if a.files.editor.HasDirtyUnder(path) {
		a.Notify(a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path."), true)
		return
	}
	a.renameFilePrompt(path)
}
func (a *App) renameFilePrompt(path string) {
	a.Prompt("Rename", filepath.Base(path), func(name string) {
		if a.files.mutating || a.files.editor.HasDirtyUnder(path) {
			a.Notify(a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path."), true)
			return
		}
		a.files.mutating = true
		a.Async(func() func() {
			p, e := workspace.RenamePath(path, name)
			if e == nil {
				a.files.editor.RenameUnder(path, p)
			}
			return func() {
				a.files.mutating = false
				a.save(e)
				if e == nil {
					a.reconcileFileRename(path, p)
					a.RefreshFiles()
				}
			}
		})
	})
}

// Source FileViewer::reconcile_rename rewrites paths on existing editor
// entities. Preserve native buffers, selections, undo/redo and scroll state,
// including every unrelated tab; a successful rename is not a reload.
func (a *App) reconcileFileRename(old, renamed string) {
	if filepath.Clean(old) == filepath.Clean(renamed) {
		return
	}
	s := a.files
	remap := func(path string) (string, bool) {
		rel, err := filepath.Rel(old, path)
		if path == "" || err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return path, false
		}
		return filepath.Join(renamed, rel), true
	}
	for path, state := range s.states {
		if next, changed := remap(path); changed {
			s.states[next] = state
			delete(s.states, path)
			if d, ok := s.editor.Active(); ok && d.Content.Path == next {
				s.focusPending = next
			}
		}
	}
	for path, version := range s.versions {
		if next, changed := remap(path); changed {
			s.versions[next] = version
			delete(s.versions, path)
		}
	}
	for path, image := range s.images {
		if next, changed := remap(path); changed {
			s.images[next] = image
			delete(s.images, path)
		}
	}
	if next, changed := remap(s.desired); changed {
		s.desired = next
		s.openGeneration++ // Invalidate a queued open of the old path.
	}
	if next, changed := remap(s.focusPending); changed {
		s.focusPending = next
	}
}

func (a *App) deleteFile(path string) {
	if a.files.editor.HasDirtyUnder(path) {
		a.Notify(a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path."), true)
		return
	}
	detail := "This permanently deletes the file or directory and its contents. This cannot be undone."
	if workspace.DeleteIsRecoverable() {
		detail = "This moves the file or directory to Trash."
	}
	a.Confirm("Delete "+filepath.Base(path)+"?", detail, func() {
		if a.files.mutating || a.files.editor.HasDirtyUnder(path) {
			a.Notify(a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path."), true)
			return
		}
		a.files.mutating = true
		a.Async(func() func() {
			e := workspace.DeletePath(context.Background(), path, true)
			if e == nil {
				a.files.editor.CloseUnder(path)
			}
			return func() {
				a.files.mutating = false
				a.save(e)
				if e == nil {
					a.RefreshFiles()
				}
			}
		})
	})
}
func (a *App) filesPage(c *ui.Context) {
	a.initFiles()
	s := a.files
	// Pinned app/view.rs places only FileViewer in the main column. The
	// ProjectPanel stays in main's right dock, and can be hidden independently.
	// A second tree here both diverged from that layout and mounted s.list
	// twice, panicking as soon as a file was opened from the visible Explorer.
	ui.Column(c).Fill().MinWidth(0).Children(func() { a.fileEditor(c) })
	for {
		select {
		case <-s.editor.Changes:
		default:
			return
		}
	}
}
func (a *App) fileEditor(c *ui.Context) {
	s := a.files
	root := ui.Column(c.Key("file-viewer")).Bind(&s.handle).Focusable().Fill().MinWidth(0)
	root.Children(func() { a.fileEditorContent(c) })
	// Like FileViewer's deferred focus target, previews/non-editor content
	// focus the viewer root; source mode consumes this in the editor below.
	if d, ok := s.editor.Active(); ok && s.focusPending == d.Content.Path {
		if d.Preview || d.Content.Mode == workspace.FileImage || d.Content.Mode == workspace.FileBinary || d.Content.Mode == workspace.FileTooLarge || d.Content.Error != "" {
			root.Focus()
			s.focusPending = ""
		} else {
			// A tab click can change Active after this frame captured the old
			// document. Let the next build focus the newly mounted editor.
			c.AnimationFrame()
		}
	}
	if !a.modalOpen() && s.handle.FocusWithin(c) {
		if c.Shortcut(ui.Cmd|ui.Shift, ui.KeyW) {
			a.Navigate("chat")
		} else if c.Shortcut(ui.Cmd, ui.KeyW) {
			if d, ok := s.editor.Active(); ok {
				a.closeFile(d.Content.Path)
			}
		} else if c.Shortcut(ui.Cmd, ui.KeyS) {
			a.saveFiles()
		}
	}
}

func (a *App) fileEditorContent(c *ui.Context) {
	s := a.files
	t := a.Theme
	docs := s.editor.Documents()
	d, ok := s.editor.Active()
	if !ok {
		a.Empty(c, "Open a file", "Select a file in the explorer to view or edit it.")
		return
	}
	ui.Row(c).Height(40).Gap(t.Space(6)).Padding(0, t.Space(16), 0, 0).Background(t.Surface).Children(func() {
		ui.ScrollHorizontal(c.Key("file-tabs")).Label("File tabs").Grow(1).MinWidth(0).FillHeight().Children(func() {
			ui.Row(c).FillHeight().Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
				for _, tab := range docs {
					tab := tab
					name := tab.Content.Display
					if tab.Dirty {
						name += " •"
					}
					ui.Row(c.Key(tab.Content.Path)).Shrink(0).AlignItems(ui.Center).Children(func() {
						b := a.button(c, name, "", tab.Content.Path == d.Content.Path).Shrink(0)
						if b.Clicked() {
							a.activateFile(tab.Content.Path)
						}
						b.ContextMenu(func(m *ui.Menu) {
							if m.Item("Close").Chosen() {
								a.closeFile(tab.Content.Path)
							}
							if m.Item("Reveal").Chosen() {
								a.Dispatch(Action{Kind: "RevealPath", Value: tab.Content.Path})
							}
						})
						label := "Close " + tab.Content.Display
						if tab.Content.Path == d.Content.Path {
							label = "Close file"
						}
						if a.iconButton(c, label, "icons/x.svg").Size(t.Px(16), t.Px(16)).Clicked() {
							a.closeFile(tab.Content.Path)
						}
					})
				}
			})
		})
		if a.iconButton(c, "Close files", "icons/x.svg").Size(t.Px(22), t.Px(22)).Clicked() {
			a.Navigate("chat")
		}
	})
	ui.Row(c).Height(34).PaddingX(10).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, d.Content.Path).MaxLines(1).FontSize(12).TextColor(t.TextMuted).Grow(1)
		if d.Content.Mode != workspace.FileBinary && d.Content.Mode != workspace.FileTooLarge {
			lines := strings.ReplaceAll(a.text("explorer.lines", "%{count} lines"), "%{count}", fmt.Sprint(len(d.Content.Lines)))
			ui.Text(c, lines).FontSize(12).TextColor(t.TextMuted).Shrink(0)
		}
		if d.Content.Mode == workspace.FileMarkdown {
			if a.button(c, "Preview", "", d.Preview).Clicked() && a.syncFileBuffers() {
				if err := s.editor.SetPreview(d.Content.Path, !d.Preview); err != nil {
					a.save(err)
				} else {
					d.Preview = !d.Preview
					s.focusPending = d.Content.Path
				}
			}
		}
		if d.Content.Editable() {
			if a.button(c, "Save", "icons/save.svg", false).Clicked() {
				a.saveFiles()
			}
		}
		ui.Switch(c, &s.wrap).Label("Wrap lines")
	})
	if d.Conflict {
		ui.Text(c, "File changed on disk; unsaved edits kept.").TextColor(t.Warning).Padding(8)
	}
	if d.LastError != "" {
		ui.Text(c, d.LastError).TextColor(t.Error).Padding(8)
	}
	if d.Content.Error != "" {
		a.Empty(c, "Cannot open file", d.Content.Error)
		return
	}
	switch d.Content.Mode {
	case workspace.FileBinary:
		a.Empty(c, "Binary file", fmt.Sprintf("%s · %d bytes", d.Content.Display, d.Content.Bytes))
		return
	case workspace.FileTooLarge:
		a.Empty(c, "File is too large", "Files larger than 2 MiB are not loaded into the editor.")
		return
	case workspace.FileImage:
		im := s.images[d.Content.Path]
		if im == nil {
			a.syncDocument(d)
			im = s.images[d.Content.Path]
		}
		if im != nil {
			ui.Scroll(c).Grow(1).Children(func() { ui.Image(c, im).MaxWidth(1200) })
		}
		return
	}
	if d.Preview && d.Content.Mode == workspace.FileMarkdown {
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() { markdown.View(c, d.Text, &a.md).Padding(24).MaxWidth(960) })
		return
	}
	st := s.states[d.Content.Path]
	if st == nil {
		a.syncDocument(d)
		st = s.states[d.Content.Path]
	}
	st.Font = ui.Font{Family: a.Fonts.CodeFamily, Size: float32(a.Prefs.EditorFontSize)}
	st.Theme = t.Code()
	st.ReadOnly = !d.Content.Editable() || s.mutating
	st.SoftWrap = s.wrap
	view := editor.View(c.Key(d.Content.Path), st).Bind(&a.surfaceFocus.editor).Label("File editor").Grow(1).FillWidth()
	if s.focusPending == d.Content.Path {
		view.Focus()
		s.focusPending = ""
	}
	if st.Buffer.Version() != s.versions[d.Content.Path] {
		a.save(s.editor.SetText(d.Content.Path, st.Buffer.String()))
		s.versions[d.Content.Path] = st.Buffer.Version()
	}
	pos := st.CursorPos()
	status := fmt.Sprintf("Ln %d, Col %d · %s", pos.Line+1, pos.Col+1, a.fileSaveStatus(d))
	if d.Content.Truncated {
		status += " · Preview truncated"
	}
	ui.Text(c, status).FontSize(12).TextColor(t.TextMuted).Padding(4, 10)
}

func (a *App) activateFile(path string) {
	s := a.files
	s.openGeneration++ // A tab click supersedes an older asynchronous open.
	s.desired = path
	a.save(s.editor.Activate(path))
	s.focusPending = path
}

// A button action can be processed before the editor's build observes its
// latest buffer version. Flush that native buffer before scheduling disk I/O.
func (a *App) syncFileBuffers() bool {
	s := a.files
	for _, d := range s.editor.Documents() {
		st := s.states[d.Content.Path]
		if st == nil || !d.Content.Editable() || st.Buffer.Version() == s.versions[d.Content.Path] {
			continue
		}
		if err := s.editor.SetText(d.Content.Path, st.Buffer.String()); err != nil {
			a.save(err)
			return false
		}
		s.versions[d.Content.Path] = st.Buffer.Version()
	}
	return true
}

func (a *App) saveFiles() {
	if a.files == nil || !a.syncFileBuffers() {
		return
	}
	ed := a.files.editor
	a.Async(func() func() { err := ed.SaveAll(); return func() { a.save(err) } })
}

func (a *App) closeFile(path string) {
	s := a.files
	if !a.syncFileBuffers() {
		return
	}
	if s.closing == nil {
		s.closing = map[string]*fileClose{}
	}
	if s.closing[path] != nil {
		return
	}
	op := &fileClose{}
	s.closing[path] = op
	a.Async(func() func() {
		e := s.editor.Close(path)
		return func() {
			if s.closing[path] != op {
				return
			}
			delete(s.closing, path)
			a.save(e)
			if e == nil {
				stillOpen := false
				for _, d := range s.editor.Documents() {
					if d.Content.Path == path {
						stillOpen = true
						break
					}
				}
				if !stillOpen {
					delete(s.states, path)
					delete(s.images, path)
				}
				if op.reopen && op.root == a.Model.Workspace && op.generation == s.openGeneration && s.desired == path {
					a.OpenFile(path)
					return
				}
				if d, ok := s.editor.Active(); ok {
					s.focusPending = d.Content.Path
				}
				if len(s.editor.Documents()) == 0 && a.page == "files" {
					a.Navigate("chat")
				}
			}
		}
	})
}
func (a *App) featurePage(c *ui.Context) {
	switch a.page {
	case "files":
		a.filesPage(c)
	case "usage":
		a.usagePage(c)
	case "git", "worktrees":
		a.Empty(c, strings.ToUpper(a.page[:1])+a.page[1:], "This native workbench module is not mounted.")
	default:
		a.Empty(c, "Unknown view", a.page)
	}
}
