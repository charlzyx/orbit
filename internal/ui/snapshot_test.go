package orbitui

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
)

func TestMain(m *testing.M) {
	dir, e := os.MkdirTemp("", "orbit-native-ui-cache-")
	if e != nil {
		panic(e)
	}
	os.Setenv("XDG_CACHE_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func fixture(t *testing.T) *App {
	t.Helper()
	if e := LoadAssets(); e != nil {
		t.Fatal(e)
	}
	s := preferences.New(t.TempDir())
	p := preferences.DefaultUI()
	p.Language = "en"
	p.ReduceMotion = true
	if e := s.SaveUI(p); e != nil {
		t.Fatal(e)
	}
	if e := s.SaveAppearance(preferences.Appearance{Mode: "dark", DarkTheme: "orbit-azure", LightTheme: "orbit-azure-light"}); e != nil {
		t.Fatal(e)
	}
	return New(Options{Store: s})
}
func TestNativeCapture(t *testing.T) {
	out := os.Getenv("ORBIT_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_CAPTURE_DIR to write reproducible native CPU screenshots")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		t.Fatal(e)
	}
	a := fixture(t)
	a.Options.Store.PiDir = "/home/orbit/.pi/agent"
	tester := ui.NewTester(a.View, 1470, 922)
	tester.SetDark(true)
	tester.Advance(time.Second)
	save := func(name string) {
		t.Helper()
		tester.Frame()
		f, e := os.Create(filepath.Join(out, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		if e = png.Encode(f, tester.Image()); e != nil {
			t.Fatal(e)
		}
	}
	save("new-task-dark")
	a.Navigate("settings")
	save("settings-general-dark")
	a.settingsPage = "appearance"
	save("settings-appearance-dark")
	a.Navigate("chat")
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "fixture", SessionName: "Native rendering fixture", Workspace: "/workspace/example", ModelID: "fixture-model", Thinking: "medium", Messages: []Message{{ID: "u1", Role: "user", Text: "Explain the native migration."}, {ID: "a1", Role: "assistant", Text: "## A native workbench\n\nThis is an explicit visual test fixture. No live model output or metrics are simulated.\n\n- Go owns the application state.\n- MyGo paints native controls.\n- Pi communicates over JSONL RPC.\n\n```go\nfunc main() {\n    runNativeWorkbench()\n}\n```\n\n| Surface | Backend |\n| --- | --- |\n| Files | Local filesystem |\n| Sessions | Pi RPC |"}}})
	save("transcript-fixture-dark")
}
func TestTokensAndCommandRegistry(t *testing.T) {
	a := fixture(t)
	if len(Palettes) != 10 {
		t.Fatalf("palettes=%d", len(Palettes))
	}
	if len(Commands) != 58 {
		t.Fatalf("commands=%d", len(Commands))
	}
	if a.Theme.Space(16) != 16 {
		t.Fatal("default spacing")
	}
	seen := map[string]bool{}
	for _, c := range Commands {
		if seen[c.ID] {
			t.Fatal("duplicate command", c.ID)
		}
		seen[c.ID] = true
	}
	for _, b := range Bindings {
		if _, _, ok := keyChord(b.Chord); !ok {
			t.Errorf("unsupported chord %q", b.Chord)
		}
	}
}
func TestNativeShortcutsAndComposer(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tester := ui.NewTester(a.View, 1470, 922)
	tester.Key(ui.Cmd, ui.KeyN)
	if len(actions) != 1 || actions[0].Kind != "NewSession" {
		t.Fatalf("new session action: %#v", actions)
	}
	tester.Key(ui.Cmd, ui.KeyComma)
	if a.page != "settings" {
		t.Fatal("settings shortcut")
	}
	a.Navigate("chat")
	tester.Frame()
	if e := tester.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tester.Type("test prompt")
	tester.Key(0, ui.KeyEnter)
	if len(actions) < 2 || actions[len(actions)-1].Value != "test prompt" {
		t.Fatalf("submit: %#v", actions)
	}
	if a.Draft != "" {
		t.Fatal("draft not cleared")
	}
}
func TestSettingsNavigationComplete(t *testing.T) {
	if len(settingsCategories) != 14 {
		t.Fatalf("settings sections=%d, want14", len(settingsCategories))
	}
}

func TestIMECompositionDoesNotSubmit(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tester := ui.NewTester(a.View, 1000, 700)
	if e := tester.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tester.Compose("你好", 2)
	tester.Key(0, ui.KeyEnter)
	if len(actions) != 0 {
		t.Fatal("IME composition submitted a prompt")
	}
	tester.Compose("", 0)
	tester.Type("你好")
	tester.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "Submit" {
		t.Fatalf("committed input not submitted: %#v", actions)
	}
}
func TestEditorPersistsNativeInput(t *testing.T) {
	a := fixture(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	if e := os.WriteFile(path, []byte("package main\n"), 0600); e != nil {
		t.Fatal(e)
	}
	a.Model.Workspace = dir
	a.initFiles()
	d, e := a.files.editor.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { a.files.editor.CloseAll() })
	a.syncDocument(d)
	a.files.loaded = dir
	a.Navigate("files")
	tester := ui.NewTester(a.View, 1200, 800)
	if e = tester.Click("File editor"); e != nil {
		t.Fatal(e)
	}
	tester.Key(ui.Cmd, ui.KeyEnd)
	tester.Type("// native edit\n")
	tester.Frame()
	d, _ = a.files.editor.Active()
	if !strings.Contains(d.Text, "native edit") {
		t.Fatalf("native buffer did not reach autosave model: %q", d.Text)
	}
	if e = a.files.editor.SaveAll(); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(path)
	if e != nil || !strings.Contains(string(b), "native edit") {
		t.Fatalf("saved %q, %v", b, e)
	}
}
func TestMarkdownAlertAndGeometry(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "fixture", Messages: []Message{{ID: "a", Role: "assistant", Text: "> [!WARNING]\n> Keep your keys private.\n"}}})
	tester := ui.NewTester(a.View, 1470, 922)
	if !tester.HasText("Warning") {
		t.Fatal("GFM alert missing")
	}
	r, ok := tester.Find("Message")
	if !ok || r.Y < 700 || r.X+r.W > 1470 {
		t.Fatalf("composer not anchored to bottom: %#v", r)
	}
	tester.SetScale(2)
	if tester.Image().Bounds().Dx() != 2940 {
		t.Fatal("2x scaling not applied")
	}
}

func TestPickerKeyboardCommit(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.Model.Models = []ModelChoice{{ID: "a", Provider: "p", Name: "Alpha"}, {ID: "b", Provider: "p", Name: "Beta"}}
	tester := ui.NewTester(a.View, 1200, 800)
	a.modelOpen = true
	tester.Frame()
	tester.Key(0, ui.KeyDown)
	tester.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Value != "b" {
		t.Fatalf("keyboard model picker: %#v", actions)
	}
	a.paletteOpen = true
	a.paletteQuery = "new session"
	tester.Frame()
	tester.Key(0, ui.KeyEnter)
	if len(actions) != 2 || actions[1].Kind != "NewSession" {
		t.Fatalf("keyboard palette: %#v", actions)
	}
}

func TestDraftSurvivesSessionSwitch(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "one"})
	a.Draft = "unsent one"
	a.SetSnapshot(Snapshot{SessionID: "two"})
	if a.Draft != "" {
		t.Fatal("new task inherited another task draft")
	}
	a.Draft = "unsent two"
	a.SetSnapshot(Snapshot{SessionID: "one"})
	if a.Draft != "unsent one" {
		t.Fatalf("draft lost: %q", a.Draft)
	}
}
