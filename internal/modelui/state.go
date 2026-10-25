// Package modelui implements Orbit's source-derived native model catalog,
// composer pickers and model-related settings. It never invents catalog entries
// or starts Pi processes: the host supplies immutable RPC facts and actions.
package modelui

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type Theme struct {
	ReduceMotion                                                          bool
	CodePx                                                                func(float32) float32
	Elevated, Secondary, Active, Hover, BorderStrong, AccentText, Overlay ui.Color
	CodeFont                                                              string
	Px                                                                    func(float32) float32
	Space                                                                 func(int) float32
}
type Options struct {
	Theme  func() Theme
	Text   func(key, fallback string) string
	Notify func(text string, isError bool)
	// A nil error means accepted for dispatch, NOT confirmed by Pi.
	OnSelectModel     func(provider, id string) error
	OnSelectThinking  func(level string) error
	OnRefresh         func() error
	OnDismiss         func(byMouse bool)
	OnDefaultsChanged func(preferences.SessionDefault)
	OnTitleChanged    func(preferences.AutoTitle)
}
type Snapshot struct {
	Models                      []agent.ModelChoice
	Current                     agent.ModelChoice
	Levels                      []string
	CurrentLevel                string
	Connected, Loading, Running bool
	Error                       string
}
type PickerKind string

const (
	ModelPicker    PickerKind = "model"
	ThinkingPicker PickerKind = "thinking"
)

type Scope struct{ Kind, Provider string }

var All = Scope{Kind: "all"}
var Recents = Scope{Kind: "recents"}
var Favorites = Scope{Kind: "favorites"}

func ProviderScope(provider string) Scope { return Scope{Kind: "provider", Provider: provider} }

type Row struct {
	Header, Provider, Level string
	Count, ModelIndex       int
	Selected                bool
}

func (r Row) Selectable() bool { return r.Header == "" }

type Group struct {
	Provider string
	Models   []int
}
type pickerState struct {
	open               bool
	kind               PickerKind
	query, lastQuery   string
	scope              Scope
	cursor             int
	scroll             ui.ScrollState
	scrollPending      bool
	suppressHoverUntil time.Time
	pointerX, pointerY float32
	pointerKnown       bool
	pointerMoved       bool
}
type selectState struct {
	list        ui.ListState
	kind        string
	query       string
	cursor      int
	scroll      ui.ScrollState
	reveal      bool
	dismissedAt time.Time
}
type State struct {
	Options               Options
	Snapshot              Snapshot
	Query, ProviderFilter string
	FavoritesOnly         bool
	Refreshing            bool
	Error, RefreshError   string
	Defaults              preferences.SessionDefault
	Title                 preferences.AutoTitle
	store                 *preferences.Store
	favorites, recents    []preferences.Model
	picker                pickerState
	selectMenu            selectState
	closed                bool
}

