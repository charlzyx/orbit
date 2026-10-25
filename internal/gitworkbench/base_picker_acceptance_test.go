package gitworkbench

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// Expected decisions are independently executed by the unchanged Rust methods
// in acceptance/run-github-state-oracle.py. No remote requests occur here.
func TestGitHubSourceDefaultBase(t *testing.T) {
	branch := func(name string) workspace.RefEntry { return workspace.RefEntry{Name: name, Kind: workspace.RefBranch} }
	for _, tc := range []struct {
		name, current, want string
		refs                []workspace.RefEntry
	}{
		{"empty", "", "main", nil},
		{"only-current", "feature", "main", []workspace.RefEntry{branch("feature")}},
		{"skip-current", "feature", "release", []workspace.RefEntry{branch("feature"), branch("release")}},
		{"prefer-master", "feature", "master", []workspace.RefEntry{branch("feature"), branch("release"), branch("master")}},
		{"prefer-main", "feature", "main", []workspace.RefEntry{branch("master"), branch("main")}},
		{"main-can-be-current", "main", "main", []workspace.RefEntry{branch("main"), branch("release")}},
		{"preferred-name-kind-agnostic", "feature", "main", []workspace.RefEntry{{Name: "main", Kind: workspace.RefTag}, branch("release")}},
		{"no-local-fallback", "feature", "main", []workspace.RefEntry{{Name: "origin/feature", Kind: workspace.RefRemote}, {Name: "v1", Kind: workspace.RefTag}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := New(t.TempDir(), Options{})
			defer s.Close()
			s.Data.Branch, s.Data.Refs = tc.current, tc.refs
			for _, ref := range tc.refs {
				if ref.Kind == workspace.RefBranch {
					s.Data.Branches = append(s.Data.Branches, ref.Name)
				}
			}
			s.openDialog("new-pull")
			if s.base != tc.want {
				t.Fatalf("default_base source result %q; Go selected %q", tc.want, s.base)
			}
		})
	}
}

func TestGitHubBaseFilterMatchesRustUnicodeAndStableSort(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	for _, name := range []string{"zeta", "Beta", "beta", "ALPHA", "Σ", "ΟΣ", "İtem"} {
		s.Data.Refs = append(s.Data.Refs, workspace.RefEntry{Name: name, Kind: workspace.RefBranch})
	}
	for _, tc := range []struct {
		query string
		want  []string
	}{
		{"  bEt  ", []string{"Beta", "beta"}},
		{"ος", []string{"ΟΣ"}},
		{"i\u0307", []string{"İtem"}},
		{"absent", nil},
	} {
		s.baseQuery = tc.query
		if got := s.baseBranches(); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("query %q: got %q want Rust %q", tc.query, got, tc.want)
		}
	}
}

func TestGitHubBasePickerEmptyDismissRepeatAndErrorDoNotMutateDraft(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	s.Data.Refs = []workspace.RefEntry{{Name: "main", Kind: workspace.RefBranch}}
	s.openDialog("new-pull")
	s.title, s.body, s.draft = "Keep title", "Keep prose", true
	s.Busy, s.Activity = true, "Refreshing"
	tt := ui.NewTester(func(c *ui.Context) { s.dialogView(c) }, 900, 950)
	if e := tt.Click("Base branch"); e != nil {
		t.Fatal(e)
	}
	popup, ok := tt.Find("PR base branches")
	chip, _ := tt.Find("Base branch")
	if !ok || popup.W != 320 || popup.Y+popup.H > chip.Y {
		t.Fatalf("source 320px popup must anchor above base chip: popup=%v chip=%v", popup, chip)
	}
	tt.Type("no-such-local-ref")
	if _, ok := tt.Find("No matching branches"); !ok {
		t.Fatal("source no-match row absent")
	}
	tt.Key(0, ui.KeyEnter)
	if s.base != "main" || !s.basePickerOpen || !s.dialogOpen {
		t.Fatal("empty picker Enter submitted or changed base")
	}
	// A load error must not invent refs, erase prose or close the picker.
	s.Busy, s.Error = false, errors.New("fixture refresh error")
	tt.Frame()
	if !s.basePickerOpen || s.title != "Keep title" || s.body != "Keep prose" || !s.draft {
		t.Fatal("loading/error transition discarded visible draft or selection")
	}
	// Test actual outside-pointer dismissal, then same-gesture suppression.
	if e := tt.Click("New pull request"); e != nil {
		t.Fatal(e)
	}
	if s.basePickerOpen || !s.dialogOpen {
		t.Fatal("outside press should dismiss only anchored popup")
	}
	if e := tt.Click("Base branch"); e != nil {
		t.Fatal(e)
	}
	if s.basePickerOpen {
		t.Fatal("source 200ms dismissal guard allowed immediate reopen")
	}
	s.baseDismissedAt = time.Now().Add(-201 * time.Millisecond)
	if e := tt.Click("Base branch"); e != nil {
		t.Fatal(e)
	}
	if !s.basePickerOpen || s.baseQuery != "" {
		t.Fatal("later reopen did not reset branch filter")
	}
	// MyGo's native popup Escape dismisses without closing its owning form;
	// this checks containment, not GPUI's unhandled PickerCancel dispatch.
	tt.Key(0, ui.KeyEscape)
	if s.basePickerOpen || !s.dialogOpen || s.title != "Keep title" {
		t.Fatal("popup Escape escaped into owning form")
	}
}

