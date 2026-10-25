package orbitui

import (
	"encoding/json"
	"fmt"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func railFixture(t *testing.T) *App {
	a := fixture(t)
	t.Cleanup(func() { waitNavigationHintWrite(t, a) })
	a.Model.Messages = nil
	for i := 0; i < 12; i++ {
		a.Model.Messages = append(a.Model.Messages, Message{ID: fmt.Sprintf("u%d", i), Role: "user", Text: fmt.Sprintf("Prompt %d", i)}, Message{ID: fmt.Sprintf("a%d", i), Role: "assistant", Text: fmt.Sprintf("Response %d ", i) + strings.Repeat("long response text. ", 120)})
	}
	return a
}
func TestNavigationRailMountedClickAndHover(t *testing.T) {
	a := railFixture(t)
	tt := ui.NewTester(a.messages, 1200, 600)
	tt.Frame()
	if _, ok := tt.Find("Conversation navigation rail"); !ok {
		t.Fatal("missing source navigation rail")
	}
	if err := tt.Click("Go to turn 3"); err != nil {
		t.Fatal(err)
	}
	first, _ := a.transcript.Visible()
	if first != 4 {
		t.Fatal("did not jump to turn 3", first)
	}
	r, ok := tt.Find("Go to turn 4")
	if !ok {
		t.Fatal("missing tick")
	}
	tt.Move(r.X+10, r.Y+6)
	if _, ok := tt.Find("Turn preview"); !ok || !tt.HasText("Prompt 3") {
		t.Fatal("missing hovered turn preview", tt.Texts())
	}
	tt.Move(1100, 590)
	if _, ok := tt.Find("Turn preview"); ok {
		t.Fatal("preview leaked outside rail")
	}
}
func TestNavigationKeyboardMovesRelativeToCurrentUserTurn(t *testing.T) {
	a := railFixture(t)
	tt := ui.NewTester(a.messages, 1200, 600)
	tt.Frame()
	a.transcript.ScrollTo(7, ui.Start)
	tt.Frame()
	a.shortcutCommand("PrevTurn", "")
	tt.Frame()
	first, _ := a.transcript.Visible()
	if first != 4 {
		t.Fatal("previous must skip current user turn from assistant row", first)
	}
}
func TestNavigationRailVisibilityHintTTLAndPersistence(t *testing.T) {
	a := railFixture(t)
	_ = a.Options.Store.Update("hints.json", func(m map[string]json.RawMessage) error { m["other_hint"] = json.RawMessage("true"); return nil })
	tt := ui.NewTester(a.messages, 1039, 600)
	tt.Frame()
	if _, ok := tt.Find("Conversation navigation rail"); ok {
		t.Fatal("rail visible below source 1040 threshold")
	}
	tt.SetSize(1040, 600)
	if _, ok := tt.Find("Navigation rail hint"); !ok {
		t.Fatal("source hint missing")
	}
	tt.Advance(10*time.Second - time.Millisecond)
	if a.railHintDismissed {
		t.Fatal("hint expired early")
	}
	tt.Advance(time.Millisecond)
	if _, ok := tt.Find("Navigation rail hint"); ok || !a.railHintDismissed {
		t.Fatal("hint TTL")
	}
	waitNavigationHintWrite(t, a)
	var hints map[string]bool
	_ = a.Options.Store.Read("hints.json", &hints)
	if !hints["rail_hint_seen"] || !hints["other_hint"] {
		t.Fatal("hint persistence clobbered sibling", hints)
	}
	reopened := New(Options{Store: a.Options.Store})
	if !reopened.railHintDismissed {
		t.Fatal("hint forgot dismissal after reopen")
	}
	a.Model.Messages = []Message{{ID: "u", Role: "user", Text: "small"}, {ID: "a", Role: "assistant", Text: "answer"}, {ID: "u2", Role: "user", Text: "another"}}
	tt.Frame()
	if _, ok := tt.Find("Conversation navigation rail"); ok {
		t.Fatal("rail visible for non-scrollable transcript", a.transcriptScroll.MaxY)
	}
}
func TestNavigationRailHoverFanAndHintExclusive(t *testing.T) {
	a := railFixture(t)
	tt := ui.NewTester(a.messages, 1200, 600)
	tt.Frame()
	r, _ := tt.Find("Go to turn 6")
	tt.Move(r.X+5, r.Y+6)
	if _, ok := tt.Find("Navigation rail hint"); ok {
		t.Fatal("hint stacked with preview")
	}
	card, ok := tt.Find("Turn preview")
	if !ok || card.W != 320 || card.H > 126 {
		t.Fatal("source preview geometry", card)
	}
	for _, v := range []struct {
		n int
		w float32
	}{{6, 32}, {5, 32 * .68}, {7, 32 * .68}, {4, 32 * .44}, {8, 32 * .44}, {3, 8}} {
		r, ok := tt.Find(fmt.Sprintf("Turn %d tick", v.n))
		if !ok || math.Abs(float64(r.W-v.w)) > .01 {
			t.Fatalf("fan tick %d: %v want %v", v.n, r, v.w)
		}
	}
	if a.railHintDismissed {
		t.Fatal("hover should not permanently dismiss hint")
	}
	tt.Move(1100, 590)
	if _, ok := tt.Find("Navigation rail hint"); !ok {
		t.Fatal("hint did not return after hover")
	}
	if err := tt.Click("Navigation rail hint"); err != nil {
		t.Fatal(err)
	}
	if !a.railHintDismissed {
		t.Fatal("hint click did not dismiss")
	}
}
func TestNavigationRailScrollAndStreamingReader(t *testing.T) {
	a := railFixture(t)
	for i := 12; i < 100; i++ {
		a.Model.Messages = append(a.Model.Messages, Message{ID: fmt.Sprintf("u%d", i), Role: "user", Text: fmt.Sprintf("Prompt %d", i)}, Message{ID: fmt.Sprintf("a%d", i), Role: "assistant", Text: strings.Repeat("answer ", 90)})
	}
	tt := ui.NewTester(a.messages, 1200, 400)
	tt.Frame()
	rail, _ := tt.Find("Conversation navigation rail")
	if rail.H != 320 || a.navigation.scroll.Y <= 0 {
		t.Fatal("source cap/active autoscroll", rail, a.navigation.scroll.Y)
	}
	if _, ok := tt.Find("Rail top fade"); !ok {
		t.Fatal("missing top fade")
	}
	before, _ := a.transcript.Visible()
	railBefore := a.navigation.scroll.Y
	tt.Scroll(rail.X+20, rail.Y+50, 0, -100)
	railAfter := a.navigation.scroll.Y
	tt.Frame()
	if railAfter >= railBefore || a.navigation.scroll.Y != railAfter {
		t.Fatal("rail wheel did not stay independent", railBefore, railAfter, a.navigation.scroll.Y)
	}
	after, _ := a.transcript.Visible()
	if before != after {
		t.Fatal("rail wheel scrolled transcript", before, after)
	}
	a.transcript.ScrollTo(20, ui.Start)
	tt.Frame()
	if a.activeUserTurn() != 20 || a.transcript.AtEnd() {
		t.Fatal("reader active tick mismatch")
	}
	first, _ := a.transcript.Visible()
	a.Model.Messages[len(a.Model.Messages)-1].Streaming = true
	a.Model.Messages[len(a.Model.Messages)-1].Text += strings.Repeat("live tail ", 200)
	tt.Frame()
	after, _ = a.transcript.Visible()
	if first != after || a.activeUserTurn() != 20 || a.transcript.AtEnd() {
		t.Fatal("stream stole historical viewport", first, after)
	}
	if err := tt.Click("Jump to latest"); err != nil {
		t.Fatal(err)
	}
	if !a.transcript.AtEnd() || a.activeUserTurn() != 198 {
		t.Fatal("tail following not restored")
	}
}
func TestNavigationRailSessionResetAndSameLengthReplacement(t *testing.T) {
	a := railFixture(t)
	a.Model.SessionID = "one"
	tt := ui.NewTester(a.messages, 1200, 600)
	tt.Frame()
	r, _ := tt.Find("Go to turn 2")
	tt.Move(r.X+5, r.Y+6)
	if a.navigation.hovered == 0 {
		t.Fatal("fixture hover")
	}
	a.SetSnapshot(Snapshot{SessionID: "two", Messages: []Message{{ID: "fresh", Role: "assistant", Text: "fresh"}}})
	tt.Frame()
	if a.navigation.hovered != 0 || a.navigation.scroll.Y != 0 {
		t.Fatal("session navigation leaked")
	}
	if _, ok := tt.Find("Turn preview"); ok {
		t.Fatal("stale preview")
	}
	a.SetSnapshot(Snapshot{SessionID: "two", Messages: []Message{{ID: "changed", Role: "user", Text: "user"}}})
	if got := a.userTurns(); len(got) != 1 || got[0] != 0 {
		t.Fatal("same length replacement stale cache", got)
	}
}
func TestNavigationSnippetGraphemesAndTurnBoundaries(t *testing.T) {
	family := "👨‍👩‍👧‍👦"
	if got := railSnippet("  "+strings.Repeat(family, 101)+" \n ", 100); got != strings.Repeat(family, 100)+"…" {
		t.Fatal("split grapheme", got)
	}
	a := fixture(t)
	a.Model.Messages = []Message{{Role: "user", Text: "  prompt\n words "}, {Role: "assistant", Text: "  "}, {Role: "assistant", Text: " first\nanswer "}, {Role: "user", Text: "next"}, {Role: "assistant", Text: "later"}}
	p, r := a.railPreview(0)
	if p != "prompt words" || r != "first answer" {
		t.Fatal(p, r)
	}
	a.Model.Messages[2].Text = ""
	_, r = a.railPreview(0)
	if r != "" {
		t.Fatal("preview crossed next user turn", r)
	}
	if p, r := a.railPreview(99); p != "" || r != "" {
		t.Fatal("stale hover indexed transcript")
	}
}
func TestNavigationKeyboardTailAndEdges(t *testing.T) {
	a := railFixture(t)
	tt := ui.NewTester(a.messages, 1200, 600)
	tt.Frame()
	a.shortcutCommand("PrevTurn", "")
	tt.Frame()
	first, _ := a.transcript.Visible()
	if first != 20 {
		t.Fatal("tail previous did not choose prior turn", first)
	}
	a.transcript.ScrollTo(0, ui.Start)
	tt.Frame()
	a.shortcutCommand("PrevTurn", "")
	tt.Frame()
	first, _ = a.transcript.Visible()
	if first != 0 {
		t.Fatal("previous edge moved", first)
	}
	a.shortcutCommand("NextTurn", "")
	tt.Frame()
	first, _ = a.transcript.Visible()
	if first != 2 {
		t.Fatal("next", first)
	}
	if !a.railHintDismissed {
		t.Fatal("keyboard did not dismiss source hint")
	}
}
func TestNavigationRailKeepsRetainedTextSelection(t *testing.T) {
	a := railFixture(t)
	a.railHintDismissed = true
	tt := ui.NewTester(a.messages, 1200, 600)
	a.transcript.ScrollTo(0, ui.Start)
	tt.Frame()
	r, ok := tt.Find("Prompt 0")
	if !ok {
		t.Fatal("prompt not visible")
	}
	tt.Press(r.X+1, r.Y+r.H/2)
	tt.Move(r.X+r.W-1, r.Y+r.H/2)
	tt.Release(r.X+r.W-1, r.Y+r.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	want := tt.Clipboard()
	if !strings.Contains(want, "Prompt") {
		t.Fatal("selection fixture failed", want, r)
	}
	if err := tt.Click("Go to turn 10"); err != nil {
		t.Fatal(err)
	}
	tt.SetClipboard("sentinel")
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != want {
		t.Fatal("rail navigation lost retained selection", want, tt.Clipboard())
	}
	a.SetSnapshot(Snapshot{SessionID: "different", Messages: a.Model.Messages})
	tt.Frame()
	tt.SetClipboard("sentinel")
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != "sentinel" {
		t.Fatal("selection crossed sessions")
	}
}
func TestNavigationRailSearchAndPreviewLocalization(t *testing.T) {
	a := railFixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	tt := ui.NewTester(a.View, 1600, 900)
	tt.Frame()
	if _, ok := tt.Find("Conversation navigation rail"); !ok {
		t.Fatal("full App fixture rail missing")
	}
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Type("Prompt 3")
	if len(a.searchMatches()) != 1 {
		t.Fatal("find lost matches")
	}
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Message") {
		t.Fatal("search did not restore focus")
	}
	r, _ := tt.Find("Go to turn 4")
	tt.Move(r.X+5, r.Y+6)
	if _, ok := tt.Find("Turn preview"); !ok {
		t.Fatal("preview missing after search")
	}
	label := orbitassets.Translate(a.catalog, a.english, "transcript_view.turn_of_total_hint", map[string]string{"current": "4", "total": "12"})
	if !tt.HasText(label) {
		t.Fatal("preview locale", label, tt.Texts())
	}
}
func TestNavigationRailCapture(t *testing.T) {
	path := os.Getenv("ORBIT_NAVIGATION_CAPTURE")
	if path == "" {
		t.Skip("capture path not requested")
	}
	a := railFixture(t)
	tt := ui.NewTester(a.View, 1600, 900)
	tt.Frame()
	r, _ := tt.Find("Go to turn 6")
	tt.Move(r.X+10, r.Y+6)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

func TestNavigationRailHintDismissDoesNotWaitForPreferenceLock(t *testing.T) {
	a := railFixture(t)
	tt := ui.NewTester(a.messages, 1200, 600)
	tt.Frame()
	locked, release, writerDone := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		writerDone <- a.Options.Store.Update("hints.json", func(m map[string]json.RawMessage) error {
			close(locked)
			<-release
			m["concurrent_hint"] = json.RawMessage("true")
			return nil
		})
	}()
	<-locked
	clickedDone := make(chan struct{})
	var blocked atomic.Bool
	go func() {
		select {
		case <-clickedDone:
		case <-time.After(time.Second):
			blocked.Store(true)
		}
		close(release)
	}()
	err := tt.Click("Navigation rail hint")
	close(clickedDone)
	if e := <-writerDone; e != nil {
		t.Fatal(e)
	}
	if err != nil {
		t.Fatal(err)
	}
	if blocked.Load() {
		t.Fatal("native hint click waited for preference store mutex")
	}
	if !a.railHintDismissed {
		t.Fatal("dismiss presentation was not optimistic")
	}
	for i := 0; i < 10; i++ {
		a.dismissRailHint()
	}
	a.SetSnapshot(Snapshot{SessionID: "new-session", Messages: []Message{{ID: "new", Role: "user", Text: "new"}}})
	waitNavigationHintWrite(t, a)
	var hints map[string]bool
	_ = a.Options.Store.Read("hints.json", &hints)
	if !hints["rail_hint_seen"] || !hints["concurrent_hint"] {
		t.Fatal("lost concurrent preference key", hints)
	}
	if a.Model.SessionID != "new-session" || !a.railHintDismissed {
		t.Fatal("hint completion mutated session ownership")
	}
}

func waitNavigationHintWrite(t *testing.T, a *App) {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for a.railHintSaving {
		select {
		case apply := <-a.updates:
			apply()
		case <-timer.C:
			t.Fatal("hint persistence worker did not finish")
		}
	}
}
func TestNavigationRailHintWriteFailureIsAsyncGlobalFeedback(t *testing.T) {
	a := railFixture(t)
	if err := os.MkdirAll(a.Options.Store.RootDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a.Options.Store.RootDir, "hints.json"), []byte("malformed {"), 0600); err != nil {
		t.Fatal(err)
	}
	a.dismissRailHint()
	if !a.railHintDismissed || !a.railHintSaving {
		t.Fatal("dismiss did not enter optimistic saving state")
	}
	a.SetSnapshot(Snapshot{SessionID: "other-session"})
	waitNavigationHintWrite(t, a)
	if !a.railHintDismissed || a.railHintSaving || a.Model.SessionID != "other-session" {
		t.Fatal("failed write corrupted presentation/owner")
	}
	if len(a.toastStack.Items()) != 1 {
		t.Fatal("preference error not surfaced on UI completion", a.toastStack.Items())
	}
	raw, err := os.ReadFile(filepath.Join(a.Options.Store.RootDir, "hints.json"))
	if err != nil || string(raw) != "malformed {" {
		t.Fatal("malformed preference silently overwritten", err, string(raw))
	}
	a.dismissRailHint()
	if a.railHintSaving {
		t.Fatal("repeated dismissal retried failed IO unexpectedly")
	}
}
