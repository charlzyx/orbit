package workspace

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const AutosaveDelay = 500 * time.Millisecond
const maxUndoStates = 512

var ErrDocumentNotOpen = errors.New("document is not open")
var ErrDocumentChanged = errors.New("document changed while saving; close again after saving")

type DocumentID uint64

type Document struct {
	ID           DocumentID  `json:"id"`
	Conflict     bool        `json:"conflict"`
	Content      FileContent `json:"content"`
	Text         string      `json:"text"`
	Dirty        bool        `json:"dirty"`
	Saving       bool        `json:"saving"`
	LastError    string      `json:"lastError,omitempty"`
	Preview      bool        `json:"preview"`
	loadSequence uint64
	saved        string
	undo         []string
	redo         []string
}

// Editor owns open tabs independently of whether Files is currently visible.
// Methods are safe from UI and autosave goroutines. Changes is coalesced.
type Editor struct {
	namespaceEpoch uint64
	reservation    *renameReservation
	renamePath     func(string, string) (string, error)
	nextID         DocumentID
	ioMu           sync.Mutex
	writeFile      func(string, string) error
	readFile       func(string) FileContent
	mu             sync.Mutex
	tabs           []*Document
	active         int
	Changes        chan struct{}
	delay          time.Duration
	closed         bool
	autosave       *time.Timer
	saveEpoch      uint64
}

func NewEditor() *Editor {
	return &Editor{active: -1, Changes: make(chan struct{}, 1), delay: AutosaveDelay, writeFile: SaveExisting}
}
func (e *Editor) notify() {
	select {
	case e.Changes <- struct{}{}:
	default:
	}
}
func documentCopy(d *Document) Document {
	c := *d
	c.Content.Lines = append([]string(nil), d.Content.Lines...)
	c.Content.ImageBytes = append([]byte(nil), d.Content.ImageBytes...)
	c.undo = nil
	c.redo = nil
	return c
}
func (e *Editor) find(path string) (int, *Document) {
	path = filepath.Clean(path)
	for i, d := range e.tabs {
		if filepath.Clean(d.Content.Path) == path {
			return i, d
		}
	}
	return -1, nil
}
func (e *Editor) Open(path string) (Document, error) { return e.open(path, true, nil) }

// OpenInactive performs the background load/register half of opening a file.
// Only a UI result that still matches its workspace/navigation generation may
// Activate it. A stale completion can retain a tab, but cannot steal selection.
func (e *Editor) OpenInactive(path string) (Document, error) { return e.open(path, false, nil) }

// ReopenClosed shares the normal reader but never adopts an existing occupant.
func (e *Editor) ReopenClosed(receipt CloseReceipt) (Document, error) {
	if receipt.ID == 0 || receipt.Path == "" || !filepath.IsAbs(receipt.Path) {
		return Document{}, ErrDocumentNotOpen
	}
	return e.open(receipt.Path, false, &receipt)
}
func (e *Editor) open(path string, activate bool, vacant *CloseReceipt) (Document, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Document{}, err
	}
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return Document{}, ErrEditorClosed
	}
	if e.namespaceBusy(abs) {
		e.mu.Unlock()
		return Document{}, ErrNamespaceBusy
	}
	if vacant != nil {
		if _, original := e.findID(vacant.ID); original != nil || vacant.ID > e.nextID || vacant.Path != abs {
			e.mu.Unlock()
			return Document{}, ErrReopenOccupied
		}
	}
	epoch := e.namespaceEpoch
	if i, d := e.find(abs); d != nil {
		if vacant != nil {
			e.mu.Unlock()
			return Document{}, ErrReopenOccupied
		}
		if activate {
			e.active = i
		}
		copy := documentCopy(d)
		e.mu.Unlock()
		return copy, nil
	}
	e.mu.Unlock()
	content := e.loadFile(abs)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return Document{}, ErrEditorClosed
	}
	if epoch != e.namespaceEpoch || e.namespaceBusy(abs) {
		return Document{}, ErrNamespaceChanged
	}
	if i, d := e.find(abs); d != nil {
		if vacant != nil {
			return Document{}, ErrReopenOccupied
		}
		if activate {
			e.active = i
		}
		return documentCopy(d), nil
	}
	e.nextID++
	d := &Document{ID: e.nextID, Content: content, Text: content.Text, saved: content.Text, Preview: content.Mode == FileMarkdown}
	e.tabs = append(e.tabs, d)
	if activate {
		e.active = len(e.tabs) - 1
	}
	e.notify()
	return documentCopy(d), nil
}
func (e *Editor) Documents() []Document {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Document, len(e.tabs))
	for i, d := range e.tabs {
		out[i] = documentCopy(d)
	}
	return out
}
func (e *Editor) Active() (Document, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.active < 0 || e.active >= len(e.tabs) {
		return Document{}, false
	}
	return documentCopy(e.tabs[e.active]), true
}
func (e *Editor) Activate(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	i, d := e.find(path)
	if d == nil {
		return ErrDocumentNotOpen
	}
	e.active = i
	e.notify()
	return nil
}

