package orbitui

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
	"strings"
	"testing"
)

const raggedTableAudit = "| A | B |\n| --- | --- |\n| one | two | [lost](https://example.invalid/extra) |\n"

func TestMountedMarkdownRaggedTablePreservesRustBodyCells(t *testing.T) {
	d := parser.Parse(raggedTableAudit)
	cells := 0
	for i := 1; i < d.Len(); i++ {
		if d.Block(i).Kind == parser.Table {
			t.Logf("table range %d:%d (ends before dropped cell)", d.Block(i).Start, d.Block(i).End)
		}
		if d.Block(i).Kind == parser.TableCell {
			cells++
		}
	}
	t.Logf("same input Go parser cells=%d (Rust source expects 2 header+3 body)", cells)
	for _, streaming := range []bool{false, true} {
		name := "settled"
		if streaming {
			name = "stream"
		}
		t.Run(name, func(t *testing.T) {
			a := fixture(t)
			var actions []Action
			a.Options.OnAction = func(x Action) { actions = append(actions, x) }
			a.SetSnapshot(Snapshot{SessionID: "table-audit", Messages: []Message{{ID: "table", Role: "assistant", Text: raggedTableAudit, Streaming: streaming}}})
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			if !tt.HasText("one") || !tt.HasText("two") {
				t.Fatal("baseline first two body cells absent")
			}
			if !tt.HasText("lost") {
				t.Fatal("third body cell and its link dropped from actual mounted transcript")
			}
			if err := tt.Click("lost"); err != nil {
				t.Fatal(err)
			}
			if len(actions) != 1 || actions[0].Kind != "OpenLink" || actions[0].Value != "https://example.invalid/extra" {
				t.Fatalf("wrong link action: %#v", actions)
			}
		})
	}
}

// Test-only candidate projection; not wired into production.
func raggedProjectionProof(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	rows := make([][]string, len(lines))
	cols := 0
	for i, line := range lines {
		s := strings.TrimSpace(line)
		s = strings.TrimPrefix(s, "|")
		s = strings.TrimSuffix(s, "|")
		rows[i] = strings.Split(s, "|")
		for j := range rows[i] {
			rows[i][j] = strings.TrimSpace(rows[i][j])
		}
		if len(rows[i]) > cols {
			cols = len(rows[i])
		}
	}
	var out strings.Builder
	for i, row := range rows {
		for len(row) < cols {
			if i == 1 {
				row = append(row, "---")
			} else {
				row = append(row, "")
			}
		}
		out.WriteString("| ")
		out.WriteString(strings.Join(row, " | "))
		out.WriteString(" |\n")
	}
	return out.String()
}

