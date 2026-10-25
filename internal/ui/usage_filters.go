package orbitui

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type usageFilterView struct {
	Open        bool
	FocusSearch bool
	Search      string
	Selected    int
	List        ui.ListState
}
type usageFilterOption struct {
	Key, Label, Sub string
	Model           workbench.UsageModelRef
}
type usageCalendarRestore struct {
	Start, End time.Time
	Preset     string
}

func toggleUsageString(values *[]string, value string) {
	if slices.Contains(*values, value) {
		*values = slices.DeleteFunc(*values, func(v string) bool { return v == value })
	} else {
		*values = append(*values, value)
	}
}
func toggleUsageModel(values *[]workbench.UsageModelRef, value workbench.UsageModelRef) {
	if slices.Contains(*values, value) {
		*values = slices.DeleteFunc(*values, func(v workbench.UsageModelRef) bool { return v == value })
	} else {
		*values = append(*values, value)
	}
}
func (s *usageState) filterOptions(kind string, search string) []usageFilterOption {
	var all []usageFilterOption
	counts := map[string]int{}
	for _, r := range s.index.Records {
		key := r.Provider
		if kind == "models" {
			key = r.Provider + "\x00" + r.Model
		}
		if kind == "workspaces" {
			key = r.Workspace
		}
		counts[key]++
	}
	switch kind {
	case "models":
		for _, ref := range s.projection.ModelRefOptions {
			all = append(all, usageFilterOption{Key: ref.Provider + "\x00" + ref.Model, Label: ref.Model, Sub: ref.Provider, Model: ref})
		}
	case "providers":
		for _, v := range s.projection.ProviderOptions {
			all = append(all, usageFilterOption{Key: v, Label: v, Sub: usageCount(uint64(counts[v]))})
		}
	case "workspaces":
		for _, v := range s.projection.WorkspaceOptions {
			all = append(all, usageFilterOption{Key: v, Label: filepath.Base(v), Sub: v})
		}
	}
	slices.SortStableFunc(all, func(a, b usageFilterOption) int {
		if counts[a.Key] != counts[b.Key] {
			return counts[b.Key] - counts[a.Key]
		}
		return strings.Compare(a.Label, b.Label)
	})
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle != "" {
		all = slices.DeleteFunc(all, func(v usageFilterOption) bool {
			return !strings.Contains(strings.ToLower(v.Label), needle) && !strings.Contains(strings.ToLower(v.Sub), needle)
		})
	}
	slices.SortStableFunc(all, func(a, b usageFilterOption) int {
		aa, bb := s.filterChecked(kind, a), s.filterChecked(kind, b)
		if aa == bb {
			return 0
		}
		if aa {
			return -1
		}
		return 1
	})
	return all
}
func (s *usageState) filterChecked(kind string, o usageFilterOption) bool {
	switch kind {
	case "models":
		return slices.Contains(s.filter.ModelRefs, o.Model) || (s.filter.Model != "" && s.filter.Model == o.Model.Model) || slices.Contains(s.filter.Models, o.Model.Model)
	case "providers":
		return s.filter.Provider == o.Key || slices.Contains(s.filter.Providers, o.Key)
	case "workspaces":
		return s.filter.Workspace == o.Key || slices.Contains(s.filter.Workspaces, o.Key)
	}
	return false
}
func (s *usageState) clearDimension(kind string) {
	switch kind {
	case "models":
		s.filter.Model = ""
		s.filter.Models = nil
		s.filter.ModelRefs = nil
	case "providers":
		s.filter.Provider = ""
		s.filter.Providers = nil
	case "workspaces":
		s.filter.Workspace = ""
		s.filter.Workspaces = nil
	}
}
func (s *usageState) toggleFilter(kind string, o usageFilterOption) {
	switch kind {
	case "models": // Normalize legacy raw IDs once, preserving their original union.
		if s.filter.Model != "" || len(s.filter.Models) > 0 {
			for _, v := range s.filterOptions(kind, "") {
				if s.filter.Model == v.Model.Model || slices.Contains(s.filter.Models, v.Model.Model) {
					if !slices.Contains(s.filter.ModelRefs, v.Model) {
						s.filter.ModelRefs = append(s.filter.ModelRefs, v.Model)
					}
				}
			}
			s.filter.Model = ""
			s.filter.Models = nil
		}
		toggleUsageModel(&s.filter.ModelRefs, o.Model)
	case "providers":
		if s.filter.Provider != "" {
			if !slices.Contains(s.filter.Providers, s.filter.Provider) {
				s.filter.Providers = append(s.filter.Providers, s.filter.Provider)
			}
			s.filter.Provider = ""
		}
		toggleUsageString(&s.filter.Providers, o.Key)
	case "workspaces":
		if s.filter.Workspace != "" {
			if !slices.Contains(s.filter.Workspaces, s.filter.Workspace) {
				s.filter.Workspaces = append(s.filter.Workspaces, s.filter.Workspace)
			}
			s.filter.Workspace = ""
		}
		toggleUsageString(&s.filter.Workspaces, o.Key)
	}
}
func (s *usageState) selectFilterOptions(kind string, options []usageFilterOption) {
	s.clearDimension(kind)
	for _, o := range options {
		s.toggleFilter(kind, o)
	}
}

