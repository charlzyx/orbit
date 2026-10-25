package orbitui

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
)

func TestRuntimeUsageUsesPiSessionAndContextSeparately(t *testing.T) {
	s := agent.Snapshot{Stats: json.RawMessage(`{"tokens":{"input":50000,"output":10000,"cacheRead":40000,"cacheWrite":5000,"total":105000},"cost":0.45,"contextUsage":{"tokens":60000,"contextWindow":200000,"percent":30}}`)}
	// A visible message is not the full session (tool/summary usage is omitted).
	s.Messages = []agent.Message{{Role: "assistant", Usage: json.RawMessage(`{"input":1,"output":2,"totalTokens":3}`)}}
	got := runtimeUsage(s)
	if !got.Available || !got.TokensAvailable || !got.CostAvailable || !got.ContextAvailable {
		t.Fatalf("flags: %+v", got)
	}
	if got.Input != 50000 || got.Output != 10000 || got.CacheRead != 40000 || got.CacheWrite != 5000 || got.Cost != .45 || got.ContextPercent != 30 || got.ContextTokens != 60000 || got.ContextWindow != 200000 {
		t.Fatalf("values: %+v", got)
	}
	rate, ok := cacheReadPercent(got)
	if !ok || math.Abs(rate-40000.0/95000*100) > .00001 {
		t.Fatal(rate, ok)
	}
}
func TestRuntimeUsageMissingOrPostCompactionIsNotZero(t *testing.T) {
	for _, raw := range []string{"", `null`, `{}`, `{"tokens":null}`, `{"tokens":{"input":12}}`, `{"tokens":{"total":99}}`, `{"contextUsage":{"contextWindow":200000,"tokens":null,"percent":null}}`, `{"contextUsage":null}`, `{"cost":"0.5"}`, `{"tokens":{"input":-3,"output":1.5}}`} {
		t.Run(raw, func(t *testing.T) {
			got := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(raw), Messages: []agent.Message{{Role: "assistant", Usage: json.RawMessage(`{"input":90000,"totalTokens":90000}`)}}})
			if got.Available || contextPercentLabel(got) != "" {
				t.Fatalf("invented fallback: %+v", got)
			}
		})
	}
	zero := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(`{"tokens":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0},"cost":0}`)})
	if !zero.TokensAvailable || !zero.CostAvailable || zero.ContextAvailable {
		t.Fatal("reported zero is different from absent", zero)
	}
}
func TestRuntimeContextFallbackClampAndSmallLabel(t *testing.T) {
	tests := []struct {
		raw     string
		percent float64
		label   string
	}{
		{`{"contextUsage":{"contextWindow":200000,"tokens":60000}}`, 30, "30%"},
		{`{"contextUsage":{"contextWindow":200000,"tokens":20}}`, .01, "<1%"},
		{`{"contextUsage":{"contextWindow":200000,"tokens":0,"percent":0}}`, 0, "0%"},
		{`{"contextUsage":{"contextWindow":100,"tokens":200,"percent":200}}`, 100, "100%"},
	}
	for _, tt := range tests {
		got := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(tt.raw)})
		if !got.ContextAvailable || got.ContextPercent != tt.percent || contextPercentLabel(got) != tt.label {
			t.Fatal(got, tt)
		}
	}
	got := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(tests[0].raw), Compacting: true})
	if got.ContextAvailable {
		t.Fatal("compaction retained old context")
	}
}
func TestLatestAssistantUsageDoesNotReuseStalePreviousOutput(t *testing.T) {
	prior := agent.Message{Role: "assistant", Usage: json.RawMessage(`{"input":100,"output":25,"cacheRead":200,"cacheWrite":50,"totalTokens":375,"cost":{"total":0.02}}`)}
	s := agent.Snapshot{Messages: []agent.Message{prior, {Role: "tool"}}}
	got := latestAssistantUsage(s)
	if !got.Available || got.Output != 25 || got.Cost != .02 || got.ContextAvailable {
		t.Fatal(got)
	}
	s.Messages = append(s.Messages, agent.Message{Role: "assistant", Streaming: true})
	if latestAssistantUsage(s).Available {
		t.Fatal("previous output reused during a new response")
	}
	s.Messages[len(s.Messages)-1].Usage = json.RawMessage(`{"input":0,"output":0,"totalTokens":0,"cost":{"total":0}}`)
	if latestAssistantUsage(s).Available {
		t.Fatal("zero streaming placeholder shown")
	}
	s.Messages[len(s.Messages)-1].Usage = json.RawMessage(`{"input":4,"output":8,"cost":0}`)
	got = latestAssistantUsage(s)
	if !got.TokensAvailable || !got.CostAvailable || got.Output != 8 {
		t.Fatal(got)
	}
}
func TestUsageNumbersRejectMalformedFields(t *testing.T) {
	for _, raw := range []string{`null`, ` null `, `"1"`, `-1`, `true`, `{}`, `1e999`} {
		if _, ok := usageNumber(json.RawMessage(raw)); ok {
			t.Fatal("accepted", raw)
		}
	}
	for _, raw := range []string{`1.2`, `9223372036854775808`, `-3`} {
		if _, ok := usageInteger(json.RawMessage(raw)); ok {
			t.Fatal("accepted token counter", raw)
		}
	}
}
func quotaPtr(n float64) *float64 { return &n }
func TestQuotaHeadlinePriorityAndErrorExclusion(t *testing.T) {
	reports := []agent.QuotaReport{
		{Provider: "active", Windows: []agent.QuotaWindow{{ID: "weekly", Label: "Weekly", UsedPercent: quotaPtr(95)}, {ID: "five_hour", Label: "5-hour", UsedPercent: quotaPtr(20)}}},
		{Provider: "other", Windows: []agent.QuotaWindow{{ID: "5h", Label: "5-hour", UsedPercent: quotaPtr(80)}}},
		{Provider: "error", Error: "expired", Windows: []agent.QuotaWindow{{ID: "5h", UsedPercent: quotaPtr(99)}}},
	}
	got := quotaSummary(reports, "active")
	if got.Provider != "active" || got.Window != "5h" || got.Text != "20%" {
		t.Fatal(got)
	}
	got = quotaSummary(reports, "unknown")
	if got.Provider != "other" || got.Text != "80%" {
		t.Fatal(got)
	}
	reports[0].Windows = reports[0].Windows[:1]
	got = quotaSummary(reports, "active")
	if got.Provider != "active" || got.Text != "95%" {
		t.Fatal(got)
	}
	reports[0].Windows = nil
	reports[0].Balances = []agent.QuotaBalance{{Amount: 110, Currency: "CNY"}}
	got = quotaSummary(reports, "active")
	if got.Text != "110 CNY" || got.Meter {
		t.Fatal(got)
	}
	reports[1].Error = "unavailable"
	got = quotaSummary(reports, "unknown")
	if got.Text != "Usage" || got.Meter {
		t.Fatal(got)
	}
}
func TestQuotaFractionsAndWindowRecognition(t *testing.T) {
	for _, label := range []string{"five_hour", "Rolling 5h", "5-hour session"} {
		if !fiveHourWindow(agent.QuotaWindow{Label: label}) {
			t.Fatal(label)
		}
	}
	if fiveHourWindow(agent.QuotaWindow{ID: "primary", Label: "Weekly"}) {
		t.Fatal("invented five-hour meaning for unlabeled primary window")
	}
	for _, tt := range []struct {
		w    agent.QuotaWindow
		want float64
		ok   bool
	}{
		{agent.QuotaWindow{UsedPercent: quotaPtr(150)}, 1, true},
		{agent.QuotaWindow{UsedPercent: quotaPtr(0)}, 0, true},
		{agent.QuotaWindow{Used: quotaPtr(20), Limit: quotaPtr(80)}, .25, true},
		{agent.QuotaWindow{Used: quotaPtr(20), Limit: quotaPtr(0)}, 0, false},
		{agent.QuotaWindow{UsedPercent: quotaPtr(math.NaN())}, 0, false},
	} {
		fraction, ok := quotaFraction(tt.w)
		if fraction != tt.want || ok != tt.ok {
			t.Fatal(fraction, ok, tt)
		}
	}
}
func TestQuotaResetHintAndAmount(t *testing.T) {
	now := time.Date(2026, 10, 25, 1, 0, 0, 0, time.UTC)
	tests := []struct {
		after time.Duration
		want  string
	}{{30 * time.Second, "resets in <1m"}, {42 * time.Minute, "resets in 42m"}, {2 * time.Hour, "resets in 2h"}, {3*time.Hour + 12*time.Minute, "resets in 3h 12m"}, {3 * 24 * time.Hour, "resets in 3d"}, {2*24*time.Hour + 4*time.Hour + 5*time.Minute, "resets in 2d 4h"}, {-time.Hour, "resets at Oct 25, 00:00"}, {8 * 24 * time.Hour, "resets at Nov 2, 01:00"}}
	for _, tt := range tests {
		if got := quotaResetHint(now.Add(tt.after).UnixMilli(), now); got != tt.want {
			t.Fatal(got, tt)
		}
	}
	if quotaAmount(110) != "110" || quotaAmount(1.234) != "1.23" {
		t.Fatal("amount formatting")
	}
}
func TestQuotaReportsAreSortedWithoutChangingInput(t *testing.T) {
	reports := []agent.QuotaReport{{Provider: "z", Note: "Not supported"}, {Provider: "empty"}, {Provider: "a", Error: "Unavailable"}}
	got := quotaReports(reports)
	if len(got) != 2 || got[0].Provider != "a" || reports[0].Provider != "z" {
		t.Fatal(got, reports)
	}
}
