package transcriptui

import (
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func nativeOptions() Options {
	return Options{FontSize: 14, CodeSize: 12.5, CodeFont: "monospace", Colors: Colors{Text: ui.Hex("#dce4f0"), Muted: ui.Hex("#92a0b5"), Accent: ui.Hex("#60a5fa"), Border: ui.Hex("#344155"), Surface: ui.Hex("#192233"), Code: ui.Hex("#111827"), Error: ui.Hex("#f87171"), Warning: ui.Hex("#fbbf24"), Success: ui.Hex("#4ade80"), Selection: ui.Hex("#315c99")}}
}
func nativeView(build func(*ui.Context)) func(*ui.Context) {
	return func(c *ui.Context) {
		ui.Column(c).Size(850, 900).Padding(20).Gap(10).Background(ui.Hex("#111827")).Selectable().Children(func() { build(c) })
	}
}
func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if e := tt.Click(label); e != nil {
		t.Fatalf("click %q: %v; texts=%v", label, e, tt.Texts())
	}
}
func TestNativeToolFailureFoldCopyRepeatAndStreaming(t *testing.T) {
	var state State
	state.Session("fixture")
	call := newTool("fixture-bash", "bash", json.RawMessage(`{"command":"go test ./..."}`))
	call.Output, _ = json.Marshal("actual fixture failure\n" + strings.Repeat("fixture log\n", 20) + "END")
	call.Failed = true
	call.Complete = true
	call.Facts = ToolFacts{Truncated: true}
	opts := nativeOptions()
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Tool(c, &state, call, opts) }), 850, 900)
	if !tt.HasText("actual fixture failure") || tt.HasText("END") {
		t.Fatal("collapsed error strip/output visibility wrong")
	}
	click(t, tt, "Run go test ./...")
	if !tt.HasText("Command") || !tt.HasText("Output") || tt.HasText("END") {
		t.Fatal("typed preview wrong")
	}
	if _, ok := tt.Find("Show all 22 lines"); !ok {
		t.Fatalf("line folding absent: %v", tt.Texts())
	}
	click(t, tt, "Show all 22 lines")
	if !tt.HasText("END") {
		t.Fatal("expand lost tail")
	}
	// The first section's copy action carries shell text, not the header preview.
	click(t, tt, "Copy section")
	if got := tt.Clipboard(); got != "go test ./..." {
		t.Fatalf("copy command=%q", got)
	}
	tt.Advance(3 * time.Second)
	click(t, tt, "Show less")
	if tt.HasText("END") {
		t.Fatal("recollapse failed")
	}
	click(t, tt, "Run go test ./...")
	if tt.HasText("Command") {
		t.Fatal("card did not close")
	}
	click(t, tt, "Run go test ./...")
	if !tt.HasText("Command") {
		t.Fatal("repeat open failed")
	}
	call.Running = true
	call.Complete = false
	call.Failed = false
	call.Output, _ = json.Marshal("updated fixture output")
	tt.Frame()
	if !tt.HasText("updated fixture output") || tt.HasText("actual fixture failure") {
		t.Fatal("stale failure/output retained")
	}
	call.Running = false
	call.Complete = true
	tt.Frame()
	state.Session("new-fixture")
	tt.Frame()
	if tt.HasText("updated fixture output") {
		t.Fatal("session reset left card open")
	}
}
func TestNativeLogicalTurnFoldAbortAndMissingUsage(t *testing.T) {
	var state State
	state.Session("fixture")
	opts := nativeOptions()
	rows := Project([]agent.Message{
		fixtureMessage(t, "a", `{"role":"assistant","content":[{"type":"thinking","thinking":"Explicit fixture thought"},{"type":"toolCall","id":"r","name":"read","arguments":{"path":"src/main.go"}}]}`),
		fixtureMessage(t, "b", `{"role":"assistant","content":"Explicit fixture answer","stopReason":"aborted","errorMessage":"Request was aborted"}`),
	})
	row := rows[0]
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Turn(c, &state, row, opts) }), 850, 900)
	if !tt.HasText("Worked") || !tt.HasText("Stopped") || !tt.HasText("Explicit fixture answer") || tt.HasText("Agent error") || tt.HasText("↑0") {
		t.Fatalf("turn markers %v", tt.Texts())
	}
	click(t, tt, "Worked")
	click(t, tt, "Ran 1 file read · 1 thought")
	if !tt.HasText("Explicit fixture thought") || !tt.HasText("src/main.go") {
		t.Fatal("hidden activity didn't open")
	}
	click(t, tt, "Thought")
	if tt.HasText("Explicit fixture thought") {
		t.Fatal("thought didn't fold")
	}
	click(t, tt, "Copy Message")
	if tt.Clipboard() != "Explicit fixture answer" {
		t.Fatal("copied work or error instead of answer")
	}
}
func TestNativeCrossMessageMarkdownSelectionAndMenu(t *testing.T) {
	md := markdown.Options{ShareSelection: true}
	source := "Second **styled** message.\n\nThird block."
	first := "First 用户 message."
	opts := nativeOptions()
	message := "Second **styled** message.\n\nThird block."
	var selected []ui.SelectedText
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Size(700, 350).Padding(20).Gap(12).Selectable()
		scope.Children(func() { ui.Text(c, first); markdown.View(c, source, &md); ui.Text(c, "Not selectable").Unselectable() })
		SelectionMenu(c, scope, message, opts)
		selected = scope.SelectedTexts()
	}, 700, 350)
	a, ok := tt.Find(first)
	if !ok {
		t.Fatal("first absent")
	}
	b, ok := tt.Find("Third block.")
	if !ok {
		t.Fatalf("third absent: %v", tt.Texts())
	}
	tt.Press(a.X, a.Y+a.H/2)
	tt.Move(b.X+min(float32(200), b.W-1), b.Y+b.H/2)
	tt.Release(b.X+min(float32(200), b.W-1), b.Y+b.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	if got, want := tt.Clipboard(), first+"\nSecond styled message.\nThird block."; got != want {
		t.Fatalf("cross-message copy=%q want %q selected %#v", got, want, selected)
	}
	source += " Appended streamed text."
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyC)
	if !strings.Contains(tt.Clipboard(), first) {
		t.Fatal("append destroyed existing selection")
	}
	if e := tt.RightClick(first); e != nil {
		t.Fatal(e)
	}
	if got := strings.Join(tt.Menu(), "|"); !strings.Contains(got, "Copy Selection") || !strings.Contains(got, "Copy Message") {
		t.Fatalf("source menu missing: %v", tt.Menu())
	}
	if e := tt.ChooseMenuItem("Copy Message"); e != nil {
		t.Fatal(e)
	}
	if tt.Clipboard() != message {
		t.Fatalf("copy message=%q", tt.Clipboard())
	}
	// Reopening and Escape must close without executing another clipboard write.
	if e := tt.RightClick(first); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyEscape)
	if tt.Clipboard() != message {
		t.Fatal("dismissal changed clipboard")
	}
}
func TestNativeSummaryAndChangedFileLimits(t *testing.T) {
	var state State
	state.Session("fixture")
	opts := nativeOptions()
	reviews := 0
	opts.Review = func() { reviews++ }
	tokens := uint64(90000)
	row := Row{ID: "summary", Summary: &SummaryRow{Text: "Fixture boundary summary", TokensBefore: &tokens}}
	var files []ChangedFile
	for i := 0; i < 14; i++ {
		files = append(files, ChangedFile{fmt.Sprintf("file-%02d.go", i), 2, 1})
	}
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Summary(c, &state, row, opts); Files(c, &state, files, opts) }), 850, 900)
	if tt.HasText("Fixture boundary summary") || tt.HasText("file-03.go") {
		t.Fatal("default folds differ")
	}
	click(t, tt, "Compaction · Compacted from 90K tokens")
	if !tt.HasText("Fixture boundary summary") {
		t.Fatal("summary failed to expand")
	}
	click(t, tt, "Show 11 more files")
	if !tt.HasText("file-11.go") || tt.HasText("file-12.go") || !tt.HasText("Showing first 12 of 14") {
		t.Fatal("source 3/12 file bound differs")
	}
	click(t, tt, "Review")
	if reviews != 1 {
		t.Fatal("review callback not exact once")
	}
}
func TestNativeTranscriptCapture(t *testing.T) {
	out := os.Getenv("ORBIT_TRANSCRIPT_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_TRANSCRIPT_CAPTURE_DIR for explicit fixture native CPU review images")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		t.Fatal(e)
	}
	var state State
	state.Session("capture-fixture")
	opts := nativeOptions()
	opts.Markdown = func(c *ui.Context, s string, stream bool) {
		markdown.View(c, s, &markdown.Options{ShareSelection: true})
	}
	call := newTool("capture-bash", "bash", json.RawMessage(`{"command":"go test ./internal/transcriptui -race"}`))
	call.Complete = true
	call.Output, _ = json.Marshal("Explicit fixture output, not a live model run.\nok   internal/transcriptui\nAll fixture cases passed.")
	edit := newTool("capture-edit", "edit", json.RawMessage(`{"path":"internal/example.go","edits":[{"oldText":"func old() {\n    return false\n}\n","newText":"func updated() {\n    return true\n}\n"}]}`))
	edit.Complete = true
	state.Toggle("tool:capture-bash", false)
	state.Toggle("tool:capture-edit", false)
	tt := ui.NewTester(nativeView(func(c *ui.Context) {
		ui.Text(c, "EXPLICIT NATIVE CPU FIXTURE · not a matched Rust capture").Unselectable().TextColor(opts.Colors.Warning)
		Tool(c, &state, call, opts)
		Tool(c, &state, edit, opts)
		Footer(c, &state, Row{ID: "footer", Role: "assistant", Steps: []Step{{Text: "Fixture answer"}}}, opts)
	}), 850, 900)
	for _, name := range []string{"typed-tools-expanded"} {
		f, e := os.Create(filepath.Join(out, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, tt.Image())
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
	}
	call.Failed = true
	call.Output, _ = json.Marshal("Fixture error: command exited with status 1")
	state.Toggle("tool:capture-bash", false)
	tt.Frame()
	f, e := os.Create(filepath.Join(out, "typed-tools-failure.png"))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if e = png.Encode(f, tt.Image()); e != nil {
		t.Fatal(e)
	}
}

