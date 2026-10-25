package gitworkbench

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// These tests mount the production view and inspect laid-out rectangles. The
// two-commit case reproduces WB07's centered history block; merely finding the
// commit text would not catch that regression. They are CPU layout tests, not
// a replacement for a same-state native screenshot comparison.
func historyLayoutFixture(t *testing.T, graph bool, count, height int) (*State, *ui.Tester) {
	t.Helper()
	s := New(t.TempDir(), Options{})
	t.Cleanup(s.Close)
	s.initialized, s.refreshed = true, time.Now()
	s.Tab = 1
	if graph {
		s.Tab = 2
	}
	s.Data.Branch = "capture-main"
	for i := range count {
		entry := workspace.CommitEntry{
			Hash: fmt.Sprintf("%040x", i+1), Short: fmt.Sprintf("%07x", i+1),
			Subject: fmt.Sprintf("Fixture commit %03d", i), Author: "Orbit Fixture", Relative: "3 weeks ago",
		}
		s.Data.History = append(s.Data.History, entry)
		s.Data.Graph = append(s.Data.Graph, workspace.GraphRow{Commit: workspace.GraphCommit{CommitEntry: entry}, LaneCount: 1})
	}
	tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1222, height)
	return s, tt
}

func historyRect(t *testing.T, tt *ui.Tester, label string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("missing %q in mounted history", label)
	}
	return r
}

func historyFilterRect(t *testing.T, tt *ui.Tester, graph bool) ui.Rect {
	t.Helper()
	if graph {
		return historyRect(t, tt, "All refs")
	}
	return historyRect(t, tt, "Commit search")
}

func TestHistoryMountedShortListStartsBelowFilters(t *testing.T) {
	for _, graph := range []bool{false, true} {
		for _, height := range []int{600, 878} {
			t.Run(fmt.Sprintf("graph=%t/height=%d", graph, height), func(t *testing.T) {
				_, tt := historyLayoutFixture(t, graph, 2, height)
				filter := historyFilterRect(t, tt, graph)
				first := historyRect(t, tt, "Fixture commit 000")
				second := historyRect(t, tt, "Fixture commit 001")
				more := historyRect(t, tt, "Load more")
				t.Logf("filter bottom=%g first=%v second=%v Load more text=%v", filter.Y+filter.H, first, second, more)
				if gap := first.Y - (filter.Y + filter.H); gap < 0 || gap > 12 {
					t.Errorf("first commit must start immediately below filters, not center in the viewport: gap=%g filter=%v first=%v", gap, filter, first)
				}
				if second.Y != first.Y+first.H || first.H != 56 || second.H != 56 {
					t.Errorf("commit rows lost their contiguous 56px layout: first=%v second=%v", first, second)
				}
				if gap := more.Y - (second.Y + second.H); gap < 0 || gap > 12 {
					t.Errorf("Load more must follow the last commit, not sit at the viewport bottom: second=%v more=%v", second, more)
				}
			})
		}
	}
}

