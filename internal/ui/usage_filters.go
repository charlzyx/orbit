package orbitui

import (
	"fmt"
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
	Key, Label string
	Model      workbench.UsageModelRef
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
	switch kind {
	case "models":
		for _, ref := range s.projection.ModelRefOptions {
			all = append(all, usageFilterOption{Key: ref.Provider + "\x00" + ref.Model, Label: ref.Model + " · " + ref.Provider, Model: ref})
		}
	case "providers":
		for _, v := range s.projection.ProviderOptions {
			all = append(all, usageFilterOption{Key: v, Label: v})
		}
	case "workspaces":
		for _, v := range s.projection.WorkspaceOptions {
			all = append(all, usageFilterOption{Key: v, Label: v})
		}
	}
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle == "" {
		return all
	}
	return slices.DeleteFunc(all, func(v usageFilterOption) bool { return !strings.Contains(strings.ToLower(v.Label), needle) })
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
func (a *App) usageDimensionFilter(c *ui.Context, kind string) {
	s := &a.usageState
	if s.filterViews == nil {
		s.filterViews = map[string]*usageFilterView{}
	}
	v := s.filterViews[kind]
	if v == nil {
		v = &usageFilterView{Selected: 0}
		s.filterViews[kind] = v
	}
	all := s.filterOptions(kind, "")
	var selected []usageFilterOption
	for _, o := range all {
		if s.filterChecked(kind, o) {
			selected = append(selected, o)
		}
	}
	label := "All " + kind
	if len(selected) == 1 {
		label = selected[0].Label
	} else if len(selected) > 1 {
		label = fmt.Sprintf("%d %s", len(selected), kind)
	}
	anchor := a.button(c, label, "icons/chevron-down.svg", v.Open).MaxWidth(200)
	if anchor.Clicked() {
		v.Open = !v.Open
		for other, state := range s.filterViews {
			if other != kind {
				state.Open = false
			}
		}
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
	ui.Popover(c.Key("usage-filter:"+kind), anchor, &v.Open, func() {
		root := ui.Column(c).Width(320).MaxHeight(360).Gap(8)
		options := s.filterOptions(kind, v.Search)
		// Capture menu navigation before the focused search field can treat Enter
		// as an input submission. Space stays editable in the search field.
		switch {
		case root.OverlayShortcut(0, ui.KeyDown):
			v.Selected = min(len(options)-1, v.Selected+1)
		case root.OverlayShortcut(0, ui.KeyUp):
			v.Selected = max(0, v.Selected-1)
		case root.OverlayShortcut(0, ui.KeyHome):
			v.Selected = 0
		case root.OverlayShortcut(0, ui.KeyEnd):
			v.Selected = len(options) - 1
		case root.OverlayShortcut(0, ui.KeyEnter):
			if v.Selected >= 0 && v.Selected < len(options) {
				s.toggleFilter(kind, options[v.Selected])
			}
		}
		root.Children(func() {
			search := ui.TextInput(c, &v.Search).Label("Filter " + kind).Placeholder("Search " + kind + "…").AlignSelf(ui.Stretch)
			if v.FocusSearch {
				search.Focus()
				v.FocusSearch = false
			}
			if search.Submitted() && v.Selected >= 0 && v.Selected < len(options) {
				s.toggleFilter(kind, options[v.Selected])
			}
			if search.Changed() {
				v.Selected = 0
			}
			options = s.filterOptions(kind, v.Search)
			ui.Row(c).Gap(8).Children(func() {
				if a.button(c, "Select all", "", false).Disabled(len(options) == 0).Clicked() {
					s.selectFilterOptions(kind, options)
				}
				if a.button(c, "Clear", "", false).Clicked() {
					s.clearDimension(kind)
				}
			})
			v.Selected = max(0, min(len(options)-1, v.Selected))
			v.List.Selected = &v.Selected
			v.List.Key = func(i int) any { return options[i].Key }
			v.List.Label = func(i int) string { return options[i].Label }
			list := ui.List(c.Key("usage-filter-options:"+kind), &v.List, len(options), func(i int) {
				o := options[i]
				checked := s.filterChecked(kind, o)
				if ui.Checkbox(c, &checked, o.Label).AlignSelf(ui.Stretch).Changed() {
					s.toggleFilter(kind, o)
				}
			}).Height(float32(max(1, min(7, len(options))) * 32)).Label("Choose " + kind)
			if (list.Submitted() || list.Shortcut(0, ui.KeySpace)) && v.Selected >= 0 && v.Selected < len(options) {
				s.toggleFilter(kind, options[v.Selected])
			}
			if len(options) == 0 {
				ui.Text(c, "No matches").TextColor(a.Theme.TextMuted)
			}
		})
	})
	if wasOpen && !v.Open {
		s.filterMenu = ""
	}
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

func (a *App) usageCalendarPicker(c *ui.Context) {
	s := &a.usageState
	if s.customMonth.IsZero() {
		s.customMonth = bucketStart(time.Now(), "month", time.Local)
	}
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		if a.iconButton(c, "Previous month", "icons/arrow-left.svg").Clicked() {
			s.customMonth = s.customMonth.AddDate(0, -1, 0)
		}
		ui.Text(c, s.customMonth.Format("January 2006")).Grow(1)
		if a.iconButton(c, "Next month", "icons/arrow-right.svg").Clicked() {
			s.customMonth = s.customMonth.AddDate(0, 1, 0)
		}
	})
	ui.Row(c).Children(func() {
		for _, name := range []string{"M", "T", "W", "T", "F", "S", "S"} {
			ui.Text(c, name).Grow(1).TextAlign(ui.Center).TextColor(a.Theme.TextMuted)
		}
	})
	grid := s.customMonth.AddDate(0, 0, -(int(s.customMonth.Weekday())+6)%7)
	for week := 0; week < 6; week++ {
		ui.Row(c).Gap(2).Children(func() {
			for day := 0; day < 7; day++ {
				at := grid.AddDate(0, 0, week*7+day)
				label := at.Format("2006-01-02")
				selected := label == s.customStart || label == s.customEnd
				b := a.button(c, fmt.Sprint(at.Day()), "", selected).Label("Choose date " + label).Height(24).Grow(1).Disabled(at.Month() != s.customMonth.Month())
				if b.Clicked() {
					if !s.customPicking {
						s.customStart = label
						s.customEnd = ""
						s.customPicking = true
					} else {
						s.customEnd = label
						s.customPicking = false
						start, _ := time.ParseInLocation("2006-01-02", s.customStart, time.Local)
						end := at
						if end.Before(start) {
							start, end = end, start
						}
						s.filter.Start, s.filter.End = start, end.AddDate(0, 0, 1)
						s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
						s.preset = "custom"
						s.calendarRestore = nil
						s.granularity = usageAutoGranularity(s.filter, s.index)
						s.customOpen = false
						a.saveUsagePrefs()
					}
				}
			}
		})
	}
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
