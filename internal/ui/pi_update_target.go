package orbitui

import (
	"context"
	"github.com/charlzyx/orbit/internal/native"
)

// startPiLaunchUpdate must not let the process-global native installer target a
// different home than the selected Agent/Store. The checker seam is test-only
// injection; production uses the unchanged native launch-update worker.
func (x *Controller) startPiLaunchUpdate(check func(context.Context, native.PiLaunchUpdateOptions) <-chan struct{}) {
	if check == nil || x.piUpdateDone != nil || x.hostStopped() || !x.nativeSetupTargetMatches() {
		return
	}
	owner := x.capturePiUpdateOwner()
	x.piUpdateDone = check(x.background, native.PiLaunchUpdateOptions{Busy: func() bool { return x.maintenance.Load() || x.Agent.AnyBusy() }, BeginUpdate: func() (func(), bool) {
		release, err := x.beginRuntimeOperation()
		if err == nil && (x.hostStopped() || x.Agent.PiBinary() != "" || !x.nativeSetupTargetMatches()) {
			release()
			return nil, false
		}
		return release, err == nil
	}, OnEvent: func(event native.PiUpdateEvent) {
		x.enqueue(func() { x.applyPiUpdateEvent(owner, event) })
	}})
}
