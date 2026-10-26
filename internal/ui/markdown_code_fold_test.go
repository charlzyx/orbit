package orbitui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown/parser"
)

func foldSource(prefix string, n int, closed bool) string {
	var b strings.Builder
	b.WriteString("```text\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "%s_LINE_%02d\n", prefix, i)
	}
	if closed {
		b.WriteString("```")
	}
	return b.String()
}
func foldTester(t *testing.T, source *string, live *bool) (*App, *ui.Tester) {
	t.Helper()
	a := fixture(t)
	a.SetSnapshot(Snapshot{RuntimeID: "fold-runtime", SessionID: "fold-session", Workspace: "fold-workspace"})
	a.md.Theme = a.Theme.Markdown()
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Selectable().Children(func() { a.chatMarkdown(c, *source, *live) }) }, 900, 1500)
	tt.Frame()
	return a, tt
}
func foldHas(tt *ui.Tester, text string) bool {
	return strings.Contains(strings.Join(tt.Texts(), "\n"), text)
}

func TestMountedCodeFoldRustThreshold(t *testing.T) {
	for _, n := range []int{24, 28, 29, 30} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			source := foldSource("BOUND", n, true)
			live := false
			_, tt := foldTester(t, &source, &live)
			label := fmt.Sprintf("Show remaining %d lines", n-24)
			_, present := tt.Find(label)
			if present != (n > 28) {
				t.Fatalf("threshold n=%d present=%v", n, present)
			}
			if n <= 28 {
				if !foldHas(tt, fmt.Sprintf("BOUND_LINE_%02d", n-1)) {
					t.Fatal("ordinary body omitted")
				}
				return
			}
			if !foldHas(tt, "BOUND_LINE_23") || foldHas(tt, "BOUND_LINE_24") {
				t.Fatal("preview not exactly24 lines", tt.Texts())
			}
			if err := tt.Click(label); err != nil {
				t.Fatal(err)
			}
			if !foldHas(tt, fmt.Sprintf("BOUND_LINE_%02d", n-1)) {
				t.Fatal("expand tail missing")
			}
			if err := tt.Click("Show less"); err != nil {
				t.Fatal(err)
			}
			if foldHas(tt, "BOUND_LINE_24") {
				t.Fatal("collapse retained tail")
			}
		})
	}
}
func TestCodeFoldPreviewLiteralLanguageAndCache(t *testing.T) {
	var lines []string
	for i := 0; i < 31; i++ {
		lines = append(lines, fmt.Sprintf("literal_%02d", i))
	}
	lines[0] = "```"
	lines[1] = "~~~~~~~~~~~~~~~~~~~~~~~~~~~~"
	lines[2] = "![not fetched](https://example.invalid/image.png)"
	lines[3] = "<script>never HTML</script>"
	lines[4] = "  \t 中文 😀  "
	source := "```````````````````````````````````````` text\n" + strings.Join(lines, "\n") + "\n````````````````````````````````````````"
	d := parser.Parse(source)
	s := codeFoldState{}
	s.prepare(d, 1)
	if len(s.blocks) != 1 {
		t.Fatal("fixture not one fenced block")
	}
	b := s.blocks[1]
	p := parser.Parse(b.preview)
	if p.Len() != 2 || p.Block(1).Kind != parser.CodeBlock || p.LineCount(1) != 24 || p.Block(1).Language() != "text" {
		t.Fatal("literal projection escaped", p.Source())
	}
	for i := 0; i < 24; i++ {
		got, _ := p.Line(1, i)
		want, _ := d.Line(1, i)
		if got != want {
			t.Fatalf("line%d=%q want%q", i, got, want)
		}
	}
	if alloc := testing.AllocsPerRun(50, func() { s.prepare(d, 1) }); alloc != 0 {
		t.Fatal("idle preparation allocated", alloc)
	}
	d2 := parser.Parse(source)
	s.prepare(d2, 1)
	if s.document != d2 {
		t.Fatal("same source different parser identity ignored")
	}
}
func TestCodeFoldActualTranscriptSelectionAndSettledOpen(t *testing.T) {
	source := foldSource("SELECT", 30, false)
	live := false
	_, tt := foldTester(t, &source, &live)
	if foldHas(tt, "SELECT_LINE_29") {
		t.Fatal("settled open fence incorrectly treated as live")
	}
	if err := tt.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSuffix(strings.TrimPrefix(source, "```text\n"), "\n")
	if tt.Clipboard() != want {
		t.Fatalf("folded Copy lost full payload %q", tt.Clipboard())
	}
	// Drag actual visible code bounds. The hidden tail is not mounted/selectable.
	preview := strings.Join(strings.Split(want, "\n")[:24], "\n")
	r, ok := tt.Find(preview)
	if !ok {
		t.Fatal("visible literal code bounds unavailable", tt.Texts())
	}
	tt.SetClipboard("sentinel")
	tt.Press(r.X, r.Y+2)
	tt.Move(r.X+r.W, r.Y+r.H-2)
	tt.Release(r.X+r.W, r.Y+r.H-2)
	tt.Key(ui.Cmd, ui.KeyC)
	if !strings.Contains(tt.Clipboard(), "SELECT_LINE_00") || strings.Contains(tt.Clipboard(), "SELECT_LINE_29") {
		t.Fatal("visible-only selection failed", tt.Clipboard())
	}
	live = true
	tt.Frame()
	if !foldHas(tt, "SELECT_LINE_29") {
		t.Fatal("semantic live body folded")
	}
	if _, ok := tt.Find("Show remaining 6 lines"); ok {
		t.Fatal("live toggle mounted")
	}
}
func TestCodeFoldExpandedStatePersistsRevisionButResetsOwner(t *testing.T) {
	source := foldSource("OLD", 30, true)
	live := false
	a, tt := foldTester(t, &source, &live)
	if err := tt.Click("Show remaining 6 lines"); err != nil {
		t.Fatal(err)
	}
	source = foldSource("NEW", 31, true)
	tt.Frame()
	if !foldHas(tt, "NEW_LINE_30") {
		t.Fatal("source edit lost expansion ordinal")
	}
	a.SetSnapshot(Snapshot{RuntimeID: "new-runtime", SessionID: "fold-session", Workspace: "fold-workspace"})
	tt.Frame()
	if foldHas(tt, "NEW_LINE_30") {
		t.Fatal("new owner inherited expansion")
	}
}
func TestCodeFoldOwnerContentAndReset(t *testing.T) {
	for _, kind := range []string{"content", "runtime", "ABA"} {
		t.Run(kind, func(t *testing.T) {
			source := foldSource("OLD", 30, true)
			live := false
			a, tt := foldTester(t, &source, &live)
			r, ok := tt.Find("Show remaining 6 lines")
			if !ok {
				t.Fatal("toggle missing")
			}
			tt.Press(r.X+r.W/2, r.Y+r.H/2)
			switch kind {
			case "content":
				source = foldSource("NEW", 30, true)
			case "runtime":
				a.SetSnapshot(Snapshot{RuntimeID: "other", SessionID: "same"})
			case "ABA":
				a.SetSnapshot(Snapshot{RuntimeID: "other", SessionID: "same"})
				a.SetSnapshot(Snapshot{RuntimeID: "fold-runtime", SessionID: "fold-session", Workspace: "fold-workspace"})
			}
			tt.Frame()
			tt.Release(r.X+r.W/2, r.Y+r.H/2)
			tt.Frame()
			if foldHas(tt, "OLD_LINE_29") || foldHas(tt, "NEW_LINE_29") {
				t.Fatal("held stale fold activated")
			}
		})
	}
}
func TestCodeFoldNestedAlertAndIndentedRemainUnfolded(t *testing.T) {
	for _, kind := range []string{"alert", "quote", "indented"} {
		t.Run(kind, func(t *testing.T) {
			source := foldSource("NEST", 30, true)
			if kind == "indented" {
				source = strings.Repeat("    INDENT_LINE\n", 30)
			} else {
				source = "> " + strings.ReplaceAll(source, "\n", "\n> ")
				if kind == "alert" {
					source = "> [!NOTE]\n" + source
				}
			}
			live := false
			_, tt := foldTester(t, &source, &live)
			if _, ok := tt.Find("Show remaining 6 lines"); ok {
				t.Fatal("nested/indented folded")
			}
			if kind != "indented" && !foldHas(tt, "NEST_LINE_29") {
				t.Fatal("nested tail omitted")
			}
		})
	}
}
func TestCodeFoldKeyboardActivation(t *testing.T) {
	source := foldSource("KEY", 30, true)
	live := false
	_, tt := foldTester(t, &source, &live)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if !foldHas(tt, "KEY_LINE_29") {
		t.Fatal("keyboard fold did not expand")
	}
}
func TestMountedCodeFoldCaptionsStayCorrect(t *testing.T) {
	source := strings.Replace(foldSource("CAPTION", 30, true), "```text", "```golang", 1)
	live := false
	_, tt := foldTester(t, &source, &live)
	if !tt.HasText("Go") {
		t.Fatal("source caption missing", tt.Texts())
	}
	if err := tt.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(tt.Clipboard(), "CAPTION_LINE_29") {
		t.Fatal("caption broke full copy")
	}
}

