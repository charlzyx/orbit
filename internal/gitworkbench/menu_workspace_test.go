package gitworkbench

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"path/filepath"
	"testing"
	"time"
)

func TestGitChangesMenuCannotRetargetWorkspace(t *testing.T) {
	for _, transition := range []string{"other", "roundtrip", "replacement", "stage"} {
		t.Run(transition, func(t *testing.T) {
			s, tt := embeddedReviewFixture(t, false)
			var opened []string
			s.Options.OpenFile = func(path string) { opened = append(opened, path) }
			tt.Frame()
			if err := tt.RightClick("long.go"); err != nil {
				t.Fatal(err)
			}
			if len(tt.Menu()) == 0 {
				t.Fatal("menu missing")
			}
			original := s.Git.Dir
			data, expanded := s.Data, s.diffExpanded
			switch transition {
			case "other", "roundtrip":
				s.SetWorkspace(t.TempDir())
				if transition == "roundtrip" {
					s.SetWorkspace(original)
				}
				s.initialized, s.refreshed = true, time.Now()
				s.Data, s.diffExpanded = data, expanded
			case "replacement":
				s.Data.Changes = append([]workspace.StatusRow(nil), s.Data.Changes...)
			case "stage":
				s.Data.Changes[0].Index = ' '
			}
			// No intervening frame: an OS-menu choice must not migrate to reused IDs.
			_ = tt.ChooseMenuItem("Open file")
			if len(opened) != 0 {
				t.Fatal("stale menu opened a file after target changed", transition, opened)
			}
			tt.CloseMenu()
			tt.Frame()
			if err := tt.RightClick("long.go"); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem("Open file"); err != nil {
				t.Fatal(err)
			}
			if len(opened) != 1 || opened[0] != filepath.Join(s.Git.Dir, "long.go") {
				t.Fatal("fresh menu did not use current workspace", opened)
			}
		})
	}
}

func TestGitChangesMenuStableFramesCancelAndFreshSnapshot(t *testing.T) {
	for _, change := range []string{"frame", "same-workspace", "cancel"} {
		t.Run(change, func(t *testing.T) {
			s, tt := embeddedReviewFixture(t, false)
			if err := tt.RightClick("long.go"); err != nil {
				t.Fatal(err)
			}
			switch change {
			case "frame":
				tt.Frame()
			case "same-workspace":
				s.SetWorkspace(s.Git.Dir)
			case "cancel":
				s.Cancel()
			}
			_ = tt.ChooseMenuItem("Copy absolute path")
			want := filepath.Join(s.Git.Dir, "long.go")
			if change == "cancel" {
				if tt.Clipboard() != "" {
					t.Fatal("cancelled epoch menu still ran", tt.Clipboard())
				}
			} else if tt.Clipboard() != want {
				t.Fatal("normal menu broken", tt.Clipboard(), want)
			}
			tt.CloseMenu()
			tt.Frame()
			if err := tt.RightClick("long.go"); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem("File history"); err != nil {
				t.Fatal(err)
			}
			if s.Tab != 1 || s.HistoryOptions.Path != "long.go" || !s.HistoryOptions.Follow {
				t.Fatal("fresh history action failed", s.Tab, s.HistoryOptions)
			}
			s.Cancel()
		})
	}
}

func TestGitHistoryMenuWorkspaceAndSnapshotIdentity(t *testing.T) {
	for _, graph := range []bool{false, true} {
		for _, change := range []string{"roundtrip", "replacement", "normal"} {
			t.Run(fmt.Sprintf("%t/%s", graph, change), func(t *testing.T) {
				s, tt := historyLayoutFixture(t, graph, 2, 878)
				var review []string
				s.Options.OpenReview = func(source, ref, path string) { review = append(review, ref) }
				tt.Frame()
				if err := tt.RightClick("Fixture commit 000"); err != nil {
					t.Fatal(err)
				}
				switch change {
				case "roundtrip":
					dir, data, tab := s.Git.Dir, s.Data, s.Tab
					s.SetWorkspace(t.TempDir())
					s.SetWorkspace(dir)
					s.Data, s.Tab = data, tab
					s.initialized, s.refreshed = true, time.Now()
				case "replacement":
					s.Data.History = append([]workspace.CommitEntry(nil), s.Data.History...)
					s.Data.Graph = append([]workspace.GraphRow(nil), s.Data.Graph...)
				}
				_ = tt.ChooseMenuItem("Review commit")
				if change == "normal" {
					if len(review) != 1 || review[0] != s.Data.History[0].Hash {
						t.Fatal("normal review menu failed", review)
					}
				} else if len(review) != 0 {
					t.Fatal("stale commit menu ran", review)
				}
			})
		}
	}
}