func TestHistoryMountedEmptyAndLoadingFollowSource(t *testing.T) {
	for _, graph := range []bool{false, true} {
		t.Run(fmt.Sprintf("graph=%t", graph), func(t *testing.T) {
			s, tt := historyLayoutFixture(t, graph, 0, 878)
			filter := historyFilterRect(t, tt, graph)
			title := historyRect(t, tt, "No commits yet")
			detail := "Commits on this branch will show up here."
			if graph {
				detail = "Commits across your local branches will show up here."
			}
			historyRect(t, tt, detail)
			// empty_note deliberately centers its message in the source. Do not
			// apply the nonempty list's top anchoring to this different state.
			if title.Y <= filter.Y+filter.H+100 || title.Y+title.H >= 878 {
				t.Errorf("empty note should occupy the body, not a commit row: filter=%v title=%v", filter, title)
			}
			assertNoMore := func() {
				t.Helper()
				if _, ok := tt.Find("Load more"); ok {
					t.Fatal("source does not offer Load more for an empty history")
				}
			}
			assertNoMore()
			if !graph {
				for _, options := range []workspace.HistoryOptions{
					{AllBranches: true}, {Author: "absent"}, {Path: "missing"}, {Grep: "no match"},
				} {
					s.HistoryOptions = options
					tt.Frame()
					historyRect(t, tt, "No commits yet")
					historyRect(t, tt, "No commits match the current filters.")
					assertNoMore()
				}
			}
			s.Busy = true
			tt.Frame()
			historyRect(t, tt, "Reading history…")
			if tt.HasText("No commits yet") || tt.HasText("No commits match") {
				t.Fatal("loading history incorrectly reports an empty result")
			}
			assertNoMore()
			s.Error = errors.New("fixture history read failed")
			tt.Frame()
			unavailable := "History unavailable"
			if graph {
				unavailable = "Graph unavailable"
			}
			historyRect(t, tt, unavailable)
			if tt.HasText("Reading history…") || tt.HasText("No commits yet") {
				t.Fatal("history error should take precedence over loading and empty results")
			}
			assertNoMore()
		})
	}
}

func TestHistoryMountedManyCommitsScrollWithLoadMore(t *testing.T) {
	for _, graph := range []bool{false, true} {
		t.Run(fmt.Sprintf("graph=%t", graph), func(t *testing.T) {
			_, tt := historyLayoutFixture(t, graph, 80, 600)
			filter := historyFilterRect(t, tt, graph)
			first := historyRect(t, tt, "Fixture commit 000")
			if gap := first.Y - (filter.Y + filter.H); gap < 0 || gap > 12 {
				t.Errorf("long list does not begin below filters: filter=%v first=%v", filter, first)
			}
			if more, ok := tt.Find("Load more"); ok && more.Y >= 0 && more.Y+more.H <= 600 {
				t.Errorf("Load more must scroll after all commits, not remain pinned in the viewport: %v", more)
			}
			tt.Scroll(600, 500, 0, 100000)
			last := historyRect(t, tt, "Fixture commit 079")
			more := historyRect(t, tt, "Load more")
			if last.Y < filter.Y+filter.H || more.Y < last.Y+last.H || more.Y > last.Y+last.H+12 || more.Y+more.H > 600 {
				t.Errorf("scrolling to end must reveal final commit followed by Load more: filter=%v last=%v more=%v", filter, last, more)
			}
			if now := historyFilterRect(t, tt, graph); now != filter {
				t.Errorf("filters scrolled with commits: before=%v after=%v", filter, now)
			}
			tt.Scroll(600, 500, 0, -100000)
			if now := historyRect(t, tt, "Fixture commit 000"); now != first {
				t.Errorf("scrolling back did not restore first commit: before=%v after=%v", first, now)
			}
		})
	}
}

