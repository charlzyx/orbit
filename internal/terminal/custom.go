package terminal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
)

const CustomColumns = 96
const CustomRows = 20

// CustomOptions connects a native ANSI grid to Pi's custom-UI RPC surface.
// No process or PTY is started. Input/resize/error callbacks run off-thread.
type CustomOptions struct {
	Font ghostty.Font
	// RowHeight is the row pitch in DIPs. The source custom surface uses
	// theme.code_px(12) text and theme.code_px(18) rows, independently of
	// the bottom terminal's font preference. Zero uses Font.Size * 1.5.
	RowHeight        float32
	Theme, DarkTheme *ghostty.Theme
	ReduceMotion     bool
	OnChange         func()
	OnInput          func(id, data string) error
	OnResize         func(id string, width, height int) error
	OnError          func(error)
}
type customFrame struct {
	lines         []string
	width, height int
	cursor        json.RawMessage
}
type CustomPanel struct {
	mu              sync.Mutex
	options         CustomOptions
	id              string
	frame           customFrame
	term            *ghostty.Terminal
	loading, closed bool
	err             error
	scale           float32
	generation      uint64
	encoded         []byte
}
type customConn struct {
	id      string
	options CustomOptions
	done    chan struct{}
	once    sync.Once
}

func (c *customConn) Read([]byte) (int, error) { <-c.done; return 0, io.EOF }
func (c *customConn) Write(p []byte) (int, error) {
	select {
	case <-c.done:
		return 0, io.ErrClosedPipe
	default:
	}
	if c.options.OnInput != nil {
		if e := c.options.OnInput(c.id, string(p)); e != nil {
			c.fail(e)
			return 0, e
		}
	}
	return len(p), nil
}
func (c *customConn) Resize(width, height int) error {
	select {
	case <-c.done:
		return io.ErrClosedPipe
	default:
	}
	if c.options.OnResize != nil {
		e := c.options.OnResize(c.id, width, min(height, CustomRows))
		if e != nil {
			c.fail(e)
		}
		return e
	}
	return nil
}
func (c *customConn) Close() error { c.once.Do(func() { close(c.done) }); return nil }
func (c *customConn) fail(e error) {
	if c.options.OnError != nil {
		c.options.OnError(e)
	}
}
func NewCustom(o CustomOptions) *CustomPanel {
	if o.Font.Size <= 0 {
		o.Font.Size = 12
	}
	if o.RowHeight <= 0 {
		o.RowHeight = o.Font.Size * 1.5
	}
	o.Font = fontAtRowHeight(o.Font, o.RowHeight)
	return &CustomPanel{options: o, scale: 1}
}

// Update takes an immutable copy of a server frame. A new ID owns a fresh
// connection so queued keys from a dismissed surface cannot target another.
func (p *CustomPanel) Update(id string, lines []string, width, height int, cursor json.RawMessage) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	changed := p.id != id
	p.frame = customFrame{lines: append([]string(nil), lines...), width: max(1, min(width, 1000)), height: max(1, min(height, 1000)), cursor: bytes.Clone(cursor)}
	if width <= 0 {
		p.frame.width = CustomColumns
	}
	if height <= 0 {
		p.frame.height = CustomRows
	}
	if len(p.frame.lines) > 2000 {
		p.frame.lines = p.frame.lines[:2000]
	}
	if changed {
		old := p.term
		p.term = nil
		p.id = id
		p.err = nil
		p.encoded = nil
		p.generation++
		p.loading = id != ""
		generation, o := p.generation, p.options
		p.mu.Unlock()
		if old != nil {
			_ = old.Close()
		}
		if id != "" {
			go p.start(id, generation, o)
		}
		return
	}
	p.paintLocked()
	p.mu.Unlock()
}
func (p *CustomPanel) start(id string, generation uint64, o CustomOptions) {
	conn := &customConn{id: id, options: o, done: make(chan struct{})}
	t, e := ghostty.New(ghostty.Options{Conn: conn, ResizeByConn: true, Font: o.Font, Theme: o.Theme, DarkTheme: o.DarkTheme, NoBlink: o.ReduceMotion, Scrollback: -1, ImageLimit: -1, NoClipboardWrite: true, ClipboardRead: ghostty.ClipboardDeny, OpenLink: func(string) bool { return true }})
	p.mu.Lock()
	if p.closed || p.generation != generation {
		p.mu.Unlock()
		if t != nil {
			_ = t.Close()
		} else {
			_ = conn.Close()
		}
		return
	}
	p.loading = false
	p.err = e
	p.term = t
	if e == nil {
		t.SetFont(p.options.Font)
		t.SetTheme(p.options.Theme, p.options.DarkTheme)
		if p.options.ReduceMotion {
			t.Feed([]byte("\x1b[?12l"))
		} else {
			t.Feed([]byte("\x1b[?12h"))
		}
		p.paintLocked()
	}
	p.mu.Unlock()
	if e != nil {
		_ = conn.Close()
		if o.OnError != nil {
			o.OnError(e)
		}
	}
	if o.OnChange != nil {
		o.OnChange()
	}
}
func (p *CustomPanel) paintLocked() {
	if p.term == nil {
		return
	}
	frame := p.frame
	rows := max(frame.height, len(frame.lines))
	encoded := encodeCustomFrame(frame.lines, frame.cursor, frame.width, rows)
	if bytes.Equal(encoded, p.encoded) {
		return
	}
	p.encoded = encoded
	p.term.Resize(frame.width, rows)
	p.term.Feed(encoded)
}
func (p *CustomPanel) ActiveTerminal() *ghostty.Terminal {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.term
}
func (p *CustomPanel) Close() error {
	p.mu.Lock()
	p.closed = true
	p.generation++
	t := p.term
	p.term = nil
	p.mu.Unlock()
	if t != nil {
		return t.Close()
	}
	return nil
}

