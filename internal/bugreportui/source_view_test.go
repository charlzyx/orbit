package bugreportui_test

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/bugreportui"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/preferences"
	orbitui "github.com/charlzyx/orbit/internal/ui"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type noTransmission struct{ calls atomic.Int32 }

func (g *noTransmission) Probe(context.Context) (workspace.GitHubStatus, error) {
	g.calls.Add(1)
	return workspace.GitHubStatus{}, nil
}
func (g *noTransmission) CurrentUser(context.Context) (workspace.GitHubUser, error) {
	g.calls.Add(1)
	return workspace.GitHubUser{}, nil
}
func (g *noTransmission) CreateIssue(context.Context, workspace.CreateIssueOptions) (string, error) {
	g.calls.Add(1)
	return "", nil
}

func sourceFixture(t *testing.T) (*bugreportui.State, *ui.Tester, *noTransmission) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if err := orbitui.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	store := preferences.New(t.TempDir())
	prefs := preferences.DefaultUI()
	prefs.Language = "en"
	prefs.ReduceMotion = true
	if err := store.SaveUI(prefs); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAppearance(preferences.Appearance{Mode: "dark", DarkTheme: "orbit-azure", LightTheme: "orbit-azure-light"}); err != nil {
		t.Fatal(err)
	}
	app := orbitui.New(orbitui.Options{Store: store, Version: "0.2.12-go-dev"})
	gh := &noTransmission{}
	s := bugreportui.New(bugreportui.Options{GitHub: gh, Home: store.Home, Version: "0.2.12-go-dev", ModelLabel: func() string { return "unknown" }, Theme: func() bugreportui.Theme {
		return bugreportui.Theme{SurfaceSecondary: app.Theme.SurfaceSecondary, Elevated: app.Theme.Elevated, TextSecondary: app.Theme.TextSecondary, Action: app.Theme.Action, ActionText: app.Theme.ActionFg, ActionHover: app.Theme.ActionHover, ActionPressed: app.Theme.ActionPressed, Px: app.Theme.Px, Space: app.Theme.Space}
	}, OpenURL: func(string) error { gh.calls.Add(1); return nil }})
	t.Cleanup(s.Close)
	app.Options.BugReport = func(c *ui.Context) { bugreportui.View(c, s).Fill() }
	app.Options.HasModal = s.HasModal
	tt := ui.NewTester(app.View, 1470, 922)
	tt.SetDark(true)
	app.Navigate("settings")
	tt.Frame()
	if err := tt.Click("Report a bug"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	return s, tt, gh
}

func settleReport(t *testing.T, s *bugreportui.State, tt *ui.Tester) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for s.Busy && time.Now().Before(deadline) {
		s.Tick()
		time.Sleep(time.Millisecond)
	}
	if s.Busy {
		t.Fatal("worker did not return")
	}
	tt.Frame()
}
func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
}
func near(t *testing.T, field string, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 1.1 {
		t.Errorf("%s: got %.2f, source %.2f", field, got, want)
	}
}

