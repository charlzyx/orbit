package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
	"time"
)

func TestAuditSidebarFolderClick(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Workspaces.GroupBy = "workspace"
	a.sidebarState.extraSessions = map[string]int{"alpha": 20}
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	for i := 0; i < 4; i++ {
		if err := tt.Click("alpha"); err != nil {
			t.Fatal(err)
		}
		want := i%2 == 0
		if a.sidebarGroupCollapsed("alpha") != want {
			t.Fatalf("click %d collapsed=%v want=%v", i, a.sidebarGroupCollapsed("alpha"), want)
		}
		_, shown := tt.Find("new task")
		if shown == want {
			t.Fatalf("click %d ordinary row visible=%v collapsed=%v", i, shown, want)
		}
	}
}

func TestAuditExplorerRepeatedFolderClick(t *testing.T) {
	a, tt, _ := mountedExplorer(t)
	for i := 0; i < 4; i++ {
		if err := tt.Click("Explorer row docs"); err != nil {
			t.Fatal(err)
		}
		want := i%2 == 0
		if a.files.expanded["docs"] != want {
			t.Fatalf("toggle%d state", i)
		}
		_, shown := tt.Find("Explorer row docs/readme.md")
		if shown != want {
			t.Fatalf("toggle%d child visible%v want%v", i, shown, want)
		}
	}
}

func TestSidebarCollapseKeepsCurrentAndExplainsBothLanguages(t *testing.T) {
	for _, language := range []string{"en", "zh-CN"} {
		t.Run(language, func(t *testing.T) {
			a := sourceSidebarFixture(t)
			a.Prefs.Language = language
			a.reloadLocale()
			a.Workspaces.GroupBy = "workspace"
			a.sidebarState.extraSessions = map[string]int{"alpha": 20}
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("alpha"); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"active task", "pin task", "running task"} {
				if _, ok := tt.Find(name); !ok {
					t.Fatal("preserved session missing", name)
				}
			}
			if _, ok := tt.Find("new task"); ok {
				t.Fatal("ordinary session not collapsed")
			}
			r, ok := tt.Find("Sidebar workspace header /work/alpha")
			if !ok {
				t.Fatal("header missing")
			}
			tt.Move(900, 100)
			tt.Frame()
			tt.Move(r.X+35, r.Y+r.H/2)
			tt.Advance(time.Second)
			if !tt.HasText(a.Tr("go.sidebar.expand_hint")) {
				t.Fatal("collapse explanation missing", language, tt.Texts())
			}
		})
	}
}
