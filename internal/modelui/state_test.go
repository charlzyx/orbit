package modelui

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
)

func fixture(t *testing.T) *State {
	t.Helper()
	s := New(preferences.New(t.TempDir()), Options{})
	s.SetSnapshot(Snapshot{Connected: true, Models: []agent.ModelChoice{
		{ID: "alpha", Provider: "first", Name: "Alpha", ContextWindow: 128000},
		{ID: "same", Provider: "second", Name: "Second Same", Reasoning: true, ContextWindow: 200000},
		{ID: "beta", Provider: "first", Name: "Beta", Reasoning: true, ContextWindow: 1000000, Raw: json.RawMessage(`{"thinkingLevelMap":{"off":null,"minimal":null,"low":"low","medium":null,"high":"high","max":"max"}}`)},
		{ID: "same", Provider: "first", Name: "First Same", Reasoning: true},
	}, Current: agent.ModelChoice{ID: "same", Provider: "second", Name: "Second Same"}, Levels: []string{"off", "low", "high"}, CurrentLevel: "high"})
	return s
}
func TestCatalogGroupsFiltersCountsAndExactIdentity(t *testing.T) {
	s := fixture(t)
	if !slices.Equal(s.Providers(), []string{"first", "second"}) {
		t.Fatal(s.Providers())
	}
	if got := s.Groups(); len(got) != 2 || !slices.Equal(got[0].Models, []int{0, 2, 3}) || !slices.Equal(got[1].Models, []int{1}) {
		t.Fatal(got)
	}
	if s.IsActive(s.Snapshot.Models[3]) || !s.IsActive(s.Snapshot.Models[1]) {
		t.Fatal("same id crossed providers")
	}
	if !s.ToggleFavorite(2) || !s.ToggleFavorite(1) {
		t.Fatal(s.Error)
	}
	if s.CountLabel() != "4 models · 2 favorite(s)" {
		t.Fatal(s.CountLabel())
	}
	s.Query = "  BeTa  "
	s.ProviderFilter = "first"
	s.FavoritesOnly = true
	if got := s.Groups(); len(got) != 1 || !slices.Equal(got[0].Models, []int{2}) {
		t.Fatal(got)
	}
	if s.CountLabel() != "1 of 4 shown · 2 favorite(s)" {
		t.Fatal(s.CountLabel())
	}
	s.Query = "no-result"
	if len(s.Groups()) != 0 || s.CountLabel() != "0 of 4 shown · 2 favorite(s)" {
		t.Fatal(s.Groups(), s.CountLabel())
	}
	s.Query = ""
	s.ProviderFilter = ""
	s.FavoritesOnly = false
	if len(s.Snapshot.Models) != 4 || len(s.Groups()) != 2 {
		t.Fatal("filters mutated catalog")
	}
	// Stale favorites are persisted identities, never fabricated catalog entries.
	if err := s.store.SetModel(false, preferences.Model{Provider: "missing", ID: "missing"}, true); err != nil {
		t.Fatal(err)
	}
	s = New(s.store, Options{})
	if len(s.Snapshot.Models) != 0 || s.CountLabel() != "No models reported by the runtime" {
		t.Fatal("invented catalog")
	}
}
func modelIndices(rows []Row) []int {
	var out []int
	for _, r := range rows {
		if r.Selectable() {
			out = append(out, r.ModelIndex)
		}
	}
	return out
}
func TestPickerSourceGroupingScopesAndDedup(t *testing.T) {
	s := fixture(t)
	for _, i := range []int{0, 1, 2} {
		if !s.ToggleFavorite(i) {
			t.Fatal(s.Error)
		}
	}
	s.recordRecent(s.Snapshot.Models[1])
	s.recordRecent(s.Snapshot.Models[2])
	s.PickerOpen(ModelPicker)
	rows := s.PickerRows()
	if !slices.Equal(modelIndices(rows), []int{2, 1, 0, 3}) || rows[0].Header != "recents" || rows[3].Header != "favorites" || rows[5].Header != "provider" {
		t.Fatal(rows)
	}
	s.SetScope(Recents)
	if got := modelIndices(s.PickerRows()); !slices.Equal(got, []int{2, 1}) {
		t.Fatal(got)
	}
	s.SetScope(Favorites)
	if got := modelIndices(s.PickerRows()); !slices.Equal(got, []int{0, 2, 1}) {
		t.Fatal(got)
	}
	// Provider-scope recents deliberately use catalog order, not MRU order.
	s.recordRecent(s.Snapshot.Models[0])
	s.SetScope(ProviderScope("first"))
	rows = s.PickerRows()
	if !slices.Equal(modelIndices(rows), []int{0, 2, 3}) || rows[0].Header != "recents" {
		t.Fatal(rows)
	}
	for _, r := range rows {
		if r.Header == "provider" {
			t.Fatal("provider scope duplicated header")
		}
	}
	s.SetPickerQuery("beta")
	if !slices.Equal(modelIndices(s.PickerRows()), []int{2}) {
		t.Fatal(s.PickerRows())
	}
}
func TestPickerMatchingNormalizationAndProviderGuard(t *testing.T) {
	m := agent.ModelChoice{ID: "minimax-m2.5", Name: "MiniMax-M2.5", Provider: "opencode-go"}
	for _, current := range []agent.ModelChoice{{ID: "minimax-m2.5", Provider: "opencode-go"}, {Name: "MiniMax M2.5", Provider: "OPENCODE-GO"}, {Name: "MiniMax M2.5"}} {
		if !IsPickerSelected(m, current) {
			t.Fatal(current)
		}
	}
	if IsPickerSelected(m, agent.ModelChoice{ID: m.ID, Provider: "other"}) {
		t.Fatal("crossed provider")
	}
}
func TestPickerIdentityOnlyFoldsASCII(t *testing.T) {
	if !equalASCIIFold("LOW", "low") || !equalASCIIFold("提供商", "提供商") {
		t.Fatal("lost ASCII or exact identity")
	}
	if equalASCIIFold("K", "K") || equalASCIIFold("s", "ſ") {
		t.Fatal("Unicode lookalikes collapsed distinct identifiers")
	}
	if IsPickerSelected(agent.ModelChoice{ID: "same", Provider: "K"}, agent.ModelChoice{ID: "same", Provider: "K"}) {
		t.Fatal("crossed custom provider identity")
	}
}

