package mcpui

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestQuotaPendingIntentDoesNotCrossPiBinaryChange(t *testing.T) {
	s, _ := popupState(t)
	s.start("held old binary load", func(ctx context.Context) (func(*State), error) { <-ctx.Done(); return nil, ctx.Err() })
	s.OpenQuotaPopup(time.Now())
	if !s.quotaProbeRequested {
		t.Fatal("fixture did not queue popup intent")
	}
	marker := filepath.Join(t.TempDir(), "new-binary-probed")
	s.SetPiBinary(fakePi(t, "touch '"+marker+"'\nprintf '%s\\n' '{\"servers\":[]}'"))
	s.PrepareQuotaPopup()
	waitState(t, s)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("old popup intent launched replacement binary", err)
	}
	if s.quotaProbeRequested || !s.checkedAt.IsZero() {
		t.Fatal("old intent/status survived binary change")
	}
}
