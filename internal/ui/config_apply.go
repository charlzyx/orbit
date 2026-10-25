package orbitui

import (
	"errors"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/mcpui"
)

// Config apply is an admitted request followed by an observed bootstrap, not a
// success toast on Restart's nil return. All fields are owned by the UI thread.
type configApplyTarget struct {
	runtime, workspace, path, session string
	token                             mcpui.ApplyToken
	epoch                             uint64
	mcp                               bool
}
type configApplyFlight struct {
	target  configApplyTarget
	runtime string
}

func (x *Controller) scheduleMCPApply(delay time.Duration) {
	x.scheduleConfigApply(delay)
	x.configApplyTarget.mcp = true
}
func (x *Controller) scheduleConfigApply(delay time.Duration) {
	s := x.Agent.Snapshot()
	// During a replacement bootstrap the authoritative ID has not arrived yet;
	// retain its known saved identity for a newer edit queued behind that reload.
	if f := x.configApplyFlight; f != nil && s.RuntimeID == f.runtime && s.ActiveSession.Path == f.target.path && s.ActiveSession.ID == "" {
		s.ActiveSession.ID = f.target.session
	}
	x.configApplyEpoch++
	x.configApplyTarget = configApplyTarget{runtime: s.RuntimeID, workspace: s.ActiveSession.Workspace, path: s.ActiveSession.Path, session: s.ActiveSession.ID, epoch: x.configApplyEpoch}
	if x.MCP != nil {
		x.configApplyTarget.token = x.MCP.ApplyToken()
	}
	x.configApplyAfter = time.Now().Add(delay)
}
func (t configApplyTarget) matches(s agent.Snapshot) bool {
	return t.workspace == s.ActiveSession.Workspace && t.path == s.ActiveSession.Path && t.session == s.ActiveSession.ID
}
func (x *Controller) tickConfigApply(now time.Time) {
	select {
	case <-x.stop:
		return
	default:
	}
	s := x.Agent.Snapshot()
	process := x.Agent.RuntimeSnapshot().Active
	if f := x.configApplyFlight; f != nil {
		// Track the exact replacement until it settles even if a newer edit
		// supersedes its ACK. Otherwise a failed bootstrap would lose retry eligibility.
		if s.RuntimeID != f.runtime || s.ActiveSession.Workspace != f.target.workspace || s.ActiveSession.Path != f.target.path || s.Connected && !f.target.matches(s) {
			x.configApplyFlight = nil
		} else if process != nil && process.Key == f.runtime && !process.Loading && process.Error != "" {
			x.configApplyFailedRuntime = s.RuntimeID
			// Preserve the newer desired token, but require explicit retry after failure.
			if x.configApplyTarget.runtime == f.runtime {
				x.configApplyAfter = time.Time{}
				if x.configApplyTarget.mcp && x.MCP != nil {
					x.MCP.ApplyFailed(x.configApplyTarget.token)
				}
			}
			if f.target.mcp && x.MCP != nil {
				x.MCP.ApplyFailed(f.target.token)
			}
			x.configApplyFlight = nil
		} else if s.Connected {
			x.configApplyFailedRuntime = ""
			if f.target.epoch == x.configApplyEpoch && f.target.mcp && x.MCP != nil {
				x.MCP.AppliedIfCurrent(f.target.token)
			}
			if f.target.epoch == x.configApplyEpoch && (!f.target.mcp || x.MCP == nil || x.MCP.ApplyCurrent(f.target.token)) {
				x.UI.Notify("Pi configuration reloaded for the current session. Other open sessions keep their running configuration.", false)
			}
			x.configApplyFlight = nil
		}
	}
	if x.configApplyAfter.IsZero() || now.Before(x.configApplyAfter) {
		return
	}
	t := x.configApplyTarget
	if t.runtime == "" || t.path == "" || t.session == "" || s.RuntimeID != t.runtime || s.ActiveSession.Workspace != t.workspace || s.ActiveSession.Path != t.path || s.Connected && !t.matches(s) || t.mcp && x.MCP != nil && !x.MCP.ApplyCurrent(t.token) {
		x.configApplyAfter = time.Time{}
		return
	}
	// Rust applies to the active settled process. An unrelated parked run does
	// not block this reload and is not stopped. An initial failed client is not
	// a valid target; an explicitly retried, previously failed reload is.
	if process == nil || process.Loading || process.Busy {
		return
	}
	if (!s.Connected || !process.Alive) && !(t.mcp && t.runtime == x.configApplyFailedRuntime) {
		x.configApplyAfter = time.Time{}
		return
	}
	if !x.sessionGate.TryRLock() {
		return
	}
	err := x.Agent.RestartIdle(t.runtime, t.path, t.session)
	x.sessionGate.RUnlock()
	if errors.Is(err, agent.ErrReloadDeferred) {
		return
	}
	x.configApplyAfter = time.Time{}
	if err != nil {
		x.UI.save(err)
		if t.mcp && x.MCP != nil {
			x.MCP.ApplyFailed(t.token)
		}
		return
	}
	next := x.Agent.Snapshot()
	if next.RuntimeID == t.runtime || next.RuntimeID == "" {
		return
	}
	x.configApplyFlight = &configApplyFlight{target: t, runtime: next.RuntimeID}
	x.UI.Notify("Pi configuration reload requested for the current session. Other open sessions keep their running configuration.", false)
}