func New(store *preferences.Store, options Options) *State {
	if store == nil {
		panic("modelui: explicit preferences store required")
	}
	recents := store.Models(true)
	if len(recents) > 5 {
		recents = recents[:5]
	}
	return &State{Options: options, store: store, favorites: store.Models(false), recents: recents, Defaults: store.SessionDefault(), Title: store.AutoTitle()}
}
func (s *State) Close() { s.closed = true; s.picker.open = false; s.selectMenu.kind = "" }
func (s *State) Tick()  {} // The host owns asynchronous catalog work.
func (s *State) label(key, fallback string, values ...string) string {
	out := fallback
	if s.Options.Text != nil {
		out = s.Options.Text(key, fallback)
	}
	for i := 0; i+1 < len(values); i += 2 {
		out = strings.ReplaceAll(out, "%{"+values[i]+"}", values[i+1])
	}
	return out
}
func cloneModel(m agent.ModelChoice) agent.ModelChoice { m.Raw = bytes.Clone(m.Raw); return m }
func (s *State) SetSnapshot(v Snapshot) {
	// Frequent unchanged get_state snapshots must not undo keyboard navigation.
	changed := !reflect.DeepEqual(s.Snapshot.Models, v.Models) || !reflect.DeepEqual(s.Snapshot.Current, v.Current) || !slices.Equal(s.Snapshot.Levels, v.Levels) || s.Snapshot.CurrentLevel != v.CurrentLevel
	v.Models = slices.Clone(v.Models)
	for i := range v.Models {
		v.Models[i] = cloneModel(v.Models[i])
	}
	v.Current = cloneModel(v.Current)
	v.Levels = slices.Clone(v.Levels)
	s.Snapshot = v
	if changed && s.picker.open {
		if s.picker.scope.Kind == "provider" && !slices.Contains(s.Providers(), s.picker.scope.Provider) || s.picker.scope == Favorites && len(s.favorites) == 0 || s.picker.scope == Recents && len(s.recents) == 0 {
			s.picker.scope = All
		}
		s.pinPicker()
	}
}
func (s *State) SetRefreshState(busy bool, detail string) {
	s.Refreshing = busy
	s.RefreshError = detail
	if !busy && detail == "" {
		s.Error = ""
	}
}
func (s *State) Refresh() bool {
	if s.closed || s.Refreshing {
		return false
	}
	if s.Snapshot.Running {
		s.fail(errors.New(s.label("settings.finish_turn_before_restart", "Finish the current turn before restarting pi.")))
		return false
	}
	if s.Options.OnRefresh == nil {
		s.fail(errors.New("Model catalog update is not connected."))
		return false
	}
	s.Refreshing = true
	s.RefreshError = ""
	if err := s.Options.OnRefresh(); err != nil {
		s.Refreshing = false
		s.RefreshError = err.Error()
		s.fail(err)
		return false
	}
	return true
}
func (s *State) fail(err error) {
	if err == nil {
		return
	}
	s.Error = err.Error()
	if s.Options.Notify != nil {
		s.Options.Notify(s.Error, true)
	}
}
func key(m agent.ModelChoice) preferences.Model {
	return preferences.Model{Provider: m.Provider, ID: m.ID}
}
func (s *State) IsFavorite(m agent.ModelChoice) bool { return slices.Contains(s.favorites, key(m)) }
func (s *State) Favorites() []preferences.Model      { return slices.Clone(s.favorites) }
func (s *State) Recents() []preferences.Model        { return slices.Clone(s.recents) }
func (s *State) ToggleFavorite(index int) bool {
	if s.closed || index < 0 || index >= len(s.Snapshot.Models) {
		return false
	}
	m := s.Snapshot.Models[index]
	if err := s.store.SetModel(false, key(m), !s.IsFavorite(m)); err != nil {
		s.fail(err)
		return false
	}
	s.favorites = s.store.Models(false)
	s.Error = ""
	if s.picker.open {
		s.pinPicker()
	}
	return true
}
func (s *State) recordRecent(m agent.ModelChoice) {
	if len(s.recents) > 0 && s.recents[0] == key(m) {
		return
	}
	if err := s.store.SetModel(true, key(m), true); err != nil {
		s.fail(err)
		return
	}
	s.recents = s.store.Models(true)
}
func (s *State) SelectModel(index int, fromPicker bool) bool {
	if s.closed || index < 0 || index >= len(s.Snapshot.Models) {
		return false
	}
	m := s.Snapshot.Models[index]
	if m.ID == "" || m.Provider == "" {
		s.fail(errors.New("Pi reported an incomplete model identity."))
		return false
	}
	if s.Options.OnSelectModel == nil {
		s.fail(errors.New("Model selection is not connected."))
		return false
	}
	s.Error = ""
	if fromPicker {
		s.recordRecent(m)
	}
	if err := s.Options.OnSelectModel(m.Provider, m.ID); err != nil {
		s.fail(err)
		return false
	}
	if fromPicker {
		s.DismissPicker(false)
	}
	return true
}
func (s *State) SelectThinking(level string) bool {
	if s.closed || !slices.Contains(s.Snapshot.Levels, level) {
		return false
	}
	if s.Options.OnSelectThinking == nil {
		s.fail(errors.New("Thinking selection is not connected."))
		return false
	}
	if err := s.Options.OnSelectThinking(level); err != nil {
		s.fail(err)
		return false
	}
	s.Error = ""
	s.DismissPicker(false)
	return true
}
func (s *State) Providers() []string {
	var out []string
	for _, m := range s.Snapshot.Models {
		if !slices.Contains(out, m.Provider) {
			out = append(out, m.Provider)
		}
	}
	return out
}
func matches(m agent.ModelChoice, q string) bool {
	return q == "" || strings.Contains(strings.ToLower(m.Name), q) || strings.Contains(strings.ToLower(m.ID), q) || strings.Contains(strings.ToLower(m.Provider), q)
}
func (s *State) Visible(index int) bool {
	if index < 0 || index >= len(s.Snapshot.Models) {
		return false
	}
	m := s.Snapshot.Models[index]
	return (s.ProviderFilter == "" || s.ProviderFilter == m.Provider) && (!s.FavoritesOnly || s.IsFavorite(m)) && matches(m, strings.ToLower(strings.TrimSpace(s.Query)))
}
func (s *State) Groups() []Group {
	var out []Group
	for _, provider := range s.Providers() {
		g := Group{Provider: provider}
		for i, m := range s.Snapshot.Models {
			if m.Provider == provider && s.Visible(i) {
				g.Models = append(g.Models, i)
			}
		}
		if len(g.Models) > 0 {
			out = append(out, g)
		}
	}
	return out
}
func (s *State) CountLabel() string {
	total, shown, favorites := len(s.Snapshot.Models), 0, 0
	if total == 0 {
		return s.label("settings.no_models_reported_by_runtime", "No models reported by the runtime")
	}
	for i, m := range s.Snapshot.Models {
		if s.Visible(i) {
			shown++
		}
		if s.IsFavorite(m) {
			favorites++
		}
	}
	fav := s.label("settings.n_favorites", "%{count} favorite(s)", "count", fmt.Sprint(favorites))
	if s.FavoritesOnly || strings.TrimSpace(s.Query) != "" || s.ProviderFilter != "" {
		return s.label("settings.count_of_shown", "%{shown} of %{total} shown · %{favorites}", "shown", fmt.Sprint(shown), "total", fmt.Sprint(total), "favorites", fav)
	}
	return s.label("settings.count_models", "%{total} models · %{favorites}", "total", fmt.Sprint(total), "favorites", fav)
}
func normalized(s string) string {
	var b strings.Builder
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// Rust eq_ignore_ascii_case folds only ASCII; Unicode lookalikes must not
// collapse distinct custom provider identities or unknown thinking levels.
func equalASCIIFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 'a' - 'A'
		}
		if y >= 'A' && y <= 'Z' {
			y += 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}