func TestNativeUsageHoverTailOwnsSingleFooter(t *testing.T) {
	var state State
	state.Session("fixture")
	opts := nativeOptions()
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.Local)
	opts.Now = func() time.Time { return now }
	row := Project([]agent.Message{fixtureMessage(t, "answer", `{"role":"assistant","content":"Final fixture answer","timestamp":"2026-10-25T12:00:00Z","usage":{"input":100,"output":30,"cacheRead":900,"cacheWrite":0,"cost":{"total":0.0012}},"durationMs":2000}`)})[0]
	files := []ChangedFile{{"fixture.go", 2, 1}}
	tt := ui.NewTester(nativeView(func(c *ui.Context) {
		suppressed := opts
		suppressed.SuppressFooter = true
		Turn(c, &state, row, suppressed)
		Tail(c, &state, files, row, opts)
	}), 850, 900)
	if !tt.HasText("Changed 1 file") || !tt.HasText("Today") {
		t.Fatalf("tail metadata missing %v", tt.Texts())
	}
	copies := 0
	for _, text := range tt.Texts() {
		if text == "Copy Message" {
			copies++
		}
	}
	if copies != 1 {
		t.Fatalf("duplicate/missing final footer count=%d", copies)
	}
	label := row.UsageLabel()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("usage label %q absent: %v", label, tt.Texts())
	}
	tt.Move(r.X+min(float32(30), r.W/2), r.Y+r.H/2)
	for _, text := range []string{"Message usage", "Cache read", "Cache write", "Total", "Rate", "$0.0012", "15 tok/s avg"} {
		if !tt.HasText(text) {
			t.Fatalf("hover lacks %q: %v", text, tt.Texts())
		}
	}
	tt.Move(5, 850)
	if tt.HasText("Message usage") {
		t.Fatal("hover card remains after leave")
	}
}
func TestNativeQuestionRecordWaitingAnsweredDeclined(t *testing.T) {
	var state State
	state.Session("fixture")
	opts := nativeOptions()
	tool := newTool("question", "ask_user_question", json.RawMessage(`{"questions":[{"header":"Method","question":"Which method?","options":[{"label":"A","description":"Alpha"},{"label":"B","description":"Beta"}]},{"question":"What else?","options":[]}]}`))
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Tool(c, &state, tool, opts) }), 850, 900)
	if !tt.HasText("Question") || !tt.HasText("Which method?") || !tt.HasText("Waiting for an answer") || tt.HasText("Arguments") {
		t.Fatal("question not specialized")
	}
	tool.Output, _ = json.Marshal(`User has answered your questions: "Which method?"="A, B". "What else?"="Custom fixture answer".`)
	tt.Frame()
	if tt.HasText("Waiting for an answer") || !tt.HasText("Custom fixture answer") {
		t.Fatal("answered record stale")
	}
	tool.Output, _ = json.Marshal("User declined to answer questions")
	tt.Frame()
	if !tt.HasText("No answer") || tt.HasText("Custom fixture answer") {
		t.Fatal("decline record stale")
	}
}
func TestNativeVirtualizedCrossRowSelectionAndEndpointLimit(t *testing.T) {
	var list ui.ListState
	var parts []ui.SelectedText
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Size(500, 180).Selectable().CopyData(SelectionData)
		scope.Children(func() {
			ui.List(c, &list, 80, func(i int) { ui.Text(c, fmt.Sprintf("Row %02d 用户文字", i)).Height(40) }).Grow(1)
		})
		parts = scope.SelectedTexts()
	}, 500, 180)
	a, ok := tt.Find("Row 00 用户文字")
	if !ok {
		t.Fatal("first virtual row missing")
	}
	b, ok := tt.Find("Row 02 用户文字")
	if !ok {
		t.Fatal("third virtual row missing")
	}
	tt.Press(a.X+1, a.Y+10)
	tt.Move(b.X+min(float32(250), b.W-1), b.Y+10)
	tt.Release(b.X+min(float32(250), b.W-1), b.Y+10)
	tt.Key(ui.Cmd, ui.KeyC)
	if got := tt.Clipboard(); got != "Row 00 用户文字\nRow 01 用户文字\nRow 02 用户文字" {
		t.Fatalf("virtual row selection %q", got)
	}
	// Concrete pinned-MyGo limitation, NOT a Rust parity pass: when virtualization
	// unmounts a selected endpoint, MyGo clears its selection (textselection.go
	// syncTextSelection). Rust snapshots SelectedBlocks and preserves copied text.
	list.ScrollToEnd()
	tt.Frame()
	tt.SetClipboard("unchanged sentinel")
	tt.Key(ui.Cmd, ui.KeyC)
	if len(parts) != 0 || tt.Clipboard() != "unchanged sentinel" {
		t.Fatal("selection behavior changed; revisit documented endpoint limit")
	}
}

func TestNativeToolKeyboardToggleAndCopy(t *testing.T) {
	var state State
	state.Session("keyboard-fixture")
	call := newTool("keyboard", "bash", json.RawMessage(`{"command":"printf keyboard"}`))
	opts := nativeOptions()
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Tool(c, &state, call, opts) }), 850, 900)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Run printf keyboard") {
		t.Fatalf("header not keyboard focusable: %v", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	if !tt.HasText("Command") {
		t.Fatal("Enter did not open card")
	}
	tt.Key(0, ui.KeySpace)
	if tt.HasText("Command") {
		t.Fatal("Space did not close card")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if tt.Clipboard() != "printf keyboard" {
		t.Fatalf("keyboard copy %q", tt.Clipboard())
	}
}
