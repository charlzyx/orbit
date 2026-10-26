package gitworkbench

import (
	"context"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCommitModelResolveIdentityAndUnknown(t *testing.T) {
	current := agent.ModelChoice{Provider: "active", ID: "session", ContextWindow: 64000}
	catalog := []agent.ModelChoice{{Provider: "one", ID: "same", ContextWindow: 32000}, {Provider: "two", ID: "same", ContextWindow: 1000000}}
	for _, tc := range []struct {
		pick    *preferences.Model
		current agent.ModelChoice
		want    agent.ModelChoice
	}{{nil, current, current}, {nil, agent.ModelChoice{}, agent.ModelChoice{}}, {&preferences.Model{Provider: "two", ID: "same"}, current, catalog[1]}, {&preferences.Model{Provider: "missing", ID: "same"}, current, agent.ModelChoice{Provider: "missing", ID: "same"}}} {
		if got := ResolveCommitModel(tc.pick, tc.current, catalog); !reflect.DeepEqual(got, tc.want) {
			t.Fatal(got, tc.want)
		}
	}
}
func commitPickerFixture(t *testing.T) (*State, *ui.Tester) {
	t.Helper()
	store := preferences.New(t.TempDir())
	s := New(t.TempDir(), Options{CommitModelStore: store, CommitModelCatalog: func() ([]agent.ModelChoice, agent.ModelChoice) {
		return []agent.ModelChoice{{ID: "same", Provider: "one", Name: "First"}, {ID: "same", Provider: "two", Name: "Second"}}, agent.ModelChoice{Provider: "active", ID: "chat"}
	}})
	t.Cleanup(s.Close)
	s.initialized = true
	s.refreshed = time.Now()
	s.Data.Changes = []workspace.StatusRow{{Path: "fixture", Index: 'M', Worktree: ' '}}
	tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1280, 900)
	return s, tt
}
func TestCommitModelMountedKeyboardFocusAndFollow(t *testing.T) {
	s, tt := commitPickerFixture(t)
	if e := tt.Click("Commit model"); e != nil {
		t.Fatal(e)
	}
	tt.Type("Second")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if s.commitModel.open || s.commitModel.selected == nil || s.commitModel.selected.Provider != "two" {
		t.Fatal(s.commitModel)
	}
	tt.Type("commit draft")
	if s.CommitMessage != "commit draft" {
		t.Fatal("selection did not return native editor focus", s.CommitMessage, tt.Texts())
	}
	if e := tt.Click("Commit model"); e != nil {
		t.Fatal(e)
	}
	tt.Type("no-match")
	if !tt.HasText("Follow active session") {
		t.Fatal(tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if s.commitModel.selected != nil {
		t.Fatal("follow not selected")
	}
	tt.Type("!")
	if !strings.Contains(s.CommitMessage, "!") {
		t.Fatal("follow focus lost")
	}
}
func TestCommitModelMountedEscapeBusyAndWorkspace(t *testing.T) {
	s, tt := commitPickerFixture(t)
	_ = tt.Click("Commit model")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	tt.Type("draft")
	if s.commitModel.open || s.CommitMessage != "draft" {
		t.Fatal("escape did not restore editor", s.CommitMessage)
	}
	s.Busy = true
	tt.Frame()
	_ = tt.Click("Commit model")
	if s.commitModel.open {
		t.Fatal("picker open during generation")
	}
	s.Busy = false
	s.refreshed = time.Now()
	tt.Frame()
	_ = tt.Click("Commit model")
	s.SetWorkspace(t.TempDir())
	if s.commitModel.open {
		t.Fatal("workspace did not close picker")
	}
}
func TestCommitModelWriteFailureAppliesOnlyInMemory(t *testing.T) {
	s, tt := commitPickerFixture(t)
	s.loadCommitModel()
	path := filepath.Join(s.Options.CommitModelStore.RootDir, "commit-model.json")
	if e := os.MkdirAll(path, 0700); e != nil {
		t.Fatal(e)
	}
	_ = tt.Click("Commit model")
	if e := tt.Click("Second · two"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if s.commitModel.selected == nil || s.commitModel.selected.Provider != "two" || s.Error == nil || !strings.Contains(s.Error.Error(), "not saved") {
		t.Fatal(s.commitModel, s.Error)
	}
	tt.Type("retained")
	if s.CommitMessage != "retained" {
		t.Fatal("save failure stole focus", s.CommitMessage)
	}
}
func TestCommitModelAdmissionFreezeAndLateResult(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(map[bool]string{true: "cancel", false: "workspace"}[cancel], func(t *testing.T) {
			g := testRepo(t)
			writeTest(t, g, "a", "fixture")
			entered := make(chan agent.ModelChoice, 1)
			release := make(chan struct{})
			done := make(chan struct{})
			active := agent.ModelChoice{Provider: "active", ID: "before", ContextWindow: 32000}
			s := New(g.Dir, Options{CommitModelStore: preferences.New(t.TempDir()), CommitModelCatalog: func() ([]agent.ModelChoice, agent.ModelChoice) { return nil, active }, GenerateSelectedCommit: func(ctx context.Context, staged bool, m agent.ModelChoice) (workspace.GeneratedText, error) {
				entered <- m
				<-release
				defer close(done)
				return workspace.GeneratedText{Text: "stale"}, ctx.Err()
			}})
			defer s.Close()
			s.IncludeUnstaged = true
			s.generateCommitFlow(false, false, false)
			active.ID = "after"
			m := <-entered
			if m.ID != "before" || m.ContextWindow != 32000 {
				t.Fatal("not admission snapshot", m)
			}
			s.selectCommitModel(&preferences.Model{Provider: "p", ID: "new"})
			if s.commitModel.selected != nil {
				t.Fatal("busy selection accepted")
			}
			if cancel {
				s.Cancel()
			} else {
				s.SetWorkspace(t.TempDir())
			}
			close(release)
			<-done
			s.Tick()
			if s.CommitMessage != "" {
				t.Fatal("late result applied", s.CommitMessage)
			}
		})
	}
}

func TestCommitModelMountedOutsideAndKeyboardReveal(t *testing.T) {
	s, tt := commitPickerFixture(t)
	s.Options.CommitModelCatalog = func() ([]agent.ModelChoice, agent.ModelChoice) {
		var models []agent.ModelChoice
		for i := 0; i < 20; i++ {
			models = append(models, agent.ModelChoice{Provider: "fixture", ID: fmt.Sprintf("m%02d", i), Name: fmt.Sprintf("Model %02d", i)})
		}
		return models, agent.ModelChoice{}
	}
	s.selectCommitModel(&preferences.Model{Provider: "fixture", ID: "m19"})
	tt.Frame()
	if e := tt.Click("Commit model"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if s.commitModel.cursor != 20 {
		t.Fatal("selected identity not highlighted", s.commitModel.cursor)
	}
	row, ok := tt.Find("Model 19 · fixture")
	panel, _ := tt.Find("Commit model picker")
	if !ok || row.Y < panel.Y || row.Y+row.H > panel.Y+panel.H+1 {
		t.Fatal("selected model not revealed", row, panel)
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if s.commitModel.selected.ID != "m18" {
		t.Fatal(s.commitModel.selected)
	}
	_ = tt.Click("Commit model")
	if e := tt.Click("Git"); e != nil {
		t.Fatal(e)
	}
	if s.commitModel.open {
		t.Fatal("outside did not close")
	}
	_ = tt.Click("Commit model")
	if s.commitModel.open {
		t.Fatal("same gesture reopened")
	}
	tt.Type("still editing")
	if s.CommitMessage != "still editing" {
		t.Fatal("dismissal guard lost message focus", s.CommitMessage)
	}
	s.commitModel.dismissed = time.Now().Add(-201 * time.Millisecond)
	_ = tt.Click("Commit model")
	if !s.commitModel.open {
		t.Fatal("later reopening blocked")
	}
	s.openDialog("new-issue")
	if s.commitModel.open {
		t.Fatal("stacked picker over modal")
	}
}

func TestCommitModelFirstRenderPreservesRealGitFailure(t *testing.T) {
	for _, selection := range []bool{false, true} {
		t.Run(fmt.Sprint(selection), func(t *testing.T) {
			s, tt := commitPickerFixture(t)
			if selection {
				s.loadCommitModel()
			}
			err := exec.Command("git", "-C", s.Git.Dir, "not-a-real-git-command").Run()
			if err == nil {
				t.Fatal("expected actual Git failure")
			}
			s.Error = err
			tt.Frame()
			if s.Error != err {
				t.Fatal("first picker render cleared Git error", s.Error)
			}
			if selection {
				s.selectCommitModel(nil)
				tt.Frame()
				if s.Error != err {
					t.Fatal("successful preference selection cleared Git error", s.Error)
				}
			}
		})
	}
}
