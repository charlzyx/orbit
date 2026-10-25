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
	contextHandle        ui.Handle
	rowHandles           map[int]*ui.Handle
	Search               string
	sortPhase            int
	Sort, observedSort   ui.SortOrder
	Page, Size, Selected int
	Hidden               []string
	List                 ui.ListState
}
type usageTableRow struct {
	usageAggregate
	Failure                   *workbench.UsageError
	Point                     *usagePoint
	Label, Stamp, Granularity string
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
		if s.breakdownWidths == nil {
			s.breakdownWidths = map[string]float32{}
		}
		t.List.Columns.Widths = s.breakdownWidths
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
func usageRowFields(r usageTableRow, kind string) []string {
	if r.Failure != nil {
		e := r.Failure
		return []string{e.At.Local().Format("Jan 2, 15:04"), e.Kind, usageModelLabel(e.Model), e.SessionTitle, e.Message}
	}
	if r.Point != nil {
		return []string{r.Label, r.Stamp}
	}
	if kind == "breakdown:tools" {
		return []string{r.Name, workbench.UsageToolClass(r.Name)}
	}
	if strings.HasPrefix(kind, "breakdown:") {
		return []string{r.Name, r.Sub}
	}
	status := "ok"
	if r.Errors+r.ToolFailures > 0 {
		status = "error failed"
	}
	return []string{r.Name, r.ID, usageWorkspaceLabel(r.Workspace), r.Workspace, usageProviderLabel(r.Provider), usageSessionModel(r.usageAggregate), status}
}
func usageRowText(r usageTableRow, kind string) string {
	return strings.Join(usageRowFields(r, kind), " ")
}
func usageRowMatches(r usageTableRow, kind, needle string) bool {
	if needle == "" {
		return true
	}
	for _, field := range usageRowFields(r, kind) {
		if strings.Contains(strings.ToLower(field), needle) {
			return true
		}
	}
	return false
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
			return usageModelLabel(e.Model), 0, true
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
		return usageWorkspaceLabel(r.Workspace), 0, true
	case "provider":
		return usageProviderLabel(r.Provider), 0, true
	case "model":
		return usageSessionModel(r.usageAggregate), 0, true
	case "started":
		return "", float64(r.Ended.UnixMilli()), false
	case "date", "time", "latest":
		return "", float64(r.At.UnixMilli()), false
	case "requests", "calls":
		return "", float64(r.Requests), false
	case "input":
		return "", float64(r.Tokens.Input), false
	case "output":
		return "", float64(r.Tokens.Output), false
	case "cache":
		return "", float64(r.Tokens.CacheRead), false
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
		if r.Timed == 0 || r.MaxDuration == 0 {
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
		if usageRowMatches(r, kind, needle) {
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
		if kind == "series" || kind == "daily" || kind == "failures" {
			if !a.At.Equal(b.At) {
				return a.At.Before(b.At)
			}
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if kind == "sessions" && a.ID != b.ID {
			return a.ID < b.ID
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
	a.Timed += b.Timed
	a.Duration += b.Duration
	a.MaxDuration = max(a.MaxDuration, b.MaxDuration)
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
		cols = append(cols, ui.TableColumn{ID: id, Title: title, Width: w, MinWidth: 48, Align: align, Fixed: id == "share", Sortable: id != "share" && id != "message" && kind != "breakdown:tools"})
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
		granularity := p.Granularity
		if kind == "daily" && granularity == "hour" {
			granularity = "day"
		}
		for i := range points {
			v := points[i]
			g := v.usageAggregate
			g.Key = v.At.Format(time.RFC3339)
			label, stamp := usageBucketLabel(v.At, granularity, a.usageState.now()), a.usageStamp(v.At, granularity)
			g.Name = stamp
			if kind == "daily" {
				g.Name = label
			}
			g.At = v.At
			out = append(out, usageTableRow{usageAggregate: g, Point: &v, Label: label, Stamp: stamp, Granularity: granularity})
		}
	case "failures":
		for i := range p.Errors[:min(400, len(p.Errors))] {
			e := p.Errors[i]
			e.SessionTitle = a.usageSessionTitle(e.SessionTitle, e.Session)
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
		if kind == "sessions" && r.Untitled {
			r.Name = a.usageSessionTitle("", r.ID)
		}
		out = append(out, usageTableRow{usageAggregate: r})
	}
	return out
}
func (a *App) usageCell(r usageTableRow, id, kind string) string {
	s := &a.usageState
	if r.Failure != nil {
		e := r.Failure
		switch id {
		case "when":
			return e.At.Local().Format("Jan 2, 15:04")
		case "kind":
			return a.text("usage.error_kind_"+e.Kind, titleCase(e.Kind))
		case "model":
			return usageModelLabel(e.Model)
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
		return usageWorkspaceLabel(r.Workspace)
	case "provider":
		return usageProviderLabel(r.Provider)
	case "model":
		if kind == "sessions" {
			return usageSessionModel(r.usageAggregate)
		}
		return r.Model
	case "time", "date":
		return r.Name
	case "started":
		if r.Ended.IsZero() {
			return "—"
		}
		if r.Ended.Local().Year() == s.now().Year() {
			return r.Ended.Local().Format("Jan 2")
		}
		return r.Ended.Local().Format("Jan 2, 2006")
	case "duration":
		if r.Started.IsZero() || r.Ended.IsZero() {
			return "—"
		}
		return usageSpan(r.Ended.Sub(r.Started))
	case "requests", "calls":
		return usageCount(uint64(r.Requests))
	case "input":
		return compactCount(r.Tokens.Input)
	case "output":
		return compactCount(r.Tokens.Output)
	case "cache":
		if kind == "sessions" && r.Tokens.CacheRead == 0 && r.Tokens.CacheWrite == 0 {
			return "—"
		}
		return compactCount(r.Tokens.CacheRead)
	case "cache-hit":
		rate, ok := usageCacheRate(r.Tokens)
		if !ok {
			return "—"
		}
		return usagePercent(rate)
	case "errors":
		if kind == "breakdown:tools" && r.ToolFailures == 0 {
			return "—"
		}
		return usageCount(uint64(r.Errors + r.ToolFailures))
	case "tools":
		return fmt.Sprint(r.ToolRuns)
	case "avg":
		if r.Timed == 0 {
			return "—"
		}
		return usageDuration(r.Duration.Seconds() / float64(r.Timed))
	case "max":
		if r.Timed == 0 || r.MaxDuration == 0 {
			return "—"
		}
		return usageDuration(r.MaxDuration.Seconds())
	case "share":
		return usagePercent(usageShare(r.usageAggregate, s.projection.Summary) * 100)
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
	all := a.usageColumns(kind, s.metric)
	required := all[0].ID
	ui.Row(c.Key("usage-toolbar:" + kind)).Gap(a.Theme.Space(8)).AlignItems(ui.Center).Children(func() {
		ui.Row(c).Grow(1).MinWidth(180).Height(a.Theme.Px(36)).Padding(0, a.Theme.Px(10)).Gap(a.Theme.Space(8)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(a.Theme.Border).Clip().Label("Usage table search head " + kind).Children(func() {
			a.icon(c, "icons/search.svg", a.Theme.Px(14), a.Theme.TextMuted)
			if ui.TextInputBase(c, &t.Search).Label("Search " + kind).Placeholder(a.usageSearchPlaceholder(kind)).Grow(1).MinWidth(0).FontSize(a.Theme.Px(14)).FocusRing(false).Changed() {
				if !strings.HasPrefix(kind, "breakdown:") {
					t.Page = 1
				}
				t.Selected = -1
			}
		})
		a.usageColumnControl(c, kind, all, required)
	})
	probe := ui.Box(c.Key("usage-table-width:" + kind)).Height(0).AlignSelf(ui.Stretch)
	width := probe.Bounds().W
	if width <= 0 {
		width, _ = c.Size()
		if a.SidebarVisible {
			width -= a.SidebarWidth
		}
		width = min(width, 1180) - 64
	}
	var cols []ui.TableColumn
	for _, col := range all {
		if col.ID == required || !slices.Contains(t.Hidden, col.ID) {
			cols = append(cols, col)
		}
	}
	if kind == "sessions" {
		cols = usageSessionColumns(cols, t.List.Columns.Widths, width)
	} else {
		cols = usageDataColumns(cols, t.List.Columns.Widths, width, kind)
	}
	result := queryUsageTable(a.usageTableRows(kind), t, kind, s.metric, s.latencyMetric)
	t.Page = result.Page
	if t.Selected >= len(result.Rows) {
		t.Selected = -1
	}
	interactive := kind == "sessions" || kind == "series" || kind == "breakdown:models" || kind == "breakdown:providers" || kind == "breakdown:workspaces"
	t.List.Selected = nil
	if interactive {
		t.List.Selected = &t.Selected
	} else {
		t.Selected = -1
	}
	for i, row := range result.Rows {
		if kind == "sessions" && ((s.filter.SessionPath != "" && row.Path == s.filter.SessionPath) || (s.filter.Session != "" && row.ID == s.filter.Session)) {
			t.Selected = i
		}
		if kind == "series" && row.Point != nil && s.filter.FocusStart.Equal(row.Point.At) && s.filter.FocusEnd.Equal(row.Point.End) {
			t.Selected = i
		}
	}
	t.List.Sort = &t.Sort
	if kind == "breakdown:tools" {
		t.List.Sort = nil
	}
	t.List.Key = func(i int) any { return result.Rows[i].Key }
	t.List.Label = func(i int) string { return usageRowText(result.Rows[i], kind) }
	t.List.CellMenu = nil

	table, body := a.usageNativeTable(c, kind, cols, result)
	if t.Selected >= 0 && t.Selected < len(result.Rows) {
		r := result.Rows[t.Selected]
		if body.Submitted() {
			if kind == "sessions" {
				a.openUsageSession(r)
			} else {
				a.activateUsageRow(kind, r)
			}
		} else if body.Changed() || body.Clicked() {
			switch kind {
			case "sessions":
				s.menu("context").Open = false
				if s.filterMenu == "context" {
					s.filterMenu = ""
				}
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
		table.Children(func() {
			ui.Box(c).Absolute().Top(40).Bottom(0).Left(0).Right(0).Center().Children(func() {
				ui.Text(c, a.usageTableEmpty(kind, false)).FontSize(a.Theme.Px(14)).TextColor(a.Theme.TextMuted)
			})
		})
	}
	filtered, pageTotals := a.usageTableTotals(kind, result)
	ui.Row(c).AlignSelf(ui.Stretch).Gap(a.Theme.Space(12)).Children(func() {
		ui.Text(c, filtered).TextColor(a.Theme.TextMuted).FontSize(a.Theme.Px(12))
		ui.Spacer(c)
		ui.Text(c, pageTotals).TextColor(a.Theme.TextMuted).FontSize(a.Theme.Px(12))
	})
	ui.Row(c.Key("usage-pager:" + kind)).Gap(8).AlignItems(ui.Center).Children(func() {
		summary := a.usageTableEmpty(kind, true)
		if result.Total > 0 {
			first := (result.Page-1)*t.Size + 1
			summary = a.usageText("usage.showing_range", "Showing %{first}–%{last} of %{total}", map[string]string{"first": fmt.Sprint(first), "last": fmt.Sprint(first + len(result.Rows) - 1), "total": usageCount(uint64(result.Total))})
		}
		ui.Text(c, summary).Grow(1).TextColor(a.Theme.TextMuted)
		ui.Text(c, a.text("view.rows", "Rows")).FontSize(12).TextColor(a.Theme.TextMuted)
		a.usagePageSizeControl(c, kind)
		ui.Text(c, fmt.Sprintf("%d/%d", result.Page, result.Pages)).FontSize(12).TextColor(a.Theme.TextMuted)
		if a.button(c, a.text("usage.previous", "Previous"), "icons/arrow-left.svg", false).Disabled(result.Page <= 1).Clicked() {
			t.Page--
			t.Selected = -1
		}
		if a.button(c, a.text("usage.next", "Next"), "icons/arrow-right.svg", false).Disabled(result.Page >= result.Pages).Clicked() {
			t.Page++
			t.Selected = -1
		}
	})
	if kind == "failures" {
		ui.Text(c, a.text("view.retries_are_unavailable_pi_records_automatic_ret", "Retries are unavailable: pi records automatic retries as live events, not in session files.")).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted)
	}
}
func (a *App) activateUsageRow(kind string, r usageTableRow) {
	s := &a.usageState
	switch kind {
	case "series":
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

func (a *App) openUsageSession(r usageTableRow) {
	if r.Path == "" {
		return
	}
	a.Dispatch(Action{Kind: "SwitchSession", Value: r.Path})
	a.Navigate("chat")
}

func usageSessionColumns(columns []ui.TableColumn, resized map[string]float32, available float32) []ui.TableColumn {
	defaults := map[string]float32{"workspace": 110, "provider": 110, "model": 156, "started": 88, "duration": 92, "requests": 92, "input": 70, "output": 76, "cache": 72, "tokens": 82, "errors": 82}
	optional := map[string]bool{"started": true, "input": true, "output": true, "cache": true}
	budget := max(360, available-48)
	width := func(id string, fallback float32) float32 {
		if w, ok := resized[id]; ok {
			return max(48, w)
		}
		return fallback
	}
	running := float32(0)
	for _, col := range columns {
		if col.ID != "title" && col.ID != "tools" && !optional[col.ID] {
			running += width(col.ID, defaults[col.ID])
		}
	}
	admitted := map[string]bool{}
	for _, col := range columns {
		if optional[col.ID] {
			w := width(col.ID, defaults[col.ID])
			if running+w+38+12+180 <= budget {
				running += w
				admitted[col.ID] = true
			}
		}
	}
	title := max(180, min(460, budget-running-38-12))
	out := make([]ui.TableColumn, 0, len(columns)+1)
	for _, col := range columns {
		if col.ID == "tools" || (optional[col.ID] && !admitted[col.ID]) {
			continue
		}
		fallback := defaults[col.ID]
		if col.ID == "title" {
			fallback = title
		}
		col.Width = width(col.ID, fallback)
		out = append(out, col)
	}
	return append(out, ui.TableColumn{ID: "open", Width: 38, MinWidth: 38, Align: ui.Center, Sortable: false, Fixed: true})
}

// Non-session plans follow the frozen source's per-table reading registers.
// Only the leading identity or trailing message absorbs available width.
func usageDataColumns(cols []ui.TableColumn, widths map[string]float32, budget float32, kind string) []ui.TableColumn {
	measures := map[string]float32{}
	flex, low, high := cols[0].ID, float32(160), float32(480)
	switch kind {
	case "series":
		measures = map[string]float32{"value": 112, "requests": 104, "tokens": 96}
	case "daily":
		measures = map[string]float32{"requests": 104, "tokens": 96, "cache-hit": 104, "errors": 84}
	case "failures":
		measures = map[string]float32{"when": 108, "kind": 96, "model": 150, "session": 200}
		for _, col := range cols {
			if col.ID == "message" {
				flex, low, high = "message", 180, 560
			}
		}
	case "breakdown:tools":
		measures = map[string]float32{"calls": 76, "errors": 80, "avg": 76, "max": 76}
		high = 460
	default:
		measures = map[string]float32{"requests": 92, "input": 80, "output": 80, "cache": 80, "tokens": 88, "share": 72}
		high = 460
	}
	fixed := float32(0)
	for _, col := range cols {
		if col.ID == flex {
			continue
		}
		w := measures[col.ID]
		if !strings.HasPrefix(kind, "breakdown:") && widths[col.ID] > 0 {
			w = widths[col.ID]
		}
		fixed += w
	}
	for i := range cols {
		w := measures[cols[i].ID]
		if cols[i].ID == flex {
			w = max(low, min(high, budget-fixed))
		}
		cols[i].Width = w
		if kind == "breakdown:tools" {
			cols[i].Fixed = true
			cols[i].Sortable = false
		}
	}
	return cols
}
