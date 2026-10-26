package orbitui

import (
	"os"
	"testing"
)

func TestRecoveryCloseBeforeRenameReceiptRejectsReplacement(t *testing.T) {
	a, _, old := mountedFile(t, "original\n")
	s := a.files
	original, _ := s.editor.Active()
	job := a.admitFileClose(old)
	if job == nil {
		t.Fatal("admission refused")
	}
	a.OpenFile(old)
	r := a.captureFileRename(old)
	next, err := s.editor.RenameDocuments(r.request("moved"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	replacement, err := s.editor.OpenInactive(old)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(replacement)
	generation := s.openGeneration
	s.focusPending = ""
	a.Async(job)
	watcherCompletion(t, a)()
	a.applyFileRename(r, next.Path)
	for _, d := range s.editor.Documents() {
		if d.ID == original.ID {
			t.Fatal("original not closed")
		}
	}
	if s.pathIDs[old] != replacement.ID {
		t.Fatal("replacement metadata changed")
	}
	if s.openGeneration != generation || s.focusPending != "" {
		t.Fatalf("old identity reopened replacement: original=%d replacement=%d generation=%d->%d focus=%q", original.ID, replacement.ID, generation, s.openGeneration, s.focusPending)
	}
}
func TestRecoveryAdmittedCloseJobTargetsOriginalID(t *testing.T) {
	for _, mode := range []string{"normal", "reopen_before_rename", "reopen_after_rename", "new_navigation"} {
		t.Run(mode, func(t *testing.T) {
			a, _, old := mountedFile(t, "original\n")
			s := a.files
			original, _ := s.editor.Active()
			oldState := s.states[old]
			job := a.admitFileClose(old)
			if job == nil {
				t.Fatal("admission refused")
			}
			if mode == "reopen_before_rename" {
				a.OpenFile(old)
			}
			r := a.captureFileRename(old)
			next, err := s.editor.RenameDocuments(r.request("moved"))
			if err != nil {
				t.Fatal(err)
			}
			a.applyFileRename(r, next.Path)
			if s.pathIDs[next.Path] != original.ID || s.states[next.Path] != oldState {
				t.Fatal("rename lost original buffer")
			}
			if err := os.WriteFile(old, []byte("replacement"), 0600); err != nil {
				t.Fatal(err)
			}
			replacement, err := s.editor.OpenInactive(old)
			if err != nil {
				t.Fatal(err)
			}
			a.syncDocument(replacement)
			st := s.states[old]
			if mode == "reopen_after_rename" {
				a.OpenFile(old)
			}
			if mode == "new_navigation" {
				a.Navigate("settings")
				watcherCompletion(t, a)()
			}
			st.Buffer.Set("replacement pending")
			ack := s.versions[old]
			generation := s.openGeneration
			s.focusPending = ""
			a.Async(job)
			watcherCompletion(t, a)()
			for _, d := range s.editor.Documents() {
				if d.ID == original.ID {
					t.Fatal("original renamed ID remained")
				}
			}
			if s.native[original.ID] != nil || s.pathIDs[old] != replacement.ID || s.states[old] != st || st.Buffer.String() != "replacement pending" || s.versions[old] != ack {
				t.Fatal("close damaged replacement ID/pending")
			}
			raw, err := os.ReadFile(old)
			if err != nil || string(raw) != "replacement" {
				t.Fatal("close saved replacement draft", string(raw), err)
			}
			if mode == "reopen_before_rename" && (s.focusPending != "" || s.openGeneration != generation) {
				t.Fatal("old reopen intent retargeted")
			}
			if mode == "new_navigation" && (a.page != "settings" || s.focusPending != "") {
				t.Fatal("old close stole navigation")
			}
			if mode == "reopen_after_rename" {
				active, _ := s.editor.Active()
				if active.ID != replacement.ID || s.focusPending != old {
					t.Fatal("explicit replacement intent failed")
				}
			}
		})
	}
}
func TestRecoveryNormalReopenAndNoopReceipt(t *testing.T) {
	for _, mode := range []string{"before_worker", "after_worker", "noop"} {
		t.Run(mode, func(t *testing.T) {
			a, _, p := mountedFile(t, "original\n")
			s := a.files
			original, _ := s.editor.Active()
			job := a.admitFileClose(p)
			if mode == "before_worker" {
				a.OpenFile(p)
			}
			if mode == "noop" {
				if err := s.editor.CloseID(original.ID); err != nil {
					t.Fatal(err)
				}
			}
			a.Async(job)
			apply := watcherCompletion(t, a)
			if mode != "before_worker" {
				a.OpenFile(p)
			}
			apply()
			if mode == "noop" {
				if len(s.editor.Documents()) != 0 || s.states[p] != nil {
					t.Fatal("noop forged reopen")
				}
				select {
				case <-a.updates:
					t.Fatal("noop queued reopen")
				default:
				}
				return
			}
			watcherCompletion(t, a)()
			d, _ := s.editor.Active()
			if d.ID == 0 || d.ID == original.ID || d.Content.Path != p || s.states[p] == nil || s.states[p].Buffer.String() != "original\n" || s.focusPending != p {
				t.Fatal("normal reopen failed", d)
			}
		})
	}
}
func TestRecoveryExplicitReplacementNeverSelectsNextID(t *testing.T) {
	a, _, p := mountedFile(t, "original")
	s := a.files
	job := a.admitFileClose(p)
	a.Async(job)
	apply := watcherCompletion(t, a)
	replacement, err := s.editor.OpenInactive(p)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(replacement)
	a.OpenFile(p)
	if err := s.editor.CloseID(replacement.ID); err != nil {
		t.Fatal(err)
	}
	next, err := s.editor.OpenInactive(p)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(next)
	s.focusPending = ""
	apply()
	if s.focusPending != "" || s.pathIDs[p] != next.ID {
		t.Fatal("missing requested ID selected next occupant")
	}
}
func TestRecoveryReopenCompletionRespectsOwnerAndNavigation(t *testing.T) {
	for _, mode := range []string{"owner", "navigation"} {
		t.Run(mode, func(t *testing.T) {
			a, _, p := mountedFile(t, "original")
			s := a.files
			job := a.admitFileClose(p)
			a.OpenFile(p)
			a.Async(job)
			watcherCompletion(t, a)()
			finish := watcherCompletion(t, a)
			if mode == "owner" {
				a.files = nil
				a.initFiles()
			} else {
				a.Navigate("settings")
				watcherCompletion(t, a)()
			}
			current := a.files
			current.focusPending = "new-intent"
			page := a.page
			finish()
			if a.files != current || a.page != page || current.focusPending != "new-intent" {
				t.Fatal("old completion stole intent")
			}
			if mode == "owner" {
				s.editor.CloseAll()
			}
		})
	}
}
func TestRecoveryCloseCompletionPreservesReplacementOwner(t *testing.T) {
	a, _, p := mountedFile(t, "original")
	old := a.files
	a.closeFile(p)
	apply := watcherCompletion(t, a)
	a.files = nil
	a.initFiles()
	s := a.files
	d, err := s.editor.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	st := s.states[p]
	st.Buffer.Set("new owner pending")
	ack := s.versions[p]
	s.focusPending = "new-owner"
	apply()
	if a.files != s || s.states[p] != st || st.Buffer.String() != "new owner pending" || s.versions[p] != ack || s.focusPending != "new-owner" {
		t.Fatal("old close mutated new owner")
	}
	if len(old.editor.Documents()) != 0 {
		t.Fatal("old close incomplete")
	}
}