func TestPickerKeyboardClampsSkipsHeadersAndUnchangedSnapshot(t *testing.T) {
	s := fixture(t)
	s.PickerOpen(ModelPicker)
	rows := s.PickerRows()
	if !rows[s.Highlighted()].Selected {
		t.Fatal("did not pin current", rows, s.Highlighted())
	}
	s.StepPicker(-1)
	moved := s.Highlighted()
	s.SetSnapshot(s.Snapshot)
	if s.Highlighted() != moved {
		t.Fatal("unchanged snapshot stole navigation")
	}
	for i := 0; i < 100; i++ {
		s.StepPicker(-1)
	}
	if s.Highlighted() != firstSelectable(s.PickerRows()) {
		t.Fatal("did not clamp start")
	}
	for i := 0; i < 100; i++ {
		s.StepPicker(1)
	}
	rows = s.PickerRows()
	if s.Highlighted() != len(rows)-1 || !rows[s.Highlighted()].Selectable() {
		t.Fatal("did not clamp end", s.Highlighted())
	}
	s.SetPickerQuery("no-result")
	if len(s.PickerRows()) != 0 || s.ConfirmPicker() {
		t.Fatal("empty filter selected")
	}
	s.SetPickerQuery("beta")
	rows = s.PickerRows()
	if rows[s.Highlighted()].ModelIndex != 2 {
		t.Fatal("query didn't choose first", rows)
	}
	s.SetPickerQuery("")
	if !s.PickerRows()[s.Highlighted()].Selected {
		t.Fatal("query clear didn't repin")
	}
	s.SetScope(ProviderScope("second"))
	next := s.Snapshot
	next.Models = next.Models[:1]
	s.SetSnapshot(next)
	if s.PickerScope() != All || s.Highlighted() != 1 {
		t.Fatal("stale provider didn't reset", s.PickerScope(), s.Highlighted())
	}
}
func TestFavoritesPersistenceFailureAndNoSelection(t *testing.T) {
	s := fixture(t)
	calls := 0
	s.Options.OnSelectModel = func(string, string) error { calls++; return nil }
	s.PickerOpen(ModelPicker)
	if !s.ToggleFavorite(3) {
		t.Fatal(s.Error)
	}
	reloaded := New(s.store, Options{})
	if !slices.Equal(reloaded.Favorites(), []preferences.Model{{Provider: "first", ID: "same"}}) {
		t.Fatal(reloaded.Favorites())
	}
	if calls != 0 || !s.PickerIsOpen() || len(s.Recents()) != 0 {
		t.Fatal("favorite selected/dismissed")
	}
	if !s.ToggleFavorite(3) || len(s.Favorites()) != 0 {
		t.Fatal("unfavorite")
	}
	path := filepath.Join(s.store.RootDir, "favorites.json")
	if err := os.WriteFile(path, []byte("{malformed"), 0600); err != nil {
		t.Fatal(err)
	}
	if s.ToggleFavorite(1) || s.Error == "" || len(s.Favorites()) != 0 {
		t.Fatal("failed save changed confirmed favorite")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "{malformed" {
		t.Fatal("overwrote malformed data")
	}
}
func TestModelDispatchAndRecentsDoNotFabricateSelection(t *testing.T) {
	s := fixture(t)
	var calls []preferences.Model
	s.Options.OnSelectModel = func(provider, id string) error {
		calls = append(calls, preferences.Model{Provider: provider, ID: id})
		return nil
	}
	before := s.Snapshot.Current
	if !s.SelectModel(3, false) || len(s.Recents()) != 0 {
		t.Fatal("settings recorded recents")
	}
	if s.Snapshot.Current.ID != before.ID || s.Snapshot.Current.Provider != before.Provider {
		t.Fatal("pretended Pi acknowledged")
	}
	s.PickerOpen(ModelPicker)
	s.SetPickerQuery("beta")
	if !s.ConfirmPicker() || s.PickerIsOpen() || s.ConfirmPicker() {
		t.Fatal("confirmation repeated after close")
	}
	if len(calls) != 2 || calls[1].ID != "beta" || s.Recents()[0].ID != "beta" {
		t.Fatal(calls, s.Recents())
	}
	p := filepath.Join(s.store.RootDir, "recent-models.json")
	stamp := time.Unix(10, 0)
	if err := os.Chtimes(p, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	s.recordRecent(s.Snapshot.Models[2])
	st, _ := os.Stat(p)
	if !st.ModTime().Equal(stamp) {
		t.Fatal("leading recent unnecessarily rewritten")
	}
	s.Options.OnSelectModel = func(string, string) error { return errors.New("runtime disconnected") }
	s.PickerOpen(ModelPicker)
	s.SetPickerQuery("alpha")
	if s.ConfirmPicker() || !s.PickerIsOpen() || !strings.Contains(s.Error, "disconnected") {
		t.Fatal("dispatch rejection lost retry", s.Error)
	}
}
func TestRecentsCapStaleModelsAndCancellation(t *testing.T) {
	s := fixture(t)
	for i := 0; i < 8; i++ {
		s.recordRecent(agent.ModelChoice{Provider: "p", ID: string(rune('a' + i))})
	}
	if len(s.Recents()) != 5 || s.Recents()[0].ID != "h" {
		t.Fatal(s.Recents())
	}
	s.PickerOpen(ModelPicker)
	s.SetScope(Recents)
	if len(s.PickerRows()) != 0 {
		t.Fatal("stale recent invented model")
	}
	n := 0
	s.Options.OnDismiss = func(mouse bool) {
		if mouse {
			n++
		}
	}
	s.DismissPicker(true)
	s.DismissPicker(true)
	if n != 1 || s.Escape() {
		t.Fatal("repeat dismissal")
	}
}
func TestThinkingCapabilitiesAndSourceDefaults(t *testing.T) {
	s := fixture(t)
	if got := CatalogThinkingLevels(s.Snapshot.Models[0]); !slices.Equal(got, []string{"off"}) {
		t.Fatal(got)
	}
	if got := CatalogThinkingLevels(s.Snapshot.Models[1]); !slices.Equal(got, []string{"off", "minimal", "low", "medium", "high"}) {
		t.Fatal(got)
	}
	if got := CatalogThinkingLevels(s.Snapshot.Models[2]); !slices.Equal(got, []string{"low", "high", "max"}) {
		t.Fatal(got)
	}
	if !s.ChooseDefaultModel(1) || !s.ChooseDefaultThinking("medium") {
		t.Fatal(s.Error)
	}
	if !s.ChooseDefaultModel(2) || s.Defaults.Thinking != "" {
		t.Fatal("unsupported thinking survived", s.Defaults)
	}
	if s.ChooseDefaultThinking("xhigh") || !s.ChooseDefaultThinking("max") {
		t.Fatal("extended capability ignored")
	}
	if !s.ChooseDefaultModel(-1) || s.Defaults != (preferences.SessionDefault{}) || s.ChooseDefaultThinking("high") {
		t.Fatal("Pi default not clear/disabled")
	}
	if _, err := os.Stat(filepath.Join(s.store.PiDir, "settings.json")); !os.IsNotExist(err) {
		t.Fatal("wrote Pi global settings")
	}
	if s.Snapshot.Current.Provider != "second" {
		t.Fatal("default changed current")
	}
}
func TestDefaultsPersistenceFailureAndTitleSelection(t *testing.T) {
	s := fixture(t)
	if !s.ChooseDefaultModel(2) || !s.ChooseDefaultThinking("high") {
		t.Fatal(s.Error)
	}
	next := New(s.store, Options{})
	if next.Defaults != s.Defaults {
		t.Fatal("defaults didn't survive reload")
	}
	before := s.Defaults
	root := s.store.RootDir
	s.store.RootDir = filepath.Join(root, "not-directory")
	if err := os.WriteFile(s.store.RootDir, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if s.ChooseDefaultModel(0) || s.Defaults != before || s.Error == "" {
		t.Fatal("failed write changed confirmed defaults")
	}
	s.store.RootDir = root
	if err := s.store.SaveAutoTitle(preferences.AutoTitle{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if !s.ChooseTitleModel(3) || s.Title.Enabled || s.Title.Model == nil || s.Title.Model.Provider != "first" {
		t.Fatal(s.Title, s.Error)
	}
	if !s.ChooseTitleModel(-1) || s.Title.Model != nil || s.Title.Enabled {
		t.Fatal("title sentinel did not preserve disabled flag")
	}
}
func TestThinkingPickerReportedLevelsAndHintSearch(t *testing.T) {
	s := fixture(t)
	var selected string
	s.Options.OnSelectThinking = func(level string) error { selected = level; return nil }
	s.Snapshot.Levels = []string{"off"}
	if s.PickerOpen(ThinkingPicker) {
		t.Fatal("one choice opened picker")
	}
	s.Snapshot.Levels = []string{"off", "low", "high"}
	if !s.PickerOpen(ThinkingPicker) || s.PickerWidth() != 300 {
		t.Fatal("thinking width/open")
	}
	s.SetPickerQuery("deeper")
	rows := s.PickerRows()
	if len(rows) != 1 || rows[0].Level != "high" {
		t.Fatal(rows)
	}
	if !s.ConfirmPicker() || selected != "high" || s.PickerIsOpen() {
		t.Fatal("raw choice or close")
	}
	if s.SelectThinking("max") {
		t.Fatal("unreported level accepted")
	}
}
func TestRefreshLifecycleGateAndSnapshotCopies(t *testing.T) {
	s := fixture(t)
	calls := 0
	s.Options.OnRefresh = func() error { calls++; return nil }
	s.Snapshot.Running = true
	if s.Refresh() || calls != 0 {
		t.Fatal("mid-turn update")
	}
	s.Snapshot.Running = false
	if !s.Refresh() || s.Refresh() || calls != 1 || !s.Refreshing {
		t.Fatal("duplicate gate")
	}
	s.SetRefreshState(false, "fixture process failed")
	if s.RefreshError == "" || len(s.Snapshot.Models) != 4 {
		t.Fatal("refresh erased old catalog")
	}
	s.Options.OnRefresh = func() error { return errors.New("missing pi") }
	if s.Refresh() || s.Refreshing || s.RefreshError != "missing pi" {
		t.Fatal("failure not actionable")
	}
	input := s.Snapshot
	s.SetSnapshot(input)
	input.Models[0].Name = "mutated"
	input.Models[2].Raw[0] = 'x'
	input.Levels[0] = "fake"
	if s.Snapshot.Models[0].Name == "mutated" || s.Snapshot.Models[2].Raw[0] == 'x' || s.Snapshot.Levels[0] == "fake" {
		t.Fatal("snapshot retained mutable payload")
	}
	s.Close()
	if s.Refresh() || s.PickerOpen(ModelPicker) || s.ToggleFavorite(0) {
		t.Fatal("closed state mutated")
	}
}
func TestSearchableSettingsOriginalIndicesAndClamping(t *testing.T) {
	options := []choice{{"Pi default", -1}, {"First Alpha", 0}, {"Second Same", 1}, {"First Beta", 2}}
	visible := visibleChoices(options, "first")
	if !slices.Equal(visible, []int{1, 3}) {
		t.Fatal(visible)
	}
	for _, c := range []struct{ current, dir, want int }{{1, -1, 1}, {1, 1, 3}, {3, 1, 3}, {2, 1, 1}, {2, -1, 3}} {
		if got := stepVisible(visible, c.current, c.dir); got != c.want {
			t.Fatal(c, got)
		}
	}
	if stepVisible(nil, 0, 1) != -1 {
		t.Fatal("empty options selected")
	}
	if !reflect.DeepEqual(options, []choice{{"Pi default", -1}, {"First Alpha", 0}, {"Second Same", 1}, {"First Beta", 2}}) {
		t.Fatal("filter mutated options")
	}
}

func TestSourceCenteredScrollAccountsForHeadersAndClamps(t *testing.T) {
	rows := []Row{{Header: "provider"}, {ModelIndex: 0}, {ModelIndex: 1}, {Header: "provider"}, {ModelIndex: 2}, {ModelIndex: 3}, {ModelIndex: 4}, {ModelIndex: 5}}
	// Full content = 2*24 + 6*50 + 7 gaps = 355; viewport306.
	if got := centeredOffset(rows, 1, 50, 24, 1, 306); got != 0 {
		t.Fatal(got)
	}
	if got := centeredOffset(rows, 7, 50, 24, 1, 306); got != 49 {
		t.Fatal(got)
	}
	if got := centeredOffset(nil, 0, 50, 24, 1, 306); got != 0 {
		t.Fatal(got)
	}
}

func TestRunningLocksOnlyComposerPickerAndReloadCapsRecents(t *testing.T) {
	s := fixture(t)
	s.Snapshot.Running = true
	if s.PickerOpen(ModelPicker) || s.PickerOpen(ThinkingPicker) {
		t.Fatal("running turn opened composer picker")
	}
	// Settings dispatch matches the source; Pi, not the catalog UI, owns whether
	// an already-running session accepts a direct model change.
	calls := 0
	s.Options.OnSelectModel = func(string, string) error { calls++; return nil }
	if !s.SelectModel(0, false) || calls != 1 {
		t.Fatal("settings path gained extra gate")
	}
	var models []preferences.Model
	for i := 0; i < 9; i++ {
		models = append(models, preferences.Model{Provider: "fixture", ID: string(rune('a' + i))})
	}
	if err := s.store.Write("recent-models.json", map[string]any{"models": models}); err != nil {
		t.Fatal(err)
	}
	if got := New(s.store, Options{}).Recents(); len(got) != 5 || got[0].ID != "a" || got[4].ID != "e" {
		t.Fatal("source load cap/order", got)
	}
}
