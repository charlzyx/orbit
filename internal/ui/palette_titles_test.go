package orbitui

import (
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestPaletteTitlesMountedToggleAndLocale(t *testing.T) {
	for _, language := range []string{"en", "zh-CN"} {
		for _, tc := range []struct{ id, query, show, hide string }{
			{"ToggleSidebar", "toggle show hide left sidebar sessions history", "command_palette.show_sidebar", "command_palette.hide_sidebar"},
			{"ToggleSidePanel", "toggle show hide right panel review git diff", "command_palette.show_side_panel", "command_palette.hide_side_panel"},
			{"ToggleProjectPanel", "explorer files project panel tree folders workspace toggle show hide", "explorer.show", "explorer.hide"},
		} {
			t.Run(language+"/"+tc.id, func(t *testing.T) {
				a := sourceSidebarFixture(t)
				a.Prefs.Language = language
				a.SidebarVisible = false
				a.ExplorerVisible = false
				a.ReviewVisible = false
				a.Options.OnAction = func(Action) {}
				a.paletteOpen = true
				a.paletteQuery = tc.query
				tt := ui.NewTester(a.View, 1222, 878)
				tt.Frame()
				if err := tt.Click(a.Tr(tc.show)); err != nil {
					t.Fatal("missing source show title", err)
				}
				tt.Frame()
				a.paletteOpen = true
				tt.Frame()
				if !tt.HasText(a.Tr(tc.hide)) {
					t.Fatal("enabled toggle lacks source alternate title", tc.id, a.Tr(tc.hide))
				}
				if err := tt.Click(a.Tr(tc.hide)); err != nil {
					t.Fatal(err)
				}
				tt.Frame()
				a.paletteOpen = true
				tt.Frame()
				if !tt.HasText(a.Tr(tc.show)) {
					t.Fatal("disabled toggle retains alternate title")
				}
			})
		}
	}
}
func TestPaletteTitlesMountedPinUsesCurrentCachedPathNotSessionList(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	a.Model.SessionPath = "/synthetic/displayed-task.jsonl"
	a.Model.SessionID = "displayed-task"
	a.Model.Sessions = nil
	a.Prefs.Language = "zh-CN"
	a.paletteOpen = true
	a.paletteQuery = "pin unpin shortlist favorite session"
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.Click(a.Tr("sidebar.pin_session")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	a.paletteOpen = true
	tt.Frame()
	if !x.pinned[a.Model.SessionPath] || !tt.HasText(a.Tr("sidebar.unpin_session")) {
		t.Fatal("pin state/cache not reflected without session index row")
	}
	if err := tt.Click(a.Tr("sidebar.unpin_session")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	a.paletteOpen = true
	tt.Frame()
	if x.pinned[a.Model.SessionPath] || !tt.HasText(a.Tr("sidebar.pin_session")) {
		t.Fatal("unpin title not refreshed")
	}
}

func titleCommand(t *testing.T, id string) CommandSpec {
	t.Helper()
	for _, cmd := range Commands {
		if cmd.ID == id {
			return cmd
		}
	}
	t.Fatal("missing source command", id)
	return CommandSpec{}
}
func TestPaletteTitlesExactOriginalKeysAndLocaleFallback(t *testing.T) {
	for _, tc := range []struct{ id, show, hide string }{
		{"ToggleSidebar", "command_palette.show_sidebar", "command_palette.hide_sidebar"},
		{"ToggleSidePanel", "command_palette.show_side_panel", "command_palette.hide_side_panel"},
		{"ToggleTerminal", "command_palette.show_terminal", "command_palette.hide_terminal"},
		{"ToggleProjectPanel", "explorer.show", "explorer.hide"},
		{"PinSession", "sidebar.pin_session", "sidebar.unpin_session"},
	} {
		cmd := titleCommand(t, tc.id)
		if cmd.TitleKey != tc.show || paletteCommandTitleKey(cmd, false) != tc.show || paletteCommandTitleKey(cmd, true) != tc.hide {
			t.Fatal("not original source keys", tc)
		}
		a := fixture(t)
		for _, language := range []string{"en", "zh-CN", "unknown-fixture-locale"} {
			a.Prefs.Language = language
			if a.Tr(tc.show) == tc.show || a.Tr(tc.hide) == tc.hide {
				t.Fatal("unresolved original catalog key", language, tc)
			}
		}
	}
	cmd := titleCommand(t, "RefreshSessions")
	if paletteCommandTitleKey(cmd, true) != cmd.TitleKey {
		t.Fatal("unrelated command title changed")
	}
}
func TestPaletteTitlesTerminalPublicClosedAndNilMounted(t *testing.T) {
	for _, withPanel := range []bool{false, true} {
		a := sourceSidebarFixture(t)
		if withPanel {
			p := orbiterm.NewPanel(orbiterm.Options{Workspace: t.TempDir(), Command: []string{"/synthetic/never-start-this-command"}})
			t.Cleanup(func() { _ = p.Close() })
			a.Options.Terminal = p
			if p.IsOpen() {
				t.Fatal("fresh panel unexpectedly open")
			}
		}
		a.paletteOpen = true
		a.paletteQuery = "toggle show hide terminal shell console command line pty"
		tt := ui.NewTester(a.View, 1222, 878)
		tt.Frame()
		if !tt.HasText(a.Tr("command_palette.show_terminal")) {
			t.Fatal("closed/nil terminal lacks show title")
		}
	}
	// Opened Terminal→hide is covered by the exact key projection table, not
	// by starting a shell/PTY or downloading native terminal libraries here.
}
func TestPaletteTitlesLiveProjectionAndEmptyPinGuard(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Prefs.Language = "en"
	a.paletteOpen = true
	a.paletteQuery = "toggle show hide right panel review git diff"
	a.page = "settings"
	a.ReviewVisible = false
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if !tt.HasText(a.Tr("command_palette.show_side_panel")) {
		t.Fatal("missing initial show title")
	}
	a.ReviewVisible = true
	tt.Frame()
	if !tt.HasText(a.Tr("command_palette.hide_side_panel")) {
		t.Fatal("retained logical-open state lost behind settings")
	}
	lookups := 0
	a.Options.SessionPinned = func(string) bool { lookups++; return true }
	a.Model.SessionPath = ""
	if a.paletteCommandTitle(titleCommand(t, "PinSession")) != a.Tr("sidebar.pin_session") || lookups != 0 {
		t.Fatal("empty path consulted pins")
	}
	a.Model.SessionPath = "/synthetic/first.jsonl"
	if a.paletteCommandTitle(titleCommand(t, "PinSession")) != a.Tr("sidebar.unpin_session") || lookups != 1 {
		t.Fatal("owner cache not consulted")
	}
	a.Options.SessionPinned = nil
	a.Model.Sessions = []SessionRow{{ID: "shared-id", Path: "/synthetic/other.jsonl", Pinned: true}, {ID: "shared-id", Path: a.Model.SessionPath, Pinned: false}}
	if a.paletteCommandTitle(titleCommand(t, "PinSession")) != a.Tr("sidebar.pin_session") {
		t.Fatal("different path pin contaminated active row")
	}
}
