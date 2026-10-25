package orbitui

import (
	"fmt"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Mount the real page and scanner-backed 33-request fixture used by the native
// paired captures, not a stand-in tree with predetermined row heights.
func TestUsageSourceDetailsMountedGeometry(t *testing.T) {
	for _, width := range []int{1470, 1100} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			a := usagePairedFixtureApp(t)
			a.usageState.mode = "details"
			tt := ui.NewTester(a.View, width, 922)
			tt.Frame()
			find := func(label string) ui.Rect {
				t.Helper()
				r, ok := tt.Find(label)
				if !ok {
					t.Fatal("missing", label)
				}
				return r
			}
			near := func(got, want float32, what string) {
				t.Helper()
				if math.Abs(float64(got-want)) > .01 {
					t.Fatalf("%s: got %.3f, want %.3f", what, got, want)
				}
			}
			theme := a.Theme
			kind := "breakdown:models"
			card := find("Usage card Breakdown")
			segments := find("Usage breakdown segments")
			chart := find("Usage breakdown chart " + kind)
			search := find("Usage table search head " + kind)
			table := find("Usage " + kind + " table")
			totals := find("Usage table totals " + kind)
			pager := find("Usage table pager " + kind)
			records := find("Usage card Records")
			near(card.W, min(1180, float32(width)-a.SidebarWidth-theme.Space(12)-2)-40, "responsive card width")
			near(card.Y, 188, "source Details band origin")
			near(segments.Y, card.Y+46+theme.Space(12), "segmented control below card header")
			near(segments.H, theme.Px(28), "medium segmented control")
			near(chart.Y, segments.Y+segments.H+theme.Space(12), "chart after tabs")
			rows := a.usageBreakdownChartRows(kind)
			if len(rows.Rows) != 4 {
				t.Fatal("real paired model identities changed", rows.Rows)
			}
			rowHeight := usageLineHeight(theme.Px(12)) + usageLineHeight(theme.Px(10)) + 2*theme.Space(4)
			for i, row := range rows.Rows {
				bar := find("Usage breakdown bar " + row.Key)
				track := find("Usage breakdown track " + row.Key)
				fill := find("Usage breakdown fill " + row.Key)
				near(bar.H, rowHeight, "source chart label line boxes")
				near(bar.Y, chart.Y+float32(i)*(rowHeight+theme.Space(2)), "source bar stride")
				near(track.Y, bar.Y+(bar.H-8)/2, "vertically centered track")
				near(track.X, bar.X+theme.Space(8)+180+theme.Space(8), "fixed identity column")
				if track.W < 48 || track.X+track.W > card.X+card.W || math.Abs(float64(fill.W/track.W)-a.usageBreakdownFraction(row, rows.Totals, false)) > .002 {
					t.Fatal("responsive track or real-data fraction changed", bar, track, fill)
				}
			}
			near(chart.H, float32(len(rows.Rows))*rowHeight+float32(len(rows.Rows)-1)*theme.Space(2)+theme.Space(12), "intrinsic chart height")
			near(search.Y, chart.Y+chart.H+theme.Space(12), "toolbar below chart")
			columns := find("Columns")
			near(columns.Y, search.Y+(search.H-columns.H)/2, "Columns centered beside search")
			near(table.Y, search.Y+search.H+theme.Space(12), "one toolbar/table gap, no measurement spacer")
			near(table.H, 40*float32(len(rows.Rows))+42, "page-sized table")
			near(totals.Y, table.Y+table.H+theme.Space(12), "totals following table")
			near(totals.H, theme.Space(8)+usageLineHeight(theme.Px(12)), "source totals inset and line height")
			near(pager.Y, totals.Y+totals.H+theme.Space(12), "pager following totals")
			near(pager.H, theme.Space(12)+theme.Px(28), "source pager inset and medium controls")
			near(card.Y+card.H, pager.Y+pager.H+theme.Space(12)+1, "content-derived card bottom")
			near(records.Y, card.Y+card.H+theme.Space(20), "Records follows Breakdown")
			if width == 1470 {
				// Rasterized source edges round to 845/865. These are a check on
				// the token-derived geometry above, never production constants.
				if math.Round(float64(card.Y+card.H)) != 845 || math.Round(float64(records.Y)) != 865 {
					t.Fatal("paired source card boundaries", card, records)
				}
			}
			var prior ui.Rect
			for i, label := range []string{"Models", "Workspaces", "Providers", "Tools"} {
				r := find(label)
				near(r.H, theme.Px(28), "joined segment height")
				if i > 0 {
					near(r.X, prior.X+prior.W, "gapless source segments")
				}
				pixel := color.NRGBAModel.Convert(tt.Image().At(int(r.X+6), int(r.Y+4))).(color.NRGBA)
				if math.Abs(float64(pixel.B)-float64(pixel.R)) > 2 {
					t.Fatal("Details tabs must use neutral fill, not blue selection", label, pixel)
				}
				prior = r
			}
			for _, col := range []string{"requests", "input", "output", "cache", "tokens"} {
				label := find("Usage header label " + kind + " " + col)
				text := find("Usage header text " + kind + " " + col)
				caret := find("Sort carets " + kind + " " + col)
				near(caret.X, label.X+label.W+theme.Space(6), "source in-flow sort caret")
				if label.W <= 0 || text.W != label.W || math.Abs(float64(text.H-usageLineHeight(theme.Px(14)))) > .01 {
					t.Fatal("header label should retain one source line inside its clip", col, label, text)
				}
			}
			if a.usageState.projection.Summary.Requests != 33 || a.usageState.projection.Summary.Tokens.Total != 55360 || !tt.HasText("Filtered totals: 33 requests · 55.4K tokens · $0.91") {
				t.Fatal("real fixture totals changed", tt.Texts())
			}
			if out := os.Getenv("ORBIT_CAPTURE_DIR"); out != "" {
				if err := os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(filepath.Join(out, fmt.Sprintf("usage-details-33-request-%d-cpu.png", width)))
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				if err = png.Encode(f, tt.Image()); err != nil {
					t.Fatal(err)
				}
			}

			// Searching the mounted table changes its chart and natural height,
			// not the range-wide accounting or a hardcoded four-row layout.
			before := a.usageState.projection.Summary
			if err := tt.Click("Search " + kind); err != nil {
				t.Fatal(err)
			}
			tt.Type("Fixture beta")
			tt.Frame()
			if len(a.usageBreakdownChartRows(kind).Rows) != 2 || a.usageState.projection.Summary != before {
				t.Fatal("search did not reduce rows independently of global totals")
			}
			filteredCard := find("Usage card Breakdown")
			near(card.H-filteredCard.H, 2*(rowHeight+theme.Space(2)+40), "search-driven intrinsic height")
			near(find("Usage card Records").Y, filteredCard.Y+filteredCard.H+theme.Space(20), "Records reflows after search")
		})
	}
}

