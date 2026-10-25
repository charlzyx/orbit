package orbitui

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func TestReviewDockedUncommittedUsesOneWorkingTreeSnapshot(t *testing.T) {
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	if err := os.WriteFile(filepath.Join(root, "seed"), []byte("staged\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Run(context.Background(), "add", "seed"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "seed"), []byte("working\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a := fixture(t)
	a.Model.Workspace = root
	a.reviewState.source = workspace.ReviewUncommitted
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	s := &a.reviewState
	if s.err != "" {
		t.Fatal(s.err)
	}
	if len(s.data.Files) != 1 {
		t.Fatalf("docked source duplicated staged/unstaged file: %#v", s.data.Files)
	}
	for _, line := range s.data.Lines {
		if line.Kind == "section_header" {
			t.Fatal("embedded-only section in docked Review", line)
		}
	}
	if s.data.Additions != 1 || s.data.Deletions != 1 {
		t.Fatal("stats counted intermediate staged state", s.data.Additions, s.data.Deletions)
	}
}

func TestMountedReviewToolbarUsesTwoSourceRows(t *testing.T) {
	a, tt := mountedReview(t)
	head, ok := tt.Find("Review header")
	if !ok || head.H != 40 {
		t.Fatal("missing source40px header", head, ok)
	}
	bar, ok := tt.Find("Review toolbar")
	if !ok || bar.H != 40 || bar.Y != head.Y+40 {
		t.Fatal("missing second source40px toolbar", head, bar, ok)
	}
	for _, label := range []string{"Collapse all", "Unwrap lines", "Side-by-side view", "Review source"} {
		r, ok := tt.Find(label)
		if !ok || r.Y < bar.Y || r.Y+r.H > bar.Y+bar.H {
			t.Fatal("control not directly in toolbar", label, r, ok)
		}
	}
	if _, ok := tt.Find("Review options"); ok {
		t.Fatal("source has direct controls, not options menu")
	}
	if err := tt.Click("Collapse all"); err != nil {
		t.Fatal(err)
	}
	if len(a.reviewState.collapsed) != len(a.reviewState.data.Files) {
		t.Fatal("collapse all did not collapse every file")
	}
	if err := tt.Click("Expand all"); err != nil {
		t.Fatal(err)
	}
	if len(a.reviewState.collapsed) != 0 {
		t.Fatal("expand all did not clear collapsed files")
	}
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	if a.reviewState.wrap {
		t.Fatal("direct wrap action failed")
	}
	if err := tt.Click("Side-by-side view"); err != nil {
		t.Fatal(err)
	}
	if !a.reviewState.split {
		t.Fatal("direct split action failed")
	}
	for _, width := range []float32{379, 380, 439, 440} {
		a.ReviewWidth = width
		tt.Frame()
		_, stats := tt.Find("Review change totals")
		_, tree := tt.Find("Toggle review file tree")
		if stats != (width >= 380) || tree != (width >= 440) {
			t.Fatal("source responsive thresholds", width, stats, tree)
		}
	}
}

func TestMountedReviewSourcePickerDismissalAndDisabledTurn(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	escapeCalls := 0
	a.Options.Escape = func() bool { escapeCalls++; return true }
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	menu, ok := tt.Find("Review source menu")
	if !ok || menu.W != 200 {
		t.Fatal("source picker must be200px", menu, ok)
	}
	for _, label := range []string{"Last Turn", "Uncommitted", "Unstaged", "Staged", "Committed", "Branch"} {
		if _, ok := tt.Find("Review source " + label); !ok {
			t.Fatal("missing source choice", label)
		}
	}
	separators := 0
	for _, text := range tt.Texts() {
		if text == "Review source separator" {
			separators++
		}
	}
	if separators != 2 {
		t.Fatal("expected separators after Last Turn and Staged", separators)
	}
	if !tt.HasText("no turns yet") {
		t.Fatal("disabled Last Turn explanation missing")
	}
	if err := tt.Click("Review source Last Turn"); err != nil {
		t.Fatal(err)
	}
	if s.source == workspace.ReviewLastTurn || !s.sourceMenuOpen {
		t.Fatal("disabled source handled click")
	}
	tt.Key(0, ui.KeyEscape)
	if s.sourceMenuOpen || !a.ReviewVisible || escapeCalls != 0 {
		t.Fatal("Escape did not dismiss only source menu")
	}
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	if s.sourceMenuOpen {
		t.Fatal("dismissal gesture reopened menu before200ms")
	}
	tt.Advance(201 * time.Millisecond)
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	if !s.sourceMenuOpen {
		t.Fatal("new source gesture did not reopen picker")
	}
	// The trigger is outside its own popup: the down dismisses, the up must
	// not reopen. Test actual pointer dispatch, rather than only state helpers.
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	if s.sourceMenuOpen {
		t.Fatal("outside trigger click reopened dismissed menu")
	}
	tt.Advance(201 * time.Millisecond)
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	header, _ := tt.Find("Review header")
	tt.Press(header.X+10, header.Y+20)
	tt.Release(header.X+10, header.Y+20)
	if s.sourceMenuOpen || !a.ReviewVisible {
		t.Fatal("outside press failed to dismiss")
	}
}

func TestMountedReviewSourceChangeLoadsFreshAndPinnedTurnMenu(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	a.Model.LatestTurn = 2
	s.source, s.turn = workspace.ReviewLastTurn, 1
	s.dataIdentity = "captured turn1"
	// Hold the collection in its loading state to inspect the frame before
	// completion; source changes must clear the stale body immediately.
	s.loading = true
	tt.Frame()
	if !tt.HasText("Turn 1") {
		t.Fatal("historical source label followed newest turn")
	}
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Review selected source Last Turn"); ok {
		t.Fatal("latest-turn checkmark incorrectly selected historical turn")
	}
	if err := tt.Click("Review source Last Turn"); err != nil {
		t.Fatal(err)
	}
	if s.source != workspace.ReviewLastTurn || s.turn != 2 || len(s.data.Lines) != 0 {
		t.Fatal("latest source click failed to replace historical identity", s.turn, len(s.data.Lines))
	}
	if !tt.HasText("Loading changes…") || tt.HasText("No changes") {
		t.Fatal("fresh in-flight source painted No changes instead of loading")
	}
	s.loading = false
	s.dataIdentity = "empty latest turn"
	tt.Frame()
	if !tt.HasText("No changes") || !tt.HasText("This turn didn't change any files.") {
		t.Fatal("settled empty snapshot lacks source description")
	}
	s.loading = true
	tt.Frame()
	if tt.HasText("Loading changes…") {
		t.Fatal("refresh of loaded empty snapshot incorrectly discarded current result")
	}
}

func TestMountedReviewSingleRefreshGlyphAndReducedMotion(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	a.Prefs.ReduceMotion = false
	s.loading = true
	tt.Frame()
	button, _ := tt.Find("Refresh review")
	glyph, ok := tt.Find("Refreshing review")
	if !ok || glyph.X < button.X || glyph.Y < button.Y || glyph.X+glyph.W > button.X+button.W || glyph.Y+glyph.H > button.Y+button.H {
		t.Fatal("feedback is not the button's own glyph", button, glyph, ok)
	}
	pixels := func() []byte {
		im := tt.Image()
		var out []byte
		for y := int(button.Y); y < int(button.Y+button.H); y++ {
			for x := int(button.X); x < int(button.X+button.W); x++ {
				v := im.RGBAAt(x, y)
				out = append(out, v.R, v.G, v.B, v.A)
			}
		}
		return out
	}
	before := pixels()
	tt.Advance(140 * time.Millisecond)
	if bytes.Equal(before, pixels()) {
		t.Fatal("refresh glyph did not rotate")
	}
	if r, _ := tt.Find("Refresh review"); r != button {
		t.Fatal("rotation reflowed refresh button", button, r)
	}
	a.Prefs.ReduceMotion = true
	tt.Frame()
	before = pixels()
	tt.Advance(140 * time.Millisecond)
	if !bytes.Equal(before, pixels()) {
		t.Fatal("reduced motion still rotates refresh glyph")
	}
	s.loading = false
	s.refreshSpinUntil = time.Time{}
	tt.Frame()
	if _, ok := tt.Find("Refreshing review"); ok {
		t.Fatal("idle glyph remained active")
	}
	if _, ok := tt.Find("Review refresh glyph"); !ok {
		t.Fatal("idle control swapped away from refresh glyph")
	}
}

func TestMountedReviewPinnedTurnSurvivesNewCompletion(t *testing.T) {
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	ctx := context.Background()
	capture := func(turn int, path string) {
		t.Helper()
		if err := g.CaptureTurnStart(ctx, "mounted-history", turn); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(path+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := g.CaptureTurn(ctx, "mounted-history", turn); err != nil {
			t.Fatal(err)
		}
	}
	capture(1, "first.txt")
	a := fixture(t)
	a.SetSnapshot(Snapshot{Workspace: root, SessionID: "mounted-history", LatestTurn: 1})
	a.OpenTurnReview(1)
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if !tt.HasText("Last Turn") {
		t.Fatal("newest completed turn should be Last Turn")
	}
	capture(2, "second.txt")
	a.SetSnapshot(Snapshot{Workspace: root, SessionID: "mounted-history", LatestTurn: 2})
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt.Frame()
	if !tt.HasText("Turn 1") || a.reviewState.data.Files[0].Path != "first.txt" {
		t.Fatal("new completion retargeted pinned review")
	}
	if err := tt.Click("Refresh review"); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if a.reviewState.turn != 1 || len(a.reviewState.data.Files) != 1 || a.reviewState.data.Files[0].Path != "first.txt" {
		t.Fatal("pointer refresh retargeted pinned review")
	}
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Review selected source Last Turn"); ok {
		t.Fatal("historical turn checked latest source")
	}
	if err := tt.Click("Review source Last Turn"); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt.Frame()
	if a.reviewState.turn != 2 || len(a.reviewState.data.Files) != 1 || a.reviewState.data.Files[0].Path != "second.txt" {
		t.Fatal("pointer source selection did not resolve latest checkpoint", a.reviewState.err, a.reviewState.data.Files)
	}
}

// Optional CPU-rendered fixtures. These are not native display captures and
// cannot establish cross-platform or pixel-perfect GPUI parity.
func TestReviewToolbarCPUEvidence(t *testing.T) {
	out := os.Getenv("ORBIT_REVIEW_CPU_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_REVIEW_CPU_CAPTURE_DIR for CPU-only review fixtures")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	a, tt := mountedReview(t)
	save := func(name string) {
		t.Helper()
		tt.Frame()
		f, err := os.Create(filepath.Join(out, name+"-cpu-only.png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	save("review-460-toolbar")
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	save("review-460-source-menu-no-turns")
	tt.Key(0, ui.KeyEscape)
	a.ReviewWidth = 379
	save("review-379-compact-toolbar")
	a.ReviewWidth = 460
	a.reviewState.clearReviewSnapshot()
	a.reviewState.loading = true
	save("review-fresh-loading")
	a.reviewState.loading = false
	a.reviewState.dataIdentity = "empty"
	save("review-settled-empty")
}

func TestMountedReviewEmptyToolbarDoesNotReflowOrReloadSameSource(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	s.clearReviewSnapshot()
	s.source = workspace.ReviewUncommitted
	s.dataIdentity = "loaded empty"
	tt.Frame()
	before, ok := tt.Find("Collapse all")
	if !ok {
		t.Fatal("empty snapshot removed collapse control")
	}
	if err := tt.Click("Collapse all"); err != nil {
		t.Fatal(err)
	}
	if len(s.collapsed) != 0 {
		t.Fatal("disabled collapse control changed state")
	}
	generation := s.generation
	if err := tt.Click("Review source"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Review source Uncommitted"); err != nil {
		t.Fatal(err)
	}
	if s.generation != generation || s.loading || s.dataIdentity != "loaded empty" {
		t.Fatal("same source selection unnecessarily reloaded/cleared snapshot")
	}
	if after, ok := tt.Find("Collapse all"); !ok || after != before {
		t.Fatal("disabled toolbar reflowed", before, after, ok)
	}
}

func TestMountedReviewTruncationDisclosureLivesInToolbar(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.data.Truncated = true
	for _, width := range []float32{300, 379, 460} {
		a.ReviewWidth = width
		tt.Frame()
		bar, _ := tt.Find("Review toolbar")
		warning, ok := tt.Find("Partial review diff")
		if !ok || warning.Y < bar.Y || warning.Y+warning.H > bar.Y+bar.H || warning.X < bar.X || warning.X+warning.W > bar.X+bar.W {
			t.Fatal("truncation warning not visible in toolbar", width, bar, warning, ok)
		}
		tt.Move(warning.X+warning.W/2, warning.Y+warning.H/2)
		tt.Advance(700 * time.Millisecond)
		if !tt.HasText("Partial diff: display limited to 50,000 lines.") {
			t.Fatal("truncation tooltip omitted limit", tt.Texts())
		}
	}
	a.reviewState.data.Truncated = false
	tt.Frame()
	if _, ok := tt.Find("Partial review diff"); ok {
		t.Fatal("partial warning remained after complete snapshot")
	}
}
