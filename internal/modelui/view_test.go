package modelui

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func nativeTester(s *State, mode string, width, height int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme := ui.DarkTheme()
		theme.FontSize = 14
		c.SetTheme(theme)
		ui.Column(c).Grow(1).AlignSelf(ui.Stretch).Background(theme.Background).LineHeight(1.618034).Children(func() {
			switch mode {
			case "models":
				s.Toolbar(c)
				ui.Scroll(c).Grow(1).AlignSelf(ui.Stretch).Children(func() { s.View(c) })
			case "picker":
				s.PickerView(c)
			case "defaults":
				ui.Row(c).Gap(12).Children(func() { s.DefaultModelControl(c); s.DefaultThinkingControl(c); s.TitleModelControl(c) })
			}
		})
	}, width, height)
}
func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
}
func TestNativeStarsNeverSelectOrCloseAndCardsDispatchOnce(t *testing.T) {
	for _, mode := range []string{"models", "picker"} {
		t.Run(mode, func(t *testing.T) {
			s := fixture(t)
			calls := 0
			s.Options.OnSelectModel = func(string, string) error { calls++; return nil }
			if mode == "picker" {
				s.PickerOpen(ModelPicker)
			}
			tt := nativeTester(s, mode, 900, 700)
			tt.Frame()
			click(t, tt, "Favorite first / alpha")
			if mode == "picker" && !tt.Focused("Model picker search") {
				t.Fatal("star stole search keyboard focus")
			}
			if calls != 0 || !s.IsFavorite(s.Snapshot.Models[0]) || mode == "picker" && !s.PickerIsOpen() {
				t.Fatal("star bubbled to selection/dismissal", calls, s.PickerIsOpen())
			}
			click(t, tt, "Unfavorite first / alpha")
			if calls != 0 || s.IsFavorite(s.Snapshot.Models[0]) {
				t.Fatal("repeat star wrong")
			}
			label := "Select model first / alpha"
			if mode == "picker" {
				label = "Choose Alpha"
			}
			click(t, tt, label)
			if calls != 1 {
				t.Fatal("single click emitted duplicate request", calls)
			}
		})
	}
}
func TestNativePickerSearchFocusArrowFilterAndEscape(t *testing.T) {
	s := fixture(t)
	var calls []string
	s.Options.OnSelectModel = func(provider, id string) error { calls = append(calls, provider+"/"+id); return nil }
	s.PickerOpen(ModelPicker)
	tt := nativeTester(s, "picker", 600, 700)
	tt.Frame()
	if !tt.Focused("Model picker search") {
		t.Fatal("search did not own focus")
	}
	tt.Type("Beta")
	tt.Key(0, ui.KeyEnter)
	if len(calls) != 1 || calls[0] != "first/beta" || s.PickerIsOpen() {
		t.Fatal("typed filtered Enter didn't select once", calls, s.PickerQuery())
	}
	tt.Key(0, ui.KeyEnter)
	if len(calls) != 1 {
		t.Fatal("repeated Enter reused dismissed picker")
	}
	s.PickerOpen(ModelPicker)
	tt.Frame()
	tt.Type("no-such-model")
	tt.Key(0, ui.KeyEnter)
	if len(calls) != 1 || !s.PickerIsOpen() {
		t.Fatal("no-match Enter submitted/dismissed")
	}
	tt.Key(0, ui.KeyEscape)
	if s.PickerIsOpen() || len(calls) != 1 {
		t.Fatal("Escape selected")
	}
	s.PickerOpen(ModelPicker)
	tt.Frame()
	tt.Key(0, ui.KeyUp)
	before := s.Highlighted()
	s.SetSnapshot(s.Snapshot)
	tt.Frame()
	if before != s.Highlighted() {
		t.Fatal("snapshot reset arrow cursor")
	}
}
func TestNativeModelsFilterEmptyAndProviderMenuOriginalChoice(t *testing.T) {
	s := fixture(t)
	tt := nativeTester(s, "models", 1000, 700)
	tt.Frame()
	click(t, tt, "Search model catalog")
	tt.Type("no-such-model")
	if _, ok := tt.Find("No models match"); !ok {
		t.Fatal("filtered-empty notice missing")
	}
	if _, ok := tt.Find("Active model"); !ok {
		t.Fatal("empty filter removed active board")
	}
	s.Query = ""
	tt.Frame()
	click(t, tt, "All providers")
	if !tt.Focused("All providers filter") {
		t.Fatal("provider filter focus")
	}
	tt.Type("Second")
	tt.Key(0, ui.KeyEnter)
	if s.ProviderFilter != "second" || s.HasModal() {
		t.Fatal("filtered original provider index", s.ProviderFilter, s.selectMenu)
	}
	if groups := s.Groups(); len(groups) != 1 || groups[0].Provider != "second" {
		t.Fatal(groups)
	}
}
func TestNativeDefaultMenusFilteredEnterAndCanceledNoWrite(t *testing.T) {
	s := fixture(t)
	tt := nativeTester(s, "defaults", 1000, 700)
	tt.Frame()
	if r, ok := tt.Find("Default thinking"); !ok || r.W <= 0 {
		t.Fatal("disabled thinking control absent")
	}
	click(t, tt, "Default model")
	if !tt.Focused("Default model filter") {
		t.Fatal("default filter focus")
	}
	tt.Type("Beta")
	tt.Key(0, ui.KeyEnter)
	if s.Defaults.Provider != "first" || s.Defaults.ModelID != "beta" || s.HasModal() {
		t.Fatal("default filtered selection", s.Defaults)
	}
	click(t, tt, "Default thinking")
	tt.Type("Max")
	tt.Key(0, ui.KeyEnter)
	if s.Defaults.Thinking != "max" {
		t.Fatal("model-derived max not selectable", s.Defaults)
	}
	before := s.Defaults
	click(t, tt, "Default model")
	tt.Type("Alpha")
	tt.Key(0, ui.KeyEscape)
	if s.Defaults != before || s.HasModal() {
		t.Fatal("cancel applied preference")
	}
	click(t, tt, "Title model")
	tt.Type("Second Same")
	tt.Key(0, ui.KeyEnter)
	if s.Title.Model == nil || s.Title.Model.Provider != "second" || s.Title.Model.ID != "same" {
		t.Fatal("title provider/id", s.Title)
	}
}
func TestNativeModelsThreeColumnCardGeometry(t *testing.T) {
	s := fixture(t)
	tt := nativeTester(s, "models", 1100, 800)
	tt.Frame()
	var rects []ui.Rect
	for _, id := range []string{"alpha", "beta", "same"} {
		r, ok := tt.Find("Select model first / " + id)
		if !ok {
			t.Fatal(id)
		}
		rects = append(rects, r)
	}
	if rects[0].Y != rects[1].Y || rects[1].Y != rects[2].Y || rects[0].X >= rects[1].X || rects[1].X >= rects[2].X {
		t.Fatal("not three-column provider grid", rects)
	}
	if rects[0].W <= 0 || rects[1].W <= 0 || rects[2].W <= 0 || rects[2].X+rects[2].W > 1100 {
		t.Fatal("cards overflow", rects)
	}
	s.PickerOpen(ModelPicker)
	picker := nativeTester(s, "picker", 1000, 800)
	picker.Frame()
	if r, ok := picker.Find("Model picker search"); !ok || r.W > 360 {
		t.Fatal("model width", r)
	}
}
func TestNativePickerLongListKeyboardRevealAndThinkingGate(t *testing.T) {
	s := fixture(t)
	s.Snapshot.Models = nil
	for i := 0; i < 60; i++ {
		s.Snapshot.Models = append(s.Snapshot.Models, agent.ModelChoice{Provider: "fixture", ID: fmt.Sprintf("model-%02d", i), Name: fmt.Sprintf("Fixture %02d", i)})
	}
	s.Snapshot.Current = s.Snapshot.Models[55]
	s.PickerOpen(ModelPicker)
	tt := nativeTester(s, "picker", 600, 700)
	tt.Frame()
	tt.Frame()
	r, ok := tt.Find("Choose Fixture 55")
	if !ok || r.Y < 0 || r.Y+r.H > 600 {
		t.Fatal("current selection not revealed", r, s.picker.scroll.Y)
	}
	want := centeredOffset(s.PickerRows(), s.Highlighted(), 26*1.618+8, 24, 1, 6*(26*1.618+9))
	if diff := s.picker.scroll.Y - want; diff < -1 || diff > 1 {
		t.Fatal("current selection not source-centered", s.picker.scroll.Y, want)
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	r, ok = tt.Find("Choose Fixture 56")
	if !ok || r.Y < 0 || r.Y+r.H > 600 {
		t.Fatal("arrow selection not revealed", r, s.picker.scroll.Y)
	}
	for i := 0; i < 10; i++ {
		tt.Key(0, ui.KeyDown)
	}
	if row := s.PickerRows()[s.Highlighted()]; row.ModelIndex != 59 {
		t.Fatal("keyboard did not clamp final", row)
	}
	s.DismissPicker(false)
	s.Snapshot.Levels = []string{"off"}
	if s.PickerOpen(ThinkingPicker) {
		t.Fatal("unavailable thinking picker opened")
	}
}
func TestNativeModelLoadingErrorAndEmptyHaveNoFabricatedRows(t *testing.T) {
	for _, snapshot := range []Snapshot{{Loading: true}, {Connected: true}, {Error: "fixture catalog failed"}, {Connected: false}} {
		s := fixture(t)
		s.SetSnapshot(snapshot)
		tt := nativeTester(s, "models", 1000, 700)
		tt.Frame()
		if _, ok := tt.Find("Select model first / alpha"); ok {
			t.Fatal("stale/invented catalog row")
		}
		if snapshot.Loading {
			if _, ok := tt.Find("Loading models…"); !ok {
				t.Fatal("missing loading")
			}
		}
		if snapshot.Error != "" {
			if _, ok := tt.Find(snapshot.Error); !ok {
				t.Fatal("missing error")
			}
		}
	}
}

// Explicit synthetic data verifies native pixels/geometry only. It is not a
// live catalog, authenticated provider result, or matched Rust comparison.
func TestNativeModelUICaptures(t *testing.T) {
	out := os.Getenv("ORBIT_MODELUI_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_MODELUI_CAPTURE_DIR for explicit native CPU fixture captures")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	save := func(name string, s *State, mode string) {
		t.Helper()
		tt := nativeTester(s, mode, 1000, 760)
		tt.Frame()
		tt.Frame()
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
	s := fixture(t)
	save("models-catalog-fixture", s, "models")
	s.ToggleFavorite(2)
	s.FavoritesOnly = true
	save("models-favorites-fixture", s, "models")
	s.Query = "no-match"
	save("models-filter-empty-fixture", s, "models")
	s.Query = ""
	s.FavoritesOnly = false
	s.recordRecent(s.Snapshot.Models[1])
	s.PickerOpen(ModelPicker)
	save("model-picker-groups-fixture", s, "picker")
	s.SetScope(Favorites)
	save("model-picker-favorites-fixture", s, "picker")
	s.DismissPicker(false)
	s.PickerOpen(ThinkingPicker)
	save("thinking-picker-fixture", s, "picker")
	if !strings.Contains(s.thinkingHint("high"), "Deeper") {
		t.Fatal("fixture hint")
	}
}

func TestNativeScopeClickKeepsSearchKeyboardOwnership(t *testing.T) {
	s := fixture(t)
	s.PickerOpen(ModelPicker)
	tt := nativeTester(s, "picker", 600, 700)
	tt.Frame()
	click(t, tt, "First")
	if !tt.Focused("Model picker search") {
		t.Fatal("scope click stole search focus")
	}
	before := s.Highlighted()
	tt.Key(0, ui.KeyDown)
	if s.Highlighted() == before {
		t.Fatal("scope click disabled arrow handling")
	}
	tt.Type("Beta")
	if rows := modelIndices(s.PickerRows()); len(rows) != 1 || rows[0] != 2 {
		t.Fatal("typing after scope failed", rows)
	}
}

func TestNativeSettingsNoMatchesAndOutsidePressDoNotSave(t *testing.T) {
	s := fixture(t)
	tt := nativeTester(s, "defaults", 1000, 700)
	tt.Frame()
	click(t, tt, "Default model")
	tt.Type("no-result")
	tt.Key(0, ui.KeyEnter)
	if !s.HasModal() || s.Defaults.ModelID != "" {
		t.Fatal("empty settings Enter applied")
	}
	if _, ok := tt.Find("No matches"); !ok {
		t.Fatal("missing no-match hint")
	}
	tt.ClickAt(900, 600)
	if s.HasModal() || s.Defaults.ModelID != "" {
		t.Fatal("outside press applied/failed dismissal")
	}
}
func TestNativeStationaryPointerCannotHijackPickerKeyboard(t *testing.T) {
	s := fixture(t)
	s.PickerOpen(ModelPicker)
	s.picker.suppressHoverUntil = time.Time{}
	tt := nativeTester(s, "picker", 600, 700)
	tt.Frame()
	r, ok := tt.Find("Choose Alpha")
	if !ok {
		t.Fatal("alpha row")
	}
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Move(r.X+r.W/2+1, r.Y+r.H/2)
	tt.Key(0, ui.KeyDown)
	before := s.Highlighted()
	tt.Frame()
	tt.Frame()
	if s.Highlighted() != before {
		t.Fatal("stationary pointer stole deliberate key cursor")
	}
}

func TestNativeSettingsRepeatedTriggerTogglesWithoutApplying(t *testing.T) {
	s := fixture(t)
	tt := nativeTester(s, "defaults", 1000, 700)
	tt.Frame()
	for i := 0; i < 3; i++ {
		click(t, tt, "Default model")
		if !s.HasModal() {
			t.Fatal("open", i)
		}
		click(t, tt, "Default model")
		if s.HasModal() || s.Defaults.ModelID != "" {
			t.Fatal("trigger close applied/reopened", i)
		}
	}
}

func TestNativeHorizontalScopesKeepFavoriteCurrentReachable(t *testing.T) {
	s := fixture(t)
	for i := 0; i < 12; i++ {
		s.Snapshot.Models = append(s.Snapshot.Models, agent.ModelChoice{ID: "fixture", Provider: fmt.Sprintf("provider-%02d", i), Name: "Fixture"})
	}
	s.PickerOpen(ModelPicker)
	tt := nativeTester(s, "picker", 600, 700)
	tt.Frame()
	r, ok := tt.Find("Favorite current")
	if !ok || r.X < 360 {
		t.Fatal("fixture should overflow horizontal scope viewport", r)
	}
	tt.Scroll(170, 52, 10000, 0)
	tt.Frame()
	r, ok = tt.Find("Favorite current")
	if !ok || r.X < 0 || r.X+r.W > 361 {
		t.Fatal("trailing scope action not reachable by horizontal scroll", r)
	}
	click(t, tt, "Favorite current")
	if !s.IsFavorite(s.Snapshot.Models[1]) || !s.PickerIsOpen() || !tt.Focused("Model picker search") {
		t.Fatal("favorite-current action selected/closed/stole focus")
	}
}

func TestNativeRefreshGlyphAndReducedMotion(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(fmt.Sprint(reduced), func(t *testing.T) {
			s := fixture(t)
			s.Options.Theme = func() Theme { return Theme{ReduceMotion: reduced} }
			s.Refreshing = true
			tt := nativeTester(s, "models", 1000, 700)
			tt.Frame()
			if _, ok := tt.Find("Catalog update progress"); !ok {
				t.Fatal("source refresh glyph absent")
			}
			first := bytes.Clone(tt.Image().Pix)
			tt.Advance(175 * time.Millisecond)
			same := bytes.Equal(first, tt.Image().Pix)
			if same != reduced {
				t.Fatalf("refresh motion=%v; reduced=%v", !same, reduced)
			}
			calls := 0
			s.Options.OnRefresh = func() error { calls++; return nil }
			click(t, tt, "Update model list")
			if calls != 0 || !s.Refreshing {
				t.Fatal("busy visual accepted duplicate refresh")
			}
		})
	}
}
