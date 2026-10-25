package orbitui

import "testing"

func TestWorkspaceServiceRefreshPreservesOwnerButMismatchAdvances(t *testing.T) {
	x := workspaceDialogFixture(t)
	initial := x.workspace
	before := x.workspaceBrowse.epoch
	x.changeWorkspaceServices(initial)
	if x.workspaceBrowse.epoch != before {
		t.Fatal("same service refresh changed owner epoch")
	}
	x.UI.Model.Workspace = "stale-projection"
	x.changeWorkspaceServices(initial)
	if x.workspaceBrowse.epoch == before {
		t.Fatal("model/service mismatch failed to advance owner")
	}
	before = x.workspaceBrowse.epoch
	x.changeWorkspaceServices(t.TempDir())
	x.changeWorkspaceServices(initial)
	if x.workspaceBrowse.epoch <= before+1 {
		t.Fatal("service roundtrip did not retain both ownership boundaries")
	}
}

func TestSameWorkspaceIntentRejectsOldSheetAndAttachmentOwner(t *testing.T) {
	x := workspaceDialogFixture(t)
	initial := x.workspace
	for x.UI.backdrop.loading {
		inlineAuditCompletion(t, x.UI)()
	}
	oldOwner := x.UI.attachmentOwner()
	initialEpoch := x.workspaceBrowse.epoch
	release := holdWorkspaceDialog(t, x, t.TempDir(), nil)
	x.SetWorkspace(initial) // explicit user intent, unlike an idempotent refresh
	if x.workspaceBrowse.epoch == initialEpoch {
		t.Fatal("same-workspace selection did not supersede pending intent")
	}
	if x.UI.ownsAttachment(oldOwner) {
		t.Fatal("old attachment owner survived explicit same-workspace intent")
	}
	release()
	drainControllerDialogUpdates(t, x, func() bool { return !x.workspaceBrowse.pending })
	if x.workspace != initial {
		t.Fatal("old native sheet replaced explicit same-workspace selection")
	}
}
