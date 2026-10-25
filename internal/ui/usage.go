package orbitui

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

type usageState struct {
	project                                  func(workbench.UsageIndex, workbench.UsageFilter, string, *time.Location) usageProjection
	lastRequest                              time.Time
	sourceFingerprint, scopeNotice           string
	sourceKnown                              bool
	noticeUntil                              time.Time
	scan                                     func(context.Context) (workbench.UsageIndex, error)
	tables                                   map[string]*usageTableState
	hiddenColumns                            []string
	breakdownSort, latencyMetric, filterMenu string
	breakdownDescending                      bool
	filterViews                              map[string]*usageFilterView
	calendarRestore                          *usageCalendarRestore
	customMonth                              time.Time
	customPicking                            bool
	calendarSelected                         int

	fixtureNow                                                time.Time
	detailError                                               *workbench.UsageError
	scanner                                                   *workbench.UsageScanner
	index                                                     workbench.UsageIndex
	filter                                                    workbench.UsageFilter
	projection                                                usageProjection
	loading, loaded, pending, projecting                      bool
	err, preset, mode, metric, granularity, breakdown, detail string
	requested, computed                                       string
	revision, generation                                      uint64
	updated                                                   time.Time
	list                                                      ui.ListState
	scroll                                                    ui.ScrollState
	selected                                                  int
	page, pageSize                                            int
	sort                                                      string
	descending                                                bool
	detailRecord                                              *workbench.UsageRecord
	detailSession                                             *usageAggregate
	customOpen                                                bool
	customStart, customEnd, customError                       string
}

