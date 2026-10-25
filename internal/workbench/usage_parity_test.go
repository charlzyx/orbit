package workbench

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestUsageExactUnsignedAndNonfiniteCounts(t *testing.T) {
	tests := []struct {
		v     any
		want  uint64
		valid bool
	}{{json.Number("18446744073709551615"), math.MaxUint64, true}, {json.Number("18446744073709551616"), 0, false}, {json.Number("1.5"), 0, false}, {json.Number("-1"), 0, false}, {json.Number("7"), 7, true}, {float64(8), 8, true}, {float64(8.5), 0, false}, {math.NaN(), 0, false}, {math.Inf(1), 0, false}, {math.Exp2(64), 0, false}, {float64(-3), 0, false}}
	for _, tt := range tests {
		got, valid := unsignedCount(tt.v)
		if got != tt.want || valid != tt.valid {
			t.Errorf("%v: %d %t", tt.v, got, valid)
		}
	}
	if addCounts(math.MaxUint64, 1) != math.MaxUint64 {
		t.Fatal("token total wrapped")
	}
	for _, v := range []any{math.NaN(), math.Inf(1), math.Exp2(63), 1.5} {
		if !ts(v).IsZero() {
			t.Fatalf("bad timestamp accepted %v", v)
		}
	}
	for _, v := range []any{math.NaN(), math.Inf(1), float64(-1), "4"} {
		if _, ok := usageCost(v); ok {
			t.Fatalf("bad cost accepted %v", v)
		}
	}
}
func TestUsagePricingAcrossFilesAndFilters(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "group", "one.jsonl"), `{"type":"session","id":"s","cwd":"/work"}
{"type":"message","timestamp":"2026-10-24T01:00:00Z","message":{"role":"assistant","provider":"p","model":"m","usage":{"input":2,"cost":{"total":1}}}}
{"type":"message","timestamp":"2026-10-25T01:00:00Z","message":{"role":"assistant","provider":"p","model":"m","usage":{"input":2,"cost":{"total":0}}}}`)
	put(t, filepath.Join(root, "group", "two.jsonl"), `{"type":"session","id":"s2","cwd":"/work"}
{"type":"message","timestamp":"2026-10-25T02:00:00Z","message":{"role":"assistant","provider":"p","model":"m","usage":{"input":3}}}
{"type":"message","timestamp":"2026-10-25T03:00:00Z","message":{"role":"assistant","provider":"other","model":"m","usage":{"input":3,"cost":0}}}`)
	idx, e := NewUsageScanner(root).Scan(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	f := UsageFilter{Start: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)}
	summary := idx.Summary(f)
	if summary.Requests != 3 || summary.PricedRequests != 2 {
		t.Fatalf("%+v", summary)
	}
	records := idx.FilteredRecords(f)
	if !records[0].HasCost || !records[0].ModelPriced || records[1].HasCost || !records[1].ModelPriced || records[2].ModelPriced {
		t.Fatal(records)
	}
}
func TestUsageCallsFailuresSourceFiltersAndSessionPaths(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "group", "session.jsonl")
	put(t, path, `{"type":"session","id":"s","cwd":"/w","timestamp":"2026-10-25T01:00:00Z"}
{"type":"message","timestamp":"2026-10-25T01:00:01Z","message":{"role":"user","content":"hello"}}
{"type":"message","timestamp":"2026-10-25T01:00:02Z","message":{"role":"assistant","provider":"p","model":"a","usage":{"input":1},"content":[{"type":"toolCall","id":"same","name":"bash"},{"type":"toolCall","id":"same","name":"read"},{"type":"toolCall","id":"pending","name":"write"}]}}
{"type":"message","timestamp":"2026-10-25T01:00:03Z","message":{"role":"assistant","provider":"p","model":"b","stopReason":"error","errorMessage":"provider failed","usage":{"input":0,"output":0,"cost":{"total":0}}}}
{"type":"message","timestamp":"2026-10-25T01:00:04Z","message":{"role":"toolResult","toolCallId":"same","toolName":"wrong","isError":true}}
{"type":"message","timestamp":"2026-10-25T01:00:05Z","message":{"role":"toolResult","toolCallId":"same"}}
{"type":"message","timestamp":"2026-10-25T01:00:06Z","message":{"role":"toolResult","toolCallId":"unmatched","toolName":"ls"}}`)
	idx, e := NewUsageScanner(root).Scan(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if len(idx.Records) != 2 || len(idx.Tools) != 4 || len(idx.Turns) != 1 {
		t.Fatal(idx)
	}
	f := UsageFilter{Provider: "p", Model: "a"}
	if summary := idx.Summary(f); summary.ToolRuns != 3 || summary.Turns != 1 || summary.Requests != 1 {
		t.Fatalf("%+v", summary)
	}
	if len(idx.FilteredTools(UsageFilter{Model: "b"})) != 1 {
		t.Fatal("unmatched tool did not use most recent request model")
	}
	tools := idx.FilteredTools(f)
	if tools[0].Name != "bash" || tools[0].Success || tools[0].Duration != 2*time.Second || tools[0].At.Second() != 2 || tools[1].Name != "read" || !tools[1].HasResult || tools[2].HasResult {
		t.Fatal(tools)
	}
	if len(idx.FilteredTurns(UsageFilter{Model: "not used"})) != 1 {
		t.Fatal("prompt incorrectly scoped to model")
	}
	for _, f := range []UsageFilter{{ErrorsOnly: true}, {CachedOnly: true}} {
		if len(idx.FilteredTurns(f)) != 0 || len(idx.FilteredTools(f)) != 0 {
			t.Fatal("request-only filters kept tools/prompts")
		}
	}
	if summary := idx.Summary(UsageFilter{ErrorsOnly: true}); summary.Requests != 1 || summary.Errors != 1 {
		t.Fatal(summary)
	}
	for _, r := range idx.Records {
		if r.SessionPath != path {
			t.Fatal("lost session file")
		}
	}
	for _, r := range idx.Tools {
		if r.SessionPath != path {
			t.Fatal("lost tool session file")
		}
	}
	if idx.Turns[0].SessionPath != path {
		t.Fatal("lost turn session file")
	}
}
func TestUsageHeaderDurationAndExactJSONIntegers(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "group", "x.jsonl")
	put(t, path, `{"type":"session","id":"s"}
{"type":"message","timestamp":"2026-10-25T01:00:02Z","message":{"role":"assistant","timestamp":1792890001000,"usage":{"input":18446744073709551615,"output":1,"totalTokens":1.5,"cost":-1}}}`)
	idx, e := readUsageFile(context.Background(), path)
	if e != nil {
		t.Fatal(e)
	}
	if len(idx.Records) != 1 || idx.Records[0].Tokens.Input != math.MaxUint64 || idx.Records[0].Tokens.Total != math.MaxUint64 || idx.Records[0].HasCost {
		t.Fatal(idx)
	}
	put(t, path, `{"type":"message","message":{"role":"assistant","usage":{"input":1}}}`)
	if _, e := readUsageFile(context.Background(), path); e == nil {
		t.Fatal("accepted headerless file")
	}
}
func TestUsageConcurrentScansAndCallerMutation(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "group", "s.jsonl"), usageFixture())
	scanner := NewUsageScanner(root)
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			idx, e := scanner.Scan(context.Background())
			if e != nil {
				t.Error(e)
				return
			}
			idx.Records[0].Tokens.Total = 999
		}()
	}
	wg.Wait()
	idx, e := scanner.Scan(context.Background())
	if e != nil || idx.Records[0].Tokens.Total != 100 {
		t.Fatal(idx, e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := scanner.Scan(ctx); e == nil {
		t.Fatal("scan ignored cancellation")
	}
	if e := os.Remove(filepath.Join(root, "group", "s.jsonl")); e != nil {
		t.Fatal(e)
	}
}

func TestUsageScannerSourceGroupDepthAndLinkedGroups(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "ignored-root.jsonl"), usageFixture())
	put(t, filepath.Join(root, "group", "kept.jsonl"), usageFixture())
	put(t, filepath.Join(root, "group", "nested", "ignored-nested.jsonl"), usageFixture())
	linked := t.TempDir()
	put(t, filepath.Join(linked, "linked.jsonl"), usageFixture())
	if err := os.Symlink(linked, filepath.Join(root, "linked-group")); err != nil {
		t.Skip(err)
	}
	scanner := NewUsageScanner(root)
	index, err := scanner.Scan(context.Background())
	if err != nil || index.Files != 2 || len(index.Sessions) != 2 {
		t.Fatalf("only direct files in ordinary or linked workspace groups should scan: %+v, %v", index, err)
	}
	for _, session := range index.Sessions {
		if strings.Contains(session.Path, "ignored") {
			t.Fatalf("non-source session depth included: %s", session.Path)
		}
	}
	if err := os.Remove(filepath.Join(root, "group", "kept.jsonl")); err != nil {
		t.Fatal(err)
	}
	index, err = scanner.Scan(context.Background())
	if err != nil || index.Files != 1 || len(index.Sessions) != 1 || index.Parsed != 0 {
		t.Fatalf("removed group file survived or linked cache was lost: %+v, %v", index, err)
	}
}

