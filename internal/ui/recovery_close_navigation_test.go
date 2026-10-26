package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"os"
	"path/filepath"
	"testing"
)

func TestRecoveryCloseCallbackPreservesNewIntent(t *testing.T) {
	for _, mode := range []string{"settings", "chat", "new_file", "reopen", "reopen_settings", "history_aba", "reopen_history_aba"} {
		t.Run(mode, func(t *testing.T) {
			a, _, old := mountedFile(t, "original\n")
			s := a.files
			original, _ := s.editor.Active()
			neighbor := filepath.Join(filepath.Dir(old), "neighbor")
			if err := os.WriteFile(neighbor, []byte("neighbor"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := s.editor.OpenInactive(neighbor); err != nil {
				t.Fatal(err)
			}
			if mode == "history_aba" || mode == "reopen_history_aba" {
				a.Navigate("settings")
				watcherCompletion(t, a)()
				a.Navigate("files")
			}
			a.closeFile(old)
			apply := watcherCompletion(t, a)
			r := a.captureFileRename(old)
			next, err := workspace.RenamePath(old, "moved")
			if err != nil {
				t.Fatal(err)
			}
			s.editor.RenameUnder(old, next)
			a.applyFileRename(r, next)
			if err := os.WriteFile(old, []byte("replacement"), 0600); err != nil {
				t.Fatal(err)
			}
			replacement, err := s.editor.OpenInactive(old)
			if err != nil {
				t.Fatal(err)
			}
			a.syncDocument(replacement)
			st := s.states[old]
			switch mode {
			case "reopen":
				a.OpenFile(old)
			case "reopen_settings":
				a.OpenFile(old)
				a.Navigate("settings")
			case "new_file":
				a.OpenFile(neighbor)
			case "history_aba":
				a.goHistory(-1)
				a.goHistory(1)
				watcherCompletion(t, a)()
			case "reopen_history_aba":
				a.OpenFile(old)
				a.goHistory(-1)
				a.goHistory(1)
				watcherCompletion(t, a)()
			default:
				a.Navigate(mode)
			}
			st.Buffer.Set("pending replacement\n")
			ack := s.versions[old]
			generation := s.openGeneration
			page := a.page
			desired := s.desired
			s.focusPending = ""
			apply()
			if s.native[original.ID] != nil || s.pathIDs[old] != replacement.ID || s.states[old] != st || st.Buffer.String() != "pending replacement\n" || s.versions[old] != ack {
				t.Fatal("old close damaged replacement identity/native pending")
			}
			if mode == "reopen" {
				active, _ := s.editor.Active()
				if active.ID != replacement.ID || s.focusPending != old {
					t.Fatal("explicit replacement reopen failed")
				}
			} else if a.page != page || s.desired != desired || s.openGeneration != generation || s.focusPending != "" {
				t.Fatalf("old close stole intent: mode=%s page=%s focus=%q", mode, a.page, s.focusPending)
			}
		})
	}
}
