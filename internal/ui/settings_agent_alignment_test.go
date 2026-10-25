package orbitui

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func TestAgentSettingsLargeFontRetryRowFitsNarrowViewports(t *testing.T) {
	const statusText = "Unknown: this pi version has no getter"
	for _, width := range []int{800, 960} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			a := agentAlignmentFixture(t)
			a.Prefs.UIFontSize = 21
			tt := ui.NewTester(a.View, width, 922)
			find := func(label string) ui.Rect {
				t.Helper()
				r, ok := tt.Find(label)
				if !ok {
					t.Fatal("missing mounted Agent element", label)
				}
				return r
			}
			row := find("Auto-retry setting row")
			control := find("Set retry policy")
			if control.W == 0 || control.Y+30 > 898 {
				// The complete Behavior card can exceed a short viewport at 21px.
				// Find reports clipped boxes, so scroll by the control's position,
				// not the truncated height of a row below the viewport.
				tt.Scroll(float32(width)-40, 500, 0, control.Y-500)
				row = find("Auto-retry setting row")
				control = find("Set retry policy")
			}
			title := find(a.Tr("settings.auto_retry"))
			status := find(statusText)
			board := find("Settings group " + a.Tr("settings.behavior"))
			t.Logf("width=%d board=%+v row=%+v title=%+v status=%+v control=%+v", width, board, row, title, status, control)
			if row.X < board.X || row.X+row.W > board.X+board.W || row.Y < 120 || row.Y+row.H > 922 || control.X+control.W > float32(width) || control.Y < row.Y || control.Y+control.H > row.Y+row.H {
				t.Fatal("mounted retry row/control not visible inside its card", board, row, control)
			}
			if status.X < title.X+title.W || status.X+status.W > control.X || status.Y < title.Y || status.Y+status.H > title.Y+title.H+.1 || status.W <= 0 {
				t.Fatal("retry status overlaps title/control or grows beyond the title line", title, status, control)
			}
			if control.W != 36 || control.H != 30 || math.Abs(float64(title.H-21*sourceDefaultLineHeight)) > .01 || math.Abs(float64(status.H-18*sourceDefaultLineHeight)) > .01 {
				t.Fatal("retry text/control remains vertically clipped after scrolling", title, status, control)
			}
			if out := os.Getenv("ORBIT_CAPTURE_DIR"); out != "" {
				if err := os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(filepath.Join(out, fmt.Sprintf("settings-agent-unknown-%d-font21-cpu.png", width)))
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				if err := png.Encode(file, tt.Image()); err != nil {
					t.Fatal(err)
				}
			}
			// The one-line status may elide at constrained widths. Its complete
			// explanation must still be available on the compact control tooltip,
			// which MyGo also exposes as the control's accessibility description.
			countStatus := func() int {
				count := 0
				for _, text := range tt.Texts() {
					if text == statusText {
						count++
					}
				}
				return count
			}
			before := countStatus()
			tt.Move(control.X+control.W/2, control.Y+control.H/2)
			tt.Advance(time.Second)
			if countStatus() != before+1 {
				t.Fatal("complete unknown status is unavailable from the mounted control tooltip", tt.Texts())
			}
			if _, ok := tt.Find("Automatic retry toggle"); ok || a.Runtime.Active.AutoRetry.Known {
				t.Fatal("narrow layout fabricated a known retry state")
			}
		})
	}
}

func agentAlignmentFixture(t *testing.T) *App {
	t.Helper()
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "agent"
	a.Model.Connected = true
	a.Model.FollowUpMode = "one-at-a-time"
	a.Model.AutoCompactionKnown = true
	a.Model.AutoCompactionEnabled = true
	a.Runtime.Active = &agent.RuntimeProcess{AutoRetry: agent.RetrySetting{Support: "unknown"}}
	return a
}

