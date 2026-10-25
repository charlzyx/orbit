// Package terminal owns Orbit's pooled login-shell terminal tabs. MyGo's native
// Ghostty view supplies the VT grid, input, selection, clipboard and paste.
package terminal

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charlzyx/orbit/internal/runtimepath"
	ghostty "github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

const (
	DefaultHeight   float32 = 260
	MinimumHeight   float32 = 96
	HeaderHeight    float32 = 32
	SplitterHeight  float32 = 6
	TabMaximumWidth float32 = 190
)

// Options can be constructed before the native event loop is ready. OnChange
// must be safe from background goroutines (e.g. Window.Invalidate).
type Options struct {
	Workspace        string
	Font             ghostty.Font
	Theme, DarkTheme *ghostty.Theme
	ReduceMotion     bool
	Height           float32
	Command          []string // Empty selects Orbit's default login shell.
	Env              []string
	OnChange         func()
	OnError          func(error)
	OnHeightChange   func(float32)
	OnBell           func()
	OpenLink         func(string) bool
}

type Profile string

const (
	Zsh   Profile = "zsh"
	Bash  Profile = "bash"
	Fish  Profile = "fish"
	Other Profile = "shell"
)

func ShellProfile(path string) Profile {
	name := strings.ToLower(filepath.Base(strings.ReplaceAll(path, "\\", "/")))
	name = strings.TrimSuffix(name, ".exe")
	switch name {
	case "zsh":
		return Zsh
	case "bash":
		return Bash
	case "fish":
		return Fish
	default:
		return Other
	}
}
func (p Profile) Icon() string {
	if p == Other {
		return "icons/terminal.svg"
	}
	return "icons/" + string(p) + ".svg"
}

// DefaultCommand follows Orbit's original shell selection, including cmd.exe
// on Windows. POSIX shells get -l so their PATH/profile is the user's own.
func DefaultCommand() []string {
	if runtime.GOOS == "windows" {
		shell := os.Getenv("COMSPEC")
		if shell == "" {
			shell = "cmd.exe"
		}
		return []string{shell}
	}
	for _, path := range []string{os.Getenv("SHELL"), "/bin/zsh", "/bin/bash", "/bin/sh"} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return []string{path, "-l"}
		}
	}
	return []string{"/bin/sh", "-l"}
}

type TabInfo struct {
	ID                                                       uint64
	Label, Title, Directory, CurrentDirectory, Tooltip, Icon string
	Profile                                                  Profile
	Loading, Exited                                          bool
	ExitCode                                                 int
	Error                                                    string
}
type State struct {
	Open       bool
	Height     float32
	Active     int
	Tabs       []TabInfo
	Restarting bool
}
type tab struct {
	TabInfo
	term       *ghostty.Terminal
	generation uint64
	command    []string
}
type Panel struct {
	mu                  sync.Mutex
	opts                Options
	workspace           string
	tabs                []*tab
	active              int
	open, closed, focus bool
	height              float32
	nextID              uint64
	restartUntil        time.Time
}

