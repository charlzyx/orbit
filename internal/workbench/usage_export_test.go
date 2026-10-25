package workbench

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestUsageJSONExportSourceSchemaAndMeasuredUnknowns(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "group", "s.jsonl"), usageFixture())
	index, e := NewUsageScanner(root).Scan(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	index.ScannedAt = now
	filter := UsageFilter{Start: now.Add(-24 * time.Hour), End: now}
	var buf bytes.Buffer
	if e := index.ExportJSON(&buf, filter, UsageExportOptions{Preset: "custom", Granularity: "day", Location: time.UTC, Now: now}); e != nil {
		t.Fatal(e)
	}
	var out map[string]any
	if e := json.Unmarshal(buf.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"generated_at_ms", "range", "summary", "previous", "series", "models", "providers", "workspaces", "sessions", "buckets", "calendar", "tools", "errors_detail", "insights"} {
		if _, ok := out[key]; !ok {
			t.Fatal("missing source export key", key)
		}
	}
	summary := out["summary"].(map[string]any)
	if summary["requests"] != float64(2) || summary["priced_requests"] != float64(2) || summary["tool_calls"] != float64(1) {
		t.Fatal(summary)
	}
	sessions := out["sessions"].([]any)
	session := sessions[0].(map[string]any)
	if session["title"] != "hi" || session["session_id"] != "s1" || session["started_ms"] != float64(time.Date(2026, 10, 25, 1, 0, 0, 0, time.UTC).UnixMilli()) || session["ended_ms"] != float64(time.Date(2026, 10, 25, 1, 0, 6, 0, time.UTC).UnixMilli()) {
		t.Fatal(session)
	}
	if len(out["calendar"].([]any)) != 366 || len(out["series"].([]any)) != 2 {
		t.Fatal("calendar/contiguous range mismatch")
	}
	first := out["series"].([]any)[0].(map[string]any)
	if first["requests"] != float64(0) || first["avg_duration_ms"] != nil {
		t.Fatal("invented missing timing", first)
	}
	tools := out["tools"].([]any)[0].(map[string]any)
	if tools["class"] != "terminal" || tools["avg_duration_ms"] != float64(1000) {
		t.Fatal(tools)
	}
	if out["previous"] != nil {
		t.Fatal("invented previous history")
	}
}
func TestUsageJSONAllTimeScopeCalendarAndNoMutation(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	index := UsageIndex{ScannedAt: now, Records: []UsageRecord{{At: now.AddDate(0, 0, -2), Session: "a", Provider: "p", Model: "m", Tokens: Tokens{Total: 5}, Cost: 1, HasCost: true}, {At: now.AddDate(0, 0, -1), Session: "a", Provider: "p", Model: "m", Tokens: Tokens{Total: 7}}}}
	var buf bytes.Buffer
	filter := UsageFilter{Start: time.UnixMilli(0), End: now, Provider: "p"}
	if e := index.ExportJSON(&buf, filter, UsageExportOptions{Preset: "all", Granularity: "day", Location: time.UTC, Now: now}); e != nil {
		t.Fatal(e)
	}
	var out map[string]any
	json.Unmarshal(buf.Bytes(), &out)
	if out["range"].(map[string]any)["start_ms"] != float64(0) || len(out["series"].([]any)) != 25 {
		t.Fatal("all-time range/series start contract")
	}
	if index.Records[0].ModelPriced || index.Records[1].ModelPriced {
		t.Fatal("export mutated caller's index")
	}
	if out["summary"].(map[string]any)["cache_hit_rate"] != nil {
		t.Fatal("invented cache measurement")
	}
	if e := index.ExportJSON(failedUsageWriter{}, filter, UsageExportOptions{Now: now}); e == nil {
		t.Fatal("ignored writer failure")
	}
}

func TestUsageJSONCalendarUsesScanAnchorAndSourceDSTWindow(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip(err)
	}
	scanned := time.Date(2026, 3, 9, 12, 0, 0, 0, zone)
	index := UsageIndex{ScannedAt: scanned}
	var buf bytes.Buffer
	if err := index.ExportJSON(&buf, UsageFilter{}, UsageExportOptions{Preset: "all", Location: zone, Now: scanned.AddDate(0, 0, 10)}); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Calendar []struct {
			Start int64 `json:"start_ms"`
		} `json:"calendar"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Calendar) == 0 {
		t.Fatal("missing fixed calendar")
	}
	first, last := time.UnixMilli(out.Calendar[0].Start).In(zone), time.UnixMilli(out.Calendar[len(out.Calendar)-1].Start).In(zone)
	if first.Format("2006-01-02 15:04") != "2025-03-09 00:00" || last.Format("2006-01-02 15:04") != "2026-03-09 00:00" {
		t.Fatalf("calendar shifted to export time or included DST padding: %s .. %s", first, last)
	}
}

type failedUsageWriter struct{}

func (failedUsageWriter) Write([]byte) (int, error) { return 0, errors.New("fixture write failure") }

func TestUsageAllTimeRangeMonthRoundThenRecomputeAndExport(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	end := time.Date(2026, 3, 10, 12, 0, 0, 0, loc)
	oldest := time.Date(2026, 3, 9, 12, 0, 0, 0, loc)
	index := UsageIndex{ScannedAt: end, Records: []UsageRecord{{At: oldest, Provider: "other", Model: "m", Tokens: Tokens{Total: 1}}, {At: end.Add(-time.Hour), Provider: "selected", Model: "m", Tokens: Tokens{Total: 2}}}}
	f := UsageFilter{Start: time.UnixMilli(0), End: end, Provider: "selected"}
	start, gotEnd, g := index.UsageChartRange(f, true, loc, time.Time{})
	if start.Format("2006-01-02 15:04") != "2026-03-01 00:00" || !gotEnd.Equal(end) || g != "day" {
		t.Fatal(start, gotEnd, g)
	}
	// The rounded start predates the spring DST change; each day is civil, not
	// an invented fixed-duration 24h lattice. Filters must not move the axis.
	var b bytes.Buffer
	if err := index.ExportJSON(&b, f, UsageExportOptions{Preset: "all", Granularity: "hour", Location: loc, Now: end}); err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	points := out["series"].([]any)
	if len(points) != 10 || out["range"].(map[string]any)["start_ms"] != float64(0) {
		t.Fatal("range schema or source leading zeros changed", len(points))
	}
	for i, raw := range points {
		p := raw.(map[string]any)
		stamp := time.UnixMilli(int64(p["start_ms"].(float64))).In(loc)
		if stamp.Hour() != 0 || stamp.Day() != i+1 {
			t.Fatal("DST bucket drift", stamp)
		}
		if i < 9 && p["requests"] != float64(0) {
			t.Fatal("scoped or leading bucket fabricated traffic", p)
		}
	}
	// A range lasting more than seventy days after month rounding uses weeks.
	index.Records[0].At = time.Date(2025, 12, 31, 12, 0, 0, 0, loc)
	start, _, g = index.UsageChartRange(f, true, loc, time.Time{})
	if start.Day() != 1 || start.Month() != time.December || g != "week" {
		t.Fatal("effective granularity not recomputed", start, g)
	}
}
