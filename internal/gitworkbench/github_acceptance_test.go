package gitworkbench

// Source-first GitHub regressions use only an executable fake gh and disposable
// real Git repositories. No credentials, network calls or remote writes occur.
import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type githubFixture struct{ binary, dir, log string }

func newGitHubFixture(t *testing.T) githubFixture {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake CLI executable uses a Unix shebang; native Windows gate remains separate")
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is required for isolated fake gh")
	}
	dir := t.TempDir()
	f := githubFixture{filepath.Join(dir, "gh-fixture"), dir, filepath.Join(dir, "argv.jsonl")}
	script := "#!" + python + "\n" + `import json, pathlib, sys, time
root=pathlib.Path(__file__).parent
a=sys.argv[1:]
with (root/'argv.jsonl').open('a') as f: f.write(json.dumps(a)+'\n')
if a==['--version']: print('gh version 2.90.0 (fixture)')
elif a[:2]==['auth','status']: print('fixture authenticated')
elif a[:2]==['api','user']: print('{"login":"fixture","name":"Local Fixture"}')
elif a[:2]==['label','list']:
 if (root/'labels-fail').exists(): print('fixture labels unavailable',file=sys.stderr);sys.exit(1)
 print('[{"name":"bug","color":"ff0000"}]')
elif len(a)>1 and a[1]=='create':
 time.sleep(0.04)
 if (root/'create-fail').exists(): print('fixture create refused',file=sys.stderr);sys.exit(1)
 print('https://example.invalid/repo/'+('issues' if a[0]=='issue' else 'pull')+'/42')
elif len(a)>1 and a[1]=='list': print('[{"number":42,"title":"Fixture item","state":"OPEN"}]')
elif len(a)>1 and a[1]=='view':
 if (root/'view-fail').exists(): print('fixture detail unavailable',file=sys.stderr);sys.exit(1)
 print('{"number":42,"title":"Fixture item","state":"OPEN","body":"Created locally in fake gh"}')
else: print('unexpected fixture argv: '+repr(a),file=sys.stderr);sys.exit(3)
`
	if err := os.WriteFile(f.binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return f
}
func (f githubFixture) flag(t *testing.T, name string, on bool) {
	t.Helper()
	p := filepath.Join(f.dir, name)
	if on {
		if e := os.WriteFile(p, []byte("1"), 0600); e != nil {
			t.Fatal(e)
		}
	} else if e := os.Remove(p); e != nil && !os.IsNotExist(e) {
		t.Fatal(e)
	}
}
func (f githubFixture) calls(t *testing.T, first, second string) [][]string {
	t.Helper()
	b, e := os.ReadFile(f.log)
	if e != nil && !os.IsNotExist(e) {
		t.Fatal(e)
	}
	var out [][]string
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		var a []string
		if e := json.Unmarshal([]byte(line), &a); e != nil {
			t.Fatal(e)
		}
		if len(a) > 1 && a[0] == first && a[1] == second {
			out = append(out, a)
		}
	}
	return out
}

func TestGitHubNativeCancelRetainsIndependentDrafts(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	s.openDialog("new-issue")
	s.title = "Issue title"
	s.body = "Issue prose"
	s.notes = "Issue notes"
	tt := ui.NewTester(func(c *ui.Context) { s.dialogView(c) }, 900, 950)
	if e := tt.Click("Cancel"); e != nil {
		t.Fatal(e)
	}
	s.openDialog("new-pull")
	s.title = "Pull title"
	s.body = "Pull prose"
	s.notes = "Pull notes"
	tt.Frame()
	if e := tt.Click("Cancel"); e != nil {
		t.Fatal(e)
	}
	s.openDialog("issue-comment")
	s.body = "Unrelated comment"
	s.dialogOpen = false
	s.openDialog("new-issue")
	if s.title != "Issue title" || s.body != "Issue prose" || s.notes != "Issue notes" {
		t.Fatalf("source Issue draft lost on cancel/reopen: %q %q %q", s.title, s.body, s.notes)
	}
	s.openDialog("new-pull")
	if s.title != "Pull title" || s.body != "Pull prose" || s.notes != "Pull notes" {
		t.Fatalf("source PR draft lost on cancel/reopen: %q %q %q", s.title, s.body, s.notes)
	}
}
func TestGitHubNativeCreateFailureKeepsFormAndRetryClearsOnce(t *testing.T) {
	for _, kind := range []string{"issue", "pull"} {
		t.Run(kind, func(t *testing.T) {
			g := testRepo(t)
			f := newGitHubFixture(t)
			f.flag(t, "create-fail", true)
			s := New(g.Dir, Options{})
			defer s.Close()
			s.GitHub.Binary = f.binary
			s.Tab = 4
			if kind == "pull" {
				s.Tab = 5
			}
			s.openDialog("new-" + kind)
			s.title = "  Title with Unicode 雪  "
			s.body = "Body\nsecond line"
			s.notes = "Keep notes"
			tt := ui.NewTester(func(c *ui.Context) { s.dialogView(c) }, 900, 950)
			if e := tt.Click("Create"); e != nil {
				t.Fatal(e)
			}
			if !s.Busy || !s.dialogOpen {
				t.Fatalf("source form must stay open during create: busy=%v open=%v", s.Busy, s.dialogOpen)
			}
			// A second direct request while busy must not spawn another mutation.
			if kind == "issue" {
				s.CreateIssue(workspace.CreateIssueOptions{Title: "duplicate"})
			} else {
				s.CreatePull(workspace.CreatePullOptions{Title: "duplicate"})
			}
			awaitIdle(t, s)
			if s.Error == nil || !s.dialogOpen || s.body != "Body\nsecond line" || s.notes != "Keep notes" {
				t.Fatalf("failed create lost form/draft: error=%v open=%v body=%q notes=%q", s.Error, s.dialogOpen, s.body, s.notes)
			}
			f.flag(t, "create-fail", false)
			tt.Frame()
			if e := tt.Click("Create"); e != nil {
				t.Fatal(e)
			}
			awaitIdle(t, s)
			if s.Error != nil || s.dialogOpen || s.title != "" || s.body != "" || s.notes != "" {
				t.Fatalf("successful retry did not clear and close: error=%v open=%v title=%q", s.Error, s.dialogOpen, s.title)
			}
			command := kind
			if kind == "pull" {
				command = "pr"
			}
			calls := f.calls(t, command, "create")
			if len(calls) != 2 {
				t.Fatalf("duplicate submit: %v", calls)
			}
			for _, a := range calls {
				for i, v := range a {
					if v == "--title" && (i+1 == len(a) || a[i+1] != "Title with Unicode 雪") {
						t.Fatalf("source trims title: %v", a)
					}
				}
			}
		})
	}
}
func TestGitHubOptionalLabelsNeverBlockIssuesOrPulls(t *testing.T) {
	for _, tab := range []int{4, 5} {
		t.Run([]string{"issues", "pulls"}[tab-4], func(t *testing.T) {
			g := testRepo(t)
			f := newGitHubFixture(t)
			f.flag(t, "labels-fail", true)
			s := New(g.Dir, Options{})
			defer s.Close()
			s.GitHub.Binary = f.binary
			s.Tab = tab
			s.Refresh()
			awaitIdle(t, s)
			if s.Error != nil {
				t.Fatalf("optional labels blocked source list: %v", s.Error)
			}
			if tab == 4 && len(s.Data.Issues) != 1 || tab == 5 && len(s.Data.Pulls) != 1 {
				t.Fatal("source list not installed", s.Data)
			}
			if tab == 5 && len(f.calls(t, "label", "list")) != 0 {
				t.Fatal("PR list unexpectedly fetches issue labels")
			}
		})
	}
}

