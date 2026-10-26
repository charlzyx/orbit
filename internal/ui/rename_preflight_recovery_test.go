package orbitui

import (
	"errors"
	"github.com/egoist/mygo/ui/editor"
	"os"
	"path/filepath"
	"testing"
)

// Exercise the actual mounted editor buffer/commit seam. Buffer.Edit is used
// deliberately; this is not a claim about a captured OS key-event interleaving.
func TestRenamePreflightRecoveryPendingNativeBuffer(t *testing.T) {
	for _, inline := range []bool{true, false} {
		t.Run(map[bool]string{true: "inline", false: "prompt"}[inline], func(t *testing.T) {
			a, _, old := mountedFile(t, "original\n")
			if inline {
				a.beginExplorerEntry(filepath.Dir(old), old, false)
			} else {
				a.renameFilePrompt(old)
			}
			st := a.files.states[old]
			st.Buffer.Edit(editor.Edit{Start: 0, End: len("original\n"), Text: "pending edit\n"})
			if a.files.editor.HasDirtyUnder(old) {
				t.Fatal("fixture domain already observed native edit")
			}
			if inline {
				entry := a.files.entry
				entry.value = "renamed.txt"
				a.commitExplorerEntry(entry)
			} else {
				a.promptCommit("renamed.txt")
			}
			if a.files.mutating {
				t.Error("rename accepted unobserved pending native edits")
				drainRegressionUpdates(t, a, func() bool { return !a.files.mutating })
				return
			}
			if inline && (a.files.entry == nil || a.files.entry.err == "") {
				t.Fatal("inline dirty refusal missing")
			}
			d, ok := a.files.editor.Active()
			if !ok || d.Content.Path != old || d.Text != "pending edit\n" {
				t.Fatal("pending buffer not synchronized", d)
			}
			if st.Buffer.String() != "pending edit\n" || !st.Buffer.CanUndo() {
				t.Fatal("native text/undo discarded")
			}
			raw, err := os.ReadFile(old)
			if err != nil {
				t.Fatal("original path disappeared", err)
			}
			// Normal debounce autosave remains allowed. Do not turn a scheduler pause
			// into a spurious failure by requiring it never to save after500ms.
			if string(raw) != "original\n" && string(raw) != "pending edit\n" {
				t.Fatal("unexpected original-path bytes", string(raw))
			}
			t.Logf("rename refused; original-path bytes=%q dirty=%v", raw, d.Dirty)
			if _, err := os.Stat(filepath.Join(filepath.Dir(old), "renamed.txt")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("rename created destination", err)
			}
		})
	}
}