func (a *App) loadUsagePrefs() {
	s := &a.usageState
	var saved struct {
		Preset, Metric, UsageMode, BreakdownTab, DetailTab, SessionSort, LatencyMetric, BreakdownSort string
		HiddenColumns                                                                                 []string
		BreakdownSortDesc                                                                             bool
		PageSize                                                                                      int
		SessionSortDesc                                                                               bool
		CustomStart, CustomEnd                                                                        int64
	}
	// Source keys use snake case, so preserve them explicitly.
	var raw map[string]json.RawMessage
	_ = a.Options.Store.Read("usage.json", &raw)
	for key, target := range map[string]any{"preset": &saved.Preset, "metric": &saved.Metric, "usage_mode": &saved.UsageMode, "breakdown_tab": &saved.BreakdownTab, "detail_tab": &saved.DetailTab, "session_sort": &saved.SessionSort, "page_size": &saved.PageSize, "session_sort_desc": &saved.SessionSortDesc, "custom_start": &saved.CustomStart, "custom_end": &saved.CustomEnd, "latency_metric": &saved.LatencyMetric, "hidden_columns": &saved.HiddenColumns, "breakdown_sort": &saved.BreakdownSort, "breakdown_sort_desc": &saved.BreakdownSortDesc} {
		if raw[key] != nil {
			_ = json.Unmarshal(raw[key], target)
		}
	}
	s.preset, s.metric, s.mode, s.breakdown, s.detail, s.sort = saved.Preset, saved.Metric, saved.UsageMode, saved.BreakdownTab, saved.DetailTab, saved.SessionSort
	if s.preset == "" {
		s.preset = "last7"
	}
	if s.metric == "" {
		s.metric = "tokens"
	}
	if s.mode != "details" {
		s.mode = "simple"
	}
	if s.breakdown == "" {
		s.breakdown = "models"
	}
	if s.detail == "" {
		s.detail = "sessions"
	}
	if s.sort == "" {
		s.sort = "tokens"
	}
	s.preset = usageValid(s.preset, "last7", "today", "yesterday", "last7", "last14", "last30", "this-week", "this-month", "previous-month", "custom", "all")
	s.metric = usageValid(s.metric, "tokens", "tokens", "requests", "input", "output", "cache", "cost", "latency", "errors")
	s.sort = usageValid(s.sort, "tokens", "title", "workspace", "provider", "model", "started", "duration", "requests", "input", "output", "cache", "tokens", "errors", "tools")
	s.detail = usageValid(s.detail, "sessions", "sessions", "daily", "failures")
	s.breakdown = usageValid(s.breakdown, "models", "models", "providers", "workspaces", "tools")
	s.latencyMetric = usageValid(saved.LatencyMetric, "average", "average", "p50", "p95", "p99")
	s.breakdownSort = usageValid(saved.BreakdownSort, "tokens", "name", "requests", "input", "output", "cache", "tokens")
	s.breakdownDescending = saved.BreakdownSortDesc
	if raw["breakdown_sort_desc"] == nil {
		s.breakdownDescending = true
	}
	s.hiddenColumns = nil
	for _, v := range saved.HiddenColumns {
		for _, col := range usageTableColumns("sessions", s.metric) {
			if v == col.ID {
				s.hiddenColumns = append(s.hiddenColumns, v)
				break
			}
		}
	}
	s.tables = nil
	s.calendarSelected = -1
	s.granularity = "day"
	s.selected = -1
	s.page = 1
	s.pageSize = saved.PageSize
	s.descending = saved.SessionSortDesc
	if raw["session_sort_desc"] == nil {
		s.descending = true
	}
	if s.pageSize != 10 && s.pageSize != 25 && s.pageSize != 50 && s.pageSize != 100 {
		s.pageSize = 25
	}
	s.filter.Start, s.filter.End = workbench.UsageRange(s.preset, time.Now())
	if s.preset == "custom" {
		if saved.CustomStart > 0 && saved.CustomEnd > saved.CustomStart {
			s.filter.Start = bucketStart(time.UnixMilli(saved.CustomStart), "day", time.Local)
			s.filter.End = bucketStart(time.UnixMilli(saved.CustomEnd), "day", time.Local).AddDate(0, 0, 1)
		} else {
			s.preset = "last7"
			s.filter.Start, s.filter.End = workbench.UsageRange(s.preset, time.Now())
		}
	}
	s.granularity = usageAutoGranularity(s.filter, s.index)
}
func (a *App) saveUsagePrefs() {
	s := &a.usageState
	values := map[string]any{"preset": s.preset, "metric": s.metric, "usage_mode": s.mode, "breakdown_tab": s.breakdown, "detail_tab": s.detail, "session_sort": s.sort, "session_sort_desc": s.descending, "page_size": s.pageSize, "latency_metric": s.latencyMetric, "hidden_columns": s.hiddenColumns, "breakdown_sort": s.breakdownSort, "breakdown_sort_desc": s.breakdownDescending}
	if s.preset == "custom" {
		values["custom_start"] = s.filter.Start.UnixMilli()
		values["custom_end"] = bucketStart(s.filter.End.Add(-time.Nanosecond), "day", time.Local).UnixMilli()
	}
	a.save(a.Options.Store.Update("usage.json", func(raw map[string]json.RawMessage) error {
		for k, v := range values {
			b, e := json.Marshal(v)
			if e != nil {
				return e
			}
			raw[k] = b
		}
		return nil
	}))
}
func (a *App) refreshUsage() {
	s := &a.usageState
	if s.loading {
		s.pending = true
		return
	}
	if s.scanner == nil {
		s.scanner = workbench.NewUsageScanner(filepath.Join(a.Options.Store.PiDir, "sessions"))
	}
	s.loading = true
	s.lastRequest = time.Now()
	scan := s.scan
	if scan == nil {
		scan = s.scanner.Scan
	}
	a.Async(func() func() {
		v, e := scan(context.Background())
		return func() {
			s.loading = false
			s.err = ""
			if e != nil {
				s.err = e.Error()
			} else {
				s.index = v
				s.granularity = usageAutoGranularity(s.filter, s.index)
				if pruneUsageFilter(v, &s.filter) {
					s.scopeNotice = "A removed session scope was cleared."
					s.noticeUntil = time.Now().Add(6 * time.Second)
				}
				s.loaded = true
				s.revision++
				s.updated = v.ScannedAt
				if s.updated.IsZero() {
					s.updated = time.Now()
				}
			}
			if s.pending {
				s.pending = false
				a.refreshUsage()
			}
		}
	})
}
func (a *App) projectUsage() {
	s := &a.usageState
	if !s.loaded {
		return
	}
	signature := fmt.Sprintf("%d|%s|%+v", s.revision, s.granularity, s.filter)
	if signature == s.requested {
		return
	}
	s.generation++
	generation := s.generation
	s.requested = signature
	s.projecting = true
	index, filter, granularity := s.index, cloneUsageFilter(s.filter), s.granularity
	project := s.project
	if project == nil {
		project = buildUsageProjection
	}
	a.Async(func() func() {
		p := project(index, filter, granularity, time.Local)
		return func() {
			if generation != s.generation {
				return
			}
			s.projection = p
			s.projecting = false
			s.computed = signature
			s.selected = -1
			s.page = 1
			for kind, t := range s.tables {
				if !strings.HasPrefix(kind, "breakdown:") {
					t.Page = 1
					t.Selected = -1
				}
			}
		}
	})
}
func (a *App) usagePage(c *ui.Context) {
	s := &a.usageState
	t := a.Theme
	if !s.loaded && !s.loading && s.err == "" {
		a.refreshUsage()
	}
	a.projectUsage()
	ui.Column(c.Key("usage-page")).Fill().MinHeight(0).Children(func() {
		ui.Row(c).Height(44).PaddingX(12).Gap(8).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			status := "Not scanned"
			if s.loading {
				status = "Refreshing…"
			} else if !s.updated.IsZero() {
				status = "Updated " + s.updated.Local().Format("15:04:05")
			}
			ui.Text(c, status).TextColor(t.TextMuted).Grow(1)
			a.button(c, a.text("usage.export", "Export"), "icons/download.svg", false).Disabled(!s.loaded || s.projecting || s.computed == "").Menu(func(m *ui.Menu) {
				for _, format := range []string{"csv", "json"} {
					title := a.text("usage.export_csv", "Filtered requests (CSV)")
					if format == "json" {
						title = a.text("usage.export_json", "Current view (JSON)")
					}
					if m.Item(title).Chosen() {
						a.Dispatch(Action{Kind: "ExportUsage", Value: format, Data: map[string]any{"index": s.index, "filter": cloneUsageFilter(s.filter), "options": workbench.UsageExportOptions{Preset: s.preset, Granularity: s.granularity, SessionSearch: s.table("sessions").Search, Location: time.Local, Now: time.Now(), Insights: a.usageInsightTexts(usageInsights(s.projection, s.preset == "all"))}}})
					}
				}
			})
			if a.button(c, a.text("common.refresh", "Refresh"), "icons/refresh.svg", false).Disabled(s.loading).Clicked() {
				a.refreshUsage()
			}
		})
		ui.Row(c).Wrap().Padding(8, 12).Gap(6).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			ui.MenuButton(c, usageRangeLabel(s.preset), func(m *ui.Menu) {
				for _, p := range []string{"today", "yesterday", "last7", "last14", "last30", "this-week", "this-month", "previous-month", "custom", "all"} {
					if m.Item(usageRangeLabel(p)).Checked(s.preset == p).Chosen() {
						if p == "custom" {
							s.customOpen = true
							s.customError = ""
							s.customPicking = false
							s.customMonth = bucketStart(time.Now(), "month", time.Local)
							s.customStart = time.Now().AddDate(0, 0, -6).Format("2006-01-02")
							s.customEnd = time.Now().Format("2006-01-02")
						} else {
							s.preset = p
							s.filter.Start, s.filter.End = workbench.UsageRange(p, time.Now())
							s.granularity = usageAutoGranularity(s.filter, s.index)
							s.calendarRestore = nil
							s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
							a.saveUsagePrefs()
						}
					}
				}
			})
			a.usageDimensionFilter(c, "workspaces")
			a.usageDimensionFilter(c, "providers")
			a.usageDimensionFilter(c, "models")
			ui.MenuButton(c, "Filters", func(m *ui.Menu) {
				if m.Item("Failed requests only").Checked(s.filter.ErrorsOnly).Chosen() {
					s.filter.ErrorsOnly = !s.filter.ErrorsOnly
					if s.filter.ErrorsOnly {
						s.filter.CachedOnly = false
					}
				}
				if m.Item("Cache hits only").Checked(s.filter.CachedOnly).Chosen() {
					s.filter.CachedOnly = !s.filter.CachedOnly
					if s.filter.CachedOnly {
						s.filter.ErrorsOnly = false
					}
				}
			})
			ui.Spacer(c)
			for _, mode := range []string{"simple", "details"} {
				if a.button(c, a.text("usage.mode_"+mode, titleCase(mode)), "", s.mode == mode).Clicked() {
					s.mode = mode
					a.saveUsagePrefs()
				}
			}
		})
		if s.err != "" {
			ui.Row(c).Padding(8, 12).Background(t.Error.Alpha(.08)).Children(func() {
				ui.Text(c, s.err).TextColor(t.Error).Grow(1).Selectable()
				if a.iconButton(c, "Dismiss usage error", "icons/x.svg").Clicked() {
					s.err = ""
				}
			})
		}
		if !s.loaded {
			if s.loading {
				a.usageSkeleton(c)
			} else {
				a.Empty(c, "Usage unavailable", "The local session history could not be loaded. Retry to check again.")
			}
			return
		}
		if s.projecting || s.computed == "" {
			a.usageSkeleton(c)
			return
		}
		ui.Scroll(c.Key("usage-body")).TrackScroll(&s.scroll).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).WidthPercent(100).MaxWidth(1180).AlignSelf(ui.Center).Padding(20).Gap(20).Children(func() {
				if len(s.index.Warnings) > 0 {
					ui.Column(c).Padding(12).Background(t.Warning.Alpha(.08)).Radius(8).Gap(4).Children(func() {
						ui.Text(c, "Some session files could not be included. The figures below cover only successfully read records.").TextColor(t.Warning)
						for _, warning := range s.index.Warnings[:min(3, len(s.index.Warnings))] {
							ui.Text(c, warning).FontSize(11).Selectable()
						}
					})
				}
				if s.scopeNotice != "" && time.Now().Before(s.noticeUntil) {
					ui.Text(c, s.scopeNotice).TextColor(t.TextMuted)
				}
				a.usageActiveFilters(c)
				if len(s.index.Records) == 0 && len(s.index.Tools) == 0 && len(s.index.Errors) == 0 && len(s.index.Warnings) > 0 {
					a.Empty(c, "Usage unavailable", "The session files could not be read. Retry after resolving the reported errors.")
					return
				}
				if s.projection.Summary.Requests == 0 && len(s.projection.Tools) == 0 && len(s.projection.Errors) == 0 {
					a.Empty(c, "No usage in this range", "No sample metrics are shown. Try another date range or clear filters.")
					return
				}
				if s.mode == "details" {
					a.usageDetails(c)
				} else {
					a.usageSimple(c)
				}
			})
		})
	})
	a.usageDialogs(c)
}
func usageRangeLabel(p string) string {
	if label := map[string]string{"today": "Today", "yesterday": "Yesterday", "last7": "Last 7 days", "last14": "Last 14 days", "last30": "Last 30 days", "this-week": "This week", "this-month": "This month", "previous-month": "Previous month", "custom": "Custom range…", "all": "All time", "all_time": "All time"}[p]; label != "" {
		return label
	}
	return p
}
func (a *App) usageSkeleton(c *ui.Context) {
	ui.Column(c).Padding(20).Gap(20).Children(func() {
		ui.Text(c, "Loading local usage…").TextColor(a.Theme.TextMuted)
		for _, h := range []float32{180, 150, 240} {
			ui.Box(c).Height(h).AlignSelf(ui.Stretch).Background(a.Theme.Elevated).Radius(12)
		}
	})
}
func (a *App) usageCard(c *ui.Context, title, subtitle string, body func()) {
	ui.Column(c).AlignSelf(ui.Stretch).Background(a.Theme.Elevated).Border(1, a.Theme.Border).Radius(12).Children(func() {
		ui.Column(c).Padding(12).Gap(4).Children(func() {
			ui.Text(c, title).FontWeight(500).FontSize(16)
			if subtitle != "" {
				ui.Text(c, subtitle).TextColor(a.Theme.TextMuted).FontSize(12)
			}
		})
		a.line(c)
		ui.Column(c).AlignSelf(ui.Stretch).Padding(12).Gap(12).Children(body)
	})
}
func (a *App) usageSimple(c *ui.Context) {
	s := &a.usageState
	p := s.projection
	ui.Text(c, a.text("usage.overview", "OVERVIEW")).FontSize(10).TextColor(a.Theme.TextMuted)
	a.usageHeatmap(c)
	a.usageSummary(c)
	a.usageCard(c, a.text("usage.over_time", "Usage over time"), "Click a bucket to focus the page. Arrow keys inspect the chart.", func() { a.usageChart(c) })
	a.usageCard(c, a.text("usage.buckets", "Buckets"), "The same buckets as the chart, with independent search and paging.", func() { a.usageTable(c, "series") })
	a.usageSignals(c)
	if p.Summary.Tokens.Total > 0 || p.CacheDataRequests > 0 {
		a.usageHealthDetails(c)
	}

}
func (a *App) usageHeatmap(c *ui.Context) {
	s := &a.usageState
	t := a.Theme
	a.usageCard(c, a.text("usage.daily_activity", "Daily activity"), "Each cell is a day over the last year. Click again to restore the previous range.", func() {
		start, first, last := usageCalendarBounds(s.index, time.Local)
		values := map[string]usagePoint{}
		maximum := 0.0
		activeValues := []float64{}
		for _, p := range s.projection.Calendar {
			values[p.At.Format("2006-01-02")] = p
			v, ok := usageMetricValue(p.usageAggregate, s.metric)
			if ok && !p.At.Before(first) && !p.At.After(last) {
				maximum = max(maximum, v)
				if v > 0 {
					activeValues = append(activeValues, v)
				}
			}
		}
		days := []time.Time{}
		for at := start; !at.After(last) || len(days)%7 != 0; at = at.AddDate(0, 0, 1) {
			days = append(days, at)
		}
		firstIndex, lastIndex := 0, len(days)-1
		for firstIndex < len(days) && days[firstIndex].Before(first) {
			firstIndex++
		}
		for lastIndex >= 0 && days[lastIndex].After(last) {
			lastIndex--
		}
		thresholds := usageHeatmapThresholds(activeValues)
		width, _ := c.Size()
		if a.SidebarVisible {
			width -= a.SidebarWidth
		}
		width = min(1180, width) - 64
		cellHeight := min(float32(22), max(float32(4), (width-28)/float32(len(days)/7)))
		ui.Row(c).Height(16).Padding(0, 0, 0, 28).Children(func() {
			previous := time.Month(0)
			for week := 0; week < len(days)/7; week++ {
				at := days[week*7]
				label := ""
				if at.Month() != previous {
					label = at.Format("Jan")
					previous = at.Month()
				}
				ui.Box(c).Grow(1).MinWidth(0).Children(func() { ui.Text(c, label).Absolute().Left(0).Top(0).NoWrap().FontSize(10).TextColor(t.TextMuted) })
			}
		})
		grid := ui.Row(c.Key("usage-calendar-grid")).Height(cellHeight * 7).Gap(0).Focusable().Label("Usage activity calendar").Description("Arrow keys inspect days. Enter selects a day. Escape clears inspection.")
		grid.Children(func() {
			ui.Column(c).Width(28).Children(func() {
				for i, name := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
					if i%2 != 0 || cellHeight < 8 {
						name = ""
					}
					ui.Text(c, name).Height(cellHeight).FontSize(10).TextColor(t.TextMuted)
				}
			})
			for week := 0; week < len(days)/7; week++ {
				ui.Column(c).Grow(1).MinWidth(0).Gap(0).Children(func() {
					for day := 0; day < 7; day++ {
						i := week*7 + day
						at := days[i]
						p, exists := values[at.Format("2006-01-02")]
						v, ok := usageMetricValue(p.usageAggregate, s.metric)
						inRange := i >= firstIndex && i <= lastIndex
						color := t.Background
						if exists && ok && maximum > 0 && inRange {
							level := usageHeatmapLevel(v, thresholds, maximum)
							color = t.Accent.Alpha([]float32{0, .24, .46, .72, 1}[level])
						}
						border := t.Border
						if i == s.calendarSelected {
							border = t.Accent
						}
						if !inRange {
							ui.Box(c).Height(cellHeight).AlignSelf(ui.Stretch)
							continue
						}
						cell := ui.ButtonBase(c).Height(cellHeight).AlignSelf(ui.Stretch).Radius(3).Background(color).Border(1, border).Label("Usage day " + at.Format("2006-01-02")).Tooltip(fmt.Sprintf("%s · %s", at.Format("Jan 2, 2006"), usageMetricText(v, ok, s.metric))).Disabled(!inRange)
						if cell.Clicked() {
							s.calendarSelected = i
							a.scopeUsageDay(at)
						}
					}
				})
			}
		})
		ui.Row(c).Gap(4).Justify(ui.End).Children(func() {
			ui.Text(c, "Less").FontSize(10).TextColor(t.TextMuted)
			for _, alpha := range []float32{0, .24, .46, .72, 1} {
				color := t.Background
				if alpha > 0 {
					color = t.Accent.Alpha(alpha)
				}
				ui.Box(c).Size(10, 10).Radius(2).Border(1, t.Border).Background(color)
			}
			ui.Text(c, "More").FontSize(10).TextColor(t.TextMuted)
		})
		move := func(delta int) {
			grid.Focus()
			if s.calendarSelected < firstIndex || s.calendarSelected > lastIndex {
				s.calendarSelected = lastIndex
			} else {
				s.calendarSelected = max(firstIndex, min(lastIndex, s.calendarSelected+delta))
			}
		}
		switch {
		case grid.Shortcut(0, ui.KeyLeft):
			move(-7)
		case grid.Shortcut(0, ui.KeyRight):
			move(7)
		case grid.Shortcut(0, ui.KeyUp):
			move(-1)
		case grid.Shortcut(0, ui.KeyDown):
			move(1)
		case grid.Shortcut(0, ui.KeyHome):
			grid.Focus()
			s.calendarSelected = firstIndex
		case grid.Shortcut(0, ui.KeyEnd):
			grid.Focus()
			s.calendarSelected = lastIndex
		case grid.Shortcut(0, ui.KeyEscape):
			s.calendarSelected = -1
		case grid.Shortcut(0, ui.KeyEnter):
			if s.calendarSelected >= firstIndex && s.calendarSelected <= lastIndex {
				a.scopeUsageDay(days[s.calendarSelected])
			}
		}
		if s.calendarSelected >= firstIndex && s.calendarSelected <= lastIndex {
			at := days[s.calendarSelected]
			v, ok := usageMetricValue(values[at.Format("2006-01-02")].usageAggregate, s.metric)
			ui.Text(c, at.Format("Jan 2, 2006")+" · "+usageMetricText(v, ok, s.metric)).FontSize(11).TextColor(t.TextMuted)
		}
	})
}
func usageMetricText(value float64, valid bool, metric string) string {
	if !valid {
		return "Not measured"
	}
	switch metric {
	case "cost":
		return usageCost(value)
	case "latency":
		return usageDuration(value)
	default:
		return compactCount(uint64(value)) + " " + metric
	}
}
func (a *App) usageChart(c *ui.Context) {
	s := &a.usageState
	t := a.Theme
	ui.Row(c).Wrap().Gap(3).Children(func() {
		for _, metric := range []string{"tokens", "requests", "input", "output", "cache", "cost", "latency", "errors"} {
			if a.button(c, a.text("usage.metric_"+metric, titleCase(metric)), "", s.metric == metric).Height(26).Disabled(!usageMetricAvailable(s.projection.Summary, metric)).Clicked() {
				s.metric = metric
				a.saveUsagePrefs()
			}
		}
		if s.metric == "latency" {
			ui.Row(c).Gap(3).Margin(0, 0, 0, 8).Children(func() {
				for _, metric := range []string{"average", "p50", "p95", "p99"} {
					_, available := s.projection.Latency.value(metric)
					label := a.text("usage.latency_"+metric, strings.ToUpper(metric))
					if a.button(c.Key("usage-latency-"+metric), label, "", s.latencyMetric == metric).Height(26).Disabled(!available).Clicked() {
						s.latencyMetric = metric
						a.saveUsagePrefs()
					}
				}
			})
		}

	})
	if s.projection.SeriesTruncated {
		ui.Text(c, "Chart limited to the first 4,000 buckets. Choose a coarser interval to show the full range.").TextColor(t.Warning)
	}
	points := s.projection.Series
	if len(points) == 0 {
		ui.Text(c, "No buckets in this range.").TextColor(t.TextMuted)
		return
	}
	labels := make([]string, len(points))
	values := make([]float64, len(points))
	valid := make([]bool, len(points))
	maximum := 0.0
	for i, p := range points {
		labels[i] = p.At.Format("Jan 2")
		if s.granularity == "hour" {
			labels[i] = p.At.Format("Jan 2 15h")
		}
		values[i], valid[i] = usageMetricValue(p.usageAggregate, s.metric)
		if s.metric == "latency" {
			values[i], valid[i] = p.Latency.value(s.latencyMetric)
		}
		if valid[i] {
			maximum = max(maximum, values[i])
		}
	}
	x := chart.NewBand(labels...).Padding(0, 0)
	y := chart.NewLinear(0, usageNiceMaximum(maximum))
	ys := make([]float32, len(points))
	base := make([]float32, len(points))
	for i, v := range values {
		ys[i] = y.At(v)
		if !valid[i] {
			ys[i] = float32(math.NaN())
		}
	}
	xs := make([]float32, len(points))
	for i := range xs {
		xs[i] = x.At(i)
	}
	ticks := x.Ticks()
	if len(ticks) > 8 {
		step := (len(ticks) + 7) / 8
		var sparse []chart.Tick
		for i := 0; i < len(ticks); i += step {
			sparse = append(sparse, ticks[i])
		}
		ticks = sparse
	}
	chart.Plot(c, chart.Axes{Left: y.Ticks(4), Bottom: ticks, Label: func(side chart.Side, tick chart.Tick) {
		label := tick.Label
		if side == chart.Left {
			switch s.metric {
			case "cost":
				label = usageCost(tick.Value)
			case "latency":
				label = usageDuration(tick.Value)
			default:
				label = compactCount(uint64(max(0, tick.Value)))
			}
		}
		ui.Text(c, label).FontSize(10).TextColor(t.TextMuted).NoWrap()
	}}, func(area ui.Element) {
		area.Label("Usage trend").Description("Left and right arrow keys inspect buckets. Enter focuses the selected bucket.")
		area.Draw(func(p *ui.Painter, r ui.Rect) {
			for quarter := 0; quarter < 4; quarter++ {
				yy := r.Y + r.H*float32(quarter)/4
				for xx := r.X; xx < r.X+r.W; xx += 6 {
					p.Line(xx, yy, min(xx+4, r.X+r.W), yy, 1, t.Border)
				}
			}
			p.Line(r.X, r.Y+r.H, r.X+r.W, r.Y+r.H, 1, t.Border)
			p.FillPath(chart.Area(r, xs, base, ys, chart.CurveLinear), t.Accent.Alpha(.14))
			p.StrokePath(chart.Line(r, xs, ys, chart.CurveLinear), 1, t.Accent)
		})
		for i := range values {
			cell := chart.Rect(c, x.Band(i), chart.Span{Lo: 0, Hi: 1}).Tooltip(usagePointReadout(points[i], s.metric, s.latencyMetric)).Label("Usage bucket " + points[i].At.Format(time.RFC3339))
			if s.filter.FocusStart.Equal(points[i].At) && s.filter.FocusEnd.Equal(points[i].End) {
				cell.Background(t.Accent.Alpha(.08))
			}
			if cell.Clicked() {
				a.focusUsageBucket(points[i].At, points[i].End)
			}
		}
		chart.Select(c, area, &s.selected, xs, func(i int) string { return labels[i] + ": " + usageMetricText(values[i], valid[i], s.metric) })
		if i := s.selected; i >= 0 && i < len(points) {
			chart.RuleX(c, xs[i], 1).Background(t.TextMuted)
			if valid[i] {
				chart.Point(c, xs[i], ys[i], 5, 5).Radius(3).Background(t.Accent)
			}
			align := ui.Center
			if xs[i] < .2 {
				align = ui.Start
			} else if xs[i] > .8 {
				align = ui.End
			}
			chart.Anchor(c, xs[i], 1, align, ui.Start, func() {
				ui.Text(c, usagePointReadout(points[i], s.metric, s.latencyMetric)).FontSize(11).Padding(5, 8).Radius(5).Background(t.Surface).Border(1, t.Border).NoWrap()
			})
			if area.Shortcut(0, ui.KeyEnter) {
				a.focusUsageBucket(points[i].At, points[i].End)
			}
		}
	}).Height(168).Gap(8).AlignSelf(ui.Stretch)
	if s.projection.Latency.Samples > 0 {
		a.usageLatencyLine(c, s.projection.Latency, a.text("view.response_time", "Response time"))
	}
	if reason := a.usageUnavailableReason(s.projection.Summary); reason != "" {
		ui.Text(c, reason).FontSize(12).TextColor(t.TextMuted)
	}
}
func (a *App) usageUnavailableReason(summary workbench.UsageSummary) string {
	var metrics, reasons []string
	for _, metric := range []string{"tokens", "requests", "input", "output", "cache", "cost", "latency", "errors"} {
		if !usageMetricAvailable(summary, metric) {
			metrics = append(metrics, metric)
		}
	}
	if summary.PricedRequests == 0 {
		reasons = append(reasons, a.text("usage.reason_needs_price_table", "cost needs a price table for these models"))
	}
	if summary.TimedRequests == 0 {
		reasons = append(reasons, a.text("usage.reason_needs_request_gaps", "response time needs measurable request gaps"))
	}
	if len(reasons) == 0 {
		return ""
	}
	return strings.NewReplacer("%{metrics}", strings.Join(metrics, ", "), "%{reasons}", strings.Join(reasons, "; ")).Replace(a.text("usage.not_shown", "%{metrics} not shown: %{reasons}"))
}
func (a *App) usageDetails(c *ui.Context) {
	s := &a.usageState
	a.usageCard(c, a.text("usage.breakdown", "Breakdown"), "Ranked usage and independently searchable table.", func() {
		ui.Row(c).Wrap().Gap(4).Children(func() {
			for _, tab := range []string{"models", "workspaces", "providers", "tools"} {
				if a.button(c, a.text("usage.tab_"+tab, titleCase(tab)), "", s.breakdown == tab).Clicked() {
					s.breakdown = tab
					a.saveUsagePrefs()
				}
			}
		})
		rows := s.projection.Models
		switch s.breakdown {
		case "workspaces":
			rows = s.projection.Workspaces
		case "providers":
			rows = s.projection.Providers
		case "tools":
			rows = s.projection.ToolGroups
		}
		for _, r := range rows[:min(8, len(rows))] {
			r := r
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				if a.button(c, r.Name, "", false).Width(200).Clicked() {
					a.activateUsageRow("breakdown:"+s.breakdown, usageTableRow{usageAggregate: r})
				}
				value, maximum := float64(r.Tokens.Total), float64(rows[0].Tokens.Total)
				if s.breakdown == "tools" {
					value, maximum = float64(r.Requests), float64(rows[0].Requests)
				}
				ui.Box(c).Grow(1).Height(8).Background(a.Theme.Background).Radius(4).Children(func() {
					if maximum > 0 {
						ui.Box(c).Height(8).WidthPercent(float32(value / maximum * 100)).Background(a.Theme.Accent).Radius(4)
					}
				})
				ui.Text(c, compactCount(uint64(value))).Font(a.Fonts.CodeFamily).Width(85)
			})
		}
		if s.breakdown == "tools" {
			a.usageLatencyLine(c, s.projection.ToolLatency, a.text("usage.tool_response_time", "Tool response time"))
		}
		a.usageTable(c, "breakdown:"+s.breakdown)
	})
	a.usageCard(c, a.text("usage.records", "Records"), "Search affects only the selected table, not the page totals.", func() {
		ui.Row(c).Wrap().Gap(4).Children(func() {
			for _, tab := range []string{"sessions", "daily", "failures"} {
				if a.button(c, a.text("usage.tab_"+tab, titleCase(tab)), "", s.detail == tab).Clicked() {
					s.detail = tab
					a.saveUsagePrefs()
				}
			}
		})
		a.usageTable(c, s.detail)
	})
}
func (a *App) usageDialogs(c *ui.Context) {
	s := &a.usageState
	if s.customOpen {
		a.modal(c.Key("usage-custom-range"), &s.customOpen, 480, func() {
			ui.Text(c, "Custom date range").FontSize(18)
			a.usageCalendarPicker(c)
			ui.Text(c, "Start date (YYYY-MM-DD)")
			ui.TextInput(c, &s.customStart).Label("Start date").AlignSelf(ui.Stretch)
			ui.Text(c, "End date, inclusive (YYYY-MM-DD)")
			ui.TextInput(c, &s.customEnd).Label("End date").AlignSelf(ui.Stretch)
			if s.customError != "" {
				ui.Text(c, s.customError).TextColor(a.Theme.Error)
			}
			ui.Row(c).Gap(8).Children(func() {
				if a.button(c, "Cancel", "", false).Clicked() {
					s.customOpen = false
				}
				if a.button(c, "Apply date range", "", true).Clicked() {
					start, e := time.ParseInLocation("2006-01-02", strings.TrimSpace(s.customStart), time.Local)
					end, e2 := time.ParseInLocation("2006-01-02", strings.TrimSpace(s.customEnd), time.Local)
					if e != nil || e2 != nil {
						s.customError = "Enter valid dates in YYYY-MM-DD format."
					} else {
						if end.Before(start) {
							start, end = end, start
						}
						s.filter.Start, s.filter.End = start, end.AddDate(0, 0, 1)
						s.preset = "custom"
						s.granularity = usageAutoGranularity(s.filter, s.index)
						s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
						s.calendarRestore = nil
						s.customOpen = false
						a.saveUsagePrefs()
					}
				}
			})
		})
	}
	if s.detailError != nil {
		open := true
		r := *s.detailError
		a.modal(c.Key("usage-failure-detail"), &open, 600, func() {
			ui.Text(c, "Failure details").FontSize(18)
			ui.Text(c, r.At.Local().Format(time.RFC1123)+" · "+r.Kind)
			ui.Text(c, r.Provider+" / "+r.Model+" "+r.Tool)
			ui.Text(c, r.SessionTitle).Selectable()
			ui.Text(c, r.Message).TextColor(a.Theme.Error).Selectable()
			ui.Row(c).Gap(8).Children(func() {
				if a.button(c, "Open session", "icons/chat.svg", false).Disabled(r.SessionPath == "").Clicked() {
					a.Dispatch(Action{Kind: "SwitchSession", Value: r.SessionPath})
					a.Navigate("chat")
					open = false
				}
				if a.button(c, "Close", "", false).Clicked() {
					open = false
				}
			})
		})
		if !open {
			s.detailError = nil
		}
	}
	if s.detailRecord != nil {
		open := true
		r := *s.detailRecord
		a.modal(c.Key("usage-record-detail"), &open, 600, func() {
			ui.Text(c, "Request details").FontSize(18)
			ui.Text(c, r.At.Local().Format(time.RFC1123))
			ui.Text(c, r.Provider+" / "+r.Model)
			ui.Text(c, r.Workspace).Selectable()
			ui.Text(c, fmt.Sprintf("%s · %d tokens", r.Outcome, r.Tokens.Total))
			if r.HasCost {
				ui.Text(c, fmt.Sprintf("Cost: $%.6f", r.Cost))
			}
			if r.HasDuration {
				ui.Text(c, "Duration: "+r.Duration.String())
			}
			if r.Error != "" {
				ui.Text(c, r.Error).TextColor(a.Theme.Error).Selectable()
			}
			if a.button(c, "Close", "", false).Clicked() {
				open = false
			}
		})
		if !open {
			s.detailRecord = nil
		}
	}
	if s.detailSession != nil {
		open := true
		r := *s.detailSession
		a.modal(c.Key("usage-session-detail"), &open, 600, func() {
			ui.Text(c, "Session details").FontSize(18)
			ui.Text(c, r.Name).Selectable()
			ui.Text(c, r.Workspace).Selectable()
			ui.Text(c, fmt.Sprintf("%d requests · %s tokens · %d errors", r.Requests, compactCount(r.Tokens.Total), r.Errors))
			ui.Row(c).Gap(8).Children(func() {
				if a.button(c, "Filter this session", "", false).Clicked() {
					s.filter.SessionPath = r.Path
					if r.Path == "" {
						s.filter.Session = r.Key
					}
					open = false
				}
				if a.button(c, "Open session", "icons/chat.svg", false).Disabled(r.Path == "").Clicked() {
					a.Dispatch(Action{Kind: "SwitchSession", Value: r.Path})
					a.Navigate("chat")
					open = false
				}
				if a.button(c, "Close", "", false).Clicked() {
					open = false
				}
			})
		})
		if !open {
			s.detailSession = nil
		}
	}
}