func TestRaggedProjectionPublicRendererProof(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	projected := raggedProjectionProof(raggedTableAudit)
	opts := markdown.Options{ShareSelection: true, Theme: a.md.Theme, OnLink: a.md.OnLink}
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Selectable().Children(func() { markdown.View(c, projected, &opts) }) }, 1200, 800)
	if !tt.HasText("lost") {
		t.Fatal("projection lost body cell")
	}
	if err := tt.Click("lost"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Value != "https://example.invalid/extra" {
		t.Fatal(actions)
	}
	first, _ := tt.Find("one")
	last, _ := tt.Find("lost")
	tt.Press(first.X, first.Y+first.H/2)
	tt.Move(last.X+last.W, last.Y+last.H/2)
	tt.Release(last.X+last.W, last.Y+last.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	if got := tt.Clipboard(); !strings.Contains(got, "one") || !strings.Contains(got, "two") || !strings.Contains(got, "lost") || strings.Contains(got, "---") {
		t.Fatalf("projection copy=%q", got)
	}
}

func TestRaggedProjectionCannotCreateNestedBlocksOrImageLoads(t *testing.T) {
	raw := "| A | B |\n| --- | --- |\n| # heading | ```sh | ![alt](file:/not-read.png) |\n| > quote | - list | <img src=\"file:/not-read.png\"> |\n"
	projected := raggedProjectionProof(raw)
	d := parser.Parse(projected)
	for i := 1; i < d.Len(); i++ {
		switch d.Block(i).Kind {
		case parser.Table, parser.TableRow, parser.TableCell:
		default:
			t.Fatalf("nested block %v", d.Block(i).Kind)
		}
	}
	opts := markdown.Options{ShareSelection: true}
	tt := ui.NewTester(func(c *ui.Context) { markdown.View(c, projected, &opts) }, 1200, 800)
	if !tt.HasText("[alt]") || !tt.HasText("# heading") {
		t.Fatal(tt.Texts())
	}
}

func TestMountedRaggedTableSelectionStreamAndOwner(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	snapshot := Snapshot{RuntimeID: "runtime-a", SessionID: "same", Workspace: "/work/a", Messages: []Message{{ID: "before", Role: "user", Text: "BEFORE TABLE"}, {ID: "table", Role: "assistant", Text: raggedTableAudit, Streaming: true}, {ID: "after", Role: "user", Text: "AFTER TABLE"}}}
	a.SetSnapshot(snapshot)
	tt := ui.NewTester(a.View, 1470, 922)
	first, ok := tt.Find("BEFORE TABLE")
	if !ok {
		t.Fatal("before missing")
	}
	last, ok := tt.Find("AFTER TABLE")
	if !ok {
		t.Fatal("after missing")
	}
	tt.Press(first.X, first.Y+first.H/2)
	tt.Move(last.X+last.W, last.Y+last.H/2)
	tt.Release(last.X+last.W, last.Y+last.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	copied := tt.Clipboard()
	for _, want := range []string{"BEFORE TABLE", "one", "two", "lost", "AFTER TABLE"} {
		if !strings.Contains(copied, want) {
			t.Fatalf("cross-message selection missing %s: %q", want, copied)
		}
	}
	if strings.Contains(copied, "---") || strings.Contains(copied, "| ") {
		t.Fatalf("synthetic syntax copied: %q", copied)
	}
	// Clear selection before testing the real link lifecycle.
	tt.Click("AFTER TABLE")
	snapshot.Messages[1].Streaming = false
	snapshot.Messages[1].Text += "\nFOLLOWING BLOCK\n"
	a.SetSnapshot(snapshot)
	tt.Frame()
	if !tt.HasText("FOLLOWING BLOCK") || !tt.HasText("lost") {
		t.Fatal("settle swallowed next block")
	}
	old, ok := tt.Find("lost")
	if !ok {
		t.Fatal("link missing")
	}
	tt.Press(old.X+old.W/2, old.Y+old.H/2)
	snapshot.RuntimeID = "runtime-b"
	snapshot.Workspace = "/work/b"
	a.SetSnapshot(snapshot)
	tt.Frame()
	tt.Release(old.X+old.W/2, old.Y+old.H/2)
	tt.Frame()
	if len(actions) != 0 {
		t.Fatalf("old held click crossed owner: %#v", actions)
	}
	if err := tt.Click("lost"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatal("fresh link not usable", actions)
	}
	if a.Model.Messages[1].Text != snapshot.Messages[1].Text {
		t.Fatal("projection mutated message source")
	}
}

func TestRaggedProjectionNormalTablesAndEscaping(t *testing.T) {
	for _, raw := range []string{"| A | B |\n| --- | --- |\n| one | two |", "| A | B |\n| --- | --- |\n| one |"} {
		if _, changed := raggedTableProjection(raw, 2); changed {
			t.Fatal("standard table intercepted")
		}
	}
	// Rust's intentionally naive split preserves the backslash as cell text.
	got := rustTableCells(`| first\|second | third |`)
	if len(got) != 3 || got[0] != `first\` || got[1] != "second" {
		t.Fatal(got)
	}
}

func TestMountedRaggedTableSafetyAndSourceCopy(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.md.LoadImage = func(string) (ui.ImageSource, error) { t.Error("table used image loader"); return nil, nil }
	raw := "| A | B |\n| --- | --- |\n| # heading | ```sh | ![alt](file:/not-read.png) |\n| > quote | - list | [unsafe](javascript:never_execute) |\n"
	a.SetSnapshot(Snapshot{SessionID: "safety", Messages: []Message{{ID: "table", Role: "assistant", Text: raw}}})
	tt := ui.NewTester(a.View, 1470, 922)
	if !tt.HasText("[alt]") || !tt.HasText("# heading") {
		t.Fatal("table inline semantics changed", tt.Texts())
	}
	if err := tt.Click("unsafe"); err != nil {
		t.Fatal(err)
	}
	// No direct ui.Link URL launch: the existing controller dispatch still owns
	// scheme rejection (run.go OpenLink), exactly like ordinary markdown links.
	if len(actions) != 1 || actions[0].Kind != "OpenLink" || actions[0].Value != "javascript:never_execute" {
		t.Fatal(actions)
	}
	if a.Model.Messages[0].Text != raw {
		t.Fatal("message mutated")
	}
}

func TestMountedRaggedTableHorizontalScroll(t *testing.T) {
	a := fixture(t)
	raw := "| A | B |\n| --- | --- |\n| " + strings.Repeat("wide first ", 12) + " | " + strings.Repeat("wide second ", 12) + " | " + strings.Repeat("wide third ", 12) + " | LASTCOLUMN |\n"
	// Mount the production hook in a narrow selectable surface so horizontal
	// movement is measured without guessing a native window coordinate.
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(300).Selectable().Children(func() { markdown.View(c, raw, &a.md) })
	}, 400, 500)
	before, ok := tt.Find("LASTCOLUMN")
	if !ok {
		t.Fatal("last cell missing")
	}
	first, ok := tt.Find(strings.Repeat("wide first ", 12))
	if !ok { // parser trims cell edge
		first, ok = tt.Find(strings.TrimSpace(strings.Repeat("wide first ", 12)))
	}
	if !ok {
		t.Fatal("first cell missing")
	}
	tt.Scroll(first.X+10, first.Y+10, 500, 0)
	tt.Frame()
	after, ok := tt.Find("LASTCOLUMN")
	if !ok {
		t.Fatal("last cell disappeared")
	}
	if after.X >= before.X {
		t.Fatalf("horizontal scroll did not move last cell: before=%v after=%v", before, after)
	}
}

