package orbitui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type usageState struct {
	scanner          *workbench.UsageScanner
	index            workbench.UsageIndex
	filter           workbench.UsageFilter
	loading, loaded  bool
	err, preset, tab string
	list             ui.ListState
}

func (a *App) refreshUsage() {
	s := &a.usageState
	if s.loading {
		return
	}
	if s.scanner == nil {
		s.scanner = workbench.NewUsageScanner(filepath.Join(a.Options.Store.PiDir, "sessions"))
	}
	s.loading = true
	a.Async(func() func() {
		v, e := s.scanner.Scan(context.Background())
		return func() {
			s.loading = false
			s.loaded = true
			s.index = v
			s.err = ""
			if e != nil {
				s.err = e.Error()
			}
		}
	})
}
func (a *App) usagePage(c *ui.Context) {
	s := &a.usageState
	t := a.Theme
	if !s.loaded && !s.loading {
		a.refreshUsage()
	}
	if s.preset == "" {
		s.preset = "all_time"
	}
	ui.Column(c).Fill().Padding(24).Gap(20).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Usage").FontSize(24).FontWeight(500).Grow(1)
			ui.MenuButton(c, s.preset, func(m *ui.Menu) {
				for _, p := range []string{"today", "yesterday", "last7", "last30", "this-month", "previous-month", "all_time"} {
					if m.Item(p).Checked(s.preset == p).Chosen() {
						s.preset = p
						s.filter.Start, s.filter.End = workbench.UsageRange(p, time.Now())
					}
				}
			})
			if a.button(c, "Refresh", "icons/refresh.svg", false).Clicked() {
				a.refreshUsage()
			}
			if a.button(c, "Export CSV…", "icons/download.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "ExportUsage", Data: map[string]any{"index": s.index, "filter": s.filter}})
			}
		})
		ui.Row(c).Gap(6).Children(func() {
			for _, tab := range []string{"Overview", "Models", "Projects", "Tools", "Requests"} {
				if a.button(c, tab, "", s.tab == tab || (s.tab == "" && tab == "Overview")).Clicked() {
					s.tab = tab
				}
			}
			ui.Spacer(c)
			ui.TextInput(c, &s.filter.Search).Placeholder("Search usage…").Label("Search usage").Width(220)
		})
		if s.loading {
			ui.Spinner(c).Size(20, 20)
		}
		if s.err != "" {
			ui.Text(c, s.err).TextColor(t.Error)
		}
		summary := s.index.Summary(s.filter)
		if summary.Requests == 0 {
			a.Empty(c, "No usage recorded", "Usage is calculated from your local pi session history. No sample metrics are shown.")
			return
		}
		if s.tab == "" || s.tab == "Overview" {
			ui.Row(c).Gap(12).Children(func() {
				for _, stat := range []struct{ title, value string }{{"Requests", fmt.Sprint(summary.Requests)}, {"Tokens", fmt.Sprint(summary.Tokens.Total)}, {"Cost", fmt.Sprintf("$%.4f", summary.Cost)}, {"Cache hit", fmt.Sprintf("%.1f%%", summary.Tokens.CacheRate())}} {
					ui.Column(c).Grow(1).Padding(16).Border(1, t.Border).Radius(8).Gap(6).Children(func() { ui.Text(c, stat.title).TextColor(t.TextMuted); ui.Text(c, stat.value).FontSize(24) })
				}
			})
			ui.Text(c, "Daily tokens").FontWeight(500)
			series := s.index.Series(s.filter, "day", time.Local)
			var maximum uint64
			for _, b := range series {
				maximum = max(maximum, b.Tokens)
			}
			ui.Row(c).Height(140).Gap(3).AlignItems(ui.End).Children(func() {
				for _, b := range series {
					h := float32(0)
					if maximum > 0 {
						h = 128 * float32(b.Tokens) / float32(maximum)
					}
					ui.Box(c).Grow(1).Height(max(2, h)).Background(t.Accent).Radius(3).Tooltip(fmt.Sprintf("%s · %d tokens", b.At.Format("Jan 2"), b.Tokens))
				}
			})
			ui.Text(c, fmt.Sprintf("%d turns · %d tool runs · %d errors · %d aborted", summary.Turns, summary.ToolRuns, summary.Errors, summary.Aborted)).TextColor(t.TextMuted)
		}
		records := []workbench.UsageRecord{}
		for _, r := range s.index.Records {
			if s.filter.Matches(r) {
				records = append(records, r)
			}
		}
		switch s.tab {
		case "Models", "Projects":
			groups := map[string]workbench.UsageSummary{}
			var order []string
			for _, r := range records {
				key := r.Provider + " / " + r.Model
				if s.tab == "Projects" {
					key = r.Workspace
				}
				v, ok := groups[key]
				if !ok {
					order = append(order, key)
				}
				v.Requests++
				v.Tokens.Total += r.Tokens.Total
				v.Cost += r.Cost
				groups[key] = v
			}
			ui.List(c, &s.list, len(order), func(i int) {
				k := order[i]
				v := groups[k]
				a.setting(c, k, fmt.Sprintf("%d requests · %d tokens", v.Requests, v.Tokens.Total), func() { ui.Text(c, fmt.Sprintf("$%.4f", v.Cost)) })
			}).Grow(1)
		case "Tools":
			ui.List(c, &s.list, len(s.index.Tools), func(i int) {
				r := s.index.Tools[i]
				status := "Succeeded"
				if !r.Success {
					status = "Failed"
				}
				a.setting(c, r.Name, r.At.Local().Format("Jan 2 15:04")+" · "+r.Workspace, func() { ui.Text(c, status) })
			}).Grow(1)
		default:
			ui.List(c, &s.list, len(records), func(i int) {
				r := records[len(records)-1-i]
				a.setting(c, r.Model, r.At.Local().Format("Jan 2 15:04")+" · "+r.Workspace, func() { ui.Text(c, fmt.Sprintf("%d tokens · $%.4f", r.Tokens.Total, r.Cost)) })
			}).Grow(1)
		}
		for _, w := range s.index.Warnings {
			ui.Text(c, w).TextColor(t.Warning).MaxLines(2)
		}
	})
}
