package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestWorkspaceResetChineseNameAndExactPersistence(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	for i := range a.Workspaces.Workspaces {
		a.Workspaces.Workspaces[i].Icon = "rocket-01"
		a.Workspaces.Workspaces[i].Tint = "blue"
	}
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("图标和颜色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("Reset"); ok {
		t.Fatal("English accessible/public name overrides Chinese text")
	}
	if err := tt.Click("重置"); err != nil {
		t.Fatal(err)
	}
	awaitWorkspaceEdit(t, a, tt)
	for _, w := range a.Options.Store.Workspaces().Workspaces {
		if w.Path == "/work/alpha" {
			if w.Icon != "" || w.Tint != "" {
				t.Fatal("target not reset", w)
			}
		} else if w.Icon != "rocket-01" || w.Tint != "blue" {
			t.Fatal("wrong workspace reset", w)
		}
	}
}

func TestWorkspaceResetStableIdentityAcrossLanguageAndSiblings(t *testing.T) {
	a := fixture(t)
	a.Prefs.Language = "en"
	a.reloadLocale()
	var id uint64
	var clicks int
	extra := false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c.Key("stable-parent")).Children(func() {
			if extra {
				ui.Text(c, "extra noninteractive sibling")
			}
			b := a.workspaceResetButton(c, false)
			id = b.ID()
			if b.Clicked() {
				clicks++
			}
		})
	}, 500, 300)
	tt.Frame()
	before := id
	if before == 0 {
		t.Fatal("missing public element ID")
	}
	if err := tt.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 || !tt.Focused("Reset") {
		t.Fatal("initial action/focus", clicks)
	}
	extra = true
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	tt.Frame()
	if id != before || !tt.Focused("重置") {
		t.Fatal("locale/sibling insertion replaced identity or focus", before, id)
	}
	if _, ok := tt.Find("Reset"); ok {
		t.Fatal("English accessible name retained")
	}
	tt.Key(0, ui.KeyEnter)
	if clicks != 2 {
		t.Fatal("localized focused action lost", clicks)
	}
	extra = false
	a.Prefs.Language = "en"
	a.reloadLocale()
	tt.Frame()
	if id != before || !tt.Focused("Reset") {
		t.Fatal("return locale lost stable identity")
	}
}

func TestWorkspaceResetHeldPressCannotOutliveOwner(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	for i := range a.Workspaces.Workspaces {
		a.Workspaces.Workspaces[i].Icon = "rocket-01"
		a.Workspaces.Workspaces[i].Tint = "blue"
	}
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("图标和颜色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	r, ok := tt.Find("重置")
	if !ok {
		t.Fatal("missing reset")
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	s := a.Model
	s.Workspace = "/work/beta"
	a.SetSnapshot(s)
	s.Workspace = "/work/alpha"
	a.SetSnapshot(s)
	tt.Frame()
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	tt.Frame()
	if a.workspaceActions.busy || a.workspaceActions.open {
		t.Fatal("stale popup action survived owner change")
	}
	for _, w := range a.Options.Store.Workspaces().Workspaces {
		if w.Icon != "rocket-01" || w.Tint != "blue" {
			t.Fatal("stale press changed stored owner", w)
		}
	}
}
