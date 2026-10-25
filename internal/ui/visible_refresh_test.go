package orbitui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestVisibleRefreshActualShortcutAndPaletteRoute(t *testing.T) {
	for _, route := range []string{"shortcut", "palette"} {
		t.Run(route, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			a := x.UI
			calls := 0
			a.Options.RefreshMCP = func() { calls++ }
			a.Options.MCP = func(c *ui.Context) { ui.Text(c, "fixture MCP visible") }
			a.page = "settings"
			a.settingsPage = "mcp"
			tt := ui.NewTester(a.View, 1222, 878)
			tt.Frame()
			before := calls
			if route == "shortcut" {
				tt.Key(ui.Cmd, ui.KeyR)
			} else {
				a.paletteOpen = true
				a.paletteQuery = "refresh reload sessions worktrees branch git explorer files usage providers mcp disk"
				tt.Frame()
				if err := tt.Click(a.Tr("command_palette.refresh_sessions")); err != nil {
					t.Fatal(err)
				}
			}
			tt.Frame()
			if calls != before+1 {
				t.Fatal("visible MCP not refreshed by actual command route", calls, before)
			}
		})
	}
}
func TestVisibleRefreshHonorsActualPageNotRetainedDockFlag(t *testing.T) {
	for _, tc := range []struct {
		page, section string
		blocked       bool
		want          string
	}{
		{"chat", "", false, "review"}, {"git", "", false, "review"}, {"worktrees", "", false, "review"},
		{"usage", "", false, "usage"}, {"settings", "mcp", false, "mcp"}, {"settings", "plugins", false, "plugins"},
		{"settings", "providers", false, "none"}, {"settings", "mcp", true, "mcp"}, {"chat", "", true, "none"}, {"usage", "", true, "none"},
	} {
		t.Run(tc.page+"/"+tc.section+map[bool]string{true: " blocked"}[tc.blocked], func(t *testing.T) {
			x := workspaceDialogFixture(t)
			a := x.UI
			a.page = tc.page
			a.settingsPage = tc.section
			a.setup.enabled = tc.blocked
			a.setup.loaded = !tc.blocked
			a.ReviewVisible = true
			a.reviewState.loading = true
			a.reviewState.pending = false
			a.usageState.loading = true
			a.usageState.pending = false
			mcp, plugins := 0, 0
			a.Options.RefreshMCP = func() { mcp++ }
			a.Options.RefreshResources = func(page string) {
				if page != "plugins" {
					t.Fatal("unexpected resource refresh", page)
				}
				plugins++
			}
			before := a.reviewState.generation
			x.Action(Action{Kind: "RefreshSessions"})
			if (a.reviewState.generation > before) != (tc.want == "review") || a.reviewState.pending != (tc.want == "review") || a.usageState.pending != (tc.want == "usage") || (mcp == 1) != (tc.want == "mcp") || plugins != 0 || x.Resources.Busy != (tc.want == "plugins") {
				t.Fatalf("wrong visible request: want=%s review=%d->%d pending=%v usage=%v mcp=%d update-check-hook=%d snapshot-busy=%v", tc.want, before, a.reviewState.generation, a.reviewState.pending, a.usageState.pending, mcp, plugins, x.Resources.Busy)
			}
		})
	}
}

func TestVisibleRefreshUsageCoalescesAndFailureKeepsGlobalLastGood(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	a.page = "usage"
	a.setup.enabled = false
	s := &a.usageState
	s.index = usageFixtureIndex()
	s.loaded = true
	s.updated = time.Now()
	before := s.index
	updated := s.updated
	revision := s.revision
	var calls atomic.Int32
	release := make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	s.scan = func(context.Context) (workbench.UsageIndex, error) {
		calls.Add(1)
		<-release
		return workbench.UsageIndex{}, errors.New("synthetic usage scan failure")
	}
	x.Action(Action{Kind: "RefreshSessions"})
	x.Action(Action{Kind: "RefreshSessions"})
	x.Action(Action{Kind: "RefreshSessions"})
	if !s.loading || !s.pending {
		t.Fatal("manual refresh bypassed existing coalescing")
	}
	a.page = "chat"
	second := t.TempDir()
	x.changeWorkspaceServices(second)
	close(release)
	drainRegressionUpdates(t, a, func() bool { return !s.loading })
	if calls.Load() != 2 || s.err != "synthetic usage scan failure" || len(s.index.Records) != len(before.Records) || !s.updated.Equal(updated) || s.revision != revision || a.page != "chat" || a.Model.Workspace != second {
		t.Fatal("failed/global refresh changed last good or newer page owner", calls.Load(), s.err, s.revision, a.page, a.Model.Workspace)
	}
}
func TestVisibleRefreshReviewRejectsPreviousWorkspaceCompletion(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	a.setup.enabled = false
	a.page = "chat"
	first, second := regressionRepo(t), regressionRepo(t)
	if err := os.WriteFile(filepath.Join(first, "first-change"), []byte("first"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(second, "second-change"), []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	x.changeWorkspaceServices(first)
	a.ReviewVisible = true
	x.Action(Action{Kind: "RefreshSessions"})
	x.changeWorkspaceServices(second)
	x.Action(Action{Kind: "RefreshSessions"})
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading && a.reviewState.root == second })
	if a.reviewState.err != "" || len(a.reviewState.data.Files) != 1 || a.reviewState.data.Files[0].Path != "second-change" {
		t.Fatal("manual refresh adopted stale review owner", a.reviewState.err, a.reviewState.data.Files)
	}
}