func NewPanel(opts Options) *Panel {
	opts.Command = append([]string(nil), opts.Command...)
	opts.Env = append([]string(nil), opts.Env...)
	if len(opts.Command) == 0 {
		opts.Command = DefaultCommand()
	}
	h := opts.Height
	if h == 0 {
		h = DefaultHeight
	}
	h = clampHeight(h)
	if opts.Font.Size == 0 {
		opts.Font.Size = 13
	}
	if opts.Font.LineHeight == 0 {
		opts.Font.LineHeight = 1.4
	}
	return &Panel{opts: opts, workspace: opts.Workspace, height: h}
}
func clampHeight(h float32) float32 {
	if math.IsNaN(float64(h)) || math.IsInf(float64(h), 0) {
		return DefaultHeight
	}
	return max(MinimumHeight, h)
}
func (p *Panel) changed() {
	if p.opts.OnChange != nil {
		p.opts.OnChange()
	}
}
func (p *Panel) failure(err error) {
	if err != nil && p.opts.OnError != nil {
		p.opts.OnError(err)
	}
}
func (p *Panel) SetWorkspace(path string) {
	p.mu.Lock()
	p.workspace = path
	p.mu.Unlock()
	p.changed()
}
func (p *Panel) SetHeight(height float32) {
	h := clampHeight(height)
	p.mu.Lock()
	changed := p.height != h
	p.height = h
	p.mu.Unlock()
	if changed {
		if p.opts.OnHeightChange != nil {
			p.opts.OnHeightChange(h)
		}
		p.changed()
	}
}
func (p *Panel) Height() float32 { p.mu.Lock(); defer p.mu.Unlock(); return p.height }
func (p *Panel) IsOpen() bool    { p.mu.Lock(); defer p.mu.Unlock(); return p.open }
func (p *Panel) TabCount() int   { p.mu.Lock(); defer p.mu.Unlock(); return len(p.tabs) }
func (p *Panel) Snapshot() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	s := State{Open: p.open, Height: p.height, Active: p.active, Restarting: time.Now().Before(p.restartUntil)}
	s.Tabs = make([]TabInfo, len(p.tabs))
	for i, t := range p.tabs {
		s.Tabs[i] = t.TabInfo
		if t.term != nil {
			s.Tabs[i].CurrentDirectory = t.term.Dir()
		}
	}
	return s
}
func (p *Panel) Toggle() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return os.ErrClosed
	}
	if p.open {
		p.open = false
		p.mu.Unlock()
		p.changed()
		return nil
	}
	if len(p.tabs) == 0 {
		p.mu.Unlock()
		return p.NewTab()
	}
	p.open = true
	p.focus = true
	p.mu.Unlock()
	p.changed()
	return nil
}

// Hide leaves every PTY alive; Close is reserved for quitting the app.
func (p *Panel) Hide()         { p.mu.Lock(); p.open = false; p.mu.Unlock(); p.changed() }
func (p *Panel) NewTab() error { return p.newTab(nil) }

// NewCommandTab runs an argv directly in a fresh PTY. It is for explicit
// interactive app actions such as gh auth login, never command-text evaluation.
func (p *Panel) NewCommandTab(command []string) error {
	if len(command) == 0 || command[0] == "" {
		return errors.New("terminal command is empty")
	}
	command = append([]string(nil), command...)
	if !strings.ContainsAny(command[0], "/\\") {
		path := runtimepath.Resolve(command[0])
		if path == "" {
			return fmt.Errorf("terminal program is not installed: %s", command[0])
		}
		command[0] = path
	}
	return p.newTab(command)
}
func (p *Panel) newTab(command []string) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return os.ErrClosed
	}
	dir := p.workspace
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if dir == "" {
		dir = "."
	}
	absolute, err := filepath.Abs(dir)
	if err == nil {
		dir = absolute
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		p.mu.Unlock()
		if err != nil {
			return err
		}
		return fmt.Errorf("terminal workspace is not a directory: %s", dir)
	}
	p.nextID++
	if len(command) == 0 {
		command = append([]string(nil), p.opts.Command...)
	}
	profile := ShellProfile(command[0])
	label := filepath.Base(dir)
	if label == "." || label == string(filepath.Separator) {
		label = dir
	}
	t := &tab{TabInfo: TabInfo{ID: p.nextID, Label: label, Directory: dir, Tooltip: string(profile) + " — " + abbreviateHome(dir), Profile: profile, Icon: profile.Icon(), Loading: true}, generation: 1, command: command}
	p.tabs = append(p.tabs, t)
	p.active = len(p.tabs) - 1
	p.open = true
	p.focus = true
	generation := t.generation
	p.mu.Unlock()
	p.changed()
	go p.start(t, generation)
	return nil
}
func abbreviateHome(path string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		if path == home {
			return "~"
		}
		if strings.HasPrefix(path, home+string(filepath.Separator)) {
			return "~" + strings.TrimPrefix(path, home)
		}
	}
	return path
}
func (p *Panel) containsLocked(t *tab, generation uint64) bool {
	if p.closed || t.generation != generation {
		return false
	}
	for _, v := range p.tabs {
		if v == t {
			return true
		}
	}
	return false
}

