package orbitui

import (
	"encoding/json"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestContextMeterMountedUnknown(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find(a.Tr("context_meter.unknown")); !ok {
		t.Fatal("missing unknown context details", tt.Texts())
	}
}

func TestContextMeterNumbersAndObservedSlices(t *testing.T) {
	for n, want := range map[int64]string{688: "688", 1000: "1K", 1900: "1.9K", 100050: "100.1K", 999999: "1000K", 1000001: "1M", 1050000: "1.1M"} {
		if got := contextTokens(n); got != want {
			t.Fatal(n, got, want)
		}
	}
	for n, want := range map[float64]string{0: "$0.00", .001: "<$0.01", 1.23: "$1.23", 123.45: "$123"} {
		if got := contextCost(n); got != want {
			t.Fatal(n, got, want)
		}
	}
	if contextHitPercent(99.96) != "99.9" || contextHitPercent(100) != "100" {
		t.Fatal("false perfect cache rate")
	}
	a := fixture(t)
	a.Model.Messages = []Message{{ID: "u", Role: "user", Text: "你好世界"}, {ID: "a", Role: "assistant", Text: "abcd"}}
	if a.conversationEstimate() != 2 {
		t.Fatal("chars/4 must count Unicode scalars")
	}
}

func TestContextMeterMountedDetailsAndDismissal(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "one", Usage: Usage{ContextAvailable: true, ContextTokensAvailable: true, ContextTokens: 800, ContextWindow: 1000, ContextPercent: 80, SessionAvailable: true, TokensAvailable: true, Input: 10, Output: 20, CacheRead: 30, CacheWrite: 40, Total: 100, TotalAvailable: true, Cost: .001, CostAvailable: true}, Messages: []Message{{ID: "u", Role: "user", Text: "abcdefgh"}}})
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "RefreshContextStats" {
		t.Fatal(actions)
	}
	for _, text := range []string{"80% Full", "~800 / 1K tokens", "Other context", "Conversation", "Session usage", "<$0.01", "Compact now"} {
		if _, ok := tt.Find(text); !ok {
			t.Fatal("missing", text, tt.Texts())
		}
	}
	panel, ok := tt.Find("Context details")
	if !ok {
		t.Fatal("missing panel")
	}
	if panel.W != 280 {
		t.Fatal("source panel width", panel)
	}
	if dir := os.Getenv("ORBIT_CONTEXT_CAPTURE_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, "context-details-headless.png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, tt.Image())
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tt.Click("Compact context"); err != nil {
		t.Fatal(err)
	}
	if actions[len(actions)-1].Kind != "CompactContext" {
		t.Fatal(actions)
	}
	a.Model.Compacting = true
	tt.Frame()
	if err := tt.Click("Compact context"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 2 {
		t.Fatal("compacting allowed duplicate", actions)
	}
	if _, ok := tt.Find("Compacting…"); !ok {
		t.Fatal(tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	if _, ok := tt.Find("Context details"); ok {
		t.Fatal("Escape left popup open")
	}
	tt.Advance(201000000)
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Context details"); !ok {
		t.Fatal("reopen missing")
	}
	a.SetSnapshot(Snapshot{SessionID: "two"})
	tt.Frame()
	if _, ok := tt.Find("Context details"); ok {
		t.Fatal("popup crossed session switch")
	}
	tt.Advance(201000000)
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Compact context"); ok {
		t.Fatal("disconnected offered compact")
	}
	if _, ok := tt.Find("Unknown"); !ok {
		t.Fatal("unknown displayed as zero")
	}
	if err := tt.Click("Close context details"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Context details"); ok {
		t.Fatal("close button left popup open")
	}
}

func TestContextMeterHoverGeometryAndThresholdColors(t *testing.T) {
	a := fixture(t)
	a.Model.Usage = Usage{ContextAvailable: true, ContextTokensAvailable: true, ContextTokens: 1, ContextWindow: 1000, ContextPercent: .1}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	ring, ok := tt.Find("Context ring")
	if !ok || ring.W != 16 || ring.H != 16 {
		t.Fatal("ring source geometry", ring)
	}
	anchor, _ := tt.Find("Context usage")
	tt.Move(anchor.X+anchor.W/2, anchor.Y+anchor.H/2)
	card, ok := tt.Find("Context summary")
	if !ok || card.W != 200 {
		t.Fatal("hover source width", card, tt.Texts())
	}
	if !tt.HasText("<1% context used") {
		t.Fatal("small nonzero usage lost", tt.Texts())
	}
	tt.Move(20, 20)
	if _, ok := tt.Find("Context summary"); ok {
		t.Fatal("hover did not close")
	}
	for _, tc := range []struct {
		percent float64
		color   ui.Color
	}{{0, a.Theme.Accent}, {69.9, a.Theme.Accent}, {70, a.Theme.Warning}, {90, a.Theme.Error}} {
		if got := a.contextFill(Usage{ContextAvailable: true, ContextPercent: tc.percent}); got != tc.color {
			t.Fatal(tc.percent, got)
		}
	}
}

func TestContextMeterPartialAndMalformedBucketsRemainUnknown(t *testing.T) {
	for _, raw := range []string{`{"tokens":{"total":37}}`, `{"tokens":{"input":-1,"output":"12","cacheRead":null,"cacheWrite":1.5}}`, `{"tokens":{}}`} {
		u := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(raw)})
		if u.InputAvailable || u.OutputAvailable || u.CacheReadAvailable || u.CacheWriteAvailable || u.TokensAvailable {
			t.Fatal("invented bucket availability", raw, u)
		}
		if raw == `{"tokens":{"total":37}}` {
			if !u.TotalAvailable || u.Total != 37 {
				t.Fatal("lost explicit total", u)
			}
		} else if u.TotalAvailable {
			t.Fatal("invented fallback total", u)
		}
		a := fixture(t)
		a.Model.Usage = u
		tt := ui.NewTester(a.View, 1470, 922)
		tt.Frame()
		if err := tt.Click("Context usage"); err != nil {
			t.Fatal(err)
		}
		if tt.HasText("0") || tt.HasText("0 · 0%") {
			t.Fatal("partial data masqueraded as zero", raw, tt.Texts())
		}
		if !tt.HasText("—") {
			t.Fatal("missing unknown markers", tt.Texts())
		}
	}
	u := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(`{"tokens":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0}}`)})
	if !u.TokensAvailable || !u.TotalAvailable || u.Total != 0 {
		t.Fatal("reported zero must stay known", u)
	}
}

func TestContextMeterOverflowTotalIsUnknown(t *testing.T) {
	u := runtimeUsage(agent.Snapshot{Stats: json.RawMessage(`{"tokens":{"input":9000000000000000000,"output":9000000000000000000,"cacheRead":0,"cacheWrite":0}}`)})
	if u.TotalAvailable || contextTokenValue(u.Total, u.TotalAvailable) != "—" {
		t.Fatal("overflow invented total", u)
	}
}

func TestContextMeterHydrationDisablesCompactUntilIdentity(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{Connected: true})
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Compact context"); err != nil {
		t.Fatal(err)
	}
	for _, v := range actions {
		if v.Kind == "CompactContext" {
			t.Fatal("unhydrated action sent")
		}
	}
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "ready"})
	tt.Frame()
	tt.Advance(201000000)
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Compact context"); err != nil {
		t.Fatal(err)
	}
	if actions[len(actions)-1].Kind != "CompactContext" {
		t.Fatal("hydrated compact remained disabled", actions)
	}
}