func TestMountedRaggedTableChangedTargetRejectsHeldPress(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	snapshot := Snapshot{RuntimeID: "runtime-a", SessionID: "same", Workspace: "/work/a", Messages: []Message{{ID: "table", Role: "assistant", Text: raggedTableAudit}}}
	a.SetSnapshot(snapshot)
	tt := ui.NewTester(a.View, 1470, 922)
	link, ok := tt.Find("lost")
	if !ok {
		t.Fatal("link missing")
	}
	tt.Press(link.X+link.W/2, link.Y+link.H/2)
	snapshot.Messages[0].Text = strings.ReplaceAll(raggedTableAudit, "example.invalid/extra", "example.invalid/replacement")
	a.SetSnapshot(snapshot)
	tt.Frame()
	tt.Release(link.X+link.W/2, link.Y+link.H/2)
	tt.Frame()
	if len(actions) != 0 {
		t.Fatalf("held press retargeted same-label changed content: %#v", actions)
	}
	if err := tt.Click("lost"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Value != "https://example.invalid/replacement" {
		t.Fatal(actions)
	}
}

func TestRaggedTableBudgetsOverflowAndRawFallback(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, v := range [][3]int{{maxInt, maxInt, 0}, {2, maxInt, 0}, {maxInt, 2, 0}, {1, 1, maxInt}, {1, 1, -1}} {
		if raggedTableWithinBudget(v[0], v[1], v[2]) {
			t.Fatal("unsafe size admitted", v)
		}
	}
	if !raggedTableWithinBudget(128, 128, 0) || raggedTableWithinBudget(129, 128, 0) {
		t.Fatal("cell budget boundary")
	}
	if !raggedTableWithinBudget(1, 1, raggedTableMaxBytes-8) || raggedTableWithinBudget(1, 1, raggedTableMaxBytes-7) {
		t.Fatal("output budget boundary")
	}
	raw := "| A | B |\n| --- | --- |\n"
	wide := "| " + strings.Repeat("x | ", 128) + "TAIL_SENTINEL |\n"
	raw += wide + strings.Repeat("| one | two |\n", 128)
	projected, handled := raggedTableProjection(raw, 2)
	if !handled || projected != "" {
		t.Fatal("overbudget ragged table should raw-fallback")
	}
	a := fixture(t)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Selectable().Children(func() { markdown.View(c, raw, &a.md) }) }, 1400, 3000)
	if !tt.HasText("Table layout limit reached — showing complete source") {
		t.Fatal("limit unmarked")
	}
	body := strings.TrimSpace(raw)
	rect, ok := tt.Find(body)
	if !ok {
		t.Fatal("complete raw source missing")
	}
	tt.Press(rect.X, rect.Y)
	tt.Move(rect.X+rect.W, rect.Y+rect.H)
	tt.Release(rect.X+rect.W, rect.Y+rect.H)
	tt.Key(ui.Cmd, ui.KeyC)
	if got := tt.Clipboard(); got != body {
		t.Fatalf("raw fallback not exactly copyable: got len%d want%d", len(got), len(body))
	}
}

func TestRaggedTableHeaderWidthBudgetRealParser(t *testing.T) {
	// MyGo treats escaped pipes as one cell; Rust's raw split sees all of them.
	raw := "| " + strings.Repeat(`h\|`, raggedTableMaxCells) + "last | B |\n| --- | --- |\n| one | two | extra |\n"
	d := parser.Parse(raw)
	table := -1
	for i := 1; i < d.Len(); i++ {
		if d.Block(i).Kind == parser.Table {
			table = i
			break
		}
	}
	if table < 0 || len(d.Block(table).Align) != 2 {
		t.Fatal("fixture did not reach real two-column parser Table")
	}
	projected, handled := raggedTableProjection(raw, 2)
	if !handled || projected != "" {
		t.Fatalf("raw header slots escaped output budget: projected bytes=%d", len(projected))
	}
}

func TestRaggedTableMalformedHeaderDelimiterRealParser(t *testing.T) {
	cases := []struct {
		name, raw string
		table     bool
	}{
		{"more header", "| A | B | C |\n| --- | --- |\n| one | two | extra |\n", false},
		{"more delimiter", "| A | B |\n| --- | --- | --- |\n| one | two | extra |\n", false},
		{"escaped header normal body", "| A\\|Z | B |\n| --- | --- |\n| one | two |\n", true},
		{"escaped delimiter", "| A | B |\n| ---\\|--- | --- |\n| one | two | extra |\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := parser.Parse(tc.raw)
			found := false
			for i := 1; i < d.Len(); i++ {
				if d.Block(i).Kind == parser.Table {
					found = true
				}
			}
			if found != tc.table {
				t.Fatalf("real parser Table=%v", found)
			}
			if tc.table {
				if _, handled := raggedTableProjection(tc.raw, 2); handled {
					t.Fatal("normal escaped-header table intercepted")
				}
			}
		})
	}
}
