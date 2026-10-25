package terminal

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCustomANSIFrameOwnsControls(t *testing.T) {
	line := "\x1b[31mRED\x1b[0m\x1b]52;c;c2VjcmV0\a\x1b[2J end"
	got := customLine(line)
	if got != "\x1b[31mRED\x1b[0m end" {
		t.Fatal(got)
	}
	frame := string(encodeCustomFrame([]string{line}, json.RawMessage(`{"row":0,"col":2}`), 96, 20))
	if strings.Contains(frame, "52;") || !strings.Contains(frame, "\x1b[1;3H\x1b[?25h") {
		t.Fatal(frame)
	}
	bad := string(encodeCustomFrame(nil, json.RawMessage(`{"row":0}`), 96, 20))
	if strings.Contains(bad, "\x1b[?25h") {
		t.Fatal("invalid cursor visible")
	}
}
func TestLiveCustomRPCGridNoPTY(t *testing.T) {
	requireTerminal(t)
	var mu sync.Mutex
	var inputs []string
	var resizes []string
	p := NewCustom(CustomOptions{OnInput: func(id, data string) error {
		mu.Lock()
		defer mu.Unlock()
		inputs = append(inputs, id+":"+data)
		return nil
	}, OnResize: func(id string, width, height int) error {
		mu.Lock()
		defer mu.Unlock()
		resizes = append(resizes, id)
		return nil
	}})
	defer p.Close()
	p.Update("first", []string{"\x1b[31mCustom color\x1b[0m", "choice"}, 96, 20, json.RawMessage(`{"row":1,"col":2}`))
	until(t, nil, "custom renderer", func() bool { return p.ActiveTerminal() != nil })
	term := p.ActiveTerminal()
	if !strings.Contains(term.Text(), "Custom color") {
		t.Fatal(term.Text())
	}
	tt := ui.NewTester(func(c *ui.Context) { CustomView(c, p).Width(700).Height(350) }, 720, 400)
	tt.Frame()
	tt.Key(0, ui.KeyUp)
	until(t, tt, "custom key", func() bool { mu.Lock(); defer mu.Unlock(); return len(inputs) > 0 && len(resizes) > 0 })
	mu.Lock()
	first := inputs[0]
	mu.Unlock()
	if first != "first:\x1b[A" {
		t.Fatal(first)
	}
	for _, key := range []ui.Key{ui.KeyEscape, ui.KeyTab, ui.KeyEnter} {
		tt.Key(0, key)
	}
	until(t, tt, "custom navigation keys", func() bool { mu.Lock(); defer mu.Unlock(); return len(inputs) >= 4 })
	mu.Lock()
	navigation := append([]string(nil), inputs[1:4]...)
	mu.Unlock()
	for i, want := range []string{"first:\x1b", "first:\t", "first:\r"} {
		if navigation[i] != want {
			t.Fatalf("navigation key %d: %q != %q", i, navigation[i], want)
		}
	}

	p.Update("first", []string{"replaced"}, 96, 20, nil)
	if strings.Contains(term.Text(), "choice") || !strings.Contains(term.Text(), "replaced") {
		t.Fatal(term.Text())
	}
	p.Update("second", []string{"new surface"}, 80, 20, nil)
	until(t, tt, "new custom renderer", func() bool { return p.ActiveTerminal() != nil && p.ActiveTerminal() != term })
	p.ActiveTerminal().Send([]byte("x"))
	until(t, tt, "new surface input", func() bool {
		mu.Lock()
		defer mu.Unlock()
		for _, s := range inputs {
			if s == "second:x" {
				return true
			}
		}
		return false
	})
}