func TestUsageSourceDetailsSegmentsRetainTableState(t *testing.T) {
	a := usagePairedFixtureApp(t)
	s := &a.usageState
	s.mode = "details"
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	models := s.table("breakdown:models")
	models.Search, models.Hidden = "Fixture beta", []string{"input"}
	models.List.Columns.Widths["name"] = 350
	tt.Frame()
	if err := tt.Click("Requests"); err != nil {
		t.Fatal(err)
	}
	wantSort := models.Sort
	for _, label := range []string{"Providers", "Models", "Models"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	if s.breakdown != "models" || models.Search != "Fixture beta" || len(models.Hidden) != 1 || models.Hidden[0] != "input" || models.List.Columns.Widths["name"] != 350 || models.Sort != wantSort {
		t.Fatal("segment changes lost independent table state", s.breakdown, models)
	}
	providers := s.table("breakdown:providers")
	if providers.Search != "" || len(providers.Hidden) != 0 || providers.List.Columns.Widths["name"] != 350 || providers.Sort != wantSort {
		t.Fatal("dimension state or shared width/sort changed", providers)
	}
	// Records uses the same source segmented treatment and saves its own tab.
	s.scroll.Y = 300
	tt.Frame()
	for _, label := range []string{"Daily", "Sessions", "Sessions"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	if s.detail != "sessions" || s.breakdown != "models" || models.Search != "Fixture beta" {
		t.Fatal("Records selector disturbed Breakdown state")
	}
	var saved struct {
		Breakdown string `json:"breakdown_tab"`
		Detail    string `json:"detail_tab"`
	}
	if err := a.Options.Store.Read("usage.json", &saved); err != nil || saved.Breakdown != "models" || saved.Detail != "sessions" {
		t.Fatal("Details tab persistence", saved, err)
	}
}

// The pinned native header clips its nowrap content at the cell boundary.
// Compare that treatment with the mounted page so a later SingleLine change
// cannot silently reserve an ellipsis and remove extra header characters.
// This is a CPU rendering regression, not a native screenshot acceptance gate.
func TestUsageSourceDetailsHeaderClipsWithoutEarlyEllipsis(t *testing.T) {
	a := usagePairedFixtureApp(t)
	a.usageState.mode = "details"
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	label, ok := tt.Find("Usage header label breakdown:models requests")
	if !ok || label.W != 44 {
		t.Fatal("source Requests content width", label, ok)
	}
	reference := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Box(c).Fill().Background(a.Theme.Elevated).Children(func() {
			ui.Box(c).Absolute().Left(label.X).Top(label.Y).Width(label.W).Height(label.H).Clip().Children(func() {
				ui.Text(c, a.usageColumnTitle("breakdown:models", "requests", "tokens", false)).NoWrap().Shrink(0).FontSize(a.Theme.Px(14)).FixedLineHeight(usageLineHeight(a.Theme.Px(14))).FontWeight(500).TextColor(a.Theme.Text)
			})
		})
	}, 1470, 922)
	reference.Frame()
	actual, want := tt.Image(), reference.Image()
	for y := int(math.Ceil(float64(label.Y))); y < int(label.Y+label.H); y++ {
		for x := int(label.X); x < int(label.X+label.W); x++ {
			if actual.RGBAAt(x, y) != want.RGBAAt(x, y) {
				t.Fatalf("header no longer matches source nowrap clipping at %d,%d: got %v want %v", x, y, actual.RGBAAt(x, y), want.RGBAAt(x, y))
			}
		}
	}
}
