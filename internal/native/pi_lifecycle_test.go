package native

import (
	"context"
	"testing"
)

func TestPiUpdateMaintenanceGuard(t *testing.T) {
	for _, tt := range []struct {
		name                  string
		ok, release, cancel   bool
		wantWork, wantRelease bool
	}{{"accepted", true, true, false, true, true}, {"rejected", false, false, false, false, false}, {"rejected cleanup", false, true, false, false, true}, {"invalid no release", true, false, false, false, false}, {"cancelled after acquisition", true, true, true, false, true}} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			worked, released := false, 0
			got := runWithPiUpdateGuard(ctx, func() (func(), bool) {
				if tt.cancel {
					cancel()
				}
				var release func()
				if tt.release {
					release = func() { released++ }
				}
				return release, tt.ok
			}, func() {
				worked = true
				if released != 0 {
					t.Fatal("guard released before update")
				}
			})
			if got != tt.wantWork || worked != tt.wantWork || (released == 1) != tt.wantRelease || released > 1 {
				t.Fatal(got, worked, released)
			}
		})
	}
	if runWithPiUpdateGuard(context.Background(), nil, func() { t.Fatal("ran without guard") }) {
		t.Fatal("missing guard accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if runWithPiUpdateGuard(ctx, func() (func(), bool) { t.Fatal("cancelled task acquired guard"); return nil, false }, func() { t.Fatal("ran cancelled update") }) {
		t.Fatal("cancelled update accepted")
	}
}