// These are CPU layout assertions against the fresh 40a8210/Rust native
// captures, not a new native screenshot acceptance or a pixel-parity claim.
func TestAgentSettingsSourceCardGeometryKeepsTruthfulRetryInline(t *testing.T) {
	for _, state := range []struct {
		name, status string
		retry        agent.RetrySetting
	}{
		{"unknown", "Unknown: this pi version has no getter", agent.RetrySetting{Support: "unknown"}},
		{"enabled", "Enabled", agent.RetrySetting{Support: "supported", Known: true, Enabled: true}},
		{"disabled", "Disabled", agent.RetrySetting{Support: "supported", Known: true}},
		{"pending unknown", "Updating…", agent.RetrySetting{Support: "unknown", Pending: true}},
		{"pending known", "Updating…", agent.RetrySetting{Support: "supported", Known: true, Enabled: true, Pending: true}},
		{"unsupported", "Unsupported by this pi runtime", agent.RetrySetting{Support: "unsupported"}},
	} {
		t.Run(state.name, func(t *testing.T) {
			a := agentAlignmentFixture(t)
			a.Runtime.Active.AutoRetry = state.retry
			tt := ui.NewTester(a.View, 1470, 922)
			for _, group := range []struct {
				key       string
				y, bottom float32
			}{
				{"settings.behavior", 121, 403},
				{"settings.session_titles", 447, 587},
				{"settings.session_defaults", 632, 773},
			} {
				r, ok := tt.Find("Settings group " + a.Tr(group.key))
				if !ok || math.Abs(float64(r.Y-group.y)) > 2 || math.Abs(float64(r.Y+r.H-group.bottom)) > 2 {
					t.Errorf("source card %s expected y %.0f bottom %.0f, got %+v", group.key, group.y, group.bottom, r)
				}
			}
			title, _ := tt.Find(a.Tr("settings.auto_retry"))
			status, ok := tt.Find(state.status)
			if !ok || !tt.HasText(state.status) || status.Y < title.Y || status.Y+status.H > title.Y+title.H+.1 {
				t.Fatal("retry knowledge must remain visible on the title line", state.status, title, status)
			}
			control := "Set retry policy"
			if state.retry.Known {
				control = "Automatic retry toggle"
			}
			r, ok := tt.Find(control)
			if !ok || r.W != 36 || r.H != 20 || r.X+r.W != 1294 {
				t.Fatal("retry control left the source switch axis", control, r)
			}
			if out := os.Getenv("ORBIT_CAPTURE_DIR"); out != "" && (state.name == "unknown" || state.name == "enabled") {
				if err := os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(filepath.Join(out, "settings-agent-"+state.name+"-cpu.png"))
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				if err := png.Encode(file, tt.Image()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestAgentSettingsRetryAcknowledgementPendingFailureAndMenuRecovery(t *testing.T) {
	a := agentAlignmentFixture(t)
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	tt := ui.NewTester(a.View, 1470, 922)
	if _, ok := tt.Find("Automatic retry toggle"); ok {
		t.Fatal("unobserved retry fabricated a Boolean switch")
	}
	if err := tt.Click("Set retry policy"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	if len(actions) != 0 {
		t.Fatal("dismissed retry menu changed policy", actions)
	}
	if err := tt.Click("Set retry policy"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Disable"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "SetAutoRetry" || actions[0].Data["enabled"] != false || a.Runtime.Active.AutoRetry.Known {
		t.Fatal("unknown retry must dispatch without inventing acknowledgement", actions, a.Runtime.Active.AutoRetry)
	}
	a.Runtime.Active.AutoRetry = agent.RetrySetting{Support: "supported", Known: true, Enabled: true}
	tt.Frame()
	if err := tt.Click("Automatic retry toggle"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 2 || actions[1].Data["enabled"] != false || !a.Runtime.Active.AutoRetry.Enabled {
		t.Fatal("known retry switch changed the observed value optimistically", actions, a.Runtime.Active.AutoRetry)
	}
	a.Runtime.Active.AutoRetry.Pending = true
	tt.Frame()
	if err := tt.Click("Automatic retry toggle"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeySpace)
	if len(actions) != 2 || !tt.HasText("Updating…") {
		t.Fatal("pending retry accepted a repeated change", actions)
	}
	a.Runtime.Active.AutoRetry.Pending = false
	a.Runtime.Active.AutoRetry.Error = "Explicit fixture rejection"
	tt.Frame()
	if !tt.HasText("Enabled") || !tt.HasText("Explicit fixture rejection") {
		t.Fatal("known failed setter lost its last acknowledged value or error")
	}
	// A timeout or restart may clear knowledge. Returning to the compact menu
	// must not retain a stale switch value from the previous acknowledged state.
	a.Runtime.Active.AutoRetry = agent.RetrySetting{Support: "unknown", Error: "Explicit fixture timeout"}
	tt.Frame()
	if _, ok := tt.Find("Automatic retry toggle"); ok || !tt.HasText("Unknown: this pi version has no getter") || !tt.HasText("Explicit fixture timeout") {
		t.Fatal("retry knowledge loss left a stale Boolean control")
	}
	if err := tt.Click("Set retry policy"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Enable"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 3 || actions[2].Data["enabled"] != true {
		t.Fatal("retry menu did not recover after knowledge loss", actions)
	}
}

func TestAgentSettingsRetryBlockedStatesAndStaleMenu(t *testing.T) {
	for _, state := range []string{"disconnected", "pending", "unsupported"} {
		t.Run(state, func(t *testing.T) {
			a := agentAlignmentFixture(t)
			var actions []Action
			a.Options.OnAction = func(action Action) { actions = append(actions, action) }
			tt := ui.NewTester(a.View, 1470, 922)
			if err := tt.Click("Set retry policy"); err != nil {
				t.Fatal(err)
			}
			switch state {
			case "disconnected":
				a.Model.Connected = false
			case "pending":
				a.Runtime.Active.AutoRetry.Pending = true
			case "unsupported":
				a.Runtime.Active.AutoRetry.Support = "unsupported"
			}
			// Simulate state changing after the OS menu opened but before its
			// choice is delivered; the callback must recheck the current gate.
			_ = tt.ChooseMenuItem("Enable")
			if len(actions) != 0 {
				t.Fatal("stale menu dispatched into a blocked runtime", actions)
			}
			tt.CloseMenu()
			tt.Frame()
			if err := tt.Click("Set retry policy"); err != nil {
				t.Fatal(err)
			}
			if len(tt.Menu()) != 0 || len(actions) != 0 {
				t.Fatal("disabled retry control reopened its menu", tt.Menu(), actions)
			}
		})
	}
}

func TestAgentSettingsSegmentSourceColorsFontsAndInteraction(t *testing.T) {
	a := agentAlignmentFixture(t)
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	tt := ui.NewTester(a.View, 1470, 922)
	composer, _ := tt.Find("Enter while running")
	followup, _ := tt.Find("Follow-up delivery")
	for _, control := range []ui.Rect{composer, followup} {
		if control.H != 28 {
			t.Fatal("source medium segment height changed", control)
		}
	}
	pixel := func(r ui.Rect, y float32) ui.Color {
		p := tt.Image().RGBAAt(int(r.X+5), int(r.Y+y))
		return ui.RGB(p.R, p.G, p.B)
	}
	neutral, blue := pixel(composer, 6), pixel(followup, 6)
	if neutral.R != neutral.G || neutral.R != neutral.B || blue.B <= blue.G || blue.G <= blue.R {
		t.Fatal("composer must be neutral and follow-up blue", neutral, blue)
	}
	neutralTop, neutralBottom := usageRaisedColors(a.Theme.Active)
	blueTop, blueBottom := usageRaisedColors(a.Theme.Selection)
	if neutral.R > neutralTop.R || neutral.R < neutralBottom.R || blue.B > blueTop.B || blue.B < blueBottom.B {
		t.Fatal("selected segments used the wrong source palette role", neutral, blue)
	}
	if neutral.R <= pixel(composer, 22).R || blue.B <= pixel(followup, 22).B {
		t.Fatal("source raised segment gradient missing")
	}
	if err := tt.Click(a.Tr("composer.steer_task")); err != nil {
		t.Fatal(err)
	}
	if a.Prefs.ComposerSendMode != "steer" || a.Options.Store.UI().ComposerSendMode != "steer" || len(actions) != 0 {
		t.Fatal("composer choice must persist locally without changing runtime delivery", a.Prefs.ComposerSendMode, actions)
	}
	tt.Key(0, ui.KeyLeft)
	if a.Prefs.ComposerSendMode != "follow_up" {
		t.Fatal("composer keyboard navigation regressed")
	}
	if err := tt.Click(a.Tr("settings.all")); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "SetFollowUpMode" || actions[0].Value != "all" || a.Model.FollowUpMode != "one-at-a-time" {
		t.Fatal("follow-up choice must dispatch without changing observed runtime state", actions)
	}
	a.Model.Connected = false
	tt.Frame()
	tt.Key(0, ui.KeyRight)
	if err := tt.Click(a.Tr("settings.all")); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatal("disabled follow-up selector dispatched", actions)
	}
	// Compare the mounted text widths to explicit source 14px/500-weight text,
	// rather than relying on a token assertion that does not exercise the UI.
	for _, size := range []float32{14, 21} {
		a.Theme.FontSize = size
		label := a.Tr("composer.queue_follow_up")
		fonts := ui.NewTester(func(c *ui.Context) {
			ui.Column(c).Fill().Font(a.Theme.Font).Children(func() {
				a.agentSettingsSegments(c, "Font fixture", "follow_up", []settingsChoice{{"follow_up", label}}, a.Theme.Active, false, func(string) {})
				ui.Text(c, label).Label("Agent segment font reference").FontSize(a.Theme.Px(14)).FontWeight(500).LineHeight(1).AlignSelf(ui.Start)
			})
		}, 1000, 400)
		button, _ := fonts.Find(label)
		reference, _ := fonts.Find("Agent segment font reference")
		group, _ := fonts.Find("Font fixture")
		if math.Abs(float64(button.W-2*a.Theme.Space(8)-reference.W)) > .01 || group.H != a.Theme.Px(28) {
			t.Fatal("Agent segment source font mismatch", size, button, reference)
		}
	}
}
