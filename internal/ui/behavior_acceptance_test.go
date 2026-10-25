package orbitui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSourceWindowRestoreBounds(t *testing.T) {
	for _, v := range []struct{ w, h, x, y int }{{0, 0, 1240, 840}, {1920, 1080, 1440, 920}, {1280, 720, 1088, 648}} {
		w, h := windowRestoreSize(v.w, v.h)
		if w != v.x || h != v.y {
			t.Fatalf("%dx%d => %dx%d", v.w, v.h, w, h)
		}
	}
}
func TestTranscriptVariableHeightFollowAndReaderPosition(t *testing.T) {
	a := fixture(t)
	for i := 0; i < 120; i++ {
		a.Model.Messages = append(a.Model.Messages, Message{ID: fmt.Sprint(i), Role: "assistant", Text: fmt.Sprintf("Message %d\n\n", i) + strings.Repeat("A variable-height paragraph with enough words to wrap. ", 1+i%8)})
	}
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if !a.transcript.AtEnd() {
		t.Fatal("transcript did not start at latest")
	}
	for n := 0; n < 12; n++ {
		tt.Scroll(700, 400, 0, -30)
	}
	if a.transcript.AtEnd() {
		t.Fatal("reader could not leave end")
	}
	first, _ := a.transcript.Visible()
	a.Model.Messages = append(a.Model.Messages, Message{ID: "new", Role: "assistant", Text: strings.Repeat("Streamed content. ", 80), Streaming: true})
	tt.Frame()
	after, _ := a.transcript.Visible()
	if after != first || a.transcript.AtEnd() {
		t.Fatalf("stream stole reader position: %d => %d", first, after)
	}
	if e := tt.Click("Jump to latest"); e != nil {
		t.Fatal(e)
	}
	if !a.transcript.AtEnd() {
		t.Fatal("jump did not resume follow")
	}
	a.Model.Messages[len(a.Model.Messages)-1].Text += strings.Repeat("Newly streamed tail. ", 40)
	tt.Frame()
	if !a.transcript.AtEnd() {
		t.Fatal("streaming growth lost follow")
	}
}
func TestComposerRepeatedEnterShiftAlternateAndRejection(t *testing.T) {
	a := fixture(t)
	var actions []Action
	reject := false
	a.Options.OnAction = func(x Action) {
		actions = append(actions, x)
		if x.Kind == "Submit" && reject {
			a.submitAccepted = false
		}
	}
	tt := ui.NewTester(a.View, 1200, 800)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Type("line one")
	tt.Key(ui.Shift, ui.KeyEnter)
	tt.Type("line two")
	if len(actions) != 0 || !strings.Contains(a.Draft, "\n") {
		t.Fatal("Shift Enter submitted or lost newline")
	}
	tt.Key(ui.Alt, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Data["mode"] != "steer" {
		t.Fatalf("alternate send: %+v", actions)
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 {
		t.Fatal("empty repeat Enter duplicated prompt")
	}
	reject = true
	tt.Type("keep failed draft")
	tt.Key(0, ui.KeyEnter)
	if a.Draft != "keep failed draft" {
		t.Fatal("rejected submit cleared draft")
	}
	reject = false
	tt.Key(0, ui.KeyEnter)
	if a.Draft != "" || len(actions) != 3 {
		t.Fatal("retry didn't submit retained text once")
	}
}
func TestLatencyPercentileRequiresSourceMinimum(t *testing.T) {
	if strings.HasPrefix(usageLatencyHint(5, 19), "p95 ") {
		t.Fatal("percentile invented with insufficient sample")
	}
	if !strings.HasPrefix(usageLatencyHint(5, 20), "p95 ") {
		t.Fatal("valid percentile omitted")
	}
}

func TestTranscriptFindKeyboardWrapTrimAndFocus(t *testing.T) {
	a := fixture(t)
	a.Model.Messages = []Message{{ID: "a", Role: "assistant", Text: "Alpha hit"}, {ID: "b", Role: "assistant", Text: "Not a match"}, {ID: "c", Role: "toolResult", ToolName: "read", ToolOutput: "Second HIT"}}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Key(ui.Cmd, ui.KeyF)
	if !tt.Focused("Find in transcript") {
		t.Fatal("find did not take focus")
	}
	tt.Type(" hit ")
	if len(a.searchMatches()) != 2 || a.searchCursor != 0 {
		t.Fatal("trim/case/tool matches failed")
	}
	tt.Key(0, ui.KeyEnter)
	if a.searchCursor != 1 {
		t.Fatal("Enter did not select next hit")
	}
	tt.Key(0, ui.KeyEnter)
	if a.searchCursor != 0 {
		t.Fatal("next did not wrap")
	}
	tt.Key(ui.Shift, ui.KeyEnter)
	if a.searchCursor != 1 {
		t.Fatal("Shift Enter did not select previous")
	}
	tt.Key(0, ui.KeyEscape)
	if a.searchOpen || a.Search != "" || !tt.Focused("Message") {
		t.Fatalf("Escape did not close/restore composer: open=%v query=%q focus=%v actions=%+v", a.searchOpen, a.Search, tt.Focused("Message"), actions)
	}
	for _, x := range actions {
		if x.Kind == "AbortRun" || x.Kind == "Submit" {
			t.Fatal("find keys escaped into agent commands")
		}
	}
}