func IsPickerSelected(model, current agent.ModelChoice) bool {
	if current.Provider != "" && !equalASCIIFold(model.Provider, current.Provider) {
		return false
	}
	for _, v := range []string{current.ID, current.Name} {
		if v != "" && (normalized(v) == normalized(model.ID) || normalized(v) == normalized(model.Name)) {
			return true
		}
	}
	return false
}
func (s *State) IsActive(m agent.ModelChoice) bool {
	return m.ID == s.Snapshot.Current.ID && m.Provider == s.Snapshot.Current.Provider
}
func (s *State) PickerOpen(kind PickerKind) bool {
	if s.closed || s.Snapshot.Running || kind != ModelPicker && kind != ThinkingPicker || kind == ThinkingPicker && len(s.Snapshot.Levels) <= 1 {
		return false
	}
	s.selectMenu.kind = ""
	s.picker = pickerState{open: true, kind: kind, scope: All}
	s.pinPicker()
	return true
}
func (s *State) PickerWidth() float32 {
	if s.picker.kind == ThinkingPicker {
		return 300
	}
	return 360
}
func (s *State) PickerKind() PickerKind { return s.picker.kind }
func (s *State) PickerIsOpen() bool     { return s.picker.open }
func (s *State) DismissPicker(byMouse bool) {
	if !s.picker.open {
		return
	}
	s.picker.open = false
	if s.Options.OnDismiss != nil {
		s.Options.OnDismiss(byMouse)
	}
}
func (s *State) HasModal() bool { return s.picker.open || s.selectMenu.kind != "" }
func (s *State) Escape() bool {
	if s.selectMenu.kind != "" {
		s.selectMenu.kind = ""
		return true
	}
	if s.picker.open {
		s.DismissPicker(false)
		return true
	}
	return false
}
func (s *State) SetScope(scope Scope) {
	if scope == s.picker.scope {
		return
	}
	s.picker.scope = scope
	s.picker.scroll.Y = 0
	s.pinPicker()
}
func (s *State) PickerScope() Scope { return s.picker.scope }
func (s *State) SetPickerQuery(query string) {
	if query == s.picker.query {
		return
	}
	s.picker.query = query
	s.picker.lastQuery = query
	s.picker.scroll.Y = 0
	if query == "" {
		s.pinPicker()
	} else {
		s.picker.cursor = firstSelectable(s.PickerRows())
		s.picker.scrollPending = true
	}
}
func (s *State) PickerQuery() string { return s.picker.query }
func (s *State) Highlighted() int    { return s.picker.cursor }
func firstSelectable(rows []Row) int {
	for i, r := range rows {
		if r.Selectable() {
			return i
		}
	}
	return -1
}
func (s *State) pinPicker() {
	rows := s.PickerRows()
	s.picker.cursor = firstSelectable(rows)
	for i, r := range rows {
		if r.Selected {
			s.picker.cursor = i
			break
		}
	}
	s.picker.scrollPending = true
	s.picker.suppressHoverUntil = time.Now().Add(400 * time.Millisecond)
}
func (s *State) StepPicker(direction int) {
	rows := s.PickerRows()
	if len(rows) == 0 {
		s.picker.cursor = -1
		return
	}
	if direction == 0 {
		return
	}
	if s.picker.cursor < 0 || s.picker.cursor >= len(rows) || !rows[s.picker.cursor].Selectable() {
		s.picker.cursor = firstSelectable(rows)
		return
	}
	delta := 1
	if direction < 0 {
		delta = -1
	}
	for i := s.picker.cursor + delta; i >= 0 && i < len(rows); i += delta {
		if rows[i].Selectable() {
			s.picker.cursor = i
			break
		}
	}
	s.picker.scrollPending = true
}
func (s *State) ConfirmPicker() bool {
	if !s.picker.open {
		return false
	}
	rows := s.PickerRows()
	i := s.picker.cursor
	if i < 0 || i >= len(rows) || !rows[i].Selectable() {
		return false
	}
	if rows[i].Level != "" {
		return s.SelectThinking(rows[i].Level)
	}
	return s.SelectModel(rows[i].ModelIndex, true)
}
func (s *State) PickerRows() []Row {
	q := strings.ToLower(s.picker.query)
	var rows []Row
	if s.picker.kind == ThinkingPicker {
		for _, level := range s.Snapshot.Levels {
			if q == "" || strings.Contains(strings.ToLower(level), q) || strings.Contains(strings.ToLower(s.thinkingHint(level)), q) {
				rows = append(rows, Row{Level: level, Selected: equalASCIIFold(level, s.Snapshot.CurrentLevel), ModelIndex: -1})
			}
		}
		return rows
	}
	scope := s.picker.scope
	models := s.Snapshot.Models
	selected := -1
	for i, m := range models {
		if IsPickerSelected(m, s.Snapshot.Current) {
			selected = i
			break
		}
	}
	add := func(i int) { rows = append(rows, Row{ModelIndex: i, Selected: selected == i}) }
	inScope := func(m agent.ModelChoice) bool {
		switch scope.Kind {
		case "favorites":
			return s.IsFavorite(m)
		case "provider":
			return scope.Provider == m.Provider
		default:
			return true
		}
	}
	included := func(m agent.ModelChoice) bool { return inScope(m) && matches(m, q) }
	find := func(k preferences.Model) int {
		for i, m := range models {
			if key(m) == k {
				return i
			}
		}
		return -1
	}
	if scope == Recents {
		for _, r := range s.recents {
			i := find(r)
			if i >= 0 && matches(models[i], q) {
				add(i)
			}
		}
		return rows
	}
	pinned := map[int]bool{}
	if scope == All || scope.Kind == "provider" {
		var recent []int
		if scope == All {
			for _, r := range s.recents {
				i := find(r)
				if i >= 0 && included(models[i]) && !slices.Contains(recent, i) {
					recent = append(recent, i)
				}
			}
		} else {
			for i, m := range models {
				if slices.Contains(s.recents, key(m)) && included(m) {
					recent = append(recent, i)
				}
			}
		}
		if len(recent) > 0 {
			rows = append(rows, Row{Header: "recents", Count: len(recent), ModelIndex: -1})
			for _, i := range recent {
				add(i)
				pinned[i] = true
			}
		}
	}
	if scope == All {
		var favorites []int
		for i, m := range models {
			if s.IsFavorite(m) && included(m) && !pinned[i] {
				favorites = append(favorites, i)
			}
		}
		if len(favorites) > 0 {
			rows = append(rows, Row{Header: "favorites", Count: len(favorites), ModelIndex: -1})
			for _, i := range favorites {
				add(i)
				pinned[i] = true
			}
		}
	}
	for _, provider := range s.Providers() {
		var group []int
		for i, m := range models {
			if m.Provider == provider && included(m) && !pinned[i] {
				group = append(group, i)
			}
		}
		if len(group) == 0 {
			continue
		}
		if scope == All {
			rows = append(rows, Row{Header: "provider", Provider: provider, Count: len(group), ModelIndex: -1})
		}
		for _, i := range group {
			add(i)
		}
	}
	return rows
}