func TestUsageRollingRangesMatchSourceDSTArithmetic(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip(err)
	}
	// Rust DateRange::for_preset subtracts fixed 86,400,000 ms, not
	// calendar days, even when yesterday crosses the spring/fall change.
	for _, now := range []time.Time{
		time.Date(2026, 3, 9, 12, 0, 0, 0, zone),
		time.Date(2026, 11, 2, 12, 0, 0, 0, zone),
	} {
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, zone)
		for preset, days := range map[string]int{"yesterday": 1, "last7": 6, "last14": 13, "last30": 29} {
			start, end := UsageRange(preset, now)
			wantEnd := now
			if preset == "yesterday" {
				wantEnd = midnight
			}
			if !start.Equal(midnight.Add(-time.Duration(days)*24*time.Hour)) || !end.Equal(wantEnd) {
				t.Fatalf("%s at %s did not match source's epoch arithmetic: %s..%s", preset, now, start, end)
			}
		}
	}
}

func TestUsageExplicitErrorsTitlesAndExactSessionScope(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "group", "title.jsonl")
	put(t, path, `{"type":"session","id":"same-id","cwd":"/w"}
{"type":"message","timestamp":"2026-10-25T01:00:00Z","message":{"role":"user","content":[{"type":"text","text":"<skill name=\"review\" location=\"/private\">Instructions"}]}}
{"type":"message","timestamp":"2026-10-25T01:00:01Z","message":{"role":"assistant","provider":"p","model":"m","errorMessage":"no usage provider failure"}}
{"type":"message","timestamp":"2026-10-25T01:00:02Z","message":{"role":"assistant","provider":"p","model":"m","usage":{"input":2},"content":[{"type":"toolCall","id":"t","name":"bash"}]}}
{"type":"message","timestamp":"2026-10-25T01:00:03Z","message":{"role":"toolResult","toolCallId":"t","isError":true}}`)
	idx, e := NewUsageScanner(root).Scan(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if len(idx.Records) != 1 || idx.Records[0].SessionTitle != "Skill: review" || idx.Tools[0].SessionTitle != "Skill: review" || idx.Turns[0].SessionTitle != "Skill: review" {
		t.Fatal(idx)
	}
	errors := idx.FilteredErrors(UsageFilter{Provider: "p", CachedOnly: true, ErrorsOnly: true, SessionPath: path})
	if len(errors) != 2 || errors[0].Kind != "tool" || errors[0].At.Second() != 3 || errors[1].Kind != "provider" || errors[1].Message != "no usage provider failure" {
		t.Fatal(errors)
	}
	if len(idx.FilteredErrors(UsageFilter{SessionPath: "elsewhere"})) != 0 || len(idx.FilteredTurns(UsageFilter{SessionPath: "elsewhere"})) != 0 {
		t.Fatal("file scope ignored")
	}
	if len([]rune(usageTextCap("  "+string(make([]rune, 170)), 160))) != 161 {
		t.Fatal("unicode cap")
	}
}

func TestUsageCSVSourceColumnsUnknownCostAndMeasuredZero(t *testing.T) {
	index := UsageIndex{Records: []UsageRecord{{At: time.Unix(0, 0), Session: "s", SessionTitle: "=formula", Provider: "p", Model: "m", Tokens: Tokens{Reasoning: 7}, Outcome: "toolUse"}, {At: time.Unix(1, 0), Session: "s", HasCost: true, HasDuration: true, Duration: 2 * time.Second}}}
	var out bytes.Buffer
	if e := index.ExportCSV(&out, UsageFilter{}); e != nil {
		t.Fatal(e)
	}
	rows, e := csv.NewReader(&out).ReadAll()
	if e != nil {
		t.Fatal(e)
	}
	if len(rows[0]) != 15 || rows[0][1] != "session_id" || rows[0][11] != "reasoning" || rows[1][2] != "'=formula" || rows[1][12] != "" || rows[1][14] != "tool_use" || rows[2][12] != "0.000000" || rows[2][13] != "2000" {
		t.Fatal(rows)
	}
}
