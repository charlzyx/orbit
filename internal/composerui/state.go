package composerui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
)

type LoadState string

const (
	Idle      LoadState = "idle"
	Loading   LoadState = "loading"
	Ready     LoadState = "ready"
	Failed    LoadState = "failed"
	Cancelled LoadState = "cancelled"
)

type FileState struct {
	State   LoadState
	Error   string
	Skipped int
	Capped  bool
}
type State struct {
	Open         bool
	Highlighted  int
	Entries      []Entry
	Trigger      Trigger
	HasTrigger   bool
	Dismissed    bool
	Files        FileState
	CatalogError string
}
type scanReply struct {
	generation uint64
	result     ScanResult
}
type pendingKey struct {
	text   string
	delta  int
	accept bool
}

// Controller is owned by the UI thread. Its worker communicates solely through
// a private result channel; no worker mutates native or app state.
type Controller struct {
	popupAnchor              [4]float32
	fontsChecked             bool
	nerdFont                 string
	state                    State
	workspace, home, session string
	commands                 []Command
	raw                      []json.RawMessage
	files                    []string
	filesCached              bool
	generation               uint64
	cancel                   context.CancelFunc
	replies                  chan scanReply
	scan                     ScanFunc
	closed                   bool
	lastKey                  string
	pending                  []pendingKey
	pendingCaret             int
	restoreCaret             bool
	text                     string
	caret                    int
	composing, focused       bool
}

func New(scan ScanFunc) *Controller { return &Controller{scan: scan} }
func (c *Controller) SetContext(workspace, home, session string) {
	if c.closed {
		return
	}
	if workspace == c.workspace && home == c.home && session == c.session {
		return
	}
	changedWorkspace := workspace != c.workspace
	c.workspace, c.home, c.session = workspace, home, session
	c.commands = nil
	c.raw = nil
	c.state.CatalogError = ""
	c.pending = nil
	c.restoreCaret = false
	c.lastKey = ""
	c.state.Open = false
	c.state.Dismissed = false
	c.state.Highlighted = 0
	if changedWorkspace {
		c.cancelScan()
		c.files = nil
		c.filesCached = false
		c.state.Files = FileState{State: Idle}
	}
}
func (c *Controller) SetCommands(raw []json.RawMessage) { c.SetCatalog(c.session, raw) }

// SetCatalog rejects a delayed command response from a different active session.
func (c *Controller) SetCatalog(session string, raw []json.RawMessage) bool {
	if c.closed || session != c.session {
		return false
	}
	equal := len(raw) == len(c.raw)
	if equal {
		for i := range raw {
			if !bytes.Equal(raw[i], c.raw[i]) {
				equal = false
				break
			}
		}
	}
	if equal {
		return true
	}
	c.raw = make([]json.RawMessage, len(raw))
	for i := range raw {
		c.raw[i] = bytes.Clone(raw[i])
	}
	c.commands = ParseCommands(raw, c.workspace, c.home)
	c.state.CatalogError = ""
	c.recompute()
	return true
}
func (c *Controller) SetCatalogError(session, message string) bool {
	if c.closed || session != c.session {
		return false
	}
	c.state.CatalogError = message
	c.recompute()
	return true
}
func (c *Controller) cancelScan() {
	c.generation++
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}
func (c *Controller) ensureFiles() {
	if c.closed || c.filesCached || c.state.Files.State == Loading {
		return
	}
	if c.workspace == "" {
		c.filesCached = true
		c.state.Files = FileState{State: Ready}
		return
	}
	if c.replies == nil {
		c.replies = make(chan scanReply, 8)
	}
	scan := c.scan
	if scan == nil {
		scan = ScanWorkspace
	}
	c.generation++
	gen := c.generation
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	root := c.workspace
	replies := c.replies
	c.state.Files = FileState{State: Loading}
	go func() {
		r := scan(ctx, root)
		select {
		case replies <- scanReply{gen, r}:
		case <-ctx.Done():
		}
	}()
}