func TestGitHubLabelCacheSurvivesTabRoundTrip(t *testing.T) {
	g := testRepo(t)
	f := newGitHubFixture(t)
	s := New(g.Dir, Options{})
	defer s.Close()
	s.GitHub.Binary = f.binary
	s.SetTab(4)
	awaitIdle(t, s)
	if s.Error != nil || len(s.Data.Labels) != 1 {
		t.Fatal(s.Error, s.Data.Labels)
	}
	s.SetTab(1)
	awaitIdle(t, s)
	s.SetTab(5)
	awaitIdle(t, s)
	s.SetTab(4)
	awaitIdle(t, s)
	s.Refresh()
	awaitIdle(t, s)
	if s.Error != nil || len(s.Data.Labels) != 1 || len(f.calls(t, "label", "list")) != 1 {
		t.Fatalf("source caches labels once across tabs: err=%v labels=%v calls=%v", s.Error, s.Data.Labels, f.calls(t, "label", "list"))
	}
}
func TestGitHubTemplateReopenAndBlankMatchSource(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	s.Data.IssueTemplates = []workspace.RepositoryTemplate{{Body: "Issue skeleton", TitlePrefix: "[Issue] "}}
	s.Data.PullTemplates = []workspace.RepositoryTemplate{{Body: "Pull skeleton", TitlePrefix: "[PR] "}}
	s.openDialog("new-issue")
	s.title = "User title"
	s.body = "Edited prose"
	s.notes = "Notes"
	s.dialogOpen = false
	s.openDialog("new-issue")
	if s.title != "User title" || s.body != "Issue skeleton" || s.notes != "Notes" {
		t.Fatalf("source lone template reapplies body but preserves title/notes: %q %q %q", s.title, s.body, s.notes)
	}
	s.body = "Pull skeleton"
	s.applyTemplate(false, -1)
	if s.body != "" {
		t.Fatal("Blank must recognize either form's known template skeleton", s.body)
	}
	s.body = "Unrelated prose"
	s.applyTemplate(false, -1)
	if s.body != "Unrelated prose" {
		t.Fatal("Blank erased user prose")
	}
	s.SetWorkspace(t.TempDir())
	s.openDialog("new-issue")
	if s.title != "User title" || s.body != "Unrelated prose" || s.notes != "Notes" {
		t.Fatal("source keeps draft inputs across workspace changes")
	}
}
func TestGitHubCreatedMutationNotRetriedWhenDetailRefreshFails(t *testing.T) {
	g := testRepo(t)
	f := newGitHubFixture(t)
	f.flag(t, "view-fail", true)
	s := New(g.Dir, Options{})
	defer s.Close()
	s.GitHub.Binary = f.binary
	s.Tab = 4
	s.openDialog("new-issue")
	s.title = "Created once"
	s.body = "Draft"
	s.CreateIssue(workspace.CreateIssueOptions{Title: s.title, Body: s.body})
	awaitIdle(t, s)
	if s.Error == nil || s.dialogOpen || s.title != "" || s.body != "" {
		t.Fatalf("successful write with failed detail load must close/clear source form: %v %v %q %q", s.Error, s.dialogOpen, s.title, s.body)
	}
	if len(f.calls(t, "issue", "create")) != 1 || s.lastAction != nil {
		t.Fatal("failed follow-up read exposed duplicate mutation retry")
	}
	if !strings.Contains(s.Status, "Issue created:") {
		t.Fatal("lost successful mutation receipt", s.Status)
	}
}
