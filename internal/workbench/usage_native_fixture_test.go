package workbench

import (
	"context"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUsageNativeCaptureFixtureRealStore(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "usage-native", "home", ".pi", "agent", "sessions")
	scanner := NewUsageScanner(root)
	for repeat := 0; repeat < 2; repeat++ {
		index, err := scanner.Scan(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		f := UsageFilter{Start: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)}
		s := index.Summary(f)
		if index.Files != 4 || index.UnreadableFiles != 0 || len(index.Sessions) != 4 || len(index.Errors) != 3 || s.Requests != 33 || s.Turns != 33 || s.ToolRuns != 4 || s.Errors != 1 || s.Aborted != 1 || s.PricedRequests != 33 || s.TimedRequests != 33 || s.Tokens.Input != 44400 || s.Tokens.Output != 4960 || s.Tokens.CacheRead != 4800 || s.Tokens.CacheWrite != 1200 || s.Tokens.Total != 55360 || math.Abs(s.Cost-.912) > 1e-10 {
			t.Fatalf("capture fixture drift repeat=%d summary=%+v sessions=%d errors=%d", repeat, s, len(index.Sessions), len(index.Errors))
		}
		for _, session := range index.Sessions {
			if !strings.HasPrefix(session.Title, "Usage fixture task ") || session.Started.After(session.Ended) {
				t.Fatal("invalid chronological capture session", session)
			}
		}
		alpha := f
		alpha.ModelRefs = []UsageModelRef{{Provider: "fixture-alpha", Model: "shared-model"}}
		beta := f
		beta.ModelRefs = []UsageModelRef{{Provider: "fixture-beta", Model: "shared-model"}}
		if index.Summary(alpha).Requests != 9 || index.Summary(beta).Requests != 8 {
			t.Fatal("duplicate model ids lost provider scope")
		}
	}
}
