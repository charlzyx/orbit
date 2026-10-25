package mcpui

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestSourceMCPUnboxedEmptyGeometryAndLoading(t *testing.T) {
	s := stateFixture(t)
	s.Servers = nil
	s.checkedAt = time.Now()
	tt := testerFor(s, 1230, 922)
	tt.Frame()
	saveSourceCapture(t, tt, "mcp-empty")
	empty, ok := tt.Find("MCP empty state")
	glyph, _ := tt.Find("MCP empty glyph")
	title, _ := tt.Find("No MCP servers configured.")
	if !ok || empty.X != 159 || empty.W != 912 || math.Abs(float64(empty.Y-173)) > 2 || glyph.W != 48 || glyph.H != 48 || math.Abs(float64(glyph.Y-empty.Y-24)) > .1 || math.Abs(float64(glyph.X+24-615)) > .1 || math.Abs(float64(title.Y-glyph.Y-54)) > .1 {
		t.Fatal("source MCP empty geometry", empty, glyph, title)
	}
	t.Logf("MCP empty=%+v glyph=%+v title=%+v", empty, glyph, title)
	// No border or filled panel may be painted around the empty state.
	bg := tt.Image().RGBAAt(10, 180)
	if got := tt.Image().RGBAAt(int(empty.X), int(empty.Y+10)); got != bg {
		t.Fatal("MCP empty state unexpectedly boxed", got, bg)
	}
	s.Busy = true
	s.Probing = true
	tt.Frame()
	if tt.HasText("MCP empty glyph") || tt.HasText("Connect an MCP server") || !tt.HasText("Checking connections…") {
		t.Fatal("probing claimed no configured servers")
	}
	s.Busy = false
	s.Probing = false
	tt.Frame()
	tt.ClickAt(empty.X+empty.W/2, title.Y+title.H+6+12*1.618034+6+14)
	if s.Editor == nil {
		t.Fatal("pointer add did not open editor")
	}
	tt.Key(0, ui.KeyEscape)
	if s.Editor != nil {
		t.Fatal("Escape did not cancel empty-state add")
	}
}

func TestSourceMCPNoMatchDoesNotChangeConfiguredServers(t *testing.T) {
	s := stateFixture(t)
	s.Selected = "test-server"
	tt := testerFor(s, 1230, 922)
	if err := tt.Click("Search MCP servers"); err != nil {
		t.Fatal(err)
	}
	tt.Type("no-match-fixture")
	if !tt.HasText("No matching servers") || !tt.HasText("Try a server name, a command, or a URL.") || tt.HasText("MCP empty state") || len(s.Servers) != 1 {
		t.Fatal("filtered configuration rendered as truly empty")
	}
	tt.Key(0, ui.KeyEnter)
	if s.Editor != nil {
		t.Fatal("search Enter edited a hidden server")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	if !tt.HasText("test-server") || len(s.Servers) != 1 {
		t.Fatal("clearing search lost configuration")
	}
}

func saveSourceCapture(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	out := os.Getenv("ORBIT_SOURCE_EMPTY_CAPTURE_DIR")
	if out == "" {
		return
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, name+"-cpu.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}