func TestGitCommitFileMenuIdentityAndNavigation(t *testing.T) {
	for _, change := range []string{"replacement", "roundtrip", "normal"} {
		t.Run(change, func(t *testing.T) {
			s, tt := historyLayoutFixture(t, false, 2, 878)
			s.CommitDetail = &workspace.CommitDetail{Hash: s.Data.History[0].Hash, Subject: "Detail", Files: []workspace.CommitFile{{Path: "old.go", Status: "M"}}}
			var opened []string
			s.Options.OpenFile = func(p string) { opened = append(opened, p) }
			tt.Frame()
			if err := tt.RightClick("M old.go"); err != nil {
				t.Fatal(err)
			}
			switch change {
			case "replacement":
				d := *s.CommitDetail
				s.CommitDetail = &d
			case "roundtrip":
				dir, data, detail := s.Git.Dir, s.Data, s.CommitDetail
				s.SetWorkspace(t.TempDir())
				s.SetWorkspace(dir)
				s.Data, s.CommitDetail, s.Tab = data, detail, 1
				s.initialized, s.refreshed = true, time.Now()
			}
			_ = tt.ChooseMenuItem("Open current file")
			if change == "normal" {
				if len(opened) != 1 {
					t.Fatal("normal current-file menu failed", opened)
				}
			} else if len(opened) != 0 {
				t.Fatal("stale commit-file menu ran", opened)
			}
			tt.CloseMenu()
			tt.Frame()
			if err := tt.RightClick("M old.go"); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem("File history (follow renames)"); err != nil {
				t.Fatal(err)
			}
			if s.HistoryOptions.Path != "old.go" || !s.HistoryOptions.Follow || s.CommitDetail != nil {
				t.Fatal("file history navigation lost source semantics", s.HistoryOptions, s.CommitDetail)
			}
			s.Cancel()
		})
	}
}

func TestGitHubMenusSnapshotIdentity(t *testing.T) {
	for _, pull := range []bool{false, true} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("pull=%t/replace=%t", pull, replace), func(t *testing.T) {
				s, tt := embeddedReviewFixture(t, false)
				s.Tab = 4
				if pull {
					s.Tab = 5
				}
				s.Data.GitHub = workspace.GitHubStatus{Installed: true, Authenticated: true, Supported: true}
				s.Data.Issues = []workspace.GitHubIssue{{Number: 1, Title: "Issue fixture", URL: "https://example.invalid/issue/1"}}
				s.Data.Pulls = []workspace.GitHubPull{{Number: 1, Title: "Pull fixture", URL: "https://example.invalid/pull/1"}}
				var opened []string
				s.Options.OpenURL = func(u string) { opened = append(opened, u) }
				tt.Frame()
				label := "Issue fixture"
				if pull {
					label = "Pull fixture"
				}
				if err := tt.RightClick(label); err != nil {
					t.Fatal(err)
				}
				if replace {
					s.Data.Issues = append([]workspace.GitHubIssue(nil), s.Data.Issues...)
					s.Data.Pulls = append([]workspace.GitHubPull(nil), s.Data.Pulls...)
				}
				_ = tt.ChooseMenuItem("Open on GitHub")
				if replace {
					if len(opened) != 0 {
						t.Fatal("stale GitHub menu ran", opened)
					}
				} else if len(opened) != 1 {
					t.Fatal("normal GitHub menu failed", opened)
				}
			})
		}
	}
}

func TestGitQueuedHistoryCannotCrossEpoch(t *testing.T) {
	for _, change := range []string{"cancel", "roundtrip", "close"} {
		t.Run(change, func(t *testing.T) {
			s := New(t.TempDir(), Options{})
			defer s.Close()
			s.pendingHistory = &historyNavigation{workspace: s.Git.Dir, generation: s.generation, path: "old.go", follow: true}
			switch change {
			case "cancel":
				s.Cancel()
			case "close":
				s.Close()
			case "roundtrip":
				dir := s.Git.Dir
				s.SetWorkspace(t.TempDir())
				s.SetWorkspace(dir)
			}
			s.applyPendingHistory()
			if s.pendingHistory != nil || s.Tab != 0 || s.HistoryOptions.Path != "" || s.Busy {
				t.Fatal("superseded navigation ran", s.Tab, s.HistoryOptions, s.Busy)
			}
		})
	}
}

func TestGitChangesMenuDoesNotConfuseSamePathStages(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	var sources []string
	s.Options.OpenReview = func(source, ref, path string) { sources = append(sources, source+":"+path) }
	tt.Frame()
	if err := tt.RightClick("long.go"); err != nil {
		t.Fatal(err)
	}
	s.Data.Changes[0].Index = ' '
	_ = tt.ChooseMenuItem("Open in Review")
	if len(sources) != 0 {
		t.Fatal("old staged menu was retargeted", sources)
	}
	tt.CloseMenu()
	tt.Frame()
	if err := tt.RightClick("long.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Open in Review"); err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || sources[0] != "unstaged:long.go" {
		t.Fatal("fresh same-path menu used wrong source", sources)
	}
}
