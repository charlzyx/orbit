package bugreportui

import (
	"context"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func finish(t *testing.T, s *State) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for s.Busy && time.Now().Before(deadline) {
		s.Tick()
		time.Sleep(time.Millisecond)
	}
	if s.Busy {
		t.Fatal("worker did not finish")
	}
	if s.Error != nil {
		t.Fatal(s.Error)
	}
}
func TestReportFormPreparationAndCancellation(t *testing.T) {
	s := New(Options{Version: "test", Home: t.TempDir()})
	defer s.Close()
	if s.IncludeDiagnostics || s.IncludeIdentity || s.CaptureConsent {
		t.Fatal("private data included by default")
	}
	s.Title = "Hang"
	s.What = "App hangs"
	s.Prepare()
	finish(t, s)
	if !s.reviewOpen || s.reviewed {
		t.Fatal("review bypassed")
	}
	s.Submit()
	if s.Busy {
		t.Fatal("submitted without confirmation")
	}
	s.Clear()
	if s.reviewOpen || s.Title != "" || s.ScreenshotCount() != 0 {
		t.Fatal("clear failed")
	}
	s.Notes = "Crashes on launch.\nDetails"
	s.Generate()
	finish(t, s)
	if s.Title != "Crashes on launch" || !strings.Contains(s.What, "Details") {
		t.Fatal(s.Title, s.What)
	}
}
func TestReportDraftOnlyExplicitNotes(t *testing.T) {
	var got string
	s := New(Options{GenerateDraft: func(ctx context.Context, k native.ReportKind, notes string) (workspace.Draft, error) {
		got = notes
		return workspace.Draft{Title: "Generated", Body: "Description"}, nil
	}})
	defer s.Close()
	s.Notes = "selected notes"
	s.What = "must not be included automatically"
	s.Generate()
	finish(t, s)
	if got != "selected notes" || s.Title != "Generated" {
		t.Fatal(got, s.Title)
	}
}
func TestRedactionBoxes(t *testing.T) {
	r, e := ParseRedactionBoxes("1,2,30,40; 5,6,7,8")
	if e != nil || len(r) != 2 || r[0].Max.X != 31 {
		t.Fatal(r, e)
	}
	for _, bad := range []string{"1,2,3", "-1,2,3,4", "1,2,0,4", "x,2,3,4"} {
		if _, e := ParseRedactionBoxes(bad); e == nil {
			t.Fatal(bad)
		}
	}
}
func TestNativeReportFormRenders(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	s := New(Options{Version: "test"})
	defer s.Close()
	tt := ui.NewTester(func(c *ui.Context) { ui.Scroll(c).Fill().Children(func() { View(c, s).Width(640) }) }, 700, 900)
	tt.Frame()
	if len(tt.Texts()) == 0 {
		t.Fatal("empty report UI")
	}
	if dir := os.Getenv("ORBIT_TEST_IMAGES"); dir != "" {
		if e := os.MkdirAll(dir, 0755); e != nil {
			t.Fatal(e)
		}
		save := func(name string) {
			f, e := os.Create(filepath.Join(dir, name))
			if e != nil {
				t.Fatal(e)
			}
			defer f.Close()
			if e = png.Encode(f, tt.Image()); e != nil {
				t.Fatal(e)
			}
		}
		save("bug-report-native-form.png")
		s.Title = "Native rendering fixture"
		s.What = "This is a local fixture. No issue is submitted."
		s.Prepare()
		finish(t, s)
		tt.Frame()
		save("bug-report-native-review.png")
	}

}

func TestClearCancelsStaleDraftResult(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})
	s := New(Options{GenerateDraft: func(ctx context.Context, _ native.ReportKind, _ string) (workspace.Draft, error) {
		close(started)
		<-ctx.Done()
		close(stopped)
		return workspace.Draft{Title: "stale"}, nil
	}})
	defer s.Close()
	s.Notes = "old request"
	s.Generate()
	<-started
	s.Clear()
	<-stopped
	s.Tick()
	if s.Title != "" || s.Busy {
		t.Fatal("cancelled worker changed cleared form")
	}
}

type filedReportGitHub struct{ calls int }

func (f *filedReportGitHub) Probe(context.Context) (workspace.GitHubStatus, error) {
	return workspace.GitHubStatus{Installed: true, Authenticated: true}, nil
}
func (f *filedReportGitHub) CurrentUser(context.Context) (workspace.GitHubUser, error) {
	return workspace.GitHubUser{}, nil
}
func (f *filedReportGitHub) CreateIssue(context.Context, workspace.CreateIssueOptions) (string, error) {
	f.calls++
	return "https://github.com/charlzyx/orbit/issues/42", nil
}
func TestSuccessfulSubmitClearsOnlyAfterExplicitReview(t *testing.T) {
	gh := &filedReportGitHub{}
	opened := ""
	s := New(Options{GitHub: gh, OpenURL: func(url string) error { opened = url; return nil }})
	defer s.Close()
	s.Title = "Test fixture"
	s.What = "No real report is submitted"
	s.Prepare()
	finish(t, s)
	if gh.calls != 0 || opened != "" {
		t.Fatal("preparation sent a report")
	}
	s.reviewed = true
	s.Submit()
	finish(t, s)
	if gh.calls != 1 || s.Title != "" || s.What != "" || !strings.Contains(s.Status, "Issue created") || opened != "https://github.com/charlzyx/orbit/issues/42" {
		t.Fatal(gh.calls, s.Title, s.Status, opened)
	}
}

func TestHasModalTracksReviewAndActualScreenshotPreview(t *testing.T) {
	var absent *State
	if absent.HasModal() {
		t.Fatal("nil reports modal")
	}
	s := New(Options{})
	defer s.Close()
	if s.HasModal() {
		t.Fatal("ordinary form is not modal")
	}
	s.reviewOpen = true
	if !s.HasModal() {
		t.Fatal("review bypasses menu guard")
	}
	s.reviewOpen = false
	s.preview = 0
	if s.HasModal() {
		t.Fatal("invalid preview should not block")
	}
	s.shots = []screenshot{{}}
	if !s.HasModal() {
		t.Fatal("screenshot preview bypasses menu guard")
	}
	s.preview = -1
	if s.HasModal() {
		t.Fatal("dismissed preview kept modal guard")
	}
}
