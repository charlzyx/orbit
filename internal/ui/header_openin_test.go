package orbitui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

func TestHeaderOpenInCachedIconAndOriginalActions(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 18, 18))
	im.Set(2, 2, color.RGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, im); err != nil {
		t.Fatal(err)
	}
	apps := []native.ExternalApp{{ID: "finder", Label: "Finder", Icon: encoded.Bytes()}, {ID: "editor", Label: "Editor", Icon: []byte("invalid")}, {ID: "empty", Label: "No icon"}}
	icons := decodeOpenAppIcons(apps)
	if len(icons) != 1 || icons["finder"] == nil {
		t.Fatal("decode must exclude unavailable icons")
	}
	// The source bytes can disappear after discovery; frames only use the cache.
	apps[0].Icon = nil
	a := fixture(t)
	a.OpenApps = apps
	a.openAppIcons = icons
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if err := tt.Click("Finder"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "OpenInApp" || actions[0].Value != "finder" {
		t.Fatalf("preferred action: %+v", actions)
	}
	actions = nil
	if err := tt.Click("Open in…"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Editor"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "OpenInApp" || actions[0].Value != "editor" {
		t.Fatalf("menu action: %+v", actions)
	}
	actions = nil
	if err := tt.Click("Open in…"); err != nil {
		t.Fatal(err)
	}
	// Headless native menus use their platform dismissal seam; raw Escape
	// is an app key here, not Cocoa menu tracking.
	if len(tt.Menu()) == 0 {
		t.Fatal("menu did not open")
	}
	tt.CloseMenu()
	tt.Frame()
	if len(actions) != 0 {
		t.Fatal("cancel dispatched", actions)
	}
	a.OpenApps = nil
	tt.Frame()
	if _, ok := tt.Find("Open in…"); ok {
		t.Fatal("empty discovery must hide controls")
	}
}

func TestChromeSidebarBoundsStableAcrossHover(t *testing.T) {
	for _, width := range []int{960, 1200, 1470} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			a := sourceSidebarFixture(t)
			a.Workspaces.GroupBy = "workspace"
			tt := ui.NewTester(a.View, width, 800)
			tt.Frame()
			header, ok := tt.Find("Sidebar workspace header /work/alpha")
			if !ok {
				t.Fatal("header")
			}
			row, ok := tt.Find("active task")
			if !ok {
				t.Fatal("session")
			}
			if header.X != row.X+8 || header.H != a.Theme.Px(26)+14 {
				t.Fatalf("shared inset/header: %+v %+v", header, row)
			}
			tt.Move(header.X+header.W/2, header.Y+header.H/2)
			tt.Frame()
			after, _ := tt.Find("Sidebar workspace header /work/alpha")
			if after != header {
				t.Fatal("hover moved header", header, after)
			}
			if err := tt.Click("alpha"); err != nil {
				t.Fatal(err)
			}
			if _, ok := tt.Find("active task"); !ok {
				t.Fatal("collapse lost current session")
			}
		})
	}
}

func TestChromeSidebarLongLabelsNarrowAndScaled(t *testing.T) {
	for _, font := range []float64{14, 22} {
		t.Run(fmt.Sprint(font), func(t *testing.T) {
			a := sourceSidebarFixture(t)
			a.Workspaces.GroupBy = "workspace"
			a.SidebarWidth = 200
			a.Prefs.UIFontSize = font
			a.Model.Sessions[8].Name = "当前任务 Very long active session title with mixed language"
			tt := ui.NewTester(a.View, 960, 800)
			tt.Frame()
			h, ok := tt.Find("Sidebar workspace header /work/alpha")
			if !ok {
				t.Fatal("header")
			}
			r, ok := tt.Find(a.Model.Sessions[8].Name)
			if !ok {
				t.Fatal("long session")
			}
			if h.X < 0 || h.X+h.W > 200 || r.X+r.W > 200 {
				t.Fatal("sidebar control escaped narrow viewport", h, r)
			}
			if err := tt.Click("alpha"); err != nil {
				t.Fatal(err)
			}
			if _, ok := tt.Find(a.Model.Sessions[8].Name); !ok {
				t.Fatal("current session lost on collapse")
			}
		})
	}
}

func TestHeaderOpenInFallbackKeepsNameTooltipAndBoundedWidth(t *testing.T) {
	for _, language := range []string{"en", "zh-CN"} {
		t.Run(language, func(t *testing.T) {
			a := fixture(t)
			a.Prefs.Language = language
			a.reloadLocale()
			name := "An unusually long external application name 外部应用程序"
			a.OpenApps = []native.ExternalApp{{ID: "external", Label: name}}
			tt := ui.NewTester(a.View, 960, 800)
			tt.Frame()
			r, ok := tt.Find(name)
			if !ok {
				t.Fatal("accessible app name lost")
			}
			if r.W > 50 || r.X+r.W > 960 {
				t.Fatal("fallback must not expand toolbar", r)
			}
			tt.Move(r.X+r.W/2, r.Y+r.H/2)
			tt.Advance(time.Second)
			want := strings.ReplaceAll(a.text("open_in.open_in_app", "Open in %{app}"), "%{app}", name)
			if !tt.HasText(want) {
				t.Fatal("fallback tooltip lost app name", tt.Texts())
			}
		})
	}
}
