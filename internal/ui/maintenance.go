package orbitui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
)

// Readers are short UI mutations or explicit isolated generation jobs. A
// maintenance writer never blocks the native event loop: new work is rejected
// with its draft intact while the operation is in progress.
func sessionMutation(kind string) bool {
	switch kind {
	case "NewSession", "Submit", "SetModel", "SetThinking", "SetAccess", "SetWorkflow", "RenameSession", "CloneSession", "DeleteSession", "Login", "Logout", "RestartRuntime", "RuntimeChanged", "ChoosePiBinary", "RefreshMcpServers", "ReconnectMcpServers", "AutoTitleChanged":
		return true
	}
	return false
}

func (x *Controller) beginRuntimeOperation() (func(), error) {
	if !x.sessionGate.TryLock() {
		return nil, errors.New("Another session transition or maintenance operation is in progress")
	}
	if x.Agent.AnyBusy() {
		x.sessionGate.Unlock()
		return nil, errors.New("Finish or stop running tasks, including background tasks, before updating pi")
	}
	x.maintenance.Store(true)
	return func() { x.maintenance.Store(false); x.sessionGate.Unlock() }, nil
}

func (x *Controller) runRetention(manual bool) {
	if !x.Store.Retention().Enabled {
		if manual {
			x.UI.Notify("Enable the session cleanup policy first.", false)
		}
		return
	}
	if !x.sessionGate.TryLock() {
		if manual {
			x.UI.Notify("Another session transition or maintenance operation is in progress.", false)
		}
		return
	}
	x.maintenance.Store(true)
	x.UI.retentionBusy = true
	live := x.Agent.LiveSessionPaths()
	svc := workbench.New(x.Store, x.workspace)
	x.UI.Async(func() func() {
		defer func() { x.maintenance.Store(false); x.sessionGate.Unlock() }()
		out, err := svc.ApplyRetention(time.Now(), live)
		return func() {
			x.UI.retentionBusy = false
			x.UI.save(err)
			if out.Archived+out.Deleted > 0 {
				x.Agent.ReloadSessions()
				x.archived = map[string]bool{}
				for _, path := range x.Store.SessionSet(true) {
					x.archived[path] = true
				}
				if x.UI.page == "usage" {
					x.UI.refreshUsage()
				}
			}
			var parts []string
			if out.Archived > 0 {
				parts = append(parts, fmt.Sprintf("%d archived", out.Archived))
			}
			if out.Deleted > 0 {
				parts = append(parts, fmt.Sprintf("%d deleted", out.Deleted))
			}
			if len(out.SkippedLive) > 0 {
				parts = append(parts, fmt.Sprintf("%d open tasks kept", len(out.SkippedLive)))
			}
			if len(out.Failed) > 0 {
				parts = append(parts, fmt.Sprintf("%d failures", len(out.Failed)))
			}
			if len(parts) > 0 {
				x.UI.Notify(strings.Join(parts, " · "), len(out.Failed) > 0)
			} else if manual && err == nil {
				x.UI.Notify("No old sessions matched the cleanup policy.", false)
			}
		}
	})
}