// The source picker confirms by dismissing, never by toggling a highlighted
// option. Selection is a row click; Select all ignores the search query.
func (a *App) usageDimensionFilter(c *ui.Context, kind string) {
	s := &a.usageState
	if s.filterViews == nil {
		s.filterViews = map[string]*usageFilterView{}
	}
	v := s.filterViews[kind]
	if v == nil {
		v = &usageFilterView{}
		s.filterViews[kind] = v
	}
	all := s.filterOptions(kind, "")
	var selected []usageFilterOption
	for _, o := range all {
		if s.filterChecked(kind, o) {
			selected = append(selected, o)
		}
	}
	label := a.text("usage.all_"+kind, "All "+kind)
	if len(selected) == 1 {
		label = selected[0].Label
	} else if len(selected) > 1 {
		label = a.usageText("usage.n_selected", "%{count} selected", map[string]string{"count": usageCount(uint64(len(selected)))})
	}
	anchor := a.button(c, label, "icons/filter.svg", len(selected) > 0).MaxWidth(240).AlignSelf(ui.Start).Border(1, a.Theme.Border)
	anchor.Children(func() { a.icon(c, "icons/chevron-down.svg", 12, a.Theme.TextMuted) })
	if anchor.Clicked() {
		s.closeUsagePopovers(kind)
		v.Open = !v.Open
		for other, state := range s.filterViews {
			if other != kind {
				state.Open = false
			}
		}
		s.rangeOpen = false
		s.customOpen = false
		if v.Open {
			s.filterMenu = kind
			v.Search = ""
			v.FocusSearch = true
			v.Selected = 0
		} else {
			s.filterMenu = ""
		}
	}
	wasOpen := v.Open
	a.usageNativePopover(c.Key("usage-filter:"+kind), anchor, &v.Open, 268, func() {
		root := ui.Column(c).AlignSelf(ui.Stretch).Gap(0)
		confirm := root.Shortcut(0, ui.KeyEnter)
		confirm = root.OverlayShortcut(0, ui.KeyEnter) || confirm
		if confirm {
			v.Open = false
			s.filterMenu = ""
		}
		root.Children(func() {
			search := ui.TextInput(c, &v.Search).Label("Filter " + kind).Placeholder(a.text("usage.search", "Search…")).AlignSelf(ui.Stretch)
			if v.FocusSearch {
				search.Focus()
				v.FocusSearch = false
			}
			if search.Submitted() {
				v.Open = false
				s.filterMenu = ""
			}
			options := s.filterOptions(kind, v.Search)
			ui.Scroll(c.Key("usage-filter-options:" + kind)).Height(float32(min(8, len(options)+1)) * 32).Children(func() {
				ui.Column(c).WidthPercent(100).Children(func() {
					if a.usagePickerRow(c.Key("all"), a.text("usage.all", "All"), "", len(selected) == 0).Clicked() {
						s.clearDimension(kind)
					}
					for _, o := range options {
						if a.usagePickerRow(c.Key(o.Key), o.Label, o.Sub, s.filterChecked(kind, o)).Clicked() {
							s.toggleFilter(kind, o)
						}
					}
				})
			})
			if len(options) == 0 && strings.TrimSpace(v.Search) != "" {
				ui.Text(c, a.usageText("usage.no_query_match", "No match for “%{query}”", map[string]string{"query": strings.TrimSpace(v.Search)})).FontSize(12).TextColor(a.Theme.TextMuted)
			}
			a.line(c)
			ui.Row(c).Gap(8).Children(func() {
				if a.button(c, a.text("usage.select_all", "Select all"), "", false).Clicked() {
					s.selectFilterOptions(kind, all)
				}
				ui.Spacer(c)
				if a.button(c, a.text("usage.clear", "Clear"), "", false).Clicked() {
					s.clearDimension(kind)
				}
			})
		})
	})
	if wasOpen && !v.Open {
		s.filterMenu = ""
	}
}
func (a *App) usagePickerRow(c *ui.Context, label, sub string, selected bool) ui.Element {
	identity := label
	if sub != "" {
		identity += " · " + sub
	}
	b := ui.ButtonBase(c).Height(a.Theme.Px(23)+2*a.Theme.Space(4)).AlignSelf(ui.Stretch).Padding(a.Theme.Space(4), a.Theme.Space(6)).Margin(0, a.Theme.Space(4)).Radius(a.Theme.Px(6)).Clip().Gap(a.Theme.Px(6)).Label(identity).Justify(ui.Start).AlignItems(ui.Center)
	if selected {
		b.Background(a.Theme.Selection)
	} else if b.Hovered() {
		b.Background(a.Theme.Hover)
	}
	b.Children(func() {
		if selected {
			a.icon(c, "icons/check.svg", 14, a.Theme.Accent)
		} else {
			ui.Box(c).Size(14, 14)
		}
		text := ui.Text(c, label).FontSize(a.Theme.Px(14)).Grow(1).MinWidth(0).SingleLine().TextColor(a.Theme.TextSecondary)
		if selected {
			text.FontWeight(500).TextColor(a.Theme.Text)
		}
		if sub != "" {
			ui.Text(c, sub).FontSize(a.Theme.Px(12)).MaxWidth(120).Margin(0, 0, 0, a.Theme.Px(10)).TextColor(a.Theme.TextMuted).SingleLine()
		}
	})
	return b
}
func (a *App) usageText(key, fallback string, vars map[string]string) string {
	text := a.text(key, fallback)
	for k, v := range vars {
		text = strings.ReplaceAll(text, "%{"+k+"}", v)
	}
	return text
}
func (a *App) scopeUsageDay(at time.Time) {
	s := &a.usageState
	end := at.AddDate(0, 0, 1)
	if s.filter.Start.Equal(at) && s.filter.End.Equal(end) && s.preset == "custom" {
		if s.calendarRestore == nil {
			return
		}
		r := s.calendarRestore
		s.filter.Start, s.filter.End, s.preset = r.Start, r.End, r.Preset
		s.calendarRestore = nil
	} else {
		s.calendarRestore = &usageCalendarRestore{s.filter.Start, s.filter.End, s.preset}
		s.filter.Start, s.filter.End, s.preset = at, end, "custom"
	}
	s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
	s.granularity = usageAutoGranularity(s.filter, s.index)
	a.saveUsagePrefs()
}
func (a *App) usageActiveFilters(c *ui.Context) {
	s := &a.usageState
	f := &s.filter
	active := f.Workspace != "" || f.Provider != "" || f.Model != "" || len(f.Workspaces)+len(f.Providers)+len(f.Models)+len(f.ModelRefs) > 0 || f.Session != "" || f.SessionPath != "" || !f.FocusStart.IsZero() || f.ErrorsOnly || f.CachedOnly
	if !active {
		return
	}
	ui.Row(c).Wrap().Gap(6).Children(func() {
		for _, kind := range []string{"workspaces", "providers", "models"} {
			for _, o := range s.filterOptions(kind, "") {
				if s.filterChecked(kind, o) {
					if a.button(c, o.Label+" ×", "", true).Clicked() {
						s.toggleFilter(kind, o)
					}
				}
			}
		}
		if f.Session != "" || f.SessionPath != "" {
			if a.button(c, "Session scope ×", "", true).Clicked() {
				f.Session, f.SessionPath = "", ""
			}
		}
		if !f.FocusStart.IsZero() {
			if a.button(c, "Bucket "+f.FocusStart.Local().Format("Jan 2 15:04")+" ×", "", true).Clicked() {
				f.FocusStart, f.FocusEnd = time.Time{}, time.Time{}
			}
		}
		if f.ErrorsOnly {
			if a.button(c, "Failed requests ×", "", true).Clicked() {
				f.ErrorsOnly = false
			}
		}
		if f.CachedOnly {
			if a.button(c, "Cached requests ×", "", true).Clicked() {
				f.CachedOnly = false
			}
		}
		if a.button(c, "Clear filters", "icons/x.svg", false).Clicked() {
			start, end := f.Start, f.End
			*f = workbench.UsageFilter{Start: start, End: end}
			s.calendarRestore = nil
		}
	})
}