// ActivateID cannot select a new document that reused a closed path.
func (e *Editor) ActivateID(id DocumentID) (Document, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	i, d := e.findID(id)
	if d == nil {
		return Document{}, ErrDocumentNotOpen
	}
	e.active = i
	e.notify()
	return documentCopy(d), nil
}
func (e *Editor) SetPreview(path string, preview bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, d := e.find(path)
	if d == nil {
		return ErrDocumentNotOpen
	}
	d.Preview = preview
	e.notify()
	return nil
}
func (e *Editor) schedule(_ *Document) {
	if e.autosave != nil {
		e.autosave.Stop()
	}
	e.saveEpoch++
	epoch := e.saveEpoch
	e.autosave = time.AfterFunc(e.delay, func() { _ = e.flushAll(epoch, true) })
}
func (e *Editor) SetText(path, text string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, d := e.find(path)
	if d == nil {
		return ErrDocumentNotOpen
	}
	return e.setText(d, text)
}

// SetTextID preserves native edit ownership across path-only rename changes.
func (e *Editor) SetTextID(id DocumentID, text string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, d := range e.tabs {
		if d.ID == id {
			return e.setText(d, text)
		}
	}
	return ErrDocumentNotOpen
}
func (e *Editor) setText(d *Document, text string) error {
	if !d.Content.Editable() {
		return ErrReadOnly
	}
	if d.Text == text {
		return nil
	}
	d.undo = append(d.undo, d.Text)
	if len(d.undo) > maxUndoStates {
		d.undo = d.undo[1:]
	}
	d.redo = nil
	d.Text = text
	d.Dirty = text != d.saved
	d.LastError = ""
	e.schedule(d)
	e.notify()
	return nil
}
func (e *Editor) Undo(path string) error { return e.history(path, false) }
func (e *Editor) Redo(path string) error { return e.history(path, true) }
func (e *Editor) history(path string, redo bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, d := e.find(path)
	if d == nil {
		return ErrDocumentNotOpen
	}
	if !d.Content.Editable() {
		return ErrReadOnly
	}
	from, to := &d.undo, &d.redo
	if redo {
		from, to = &d.redo, &d.undo
	}
	if len(*from) == 0 {
		return nil
	}
	*to = append(*to, d.Text)
	d.Text = (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]
	d.Dirty = d.Text != d.saved
	e.schedule(d)
	e.notify()
	return nil
}

