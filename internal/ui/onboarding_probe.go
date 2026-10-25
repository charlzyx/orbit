package orbitui

import (
	"context"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
	"os"
	"time"
)

type setupProbeResult struct {
	dependencies []native.Dependency
	facts        []native.HostFact
}
type setupProbeState struct {
	generation uint64
	binary     string
	cancel     context.CancelFunc
	probe      func(context.Context, string) (setupProbeResult, error)
	resume     func(agent.WorkflowTarget, string) bool
}

func (x *Controller) probeSetup(ctx context.Context, binary string) (setupProbeResult, error) {
	if binary == "" {
		binary = os.Getenv("PI_BIN")
	}
	deps := native.CheckDependenciesFor(ctx, binary)
	p := native.PiInstallation{Kind: native.PiOther}
	for _, d := range deps {
		if d.Bin == "pi" {
			p.Bin, p.Version, p.Error = d.Path, d.Version, d.Error
			if p.Bin != "" {
				p.Kind, p.Root = native.ClassifyPiInstallation(p.Bin, x.Store.PiDir)
			}
		}
	}
	facts := native.RuntimeHostFacts(native.ProbeHost(ctx), x.Store.PiDir, x.Store.RootDir)
	facts = append(facts, native.PiInstallFacts(p)...)
	return setupProbeResult{deps, facts}, ctx.Err()
}
func (x *Controller) invalidateSetupProbe() {
	x.setupProbe.generation++
	if x.setupProbe.cancel != nil {
		x.setupProbe.cancel()
		x.setupProbe.cancel = nil
	}
	x.UI.setup.refreshing = false
	x.UI.setup.loaded = false
}
func (x *Controller) refreshSetup() {
	if x.hostStopped() || x.setupProbe.cancel != nil {
		return
	}
	s := &x.UI.setup
	wasMissing := s.loaded && (s.error != "" || !native.AllRequiredInstalled(s.dependencies))
	s.enabled, s.refreshing = true, true
	binary := x.Agent.PiBinary()
	target := x.Agent.CaptureWorkflowTarget()
	x.setupProbe.generation++
	generation := x.setupProbe.generation
	x.setupProbe.binary = binary
	ctx, cancel := context.WithTimeout(x.background, 20*time.Second)
	x.setupProbe.cancel = cancel
	probe := x.setupProbe.probe
	if probe == nil {
		probe = x.probeSetup
	}
	go func() {
		result, err := probe(ctx, binary)
		cancel()
		x.enqueue(func() {
			if x.hostStopped() || x.setupProbe.generation != generation || x.Agent.PiBinary() != binary {
				return
			}
			x.setupProbe.cancel = nil
			s.refreshing = false
			s.loaded = true
			if err != nil {
				s.error = err.Error()
				return
			}
			s.error = ""
			s.dependencies = append([]native.Dependency(nil), result.dependencies...)
			s.facts = append([]native.HostFact(nil), result.facts...)
			if wasMissing && native.AllRequiredInstalled(s.dependencies) && x.Agent.CaptureWorkflowTarget() == target {
				if !x.sessionGate.TryRLock() {
					return
				}
				defer x.sessionGate.RUnlock()
				resume := x.setupProbe.resume
				if resume == nil {
					resume = x.Agent.ResumeSetupFor
				}
				resume(target, binary)
			}
		})
	}()
}