// start is off the UI thread: even the first verified libghostty-vt download
// cannot stall a frame. A tab closed during startup never leaks its process.
func (p *Panel) start(t *tab, generation uint64) {
	p.mu.Lock()
	if !p.containsLocked(t, generation) {
		p.mu.Unlock()
		return
	}
	o := p.opts
	o.Command = append([]string(nil), t.command...)
	o.Env = append([]string{"PATH=" + runtimepath.AugmentedPath(filepath.Dir(o.Command[0]))}, o.Env...)
	dir := t.Directory
	p.mu.Unlock()
	term, err := ghostty.New(ghostty.Options{
		Command: o.Command, Dir: dir, Env: o.Env, Font: o.Font, Theme: o.Theme, DarkTheme: o.DarkTheme,
		NoBlink: o.ReduceMotion, Scrollback: 10 << 20, Padding: ghostty.Padding{X: 8, Y: 4},
		ClipboardRead: ghostty.ClipboardDeny, OnBell: o.OnBell, OpenLink: o.OpenLink,
		OnKey: p.HandleKey,
		OnTitle: func(title string) {
			p.mu.Lock()
			valid := p.containsLocked(t, generation)
			if valid {
				t.Title = title
				t.Label = title
				if title == "" {
					t.Label = filepath.Base(t.Directory)
				}
			}
			p.mu.Unlock()
			if valid {
				p.changed()
			}
		},
		OnExit: func(code int) {
			p.mu.Lock()
			valid := p.containsLocked(t, generation)
			if valid {
				t.Exited = true
				t.ExitCode = code
			}
			p.mu.Unlock()
			if valid {
				p.changed()
			}
		},
	})
	p.mu.Lock()
	valid := p.containsLocked(t, generation)
	if valid {
		t.Loading = false
		if err != nil {
			t.Error = err.Error()
			t.Exited = true
			t.ExitCode = -1
		} else {
			// Appearance may have changed while the first library load ran.
			term.SetFont(p.opts.Font)
			term.SetTheme(p.opts.Theme, p.opts.DarkTheme)
			t.term = term
		}
	}
	p.mu.Unlock()
	if !valid {
		if term != nil {
			_ = term.Close()
		}
		return
	}
	p.failure(err)
	p.changed()
}
func (p *Panel) ActiveTerminal() *ghostty.Terminal {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.tabs) == 0 {
		return nil
	}
	return p.tabs[p.active].term
}
func (p *Panel) Select(index int) {
	p.mu.Lock()
	if index < 0 || index >= len(p.tabs) {
		p.mu.Unlock()
		return
	}
	p.active = index
	p.focus = true
	p.mu.Unlock()
	p.changed()
}
func (p *Panel) Cycle(delta int) {
	p.mu.Lock()
	n := len(p.tabs)
	if n < 2 {
		p.mu.Unlock()
		return
	}
	p.active = ((p.active+delta)%n + n) % n
	p.focus = true
	p.mu.Unlock()
	p.changed()
}
func (p *Panel) CloseActive() error {
	p.mu.Lock()
	index := p.active
	p.mu.Unlock()
	return p.CloseTab(index)
}
func (p *Panel) CloseTab(index int) error {
	p.mu.Lock()
	if index < 0 || index >= len(p.tabs) {
		p.mu.Unlock()
		return nil
	}
	t := p.tabs[index]
	p.tabs = append(p.tabs[:index], p.tabs[index+1:]...)
	if len(p.tabs) == 0 {
		p.active = 0
		p.open = false
	} else {
		if index < p.active {
			p.active--
		}
		p.active = min(p.active, len(p.tabs)-1)
		p.focus = true
	}
	p.mu.Unlock()
	var err error
	if t.term != nil {
		err = t.term.Close()
	}
	p.changed()
	return err
}
func (p *Panel) Restart() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return os.ErrClosed
	}
	if len(p.tabs) == 0 {
		p.mu.Unlock()
		return nil
	}
	t := p.tabs[p.active]
	old := t.term
	t.term = nil
	t.generation++
	g := t.generation
	t.Loading = true
	t.Exited = false
	t.ExitCode = 0
	t.Error = ""
	t.Title = ""
	t.Label = filepath.Base(t.Directory)
	p.focus = true
	p.restartUntil = time.Now().Add(650 * time.Millisecond)
	p.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	p.changed()
	go p.start(t, g)
	time.AfterFunc(650*time.Millisecond, p.changed)
	return nil
}
func (p *Panel) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.open = false
	tabs := p.tabs
	p.tabs = nil
	p.mu.Unlock()
	var errs []error
	for _, t := range tabs {
		if t.term != nil {
			errs = append(errs, t.term.Close())
		}
	}
	p.changed()
	return errors.Join(errs...)
}
func (p *Panel) SetTheme(light, dark *ghostty.Theme) {
	p.mu.Lock()
	p.opts.Theme = light
	p.opts.DarkTheme = dark
	var terms []*ghostty.Terminal
	for _, t := range p.tabs {
		if t.term != nil {
			terms = append(terms, t.term)
		}
	}
	p.mu.Unlock()
	for _, t := range terms {
		t.SetTheme(light, dark)
	}
	p.changed()
}
func (p *Panel) SetFont(font ghostty.Font) {
	p.mu.Lock()
	p.opts.Font = font
	var terms []*ghostty.Terminal
	for _, t := range p.tabs {
		if t.term != nil {
			terms = append(terms, t.term)
		}
	}
	p.mu.Unlock()
	for _, t := range terms {
		t.SetFont(font)
	}
	p.changed()
}