// CatalogThinkingLevels matches frozen Orbit/Pi's reasoning + thinkingLevelMap
// contract. Extended levels are opt-in; null explicitly removes a capability.
func CatalogThinkingLevels(model agent.ModelChoice) []string {
	return agent.CatalogThinkingLevels(model)
}

func (s *State) DefaultModelIndex() int {
	for i, m := range s.Snapshot.Models {
		if m.ID == s.Defaults.ModelID && (s.Defaults.Provider == "" || m.Provider == s.Defaults.Provider) {
			return i
		}
	}
	return -1
}
func (s *State) DefaultThinkingLevels() []string {
	i := s.DefaultModelIndex()
	if i < 0 {
		return nil
	}
	return CatalogThinkingLevels(s.Snapshot.Models[i])
}
func (s *State) ChooseDefaultModel(index int) bool {
	if s.closed || index < -1 || index >= len(s.Snapshot.Models) {
		return false
	}
	next := preferences.SessionDefault{}
	if index >= 0 {
		m := s.Snapshot.Models[index]
		next.Provider = m.Provider
		next.ModelID = m.ID
		if slices.Contains(CatalogThinkingLevels(m), s.Defaults.Thinking) {
			next.Thinking = s.Defaults.Thinking
		}
	}
	return s.saveDefault(next)
}
func (s *State) ChooseDefaultThinking(level string) bool {
	if s.closed || s.DefaultModelIndex() < 0 || level != "" && !slices.Contains(s.DefaultThinkingLevels(), level) {
		return false
	}
	next := s.Defaults
	next.Thinking = level
	return s.saveDefault(next)
}
func (s *State) saveDefault(next preferences.SessionDefault) bool {
	if err := s.store.SaveSessionDefault(next); err != nil {
		s.fail(err)
		return false
	}
	s.Defaults = next
	s.Error = ""
	if s.Options.OnDefaultsChanged != nil {
		s.Options.OnDefaultsChanged(next)
	}
	return true
}
func (s *State) SetTitlePreferences(title preferences.AutoTitle) {
	s.Title = title
	if title.Model != nil {
		model := *title.Model
		s.Title.Model = &model
	}
}
func (s *State) ChooseTitleModel(index int) bool {
	if s.closed || index < -1 || index >= len(s.Snapshot.Models) {
		return false
	}
	// Preserve an enabled toggle changed by another host control since last frame.
	next := s.store.AutoTitle()
	next.Model = nil
	if index >= 0 {
		m := key(s.Snapshot.Models[index])
		next.Model = &m
	}
	if err := s.store.SaveAutoTitle(next); err != nil {
		s.fail(err)
		return false
	}
	s.Title = next
	s.Error = ""
	if s.Options.OnTitleChanged != nil {
		s.Options.OnTitleChanged(next)
	}
	return true
}
func ProviderName(id string) string {
	for _, p := range workbench.BuiltinProviders() {
		if p.ID == id {
			return p.Name
		}
	}
	known := map[string]string{"zai-coding-cn": "Z.ai Coding (CN)", "moonshotai-cn": "Moonshot AI (CN)", "minimax-cn": "MiniMax (CN)", "google-vertex": "Google Vertex", "opencode": "opencode", "opencode-go": "opencode Go", "ant-ling": "Ant Ling", "baseten": "Baseten", "cerebras": "Cerebras", "fireworks": "Fireworks", "huggingface": "Hugging Face", "together": "Together", "xiaomi": "Xiaomi"}
	if name := known[id]; name != "" {
		return name
	}
	parts := strings.FieldsFunc(id, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		r := []rune(p)
		if len(r) > 0 {
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, " ")
}
func formatTokens(n int) string {
	scale, suffix := 1, ""
	if n >= 1000000 {
		scale, suffix = 1000000, "M"
	} else if n >= 1000 {
		scale, suffix = 1000, "K"
	}
	if scale == 1 {
		return fmt.Sprint(n)
	}
	whole, tenths := n/scale, ((n%scale)+scale/20)/(scale/10)
	if tenths == 10 {
		whole++
		tenths = 0
	}
	if tenths == 0 {
		return fmt.Sprintf("%d%s", whole, suffix)
	}
	return fmt.Sprintf("%d.%d%s", whole, tenths, suffix)
}
func displayLevel(s string) string {
	chars := []rune(s)
	if len(chars) == 0 {
		return ""
	}
	chars[0] = unicode.ToUpper(chars[0])
	return string(chars)
}
func (s *State) thinkingHint(level string) string {
	key, fallback := "reasoning_custom", "Custom reasoning level"
	switch strings.ToLower(level) {
	case "off", "none":
		key, fallback = "reasoning_none", "No reasoning"
	case "minimal", "low":
		key, fallback = "reasoning_fast", "Fast reasoning"
	case "medium":
		key, fallback = "reasoning_balanced", "Balanced reasoning"
	case "high":
		key, fallback = "reasoning_deeper", "Deeper reasoning"
	case "xhigh", "ultra":
		key, fallback = "reasoning_extensive", "Extensive reasoning"
	case "max":
		key, fallback = "reasoning_maximum", "Maximum reasoning"
	case "auto":
		key, fallback = "reasoning_auto", "Model chooses depth"
	}
	return s.label("model_selector."+key, fallback)
}