func TestCodeFoldMultipleBlocksMessagesAndOrdinal(t *testing.T) {
	a := fixture(t)
	a.md.Theme = a.Theme.Markdown()
	a.SetSnapshot(Snapshot{SessionID: "multiple-fold"})
	source := foldSource("A", 30, true) + "\n\nparagraph\n\n" + foldSource("B", 31, true)
	other := foldSource("C", 32, true)
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Children(func() {
			a.chatMarkdown(c.Key("first-message"), source, false)
			a.chatMarkdown(c.Key("second-message"), other, false)
		})
	}, 900, 3000)
	if err := tt.Click("Show remaining 6 lines"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Show remaining 7 lines"); err != nil {
		t.Fatal(err)
	}
	if !foldHas(tt, "A_LINE_29") || !foldHas(tt, "B_LINE_30") || foldHas(tt, "C_LINE_31") {
		t.Fatal("block/message expansion collided", tt.Texts())
	}
	source = "new paragraph\n\n" + foldSource("RESET", 30, true)
	tt.Frame()
	if foldHas(tt, "RESET_LINE_29") {
		t.Fatal("different ordinal retained removed expansion")
	}
}

func TestCodeFoldRealListRemeasuresAndPreservesScrolledAnchor(t *testing.T) {
	a := fixture(t)
	s := Snapshot{SessionID: "fold-list"}
	for i := 0; i < 8; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprintf("row%d", i), Role: "assistant", Text: foldSource(fmt.Sprintf("R%d", i), 29, true)})
	}
	s.Messages[3].Text = foldSource("R3", 30, true)
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	a.transcript.ScrollTo(3, ui.Start)
	tt.Frame()
	tt.Frame()
	first, _ := a.transcript.Visible()
	if first != 3 {
		t.Fatal("fixture did not anchor requested actual List row", first)
	}
	visibleBody := strings.TrimSuffix(strings.TrimPrefix(foldSource("R3", 24, false), "```text\n"), "\n")
	fullBody := strings.TrimSuffix(strings.TrimPrefix(foldSource("R3", 30, false), "```text\n"), "\n")
	before, ok := tt.Find(visibleBody)
	if !ok {
		t.Fatal("visible code card absent")
	}
	if err := tt.Click("Show remaining 6 lines"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	after, ok := tt.Find(fullBody)
	if !ok || after.H <= before.H {
		t.Fatal("actual list code height did not grow", before, after)
	}
	visible, _ := a.transcript.Visible()
	if visible != first || after.Y != before.Y {
		t.Fatal("fold changed list anchor", first, visible, before, after)
	}
	if err := tt.Click("Show less"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	collapsed, ok := tt.Find(visibleBody)
	if !ok {
		t.Fatal("collapsed actual body missing")
	}
	if collapsed.H >= after.H {
		t.Fatal("list failed to shrink", collapsed, after)
	}
	visible, _ = a.transcript.Visible()
	if visible != first {
		t.Fatal("collapse moved first visible row", visible)
	}
}

func TestCodeFoldRealListLiveTailFollowsAppendAndSettle(t *testing.T) {
	a := fixture(t)
	s := Snapshot{SessionID: "fold-tail", Running: true}
	for i := 0; i < 5; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprintf("u%d", i), Role: "user", Text: strings.Repeat("earlier text ", 50)})
	}
	s.Messages = append(s.Messages, Message{ID: "live", Role: "assistant", Text: foldSource("LIVE", 30, false), Streaming: true})
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Frame()
	if !a.transcript.AtEnd() {
		t.Fatal("initial live tail not followed")
	}
	s.Messages[len(s.Messages)-1].Text = foldSource("LIVE", 40, false)
	a.SetSnapshot(s)
	tt.Frame()
	tt.Frame()
	if !a.transcript.AtEnd() || !foldHas(tt, "LIVE_LINE_39") {
		t.Fatal("append lost live tail", a.transcript.AtEnd())
	}
	s.Running = false
	s.Messages[len(s.Messages)-1].Streaming = false
	a.SetSnapshot(s)
	tt.Frame()
	tt.Frame()
	if !a.transcript.AtEnd() || foldHas(tt, "LIVE_LINE_39") {
		t.Fatal("settling fold lost tail or kept hidden code", a.transcript.AtEnd())
	}
	if _, ok := tt.Find("Show remaining 16 lines"); !ok {
		t.Fatal("settled fold toggle missing")
	}
}