func usageValid(v, fallback string, values ...string) string {
	v = strings.TrimSpace(v)
	for _, candidate := range values {
		if candidate == v {
			return v
		}
	}
	return fallback
}
func usageAutoGranularity(f workbench.UsageFilter, index workbench.UsageIndex) string {
	_, _, granularity := index.UsageChartRange(f, !f.Start.IsZero() && f.Start.UnixMilli() <= 0, time.Local, time.Time{})
	return granularity
}

func cloneUsageFilter(f workbench.UsageFilter) workbench.UsageFilter {
	f.ModelRefs = append([]workbench.UsageModelRef(nil), f.ModelRefs...)
	f.Models = append([]string(nil), f.Models...)
	f.Providers = append([]string(nil), f.Providers...)
	f.Workspaces = append([]string(nil), f.Workspaces...)
	return f
}

func (a *App) focusUsageBucket(start, end time.Time) {
	s := &a.usageState
	if s.filter.FocusStart.Equal(start) && s.filter.FocusEnd.Equal(end) {
		s.filter.FocusStart, s.filter.FocusEnd = time.Time{}, time.Time{}
	} else {
		s.filter.FocusStart, s.filter.FocusEnd = start, end
	}
}

func (a *App) openUsage() {
	s := &a.usageState
	s.scroll.X, s.scroll.Y = 0, 0
	if !s.loaded || s.updated.IsZero() || time.Since(s.updated) >= 60*time.Second {
		a.refreshUsage()
	}
}
func (a *App) closeUsage() {
	s := &a.usageState
	s.filterMenu = ""
	for _, v := range s.filterViews {
		v.Open = false
	}
	s.customOpen = false
	s.customPicking = false
	s.detailSession = nil
	s.detailError = nil
	s.detailRecord = nil
	s.selected = -1
	s.calendarSelected = -1
}
func (a *App) syncUsageStore() {
	s := &a.usageState
	if a.page != "usage" {
		return
	}
	var fingerprint strings.Builder
	for _, session := range a.Model.Sessions {
		fmt.Fprintf(&fingerprint, "%s|%d;", session.Path, session.Updated.UnixNano())
	}
	value := fingerprint.String()
	if !s.sourceKnown {
		s.sourceKnown = true
		s.sourceFingerprint = value
		return
	}
	if value != s.sourceFingerprint && time.Since(s.lastRequest) >= 3*time.Second {
		s.sourceFingerprint = value
		a.refreshUsage()
	}
}

