package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestSourceCodeCaptionsPreserveUnknownAndTokenizerTag(t *testing.T) {
	a := fixture(t)
	for _, tc := range []struct{ tag, want string }{{" TS ", "TypeScript"}, {"rs", "Rust"}, {"jsx", "JSX"}, {"cs", "C#"}, {"scala", "scala"}, {"objc", "objc"}, {"brainfuck", "brainfuck"}, {"  Unknown  ", "  Unknown  "}, {"Kt", "Kt"}} {
		if got := a.codeCaption(tc.tag); got != tc.want {
			t.Fatalf("%q caption %q want %q", tc.tag, got, tc.want)
		}
	}
	for tag, want := range sourceCodeCaptions {
		if got := a.codeCaption(tag); got != want {
			t.Fatal(tag, got, want)
		}
	}
}

func TestMountedCodeSourceCaptionsAndLocale(t *testing.T) {
	for _, tc := range []struct{ tag, want string }{{"ts", "TypeScript"}, {"rs", "Rust"}, {"", "Plain text"}, {"Mermaid", "Mermaid · source (preview unavailable)"}, {"brainfuck", "brainfuck"}} {
		t.Run(tc.tag, func(t *testing.T) {
			a := fixture(t)
			a.SetSnapshot(Snapshot{SessionID: "caption", Messages: []Message{{ID: "code", Role: "assistant", Text: "```" + tc.tag + "\nraw payload\n```"}}})
			tt := ui.NewTester(a.View, 1470, 922)
			if !tt.HasText(tc.want) {
				t.Fatal("caption missing", tc.want)
			}
			if err := tt.Click("Copy code"); err != nil {
				t.Fatal(err)
			}
			if tt.Clipboard() != "raw payload" {
				t.Fatal("caption affected copy", tt.Clipboard())
			}
			if tc.tag == "" || tc.tag == "Mermaid" {
				// Exercise the real Appearance language setter against fixture temp Store;
				// it updates preferences, reloads the catalog and persists the choice.
				a.Navigate("settings")
				a.selectSettingsPage("appearance")
				tt.Frame()
				if err := tt.Click("Language selector"); err != nil {
					t.Fatal(err)
				}
				if err := tt.Click("简体中文"); err != nil {
					t.Fatal(err)
				}
				if a.Prefs.Language != "zh-CN" || a.Options.Store.UI().Language != "zh-CN" {
					t.Fatal("real locale setter did not persist Chinese")
				}
				a.Navigate("chat")
				tt.Frame()
				key := "transcript_view.plain_text"
				if tc.tag == "Mermaid" {
					key = "transcript_view.mermaid_source"
				}
				translated := a.Tr(key)
				if translated == tc.want || !tt.HasText(translated) {
					t.Fatal("caption did not follow locale", translated)
				}
			}
		})
	}
}
