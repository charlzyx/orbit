package orbitui

import (
	"bytes"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui/editor"
	"os"
	"testing"
)

func TestFileEditorFailedSyncRetainsVersionAndRetries(t *testing.T) {
	a, tt, path := mountedFile(t, "original\n")
	st := a.files.states[path]
	observed := a.files.versions[path]
	st.Buffer.Edit(editor.Edit{Start: 0, End: len("original\n"), Text: "pending native\n"})
	pending := st.Buffer.Version()
	if pending == observed {
		t.Fatal("fixture has no pending native edit")
	}
	// A real external lossy-file transition makes the real Editor reject edits.
	// Avoid watcher syncDocument here: this fixture isolates the renderer's
	// version-acknowledgement boundary, not the entire watcher lifecycle.
	invalid := []byte{0xff, '\n'}
	if err := os.WriteFile(path, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.files.editor.Reload(path); err != nil {
		t.Fatal(err)
	}
	before, _ := a.files.editor.Active()
	if !before.Content.Lossy || before.Content.Editable() {
		t.Fatal("fixture did not become genuinely read-only", before)
	}
	tt.Frame()
	failed, _ := a.files.editor.Active()
	if failed.Text != before.Text || !failed.Content.Lossy {
		t.Fatal("failed sync altered domain text/mode", failed)
	}
	t.Logf("failure: domain=%q native=%q dirty=%v observed=%d current=%d", failed.Text, st.Buffer.String(), failed.Dirty, a.files.versions[path], st.Buffer.Version())
	if a.files.versions[path] != observed {
		t.Error("failed SetText acknowledged pending buffer version")
	}
	if st.Buffer.String() != "pending native\n" || !st.Buffer.CanUndo() {
		t.Fatal("failed sync lost native text/undo")
	}
	found := false
	for _, toast := range a.toastStack.Items() {
		if toast.Body == workspace.ErrReadOnly.Error() {
			found = true
		}
	}
	if !found {
		t.Fatal("real SetText ErrReadOnly was not surfaced")
	}
	raw, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, invalid) {
		t.Fatal("failed edit wrote read-only content", raw, err)
	}
	if err := os.WriteFile(path, []byte("restored\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.files.editor.Reload(path); err != nil {
		t.Fatal(err)
	}
	restored, _ := a.files.editor.Active()
	if !restored.Content.Editable() {
		t.Fatal("fixture did not recover")
	}
	tt.Frame()
	after, _ := a.files.editor.Active()
	t.Logf("retry: domain=%q native=%q dirty=%v observed=%d current=%d", after.Text, st.Buffer.String(), after.Dirty, a.files.versions[path], st.Buffer.Version())
	if after.Text != "pending native\n" {
		t.Errorf("legal next-frame retry lost pending edit: %q", after.Text)
	}
	if a.files.versions[path] != pending {
		t.Error("successful retry did not acknowledge buffer")
	}
	if st.Buffer.String() != "pending native\n" || !st.Buffer.CanUndo() {
		t.Fatal("retry lost native undo")
	}
}