// KeyAction is explicit about the OS so its shell-preserving bindings can be
// regression-tested from any build host. Ctrl+Tab remains the shell's key.
type Action uint8

const (
	NoAction Action = iota
	NewTabAction
	CloseTabAction
	NextTabAction
	PreviousTabAction
)

func KeyAction(goos string, mods ui.Modifiers, key ui.Key) Action {
	primary := ui.Ctrl
	tabmods := ui.Ctrl | ui.Shift
	if goos == "darwin" {
		primary = ui.Super
		tabmods = ui.Super
	}
	if mods == tabmods {
		switch key {
		case ui.KeyT:
			return NewTabAction
		case ui.KeyW:
			return CloseTabAction
		}
	}
	if mods == primary|ui.Shift {
		switch key {
		case ui.KeyBracketRight:
			return NextTabAction
		case ui.KeyBracketLeft:
			return PreviousTabAction
		}
	}
	return NoAction
}

// HandleKey belongs only to the focused terminal's OnKey, never the global
// window keymap; in particular it does not swallow Ctrl+W/T or Cmd+J.
func (p *Panel) HandleKey(mods ui.Modifiers, key ui.Key) bool {
	switch KeyAction(runtime.GOOS, mods, key) {
	case NewTabAction:
		p.failure(p.NewTab())
	case CloseTabAction:
		p.failure(p.CloseActive())
	case NextTabAction:
		p.Cycle(1)
	case PreviousTabAction:
		p.Cycle(-1)
	default:
		return false
	}
	return true
}

// View draws only the active grid. The shell owns the tab strip, resize handle
// and error/exit cards, with the constants above matching Orbit's geometry.
func View(c *ui.Context, p *Panel) ui.Element {
	p.mu.Lock()
	if !p.open || len(p.tabs) == 0 || p.tabs[p.active].term == nil {
		p.mu.Unlock()
		return ui.Box(c)
	}
	t := p.tabs[p.active]
	term, id := t.term, t.ID
	focus := p.focus
	p.focus = false
	p.mu.Unlock()
	e := ghostty.View(c, term).Key(id).Fill()
	if focus {
		e.Focus()
	}
	return e
}

// SetReduceMotion changes the default for new tabs and the current cursor
// mode without restarting any shell. Programs retain control of later cursor
// sequences, as required by the terminal protocol.
func (p *Panel) SetReduceMotion(reduce bool) {
	p.mu.Lock()
	if p.opts.ReduceMotion == reduce {
		p.mu.Unlock()
		return
	}
	p.opts.ReduceMotion = reduce
	var terms []*ghostty.Terminal
	for _, t := range p.tabs {
		if t.term != nil {
			terms = append(terms, t.term)
		}
	}
	p.mu.Unlock()
	seq := []byte("\x1b[?12h")
	if reduce {
		seq = []byte("\x1b[?12l")
	}
	for _, t := range terms {
		t.Feed(seq)
	}
	p.changed()
}
