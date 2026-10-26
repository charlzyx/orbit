package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestMountedFileHeldSnapshotAfterRealCloseDoesNotReviveEditor(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	docs := a.files.editor.Documents()
	snapshot, ok := a.files.editor.Active()
	if !ok {
		t.Fatal("no active snapshot")
	}
	a.closeFile(path)
	finish := watcherCompletion(t, a) // Real asynchronous domain Close completed.
	finish()                          // Real close completion prunes its native state.
	if len(a.files.editor.Documents()) != 0 || a.files.states[path] != nil {
		t.Fatal("real close did not settle")
	}
	// Resume the same production render body with its earlier Active snapshot.
	// Controlled boundary, not a claim to recreate the exact hosted scheduler.
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Children(func() { a.fileEditorDocument(c, docs, snapshot) })
	}, 1000, 700)
	tt.Frame()
	if a.files.states[path] != nil || a.files.native[snapshot.ID] != nil || a.files.pathIDs[path] != 0 {
		t.Fatal("stale render resurrected closed native identity")
	}
}

func TestMountedFileStaleSnapshotCannotRenderReopenedPath(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	docs := a.files.editor.Documents()
	snapshot, _ := a.files.editor.Active()
	a.closeFile(path)
	watcherCompletion(t, a)()
	a.OpenFile(path)
	drainRegressionUpdates(t, a, func() bool {
		d, ok := a.files.editor.Active()
		return ok && d.ID != snapshot.ID && a.files.states[path] != nil
	})
	live, _ := a.files.editor.Active()
	st := a.files.states[path]
	stale := true
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Children(func() {
			if stale {
				a.fileEditorDocument(c, docs, snapshot)
			} else {
				a.fileEditorContent(c)
			}
		})
	}, 1000, 700)
	tt.Frame()
	if _, found := tt.Find("File editor"); found {
		t.Fatal("stale snapshot rendered the replacement path's editor")
	}
	if a.files.states[path] != st || a.files.pathIDs[path] != live.ID || st.Buffer.String() != "original\n" {
		t.Fatal("stale snapshot altered replacement identity")
	}
	stale = false
	tt.Frame()
	if _, found := tt.Find("File editor"); !found {
		t.Fatal("fresh frame did not recover live editor")
	}
	if a.files.states[path] != st {
		t.Fatal("fresh frame replaced live native state")
	}
}
