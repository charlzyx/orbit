package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeSyncRejectsClosedAndRenamedSnapshots(t *testing.T) {
	for _, mode := range []string{"closed", "renamed_again"} {
		t.Run(mode, func(t *testing.T) {
			a, _, old := mountedFile(t, "original\n")
			snapshot, _ := a.files.editor.Active()
			if mode == "closed" {
				a.closeFile(old)
				drainRegressionUpdates(t, a, func() bool { return a.files.states[old] == nil })
			} else {
				first, err := workspaceRenameProbe(a, old, "first.txt")
				if err != nil {
					t.Fatal(err)
				}
				snapshot, _ = a.files.editor.Active()
				a.bindFileNative(snapshot)
				old = first
				next, err := workspaceRenameProbe(a, old, "new.txt")
				if err != nil {
					t.Fatal(err)
				}
				d, _ := a.files.editor.Active()
				a.bindFileNative(d)
				if a.files.states[next] == nil {
					t.Fatal("live target missing")
				}
			}
			a.syncDocument(snapshot)
			if a.files.pathIDs[old] == snapshot.ID || a.files.states[old] != nil {
				t.Fatal("stale snapshot resurrected old native binding")
			}
			if mode == "closed" && a.files.native[snapshot.ID] != nil {
				t.Fatal("stale snapshot resurrected closed ID")
			}
		})
	}
}

func TestOpenCompletionPreservesAlreadyMountedPendingState(t *testing.T) {
	a, _, old := mountedFile(t, "original\n")
	p := filepath.Join(filepath.Dir(old), "other.txt")
	if err := os.WriteFile(p, []byte("disk\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a.OpenFile(p)
	apply := watcherCompletion(t, a)
	var found bool
	for _, d := range a.files.editor.Documents() {
		if d.Content.Path == p {
			a.syncDocument(d)
			found = true
		}
	}
	if !found {
		t.Fatal("real OpenInactive did not register document")
	}
	// Explicit adapter seam: pre-existing native state acquires a late edit.
	// No claim an OS key sequence can focus this inactive tab without changing generation.
	st := a.files.states[p]
	ack := a.files.versions[p]
	st.Buffer.Set("native pending\n")
	apply()
	if st.Buffer.String() != "native pending\n" || a.files.versions[p] != ack {
		t.Fatal("real queued OpenFile completion replaced/acknowledged pending native state")
	}
}

func workspaceRenameProbe(a *App, old, name string) (string, error) {
	p, err := workspace.RenamePath(old, name)
	if err == nil {
		a.files.editor.RenameUnder(old, p)
	}
	return p, err
}

func TestRenameInvalidatesQueuedOldPathOpenIntent(t *testing.T) {
	a, _, existing := mountedFile(t, "existing\n")
	old := filepath.Join(filepath.Dir(existing), "queued.txt")
	if err := os.WriteFile(old, []byte("queued original"), 0600); err != nil {
		t.Fatal(err)
	}
	a.OpenFile(old)
	oldApply := watcherCompletion(t, a)
	generation := a.files.openGeneration
	r := a.captureFileRename(old)
	renamed, err := workspaceRenameProbe(a, old, "renamed.txt")
	if err != nil {
		t.Fatal(err)
	}
	a.applyFileRename(r, renamed)
	// Old namespace now names another actual registered document. The older
	// OpenFile completion must not activate this unrelated replacement ID.
	if err := os.WriteFile(old, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.files.editor.OpenInactive(old)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := a.files.editor.Active()
	focus := a.files.focusPending
	oldApply()
	after, _ := a.files.editor.Active()
	if after.ID == replacement.ID || after.ID != before.ID || a.files.focusPending != focus {
		t.Fatal("old open completion took focus for replacement oldpath")
	}
	if a.files.desired != renamed || a.files.openGeneration == generation {
		t.Fatal("rename did not invalidate/remap accepted old-path navigation intent")
	}
}

func TestNativeReadonlyErrorPendingDoesNotAdvanceRecord(t *testing.T) {
	for _, mode := range []string{"lossy", "missing"} {
		t.Run(mode, func(t *testing.T) {
			a, _, p := mountedFile(t, "original\n")
			d, _ := a.files.editor.Active()
			st := a.files.states[p]
			ack := a.files.versions[p]
			st.Buffer.Set("pending\n")
			var err error
			if mode == "lossy" {
				err = os.WriteFile(p, []byte{0xff}, 0600)
			} else {
				err = os.Remove(p)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := a.files.editor.Reload(p); err != nil {
				t.Fatal(err)
			}
			d, _ = a.files.editor.Active()
			a.syncDocument(d)
			if st.Buffer.String() != "pending\n" || a.files.versions[p] != ack || a.files.native[d.ID].version != ack {
				t.Fatal("readonly/error pending overwritten or record acknowledged")
			}
		})
	}
}

func TestRenameImageNativeCacheBelongsToDocumentID(t *testing.T) {
	a, _, existing := mountedFile(t, "existing\n")
	old := filepath.Join(filepath.Dir(existing), "image.png")
	writeImage := func(p string, c color.RGBA) {
		t.Helper()
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		im := image.NewRGBA(image.Rect(0, 0, 2, 2))
		im.SetRGBA(0, 0, c)
		if err := png.Encode(f, im); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	writeImage(old, color.RGBA{R: 255, A: 255})
	d, err := a.files.editor.Open(old)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	original := a.files.images[old]
	if original == nil {
		t.Fatal("real PNG did not decode")
	}
	receipt := a.captureFileRename(old)
	renamed, err := workspaceRenameProbe(a, old, "renamed.png")
	if err != nil {
		t.Fatal(err)
	}
	writeImage(old, color.RGBA{B: 255, A: 255})
	fresh, err := a.files.editor.Open(old)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(fresh)
	replacement := a.files.images[old]
	if replacement == nil || replacement == original {
		t.Fatal("new document reused image source")
	}
	a.applyFileRename(receipt, renamed)
	if a.files.images[renamed] != original || a.files.images[old] != replacement {
		t.Fatal("receipt crossed image ownership")
	}
	a.closeFile(old)
	drainRegressionUpdates(t, a, func() bool { return a.files.states[old] == nil })
	if a.files.native[fresh.ID] != nil {
		t.Fatal("closed image ID retained cache")
	}
	a.syncDocument(fresh)
	if a.files.images[old] != nil || a.files.native[fresh.ID] != nil {
		t.Fatal("closed snapshot revived image cache")
	}
}
