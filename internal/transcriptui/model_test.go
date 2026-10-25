package transcriptui

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/pirpc"
)

// All examples below are explicit protocol fixtures, not live model results.
func fixtureMessage(t *testing.T, id, raw string) agent.Message {
	t.Helper()
	if !json.Valid([]byte(raw)) {
		t.Fatal("invalid fixture")
	}
	o := obj([]byte(raw))
	return agent.Message{ID: id, Role: str(o["role"]), ToolName: str(o["toolName"]), ToolCallID: str(o["toolCallId"]), Raw: json.RawMessage(raw), Usage: o["usage"]}
}
func TestProjectSourceLogicalTurnsAndBoundary(t *testing.T) {
	messages := []agent.Message{
		fixtureMessage(t, "u", `{"role":"user","content":"Test this"}`),
		fixtureMessage(t, "s", `{"role":"system","content":"not a chat row"}`),
		fixtureMessage(t, "a1", `{"role":"assistant","content":[{"type":"thinking","thinking":"Inspect."},{"type":"toolCall","id":"run-1","name":"bash","arguments":{"command":"go test ./..."}}],"usage":{"input":10,"output":2,"cacheRead":4,"cost":{"total":0.001}},"durationMs":1000}`),
		fixtureMessage(t, "tool", `{"role":"toolResult","toolCallId":"run-1","toolName":"bash","content":[{"type":"text","text":"ok"},{"type":"text","text":"pass"},{"type":"image","mimeType":"image/png","data":"secret fixture bytes"}],"details":{"truncation":{"truncated":true,"outputLines":2,"totalLines":9}}}`),
		fixtureMessage(t, "a2", `{"role":"assistant","content":[{"type":"text","text":"Done."}],"usage":{"input":3,"output":6,"cost":0.002},"durationMs":2000,"timestamp":"2026-10-25T05:00:00Z"}`),
		fixtureMessage(t, "b", `{"role":"compactionSummary","summary":"Earlier work","tokensBefore":90000}`),
		fixtureMessage(t, "a3", `{"role":"assistant","content":"After boundary"}`),
	}
	rows := Project(messages)
	if len(rows) != 4 {
		t.Fatalf("rows %d: %#v", len(rows), rows)
	}
	r := rows[1]
	if r.ID != "a1" || len(r.Steps) != 2 || r.Text() != "Done." {
		t.Fatalf("logical row %#v", r)
	}
	tool := r.Steps[0].Tools[0]
	if got := display(tool.Output, 100); got != "ok\npass" {
		t.Fatalf("output=%q", got)
	}
	if !tool.Complete || !tool.Facts.Truncated || *tool.Facts.TotalLines != 9 {
		t.Fatalf("tool facts %#v", tool)
	}
	if strings.Contains(r.SearchText(), "secret fixture") {
		t.Fatal("image base64 leaked into searchable output")
	}
	if u := r.Usage(); u.Input != 13 || u.Output != 8 || u.Total != 25 || u.Cost == nil || *u.Cost != .003 {
		t.Fatalf("usage %#v", u)
	}
	if d := r.GenerationTime(); d == nil || *d != 3*time.Second {
		t.Fatal("provider generation duration lost")
	}
	if r.Elapsed != nil {
		t.Fatal("snapshot fabricated elapsed wall clock")
	}
	if rows[2].Summary == nil || rows[2].Summary.Text != "Earlier work" || rows[2].Text() != "" {
		t.Fatal("boundary rendered as assistant answer")
	}
}
func TestTypedToolsAndOutputBounds(t *testing.T) {
	shell := newTool("run", "bash", json.RawMessage(`{"command":"printf 'hello'\nprintf 'world'"}`))
	if cmd, ok := shell.Command(); !ok || cmd != "printf 'hello'\nprintf 'world'" {
		t.Fatalf("command %q %v", cmd, ok)
	}
	if shell.Preview() != "printf 'hello' printf 'world'" {
		t.Fatal(shell.Preview())
	}
	edit := newTool("edit", "edit", json.RawMessage(`{"path":"main.go","edits":[{"oldText":"one\ntwo\n","newText":"one\nthree\n"},{"oldText":"x","newText":"y"}]}`))
	if edit.Added != 4 || edit.Removed != 4 {
		t.Fatal("source line count changed")
	}
	if got := DiffText(edit.Diff()); got != " one\n-two\n+three\n-x\n+y\n" {
		t.Fatalf("patch %q", got)
	}
	if len(edit.Sections()) != 0 {
		t.Fatal("diff tool redundantly shows raw arguments")
	}
	for _, n := range []int{1, 16, 17, 401} {
		s := strings.Repeat("line\n", n-1) + "last"
		w := SectionLines(s, false)
		expected := n
		if n > 16 {
			expected = 12
		}
		if len(w.Visible) != expected {
			t.Fatalf("%d preview %d", n, len(w.Visible))
		}
		expanded := SectionLines(s, true)
		if len(expanded.Visible) != min(n, 400) || expanded.Clipped != (n > 400) {
			t.Fatalf("%d expanded %#v", n, expanded)
		}
	}
	if got := snippet(strings.Repeat("👨‍👩‍👧‍👦", 73), 72); got != strings.Repeat("👨‍👩‍👧‍👦", 72)+"…" {
		t.Fatalf("preview split grapheme: bytes %d want %d; tail %q", len(got), len(strings.Repeat("👨‍👩‍👧‍👦", 72)+"…"), got[max(0, len(got)-60):])
	}
	partial := newTool("partial", "bash", json.RawMessage(`{"command":"incom`))
	if partial.Arguments != nil || partial.Preview() != "" {
		t.Fatal("rendered incomplete JSON as final arguments")
	}
}
func TestSourceFailuresAbortAndMissingMetrics(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"input":0,"output":0}`, `{"cost":null}`, `{"input":-1}`} {
		if ParseUsage(json.RawMessage(raw)) != nil {
			t.Fatalf("invented usage for %s", raw)
		}
	}
	rows := Project([]agent.Message{fixtureMessage(t, "a", `{"role":"assistant","content":"Partial","stopReason":"aborted","errorMessage":"Request was aborted"}`)})
	if len(rows) != 1 || !rows[0].Aborted || rows[0].Error != "" || rows[0].UsageLabel() != "" {
		t.Fatalf("aborted %#v", rows)
	}
	failed := Project([]agent.Message{fixtureMessage(t, "e", `{"role":"assistant","content":[],"stopReason":"error","errorMessage":"  Actual provider failure  "}`)})
	if failed[0].Error != "Actual provider failure" {
		t.Fatal("lost failure")
	}
}
func TestLiveToolEnvelopeUpdatesAreCumulativeNotDuplicated(t *testing.T) {
	base := fixtureMessage(t, "a", `{"role":"assistant","content":[{"type":"toolCall","id":"x","name":"bash","arguments":{"command":"test"}}]}`)
	for _, f := range []struct {
		raw, want       string
		running, failed bool
	}{
		{`{"type":"tool_execution_start","toolCallId":"x","toolName":"bash","args":{"command":"test"}}`, "", true, false},
		{`{"type":"tool_execution_update","toolCallId":"x","toolName":"bash","partialResult":{"content":[{"type":"text","text":"one"}]}}`, "one", true, false},
		{`{"type":"tool_execution_update","toolCallId":"x","toolName":"bash","partialResult":{"content":[{"type":"text","text":"one\ntwo"}]}}`, "one\ntwo", true, false},
		{`{"type":"tool_execution_end","toolCallId":"x","toolName":"bash","result":{"content":[{"type":"text","text":"exit 1"}]},"isError":true}`, "exit 1", false, true},
	} {
		e, err := pirpc.ParseEvent([]byte(f.raw))
		if err != nil {
			t.Fatal(err)
		}
		m := agent.Message{ID: "tool:x", Role: "tool", ToolCallID: e.String("toolCallId"), ToolName: e.String("toolName"), Streaming: f.running, Raw: e.Raw}
		for range 3 {
			rows := Project([]agent.Message{base, m})
			if len(rows) != 1 || len(rows[0].Steps[0].Tools) != 1 {
				t.Fatal("duplicate tool")
			}
			tool := rows[0].Steps[0].Tools[0]
			if display(tool.Output, 100) != f.want || tool.Running != f.running || tool.Failed != f.failed || tool.Arguments == nil {
				t.Fatalf("bad update %#v", tool)
			}
		}
	}
}
func TestSourcePostAnswerGroupingAndChangedFiles(t *testing.T) {
	rows := Project([]agent.Message{
		fixtureMessage(t, "a", `{"role":"assistant","content":"First answer"}`),
		fixtureMessage(t, "b", `{"role":"assistant","content":[{"type":"thinking","thinking":"Next"},{"type":"toolCall","id":"e","name":"edit","arguments":{"path":"a.go","edits":[{"oldText":"x","newText":"y"}]}}]}`),
		fixtureMessage(t, "c", `{"role":"assistant","content":[{"type":"toolCall","id":"w","name":"write","arguments":{"path":"a.go","content":"x\ny"}}]}`),
		fixtureMessage(t, "d", `{"role":"assistant","content":"Final answer"}`),
	})
	r := rows[0]
	want := []Part{{"text", 0, 1, false, false}, {"work", 1, 3, false, false}, {"text", 3, 4, false, false}}
	if !reflect.DeepEqual(r.Parts(), want) {
		t.Fatalf("parts %#v", r.Parts())
	}
	if got := r.ChangedFiles(); !reflect.DeepEqual(got, []ChangedFile{{"a.go", 3, 1}}) {
		t.Fatalf("files %#v", got)
	}
}
func TestStateSessionResetAndIndependentToggles(t *testing.T) {
	var s State
	s.Session("a")
	s.Toggle("tool:x", false)
	s.Toggle("tool:x:output", false)
	if !s.Open("tool:x", false) || !s.Open("tool:x:output", false) {
		t.Fatal("independent toggles lost")
	}
	s.Session("a")
	if !s.Open("tool:x", false) {
		t.Fatal("snapshot reset state")
	}
	s.Toggle("tool:x", false)
	if s.Open("tool:x", false) || !s.Open("tool:x:output", false) {
		t.Fatal("repeated toggle changed child")
	}
	s.Session("b")
	if s.Open("tool:x:output", false) {
		t.Fatal("stale state leaked across session")
	}
}

func TestObservedWallTimeDoesNotInventReloadTiming(t *testing.T) {
	var s State
	s.Session("s")
	now := time.Unix(1000, 0)
	history := []Row{{ID: "old", Role: "assistant", Steps: []Step{{Text: "old"}}}}
	if got := s.Observe(history, false, now); got[0].Elapsed != nil {
		t.Fatal("historical clock invented")
	}
	live := append(history, Row{ID: "new", Role: "assistant", Streaming: true})
	got := s.Observe(live, true, now)
	if got[1].Elapsed == nil || *got[1].Elapsed != 0 {
		t.Fatal("live clock not started")
	}
	live[1].Streaming = false // the model step ended but its tool/roundtrip continues
	got = s.Observe(live, true, now.Add(4*time.Second))
	if !got[1].Streaming || *got[1].Elapsed != 4*time.Second {
		t.Fatal("tool gap prematurely ended turn")
	}
	got = s.Observe(live, false, now.Add(7*time.Second))
	if got[1].Streaming || *got[1].Elapsed != 7*time.Second {
		t.Fatal("settled clock failed")
	}
	got = s.Observe(live, false, now.Add(time.Hour))
	if *got[1].Elapsed != 7*time.Second {
		t.Fatal("repeated snapshot inflated duration")
	}
	s.Session("other")
	got = s.Observe(live, false, now.Add(time.Hour))
	if got[1].Elapsed != nil {
		t.Fatal("session clock leaked")
	}
}

func TestSourceMetricFormattingAndDateBoundaries(t *testing.T) {
	for n, want := range map[uint64]string{0: "0", 999: "999", 1000: "1K", 1050: "1.1K", 999950: "1000K", 1000000: "1M", 1550000: "1.6M"} {
		if got := FormatTokens(n); got != want {
			t.Fatalf("tokens %d: %s want %s", n, got, want)
		}
	}
	if hitPercent(99.99) != "99.9" || hitPercent(50) != "50" {
		t.Fatal("rounded partial cache hit to 100%")
	}
	if formatCost(.0012) != "$0.0012" || formatCost(12.345) != "$12.35" {
		t.Fatal("source cost precision")
	}
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.Local)
	for _, tc := range []struct {
		at   time.Time
		want string
	}{{now, "Today 12:00 PM"}, {now.AddDate(0, 0, -1), "Yesterday 12:00 PM"}, {now.AddDate(0, 0, -3), "Oct 22"}} {
		if got := SummaryTimestamp(tc.at, now, Options{}); got != tc.want {
			t.Fatalf("timestamp %q", got)
		}
	}
}
func TestQuestionMalformedFallbackAndExactAnswer(t *testing.T) {
	for _, args := range []string{`{}`, `{"questions":[{}]}`, `{"questions":"bad"}`} {
		if qs := newTool("x", "ask_user_question", json.RawMessage(args)).Questions(); len(qs) != 0 {
			t.Fatal("malformed questions not generic")
		}
	}
	q := newTool("x", "ask_user_question", json.RawMessage(`{"questions":[{"question":"Q","options":[{"label":"A"},{"label":"B"}]}]}`))
	q.Output, _ = json.Marshal(`User has answered your questions: "Q"="A, B".`)
	if answer := q.QuestionAnswer("Q"); answer != "A, B" || !optionAnswer(q.Questions()[0], answer) {
		t.Fatalf("answer %q", answer)
	}
	if q.QuestionAnswer("Other") != "" {
		t.Fatal("answer matched wrong question")
	}
}

func TestSnapshotToolImagesAndUserMetadataStayOutOfText(t *testing.T) {
	tool := ToolFromMessage(fixtureMessage(t, "t", `{"role":"toolResult","toolCallId":"x","content":[{"type":"image","data":"not shown","mimeType":"image/png"}]}`))
	if tool.Output != nil {
		t.Fatal("image-only snapshot leaked a protocol envelope")
	}
	for _, tc := range []struct{ in, out string }{
		{"Look at this\n\n[Image: original 400x400 resized]", "Look at this"},
		{"Normal\n\n[Image omitted: too large]", "Normal"},
		{`<skill name="review" location="/tmp/SKILL.md">internal body</skill> Please review`, "/skill:review Please review"},
		{"<skill unfinished ordinary prose", "<skill unfinished ordinary prose"},
	} {
		if got := compactUserPrompt(tc.in); got != tc.out {
			t.Fatalf("prompt %q want %q", got, tc.out)
		}
	}
}

func TestMeasuredAndEstimatedThinkingClocksAndRates(t *testing.T) {
	base := time.Unix(1000, 0)
	gen, measured := 2*time.Second, 1500*time.Millisecond
	msg := fixtureMessage(t, "clock", `{"role":"assistant","content":[{"type":"thinking","thinking":"fixture reasoning"},{"type":"text","text":"fixture answer"}],"usage":{"input":2,"output":12},"durationMs":4000}`)
	msg.GenerationTime = &gen
	msg.ThinkingDuration = &measured
	msg.RequestStartedAt = base
	msg.Streaming = true
	row := Project([]agent.Message{msg})[0]
	if d := row.GenerationTime(); d == nil || *d != 4*time.Second {
		t.Fatal("Pi duration did not override client fallback")
	}
	if d, estimated := StepThinkingDuration(row.Steps, 0); d == nil || *d != measured || estimated {
		t.Fatal("measured thinking clock lost")
	}
	if row.LiveRate(base.Add(3*time.Second)) != "4.0" {
		t.Fatal("live rate did not use actual request clock")
	}
	*row.Steps[0].ThinkingDuration = 99 * time.Second
	if *msg.ThinkingDuration != 1500*time.Millisecond {
		t.Fatal("projection aliased clock pointer")
	}
	rows := []Step{{Thinking: "historical reason", Timestamp: base}, {Timestamp: base.Add(8 * time.Second)}}
	d, estimated := StepThinkingDuration(rows, 0)
	if d == nil || *d != 8*time.Second || !estimated {
		t.Fatal("source adjacent-timestamp estimate missing")
	}
	rows[1].Timestamp = base.Add(-time.Second)
	if d, _ := StepThinkingDuration(rows, 0); d != nil {
		t.Fatal("negative timestamp gap became duration")
	}
	noUsage := Row{Streaming: true, Steps: []Step{{Text: strings.Repeat("chars are not reported tokens", 30), RequestStartedAt: base}}}
	if noUsage.LiveRate(base.Add(3*time.Second)) != "" {
		t.Fatal("fabricated token rate from character count")
	}
	row.Steps[0].Duration = nil
	if d := row.GenerationTime(); d == nil || *d != gen {
		t.Fatal("measured fallback not used when provider omitted duration")
	}
	row.Steps = append(row.Steps, Step{RequestStartedAt: base.Add(8 * time.Second), Usage: &Usage{Output: 8}})
	if row.LiveRate(base.Add(10*time.Second)) != "5.0" {
		t.Fatal("live rate did not sum prior measured step and current request")
	}
}
