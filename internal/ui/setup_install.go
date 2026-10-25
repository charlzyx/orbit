package orbitui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
)

// This coordinator is deliberately not connected to a button yet. An explicit
// confirmation must capture this intent before installation can be admitted.
type setupInstallIntent struct {
	target     agent.WorkflowTarget
	binary     string
	generation uint64
	migrate    bool
}
type setupInstallReceipt struct {
	installation    native.PiInstallation
	dependencies    []native.Dependency
	output, warning string
	err             error
	verified        bool
}
type setupInstallOperation struct {
	intent        setupInstallIntent
	done          chan struct{}
	runnerSettled chan struct{}
	cancel        context.CancelFunc
}
type setupInstallState struct {
	pending *setupInstallOperation
	receipt setupInstallReceipt
}
type setupInstallBackend struct {
	supported    func() bool
	detect       func(context.Context) native.PiInstallation
	run          func(context.Context, bool) (native.CommandOutput, error)
	dependencies func(context.Context) []native.Dependency
}

func nativeSetupInstallBackend() setupInstallBackend {
	return setupInstallBackend{native.OfficialPiInstallSupported, native.DetectPiInstallation, native.InstallPi, native.CheckDependencies}
}
func (x *Controller) captureSetupInstallIntent(migrate bool) setupInstallIntent {
	return setupInstallIntent{x.Agent.CaptureWorkflowTarget(), x.Agent.PiBinary(), x.setupProbe.generation, migrate}
}
func (x *Controller) ownsSetupInstallIntent(i setupInstallIntent) bool {
	return !x.hostStopped() && x.Agent.CaptureWorkflowTarget() == i.target && x.Agent.PiBinary() == i.binary && x.setupProbe.generation == i.generation
}
func (x *Controller) beginSetupInstall(i setupInstallIntent, backend setupInstallBackend) error {
	if !x.ownsSetupInstallIntent(i) {
		return errors.New("Setup confirmation is stale; refresh and try again")
	}
	if i.binary != "" || os.Getenv("PI_BIN") != "" {
		return errors.New("A custom Pi launcher is selected; managed installation would not repair that launcher")
	}
	if x.setupInstall.pending != nil {
		return errors.New("A setup installation is already in progress")
	}
	if backend.supported == nil || backend.detect == nil || backend.run == nil || backend.dependencies == nil {
		return errors.New("Setup installation backend is incomplete")
	}
	if !backend.supported() {
		return errors.New("The official Pi installer supports macOS and Linux only")
	}
	release, err := x.beginRuntimeOperation()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(x.background, 10*time.Minute)
	op := &setupInstallOperation{intent: i, done: make(chan struct{}), runnerSettled: make(chan struct{}), cancel: cancel}
	x.setupInstall.pending = op
	x.setupInstall.receipt = setupInstallReceipt{}
	var once sync.Once
	finish := func() { once.Do(func() { release(); close(op.done) }) }
	// A cancellation request is not proof the installer stopped. Release only
	// after the runner has returned, including when it ignores cancellation.
	go func() {
		defer cancel()
		receipt := runSetupInstall(ctx, i, backend)
		close(op.runnerSettled)
		applied := make(chan struct{})
		x.enqueue(func() {
			defer close(applied)
			defer finish()
			if x.setupInstall.pending != op {
				return
			}
			x.setupInstall.pending = nil
			if x.ownsSetupInstallIntent(i) {
				x.setupInstall.receipt = receipt
			}
		})
		select {
		case <-applied:
		case <-x.background.Done():
			finish()
		}
	}()
	return nil
}
func runSetupInstall(ctx context.Context, i setupInstallIntent, b setupInstallBackend) setupInstallReceipt {
	var r setupInstallReceipt
	if err := ctx.Err(); err != nil {
		r.err = err
		return r
	}
	before := b.detect(ctx)
	if err := ctx.Err(); err != nil {
		r.err = err
		return r
	}
	switch native.PlanPiInstall(before) {
	case native.PiAlreadyManaged:
		r.err = errors.New("Pi is already managed; use update instead")
		return r
	case native.PiInstallConflict:
		if before.Kind != native.PiNPM {
			r.err = errors.New("Another package manager owns Pi; its installation will not be replaced")
			return r
		}
		if !i.migrate {
			r.err = errors.New("Replacing the npm installation requires explicit migration confirmation")
			return r
		}
	}
	output, runErr := b.run(ctx, i.migrate)
	r.output = strings.TrimSpace(string(output.Stdout) + "\n" + string(output.Stderr))
	if err := ctx.Err(); err != nil {
		r.err = err
		return r
	}
	r.installation = b.detect(ctx)
	r.dependencies = append([]native.Dependency(nil), b.dependencies(ctx)...)
	if err := ctx.Err(); err != nil {
		r.err = err
		return r
	}
	if runErr != nil {
		r.err = runErr
		return r
	}
	if output.ExitCode != 0 {
		r.err = fmt.Errorf("Installer exited with status %d", output.ExitCode)
		return r
	}
	if !r.installation.Installed() || r.installation.Kind != native.PiManaged || r.installation.Version == "" || r.installation.Error != "" {
		r.err = errors.New("Installer finished, but a working managed Pi installation could not be verified")
		return r
	}
	r.verified = true
	if !r.installation.OnPath {
		r.warning = "The managed Pi installation is not on the process PATH"
	}
	return r
}
