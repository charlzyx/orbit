package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workbench"
	"image"
	"image/color"
	"math"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestUsageSourceDefaultSortQuirkAndNoHeaderReorder(t *testing.T) {
	a, tt := usageTableFixture(t, "sessions")
	tab := a.usageState.table("sessions")
	tab.Page = 2
	tt.Frame()
	for i := 0; i < 3; i++ {
		if err := tt.Click("Tokens"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if tab.Sort != usageDefaultSort("sessions") || tab.Page != 2 {
			t.Fatalf("active source default must remain identical, got %+v page %d", tab.Sort, tab.Page)
		}
	}
	if err := tt.Click("Session"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tab.Sort.Column != "title" || tab.Sort.Descending || tab.Page != 1 {
		t.Fatal(tab.Sort, tab.Page)
	}
	h, _ := tt.Find("Session")
	tt.Press(float32(h.X+25), float32(h.Y+20))
	tt.Move(float32(h.X+h.W+90), float32(h.Y+20))
	tt.Release(float32(h.X+h.W+90), float32(h.Y+20))
	if len(tab.List.Columns.Order) != 0 {
		t.Fatal("source does not support header reorder", tab.List.Columns.Order)
	}
	grip, ok := tt.Find("Resize sessions title")
	if !ok || grip.W != 5 {
		t.Fatal("source resize grip", grip, ok)
	}
	before := tab.Sort
	tt.Press(float32(grip.X+1), float32(grip.Y+20))
	tt.Move(float32(headerLeftForResize(tt)), float32(grip.Y+20))
	tt.Release(float32(headerLeftForResize(tt)), float32(grip.Y+20))
	if tab.List.Columns.Widths["title"] != 48 || tab.Sort != before {
		t.Fatal("resize clamp or sort isolation", tab.List.Columns.Widths, tab.Sort, before)
	}
	if got := usageNextSort("series", usageDefaultSort("series"), "time"); !got.Descending {
		t.Fatal("ascending default should advance to descending", got)
	}
	if got := usageNextSort("series", ui.SortOrder{Column: "time", Descending: true}, "time"); got != usageDefaultSort("series") {
		t.Fatal(got)
	}
}
func colorsInside(img image.Image, r image.Rectangle) int {
	seen := map[color.NRGBA]bool{}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			seen[color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)] = true
		}
	}
	return len(seen)
}
func TestUsageSourceNativeCellGeometryAndHoverOpen(t *testing.T) {
	a := fixtureUsageApp(t)
	tab := a.usageState.table("sessions")
	a.usageState.projection.Sessions = a.usageState.projection.Sessions[:1]
	row := a.usageState.projection.Sessions[0]
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Padding(10).Children(func() { a.usageTable(c, "sessions") })
	}, 1500, 500)
	tt.Frame()
	header, ok := tt.Find("Session")
	if !ok {
		t.Fatal("header missing")
	}
	text, ok := tt.Find(row.Name)
	if !ok {
		t.Fatal("session title missing")
	}
	if math.Abs(float64(text.X-header.X-16)) > .01 {
		t.Fatal("source cell padding not 16", text, header)
	}
	button, ok := tt.Find("Open " + row.Name)
	if !ok || button.W != 22 || button.H != 22 {
		t.Fatal("source Open button geometry", button, ok)
	}
	inner := image.Rect(int(button.X+4), int(button.Y+4), int(button.X+button.W-4), int(button.Y+button.H-4))
	tt.Move(0, 0)
	tt.Frame()
	hidden := colorsInside(tt.Image(), inner)
	tt.Move(float32(text.X+4), float32(text.Y+text.H/2))
	tt.Frame()
	shown := colorsInside(tt.Image(), inner)
	if hidden > 1 || shown <= hidden {
		t.Fatal("Open arrow should paint only on row hover", hidden, shown)
	}
	if err := tt.Click("Open " + row.Name); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "SwitchSession" || actions[0].Value != row.Path || a.usageState.filter.SessionPath != "" || tab.Selected != -1 {
		t.Fatal("Open action accidentally scoped/selected row", actions, a.usageState.filter, tab.Selected)
	}
	if usageClipCell("汉字abcde", 48, 13) != "汉字…" {
		t.Fatal("source codepoint clipping changed", usageClipCell("汉字abcde", 48, 13))
	}
}

func headerLeftForResize(tt *ui.Tester) float32 { r, _ := tt.Find("Session"); return r.X }

func TestUsageSourceSharedBreakdownWidthsAndIndependentSearchPage(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	s.projection.Models = nil
	s.projection.Providers = nil
	for i := 0; i < 35; i++ {
		r := usageAggregate{Key: fmt.Sprint(i), Name: fmt.Sprintf("Match %02d", i), Requests: 1, Tokens: workbench.Tokens{Total: uint64(i + 1)}}
		s.projection.Models = append(s.projection.Models, r)
		s.projection.Providers = append(s.projection.Providers, r)
	}
	kind := "breakdown:models"
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(10).Gap(8).Children(func() { a.usageTable(c, kind) }) }, 1000, 700)
	tt.Frame()
	models := s.table(kind)
	models.Page = 2
	tt.Frame()
	head, ok := tt.Find("Usage table search head " + kind)
	if !ok || head.H != 36 || head.W < 800 {
		t.Fatal("source search should grow beside Columns", head)
	}
	if err := tt.Click("Search " + kind); err != nil {
		t.Fatal(err)
	}
	tt.Type("Match")
	tt.Frame()
	if models.Page != 2 {
		t.Fatal("source breakdown search does not reset page", models.Page)
	}
	grip, ok := tt.Find("Resize " + kind + " name")
	if !ok {
		t.Fatal("no name grip")
	}
	tt.Press(grip.X+1, grip.Y+20)
	tt.Move(grip.X-40, grip.Y+20)
	tt.Release(grip.X-40, grip.Y+20)
	tt.Frame()
	width := models.List.Columns.Widths["name"]
	if width < 48 || width == 0 {
		t.Fatal("no resize override", width)
	}
	models.Hidden = []string{"input"}
	kind = "breakdown:providers"
	tt.Frame()
	providers := s.table(kind)
	if providers.Search != "" || providers.Page != 1 || len(providers.Hidden) != 0 || providers.List.Columns.Widths["name"] != width {
		t.Fatal("width scope versus view state", providers, width)
	}
	header, _ := tt.Find("Name")
	if header.W != width {
		t.Fatal("shared override not painted", header, width)
	}
	if len(s.table("sessions").List.Columns.Widths) != 0 || len(s.table("breakdown:tools").List.Columns.Widths) != 0 {
		t.Fatal("breakdown widths leaked into unrelated table")
	}
	kind = "breakdown:models"
	tt.Frame()
	if models.Search != "Match" || models.Page != 2 || !slices.Contains(models.Hidden, "input") {
		t.Fatal("dimension view not restored", models)
	}
}
