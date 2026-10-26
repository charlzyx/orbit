package orbitui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func providerSaveAuditFixture(t *testing.T) (*App, *ui.Tester, map[string]int) {
	t.Helper()
	a := fixture(t)
	a.Options.PiBinary = "/nonexistent-fixture-pi"
	a.Navigate("settings")
	a.settingsPage = "providers"
	counts := map[string]int{}
	a.Options.OnAction = func(action Action) { counts[action.Kind]++ }
	tt := ui.NewTester(a.View, 1000, 800)
	tt.Frame()
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading && !a.settingsLoading })
	tt.Frame()
	if err := tt.Click("Add provider…"); err != nil {
		t.Fatal(err)
	}
	d := a.providerState.editor
	if d == nil {
		t.Fatal("real Add provider did not open")
	}
	tt.Type("fixture-provider")
	d.edit.BaseURL = "https://example.invalid/v1"
	d.models = "fixture-model"
	tt.Frame()
	for k := range counts {
		delete(counts, k)
	}
	return a, tt, counts
}
func holdProviderSaveCompletion(t *testing.T, a *App, tt *ui.Tester) func() {
	t.Helper()
	if a.backdrop.loading || a.settingsLoading || len(a.updates) != 0 {
		t.Fatal("save fixture has unsettled initial work")
	}
	// A real Click settles multiple frames. An already completed error may
	// otherwise be drained before this helper can hold its UI callback.
	// drainUpdates uses len(queue), so an unbuffered queue stays unavailable
	// to rendered frames until this helper explicitly receives the callback.
	original := a.updates
	held := make(chan func())
	a.updates = held
	if err := tt.Click("Save provider"); err != nil {
		a.updates = original // No input was delivered when target lookup failed.
		t.Fatal(err)
	}
	select {
	case apply := <-a.updates:
		a.updates = original // Restore before apply can schedule refresh jobs.
		if a.providerState.editor == nil || !a.providerState.editor.busy {
			t.Fatal("actual save completion arrived without held busy editor")
		}
		return apply
	case <-time.After(3 * time.Second):
		// Do not close or restore a queue a late Async sender may still use.
		// Cleanup can discard an already-arriving callback without applying it;
		// a genuinely unsettled I/O remains a failure, not a cancellable job.
		t.Cleanup(func() {
			select {
			case <-held:
				t.Log("discarded late save completion; fixture queue left unchanged")
			default:
				t.Log("save I/O still unsettled; fixture queue left unchanged, no unsafe restore/close")
			}
		})
		t.Fatal("save did not finish actual I/O")
		return nil
	}
}
func dismissProviderSaveAudit(t *testing.T, a *App, tt *ui.Tester, how string) {
	t.Helper()
	if how == "escape" {
		tt.Key(0, ui.KeyEscape)
	} else {
		tt.ClickAt(5, 5)
	}
	if a.providerState.editor != nil {
		t.Fatal("actual dismissal did not close", how)
	}
	if err := tt.Click("Add provider…"); err != nil {
		t.Fatal(err)
	}
	tt.Type("new-unsaved")
	if a.providerState.editor == nil || a.providerState.editor.edit.ID != "new-unsaved" {
		t.Fatal("replacement draft not mounted")
	}
}
func TestAuditProviderSavedCompletionSurvivesDismissal(t *testing.T) {
	for _, how := range []string{"escape", "scrim", "same-editor"} {
		t.Run(how, func(t *testing.T) {
			a, tt, counts := providerSaveAuditFixture(t)
			old := a.providerState.editor
			apply := holdProviderSaveCompletion(t, a, tt)
			providers, e := a.Service().Providers()
			if e != nil || len(providers) != 1 || providers[0].ID != "fixture-provider" {
				t.Fatal("actual save missing", providers, e)
			}
			var replacement *providerDraft
			if how != "same-editor" {
				dismissProviderSaveAudit(t, a, tt, how)
				replacement = a.providerState.editor
			}
			apply()
			tt.Frame()
			if counts["RuntimeConfigChanged"] != 1 || counts["RefreshProviders"] != 1 {
				t.Errorf("persisted save lost successful completion effects: %v", counts)
			}
			if how != "same-editor" && (a.providerState.editor != replacement || replacement.edit.ID != "new-unsaved") {
				t.Error("old completion consumed replacement draft")
			}
			if how == "same-editor" && a.providerState.editor == old {
				t.Error("same editor success did not close")
			}
			drainRegressionUpdates(t, a, func() bool { return !a.settingsLoading })
		})
	}
}
func TestAuditProviderFailedCompletionDoesNotApplyOrConsumeNewDraft(t *testing.T) {
	a, tt, counts := providerSaveAuditFixture(t)
	p := filepath.Join(a.Options.Store.PiDir, "models.json")
	if err := os.MkdirAll(p, 0700); err != nil {
		t.Fatal(err)
	}
	apply := holdProviderSaveCompletion(t, a, tt)
	dismissProviderSaveAudit(t, a, tt, "escape")
	replacement := a.providerState.editor
	apply()
	tt.Frame()
	if counts["RuntimeConfigChanged"] != 0 || counts["RefreshProviders"] != 0 {
		t.Fatal("failed write reported success", counts)
	}
	if a.providerState.editor != replacement || replacement.edit.ID != "new-unsaved" {
		t.Fatal("failed old save consumed draft")
	}
	if st, err := os.Stat(p); err != nil || !st.IsDir() {
		t.Fatal("failure fixture mutated", st, err)
	}
}
func TestAuditProviderCancelBeforeSaveDoesNotWrite(t *testing.T) {
	for _, how := range []string{"escape", "scrim", "cancel-button"} {
		t.Run(how, func(t *testing.T) {
			a, tt, counts := providerSaveAuditFixture(t)
			switch how {
			case "escape":
				tt.Key(0, ui.KeyEscape)
			case "scrim":
				tt.ClickAt(5, 5)
			default:
				if err := tt.Click("Cancel"); err != nil {
					t.Fatal(err)
				}
			}
			if a.providerState.editor != nil {
				t.Fatal("cancel retained draft")
			}
			if _, err := os.Stat(filepath.Join(a.Options.Store.PiDir, "models.json")); !os.IsNotExist(err) {
				t.Fatal("cancel wrote config", err)
			}
			if len(counts) != 0 {
				t.Fatal(fmt.Sprint(counts))
			}
		})
	}
}