// Poll is nonblocking and may be called from an app heartbeat or native build.
func (c *Controller) Poll() bool {
	changed := false
	for c.replies != nil {
		select {
		case reply := <-c.replies:
			if c.closed || reply.generation != c.generation {
				continue
			}
			changed = true
			c.cancel = nil
			c.files = append([]string(nil), reply.result.Files...)
			c.filesCached = true
			c.state.Files = FileState{State: Ready, Skipped: reply.result.Skipped, Capped: reply.result.Capped}
			if reply.result.Err != nil {
				c.state.Files.State = Failed
				c.state.Files.Error = reply.result.Err.Error()
				if errors.Is(reply.result.Err, context.Canceled) {
					c.state.Files.State = Cancelled
				}
			}
		default:
			if changed {
				c.recompute()
			}
			return changed
		}
	}
	return changed
}
func (c *Controller) RefreshFiles() {
	if c.closed {
		return
	}
	c.cancelScan()
	c.files = nil
	c.filesCached = false
	c.state.Files = FileState{State: Idle}
	if c.state.HasTrigger && c.state.Trigger.Kind == At {
		c.ensureFiles()
	}
	c.recompute()
}
func (c *Controller) Close() {
	if c.closed {
		return
	}
	c.cancelScan()
	c.closed = true
	c.pending = nil
	c.state.Open = false
	c.state.Files.State = Cancelled
}
func (c *Controller) Sync(text string, caret int, focused, composing bool) {
	c.Poll()
	c.text = text
	c.caret = min(max(caret, 0), len([]rune(text)))
	c.focused = focused
	c.composing = composing
	trigger, ok := DetectTrigger(text, c.caret)
	key := ""
	if ok {
		key = string(rune(trigger.Kind)) + ":" + trigger.Query
	}
	if key != c.lastKey {
		c.state.Dismissed = false
	}
	c.lastKey = key
	c.state.Trigger = trigger
	c.state.HasTrigger = ok
	if ok && trigger.Kind == At {
		c.ensureFiles()
	}
	c.recompute()
}
func (c *Controller) recompute() {
	c.state.Entries = nil
	if c.state.HasTrigger {
		if c.state.Trigger.Kind == Slash {
			c.state.Entries = FilterEntries(c.state.Trigger.Query, nil, c.commands, Limit)
		} else {
			c.state.Entries = FilterEntries(c.state.Trigger.Query, c.files, nil, Limit)
		}
	} else {
		c.state.Highlighted = 0
	}
	c.state.Highlighted = min(c.state.Highlighted, max(0, len(c.state.Entries)-1))
	c.state.Open = !c.closed && !c.composing && !c.state.Dismissed && len(c.state.Entries) > 0
}
func (c *Controller) Snapshot() State {
	s := c.state
	s.Entries = append([]Entry(nil), s.Entries...)
	return s
}
func (c *Controller) Dismiss() bool {
	if !c.state.Open {
		return false
	}
	c.state.Open = false
	c.state.Dismissed = true
	c.pending = nil
	return true
}
func (c *Controller) Move(delta int) bool {
	n := len(c.state.Entries)
	if !c.state.Open || n == 0 {
		return false
	}
	c.state.Highlighted = ((c.state.Highlighted+delta)%n + n) % n
	return true
}

// Commit returns a text edit only after current text/caret/focus/IME validation.
func (c *Controller) Commit(index int) (string, int, bool) {
	if !c.state.Open || !c.focused || c.composing || index < 0 || index >= len(c.state.Entries) {
		return "", 0, false
	}
	t, ok := DetectTrigger(c.text, c.caret)
	if !ok || t != c.state.Trigger {
		return "", 0, false
	}
	entry := c.state.Entries[index]
	if entry.Kind != t.Kind {
		return "", 0, false
	}
	text := []rune(c.text)
	insert := []rune(entry.Insertion())
	next := string(text[:t.Start]) + string(insert) + string(text[t.End:])
	caret := t.Start + len(insert)
	c.Dismiss()
	c.lastKey = ""
	c.text = next
	c.caret = caret
	return next, caret, true
}

func (c *Controller) IsOpen() bool { return c.state.Open }
