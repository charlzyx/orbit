package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
	"time"
)

func TestQuotaContextMountedMutualExclusion(t *testing.T) {
	a := fixture(t)
	state := quotaMCPFixture(t, a, true)
	percent := 25.0
	a.Model.Quotas = []agent.QuotaReport{{Provider: "fixture", Windows: []agent.QuotaWindow{{Label: "usage", UsedPercent: &percent}}}}
	summary := quotaSummary(a.Model.Quotas, "")
	quotaLabel := strings.TrimSpace(strings.Join([]string{summary.Provider, summary.Window, summary.Text}, " "))
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	if !a.contextOpen {
		t.Fatal("context did not open")
	}
	if err := tt.Click(quotaLabel); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.quotaOpen || a.contextOpen {
		t.Fatalf("context to quota: context=%v quota=%v", a.contextOpen, a.quotaOpen)
	}
	if err := tt.Click(a.text("session.tab_mcp", "MCP")); err != nil {
		t.Fatal(err)
	}
	waitQuotaMCP(t, state)
	tt.Advance(250 * time.Millisecond)
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.contextOpen || a.quotaOpen {
		t.Fatalf("quota to context: context=%v quota=%v", a.contextOpen, a.quotaOpen)
	}
}
