package orbitui

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func queueTester(a *App, width int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Gap(0).LineHeight(sourceDefaultLineHeight).Children(func() {
			a.queueBar(c)
			ui.Box(c).Height(10).Label("After queue")
		})
	}, width, 600)
}

func TestQueueProjectionRetainsKindsOrderAndOwnership(t *testing.T) {
	s := agent.Snapshot{Queue: agent.Queue{Steering: []string{"s1", "s2"}, FollowUp: []string{"f1", "f2"}}}
	v := adaptSnapshot(s, "", nil, nil)
	if !reflect.DeepEqual(v.Queue, s.Queue) || !reflect.DeepEqual(v.Queued, []string{"s1", "s2", "f1", "f2"}) {
		t.Fatal("queue kinds or order lost", v.Queue, v.Queued)
	}
	s.Queue.Steering[0], s.Queue.FollowUp[0] = "changed steering", "changed follow-up"
	if v.Queue.Steering[0] != "s1" || v.Queue.FollowUp[0] != "f1" {
		t.Fatal("projection aliases mutable source", v.Queue)
	}
	a := fixture(t)
	a.Model = Snapshot{Queued: []string{"legacy fixture"}}
	if got := a.queueContents(); !reflect.DeepEqual(got.FollowUp, a.Model.Queued) || len(got.Steering) != 0 {
		t.Fatal("legacy fixture fallback changed", got)
	}
	a.Model = adaptSnapshot(agent.Snapshot{}, "", nil, nil)
	a.Model.Queued = []string{"stale flattened fixture"}
	if got := a.queueContents(); len(got.Steering)+len(got.FollowUp) != 0 {
		t.Fatal("explicit empty queue must take precedence over fixture fallback", got)
	}
}

func TestQueueSourceHeaderTooltipAndSingleClearAction(t *testing.T) {
	cases := []struct {
		name            string
		queue           agent.Queue
		heading, detail string
	}{
		{"steering", agent.Queue{Steering: []string{"Steer one", "Steer two"}}, "STEERING · 2", "Steering the running turn (2)"},
		{"follow-up", agent.Queue{FollowUp: []string{"Follow one", "Follow two"}}, "QUEUED · 2", "Queued — sends after the current task finishes (2)"},
		{"mixed", agent.Queue{Steering: []string{"Steer one", "Steer two"}, FollowUp: []string{"Follow one", "Follow two"}}, "QUEUED · 4", "Steering now; follow-ups send after the task finishes (4)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := fixture(t)
			a.Model.Queue = tc.queue
			var actions []Action
			a.Options.OnAction = func(v Action) { actions = append(actions, v) }
			tt := queueTester(a, 600)
			tt.Frame()
			if !tt.HasText(tc.heading) {
				t.Fatal("wrong dominant-state label/count", tt.Texts())
			}
			clearCount := 0
			for _, text := range tt.Texts() {
				if text == "Clear queue" {
					clearCount++
				}
			}
			if clearCount != 1 {
				t.Fatal("source queue must have one header clear action", clearCount, tt.Texts())
			}
			lastY := float32(-1)
			for _, message := range append(append([]string{}, tc.queue.Steering...), tc.queue.FollowUp...) {
				r, ok := tt.Find(message)
				if !ok || r.Y <= lastY {
					t.Fatal("rows not in source steering-then-follow-up order", message, r)
				}
				lastY = r.Y
				if err := tt.Click(message); err != nil {
					t.Fatal(err)
				}
			}
			if len(actions) != 0 {
				t.Fatal("read-only queue row triggered action", actions)
			}
			label, _ := tt.Find(tc.heading)
			tt.Move(label.X+label.W/2, label.Y+label.H/2)
			tt.Advance(time.Second)
			if !tt.HasText(tc.detail) {
				t.Fatal("source delivery detail tooltip absent", tc.detail, tt.Texts())
			}
			if err := tt.Click("Clear queue"); err != nil {
				t.Fatal(err)
			}
			if len(actions) != 1 || actions[0].Kind != "ClearQueue" {
				t.Fatal("header did not dispatch exactly one clear", actions)
			}
			a.Model.Queue = agent.Queue{}
			tt.Frame()
			if _, exists := tt.Find("Queue bar"); exists {
				t.Fatal("empty source queue still rendered")
			}
		})
	}
}

