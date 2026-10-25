package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestComposerAddMenuCancelRestoresDraftFocus(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("before")
	if err := tt.Click("Add attachment"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.Frame()
	tt.Type(" after")
	if a.Draft != "before after" {
		t.Fatalf("cancel did not restore focus: %q", a.Draft)
	}
}
func TestComposerAddMenuPickerChoiceRestoresDraftFocus(t *testing.T) {
	for _, label := range []string{"Attach image…", "Attach files…", "Paste image"} {
		t.Run(label, func(t *testing.T) {
			a := fixture(t)
			a.Options.OnAction = func(Action) {}
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			tt.Type("before")
			if err := tt.Click("Add attachment"); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem(label); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			tt.Type(" after")
			if a.Draft != "before after" {
				t.Fatalf("choice did not restore focus: %q", a.Draft)
			}
		})
	}
}

func TestComposerAddMenuStaleChoiceCannotFocusUnrelatedPage(t *testing.T) {
	a := fixture(t)
	generation := uint64(1)
	a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, generation, true }
	hidden := false
	other := ""
	tt := ui.NewTester(func(c *ui.Context) {
		if hidden {
			ui.TextInput(c, &other).Label("Unrelated page input")
		} else {
			a.View(c)
		}
	}, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("draft")
	if err := tt.Click("Add attachment"); err != nil {
		t.Fatal(err)
	}
	hidden = true
	generation++
	tt.Frame()
	if err := tt.Click("Unrelated page input"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(a.text("composer.mention_file", "Mention file")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Type("still here")
	if other != "still here" || a.Draft != "draft" || len(a.pendingAttachmentPaths) != 0 {
		t.Fatal("stale menu took unrelated focus", other, a.Draft)
	}
	hidden = false
	tt.Frame()
	if a.Draft != "draft" {
		t.Fatal("late choice resurrected on return", a.Draft)
	}
}
