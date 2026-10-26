package orbitui

import (
	"os"
	"path/filepath"
	"testing"
)

// Current-source RED only: real UI rename worker, explicit completion ordering.
func TestRenameOldPathReopenBeforeCompletionIdentityProbe(t *testing.T) {
	for _, render := range []bool{false, true} {
		t.Run(map[bool]string{false: "no_intermediate_render", true: "render_before_reopen"}[render], func(t *testing.T) {
			a, tt, old := mountedFile(t, "original renamed document\n")
			original := a.files.states[old]
			renamed := filepath.Join(filepath.Dir(old), "renamed.txt")
			a.renameFilePrompt(old)
			a.promptCommit("renamed.txt")
			apply := watcherCompletion(t, a) // Real filesystem + domain rename done, UI reconcile held.
			d, ok := a.files.editor.Active()
			if !ok || d.Content.Path != renamed {
				t.Fatal("rename domain did not complete", d)
			}
			if render {
				tt.Frame()
			}
			if err := os.WriteFile(old, []byte("new old-path document\n"), 0600); err != nil {
				t.Fatal(err)
			}
			a.OpenFile(old)
			drainRegressionUpdates(t, a, func() bool {
				d, ok := a.files.editor.Active()
				return ok && d.Content.Path == old && a.files.states[old] != nil && a.files.states[old].Buffer.String() == "new old-path document\n"
			})
			reopened := a.files.states[old]
			t.Logf("before old callback: originalNative=%q reusedByReopen=%v renamedCacheExists=%v", original.Buffer.String(), original == reopened, a.files.states[renamed] != nil)
			if reopened == original {
				t.Error("new oldpath document reused original renamed native state before callback")
			}
			apply()
			t.Logf("after callback: oldCacheExists=%v renamedCacheIsReopened=%v desired=%q", a.files.states[old] != nil, a.files.states[renamed] == reopened, a.files.desired)
			if a.files.states[old] != reopened {
				t.Error("old callback moved the newly opened oldpath state")
			}
			if a.files.states[renamed] != original || original.Buffer.String() != "original renamed document\n" {
				t.Error("old callback did not preserve original renamed native identity/text")
			}
			for path, want := range map[string]string{old: "new old-path document\n", renamed: "original renamed document\n"} {
				raw, err := os.ReadFile(path)
				if err != nil || string(raw) != want {
					t.Errorf("disk %s=%q err=%v", filepath.Base(path), raw, err)
				}
			}
			// No editor frame after the observed corruption: do not turn this state
			// identity probe into an accidental write of the wrong native draft.
		})
	}
}

func TestRenameReceiptPreservesPendingUndoAndNewNavigation(t *testing.T) {
	a, _, old := mountedFile(t, "original\n")
	original := a.files.states[old]
	d, _ := a.files.editor.Active()
	id := d.ID
	a.renameFilePrompt(old)
	a.promptCommit("renamed.txt")
	apply := watcherCompletion(t, a)
	// A native edit after preflight is still owned by the original document.
	original.Buffer.Set("late native draft\n")
	pending := original.Buffer.Version()
	if err := os.WriteFile(old, []byte("new occupant\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a.OpenFile(old)
	drainRegressionUpdates(t, a, func() bool {
		d, ok := a.files.editor.Active()
		return ok && d.Content.Path == old && d.ID != id && a.files.states[old] != nil
	})
	newer := a.files.states[old]
	generation := a.files.openGeneration
	focus := a.files.focusPending
	apply()
	renamed := filepath.Join(filepath.Dir(old), "renamed.txt")
	if a.files.states[renamed] != original || original.Buffer.String() != "late native draft\n" || !original.Buffer.CanUndo() {
		t.Fatal("receipt lost original pending/undo")
	}
	if a.files.versions[renamed] == pending {
		t.Fatal("receipt acknowledged pending edit")
	}
	if a.files.states[old] != newer || newer == original || newer.Buffer.String() != "new occupant\n" {
		t.Fatal("receipt reused new occupant state")
	}
	if a.files.desired != old || a.files.focusPending != focus || a.files.openGeneration != generation {
		t.Fatal("old receipt changed newer navigation")
	}
	if !a.syncFileBuffers() {
		t.Fatal("ID native synchronization failed")
	}
	if !original.Buffer.Undo() || original.Buffer.String() != "original\n" || newer.Buffer.String() != "new occupant\n" {
		t.Fatal("Undo no longer belongs to original ID")
	}
	if !original.Buffer.Redo() || original.Buffer.String() != "late native draft\n" {
		t.Fatal("Redo lost original draft")
	}

	for _, d := range a.files.editor.Documents() {
		if d.ID == id && d.Text != "late native draft\n" {
			t.Fatal("ID native synchronization targeted wrong doc")
		}
	}
}

func TestRenameOpenExistingIDRetainsPendingBeforeCallback(t *testing.T) {
	a, _, old := mountedFile(t, "original\n")
	st := a.files.states[old]
	a.renameFilePrompt(old)
	a.promptCommit("renamed.txt")
	apply := watcherCompletion(t, a)
	st.Buffer.Set("native pending\n")
	version := st.Buffer.Version()
	renamed := filepath.Join(filepath.Dir(old), "renamed.txt")
	a.OpenFile(renamed) // Existing domain ID, still missing its new path index.
	if a.files.states[renamed] != st || st.Buffer.String() != "native pending\n" || a.files.versions[renamed] == version {
		t.Fatal("existing ID open reset pending native state")
	}
	apply()
	if a.files.states[renamed] != st || st.Buffer.String() != "native pending\n" {
		t.Fatal("completion reset pending native state")
	}
}
