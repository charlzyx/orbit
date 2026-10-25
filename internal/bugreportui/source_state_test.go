package bugreportui

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/workspace"
)

func waitResult(t *testing.T, s *State) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for s.Busy && time.Now().Before(deadline) {
		s.Tick()
		time.Sleep(time.Millisecond)
	}
	if s.Busy {
		t.Fatal("report worker did not return")
	}
}

func TestSourceRequiredFieldsAndKindTransitions(t *testing.T) {
	var calls atomic.Int32
	s := New(Options{GenerateDraft: func(context.Context, native.ReportKind, string) (workspace.Draft, error) {
		calls.Add(1)
		return workspace.Draft{}, nil
	}})
	defer s.Close()
	if s.Kind != native.ReportBug || !s.RedactSensitive {
		t.Fatal("source kind or retained privacy default changed")
	}
	s.Generate()
	if s.Error == nil || s.Error.Error() != "Add some context before generating a draft." || s.Busy || calls.Load() != 0 {
		t.Fatal("empty context did not fail locally", s.Error)
	}
	s.Title, s.What, s.Steps, s.Notes = "keep", "body", "steps", "notes"
	for _, test := range []struct {
		kind        native.ReportKind
		what, steps string
	}{{native.ReportFeature, "What problem does this solve?", "Proposed solution"}, {native.ReportOther, "Details", "Additional details"}, {native.ReportBug, "What happened?", "Steps to reproduce"}} {
		s.Error = errors.New("previous error")
		s.SelectKind(test.kind)
		what, steps := s.fieldLabels()
		if what != test.what || steps != test.steps || s.Error != nil || s.Title != "keep" || s.What != "body" || s.Steps != "steps" || s.Notes != "notes" {
			t.Fatal(test, what, steps, s.Error)
		}
	}
	s.Title = " "
	s.Prepare()
	if s.Error == nil || s.Error.Error() != "Add a title for the issue." || s.Busy || s.reviewOpen {
		t.Fatal("missing title not preserved as inline validation", s.Error)
	}
	s.Title, s.What = "valid", "\n"
	s.Prepare()
	if s.Error == nil || s.Error.Error() != "Describe what happened." || s.Busy || s.reviewOpen {
		t.Fatal("missing description not preserved as inline validation", s.Error)
	}
	s.SelectKind(native.ReportFeature)
	s.IncludeDiagnostics = true
	s.Clear()
	if s.Kind != native.ReportFeature || !s.IncludeDiagnostics || s.Error != nil || s.Title != "" || s.What != "" || s.Steps != "" || s.Notes != "" {
		t.Fatal("clear must keep report kind and privacy choices, and clear draft/error")
	}
}

func TestDraftFailureKeepsAllUserFields(t *testing.T) {
	s := New(Options{GenerateDraft: func(context.Context, native.ReportKind, string) (workspace.Draft, error) {
		return workspace.Draft{}, errors.New("fixture model unavailable")
	}})
	defer s.Close()
	s.Notes, s.Title, s.What, s.Steps = "input notes", "old title", "old body", "old steps"
	s.Generate()
	waitResult(t, s)
	if s.Error == nil || s.Title != "old title" || s.What != "old body" || s.Notes != "input notes" || s.Steps != "old steps" {
		t.Fatal("generation error lost report", s.Error)
	}
	s.Options.GenerateDraft = func(context.Context, native.ReportKind, string) (workspace.Draft, error) {
		return workspace.Draft{Title: "new title", Body: "new body"}, nil
	}
	s.Generate()
	finish(t, s)
	if s.Error != nil || s.Title != "new title" || s.What != "new body" || s.Notes != "input notes" || s.Steps != "old steps" {
		t.Fatal("retry did not replace only the generated fields")
	}
}

func TestCancelWaitsForWorkerAndRejectsRepeatedGeneration(t *testing.T) {
	started, cancelled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	s := New(Options{GenerateDraft: func(ctx context.Context, _ native.ReportKind, _ string) (workspace.Draft, error) {
		calls.Add(1)
		close(started)
		<-ctx.Done()
		close(cancelled)
		<-release
		return workspace.Draft{Title: "late title", Body: "late body"}, nil
	}})
	defer s.Close()
	s.Title, s.What, s.Notes = "keep title", "keep body", "explicit notes"
	s.Generate()
	<-started
	s.Generate()
	s.Cancel()
	<-cancelled
	s.Generate()
	if !s.Busy || !s.cancelling || calls.Load() != 1 {
		t.Fatal("cancellation released duplicate-operation gate early")
	}
	close(release)
	finish(t, s)
	if s.Busy || s.cancelling || s.Title != "keep title" || s.What != "keep body" || s.Notes != "explicit notes" || !strings.Contains(s.Status, "Cancelled") {
		t.Fatal("cancelled result replaced preserved report", s.Status)
	}
}

func imageFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(file, image.NewNRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSourceScreenshotCapDeduplicationAndPartialReadFailure(t *testing.T) {
	s := New(Options{})
	defer s.Close()
	dir := t.TempDir()
	one := imageFile(t, dir, "first.png")
	s.AddFiles([]string{one, one, filepath.Join(dir, "missing.png")})
	waitResult(t, s)
	if s.Error == nil || len(s.shots) != 1 {
		t.Fatal("valid selection discarded or duplicate kept", s.Error, len(s.shots))
	}
	var files []string
	for i := 0; i < 8; i++ {
		files = append(files, imageFile(t, dir, fmt.Sprintf("shot-%d.png", i)))
	}
	s.AddFiles(files)
	finish(t, s)
	if len(s.shots) != native.MaxReportScreenshots {
		t.Fatal("source selection must retain first six unique screenshots", len(s.shots))
	}
	s.AddFiles([]string{one})
	if s.Error == nil || s.Busy || len(s.shots) != 6 {
		t.Fatal("limit guard missing")
	}
	s.RemoveScreenshot(2)
	if len(s.shots) != 5 {
		t.Fatal("remove did not restore screenshot capacity")
	}
	s.AddFiles([]string{one})
	finish(t, s)
	if len(s.shots) != 5 {
		t.Fatal("existing screenshot was added again")
	}
	s.Options.PickScreenshots = func(context.Context) ([]string, error) { return nil, nil }
	s.pickScreenshots()
	finish(t, s)
	if len(s.shots) != 5 {
		t.Fatal("cancelled picker changed selected files")
	}
}

type guardedGitHub struct {
	probes, identities, creates atomic.Int32
	createError                 error
}

func (g *guardedGitHub) Probe(context.Context) (workspace.GitHubStatus, error) {
	g.probes.Add(1)
	return workspace.GitHubStatus{Installed: true, Authenticated: true}, nil
}
func (g *guardedGitHub) CurrentUser(context.Context) (workspace.GitHubUser, error) {
	g.identities.Add(1)
	return workspace.GitHubUser{Login: "fixture-user"}, nil
}
func (g *guardedGitHub) CreateIssue(context.Context, workspace.CreateIssueOptions) (string, error) {
	g.creates.Add(1)
	if g.createError != nil {
		return "", g.createError
	}
	return "https://github.com/charlzyx/orbit/issues/42", nil
}

func TestPreparingNeverTransmitsAndIdentityRevocationIsImmediate(t *testing.T) {
	gh := &guardedGitHub{}
	var opens atomic.Int32
	s := New(Options{GitHub: gh, OpenURL: func(string) error { opens.Add(1); return nil }})
	defer s.Close()
	s.Title, s.What = "private fixture", "notes"
	s.IncludeDiagnostics = true
	s.IncludeIdentity = false
	s.Diagnostics = native.Diagnostics{OrbitVersion: "fixture", GHUser: "stale-identity"}
	s.Prepare()
	finish(t, s)
	if strings.Contains(s.prepared.BrowserBody, "stale-identity") || !strings.Contains(s.prepared.BrowserBody, "fixture") {
		t.Fatal("identity opt-out did not scrub cached diagnostics")
	}
	s.Submit()
	if gh.probes.Load() != 0 || gh.identities.Load() != 0 || gh.creates.Load() != 0 || opens.Load() != 0 {
		t.Fatal("preparation or unconfirmed submit transmitted")
	}
	if !s.Escape() || s.HasModal() || s.Title != "private fixture" || s.What != "notes" {
		t.Fatal("review dismissal discarded report")
	}
	s.optionsOpen = true
	s.CaptureConsent = true
	if !s.Escape() || s.CaptureConsent || s.HasModal() {
		t.Fatal("privacy dismissal kept capture consent")
	}
}

func TestSubmissionFailurePreservesReportAndRequiresFreshReview(t *testing.T) {
	gh := &guardedGitHub{createError: errors.New("fixture upload failed")}
	var opens atomic.Int32
	s := New(Options{GitHub: gh, OpenURL: func(string) error { opens.Add(1); return nil }})
	defer s.Close()
	s.Title, s.What, s.Notes, s.Steps = "title", "body", "notes", "steps"
	s.Prepare()
	finish(t, s)
	s.reviewed = true
	s.Submit()
	waitResult(t, s)
	if s.Error == nil || s.Title != "title" || s.What != "body" || s.Notes != "notes" || s.Steps != "steps" || s.reviewOpen || gh.creates.Load() != 1 || opens.Load() != 0 {
		t.Fatal("failed submit did not preserve report")
	}
	s.Submit()
	if s.Busy || gh.creates.Load() != 1 {
		t.Fatal("failure allowed an implicit retry")
	}
	s.Prepare()
	finish(t, s)
	if s.reviewed {
		t.Fatal("retry reused prior confirmation")
	}
}