func TestQueueSourceLocalizedLabelsAndTooltip(t *testing.T) {
	a := fixture(t)
	var err error
	a.catalog, err = orbitassets.Locale("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	a.Model.Queue = agent.Queue{Steering: []string{"First"}, FollowUp: []string{"Second"}}
	tt := queueTester(a, 600)
	tt.Frame()
	for _, key := range []string{"view.queue_steer", "view.queue_follow_up", "view.clear_queue"} {
		if !tt.HasText(strings.ToUpper(a.Tr(key))) && !tt.HasText(a.Tr(key)) {
			t.Fatal("queue label bypassed selected catalog", key, tt.Texts())
		}
	}
	heading, detail := a.queueHeading(a.Model.Queue)
	if !tt.HasText(heading) || strings.HasPrefix(heading, "QUEUED") || strings.Contains(detail, "%{count}") {
		t.Fatal("localized heading/count failed", heading, detail)
	}
	label, _ := tt.Find(heading)
	tt.Move(label.X+label.W/2, label.Y+label.H/2)
	tt.Advance(time.Second)
	if !tt.HasText(detail) {
		t.Fatal("localized queue tooltip absent", detail, tt.Texts())
	}
}

// These are source-derived CPU geometry assertions, not native OS or pixel parity.
func TestQueueSourceGeometryUsesLiteralAndScaledMetrics(t *testing.T) {
	for _, font := range []float32{14, 20} {
		for _, density := range []int{80, 100, 120} {
			t.Run(fmt.Sprintf("font=%g/density=%d", font, density), func(t *testing.T) {
				a := fixture(t)
				a.Theme.FontSize, a.Theme.Density = font, density
				long := strings.Repeat("A bounded two-line queued message. ", 15)
				a.Model.Queue = agent.Queue{Steering: []string{long}, FollowUp: []string{"Follow-up body"}}
				tt := queueTester(a, 420)
				tt.Frame()
				find := func(label string) ui.Rect {
					t.Helper()
					r, ok := tt.Find(label)
					if !ok {
						t.Fatal("missing", label, tt.Texts())
					}
					return r
				}
				close := func(name string, got, want float32) {
					t.Helper()
					if math.Abs(float64(got-want)) > .1 {
						t.Fatalf("%s = %g, want %g", name, got, want)
					}
				}
				bar, header, clear := find("Queue bar"), find("Queue header"), find("Clear queue")
				steer, body, follow, after := find("STEER"), find(long), find("Follow-up body"), find("After queue")
				followKind := find("FOLLOW-UP")
				close("literal header height", header.H, 30)
				close("compact clear width", clear.W, a.Theme.Px(18))
				close("compact clear height", clear.H, a.Theme.Px(18))
				close("horizontal row padding", steer.X-bar.X, 1+a.Theme.Space(12))
				close("literal type/body gap", body.X-steer.X-steer.W, 10)
				close("row top separator and padding", body.Y-header.Y-header.H, 8)
				// MyGo currently floors line boxes at the native font metrics:
				// this is 17 DIP at source font 12 despite FixedLineHeight(16).
				// Preserve the source request, prove the two-line clamp relative
				// to an actual one-line row, and report rather than hide that gap.
				close("two-line clamp", body.H, 2*follow.H)
				if math.Abs(float64(follow.H-a.Theme.Px(16))) > .1 {
					t.Logf("native text metric limitation: requested source line height %g, measured %g", a.Theme.Px(16), follow.H)
				}
				close("next row separator and padding", follow.Y-body.Y-body.H, 15)
				close("queue bottom padding and border", bar.Y+bar.H-max(follow.Y+follow.H, followKind.Y+followKind.H), 8)
				close("source card bottom margin", after.Y-bar.Y-bar.H, a.Theme.Space(8))
			})
		}
	}
}
