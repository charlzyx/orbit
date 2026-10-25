package orbitui

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type usageTableState struct {
	Search               string
	sortPhase            int
	Sort, observedSort   ui.SortOrder
	Page, Size, Selected int
	Hidden               []string
	List                 ui.ListState
}
type usageTableRow struct {
	usageAggregate
	Failure *workbench.UsageError
	Point   *usagePoint
}
type usageTableResult struct {
	Rows               []usageTableRow
	Total, Page, Pages int
	Totals, PageTotals usageAggregate
}

func usageDefaultSort(kind string) ui.SortOrder {
	switch kind {
	case "series":
		return ui.SortOrder{Column: "time"}
	case "daily":
		return ui.SortOrder{Column: "date", Descending: true}
	case "failures":
		return ui.SortOrder{Column: "when", Descending: true}
	case "breakdown:tools":
		return ui.SortOrder{Column: "calls", Descending: true}
	default:
		return ui.SortOrder{Column: "tokens", Descending: true}
	}
}
func (s *usageState) table(kind string) *usageTableState {
	if s.tables == nil {
		s.tables = map[string]*usageTableState{}
	}
	if t := s.tables[kind]; t != nil {
		return t
	}
	t := &usageTableState{Page: 1, Size: 25, Selected: -1, Sort: usageDefaultSort(kind)}
	if strings.HasPrefix(kind, "breakdown:") || kind == "series" {
		t.Size = 10
	}
	if kind == "sessions" {
		t.Sort = ui.SortOrder{Column: s.sort, Descending: s.descending}
		t.Size = s.pageSize
		t.Hidden = append([]string(nil), s.hiddenColumns...)
	}
	if strings.HasPrefix(kind, "breakdown:") && kind != "breakdown:tools" {
		t.Sort = ui.SortOrder{Column: s.breakdownSort, Descending: s.breakdownDescending}
	}
	t.observedSort = t.Sort
	s.tables[kind] = t
	return t
}
func (t *usageTableState) setPageSize(n int) {
	if slices.Contains([]int{10, 25, 50, 100}, n) && t.Size != n {
		t.Size = n
		t.Page = 1
		t.Selected = -1
	}
}
func (t *usageTableState) toggleColumn(id, required string) {
	if id == required {
		return
	}
	if slices.Contains(t.Hidden, id) {
		t.Hidden = slices.DeleteFunc(t.Hidden, func(s string) bool { return s == id })
	} else {
		t.Hidden = append(t.Hidden, id)
	}
}
func usageRowText(r usageTableRow, kind string) string {
	if r.Failure != nil {
		e := r.Failure
		return e.At.Local().Format("Jan 2, 15:04") + " " + e.Kind + " " + e.Model + " " + e.SessionTitle + " " + e.Message
	}
	if r.Point != nil {
		return r.Point.At.Local().Format("Jan 2 2006 January Monday 15:04") + " " + r.Name
	}
	status := "ok"
	if r.Errors+r.ToolFailures > 0 {
		status = "error failed"
	}
	return r.Name + " " + r.ID + " " + r.Workspace + " " + r.Provider + " " + r.Model + " " + status
}
func usageRowField(r usageTableRow, id, metric, latency string) (string, float64, bool) {
	if r.Failure != nil {
		e := r.Failure
		switch id {
		case "when":
			return "", float64(e.At.UnixMilli()), false
		case "kind":
			return e.Kind, 0, true
		case "model":
			return e.Model, 0, true
		case "session":
			return e.SessionTitle, 0, true
		case "message":
			return e.Message, 0, true
		}
	}
	switch id {
	case "title", "name", "tool":
		return r.Name, 0, true
	case "workspace":
		return r.Workspace, 0, true
	case "provider":
		return r.Provider, 0, true
	case "model":
		return r.Model, 0, true
	case "started":
		return "", float64(r.Started.UnixMilli()), false
	case "date", "time", "latest":
		return "", float64(r.At.UnixMilli()), false
	case "requests", "calls":
		return "", float64(r.Requests), false
	case "input":
		return "", float64(r.Tokens.Input), false
	case "output":
		return "", float64(r.Tokens.Output), false
	case "cache":
		return "", float64(r.Tokens.CacheRead + r.Tokens.CacheWrite), false
	case "cache-hit":
		rate, ok := usageCacheRate(r.Tokens)
		if !ok {
			return "", -1, false
		}
		return "", rate, false
	case "cost":
		if r.Priced == 0 {
			return "", -1, false
		}
		return "", r.Cost, false
	case "errors":
		return "", float64(r.Errors + r.ToolFailures), false
	case "tools":
		return "", float64(r.ToolRuns), false
	case "duration":
		return "", max(0, r.Ended.Sub(r.Started).Seconds()), false
	case "avg":
		if r.Timed == 0 {
			return "", -1, false
		}
		return "", r.Duration.Seconds() / float64(r.Timed), false
	case "max":
		if r.Timed == 0 {
			return "", -1, false
		}
		return "", r.MaxDuration.Seconds(), false
	case "value":
		if metric == "latency" && r.Point != nil {
			v, ok := r.Point.Latency.value(latency)
			if !ok {
				return "", -1, false
			}
			return "", v, false
		}
		v, ok := usageMetricValue(r.usageAggregate, metric)
		if !ok {
			return "", -1, false
		}
		return "", v, false
	}
	return "", float64(r.Tokens.Total), false
}
func queryUsageTable(rows []usageTableRow, t *usageTableState, kind, metric, latency string) usageTableResult {
	needle := strings.ToLower(strings.TrimSpace(t.Search))
	out := make([]usageTableRow, 0, len(rows))
	for _, r := range rows {
		if needle == "" || strings.Contains(strings.ToLower(usageRowText(r, kind)), needle) {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		as, av, text := usageRowField(a, t.Sort.Column, metric, latency)
		bs, bv, _ := usageRowField(b, t.Sort.Column, metric, latency)
		cmp := 0
		if text {
			cmp = strings.Compare(strings.ToLower(as), strings.ToLower(bs))
		} else if av < bv {
			cmp = -1
		} else if av > bv {
			cmp = 1
		}
		if cmp != 0 {
			if t.Sort.Descending {
				return cmp > 0
			}
			return cmp < 0
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Key < b.Key
	})
	result := usageTableResult{Total: len(out)}
	size := max(1, t.Size)
	result.Pages = max(1, (len(out)+size-1)/size)
	result.Page = max(1, min(result.Pages, t.Page))
	start := min(len(out), (result.Page-1)*size)
	end := min(len(out), start+size)
	result.Rows = out[start:end]
	for _, r := range out {
		mergeUsageAggregate(&result.Totals, r.usageAggregate)
	}
	for _, r := range result.Rows {
		mergeUsageAggregate(&result.PageTotals, r.usageAggregate)
	}
	return result
}
func mergeUsageAggregate(a *usageAggregate, b usageAggregate) {
	a.Requests += b.Requests
	a.Errors += b.Errors
	a.ToolFailures += b.ToolFailures
	a.ToolRuns += b.ToolRuns
	a.Tokens.Total = usageAdd(a.Tokens.Total, b.Tokens.Total)
	a.Tokens.Input = usageAdd(a.Tokens.Input, b.Tokens.Input)
	a.Tokens.Output = usageAdd(a.Tokens.Output, b.Tokens.Output)
	a.Tokens.CacheRead = usageAdd(a.Tokens.CacheRead, b.Tokens.CacheRead)
	a.Tokens.CacheWrite = usageAdd(a.Tokens.CacheWrite, b.Tokens.CacheWrite)
	a.Cost += b.Cost
	a.Priced += b.Priced
}
func usageTableColumns(kind, metric string) []ui.TableColumn {
	ids := []string{"title", "workspace", "provider", "model", "started", "duration", "requests", "input", "output", "cache", "tokens", "errors", "tools"}
	switch kind {
	case "series":
		ids = []string{"time", "value", "requests", "tokens"}
	case "daily":
		ids = []string{"date", "requests", "tokens", "cache-hit", "errors"}
	case "failures":
		ids = []string{"when", "kind", "model", "session", "message"}
	case "breakdown:tools":
		ids = []string{"tool", "calls", "errors", "avg", "max"}
	default:
		if strings.HasPrefix(kind, "breakdown:") {
			ids = []string{"name", "requests", "input", "output", "cache", "tokens", "share"}
		}
	}
	titles := map[string]string{"title": "Session", "name": "Name", "tool": "Tool", "when": "When", "kind": "Type", "cache-hit": "Cache hit", "value": titleCase(metric), "avg": "Avg", "max": "Max"}
	var cols []ui.TableColumn
	for i, id := range ids {
		title := titles[id]
		if title == "" {
			title = titleCase(id)
		}
		w := float32(96)
		if i == 0 {
			w = 220
		}
		if id == "model" || id == "session" {
			w = 160
		}
		if id == "message" {
			w = 320
		}
		align := ui.End
		if slices.Contains([]string{"title", "name", "tool", "workspace", "provider", "model", "when", "date", "time", "started", "kind", "session", "message"}, id) {
			align = ui.Start
		}
		cols = append(cols, ui.TableColumn{ID: id, Title: title, Width: w, MinWidth: 48, Align: align, Sortable: id != "share"})
	}
	return cols
}
func (a *App) usageTableRows(kind string) []usageTableRow {
	p := &a.usageState.projection
	var out []usageTableRow
	var groups []usageAggregate
	switch kind {
	case "sessions":
		groups = p.Sessions
	case "series", "daily":
		points := p.Series
		if kind == "daily" {
			points = p.Daily
		}
		for i := range points {
			v := points[i]
			g := v.usageAggregate
			g.Key = v.At.Format(time.RFC3339)
			g.Name = v.At.Local().Format("Jan 2, 2006 15:04")
			g.At = v.At
			out = append(out, usageTableRow{usageAggregate: g, Point: &v})
		}
	case "failures":
		for i := range p.Errors[:min(400, len(p.Errors))] {
			e := p.Errors[i]
			out = append(out, usageTableRow{usageAggregate: usageAggregate{Key: fmt.Sprintf("%s|%s|%d|%s", e.SessionPath, e.Kind, e.At.UnixNano(), e.Message), Name: e.SessionTitle, ID: e.Session, Path: e.SessionPath, Provider: e.Provider, Model: e.Model, Workspace: e.Workspace, At: e.At}, Failure: &e})
		}
	case "breakdown:models":
		groups = p.Models
	case "breakdown:providers":
		groups = p.Providers
	case "breakdown:workspaces":
		groups = p.Workspaces
	case "breakdown:tools":
		groups = p.ToolGroups
	}
	for _, r := range groups {
		out = append(out, usageTableRow{usageAggregate: r})
	}
	return out
}
func (a *App) usageCell(r usageTableRow, id string) string {
	s := &a.usageState
	if r.Failure != nil {
		e := r.Failure
		switch id {
		case "when":
			return e.At.Local().Format("Jan 2, 15:04")
		case "kind":
			return titleCase(e.Kind)
		case "model":
			return e.Model
		case "session":
			return e.SessionTitle
		case "message":
			return e.Message
		}
	}
	switch id {
	case "title", "name", "tool":
		return r.Name
	case "workspace":
		return r.Workspace
	case "provider":
		return r.Provider
	case "model":
		return r.Model
	case "time", "date":
		return r.Name
	case "started":
		if r.Started.IsZero() {
			return "—"
		}
		return r.Started.Local().Format("Jan 2 15:04")
	case "duration":
		if r.Started.IsZero() || r.Ended.IsZero() {
			return "—"
		}
		return max(time.Duration(0), r.Ended.Sub(r.Started)).Round(time.Second).String()
	case "requests", "calls":
		return fmt.Sprint(r.Requests)
	case "input":
		return compactCount(r.Tokens.Input)
	case "output":
		return compactCount(r.Tokens.Output)
	case "cache":
		return compactCount(usageAdd(r.Tokens.CacheRead, r.Tokens.CacheWrite))
	case "cache-hit":
		rate, ok := usageCacheRate(r.Tokens)
		if !ok {
			return "—"
		}
		return fmt.Sprintf("%.1f%%", rate)
	case "errors":
		return fmt.Sprint(r.Errors + r.ToolFailures)
	case "tools":
		return fmt.Sprint(r.ToolRuns)
	case "avg":
		if r.Timed == 0 {
			return "—"
		}
		return fmt.Sprintf("%.2fs", r.Duration.Seconds()/float64(r.Timed))
	case "max":
		if r.Timed == 0 {
			return "—"
		}
		return r.MaxDuration.Round(time.Millisecond).String()
	case "share":
		return fmt.Sprintf("%.1f%%", usageShare(r.usageAggregate, s.projection.Summary)*100)
	case "value":
		v, ok := usageMetricValue(r.usageAggregate, s.metric)
		if s.metric == "latency" && r.Point != nil {
			v, ok = r.Point.Latency.value(s.latencyMetric)
		}
		return usageMetricText(v, ok, s.metric)
	}
	return compactCount(r.Tokens.Total)
}
func (a *App) usageTable(c *ui.Context, kind string) {
	s := &a.usageState
	t := s.table(kind)
	all := usageTableColumns(kind, s.metric)
	required := all[0].ID
	// Native MyGo sorts are two-state; normalize a descending-to-ascending third
	// click into the frozen source's default ordering, before querying the rows.
	if t.Sort != t.observedSort {
		if t.Sort.Column == t.observedSort.Column && t.sortPhase == 2 {
			t.Sort = usageDefaultSort(kind)
			t.sortPhase = 0
		} else if t.Sort.Column == t.observedSort.Column && t.sortPhase == 1 {
			t.sortPhase = 2
		} else {
			t.sortPhase = 1
		}
		t.observedSort = t.Sort
		t.Page = 1
		t.Selected = -1
		if kind == "sessions" {
			s.sort, s.descending = t.Sort.Column, t.Sort.Descending
			a.saveUsagePrefs()
		}
		if strings.HasPrefix(kind, "breakdown:") && kind != "breakdown:tools" {
			s.breakdownSort, s.breakdownDescending = t.Sort.Column, t.Sort.Descending
			for key, other := range s.tables {
				if strings.HasPrefix(key, "breakdown:") && key != "breakdown:tools" {
					other.Sort = t.Sort
					other.observedSort = t.Sort
				}
			}
			a.saveUsagePrefs()
		}
	}
	ui.Row(c.Key("usage-toolbar:" + kind)).Wrap().Gap(8).Children(func() {
		if ui.TextInput(c, &t.Search).Label("Search " + kind).Placeholder("Search…").Width(200).Changed() {
			t.Page = 1
			t.Selected = -1
		}
		ui.Spacer(c)
		ui.MenuButton(c, "Columns", func(m *ui.Menu) {
			for _, col := range all {
				if m.Item(col.Title).Checked(!slices.Contains(t.Hidden, col.ID) || col.ID == required).Disabled(col.ID == required).Chosen() {
					t.toggleColumn(col.ID, required)
					if kind == "sessions" {
						s.hiddenColumns = append([]string(nil), t.Hidden...)
						a.saveUsagePrefs()
					}
				}
			}
			m.Separator()
			if m.Item("Show all columns").Chosen() {
				t.Hidden = nil
				if kind == "sessions" {
					s.hiddenColumns = nil
					a.saveUsagePrefs()
				}
			}
			if m.Item("Reset column widths").Chosen() {
				t.List.Columns = ui.TableLayout{}
			}
		})
		ui.MenuButton(c, fmt.Sprintf("%d rows", t.Size), func(m *ui.Menu) {
			for _, size := range []int{10, 25, 50, 100} {
				if m.Item(fmt.Sprint(size)).Checked(t.Size == size).Chosen() {
					t.setPageSize(size)
					if kind == "sessions" {
						s.pageSize = t.Size
						a.saveUsagePrefs()
					}
				}
			}
		})
	})
	var cols []ui.TableColumn
	for _, col := range all {
		if col.ID == required || !slices.Contains(t.Hidden, col.ID) {
			cols = append(cols, col)
		}
	}
	result := queryUsageTable(a.usageTableRows(kind), t, kind, s.metric, s.latencyMetric)
	t.Page = result.Page
	if t.Selected >= len(result.Rows) {
		t.Selected = -1
	}
	t.List.Selected = &t.Selected
	t.List.Sort = &t.Sort
	t.List.Key = func(i int) any { return result.Rows[i].Key }
	t.List.Label = func(i int) string { return usageRowText(result.Rows[i], kind) }
	t.List.CellMenu = func(m *ui.Menu, row, col int) { a.usageRowMenu(m, kind, result.Rows[row]) }
	table := ui.Table(c.Key("usage-table:"+kind), &t.List, cols, len(result.Rows), func(row, col int) {
		r := result.Rows[row]
		text := a.usageCell(r, cols[col].ID)
		ui.Text(c, text).SingleLine().Tooltip(text)
	}).Height(float32(max(3, len(result.Rows))*40 + 40)).AlignSelf(ui.Stretch).Label("Usage " + kind + " table")
	if t.Selected >= 0 && t.Selected < len(result.Rows) {
		r := result.Rows[t.Selected]
		if table.Submitted() {
			if kind == "sessions" {
				a.openUsageSession(r)
			} else {
				a.activateUsageRow(kind, r)
			}
		} else if table.Changed() || (table.Clicked() && t.Sort == t.observedSort) {
			switch kind {
			case "sessions":
				s.filter.SessionPath = r.Path
				s.filter.Session = ""
				if r.Path == "" {
					s.filter.Session = r.ID
				}
			case "series", "breakdown:models", "breakdown:providers", "breakdown:workspaces":
				a.activateUsageRow(kind, r)
			}
		}
	}
	if result.Total == 0 {
		ui.Text(c, "No matching rows.").TextColor(a.Theme.TextMuted)
	}
	if kind != "failures" {
		ui.Text(c, fmt.Sprintf("Filtered: %d requests · %s tokens    Page: %d requests · %s tokens", result.Totals.Requests, compactCount(result.Totals.Tokens.Total), result.PageTotals.Requests, compactCount(result.PageTotals.Tokens.Total))).TextColor(a.Theme.TextMuted).FontSize(11)
	}
	ui.Row(c.Key("usage-pager:" + kind)).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, fmt.Sprintf("%d records · page %d of %d", result.Total, result.Page, result.Pages)).Grow(1).TextColor(a.Theme.TextMuted)
		if a.button(c, "Previous", "icons/arrow-left.svg", false).Disabled(result.Page <= 1).Clicked() {
			t.Page--
			t.Selected = -1
		}
		if a.button(c, "Next", "icons/arrow-right.svg", false).Disabled(result.Page >= result.Pages).Clicked() {
			t.Page++
			t.Selected = -1
		}
	})
}
func (a *App) activateUsageRow(kind string, r usageTableRow) {
	s := &a.usageState
	switch kind {
	case "sessions":
		copy := r.usageAggregate
		s.detailSession = &copy
	case "failures":
		copy := *r.Failure
		s.detailError = &copy
	case "series", "daily":
		if r.Point != nil {
			a.focusUsageBucket(r.Point.At, r.Point.End)
		}
	case "breakdown:models":
		s.toggleFilter("models", usageFilterOption{Model: workbench.UsageModelRef{Provider: r.Provider, Model: r.Model}})
	case "breakdown:providers":
		s.toggleFilter("providers", usageFilterOption{Key: r.Provider})
	case "breakdown:workspaces":
		s.toggleFilter("workspaces", usageFilterOption{Key: r.Workspace})
	}
}
func (a *App) usageRowMenu(m *ui.Menu, kind string, r usageTableRow) {
	s := &a.usageState
	if kind == "sessions" || kind == "failures" {
		if m.Item("Open session").Disabled(r.Path == "").Chosen() {
			a.Dispatch(Action{Kind: "SwitchSession", Value: r.Path})
			a.Navigate("chat")
		}
		if m.Item("Scope to this session").Chosen() {
			s.filter.SessionPath = r.Path
			s.filter.Session = ""
			if r.Path == "" {
				s.filter.Session = r.ID
			}
		}
		if m.Item("Copy session ID").Chosen() {
			a.Dispatch(Action{Kind: "CopyText", Value: r.ID})
		}
		m.Separator()
		if m.Item("Filter by provider: " + r.Provider).Chosen() {
			toggleUsageString(&s.filter.Providers, r.Provider)
		}
		if m.Item("Filter by model: " + r.Model).Chosen() {
			toggleUsageModel(&s.filter.ModelRefs, workbench.UsageModelRef{Provider: r.Provider, Model: r.Model})
		}
		if m.Item("Filter by workspace").Chosen() {
			toggleUsageString(&s.filter.Workspaces, r.Workspace)
		}
	} else if kind != "breakdown:tools" {
		if m.Item("Focus these results").Chosen() {
			a.activateUsageRow(kind, r)
		}
	}
}

func (a *App) openUsageSession(r usageTableRow) {
	if r.Path == "" {
		return
	}
	a.Dispatch(Action{Kind: "SwitchSession", Value: r.Path})
	a.Navigate("chat")
}
