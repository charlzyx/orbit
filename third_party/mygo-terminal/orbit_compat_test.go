package terminal

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestOrbitHiddenTerminalKeepsStateWithoutInvalidating(t *testing.T) {
	loadLib(t)
	term, err := New(Options{Conn: newPipe(), NoBlink: true})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	visible, frames := true, 0
	tt := ui.NewTester(func(c *ui.Context) {
		frames++
		if visible {
			View(c, term).Fill()
		}
	}, 640, 240)
	term.Feed([]byte("visible\r\n"))
	tt.Advance(time.Second)
	visible = false
	term.SetVisible(false)
	tt.Frame()
	before := frames
	term.Feed([]byte("hidden output remains in VT\r\n"))
	tt.Advance(time.Second)
	if frames != before {
		t.Fatalf("hidden terminal requested %d frames", frames-before)
	}
	if !strings.Contains(term.Text(), "hidden output remains in VT") {
		t.Fatal("hidden terminal stopped consuming output")
	}
	visible = true
	term.SetVisible(true)
	tt.Advance(time.Second)
	if frames <= before {
		t.Fatal("showing terminal did not request a fresh frame")
	}
}

func TestOrbitVisibilityConcurrentOutput(t *testing.T) {
	loadLib(t)
	term, err := New(Options{Conn: newPipe(), NoBlink: true})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	var frames atomic.Int64
	draw := func() { frames.Add(1) }
	term.draw.Store(&draw)
	term.SetVisible(false)
	for range 100 {
		term.Feed([]byte("x"))
	}
	if frames.Load() != 0 {
		t.Fatal("hidden output reached persistent services")
	}
	term.SetVisible(true)
	if frames.Load() != 1 {
		t.Fatal("visibility transition did not invalidate exactly once")
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 1000 {
			term.Feed([]byte("live\r\n"))
		}
	}()
	for range 1000 {
		term.SetVisible(false)
		term.SetVisible(true)
	}
	<-done
	term.SetVisible(false)
	before := frames.Load()
	term.Feed([]byte("settled hidden\r\n"))
	if frames.Load() != before {
		t.Fatal("visibility race left invalidation enabled")
	}
}

func TestOrbitNativeLineBudget(t *testing.T) {
	loadLib(t)
	for _, columns := range []int{1, 10, 80, 240} {
		term, err := New(Options{Conn: newPipe(), ScrollbackLines: 10000})
		if err != nil {
			t.Fatal(err)
		}
		term.Resize(columns, 24)
		var text strings.Builder
		for i := 0; i < 20050; i++ {
			text.WriteString("x\r\n")
		}
		term.Feed([]byte(text.String()))
		t.Logf("%d columns: retained %d physical history rows for configured 10000", columns, term.ScrollbackRows())
		if rows := term.ScrollbackRows(); rows <= 0 || rows > 20050 {
			t.Fatalf("line budget was not applied: %d", rows)
		}
		term.Feed([]byte("\x1b[?1049hALT\x1b[?1049l"))
		if strings.Contains(term.Text(), "ALT") {
			t.Fatal("alternate screen leaked into main history")
		}
		_ = term.Close()
	}
}