func TestHistoryMountedCommitExpansionAndLoadMore(t *testing.T) {
	for _, graph := range []bool{false, true} {
		t.Run(fmt.Sprintf("graph=%t", graph), func(t *testing.T) {
			g := testRepo(t)
			for _, subject := range []string{"Older fixture commit", "Newest fixture commit"} {
				writeTest(t, g, "a", subject)
				if _, err := g.Commit(context.Background(), subject, true); err != nil {
					t.Fatal(err)
				}
			}
			s := New(g.Dir, Options{})
			defer s.Close()
			s.Tab = 1
			if graph {
				s.Tab = 2
			}
			s.Refresh()
			awaitIdle(t, s)
			if s.Error != nil {
				t.Fatal(s.Error)
			}
			tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1222, 878)
			first := historyRect(t, tt, "Newest fixture commit")
			second := historyRect(t, tt, "Older fixture commit")
			selectFirst := func() {
				// Avoid the nested author filter button while exercising the row's
				// actual selection callback, rather than assigning CommitDetail.
				tt.ClickAt(first.X+first.W-20, first.Y+first.H/2)
				awaitIdle(t, s)
				tt.Frame()
			}
			selectFirst()
			if s.Error != nil || s.CommitDetail == nil || s.CommitDetail.Subject != "Newest fixture commit" {
				t.Fatalf("commit selection failed: detail=%v error=%v", s.CommitDetail, s.Error)
			}
			detail := historyRect(t, tt, s.CommitDetail.Hash)
			expandedSecond := historyRect(t, tt, "Older fixture commit")
			more := historyRect(t, tt, "Load more")
			if firstNow := historyRect(t, tt, "Newest fixture commit"); firstNow != first {
				t.Errorf("expanding first commit moved its header: before=%v after=%v", first, firstNow)
			}
			if detail.Y < first.Y+first.H || detail.Y+detail.H > expandedSecond.Y || expandedSecond.Y <= second.Y {
				t.Errorf("selected detail must expand inline before next commit: first=%v detail=%v second=%v", first, detail, expandedSecond)
			}
			if gap := more.Y - (expandedSecond.Y + expandedSecond.H); gap < 0 || gap > 12 {
				t.Errorf("Load more detached from expanded list content: second=%v more=%v", expandedSecond, more)
			}
			if err := tt.Click("Close"); err != nil {
				t.Fatal(err)
			}
			if s.CommitDetail != nil || historyRect(t, tt, "Older fixture commit") != second {
				t.Fatal("closing detail did not restore collapsed row geometry")
			}
			selectFirst()
			selectFirst()
			if s.CommitDetail != nil || historyRect(t, tt, "Older fixture commit") != second {
				t.Fatal("repeated row selection did not collapse detail")
			}

			limit, graphLimit := s.HistoryOptions.Limit, s.GraphLimit
			s.Busy = true
			tt.Frame()
			if err := tt.Click("Load more"); err != nil {
				t.Fatal(err)
			}
			if s.HistoryOptions.Limit != limit || s.GraphLimit != graphLimit {
				t.Fatal("disabled Load more changed limits")
			}
			s.Busy = false
			tt.Frame()
			if err := tt.Click("Load more"); err != nil {
				t.Fatal(err)
			}
			awaitIdle(t, s)
			tt.Frame()
			if graph {
				graphLimit += 40
			} else {
				limit += 40
			}
			if s.Error != nil || s.HistoryOptions.Limit != limit || s.GraphLimit != graphLimit {
				t.Fatalf("Load more did not refresh the correct list: history=%d graph=%d error=%v", s.HistoryOptions.Limit, s.GraphLimit, s.Error)
			}
			if historyRect(t, tt, "Newest fixture commit") != first || historyRect(t, tt, "Older fixture commit") != second {
				t.Fatal("refreshing an exhausted list displaced its existing commits")
			}
			if !graph {
				if err := tt.Click("All branches"); err != nil {
					t.Fatal(err)
				}
				if !s.HistoryOptions.AllBranches {
					t.Fatal("all-branches toggle stopped updating history filters")
				}
				if err := tt.Click("Commit search"); err != nil {
					t.Fatal(err)
				}
				tt.Type("Older fixture")
				if err := tt.Click("Apply"); err != nil {
					t.Fatal(err)
				}
				awaitIdle(t, s)
				tt.Frame()
				if s.Error != nil || len(s.Data.History) != 1 || historyRect(t, tt, "Older fixture commit").Y != first.Y {
					t.Fatalf("search did not keep the filtered commit at the top: history=%v error=%v", s.Data.History, s.Error)
				}
				if err := tt.Click("Clear filters"); err != nil {
					t.Fatal(err)
				}
				awaitIdle(t, s)
				tt.Frame()
				if s.HistoryOptions.Grep != "" || s.HistoryOptions.AllBranches || historyRect(t, tt, "Newest fixture commit") != first || historyRect(t, tt, "Older fixture commit") != second {
					t.Fatal("clearing filters did not restore the two-commit top-anchored layout")
				}
			}
		})
	}
}