func (s *usageState) now() time.Time {
	if !s.fixtureNow.IsZero() {
		return s.fixtureNow
	}
	return time.Now()
}
func (a *App) usagePresetLabel(p string) string {
	key := strings.ReplaceAll(p, "-", "_")
	if p == "all" {
		key = "all_time"
	}
	return a.text("usage.range_"+key, usageRangeLabel(p))
}
func usageCustomLabel(start, end time.Time) string {
	if start.Year() == end.Year() && start.YearDay() == end.YearDay() {
		return start.Format("Jan 2, 2006")
	}
	if start.Year() == end.Year() {
		return start.Format("Jan 2") + " – " + end.Format("Jan 2, 2006")
	}
	return start.Format("Jan 2, 2006") + " – " + end.Format("Jan 2, 2006")
}
func (a *App) openUsageCalendar() {
	s := &a.usageState
	start, end := s.filter.Start, s.filter.End.Add(-time.Nanosecond)
	if s.preset == "all" && len(s.index.Records) > 0 {
		start, end = s.index.Records[0].At, s.index.Records[0].At
		for _, r := range s.index.Records {
			if r.At.Before(start) {
				start = r.At
			}
			if r.At.After(end) {
				end = r.At
			}
		}
	}
	if start.IsZero() {
		start = s.now()
	}
	if end.Before(start) {
		end = start
	}
	start, end = bucketStart(start, "day", time.Local), bucketStart(end, "day", time.Local)
	s.customStart, s.customEnd = start.Format("2006-01-02"), end.Format("2006-01-02")
	s.customMonth = bucketStart(end, "month", time.Local)
	s.customPicking = false
	s.customOpen = true
}
func (a *App) usageRangeControl(c *ui.Context) {
	s := &a.usageState
	label := a.usagePresetLabel(s.preset)
	if s.preset == "custom" {
		label = usageCustomLabel(s.filter.Start, s.filter.End.Add(-time.Nanosecond))
	}
	anchor := a.button(c, label, "icons/clock.svg", s.preset != "last7").AlignSelf(ui.Start).Border(1, a.Theme.Border)
	anchor.Children(func() { a.icon(c, "icons/chevron-down.svg", 12, a.Theme.TextMuted) })
	if anchor.Clicked() {
		s.closeUsagePopovers("range")
		s.rangeOpen = !s.rangeOpen
		s.customOpen = false
		for _, v := range s.filterViews {
			v.Open = false
		}
		if s.rangeOpen {
			s.filterMenu = "range"
		} else {
			s.filterMenu = ""
		}
	}
	wasOpen := s.rangeOpen
	a.usageNativePopover(c.Key("usage-range-menu"), anchor, &s.rangeOpen, 244, func() {
		root := ui.Column(c).AlignSelf(ui.Stretch).Gap(0)
		confirm := root.Shortcut(0, ui.KeyEnter)
		confirm = root.OverlayShortcut(0, ui.KeyEnter) || confirm
		if confirm {
			s.rangeOpen = false
			s.customOpen = false
			s.filterMenu = ""
		}
		root.Children(func() {
			for _, preset := range []string{"today", "yesterday", "last7", "last14", "last30", "this-week", "this-month", "previous-month", "custom", "all"} {
				hintKey := map[string]string{"today": "compare_with_yesterday", "last7": "vs_previous_7_days", "last14": "vs_previous_14_days", "last30": "vs_previous_30_days", "previous-month": "complete_month", "all": "no_comparison"}[preset]
				hint := ""
				if hintKey != "" {
					hint = a.text("usage."+hintKey, "")
				}
				if a.usagePickerRow(c.Key("range-"+preset), a.usagePresetLabel(preset), hint, s.preset == preset).Clicked() {
					if preset == "custom" {
						a.openUsageCalendar()
					} else {
						s.preset = preset
						s.filter.Start, s.filter.End = workbench.UsageRange(preset, s.now())
						s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
						s.calendarRestore = nil
						s.granularity = usageAutoGranularity(s.filter, s.index)
						s.rangeOpen = false
						s.customOpen = false
						s.filterMenu = ""
						a.saveUsagePrefs()
					}
				}
			}
			if s.customOpen {
				a.line(c)
				a.usageCalendarPicker(c)
			}
		})
	})
	if wasOpen && !s.rangeOpen {
		s.customOpen = false
		s.filterMenu = ""
	}
}
func (a *App) usageCalendarPicker(c *ui.Context) {
	s := &a.usageState
	if s.customMonth.IsZero() {
		s.customMonth = bucketStart(s.now(), "month", time.Local)
	}
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		prev := ui.ButtonBase(c).Size(26, 26).Label("Calendar previous month").Tooltip(a.text("usage.previous_month", "Previous month"))
		prev.Children(func() { a.icon(c, "icons/chevron-right.svg", 14, a.Theme.TextMuted).Rotate(180) })
		if prev.Clicked() {
			s.customMonth = s.customMonth.AddDate(0, -1, 0)
		}
		ui.Text(c, s.customMonth.Format("January 2006")).FontSize(12).Grow(1).TextAlign(ui.Center)
		if a.iconButton(c, a.text("usage.next_month", "Next month"), "icons/chevron-right.svg").Label("Calendar next month").Clicked() {
			s.customMonth = s.customMonth.AddDate(0, 1, 0)
		}
	})
	ui.Row(c).Children(func() {
		for _, name := range []string{"M", "T", "W", "T", "F", "S", "S"} {
			ui.Text(c, name).FontSize(11).Grow(1).TextAlign(ui.Center).TextColor(a.Theme.TextMuted)
		}
	})
	offset := (int(s.customMonth.Weekday()) + 6) % 7
	days := s.customMonth.AddDate(0, 1, -1).Day()
	grid := s.customMonth.AddDate(0, 0, -offset)
	today := bucketStart(s.now(), "day", time.Local)
	for week := 0; week < (offset+days+6)/7; week++ {
		ui.Row(c).Gap(2).Children(func() {
			for day := 0; day < 7; day++ {
				at := grid.AddDate(0, 0, week*7+day)
				if at.Month() != s.customMonth.Month() {
					ui.Box(c).Height(26).Grow(1)
					continue
				}
				label := at.Format("2006-01-02")
				selected := label == s.customStart || label == s.customEnd
				b := a.button(c, fmt.Sprint(at.Day()), "", selected).Label("Choose date " + label).Height(26).Padding(2).Justify(ui.Center).Grow(1)
				if !selected && s.customEnd != "" && label > s.customStart && label < s.customEnd {
					b.Background(a.Theme.Hover)
				}
				if !selected && at.Equal(today) {
					b.TextColor(a.Theme.Accent)
				}
				if b.Clicked() {
					if !s.customPicking {
						s.customStart = label
						s.customEnd = ""
						s.customPicking = true
					} else {
						start, _ := time.ParseInLocation("2006-01-02", s.customStart, time.Local)
						end := at
						if end.Before(start) {
							start, end = end, start
						}
						s.customStart, s.customEnd = start.Format("2006-01-02"), end.Format("2006-01-02")
						s.filter.Start, s.filter.End = start, end.AddDate(0, 0, 1)
						s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
						s.preset = "custom"
						s.calendarRestore = nil
						s.granularity = usageAutoGranularity(s.filter, s.index)
						s.customPicking = false
						s.customOpen = false
						s.rangeOpen = false
						s.filterMenu = ""
						a.saveUsagePrefs()
					}
				}
			}
		})
	}
	hint := a.text("usage.pick_start_date", "Pick the start date")
	if s.customStart != "" {
		hint = a.text("usage.pick_end_date", "Pick the end date")
	}
	if s.customStart != "" && s.customEnd != "" {
		from, _ := time.ParseInLocation("2006-01-02", s.customStart, time.Local)
		to, _ := time.ParseInLocation("2006-01-02", s.customEnd, time.Local)
		hint = usageCustomLabel(from, to)
	}
	ui.Text(c, hint).FontSize(11).TextColor(a.Theme.TextMuted).Padding(4, 8)
}