func TestGitHubWorkspaceChangeRetainsSourceDraftsAndLabelCache(t *testing.T) {
	a, b := testRepo(t), testRepo(t)
	f := newGitHubFixture(t)
	s := New(a.Dir, Options{})
	defer s.Close()
	s.GitHub.Binary = f.binary
	s.SetTab(4)
	awaitIdle(t, s)
	s.openDialog("new-issue")
	s.title, s.body, s.notes = "Issue title", "Issue prose", "Issue notes"
	s.openDialog("new-pull")
	s.title, s.body, s.notes, s.draft = "PR title", "PR prose", "PR notes", true
	s.Data.IssueTemplates = []workspace.RepositoryTemplate{{Body: "old template"}}
	s.Data.PullTemplates = []workspace.RepositoryTemplate{{Body: "old pull template"}}
	s.SetWorkspace(b.Dir)
	if s.dialogOpen || len(s.Data.IssueTemplates)+len(s.Data.PullTemplates) != 0 {
		t.Fatal("workspace change must close forms and clear repository templates")
	}
	s.openDialog("new-issue")
	if s.title != "Issue title" || s.body != "Issue prose" || s.notes != "Issue notes" {
		t.Fatalf("Rust retains Issue inputs across workspace changes: %q %q %q", s.title, s.body, s.notes)
	}
	s.openDialog("new-pull")
	if s.title != "PR title" || s.body != "PR prose" || s.notes != "PR notes" || !s.draft {
		t.Fatalf("Rust retains PR inputs and draft flag: %q %q %q draft=%v", s.title, s.body, s.notes, s.draft)
	}
	s.Refresh()
	awaitIdle(t, s)
	if s.Error != nil || len(s.Data.Labels) != 1 || len(f.calls(t, "label", "list")) != 1 {
		t.Fatalf("Rust keeps its label cache across roots: %v %v calls=%v", s.Error, s.Data.Labels, f.calls(t, "label", "list"))
	}
	if len(f.calls(t, "issue", "create"))+len(f.calls(t, "pr", "create")) != 0 {
		t.Fatal("retaining a draft must never submit it")
	}
	// Workspace retention is in-memory only. A new panel starts with no draft.
	fresh := New(b.Dir, Options{})
	defer fresh.Close()
	fresh.openDialog("new-issue")
	if fresh.title != "" || fresh.body != "" || fresh.notes != "" || len(fresh.Data.Labels) != 0 {
		t.Fatal("source retention was incorrectly persisted across panel lifetime")
	}
}

