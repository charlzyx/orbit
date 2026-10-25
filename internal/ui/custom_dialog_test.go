package orbitui

import (
	"os"
	"sync"
	"testing"
	"time"

	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	ghostty "github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

func TestCustomDialogDoesNotDismissEscape(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "custom", Method: "custom", Title: "Component"}})
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if len(actions) != 0 || a.Model.Dialog == nil {
		t.Fatalf("Escape dismissed custom component: %v", actions)
	}
	if e := tt.Click("Cancel"); e != nil {
		t.Fatal(e)
	}
	if len(actions) != 1 || actions[0].Kind != "RespondDialog" || actions[0].Data["cancelled"] != true {
		t.Fatalf("explicit cancellation missing: %v", actions)
	}
}

func TestLiveCustomDialogForwardsEscape(t *testing.T) {
	if os.Getenv("ORBIT_TERMINAL_TEST") != "1" {
		t.Skip("set ORBIT_TERMINAL_TEST=1 for the pinned Ghostty native input integration")
	}
	if e := ghostty.Load(); e != nil {
		t.Fatal(e)
	}
	a := fixture(t)
	var mu sync.Mutex
	var inputs []string
	p := orbiterm.NewCustom(orbiterm.CustomOptions{OnInput: func(id, data string) error {
		mu.Lock()
		defer mu.Unlock()
		inputs = append(inputs, id+":"+data)
		return nil
	}})
	defer p.Close()
	a.Options.Custom = p
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "custom", Method: "custom", Title: "Component"}})
	p.Update("custom", []string{"Select a choice"}, 96, 20, nil)
	deadline := time.Now().Add(10 * time.Second)
	for p.ActiveTerminal() == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if p.ActiveTerminal() == nil {
		t.Fatal("custom renderer did not initialize")
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	for time.Now().Before(deadline) {
		mu.Lock()
		ready := len(inputs) > 0
		mu.Unlock()
		if ready {
			break
		}
		time.Sleep(time.Millisecond)
	}
	mu.Lock()
	got := append([]string(nil), inputs...)
	mu.Unlock()
	if len(got) != 1 || got[0] != "custom:\x1b" {
		t.Fatalf("Escape did not reach Pi component: %q", got)
	}
	if len(actions) != 0 || a.Model.Dialog == nil {
		t.Fatalf("Escape cancelled component: %v", actions)
	}
}