func pruneUsageFilter(index workbench.UsageIndex, f *workbench.UsageFilter) bool {
	models := map[workbench.UsageModelRef]bool{}
	providers, workspaces, sessions, paths := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	add := func(session, path, workspace, provider, model string) {
		sessions[session] = true
		paths[path] = true
		workspaces[workspace] = true
		providers[provider] = true
		models[workbench.UsageModelRef{Provider: provider, Model: model}] = true
	}
	for _, r := range index.Records {
		add(r.Session, r.SessionPath, r.Workspace, r.Provider, r.Model)
	}
	for _, r := range index.Tools {
		add(r.Session, r.SessionPath, r.Workspace, r.Provider, r.Model)
	}
	for _, r := range index.Errors {
		add(r.Session, r.SessionPath, r.Workspace, r.Provider, r.Model)
	}
	for _, r := range index.Sessions {
		sessions[r.ID] = true
		paths[r.Path] = true
		workspaces[r.Workspace] = true
	}
	cleared := false
	if f.SessionPath != "" && !paths[f.SessionPath] {
		f.SessionPath = ""
		cleared = true
	}
	if f.Session != "" && !sessions[f.Session] {
		f.Session = ""
		cleared = true
	}
	f.ModelRefs = slices.DeleteFunc(f.ModelRefs, func(ref workbench.UsageModelRef) bool { return !models[ref] })
	f.Providers = slices.DeleteFunc(f.Providers, func(v string) bool { return !providers[v] })
	f.Workspaces = slices.DeleteFunc(f.Workspaces, func(v string) bool { return !workspaces[v] })
	if f.Provider != "" && !providers[f.Provider] {
		f.Provider = ""
	}
	if f.Workspace != "" && !workspaces[f.Workspace] {
		f.Workspace = ""
	}
	hasModel := func(v string) bool {
		for ref := range models {
			if ref.Model == v {
				return true
			}
		}
		return false
	}
	f.Models = slices.DeleteFunc(f.Models, func(v string) bool { return !hasModel(v) })
	if f.Model != "" && !hasModel(f.Model) {
		f.Model = ""
	}
	return cleared
}
