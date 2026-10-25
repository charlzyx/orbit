package composerui

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"golang.org/x/image/font/sfnt"
)

type nativeFixture struct {
	controller               *Controller
	draft, other             string
	submissions              []string
	input                    ui.Handle
	caret, anchor            int
	setCaret                 *[2]int
	style                    Style
	inputBounds, popupAnchor ui.Rect
}

func nativeStyle() Style {
	return Style{FontSize: 14, Density: 100, Dark: true, Surface: ui.Hex("#1b1f26"), Border: ui.Hex("#363c46"), Selection: ui.Hex("#2d4360"), Hover: ui.Hex("#272f3a"), Text: ui.Hex("#eef2f8"), Secondary: ui.Hex("#c7cdd5"), Muted: ui.Hex("#7e8999"), Badge: ui.Hex("#303640"), CommandColor: ui.Hex("#88baff"), FileColor: ui.Hex("#ffc188")}
}
func newNative(t *testing.T) (*nativeFixture, *ui.Tester) {
	t.Helper()
	f := &nativeFixture{controller: New(nil), style: nativeStyle()}
	t.Cleanup(f.controller.Close)
	f.controller.SetCommands(raws(`{"name":"review","description":"Review changes","source":"skill"}`, `{"name":"reset","description":"Reset context","source":"builtin"}`, `{"name":"rename","description":"Rename session","sourceInfo":{"scope":"user"}}`))
	tt := ui.NewTester(f.view, 1000, 700)
	tt.Advance(time.Second)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	return f, tt
}
func (f *nativeFixture) view(c *ui.Context) {
	ui.Column(c).Fill().Padding(20).Children(func() {
		ui.TextInput(c.Key("other"), &f.other).Label("Other input")
		ui.Spacer(c)
		anchor := ui.Column(c.Key("composer")).AlignSelf(ui.Stretch).Padding(14, 16, 10).Radius(18).Border(1, f.style.Border).Background(f.style.Surface)
		var input ui.Element
		anchor.Children(func() {
			input = ui.TextAreaBase(c.Key("message"), &f.draft).Bind(&f.input).Label("Message").Lines(2, 10).AlignSelf(ui.Stretch)
			if f.setCaret != nil {
				input.SetTextSelection(f.setCaret[0], f.setCaret[1])
				f.setCaret = nil
			}
			if f.controller.SyncInput(input, &f.draft) {
				c.AnimationFrame()
			}
			input.OnTextKey(func(k ui.TextKey) bool {
				if f.controller.HandleKey(k) {
					return true
				}
				if k.Key == ui.KeyEnter && (k.Mods == 0 || k.Mods == ui.Alt || k.Mods == ui.Cmd|ui.Shift) {
					f.submissions = append(f.submissions, k.Text())
					k.SetText("")
					return true
				}
				return false
			})
			PaintMentions(input, f.draft, f.style)
			f.anchor, f.caret = input.TextAnchorCaret()
			f.inputBounds = input.Bounds()
			if ui.Button(c, "Send").Clicked() {
				if !input.Composing() && !f.controller.CommitInput(input, &f.draft) {
					f.submissions = append(f.submissions, f.draft)
					f.draft = ""
				}
			}
		})
		f.popupAnchor = anchor.Bounds()
		f.controller.Popup(c, anchor, input, &f.draft, f.style)
	})
}
func TestNativeKeyboardCompletionAndRepeat(t *testing.T) {
	f, tt := newNative(t)
	tt.Type("/re")
	if !f.controller.IsOpen() {
		t.Fatal(f.controller.Snapshot())
	}
	tt.Key(0, ui.KeyUp)
	if f.controller.Snapshot().Highlighted != 2 {
		t.Fatal(f.controller.Snapshot())
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if f.draft != "/review " || len(f.submissions) != 0 || f.caret != 8 {
		t.Fatal(f.draft, f.caret, f.submissions)
	}
	tt.Key(0, ui.KeyEnter)
	if len(f.submissions) != 1 || f.submissions[0] != "/review " {
		t.Fatal(f.submissions)
	}
	tt.Type("/res")
	tt.Key(0, ui.KeyTab)
	if f.draft != "/reset " {
		t.Fatal(f.draft)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("/rev")
	tt.Key(ui.Alt, ui.KeyEnter)
	if f.draft != "/review " || len(f.submissions) != 1 {
		t.Fatal(f.draft, f.submissions)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("/ren")
	tt.Key(ui.Cmd|ui.Shift, ui.KeyEnter)
	if f.draft != "/rename " || len(f.submissions) != 1 {
		t.Fatal(f.draft, f.submissions)
	}
}
func TestNativeEscapeOutsideFocusAndIME(t *testing.T) {
	f, tt := newNative(t)
	tt.Type("/re")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if f.controller.IsOpen() {
		t.Fatal("Escape reopened")
	}
	tt.Type("v")
	if !f.controller.IsOpen() {
		t.Fatal("edit did not rearm")
	}
	if err := tt.Click("Other input"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEnter)
	if f.draft != "/rev" || len(f.submissions) != 0 {
		t.Fatal("other input committed composer")
	}
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Compose("你", 1)
	tt.Key(0, ui.KeyEnter)
	if len(f.submissions) != 0 || f.draft != "/rev" {
		t.Fatal("IME committed", f.draft)
	}
	tt.Compose("", 0)
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("/re")
	tt.Key(ui.Shift, ui.KeyEnter)
	if f.draft != "/re\n" {
		t.Fatal("Shift Enter swallowed", f.draft)
	}
}
func TestNativePointerCommitMidTokenAndTokenPainting(t *testing.T) {
	f, tt := newNative(t)
	tt.Type("/remainder args")
	f.setCaret = &[2]int{3, 3}
	tt.Frame()
	if !f.controller.IsOpen() {
		t.Fatal("midword popup closed")
	}
	if err := tt.Click("/review"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if f.draft != "/review  args" || f.caret != 8 || len(f.submissions) != 0 {
		t.Fatal(f.draft, f.caret, f.submissions)
	}
}
func TestNativeFileCompletionActualFilesystem(t *testing.T) {
	f, tt := newNative(t)
	root := t.TempDir()
	file(t, root, "src/文件.go")
	f.controller.SetContext(root, "", "s")
	tt.Type("你好 @文件尾")
	f.setCaret = &[2]int{6, 6}
	tt.Frame()
	waitScan(t, f.controller)
	tt.Frame()
	tt.Key(0, ui.KeyTab)
	if f.draft != "你好 @src/文件.go 尾" || f.caret != 14 {
		t.Fatal(f.draft, f.caret)
	}
}
func TestNativeExactUndoRedoAndUnicodeCaret(t *testing.T) {
	if !NativeRangeEditsAvailable() {
		if os.Getenv("ORBIT_REQUIRE_COMPOSER_PATCH") == "1" {
			t.Fatal("reviewed MyGo range edit patch required")
		}
		t.Skip("unpatched MyGo SetText merges undo; exact range-edit patch not linked")
	}
	f, tt := newNative(t)
	tt.Type("/remainder 😀 args")
	f.setCaret = &[2]int{3, 3}
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if f.draft != "/review  😀 args" || f.caret != 8 {
		t.Fatal(f.draft, f.caret)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if f.draft != "/remainder 😀 args" || f.caret != 3 {
		t.Fatal("completion undo", f.draft, f.caret)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyZ)
	if f.draft != "/review  😀 args" || f.caret != 8 {
		t.Fatal("completion redo", f.draft, f.caret)
	}
	tt.Type("tail")
	tt.Key(ui.Cmd, ui.KeyZ)
	if f.draft != "/review  😀 args" {
		t.Fatal("typing merged into completion", f.draft)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if f.draft != "/remainder 😀 args" {
		t.Fatal("second undo", f.draft)
	}
}
func TestNativePopupCPUCaptureAndGeometry(t *testing.T) {
	f, tt := newNative(t)
	tt.Type("/re")
	bounds, ok := tt.Find("Composer completions")
	if !ok {
		t.Fatal("popup not found")
	}
	if bounds.W != 960 || bounds.H <= 90 || bounds.Y+bounds.H > f.popupAnchor.Y {
		t.Fatal("popup geometry", bounds, f.popupAnchor)
	}
	if gap := f.popupAnchor.Y - bounds.Y - bounds.H; gap != 5 {
		t.Fatalf("source gap=5, got %.1f", gap)
	}
	if f.style.RowHeight() != 31 {
		t.Fatal("source row height", f.style.RowHeight())
	}
	if dir := os.Getenv("ORBIT_COMPOSER_CAPTURE_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, "composer-slash-native-cpu.png")
		out, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()
		if err := png.Encode(out, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
	if strings.Contains(f.draft, " ") {
		t.Fatal("render changed draft")
	}
}

func TestNativeClosedTabAndEmptyFilter(t *testing.T) {
	f, tt := newNative(t)
	tt.Type("/zz-no-result")
	if f.controller.IsOpen() {
		t.Fatal("empty match popup")
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Send") || f.draft != "/zz-no-result" {
		t.Fatal("closed Tab did not traverse focus", f.draft)
	}
}
func TestMentionColorComplementAndNeutral(t *testing.T) {
	if got := MentionFileColor(ui.RGB(255, 0, 0)); got != ui.RGB(0, 255, 255) {
		t.Fatal(got)
	}
	if got := MentionFileColor(ui.RGB(128, 128, 128)); got.R != 128 || got.G != 128 || got.B != 128 || got.A != 179 {
		t.Fatal(got)
	}
}

func TestNativeBundledNerdFileCapture(t *testing.T) {
	data, err := orbitassets.Asset("fonts/SymbolsNerdFont-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	font, err := sfnt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buffer sfnt.Buffer
	index, err := font.GlyphIndex(&buffer, '\uea60')
	if err != nil || index == 0 {
		t.Fatal("bundled Nerd probe glyph missing", err)
	}
	if err = ui.RegisterFont(data, ""); err != nil {
		t.Fatal(err)
	}
	f, tt := newNative(t)
	tt.SetDark(true)
	root := t.TempDir()
	for _, path := range []string{"src/main.go", "src/App.tsx", "README.md"} {
		file(t, root, path)
	}
	f.controller.SetContext(root, "", "s")
	tt.Type("@")
	waitScan(t, f.controller)
	tt.Frame()
	if f.controller.nerdFont == "" {
		t.Fatal("registered Nerd Font was not selected")
	}
	if !f.controller.IsOpen() {
		t.Fatal("file popup not open")
	}
	if dir := os.Getenv("ORBIT_COMPOSER_CAPTURE_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		out, err := os.Create(filepath.Join(dir, "composer-files-native-cpu.png"))
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()
		if err := png.Encode(out, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNativeSourceWindowInsetBeforePlacement(t *testing.T) {
	f, tt := newNative(t)
	tt.Type("/re")
	tt.SetSize(700, 300)
	tt.Frame()
	tt.Advance(time.Second)
	popup, ok := tt.Find("Composer completions")
	if !ok {
		t.Fatal("popup vanished on resize")
	}
	if popup.X != 8 || popup.X+popup.W != 692 || popup.Y < 8 || popup.Y+popup.H > 292 {
		t.Fatal("source 8px clamp", popup)
	}
	want := popupBounds(ui.Rect{X: 500, Y: 20, W: 200, H: 100}, 700, 300, 3, f.style)
	if want.X != 8 || want.Y != 8 || want.W != 684 || want.H != 103 {
		t.Fatal(want)
	}
}
