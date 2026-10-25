package orbitui

import (
	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
)

type piUpdateOwner struct {
	manager    *agent.Manager
	binary     string
	generation uint64
}

func (x *Controller) capturePiUpdateOwner() piUpdateOwner {
	return piUpdateOwner{x.Agent, x.Agent.PiBinary(), x.setupProbe.generation}
}

// applyPiUpdateEvent runs only on the host owner. Installation state is global:
// switching tasks does not invalidate it, but a newer probe/binary/host does.
func (x *Controller) applyPiUpdateEvent(owner piUpdateOwner, event native.PiUpdateEvent) {
	if x.hostStopped() || x.Agent != owner.manager || x.Agent.PiBinary() != owner.binary || x.setupProbe.generation != owner.generation || x.setupInstall.pending != nil {
		return
	}
	switch event.State {
	case "updated", "failed", "unchanged":
		x.applyPiUpdateFacts(event)
	}
	switch event.State {
	case "updated":
		key := "pi_update.updated"
		if event.Installation.Bin != "" {
			key = "pi_update.updated_path"
		}
		body := orbitassets.Translate(x.UI.catalog, x.UI.english, key, map[string]string{"version": event.Installation.Version, "path": event.Installation.Bin})
		x.UI.toastStack.Push(native.ToastSuccess, "", body)
		if !event.Installation.OnPath {
			x.UI.Notify(x.UI.Tr("pi_update.installed_not_on_path"), true)
		}
		// Source finish_pi_update keeps every existing Pi process. In particular,
		// an idle task is not permission to restart it or replace its opaque owner.
	case "failed":
		x.UI.Notify("Pi update failed: "+event.Detail, true)
	case "migration_required":
		x.UI.notifyPiMigrationRequired(event.Latest)
	case "available", "deferred", string(native.PiUnsupportedManager), string(native.PiUnknownVersion):
		x.UI.notifyPiUpdateStatus(event)
	}
}

func (x *Controller) applyPiUpdateFacts(event native.PiUpdateEvent) {
	if len(event.Dependencies) == 0 {
		return
	}
	// A matching verified receipt supersedes the older in-flight probe. Cancel
	// it before adoption so its delayed completion cannot restore stale facts.
	x.invalidateSetupProbe()
	s := &x.UI.setup
	s.dependencies = append([]native.Dependency(nil), event.Dependencies...)
	s.installation = event.Installation
	s.loaded, s.error = true, ""
	facts := make([]native.HostFact, 0, len(s.facts)+3)
	for _, f := range s.facts {
		switch f.LabelKey {
		case "onboarding.pi_install", "onboarding.pi_path", "onboarding.pi_version":
			continue
		}
		facts = append(facts, f)
	}
	s.facts = append(facts, native.PiInstallFacts(event.Installation)...)
	// Required Node readiness remains independent. Never resume/start a task.
}