func TestProviderSavedCompletionPreservesWorkspaceBoundary(t *testing.T) {
	a, tt, counts := providerSaveAuditFixture(t)
	old := a.providerState.editor
	apply := holdProviderSaveCompletion(t, a, tt)
	// This transient marker is not included in the already-completed write.
	old.key = "unsent-fixture-draft-marker"
	dismissProviderSaveAudit(t, a, tt, "escape")
	replacement := a.providerState.editor
	// Cancel clears the old marker; restore one solely to observe success cleanup.
	old.key = "unsent-fixture-draft-marker"
	snapshot := a.Model
	snapshot.Workspace = t.TempDir()
	a.SetSnapshot(snapshot)
	tt.Frame()
	apply()
	tt.Frame()
	if counts["RuntimeConfigChanged"] != 0 || counts["RefreshProviders"] != 0 {
		t.Fatal("old workspace completion applied to another target", counts)
	}
	if a.providerState.editor != replacement || replacement.edit.ID != "new-unsaved" {
		t.Fatal("workspace boundary consumed new draft")
	}
	if old.key != "" {
		t.Fatal("completed old draft retained transient key")
	}
	providers, e := a.Service().Providers()
	if e != nil || len(providers) != 1 || providers[0].ID != "fixture-provider" {
		t.Fatal("global store mutation missing", providers, e)
	}
}
func TestProviderSavedCompletionSameWorkspaceTaskChangeKeepsNewDraft(t *testing.T) {
	a, tt, counts := providerSaveAuditFixture(t)
	apply := holdProviderSaveCompletion(t, a, tt)
	dismissProviderSaveAudit(t, a, tt, "escape")
	replacement := a.providerState.editor
	snapshot := a.Model
	snapshot.RuntimeID = "new-fixture-runtime"
	snapshot.SessionID = "new-fixture-task"
	a.SetSnapshot(snapshot)
	tt.Frame()
	apply()
	tt.Frame()
	if counts["RuntimeConfigChanged"] != 1 || counts["RefreshProviders"] != 1 {
		t.Fatal("same PiDir successful config not refreshed", counts)
	}
	if a.providerState.editor != replacement || replacement.edit.ID != "new-unsaved" {
		t.Fatal("task change consumed new draft")
	}
	drainRegressionUpdates(t, a, func() bool { return !a.settingsLoading })
}
func TestProviderFailedSaveKeepsOriginalErrorDraft(t *testing.T) {
	a, tt, counts := providerSaveAuditFixture(t)
	if err := os.MkdirAll(filepath.Join(a.Options.Store.PiDir, "models.json"), 0700); err != nil {
		t.Fatal(err)
	}
	old := a.providerState.editor
	apply := holdProviderSaveCompletion(t, a, tt)
	apply()
	tt.Frame()
	if a.providerState.editor != old || old.error == "" || old.edit.ID != "fixture-provider" {
		t.Fatal("failure lost original error/draft")
	}
	if counts["RuntimeConfigChanged"] != 0 || counts["RefreshProviders"] != 0 {
		t.Fatal("failure emitted success", counts)
	}
}