func TestCodeFoldToggleClearsOldSelection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "fold-selection"})
	a.md.Theme = a.Theme.Markdown()
	source := foldSource("DRAG", 30, true)
	selectedCount := 0
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Selectable()
		scope.Children(func() { a.chatMarkdown(c, source, false) })
		selectedCount = len(scope.SelectedTexts())
	}, 900, 1200)
	body := strings.TrimSuffix(strings.TrimPrefix(foldSource("DRAG", 24, false), "```text\n"), "\n")
	r, ok := tt.Find(body)
	if !ok || r.H <= 0 {
		t.Fatal("actual preview bounds absent")
	}
	tt.Press(r.X, r.Y+2)
	tt.Move(r.X+r.W, r.Y+r.H-2)
	tt.Release(r.X+r.W, r.Y+r.H-2)
	if selectedCount == 0 {
		t.Fatal("actual preview drag made no selection")
	}
	if err := tt.Click("Show remaining 6 lines"); err != nil {
		t.Fatal(err)
	}
	if selectedCount != 0 {
		t.Fatal("ordinary fold control retained old range")
	}
	if err := tt.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(tt.Clipboard(), "DRAG_LINE_29") {
		t.Fatal("expanded copy lost full body")
	}
}

func TestCodeFoldMetadataDisabledAndStreaming(t *testing.T) {
	d := parser.Parse(foldSource("LIVE", 30, true))
	for _, mode := range []struct{ live, allow bool }{{true, true}, {false, false}, {true, false}} {
		s := codeFoldState{}
		s.prepareEligible(d, 1, mode.live, mode.allow)
		if s.document != nil || s.blocks != nil || s.revision != 0 {
			t.Fatal("ineligible renderer materialized fold metadata")
		}
	}
	s := codeFoldState{}
	s.prepareEligible(d, 1, false, true)
	if len(s.blocks) != 1 {
		t.Fatal("eligible renderer did not prepare")
	}
}