func TestGitHubNativeBasePickerFiltersLocalBranchesWithoutImplicitSelection(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	s.Data.Branch = "feature"
	s.Data.Refs = []workspace.RefEntry{{Name: "feature", Kind: workspace.RefBranch}, {Name: "zeta", Kind: workspace.RefBranch}, {Name: "Beta", Kind: workspace.RefBranch}, {Name: "beta", Kind: workspace.RefBranch}, {Name: "ALPHA", Kind: workspace.RefBranch}, {Name: "beta-remote", Kind: workspace.RefRemote}, {Name: "beta-tag", Kind: workspace.RefTag}}
	s.Data.Branches = []string{"feature", "zeta", "Beta", "beta", "ALPHA"}
	s.openDialog("new-pull")
	s.base = "feature"
	tt := ui.NewTester(func(c *ui.Context) { s.dialogView(c) }, 900, 950)
	if e := tt.Click("Base branch"); e != nil {
		t.Fatal(e)
	}
	if _, ok := tt.Find("Local branches"); !ok {
		t.Fatal("source anchored local-branch popup is missing")
	}
	if _, ok := tt.Find("beta-remote"); ok {
		t.Fatal("remote ref leaked into local-only PR picker")
	}
	if _, ok := tt.Find("beta-tag"); ok {
		t.Fatal("tag leaked into local-only PR picker")
	}
	alpha, _ := tt.Find("ALPHA")
	beta, _ := tt.Find("Beta")
	zeta, _ := tt.Find("zeta")
	if !(alpha.Y < beta.Y && beta.Y < zeta.Y) {
		t.Fatalf("source case-insensitive sorting absent: %v %v %v", alpha, beta, zeta)
	}
	tt.Type("  bEt  ")
	if _, ok := tt.Find("ALPHA"); ok {
		t.Fatal("branch filter did not trim and match case-insensitively")
	}
	if _, ok := tt.Find("Beta"); !ok {
		t.Fatal("branch search input did not receive focus")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if s.base != "feature" || s.Busy || !s.dialogOpen {
		t.Fatal("unhandled source Picker actions must not pick a branch or submit the form")
	}
	if e := tt.Click("Beta"); e != nil {
		t.Fatal(e)
	}
	if s.base != "Beta" {
		t.Fatal("pointer selection did not update base", s.base)
	}
	if _, ok := tt.Find("Local branches"); ok {
		t.Fatal("selection did not dismiss popup")
	}
}

func TestGitHubNativeSelectedBaseSurvivesCreateFailureAndExactArgvRetry(t *testing.T) {
	g := testRepo(t)
	f := newGitHubFixture(t)
	f.flag(t, "create-fail", true)
	s := New(g.Dir, Options{})
	defer s.Close()
	s.GitHub.Binary = f.binary
	s.Tab = 5
	s.Data.Refs = []workspace.RefEntry{{Name: "main", Kind: workspace.RefBranch}, {Name: "release/雪", Kind: workspace.RefBranch}}
	s.openDialog("new-pull")
	s.title, s.body, s.draft = "  Chosen base PR  ", "Do not discard", true
	tt := ui.NewTester(func(c *ui.Context) { s.dialogView(c) }, 900, 950)
	for _, label := range []string{"Base branch", "release/雪", "Create"} {
		if e := tt.Click(label); e != nil {
			t.Fatal(e)
		}
	}
	if !s.Busy || !s.dialogOpen {
		t.Fatal("create must remain visibly pending")
	}
	// Repeated native click is disabled and cannot create a second process.
	_ = tt.Click("Create")
	awaitIdle(t, s)
	tt.Frame()
	if s.Error == nil || !s.dialogOpen || s.base != "release/雪" || !s.draft || s.body != "Do not discard" {
		t.Fatalf("failed PR changed selected fields: %v base=%q draft=%v body=%q", s.Error, s.base, s.draft, s.body)
	}
	f.flag(t, "create-fail", false)
	if e := tt.Click("Create"); e != nil {
		t.Fatal(e)
	}
	awaitIdle(t, s)
	calls := f.calls(t, "pr", "create")
	if len(calls) != 2 {
		t.Fatal("expected failed request and explicit retry only", calls)
	}
	for _, args := range calls {
		value := func(key string) string {
			for i, v := range args {
				if v == key && i+1 < len(args) {
					return args[i+1]
				}
			}
			return ""
		}
		hasDraft := false
		for _, arg := range args {
			hasDraft = hasDraft || arg == "--draft"
		}
		if value("--base") != "release/雪" || value("--title") != "Chosen base PR" || value("--body") != "Do not discard" || !hasDraft {
			t.Fatalf("native selected fields did not reach exact gh argv: %q", args)
		}
	}
	if s.Error != nil || s.dialogOpen || s.title != "" || s.body != "" {
		t.Fatal("successful explicit retry did not clear/close", s.Error)
	}
}
