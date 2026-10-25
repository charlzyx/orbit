package orbitui

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFilteredPickersResetAndClampKeyboardSelection(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	for i := 0; i < 8; i++ {
		a.Model.Models = append(a.Model.Models, ModelChoice{ID: fmt.Sprint(i), Provider: "provider", Name: fmt.Sprintf("Model %d", i)})
	}
	tt := ui.NewTester(a.View, 1200, 800)
	a.modelOpen = true
	tt.Frame()
	for i := 0; i < 6; i++ {
		tt.Key(0, ui.KeyDown)
	}
	tt.Type("Model 2")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Value != "2" {
		t.Fatalf("filtered model not selected: %+v", actions)
	}
	a.modelOpen = true
	a.modelQuery = ""
	tt.Frame()
	tt.Type("no-such-model")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || !a.modelOpen {
		t.Fatal("empty model filter submitted")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	tt.Type("Model 7")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 2 || actions[1].Value != "7" {
		t.Fatalf("filter recovery failed: %+v", actions)
	}
	a.paletteOpen = true
	a.paletteQuery = ""
	tt.Frame()
	for i := 0; i < 6; i++ {
		tt.Key(0, ui.KeyDown)
	}
	if a.commandCursor == 0 {
		t.Fatal("palette keyboard cursor reset on every frame")
	}
	tt.Type("new session")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 3 || actions[2].Kind != "NewSession" {
		t.Fatalf("filtered command not selected: %+v", actions)
	}
}

func TestNativeMenuCannotMutateBehindModal(t *testing.T) {
	for _, mode := range []string{"prompt", "confirm", "custom", "secret", "json", "image", "palette", "model"} {
		t.Run(mode, func(t *testing.T) {
			a := fixture(t)
			var actions []Action
			a.Options.OnAction = func(action Action) { actions = append(actions, action) }
			a.Draft = "keep"
			switch mode {
			case "prompt":
				a.promptOpen = true
			case "confirm":
				a.confirmOpen = true
			case "custom":
				a.Model.Dialog = &DialogRequest{Method: "custom"}
			case "secret":
				a.secretOpen = true
			case "json":
				a.jsonOpen = true
			case "image":
				a.imageOpen = true
			case "palette":
				a.paletteOpen = true
			case "model":
				a.modelOpen = true
			}
			for _, command := range []string{"NewSession", "OpenSettings", "ToggleTerminal"} {
				a.nativeCommand(command)
			}
			if len(actions) > 0 || a.page != "chat" || a.Draft != "keep" {
				t.Fatal("native accelerator bypassed modal", mode, actions, a.page)
			}
		})
	}
}

func TestMinimumWindowModalsKeepActionsReachable(t *testing.T) {
	for _, mode := range []string{"custom", "json", "model", "prompt", "confirm"} {
		t.Run(mode, func(t *testing.T) {
			a := fixture(t)
			switch mode {
			case "custom":
				a.Model.Dialog = &DialogRequest{ID: "component", Method: "custom", Title: "Extension"}
			case "json":
				a.jsonOpen = true
			case "model":
				a.modelOpen = true
			case "prompt":
				a.Prompt("Prompt", "", nil)
			case "confirm":
				a.Confirm("Confirmation", "detail", nil)
			}
			tt := ui.NewTester(a.View, 800, 560)
			tt.Frame()
			label := "Cancel"
			if mode == "model" {
				label = "Model picker search"
			}
			r, ok := tt.Find(label)
			if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 800 || r.Y+r.H > 560 {
				t.Fatalf("%s is clipped at minimum size: %+v, found=%v", label, r, ok)
			}
			if mode != "model" {
				if e := tt.Click(label); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}

func TestModalTransitionRestoresComposerFocus(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	a.paletteOpen = true
	a.paletteQuery = "rename"
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if !a.promptOpen || a.paletteOpen || !tt.Focused("Rename task") {
		t.Fatal("palette-to-prompt focus transition failed")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.promptOpen || !tt.Focused("Message") {
		t.Fatal("composer focus not restored after nested modal dismissal")
	}
}

func TestFractionalScaleKeepsLogicalComposerGeometry(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	base, ok := tt.Find("Message")
	if !ok {
		t.Fatal("missing composer")
	}
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		tt.SetScale(scale)
		r, ok := tt.Find("Message")
		if !ok {
			t.Fatal("composer missing at scale", scale)
		}
		for i, d := range []float32{r.X - base.X, r.Y - base.Y, r.W - base.W, r.H - base.H} {
			if math.Abs(float64(d)) > 1 {
				t.Fatalf("scale %g changed DIP dimension %d by %g", scale, i, d)
			}
		}
		if tt.Image().Bounds().Dx() != int(math.Ceil(float64(1470*scale))) {
			t.Fatal("wrong raster width at scale", scale)
		}
	}
}

func TestHiddenSidebarReservesNativeTrafficLights(t *testing.T) {
	a := fixture(t)
	a.SidebarVisible = false
	tt := ui.NewTester(a.View, 1200, 800)
	tt.SetTitleBar(ui.TitleBar{Left: 80, Height: 44})
	tt.Frame()
	r, ok := tt.Find("Show sidebar")
	if !ok || r.X < 80 {
		t.Fatalf("header controls overlap native traffic lights: %+v", r)
	}
}
