package orbitui

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestUsageSourceCatalogIdentityTieEmptyModelAndNativeContext(t *testing.T) {
	root := t.TempDir()
	put := func(name, body string) {
		t.Helper()
		path := filepath.Join(root, "g", name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	put("a.jsonl", `{"type":"session","id":"catalog","cwd":"/catalog","timestamp":"2026-10-01T00:00:00Z"}
{"type":"message","message":{"role":"assistant","provider":"p","model":"zebra"}}
{"type":"message","message":{"role":"assistant","provider":"p","model":"alpha"}}
{"type":"message","message":{"role":"assistant","provider":"catalog-only","model":"not-billed"}}
`)
	put("b.jsonl", `{"type":"session","id":"tie","cwd":"/ties","timestamp":"2026-10-01T00:00:00Z"}
{"type":"message","timestamp":"2026-10-01T00:00:01Z","message":{"role":"user","content":"Tie session"}}
{"type":"message","timestamp":"2026-10-01T00:00:02Z","message":{"role":"assistant","provider":"p","model":"alpha","usage":{"input":10}}}
{"type":"message","timestamp":"2026-10-01T00:00:03Z","message":{"role":"assistant","provider":"p","model":"zebra","usage":{"input":10}}}
`)
	put("c.jsonl", `{"type":"session","id":"empty-model","cwd":"/empty-model","timestamp":"2026-10-01T00:00:00Z"}
{"type":"message","timestamp":"2026-10-01T00:00:01Z","message":{"role":"user","content":"Empty model session"}}
{"type":"message","timestamp":"2026-10-01T00:00:02Z","message":{"role":"assistant","usage":{"input":7}}}
`)
	put("d.jsonl", `{"type":"session","id":"user-only","cwd":"/user-only","timestamp":"2026-10-01T00:00:00Z"}
{"type":"message","timestamp":"2026-10-01T00:00:01Z","message":{"role":"user","content":"No assistant"}}
`)
	scanner := workbench.NewUsageScanner(root)
	a := fixtureUsageApp(t)
	s := &a.usageState
	for repeat := 0; repeat < 2; repeat++ {
		index, err := scanner.Scan(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		s.index = index
		s.filter = workbench.UsageFilter{}
		s.projection = buildUsageProjection(index, s.filter, "day", time.Local)
		var empty usageTableRow
		for _, row := range a.usageTableRows("sessions") {
			if row.ID == "tie" && (row.Model != "alpha" || a.usageCell(row, "model", "sessions") != "alpha · P") {
				t.Fatal("source intern ordinal tie lost", row)
			}
			if row.ID == "empty-model" {
				empty = row
			}
		}
		if !empty.HasModel || a.usageCell(empty, "model", "sessions") != " · unknown" {
			t.Fatal("empty actual identity became missing", empty)
		}
		if !slices.Contains(s.projection.WorkspaceOptions, "/user-only") {
			t.Fatal("valid user-only workspace absent")
		}
		options := s.filterOptions("models", "")
		if len(options) != 4 {
			t.Fatal("unbilled or empty model absent", options)
		}
		knownOnly := workbench.UsageModelRef{Provider: "catalog-only", Model: "not-billed"}
		s.filter.ModelRefs = []workbench.UsageModelRef{knownOnly}
		if pruneUsageFilter(s.index, &s.filter) || !slices.Contains(s.filter.ModelRefs, knownOnly) {
			t.Fatal("catalog-only model pruned")
		}
		s.filter = workbench.UsageFilter{}
		var unknown usageFilterOption
		for _, o := range s.filterOptions("providers", "") {
			if o.Key == "" {
				unknown = o
			}
		}
		if unknown.Label != "unknown" || s.filterChecked("providers", unknown) {
			t.Fatal("empty legacy provider made unknown selected", unknown)
		}
		s.toggleFilter("providers", unknown)
		if len(s.filter.Providers) != 1 || s.filter.Providers[0] != "" || index.Summary(s.filter).Requests != 1 {
			t.Fatal("empty provider scope lost", s.filter)
		}
		s.filter = workbench.UsageFilter{}
		s.table("sessions").Search = "Empty model session"
		tt := ui.NewTester(func(c *ui.Context) {
			c.SetTheme(a.Theme.Native())
			ui.Column(c).Fill().Padding(10).Children(func() { a.usageTable(c, "sessions") })
		}, 1200, 500)
		tt.Frame()
		if err := tt.RightClick("Empty model session"); err != nil {
			t.Fatal(err)
		}
		if err := tt.Click("Filter by model:  · unknown"); err != nil {
			t.Fatal(err, tt.Texts())
		}
		if len(s.filter.ModelRefs) != 1 || s.filter.ModelRefs[0] != (workbench.UsageModelRef{}) || index.Summary(s.filter).Requests != 1 {
			t.Fatal("native empty model context did not use exact identity", s.filter)
		}
		s.closeUsagePopovers("")
	}
}

func TestUsageSourceLocalizedUntitledSessionChipsAndExport(t *testing.T) {
	a := fixtureUsageApp(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	s := &a.usageState
	at := s.now()
	id := "汉字🙂abcdefghi"
	s.index = workbench.UsageIndex{Sessions: []workbench.UsageSession{{ID: id, Path: "/session", Workspace: "/workspace"}}, Records: []workbench.UsageRecord{{At: at, Session: id, SessionPath: "/session", Workspace: "/workspace", Provider: "first", Model: "m", Tokens: workbench.Tokens{Total: 10}}, {At: at, Session: id, SessionPath: "/session", Workspace: "/workspace", Provider: "second", Model: "m", Tokens: workbench.Tokens{Total: 1}}}, Errors: []workbench.UsageError{{At: at, Session: id, SessionPath: "/session", Kind: "provider", Message: "failure"}}}
	s.projection = buildUsageProjection(s.index, workbench.UsageFilter{}, "day", time.Local)
	rows := a.usageTableRows("sessions")
	if len(rows) != 1 || rows[0].Name != "会话 汉字🙂abcde" {
		t.Fatal(rows)
	}
	failures := a.usageTableRows("failures")
	if failures[0].Failure.SessionTitle != "会话 汉字🙂abcde" || s.index.Errors[0].SessionTitle != "" {
		t.Fatal("display fallback mutated raw session title", failures, s.index.Errors)
	}
	if !usageRowMatches(rows[0], "sessions", "会话") || !usageRowMatches(failures[0], "failures", "会话") {
		t.Fatal("localized fallback not searchable")
	}
	s.filter = workbench.UsageFilter{Start: at.AddDate(0, 0, -2), End: at.AddDate(0, 0, 1), SessionPath: "/session", FocusStart: at, FocusEnd: at.AddDate(0, 0, 1), Providers: []string{"second", "first"}, ModelRefs: []workbench.UsageModelRef{{Provider: "first", Model: "m"}}}
	options := s.selectedFilterOptions("providers")
	if len(options) != 2 || options[0].Key != "second" {
		t.Fatal("chips reordered by request rank", options)
	}
	all := s.allFilterOptions("providers")
	if len(all) != 3 || all[0].Key != "first" || all[1].Key != "second" {
		t.Fatal("Select all ignored source catalog order", all)
	}
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Padding(10).Children(func() { a.usageActiveFilters(c) })
	}, 2000, 150)
	tt.Frame()
	sessionLabel := a.usageText("usage.chip_session", "Session: %{name}", map[string]string{"name": "会话 汉字🙂abcde"})
	firstLabel := a.usageText("usage.chip_provider", "Provider: %{name}", map[string]string{"name": "First"})
	secondLabel := a.usageText("usage.chip_provider", "Provider: %{name}", map[string]string{"name": "Second"})
	focus, _ := tt.Find(a.usageStamp(at, "day"))
	session, _ := tt.Find(sessionLabel)
	second, _ := tt.Find(secondLabel)
	first, _ := tt.Find(firstLabel)
	if session.W == 0 || focus.X >= session.X || session.X >= second.X || second.X >= first.X {
		t.Fatal("source chip order", focus, session, second, first, tt.Texts())
	}
	if err := tt.Click(firstLabel); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(s.filter.Providers, "first") {
		t.Fatal("clicking chip label removed source scope")
	}
	clear, ok := tt.Find("Clear " + firstLabel)
	if !ok || clear.W != 16 || clear.H != 16 {
		t.Fatal("separate source clear control", clear)
	}
	if err := tt.Click("Clear " + firstLabel); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(s.filter.Providers, "first") || !slices.Contains(s.filter.Providers, "second") {
		t.Fatal(s.filter)
	}
	before := s.filter
	if err := tt.Click(a.text("usage.clear_all", "Clear all")); err != nil {
		t.Fatal(err)
	}
	if !s.filter.Start.Equal(before.Start) || !s.filter.End.Equal(before.End) || len(s.filter.Providers) != 0 || s.filter.SessionPath != "" || !s.filter.FocusStart.IsZero() {
		t.Fatal("Clear all changed date or retained narrowing", s.filter)
	}
	opts := a.usageExportOptions()
	if opts.SessionTitles["/session"] != "会话 汉字🙂abcde" {
		t.Fatal(opts.SessionTitles)
	}
	var b bytes.Buffer
	if err := s.index.ExportJSON(&b, workbench.UsageFilter{}, opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "会话 汉字🙂abcde") {
		t.Fatal("UI-resolved localized title lost by actual exporter", b.String())
	}
}
