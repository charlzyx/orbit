package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func chromeStyleFixture(t *testing.T) *App {
	a := sourceSidebarFixture(t)
	a.Workspaces.GroupBy = "workspace"
	a.Model.SessionName = "Native UI comparison fixture"
	a.Model.Connected = true
	a.OpenApps = []native.ExternalApp{{ID: "finder", Label: "Finder", Target: "com.apple.finder"}}
	row := transcriptui.Row{ID: "answer", Role: "assistant", Steps: []transcriptui.Step{{Text: "## Implementation verified\n\nThis is a synthetic visual fixture, not a user conversation.\n\n- Keep native input and clipboard behavior.\n- Preserve existing data and session ownership.", Thinking: "A synthetic thought used only to test the disclosure layout."}}}
	a.Model.Messages = []Message{{ID: "answer", Role: "assistant", Logical: &row, Text: row.Text()}, {ID: "changed", Role: "changedFiles", ChangedFiles: []transcriptui.ChangedFile{{Path: "internal/ui/chat.go", Added: 24, Removed: 8}, {Path: "internal/ui/sidebar.go", Added: 12, Removed: 3}, {Path: "README.md", Added: 4}, {Path: "main.go", Removed: 1}}}}
	return a
}
func TestChromeStyleCPUComparison(t *testing.T) {
	for _, mode := range []string{"light", "dark"} {
		t.Run(mode, func(t *testing.T) {
			a := chromeStyleFixture(t)
			a.Appearance.Mode = mode
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			if out := os.Getenv("ORBIT_CHROME_CAPTURE_DIR"); out != "" {
				if err := os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(filepath.Join(out, "chrome-"+mode+".png"))
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				if err = png.Encode(f, tt.Image()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
func BenchmarkChromeFrameInput(b *testing.B) {
	if err := LoadAssets(); err != nil {
		b.Fatal(err)
	}
	for _, n := range []int{0, 1000} {
		for _, input := range []bool{false, true} {
			b.Run(fmt.Sprintf("history%d/input%t", n, input), func(b *testing.B) {
				store := preferences.New(b.TempDir())
				prefs := preferences.DefaultUI()
				prefs.Language = "en"
				prefs.ReduceMotion = true
				if err := store.SaveUI(prefs); err != nil {
					b.Fatal(err)
				}
				a := New(Options{Store: store})
				msgs := make([]Message, n)
				for i := range msgs {
					msgs[i] = Message{ID: fmt.Sprint(i), Role: "assistant", Text: "Synthetic history text with **formatting** and a stable layout."}
				}
				a.SetSnapshot(Snapshot{SessionID: "bench", Messages: msgs})
				tt := ui.NewTester(a.View, 1200, 800)
				tt.Frame()
				if err := tt.Click("Message"); err != nil {
					b.Fatal(err)
				}
				tt.Type("fixture")
				tt.Advance(time.Second)
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					if input {
						tt.Type("x")
						tt.Key(0, ui.KeyBackspace)
					} else {
						tt.Frame()
					}
				}
			})
		}
	}
}