func TestSourceReportBoardGeometryAndFooterVisibility(t *testing.T) {
	s, tt, gh := sourceFixture(t)
	board, ok := tt.Find("Report form board")
	if !ok {
		t.Fatal("source board missing", tt.Texts())
	}
	near(t, "board X", board.X, 399)
	near(t, "board Y", board.Y, 143)
	near(t, "board width", board.W, 912)
	near(t, "board height", board.H, 676)
	for _, label := range []string{"Issue type", "Bug report", "Feature request", "Other", "Context", "Using unknown", "Title", "What happened?", "Steps to reproduce", "Screenshots", "Generate draft", "Add screenshot", "Clear", "Create issue on GitHub"} {
		if !tt.HasText(label) {
			t.Error("missing source control", label)
		}
	}
	for label, wantY := range map[string]float32{"Bug report": 156, "Generate draft": 209, "Add screenshot": 756, "Clear": 839, "Create issue on GitHub": 839} {
		r, ok := tt.Find(label)
		if !ok {
			t.Error("missing", label)
			continue
		}
		near(t, label+" Y", r.Y, wantY)
		near(t, label+" height", r.H, 28)
		if r.Y+r.H > 922 {
			t.Error("footer clipped", label, r)
		}
	}
	if tt.HasText("Redact recognizable secrets") || tt.HasText("Include build and runtime diagnostics") || tt.HasText("Capture this window") || tt.HasText("attached automatically") {
		t.Fatal("privacy options displaced source form or made an untrue promise")
	}
	if s.IncludeDiagnostics || s.IncludeIdentity || s.CaptureConsent || !s.RedactSensitive || gh.calls.Load() != 0 {
		t.Fatal("opening source form changed privacy/transmitted data")
	}
	if dir := os.Getenv("ORBIT_TEST_IMAGES"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, "bug-report-source-form-cpu.png"))
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		f.Close()
		var rows []struct {
			Label string
			Rect  ui.Rect
		}
		for _, label := range []string{"Report form board", "Issue type", "Context", "Title", "What happened?", "Steps to reproduce", "Screenshots", "Using unknown", "Bug report", "Feature request", "Other", "Generate draft", "Add screenshot", "Clear", "Create issue on GitHub"} {
			r, _ := tt.Find(label)
			rows = append(rows, struct {
				Label string
				Rect  ui.Rect
			}{label, r})
		}
		data, _ := json.MarshalIndent(rows, "", "  ")
		if err = os.WriteFile(filepath.Join(dir, "bug-report-source-geometry.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSourceReportPointerKeyboardValidationAndReview(t *testing.T) {
	s, tt, gh := sourceFixture(t)
	click(t, tt, "Generate draft")
	if s.Error == nil || s.Error.Error() != "Add some context before generating a draft." {
		t.Fatal("empty generate did not show source error")
	}
	click(t, tt, "Feature request")
	saveFixture(t, tt, "bug-report-feature-cpu")
	if s.Kind != native.ReportFeature || s.Error != nil || !tt.HasText("What problem does this solve?") || !tt.HasText("Proposed solution") {
		t.Fatal("feature segment did not change source labels")
	}
	// Pointer selects Other; keyboard Space activates the focused real button.
	click(t, tt, "Other")
	s.SelectKind(native.ReportBug)
	tt.Frame()
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	if s.Kind != native.ReportOther || !tt.HasText("Additional details") {
		t.Fatal("kind segment is not keyboard operable")
	}
	click(t, tt, "Bug report")
	contextRect, _ := tt.Find("Context")
	tt.ClickAt(contextRect.X+80, contextRect.Y+48)
	tt.Type("A local fixture crashes.\nOnly my notes are used.")
	tt.Frame()
	if !strings.Contains(s.Notes, "local fixture") {
		t.Fatal("whole multiline field did not accept pointer/typing", s.Notes)
	}
	tt.Key(0, ui.KeyTab)
	tt.Type("Keyboard title")
	tt.Frame()
	if s.Title != "Keyboard title" {
		t.Fatal("Tab from context did not focus Title", s.Title)
	}
	click(t, tt, "Generate draft")
	settleReport(t, s, tt)
	if s.Title != "A local fixture crashes" || !strings.Contains(s.What, "Only my notes") {
		t.Fatal("local fallback failed", s.Title, s.What)
	}
	click(t, tt, "Create issue on GitHub")
	settleReport(t, s, tt)
	if !s.HasModal() || !tt.HasText("Review public GitHub report") || gh.calls.Load() != 0 {
		t.Fatal("review did not gate external transmission")
	}
	saveFixture(t, tt, "bug-report-review-cpu")
	click(t, tt, "Copy reviewed report")
	if !strings.Contains(tt.Clipboard(), "A local fixture crashes") || gh.calls.Load() != 0 {
		t.Fatal("local review copy was not local")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.HasModal() || s.Title == "" || s.What == "" {
		t.Fatal("Escape did not dismiss review without clearing fields")
	}
	click(t, tt, "Report privacy options")
	if !s.HasModal() || !tt.HasText("Redact recognizable secrets and my home path") || gh.calls.Load() != 0 {
		t.Fatal("privacy surface missing or contacted GitHub")
	}
	saveFixture(t, tt, "bug-report-privacy-cpu")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.HasModal() {
		t.Fatal("Escape did not dismiss privacy surface")
	}
	click(t, tt, "Clear")
	if s.Title != "" || s.Notes != "" || s.What != "" || s.Error != nil || gh.calls.Load() != 0 {
		t.Fatal("clear lost privacy or sent a report")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if tt.HasText("Report form board") {
		t.Fatal("idle report consumed Escape instead of letting Settings close")
	}
}

func TestSelectedModelCaptionTracksActualModel(t *testing.T) {
	s, tt, _ := sourceFixture(t)
	for _, model := range []string{"Claude fixture", "…", ""} {
		s.Options.ModelLabel = func() string { return model }
		tt.Frame()
		if model == "Claude fixture" {
			if !tt.HasText("Using Claude fixture") {
				t.Fatal("active model caption missing")
			}
		} else if tt.HasText("Using ") {
			t.Fatal("source hides empty/loading model caption", model)
		}
	}
}

func saveFixture(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	if dir := os.Getenv("ORBIT_TEST_IMAGES"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReportScreenshotPointerPreviewRedactionAndLimit(t *testing.T) {
	s, tt, gh := sourceFixture(t)
	dir := t.TempDir()
	var paths []string
	for i := 0; i < 7; i++ {
		path := filepath.Join(dir, fmt.Sprintf("fixture-%d.png", i))
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(f, fixtureImage()); err != nil {
			t.Fatal(err)
		}
		f.Close()
		paths = append(paths, path)
	}
	var picks atomic.Int32
	s.Options.PickScreenshots = func(context.Context) ([]string, error) { picks.Add(1); return paths, nil }
	click(t, tt, "Add screenshot")
	settleReport(t, s, tt)
	if s.ScreenshotCount() != 6 || s.Error != nil || picks.Load() != 1 {
		t.Fatal("picker did not retain source cap", s.ScreenshotCount(), s.Error)
	}
	click(t, tt, "Add screenshot")
	if picks.Load() != 1 {
		t.Fatal("six-image cap button allowed a second picker")
	}
	saveFixture(t, tt, "bug-report-screenshots-cpu")
	click(t, tt, "Preview / redact fixture-0.png")
	if !s.HasModal() || !tt.HasText("Redaction boxes") {
		t.Fatal("screenshot preview inaccessible")
	}
	saveFixture(t, tt, "bug-report-screenshot-preview-cpu")
	click(t, tt, "Redaction boxes: x,y,width,height; x,y,width,height")
	tt.Type("1,1,3,3")
	tt.Frame()
	click(t, tt, "Apply redaction")
	settleReport(t, s, tt)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	click(t, tt, "Done")
	if s.HasModal() {
		t.Fatal("preview dismissal did not restore form")
	}
	click(t, tt, "Remove fixture-0.png")
	if s.ScreenshotCount() != 5 || gh.calls.Load() != 0 {
		t.Fatal("remove failed or screenshots implicitly transmitted")
	}
}

func TestReportGeneratingCancelAndRepeatedClick(t *testing.T) {
	s, tt, gh := sourceFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	s.Options.GenerateDraft = func(ctx context.Context, _ native.ReportKind, _ string) (workspace.Draft, error) {
		calls.Add(1)
		close(started)
		<-ctx.Done()
		<-release
		return workspace.Draft{Title: "late"}, nil
	}
	s.Notes, s.Title, s.What = "explicit notes", "keep title", "keep body"
	tt.Frame()
	click(t, tt, "Generate draft")
	<-started
	if !tt.HasText("Generating…") {
		t.Fatal("source loading label missing")
	}
	click(t, tt, "Generating…")
	if calls.Load() != 1 {
		t.Fatal("repeated generate started a second request")
	}
	saveFixture(t, tt, "bug-report-generating-cpu")
	click(t, tt, "Cancel")
	if !s.Busy {
		t.Fatal("cancel unlocked report before worker drained")
	}
	close(release)
	settleReport(t, s, tt)
	if s.Title != "keep title" || s.What != "keep body" || s.Busy || !tt.HasText("Cancelled") || gh.calls.Load() != 0 {
		t.Fatal("cancelled generation changed report or transmitted", s.Status)
	}
}

func fixtureImage() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	return img
}