func usageMetricAvailable(summary workbench.UsageSummary, metric string) bool {
	switch metric {
	case "cost":
		return summary.PricedRequests > 0
	case "latency":
		return summary.TimedRequests > 0
	case "tokens", "cache":
		return summary.Tokens.Total > 0
	default:
		return true
	}
}

func usageNiceMaximum(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
		return 1
	}
	base := math.Pow(10, math.Floor(math.Log10(v)))
	n := v / base
	for _, step := range []float64{1, 2, 2.5, 5, 10} {
		if n <= step {
			return step * base
		}
	}
	return 10 * base
}
func usagePointReadout(p usagePoint, metric, latency string) string {
	v, ok := usageMetricValue(p.usageAggregate, metric)
	if metric == "latency" {
		v, ok = p.Latency.value(latency)
	}
	text := fmt.Sprintf("%s · %s\n%d requests · %s tokens\nInput %s · Output %s\nCache read %s · Cache write %s", p.At.Local().Format("Jan 2, 2006 15:04"), usageMetricText(v, ok, metric), p.Requests, usageExact(p.Tokens.Total), usageExact(p.Tokens.Input), usageExact(p.Tokens.Output), usageExact(p.Tokens.CacheRead), usageExact(p.Tokens.CacheWrite))
	if p.Errors > 0 {
		text += fmt.Sprintf("\n%d failed requests", p.Errors)
	}
	return text
}
