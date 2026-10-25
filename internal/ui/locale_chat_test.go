package orbitui

import (
	"github.com/egoist/mygo/ui"
	"image"
	"slices"
	"testing"
)

func TestP20ChatChineseMenusPreservePublicIdentity(t *testing.T) {
	a := fixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	a.access = "supervised"
	a.workflow = "plan"
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if !tt.HasText("受监督") || !tt.HasText("计划") {
		t.Fatal("English composer display")
	}
	if err := tt.Click("Supervised"); err != nil {
		t.Fatal("public label changed", err)
	}
	if err := tt.ChooseMenuItem("自动接受编辑"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 1 || actions[0].Kind != "SetAccess" || actions[0].Value != "auto-accept-edits" {
		t.Fatal(actions)
	}
	if err := tt.Click("Plan"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("提问"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 2 || actions[1].Kind != "SetWorkflow" || actions[1].Value != "ask" {
		t.Fatal(actions)
	}
	a.Prefs.Language = "en"
	a.reloadLocale()
	tt.Frame()
	if !tt.HasText("Auto-accept edits") || !tt.HasText("Ask") {
		t.Fatal("English reload")
	}
}

func TestP20ChatChineseSearchPlaceholderKeepsFocusLabel(t *testing.T) {
	a := fixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	a.searchOpen = true
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	// Native input placeholders are paint, not Tester.Texts labels. Compare only
	// the empty input's text region across locale reloads, not the rest of the UI.
	r, ok := tt.Find("Find in transcript")
	if !ok {
		t.Fatal("missing input")
	}
	before := tt.Image()
	a.Prefs.Language = "en"
	a.reloadLocale()
	tt.Frame()
	after := tt.Image()
	crop := image.Rect(int(r.X)+12, int(r.Y)+3, int(r.X)+240, int(r.Y+r.H)-3).Intersect(before.Bounds()).Intersect(after.Bounds())
	changed := 0
	for y := crop.Min.Y; y < crop.Max.Y; y++ {
		for x := crop.Min.X; x < crop.Max.X; x++ {
			if before.RGBAAt(x, y) != after.RGBAAt(x, y) {
				changed++
			}
		}
	}
	if changed < 20 {
		t.Fatal("placeholder paint did not change across locales", changed, crop)
	}
	if _, ok := tt.Find("Find in transcript"); !ok {
		t.Fatal("public search label lost")
	}
}

func TestP20WorkspaceChineseMenuKeepsOwnerAndLocalizedReset(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(tt.Menu(), "从侧边栏移除") {
		t.Fatal(tt.Menu())
	}
	if err := tt.ChooseMenuItem("复制路径"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "/work/alpha" {
		t.Fatal(tt.Clipboard())
	}
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("图标和颜色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("图标和颜色") || !tt.HasText("重置") {
		t.Fatal("English appearance display")
	}
	if _, ok := tt.Find("Workspace appearance"); !ok {
		t.Fatal("public popup label")
	}
	if err := tt.Click(a.Tr("sidebar.reset")); err != nil {
		t.Fatal("localized Reset action", err)
	}
	awaitWorkspaceEdit(t, a, tt)
	if a.workspaceEntry("/work/alpha") == nil {
		t.Fatal("reset removed owner")
	}
}

func TestP20DisplayUnknownValuesPreserveFallback(t *testing.T) {
	a := fixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	if a.accessDisplay("future-mode", "Future mode") != "Future mode" || a.workflowDisplay("future-mode") != "Future-mode" {
		t.Fatal("unknown value turned into untranslated key")
	}
	if a.accessDisplay("full", "Full access") != a.accessDisplay("full-access", "Full access") {
		t.Fatal("legacy access alias display drift")
	}
	if a.workflowDisplay("") != a.Tr("workflow.build") {
		t.Fatal("empty workflow fallback")
	}
}