func TestCodeFoldExpansionSurvivesThresholdRoundtrip(t *testing.T) {
	source := foldSource("EDIT", 29, true)
	live := false
	_, tt := foldTester(t, &source, &live)
	if err := tt.Click("Show remaining 5 lines"); err != nil {
		t.Fatal(err)
	}
	source = foldSource("EDIT", 28, true)
	tt.Frame()
	if !foldHas(tt, "EDIT_LINE_27") {
		t.Fatal("short body lost")
	}
	source = foldSource("EDIT", 29, true)
	tt.Frame()
	if !foldHas(tt, "EDIT_LINE_28") {
		t.Fatal("same block expansion lost across threshold")
	}
	if _, ok := tt.Find("Show less"); !ok {
		t.Fatal("expansion control state lost")
	}
}

func TestCodeFoldReplacementRevokesPreviewSelection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "replace"})
	a.md.Theme = a.Theme.Markdown()
	source := foldSource("OLD", 30, true)
	selected := []ui.SelectedText(nil)
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Selectable()
		scope.Children(func() { a.chatMarkdown(c, source, false) })
		selected = scope.SelectedTexts()
	}, 900, 1200)
	body := strings.TrimSuffix(strings.TrimPrefix(foldSource("OLD", 24, false), "```text\n"), "\n")
	r, ok := tt.Find(body)
	if !ok || r.H <= 0 {
		t.Fatal("preview unavailable")
	}
	tt.Press(r.X, r.Y+2)
	tt.Move(r.X+r.W, r.Y+r.H-2)
	tt.Release(r.X+r.W, r.Y+r.H-2)
	if len(selected) == 0 {
		t.Fatal("actual selection absent")
	}
	source = foldSource("NEW", 30, true)
	tt.Frame()
	if len(selected) != 0 {
		t.Fatalf("old selection remapped into replacement: %+v", selected)
	}
	if !foldHas(tt, "NEW_LINE_23") || foldHas(tt, "NEW_LINE_24") {
		t.Fatal("replacement prefix wrong")
	}
}