// custom SGR is display data, not arbitrary terminal commands. Retain ANSI
// styles but discard OSC clipboard/title/links, cursor moves and control bytes.
// The host owns row placement and the separate, 0-based cursor field.
func customLine(line string) string {
	var out strings.Builder
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			start := i
			i++
			if i >= len(line) {
				break
			}
			kind := line[i]
			i++
			if kind == '[' {
				for i < len(line) && !(line[i] >= 0x40 && line[i] <= 0x7e) {
					i++
				}
				if i < len(line) {
					if line[i] == 'm' && i-start < 128 {
						out.WriteString(line[start : i+1])
					}
					i++
				}
				continue
			}
			if kind == ']' || kind == 'P' || kind == '_' || kind == '^' {
				for i < len(line) {
					if line[i] == 7 {
						i++
						break
					}
					if line[i] == 0x1b && i+1 < len(line) && line[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
				continue
			}
			continue
		}
		ch := line[i]
		if ch == '\t' {
			out.WriteByte(ch)
		} else if ch >= 32 && ch != 127 {
			out.WriteByte(ch)
		}
		i++
	}
	return out.String()
}
func encodeCustomFrame(lines []string, cursor json.RawMessage, width, height int) []byte {
	var b strings.Builder
	b.WriteString("\x1b[?2026h\x1b[?25l\x1b[?7l\x1b[0m\x1b[2J\x1b[H")
	for i, line := range lines {
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[0m%s\x1b[K", i+1, customLine(line))
	}
	var caret struct{ Row, Col *int }
	if json.Unmarshal(cursor, &caret) == nil && caret.Row != nil && caret.Col != nil && *caret.Row >= 0 && *caret.Col >= 0 && *caret.Row < height && *caret.Col < width {
		fmt.Fprintf(&b, "\x1b[%d;%dH\x1b[?25h", *caret.Row+1, *caret.Col+1)
	}
	b.WriteString("\x1b[0m\x1b[?2026l")
	return []byte(b.String())
}

// CustomView includes scrollable overflow because Pi's row budget is a hint.
// The host owns the surrounding modal header and explicit cancel response.
func CustomView(c *ui.Context, p *CustomPanel) ui.Element {
	p.mu.Lock()
	t, loading, e, rows, pitch, scale, id := p.term, p.loading, p.err, max(p.frame.height, len(p.frame.lines)), p.options.RowHeight, p.scale, p.id
	p.mu.Unlock()
	if t == nil {
		label := "No custom UI surface"
		if loading {
			label = "Loading terminal renderer…"
		}
		if e != nil {
			label = e.Error()
		}
		return ui.Text(c, label)
	}
	height := float32(rows)*gridRowHeight(pitch, scale) + 8
	services := c.Services()
	// The public frame context has no scale accessor. Read the actual scale
	// from our own paint node (never replace Ghostty's painter), and rebuild
	// once on monitor/scale changes so the last row is not clipped.
	return ui.Scroll(c).Draw(func(painter *ui.Painter, _ ui.Rect) {
		p.mu.Lock()
		changed := p.scale != painter.Scale()
		p.scale = painter.Scale()
		p.mu.Unlock()
		if changed {
			services.Invalidate()
		}
	}).Children(func() { ghostty.View(c, t).Key(id).Height(height).AlignSelf(ui.Stretch).AutoFocus() })
}

// SetAppearance updates an already-open surface. Call when preferences change,
// not each frame; the native view invalidates after font/theme updates.
func (p *CustomPanel) SetAppearance(font ghostty.Font, light, dark *ghostty.Theme, reduce bool) {
	p.mu.Lock()
	font = fontAtRowHeight(font, p.options.RowHeight)
	p.options.Font = font
	p.options.Theme = light
	p.options.DarkTheme = dark
	changedMotion := p.options.ReduceMotion != reduce
	p.options.ReduceMotion = reduce
	t := p.term
	p.mu.Unlock()
	if t != nil {
		t.SetFont(font)
		t.SetTheme(light, dark)
		if changedMotion {
			if reduce {
				t.Feed([]byte("\x1b[?12l"))
			} else {
				t.Feed([]byte("\x1b[?12h"))
			}
		}
	}
}

// SetRowHeight changes the source custom grid's DIP pitch without restarting
// its RPC connection. Use with SetAppearance when the editor scale changes.
func (p *CustomPanel) SetRowHeight(height float32) {
	if height <= 0 {
		return
	}
	p.mu.Lock()
	p.options.RowHeight = height
	p.options.Font = fontAtRowHeight(p.options.Font, height)
	t, font, change := p.term, p.options.Font, p.options.OnChange
	p.mu.Unlock()
	if t != nil {
		t.SetFont(font)
	}
	if change != nil {
		change()
	}
}