// saveDocument holds the I/O gate but never the state lock during disk work.
// New keystrokes stay responsive and are left dirty when an older write ends.
func (e *Editor) saveDocument(d *Document) error {
	e.mu.Lock()
	path := d.Content.Path
	_, live := e.find(path)
	if live != d || !d.Dirty {
		e.mu.Unlock()
		return nil
	}
	if !d.Content.Editable() {
		e.mu.Unlock()
		return ErrReadOnly
	}
	text := d.Text
	d.Saving = true
	e.notify()
	write := e.writeFile
	if write == nil {
		write = SaveExisting
	}
	e.mu.Unlock()
	err := write(path, text)
	e.mu.Lock()
	defer e.mu.Unlock()
	_, live = e.find(path)
	if live != d {
		return err
	}
	d.Saving = false
	if err != nil {
		d.LastError = err.Error()
		e.notify()
		return err
	}
	d.saved = text
	d.Dirty = d.Text != text
	d.LastError = ""
	d.Conflict = false
	d.Content.Text = text
	d.Content.Bytes = int64(len(text))
	d.Content.Lines = strings.Split(text, "\n")
	if strings.HasSuffix(text, "\n") {
		d.Content.Lines = d.Content.Lines[:len(d.Content.Lines)-1]
	}
	e.notify()
	return nil
}
func (e *Editor) Save(path string) error {
	e.mu.Lock()
	_, d := e.find(path)
	var id DocumentID
	if d != nil {
		id = d.ID
	}
	e.mu.Unlock()
	return e.SaveID(id)
}
func (e *Editor) SaveID(id DocumentID) error {
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return ErrEditorClosed
	}
	_, d := e.findID(id)
	e.mu.Unlock()
	if d == nil {
		return ErrDocumentNotOpen
	}
	return e.saveDocument(d)
}
func (e *Editor) flushAll(epoch uint64, autosave bool) error {
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	if e.closed || (autosave && epoch != e.saveEpoch) {
		e.mu.Unlock()
		return nil
	}
	tabs := append([]*Document(nil), e.tabs...)
	e.mu.Unlock()
	var errs []error
	for _, d := range tabs {
		if err := e.saveDocument(d); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
func (e *Editor) SaveAll() error { return e.flushAll(0, false) }

// A nonzero receipt is produced only by an actual successful ID removal.
type CloseReceipt struct {
	ID   DocumentID
	Path string
}

func (e *Editor) Close(path string) error {
	e.mu.Lock()
	_, d := e.find(path)
	var id DocumentID
	if d != nil {
		id = d.ID
	}
	e.mu.Unlock()
	if id == 0 {
		return nil
	}
	return e.CloseID(id)
}
func (e *Editor) CloseID(id DocumentID) error { _, err := e.CloseIDReceipt(id); return err }
func (e *Editor) CloseIDReceipt(id DocumentID) (CloseReceipt, error) {
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	_, d := e.findID(id)
	e.mu.Unlock()
	if d == nil {
		return CloseReceipt{}, nil
	}
	if err := e.saveDocument(d); err != nil {
		return CloseReceipt{}, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	i, live := e.findID(id)
	if live != d {
		return CloseReceipt{}, nil
	}
	if d.Dirty {
		return CloseReceipt{}, ErrDocumentChanged
	}
	receipt := CloseReceipt{d.ID, d.Content.Path}
	e.tabs = append(e.tabs[:i], e.tabs[i+1:]...)
	e.namespaceEpoch++
	if i < e.active {
		e.active--
	}
	if e.active >= len(e.tabs) {
		e.active = len(e.tabs) - 1
	}
	e.notify()
	return receipt, nil
}
func (e *Editor) CloseAll() error {
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	tabs := append([]*Document(nil), e.tabs...)
	e.mu.Unlock()
	var errs []error
	for _, d := range tabs {
		if err := e.saveDocument(d); err != nil {
			errs = append(errs, err)
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	for _, d := range e.tabs {
		if d.Dirty {
			return ErrDocumentChanged
		}
	}
	e.tabs = nil
	e.active = -1
	e.closed = true
	if e.autosave != nil {
		e.autosave.Stop()
	}
	e.saveEpoch++
	e.notify()
	return nil
}
func (e *Editor) Reload(path string) error {
	e.mu.Lock()
	_, d := e.find(path)
	var id DocumentID
	if d != nil {
		id = d.ID
	}
	e.mu.Unlock()
	return e.ReloadID(id)
}
func (e *Editor) ReloadID(id DocumentID) error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return ErrEditorClosed
	}
	_, d := e.findID(id)
	if d == nil {
		e.mu.Unlock()
		return ErrDocumentNotOpen
	}
	currentPath := d.Content.Path
	if e.namespaceBusy(currentPath) {
		e.mu.Unlock()
		return ErrNamespaceBusy
	}
	epoch := e.namespaceEpoch
	d.loadSequence++
	sequence := d.loadSequence
	e.mu.Unlock()
	content := e.loadFile(currentPath)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrEditorClosed
	}
	_, live := e.findID(id)
	if live != d {
		return ErrDocumentNotOpen
	}
	if epoch != e.namespaceEpoch || d.Content.Path != currentPath || e.namespaceBusy(currentPath) {
		return ErrNamespaceChanged
	}
	if d.loadSequence != sequence {
		return ErrReadSuperseded
	}
	if content.Error == "" && content.Text == d.Text && content.Mode != FileImage && content.Mode != FileBinary && content.Mode != FileTooLarge {
		// A self-save echo (or matching external save) must not reset undo,
		// cursor or scroll state in the native adapter.
		d.Content = content
		d.saved = content.Text
		d.Dirty = false
		d.Conflict = false
		e.notify()
		return nil
	}
	if d.Dirty || d.Saving {
		d.Conflict = content.Error != "" || content.Text != d.saved
		e.notify()
		return nil
	}
	d.Content = content
	d.Text = content.Text
	d.saved = content.Text
	d.undo = nil
	d.redo = nil
	d.Conflict = false
	e.notify()
	return nil
}
func pathUnder(path, prefix string) bool {
	rel, err := filepath.Rel(filepath.Clean(prefix), filepath.Clean(path))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func (e *Editor) HasDirtyUnder(prefix string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, d := range e.tabs {
		if d.Dirty && pathUnder(d.Content.Path, prefix) {
			return true
		}
	}
	return false
}
func (e *Editor) RenameUnder(old, new string) {
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.namespaceEpoch++
	for _, d := range e.tabs {
		if pathUnder(d.Content.Path, old) {
			rel, _ := filepath.Rel(old, d.Content.Path)
			d.Content.Path = filepath.Join(new, rel)
			d.Content.Display = filepath.Base(d.Content.Path)
		}
	}
	e.notify()
}

// CloseUnder is called only after the external deletion has succeeded. It
// deliberately drops affected buffers without recreating deleted files.
func (e *Editor) CloseUnder(prefix string) {
	e.ioMu.Lock()
	defer e.ioMu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.namespaceEpoch++ // Fence held reads even when deletion had zero registered tabs.
	out := e.tabs[:0]
	for _, d := range e.tabs {
		if !pathUnder(d.Content.Path, prefix) {
			out = append(out, d)
		}
	}
	e.tabs = out
	if e.active >= len(out) {
		e.active = len(out) - 1
	}
	e.notify()
}

func (e *Editor) loadFile(path string) FileContent {
	if e.readFile != nil {
		return e.readFile(path)
	}
	return LoadFile(path)
}
